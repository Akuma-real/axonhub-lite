package orchestrator

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/ent/enttest"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/xcache"
	"github.com/looplj/axonhub/internal/server/biz"
	"github.com/looplj/axonhub/llm"
	"github.com/looplj/axonhub/llm/httpclient"
	"github.com/looplj/axonhub/llm/pipeline"
	"github.com/looplj/axonhub/llm/pipeline/stream"
	"github.com/looplj/axonhub/llm/streams"
	"github.com/looplj/axonhub/llm/transformer/openai"
)

type mockStrategy struct {
	name  string
	score float64
}

func (m *mockStrategy) Score(ctx context.Context, channel *biz.Channel) float64 {
	return m.score
}

func (m *mockStrategy) ScoreWithDebug(ctx context.Context, channel *biz.Channel) (float64, StrategyScore) {
	return m.score, StrategyScore{
		StrategyName: m.name,
		Score:        m.score,
		Details:      map[string]any{"fixed_score": m.score},
	}
}

func (m *mockStrategy) Name() string {
	return m.name
}

type mockMetricsProvider struct {
	metrics map[int]*biz.AggregatedMetrics
	err     error
}

func (m *mockMetricsProvider) GetChannelMetrics(ctx context.Context, channelID int) (*biz.AggregatedMetrics, error) {
	if m.err != nil {
		return nil, m.err
	}

	if metrics, ok := m.metrics[channelID]; ok {
		return metrics, nil
	}

	return &biz.AggregatedMetrics{}, nil
}

type mockRetryPolicyProvider struct {
	policy *biz.RetryPolicy
}

func (m *mockRetryPolicyProvider) RetryPolicyOrDefault(ctx context.Context) *biz.RetryPolicy {
	return m.policy
}

type mockSystemService struct {
	retryPolicy *biz.RetryPolicy
}

func (m *mockSystemService) RetryPolicyOrDefault(ctx context.Context) *biz.RetryPolicy {
	if m.retryPolicy != nil {
		return m.retryPolicy
	}

	return &biz.RetryPolicy{
		Enabled:                 false,
		MaxChannelRetries:       3,
		MaxSingleChannelRetries: 2,
		RetryDelayMs:            1000,
	}
}

type mockSelectionTracker struct {
	selections map[int]int
}

func (m *mockSelectionTracker) IncrementChannelSelection(channelID int) {
	if m.selections == nil {
		m.selections = make(map[int]int)
	}

	m.selections[channelID]++
}

type mockTraceProvider struct {
	lastSuccessChannel map[int]int
	err                error
}

func (m *mockTraceProvider) GetLastSuccessfulChannelID(ctx context.Context, traceID int) (int, error) {
	if m.err != nil {
		return 0, m.err
	}

	if channelID, ok := m.lastSuccessChannel[traceID]; ok {
		return channelID, nil
	}

	return 0, nil
}

type mockTransformer struct {
	aggregatedResponse []byte
	aggregatedMeta     llm.ResponseMeta
	aggregatedErr      error
	apiFormat          llm.APIFormat
}

func (m *mockTransformer) TransformRequest(ctx context.Context, req *llm.Request) (*httpclient.Request, error) {
	body, err := json.Marshal(map[string]any{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": 0.5,
		"max_tokens":  1000,
	})
	if err != nil {
		return nil, err
	}

	return &httpclient.Request{
		Method: "POST",
		URL:    "https://api.example.com/v1/chat/completions",
		Body:   body,
	}, nil
}

func (m *mockTransformer) TransformResponse(ctx context.Context, resp *httpclient.Response) (*llm.Response, error) {
	return &llm.Response{}, nil
}

func (m *mockTransformer) TransformStream(ctx context.Context, req *httpclient.Request, stream streams.Stream[*httpclient.StreamEvent]) (streams.Stream[*llm.Response], error) {
	return nil, nil
}

func (m *mockTransformer) TransformError(ctx context.Context, err *httpclient.Error) *llm.ResponseError {
	return nil
}

func (m *mockTransformer) AggregateStreamChunks(ctx context.Context, _ *httpclient.Request, chunks []*httpclient.StreamEvent) ([]byte, llm.ResponseMeta, error) {
	return m.aggregatedResponse, m.aggregatedMeta, m.aggregatedErr
}

func (m *mockTransformer) APIFormat() llm.APIFormat {
	if m.apiFormat != "" {
		return m.apiFormat
	}

	return llm.APIFormatOpenAIChatCompletion
}

type mockExecutor struct {
	response      *httpclient.Response
	streamEvents  []*httpclient.StreamEvent
	err           error
	requestCalled bool
	lastRequest   *httpclient.Request
}

type mockStream struct {
	events     []*httpclient.StreamEvent
	currentIdx int
	closed     bool
	err        error
}

func (m *mockStream) Next() bool {
	if m.currentIdx >= len(m.events) {
		return false
	}
	m.currentIdx++
	return true
}

func (m *mockStream) Current() *httpclient.StreamEvent {
	if m.currentIdx == 0 || m.currentIdx > len(m.events) {
		return nil
	}
	return m.events[m.currentIdx-1]
}

func (m *mockStream) Err() error {
	return m.err
}

func (m *mockStream) Close() error {
	m.closed = true
	return nil
}

func (m *mockExecutor) Do(ctx context.Context, request *httpclient.Request) (*httpclient.Response, error) {
	m.requestCalled = true
	m.lastRequest = request

	if m.err != nil {
		return nil, m.err
	}

	return m.response, nil
}

func (m *mockExecutor) DoStream(ctx context.Context, request *httpclient.Request) (streams.Stream[*httpclient.StreamEvent], error) {
	m.requestCalled = true
	m.lastRequest = request

	if m.err != nil {
		return nil, m.err
	}

	return streams.SliceStream(m.streamEvents), nil
}

type staticChannelSelector struct {
	candidates []*ChannelModelsCandidate
}

func (s *staticChannelSelector) Select(ctx context.Context, req *llm.Request) ([]*ChannelModelsCandidate, error) {
	return s.candidates, nil
}

func setupTest(t *testing.T) (context.Context, *ent.Client) {
	t.Helper()

	ctx := authz.WithTestBypass(context.Background())
	client := enttest.NewEntClient(t, "sqlite3", "file:ent?mode=memory&_fk=0")
	t.Cleanup(func() { client.Close() })

	ctx = ent.NewContext(ctx, client)
	return ctx, client
}

func newTestSystemService(client *ent.Client) *biz.SystemService {
	return biz.NewSystemService(biz.SystemServiceParams{
		CacheConfig: xcache.Config{Mode: xcache.ModeMemory},
		Ent:         client,
	})
}

func newTestChannelService(client *ent.Client) *biz.ChannelService {
	systemService := newTestSystemService(client)
	return biz.NewChannelService(biz.ChannelServiceParams{
		Ent:           client,
		SystemService: systemService,
	})
}

func newTestChannelServiceForChannels(client *ent.Client) *biz.ChannelService {
	return newTestChannelService(client)
}

func newTestModelService(client *ent.Client) *biz.ModelService {
	return biz.NewModelService(biz.ModelServiceParams{
		Ent: client,
	})
}

func newTestRequestService(client *ent.Client) *biz.RequestService {
	systemService := newTestSystemService(client)
	return newTestRequestServiceForChannels(client, systemService)
}

func newTestRequestServiceForChannels(client *ent.Client, systemService *biz.SystemService) *biz.RequestService {
	channelService := biz.NewChannelServiceForTest(client)
	usageLogService := biz.NewUsageLogService(client, systemService, channelService)
	return biz.NewRequestService(client, systemService, usageLogService, biz.NewLiveStreamRegistry())
}

func setupTestServices(t *testing.T, client *ent.Client) (*biz.ChannelService, *biz.RequestService, *biz.SystemService, *biz.UsageLogService) {
	t.Helper()

	systemService := newTestSystemService(client)
	channelService := biz.NewChannelServiceForTest(client)
	usageLogService := biz.NewUsageLogService(client, systemService, channelService)
	requestService := biz.NewRequestService(client, systemService, usageLogService, biz.NewLiveStreamRegistry())

	return channelService, requestService, systemService, usageLogService
}

func newTestLoadBalancedSelector(
	channelService *biz.ChannelService,
	client *ent.Client,
	systemService *biz.SystemService,
	requestService *biz.RequestService,
) CandidateSelector {
	strategies := []LoadBalanceStrategy{
		NewTraceAwareStrategy(requestService),
		NewErrorAwareStrategy(channelService),
		NewWeightRoundRobinStrategy(channelService),
		NewLatencyAwareStrategy(channelService),
	}
	loadBalancer := NewLoadBalancer(systemService, nil, strategies...)

	modelService := newTestModelService(client)
	baseSelector := NewDefaultSelector(channelService, modelService, systemService)

	return WithLoadBalancedSelector(baseSelector, loadBalancer, systemService)
}

func createTestChannels(t *testing.T, ctx context.Context, client *ent.Client) []*ent.Channel {
	t.Helper()

	channels := make([]*ent.Channel, 0)

	ch1, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("High Weight Channel").
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{APIKey: "test-key-1"}).
		SetSupportedModels([]string{"gpt-4", "gpt-3.5-turbo"}).
		SetDefaultTestModel("gpt-4").
		SetOrderingWeight(100).
		SetStatus(channel.StatusEnabled).
		Save(ctx)
	require.NoError(t, err)
	channels = append(channels, ch1)

	ch2, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("Medium Weight Channel").
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{APIKey: "test-key-2"}).
		SetSupportedModels([]string{"gpt-4", "gpt-3.5-turbo"}).
		SetDefaultTestModel("gpt-4").
		SetOrderingWeight(50).
		SetStatus(channel.StatusEnabled).
		Save(ctx)
	require.NoError(t, err)
	channels = append(channels, ch2)

	ch3, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("Low Weight Channel").
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{APIKey: "test-key-3"}).
		SetSupportedModels([]string{"gpt-4", "gpt-3.5-turbo"}).
		SetDefaultTestModel("gpt-4").
		SetOrderingWeight(25).
		SetStatus(channel.StatusEnabled).
		Save(ctx)
	require.NoError(t, err)
	channels = append(channels, ch3)

	ch4, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("Disabled Channel").
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{APIKey: "test-key-4"}).
		SetSupportedModels([]string{"gpt-4", "gpt-3.5-turbo"}).
		SetDefaultTestModel("gpt-4").
		SetOrderingWeight(75).
		SetStatus(channel.StatusDisabled).
		Save(ctx)
	require.NoError(t, err)
	channels = append(channels, ch4)

	return channels
}

func getCandidateNameByID(result []*ChannelModelsCandidate, channelID int) string {
	for _, c := range result {
		if c.Channel.ID == channelID {
			return c.Channel.Name
		}
	}

	return "unknown"
}

func createTestChannel(t *testing.T, ctx context.Context, client *ent.Client) *ent.Channel {
	t.Helper()

	ch, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("Test OpenAI Channel").
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{APIKey: "test-api-key"}).
		SetSupportedModels([]string{"gpt-4", "gpt-3.5-turbo"}).
		SetDefaultTestModel("gpt-3.5-turbo").
		Save(ctx)
	require.NoError(t, err)

	return ch
}

func buildMockOpenAIResponse(id, model, content string, promptTokens, completionTokens int) []byte {
	resp := map[string]any{
		"id":      id,
		"object":  "chat.completion",
		"created": 1234567890,
		"model":   model,
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]any{
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
			"total_tokens":      promptTokens + completionTokens,
		},
	}

	body, _ := json.Marshal(resp)
	return body
}

func buildTestRequest(model, content string, stream bool) *httpclient.Request {
	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]any{
			{
				"role":    "user",
				"content": content,
			},
		},
		"stream": stream,
	}

	body, _ := json.Marshal(reqBody)

	return &httpclient.Request{
		Method: "POST",
		URL:    "/v1/chat/completions",
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: body,
	}
}

func channelsToTestCandidates(channels []*biz.Channel, model string) []*ChannelModelsCandidate {
	candidates := make([]*ChannelModelsCandidate, 0, len(channels))
	for _, ch := range channels {
		entries := ch.GetModelEntries()

		entry, ok := entries[model]
		if !ok {
			continue
		}

		candidates = append(candidates, &ChannelModelsCandidate{
			Channel:  ch,
			Priority: 0,
			Models:   []biz.ChannelModelEntry{entry},
		})
	}

	return candidates
}

func newTestOrchestrator(
	t *testing.T,
	channelSelector CandidateSelector,
	client *ent.Client,
	executor pipeline.Executor,
) *ChatCompletionOrchestrator {
	t.Helper()

	channelService, requestService, systemService, usageLogService := setupTestServices(t, client)

	return &ChatCompletionOrchestrator{
		channelSelector:       channelSelector,
		Inbound:               openai.NewInboundTransformer(),
		RequestService:        requestService,
		ChannelService:        channelService,
		SystemService:         systemService,
		UsageLogService:       usageLogService,
		LiveStreamRegistry:    biz.NewLiveStreamRegistry(),
		PipelineFactory:       pipeline.NewFactory(executor),
		ModelMapper:           NewModelMapper(),
		channelLimiterManager: NewChannelLimiterManager(),
		Middlewares: []pipeline.Middleware{
			stream.EnsureUsage(),
		},
	}
}
