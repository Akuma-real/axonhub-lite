package gql

import (
	"errors"

	"github.com/99designs/gqlgen/graphql"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/server/biz"
	"github.com/looplj/axonhub/internal/server/gc"
	"github.com/looplj/axonhub/internal/server/orchestrator"
	"github.com/looplj/axonhub/internal/server/scheduler"
	"github.com/looplj/axonhub/llm/httpclient"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// ErrNotOwner is returned when a non-owner user attempts an owner-only operation.
var ErrNotOwner = errors.New("permission denied: owner access required")

// Resolver is the resolver root.
type Resolver struct {
	client                         *ent.Client
	authService                    *biz.AuthService
	apiKeyService                  *biz.APIKeyService
	userService                    *biz.UserService
	systemService                  *biz.SystemService
	channelService                 *biz.ChannelService
	requestService                 *biz.RequestService
	channelOverrideTemplateService *biz.ChannelOverrideTemplateService
	modelService                   *biz.ModelService
	channelProbeService            *biz.ChannelProbeService
	providerQuotaService           *biz.ProviderQuotaService
	scheduler                      *scheduler.Scheduler
	modelFetcher                   *biz.ModelFetcher
	defaultSelector                *orchestrator.DefaultSelector
	candidateSelectorDiagnostics   *orchestrator.CandidateSelectorDiagnostics
	channelLimiterManager          *orchestrator.ChannelLimiterManager
	TestChannelOrchestrator        *orchestrator.TestChannelOrchestrator
	gcWorker                       *gc.Worker
}

// NewSchema creates a graphql executable schema.
func NewSchema(
	client *ent.Client,
	authService *biz.AuthService,
	apiKeyService *biz.APIKeyService,
	userService *biz.UserService,
	systemService *biz.SystemService,
	channelService *biz.ChannelService,
	requestService *biz.RequestService,
	usageLogService *biz.UsageLogService,
	channelOverrideTemplateService *biz.ChannelOverrideTemplateService,
	modelService *biz.ModelService,
	channelProbeService *biz.ChannelProbeService,
	providerQuotaService *biz.ProviderQuotaService,
	scheduler *scheduler.Scheduler,
	defaultSelector *orchestrator.DefaultSelector,
	candidateSelectorDiagnostics *orchestrator.CandidateSelectorDiagnostics,
	channelLimiterManager *orchestrator.ChannelLimiterManager,
	httpClient *httpclient.HttpClient,
	gcWorker *gc.Worker,
) graphql.ExecutableSchema {
	modelFetcher := biz.NewModelFetcher(httpClient, channelService)

	return NewExecutableSchema(Config{
		Resolvers: &Resolver{
			client:                         client,
			authService:                    authService,
			apiKeyService:                  apiKeyService,
			userService:                    userService,
			systemService:                  systemService,
			channelService:                 channelService,
			requestService:                 requestService,
			channelOverrideTemplateService: channelOverrideTemplateService,
			modelService:                   modelService,
			channelProbeService:            channelProbeService,
			providerQuotaService:           providerQuotaService,
			scheduler:                      scheduler,
			modelFetcher:                   modelFetcher,
			defaultSelector:                defaultSelector,
			candidateSelectorDiagnostics:   candidateSelectorDiagnostics,
			channelLimiterManager:          channelLimiterManager,
			TestChannelOrchestrator:        orchestrator.NewTestChannelOrchestrator(channelService, requestService, systemService, usageLogService, httpClient),
			gcWorker:                       gcWorker,
		},
	})
}
