package codex

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/looplj/axonhub/llm"
	"github.com/looplj/axonhub/llm/auth"
	"github.com/looplj/axonhub/llm/httpclient"
	"github.com/looplj/axonhub/llm/oauth"
	"github.com/looplj/axonhub/llm/pipeline"
	"github.com/looplj/axonhub/llm/streams"
	"github.com/looplj/axonhub/llm/transformer"
	"github.com/looplj/axonhub/llm/transformer/openai/responses"
	"github.com/looplj/axonhub/llm/transformer/shared"
)

const (
	codexBaseURL                     = "https://chatgpt.com/backend-api/codex#"
	codexAPIURL                      = "https://chatgpt.com/backend-api/codex/responses"
	codexCompactEmulatedKey          = "codex_compact_emulated"
	codexCompactInstructionsKey      = "codex_compact_instructions"
	codexCompactOriginalAPIFormatKey = "codex_compact_original_api_format"
	compactModeEmulated              = "emulated"
	compactModeNative                = "native"
)

// OutboundTransformer implements transformer.Outbound for Codex proxy.
// It always talks to the Codex Responses upstream (SSE only) and adapts requests accordingly.
//
//nolint:containedctx // It is used as a transformer.
type OutboundTransformer struct {
	tokens oauth.TokenGetter

	// reuse existing Responses outbound for payload building.
	responsesOutbound *responses.OutboundTransformer
}

var (
	_ transformer.Outbound               = (*OutboundTransformer)(nil)
	_ pipeline.ChannelCustomizedExecutor = (*OutboundTransformer)(nil)
)

type Params struct {
	TokenProvider oauth.TokenGetter
	BaseURL       string
}

func NewOutboundTransformer(params Params) (*OutboundTransformer, error) {
	if params.TokenProvider == nil {
		return nil, errors.New("token provider is required")
	}

	baseURL := params.BaseURL
	// Compatibility with old codex channel base url.
	if baseURL == "" || baseURL == "https://api.openai.com/v1" {
		baseURL = codexBaseURL
	}

	// The underlying responses outbound requires baseURL/apiKey. We only need its request body logic.
	// Use a dummy config and then override URL/auth.
	ro, err := responses.NewOutboundTransformerWithConfig(&responses.Config{
		BaseURL:        baseURL,
		APIKeyProvider: auth.NewStaticKeyProvider("dummy"),
	})
	if err != nil {
		return nil, err
	}

	return &OutboundTransformer{
		tokens:            params.TokenProvider,
		responsesOutbound: ro,
	}, nil
}

func (t *OutboundTransformer) APIFormat() llm.APIFormat {
	return llm.APIFormatOpenAIResponse
}

func (t *OutboundTransformer) TransformError(ctx context.Context, rawErr *httpclient.Error) *llm.ResponseError {
	return t.responsesOutbound.TransformError(ctx, rawErr)
}

func (t *OutboundTransformer) TransformRequest(ctx context.Context, llmReq *llm.Request) (*httpclient.Request, error) {
	if llmReq == nil {
		return nil, errors.New("request is nil")
	}

	rawSessionID := ""
	rawOriginator := ""
	rawUserAgent := ""
	rawTurnMetadata := ""

	var rawHeaders http.Header

	if llmReq.RawRequest != nil && llmReq.RawRequest.Headers != nil {
		rawHeaders = llmReq.RawRequest.Headers
		rawSessionID = llmReq.RawRequest.Headers.Get(SessionHeader)
		rawOriginator = llmReq.RawRequest.Headers.Get("Originator")
		rawUserAgent = llmReq.RawRequest.Headers.Get("User-Agent")
		rawTurnMetadata = llmReq.RawRequest.Headers.Get(TurnMetadataHeader)
	}

	creds, err := t.tokens.Get(ctx)
	if err != nil {
		return nil, err
	}

	// Parse account ID from access token JWT.
	accountID := ExtractChatGPTAccountIDFromJWT(creds.AccessToken)

	// Clone request so we do not mutate upstream pipeline state.
	reqCopy := *llmReq

	// Codex expects Responses API payload with some strict rules.
	// Always enable stream except for compact requests and disable store.
	//nolint: exhaustive // We only care about compact requests.
	switch reqCopy.RequestType {
	case llm.RequestTypeCompact:
		reqCopy.Stream = lo.ToPtr(false)
		if compactModeFromRequest(&reqCopy) != compactModeNative {
			prepareEmulatedCompactRequest(&reqCopy)
		}
	default:
		reqCopy.Stream = lo.ToPtr(true)
	}

	reqCopy.Store = lo.ToPtr(false)

	// Codex recommends parallel tool calls.
	reqCopy.ParallelToolCalls = lo.ToPtr(true)

	// Ask for encrypted reasoning content so the downstream can surface reasoning blocks.
	if reqCopy.TransformerMetadata == nil {
		reqCopy.TransformerMetadata = map[string]any{}
	}

	if _, ok := reqCopy.TransformerMetadata["include"]; !ok {
		reqCopy.TransformerMetadata["include"] = []string{"reasoning.encrypted_content"}
	}

	if reqCopy.ReasoningSummary == nil || *reqCopy.ReasoningSummary == "" {
		// Enable reasoning summary for Codex CLI requests.
		reqCopy.ReasoningSummary = lo.ToPtr("auto")
	}

	// Codex Responses rejects token limit fields, so strip them out.
	reqCopy.MaxCompletionTokens = nil
	reqCopy.MaxTokens = nil

	reqCopy.Metadata = nil

	reqCopy.TransformOptions.ArrayInputs = lo.ToPtr(true)

	hreq, err := t.responsesOutbound.TransformRequest(ctx, &reqCopy)
	if err != nil {
		return nil, err
	}

	// Overwrite auth.
	hreq.Auth = &httpclient.AuthConfig{Type: httpclient.AuthTypeBearer, APIKey: creds.AccessToken}
	// Compact requests are emulated through non-streaming Responses calls.
	if llmReq.RequestType == llm.RequestTypeCompact {
		hreq.Headers.Set("Accept", "application/json")
	} else {
		hreq.Headers.Set("Accept", "text/event-stream")
	}

	hreq.Headers.Del("User-Agent")

	if rawOriginator != "" {
		hreq.Headers.Set("Originator", rawOriginator)
	} else {
		hreq.Headers.Set("Originator", AxonHubOriginator)
	}

	if rawUserAgent != "" {
		hreq.Headers.Set("User-Agent", rawUserAgent)
	}

	for _, header := range PassthroughHeaders {
		if value := rawHeaders.Get(header); value != "" {
			hreq.Headers.Set(header, value)
		}
	}

	if rawSessionID != "" {
		hreq.Headers.Set(SessionHeader, rawSessionID)
	} else if sessionID := ExtractSessionIDFromTurnMetadata(rawTurnMetadata); sessionID != "" {
		hreq.Headers.Set(SessionHeader, sessionID)
	} else if hreq.Headers.Get(SessionHeader) == "" {
		if sessionID, ok := shared.GetSessionID(ctx); ok {
			hreq.Headers.Set(SessionHeader, sessionID)
		} else {
			hreq.Headers.Set(SessionHeader, uuid.NewString())
		}
	}

	if accountID != "" {
		hreq.Headers.Set("Chatgpt-Account-Id", accountID)
	}

	return hreq, nil
}

func (t *OutboundTransformer) TransformResponse(ctx context.Context, httpResp *httpclient.Response) (*llm.Response, error) {
	// Codex upstream returns Responses API response.
	llmResp, err := t.responsesOutbound.TransformResponse(ctx, httpResp)
	if err != nil {
		return nil, err
	}

	if httpResp == nil || httpResp.Request == nil || httpResp.Request.TransformerMetadata == nil {
		return llmResp, nil
	}

	emulated, _ := httpResp.Request.TransformerMetadata[codexCompactEmulatedKey].(bool)
	if !emulated {
		return llmResp, nil
	}

	instructions, _ := httpResp.Request.TransformerMetadata[codexCompactInstructionsKey].(string)
	return wrapEmulatedCompactResponse(llmResp, instructions), nil
}

func (t *OutboundTransformer) TransformStream(ctx context.Context, req *httpclient.Request, streamIn streams.Stream[*httpclient.StreamEvent]) (streams.Stream[*llm.Response], error) {
	return t.responsesOutbound.TransformStream(ctx, req, streamIn)
}

func (t *OutboundTransformer) AggregateStreamChunks(ctx context.Context, req *httpclient.Request, chunks []*httpclient.StreamEvent) ([]byte, llm.ResponseMeta, error) {
	return t.responsesOutbound.AggregateStreamChunks(ctx, req, chunks)
}

func (t *OutboundTransformer) CustomizeExecutor(executor pipeline.Executor) pipeline.Executor {
	return &codexExecutor{
		inner:       executor,
		transformer: t,
	}
}

type codexExecutor struct {
	inner       pipeline.Executor
	transformer *OutboundTransformer
}

func (e *codexExecutor) Do(ctx context.Context, request *httpclient.Request) (*httpclient.Response, error) {
	if request.RequestType == string(llm.RequestTypeCompact) || isEmulatedCompactRequest(request) {
		return e.inner.Do(ctx, request)
	}

	stream, err := e.inner.DoStream(ctx, request)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = stream.Close()
	}()

	var chunks []*httpclient.StreamEvent

	for stream.Next() {
		ev := stream.Current()
		if ev == nil {
			continue
		}

		chunks = append(chunks, &httpclient.StreamEvent{
			Type:        ev.Type,
			LastEventID: ev.LastEventID,
			Data:        append([]byte(nil), ev.Data...),
		})
	}

	if err := stream.Err(); err != nil {
		return nil, err
	}

	body, _, err := e.transformer.AggregateStreamChunks(ctx, request, chunks)
	if err != nil {
		return nil, err
	}

	return &httpclient.Response{
		StatusCode: http.StatusOK,
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body:    body,
		Request: request,
	}, nil
}

func (e *codexExecutor) DoStream(ctx context.Context, request *httpclient.Request) (streams.Stream[*httpclient.StreamEvent], error) {
	return e.inner.DoStream(ctx, request)
}

func isEmulatedCompactRequest(request *httpclient.Request) bool {
	if request == nil || request.TransformerMetadata == nil {
		return false
	}

	emulated, _ := request.TransformerMetadata[codexCompactEmulatedKey].(bool)
	return emulated
}

func compactModeFromRequest(req *llm.Request) string {
	if req == nil || req.TransformerMetadata == nil {
		return compactModeEmulated
	}

	mode, _ := req.TransformerMetadata["codex_compact_mode"].(string)
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case compactModeNative:
		return compactModeNative
	default:
		return compactModeEmulated
	}
}

func prepareEmulatedCompactRequest(req *llm.Request) {
	if req == nil || req.Compact == nil {
		return
	}

	if req.TransformerMetadata == nil {
		req.TransformerMetadata = map[string]any{}
	}

	req.TransformerMetadata[codexCompactEmulatedKey] = true
	req.TransformerMetadata[codexCompactInstructionsKey] = req.Compact.Instructions
	req.TransformerMetadata[codexCompactOriginalAPIFormatKey] = string(req.APIFormat)

	req.Messages = append([]llm.Message(nil), req.Compact.Input...)
	if req.Compact.Instructions != "" {
		req.Messages = append([]llm.Message{{
			Role: "system",
			Content: llm.MessageContent{
				Content: lo.ToPtr(req.Compact.Instructions),
			},
		}}, req.Messages...)
	}
	req.PromptCacheKey = lo.ToPtr(req.Compact.PromptCacheKey)
	req.RequestType = llm.RequestTypeChat
	req.APIFormat = llm.APIFormatOpenAIResponse
	req.Compact = nil
}

func wrapEmulatedCompactResponse(resp *llm.Response, instructions string) *llm.Response {
	if resp == nil {
		return nil
	}

	output := make([]llm.Message, 0, len(resp.Choices))
	for _, choice := range resp.Choices {
		if choice.Message != nil {
			output = append(output, *choice.Message)
		}
	}

	if len(output) == 0 {
		output = []llm.Message{{
			Role: "assistant",
			Content: llm.MessageContent{
				Content: lo.ToPtr(""),
			},
		}}
	}

	resp.RequestType = llm.RequestTypeCompact
	resp.APIFormat = llm.APIFormatOpenAIResponseCompact
	resp.Object = "response.compaction"
	resp.Compact = &llm.CompactResponse{
		ID:           resp.ID,
		CreatedAt:    resp.Created,
		Object:       "response.compaction",
		Instructions: instructions,
		Output:       output,
	}

	return resp
}
