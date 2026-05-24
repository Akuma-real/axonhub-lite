import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { graphqlRequest } from '@/gql/graphql';
import { toast } from 'sonner';
import i18n from '@/lib/i18n';
import { useErrorHandler } from '@/hooks/use-error-handler';
import type { ProxyConfig } from '@/features/channels/data/schema';
import type { ModelAssociation } from '@/features/models/data/schema';

// GraphQL queries and mutations
const SYSTEM_VERSION_QUERY = `
  query SystemVersion {
    systemVersion {
      version
      commit
      buildTime
      goVersion
      platform
      uptime
    }
  }
`;

export const CHECK_FOR_UPDATE_QUERY = `
  query CheckForUpdate {
    checkForUpdate {
      currentVersion
      latestVersion
      hasUpdate
      releaseUrl
    }
  }
`;

const GET_CACHE_DIAGNOSTICS_QUERY = `
  query GetCacheDiagnostics($input: GetCacheDiagnosticsInput) {
    getCacheDiagnostics(input: $input) {
      fileName
      content
      targets
    }
  }
`;

const CLEAR_CACHE_MUTATION = `
  mutation ClearCache($input: ClearCacheInput!) {
    clearCache(input: $input) {
      success
      message
      targets
    }
  }
`;

const BRAND_SETTINGS_QUERY = `
  query BrandSettings {
    brandSettings {
      brandName
      brandLogo
    }
  }
`;

const STORAGE_POLICY_QUERY = `
  query StoragePolicy {
    storagePolicy {
      storeChunks
      livePreview
      storeRequestBody
      storeResponseBody
      cleanupOptions {
        resourceType
        enabled
        cleanupDays
      }
    }
  }
`;

const UPDATE_BRAND_SETTINGS_MUTATION = `
  mutation UpdateBrandSettings($input: UpdateBrandSettingsInput!) {
    updateBrandSettings(input: $input)
  }
`;

const UPDATE_STORAGE_POLICY_MUTATION = `
  mutation UpdateStoragePolicy($input: UpdateStoragePolicyInput!) {
    updateStoragePolicy(input: $input)
  }
`;

const RETRY_POLICY_QUERY = `
  query RetryPolicy {
    retryPolicy {
      maxChannelRetries
      maxSingleChannelRetries
      retryDelayMs
      loadBalancerStrategy
      enabled
      emptyResponseDetection
      upstreamErrorPolicy {
        mode
        customMessage
      }
      autoDisableChannel {
        enabled
        statuses {
          status
          times
        }
      }
    }
  }
`;

const UPDATE_RETRY_POLICY_MUTATION = `
  mutation UpdateRetryPolicy($input: UpdateRetryPolicyInput!) {
    updateRetryPolicy(input: $input)
  }
`;

const WEBHOOK_NOTIFIER_CONFIG_QUERY = `
  query WebhookNotifierConfig {
    webhookNotifierConfig {
      targets {
        name
        enabled
        url
        proxy {
          type
          url
          username
          password
        }
        timeoutMs
        headers {
          key
          value
        }
        body
      }
      subscriptions {
        event
        targetNames
      }
    }
  }
`;

const UPDATE_WEBHOOK_NOTIFIER_CONFIG_MUTATION = `
  mutation UpdateWebhookNotifierConfig($input: WebhookNotifierConfigInput!) {
    updateWebhookNotifierConfig(input: $input)
  }
`;

const ONBOARDING_INFO_QUERY = `
  query OnboardingInfo {
    onboardingInfo {
      onboarded
      completedAt
      systemModelSetting {
        onboarded
        completedAt
      }
      autoDisableChannel {
        onboarded
        completedAt
      }
    }
  }
`;

const COMPLETE_ONBOARDING_MUTATION = `
  mutation CompleteOnboarding($input: CompleteOnboardingInput!) {
    completeOnboarding(input: $input)
  }
`;

const COMPLETE_SYSTEM_MODEL_SETTING_ONBOARDING_MUTATION = `
  mutation CompleteSystemModelSettingOnboarding($input: CompleteSystemModelSettingOnboardingInput!) {
    completeSystemModelSettingOnboarding(input: $input)
  }
`;

const COMPLETE_AUTO_DISABLE_CHANNEL_ONBOARDING_MUTATION = `
  mutation CompleteAutoDisableChannelOnboarding($input: CompleteAutoDisableChannelOnboardingInput!) {
    completeAutoDisableChannelOnboarding(input: $input)
  }
`;

const TRIGGER_GC_CLEANUP_MUTATION = `
  mutation triggerGcCleanup($input: TriggerGcCleanupInput!) {
    triggerGcCleanup(input: $input)
  }
`;

const PREVIEW_GC_CLEANUP_QUERY = `
  query previewGcCleanup($input: TriggerGcCleanupInput!) {
    previewGcCleanup(input: $input) {
      resourceType
      estimatedCount
      cutoffTime
      retentionDays
    }
  }
`;

// Types
export interface BrandSettings {
  brandName?: string;
  brandLogo?: string;
}

export interface SystemGeneralSettings {
  currencyCode: string;
  timezone: string;
}

export interface UpdateSystemGeneralSettingsInput {
  currencyCode?: string;
  timezone?: string;
}

export interface StoragePolicy {
  storeChunks: boolean;
  livePreview: boolean;
  storeRequestBody: boolean;
  storeResponseBody: boolean;
  cleanupOptions: CleanupOption[];
}

export interface CleanupOption {
  resourceType: string;
  enabled: boolean;
  cleanupDays: number;
}

export interface UpdateBrandSettingsInput {
  brandName?: string;
  brandLogo?: string;
}

export interface UpdateStoragePolicyInput {
  storeChunks?: boolean;
  livePreview?: boolean;
  storeRequestBody?: boolean;
  storeResponseBody?: boolean;
  cleanupOptions?: CleanupOptionInput[];
}

export interface CleanupOptionInput {
  resourceType: string;
  enabled: boolean;
  cleanupDays: number;
}

export interface TriggerGcCleanupInput {
  requestsCleanupDays: number;
  usageLogsCleanupDays: number;
}

export interface GcCleanupPreviewItem {
  resourceType: string;
  estimatedCount: number;
  cutoffTime: string;
  retentionDays: number;
}

export interface AutoDisableChannelStatus {
  status: number;
  times: number;
}

export interface WebhookHeader {
  key: string;
  value: string;
}

export interface WebhookTarget {
  name: string;
  enabled: boolean;
  url: string;
  proxy?: ProxyConfig | null;
  timeoutMs: number;
  headers: WebhookHeader[];
  body: string;
}

export interface WebhookSubscription {
  event: string;
  targetNames: string[];
}

export interface WebhookNotifierConfig {
  targets: WebhookTarget[];
  subscriptions: WebhookSubscription[];
}

export interface AutoDisableChannel {
  enabled: boolean;
  statuses: AutoDisableChannelStatus[];
}

export interface RetryPolicy {
  maxChannelRetries: number;
  maxSingleChannelRetries: number;
  retryDelayMs: number;
  loadBalancerStrategy: string;
  enabled: boolean;
  autoDisableChannel: AutoDisableChannel;
  emptyResponseDetection: boolean;
  upstreamErrorPolicy: UpstreamErrorPolicy;
}

export interface UpstreamErrorPolicy {
  mode: string;
  customMessage: string;
}

export interface AutoDisableChannelStatusInput {
  status: number;
  times: number;
}

export interface AutoDisableChannelInput {
  enabled?: boolean;
  statuses?: AutoDisableChannelStatusInput[];
}

export interface RetryPolicyInput {
  maxChannelRetries?: number;
  maxSingleChannelRetries?: number;
  retryDelayMs?: number;
  loadBalancerStrategy?: string;
  enabled?: boolean;
  autoDisableChannel?: AutoDisableChannelInput;
  emptyResponseDetection?: boolean;
  upstreamErrorPolicy?: Partial<UpstreamErrorPolicy>;
}

export interface SystemModelSettingOnboarding {
  onboarded: boolean;
  completedAt?: string;
}

export interface AutoDisableChannelOnboarding {
  onboarded: boolean;
  completedAt?: string;
}

export interface OnboardingInfo {
  onboarded: boolean;
  completedAt?: string;
  systemModelSetting?: SystemModelSettingOnboarding;
  autoDisableChannel?: AutoDisableChannelOnboarding;
}

export interface CompleteOnboardingInput {
  dummy?: string;
}

export interface CompleteSystemModelSettingOnboardingInput {
  dummy?: string;
}

export interface CompleteAutoDisableChannelOnboardingInput {
  dummy?: string;
}

export interface SystemVersion {
  version: string;
  commit: string;
  buildTime: string;
  goVersion: string;
  platform: string;
  uptime: string;
}

export interface VersionCheck {
  currentVersion: string;
  latestVersion: string;
  hasUpdate: boolean;
  releaseUrl: string;
}

export type DiagnosticsTarget = 'CHANNEL_CACHE';

export interface GetCacheDiagnosticsPayload {
  fileName: string;
  content: string;
  targets: DiagnosticsTarget[];
}

export interface ClearCachePayload {
  success: boolean;
  message: string;
  targets: DiagnosticsTarget[];
}

// Hooks
export function useBrandSettings() {
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['brandSettings'],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ brandSettings: BrandSettings }>(BRAND_SETTINGS_QUERY);
        return data.brandSettings;
      } catch (error) {
        handleError(error, i18n.t('common.errors.internalServerError'));
        throw error;
      }
    },
  });
}

export function useStoragePolicy() {
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['storagePolicy'],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ storagePolicy: StoragePolicy }>(STORAGE_POLICY_QUERY);
        return data.storagePolicy;
      } catch (error) {
        handleError(error, i18n.t('common.errors.internalServerError'));
        throw error;
      }
    },
  });
}

export function useUpdateBrandSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdateBrandSettingsInput) => {
      const data = await graphqlRequest<{ updateBrandSettings: boolean }>(UPDATE_BRAND_SETTINGS_MUTATION, { input });
      return data.updateBrandSettings;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['brandSettings'] });
      toast.success(i18n.t('common.success.systemUpdated'));
    },
    onError: () => {
      toast.error(i18n.t('common.errors.systemUpdateFailed'));
    },
  });
}

export function useUpdateStoragePolicy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdateStoragePolicyInput) => {
      const data = await graphqlRequest<{ updateStoragePolicy: boolean }>(UPDATE_STORAGE_POLICY_MUTATION, { input });
      return data.updateStoragePolicy;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['storagePolicy'] });
      toast.success(i18n.t('common.success.systemUpdated'));
    },
    onError: () => {
      toast.error(i18n.t('common.errors.systemUpdateFailed'));
    },
  });
}

export function useTriggerGcCleanup() {
  return useMutation({
    mutationFn: async (input: TriggerGcCleanupInput) => {
      const data = await graphqlRequest<{ triggerGcCleanup: boolean }>(TRIGGER_GC_CLEANUP_MUTATION, { input });
      return data.triggerGcCleanup;
    },
    onSuccess: () => {
      toast.success(i18n.t('system.storage.policy.runCleanupSuccess'));
    },
    onError: () => {
      toast.error(i18n.t('system.storage.policy.runCleanupError'));
    },
  });
}

export function usePreviewGcCleanup() {
  return useMutation({
    mutationFn: async (input: TriggerGcCleanupInput) => {
      const data = await graphqlRequest<{ previewGcCleanup: GcCleanupPreviewItem[] }>(PREVIEW_GC_CLEANUP_QUERY, { input });
      return data.previewGcCleanup;
    },
  });
}

export function useRetryPolicy() {
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['retryPolicy'],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ retryPolicy: RetryPolicy }>(RETRY_POLICY_QUERY);
        return data.retryPolicy;
      } catch (error) {
        handleError(error, i18n.t('common.errors.internalServerError'));
        throw error;
      }
    },
  });
}

export function useUpdateRetryPolicy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: RetryPolicyInput) => {
      const data = await graphqlRequest<{ updateRetryPolicy: boolean }>(UPDATE_RETRY_POLICY_MUTATION, { input });
      return data.updateRetryPolicy;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['retryPolicy'] });
      toast.success(i18n.t('common.success.systemUpdated'));
    },
    onError: () => {
      toast.error(i18n.t('common.errors.systemUpdateFailed'));
    },
  });
}

export function useWebhookNotifierConfig() {
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['webhookNotifierConfig'],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ webhookNotifierConfig: WebhookNotifierConfig }>(WEBHOOK_NOTIFIER_CONFIG_QUERY);
        return data.webhookNotifierConfig;
      } catch (error) {
        handleError(error, i18n.t('common.errors.internalServerError'));
        throw error;
      }
    },
  });
}

export function useUpdateWebhookNotifierConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: WebhookNotifierConfig) => {
      const data = await graphqlRequest<{ updateWebhookNotifierConfig: boolean }>(UPDATE_WEBHOOK_NOTIFIER_CONFIG_MUTATION, { input });
      return data.updateWebhookNotifierConfig;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhookNotifierConfig'] });
      toast.success(i18n.t('common.success.systemUpdated'));
    },
    onError: () => {
      toast.error(i18n.t('common.errors.systemUpdateFailed'));
    },
  });
}

export function useOnboardingInfo() {
  return useQuery({
    queryKey: ['onboardingInfo'],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ onboardingInfo: OnboardingInfo | null }>(ONBOARDING_INFO_QUERY);
        return data.onboardingInfo;
      } catch (_error) {
        return {
          onboarded: true,
          completedAt: new Date().toISOString(),
        };
      }
    },
  });
}

export function useCompleteOnboarding() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input?: CompleteOnboardingInput) => {
      const data = await graphqlRequest<{ completeOnboarding: boolean }>(COMPLETE_ONBOARDING_MUTATION, { input: input || {} });
      return data.completeOnboarding;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['onboardingInfo'] });
    },
    onError: () => {
      toast.error(i18n.t('common.errors.onboardingFailed'));
    },
  });
}

export function useCompleteSystemModelSettingOnboarding() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input?: CompleteSystemModelSettingOnboardingInput) => {
      const data = await graphqlRequest<{ completeSystemModelSettingOnboarding: boolean }>(
        COMPLETE_SYSTEM_MODEL_SETTING_ONBOARDING_MUTATION,
        { input: input || {} }
      );
      return data.completeSystemModelSettingOnboarding;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['onboardingInfo'] });
    },
    onError: () => {
      toast.error(i18n.t('common.errors.onboardingFailed'));
    },
  });
}

export function useCompleteAutoDisableChannelOnboarding() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input?: CompleteAutoDisableChannelOnboardingInput) => {
      const data = await graphqlRequest<{ completeAutoDisableChannelOnboarding: boolean }>(
        COMPLETE_AUTO_DISABLE_CHANNEL_ONBOARDING_MUTATION,
        { input: input || {} }
      );
      return data.completeAutoDisableChannelOnboarding;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['onboardingInfo'] });
    },
    onError: () => {
      toast.error(i18n.t('common.errors.onboardingFailed'));
    },
  });
}

export function useSystemVersion() {
  return useQuery({
    queryKey: ['systemVersion'],
    queryFn: async () => {
      const data = await graphqlRequest<{ systemVersion: SystemVersion }>(SYSTEM_VERSION_QUERY);
      return data.systemVersion;
    },
  });
}

export function useCheckForUpdate() {
  return useQuery({
    queryKey: ['checkForUpdate'],
    queryFn: async () => {
      const data = await graphqlRequest<{ checkForUpdate: VersionCheck }>(CHECK_FOR_UPDATE_QUERY);
      return data.checkForUpdate;
    },
    retry: false,
    staleTime: 60 * 60 * 1000, // 1 hour
  });
}

export function useExportCacheDiagnostics() {
  const { handleError } = useErrorHandler();

  return useMutation({
    mutationFn: async () => {
      const data = await graphqlRequest<{ getCacheDiagnostics: GetCacheDiagnosticsPayload }>(
        GET_CACHE_DIAGNOSTICS_QUERY,
        { input: { targets: ['CHANNEL_CACHE'] } }
      );
      return data.getCacheDiagnostics;
    },
    onSuccess: (data) => {
      const blob = new Blob([data.content], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = data.fileName;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      toast.success(i18n.t('system.diagnostics.cache.exportSuccess'));
    },
    onError: (error) => {
      handleError(error, i18n.t('system.diagnostics.cache.exportFailed'));
    },
  });
}

export function useClearCache() {
  const { handleError } = useErrorHandler();

  return useMutation({
    mutationFn: async () => {
      const data = await graphqlRequest<{ clearCache: ClearCachePayload }>(CLEAR_CACHE_MUTATION, {
        input: { targets: ['CHANNEL_CACHE'] },
      });
      return data.clearCache;
    },
    onSuccess: (data) => {
      if (data.success) {
        toast.success(i18n.t('system.diagnostics.cache.clearSuccess'));
        return;
      }

      toast.error(data.message || i18n.t('system.diagnostics.cache.clearFailed'));
    },
    onError: (error) => {
      handleError(error, i18n.t('system.diagnostics.cache.clearFailed'));
    },
  });
}

// Model Settings
const MODEL_SETTINGS_QUERY = `
  query ModelSettings {
    systemModelSettings {
      fallbackToChannelsOnModelNotFound
      queryAllChannelModels
      defaultModelAPIIncludeAll
      autoReasoningEffort
      modelBlacklistRegex
      developerSettings {
        developer
        associations {
          type
          priority
          disabled
          when {
            enabled
            condition {
              type
              logic
              field
              operator
              value
              conditions {
                type
                logic
                field
                operator
                value
                conditions {
                  type
                  logic
                  field
                  operator
                  value
                }
              }
            }
          }
          channelModel {
            channelId
            modelId
          }
          channelRegex {
            channelId
            pattern
          }
          regex {
            pattern
            exclude {
              channelNamePattern
              channelIds
              channelTags
            }
          }
          modelId {
            modelId
            exclude {
              channelNamePattern
              channelIds
              channelTags
            }
          }
          channelTagsModel {
            channelTags
            modelId
          }
          channelTagsRegex {
            channelTags
            pattern
          }
        }
      }
    }
  }
`;

const UPDATE_MODEL_SETTINGS_MUTATION = `
  mutation UpdateModelSettings($input: UpdateSystemModelSettingsInput!) {
    updateSystemModelSettings(input: $input)
  }
`;

const CHANNEL_SETTINGS_QUERY = `
  query SystemChannelSettings {
    systemChannelSettings {
      probe {
        enabled
        frequency
      }
      autoSync {
        frequency
      }
    }
  }
`;

const UPDATE_CHANNEL_SETTINGS_MUTATION = `
  mutation UpdateChannelSettings($input: UpdateSystemChannelSettingsInput!) {
    updateSystemChannelSettings(input: $input)
  }
`;

const SYSTEM_GENERAL_SETTINGS_QUERY = `
  query SystemGeneralSettings {
    systemGeneralSettings {
      currencyCode
      timezone
    }
  }
`;

const UPDATE_SYSTEM_GENERAL_SETTINGS_MUTATION = `
  mutation UpdateSystemGeneralSettings($input: UpdateSystemGeneralSettingsInput!) {
    updateSystemGeneralSettings(input: $input)
  }
`;

export interface ModelSettings {
  fallbackToChannelsOnModelNotFound: boolean;
  queryAllChannelModels: boolean;
  defaultModelAPIIncludeAll: boolean;
  autoReasoningEffort: boolean;
  modelBlacklistRegex: string;
  developerSettings: DeveloperModelSettings[];
}

export interface UpdateModelSettingsInput {
  fallbackToChannelsOnModelNotFound?: boolean;
  queryAllChannelModels?: boolean;
  defaultModelAPIIncludeAll?: boolean;
  autoReasoningEffort?: boolean;
  modelBlacklistRegex?: string;
  developerSettings?: DeveloperModelSettings[];
}

export interface DeveloperModelSettings {
  developer: string;
  associations: ModelAssociation[];
}

export function useModelSettings() {
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['modelSettings'],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ systemModelSettings: ModelSettings }>(MODEL_SETTINGS_QUERY);
        return data.systemModelSettings;
      } catch (error) {
        handleError(error, i18n.t('common.errors.internalServerError'));
        throw error;
      }
    },
  });
}

export function useUpdateModelSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdateModelSettingsInput) => {
      const data = await graphqlRequest<{ updateSystemModelSettings: boolean }>(UPDATE_MODEL_SETTINGS_MUTATION, { input });
      return data.updateSystemModelSettings;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['modelSettings'] });
      queryClient.invalidateQueries({ queryKey: ['models'] });
      toast.success(i18n.t('common.success.systemUpdated'));
    },
    onError: () => {
      toast.error(i18n.t('common.errors.systemUpdateFailed'));
    },
  });
}

export type ProbeFrequency = 'ONE_MINUTE' | 'FIVE_MINUTES' | 'THIRTY_MINUTES' | 'ONE_HOUR';

export type AutoSyncFrequency = 'ONE_HOUR' | 'SIX_HOURS' | 'ONE_DAY';

export interface ChannelProbeSetting {
  enabled: boolean;
  frequency: ProbeFrequency;
}

export interface ChannelModelAutoSyncSetting {
  frequency: AutoSyncFrequency;
}

export interface ChannelSetting {
  probe: ChannelProbeSetting;
  autoSync: ChannelModelAutoSyncSetting;
}

export interface UpdateChannelProbeSettingInput {
  enabled?: boolean;
  frequency?: ProbeFrequency;
}

export interface UpdateChannelModelAutoSyncSettingInput {
  frequency?: AutoSyncFrequency;
}

export interface UpdateSystemChannelSettingsInput {
  probe?: UpdateChannelProbeSettingInput;
  autoSync?: UpdateChannelModelAutoSyncSettingInput;
}

export function useChannelSetting() {
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['channelSetting'],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ systemChannelSettings: ChannelSetting }>(CHANNEL_SETTINGS_QUERY);
        return data.systemChannelSettings;
      } catch (error) {
        handleError(error, i18n.t('common.errors.internalServerError'));
        throw error;
      }
    },
  });
}

export function useUpdateChannelSetting() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdateSystemChannelSettingsInput) => {
      const data = await graphqlRequest<{ updateSystemChannelSettings: boolean }>(UPDATE_CHANNEL_SETTINGS_MUTATION, { input });
      return data.updateSystemChannelSettings;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['channelSetting'] });
      queryClient.invalidateQueries({ queryKey: ['channelProbeData'] });
      toast.success(i18n.t('common.success.systemUpdated'));
    },
    onError: () => {
      toast.error(i18n.t('common.errors.systemUpdateFailed'));
    },
  });
}

export function useGeneralSettings() {
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['generalSettings'],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ systemGeneralSettings: SystemGeneralSettings }>(SYSTEM_GENERAL_SETTINGS_QUERY);
        return data.systemGeneralSettings;
      } catch (error) {
        handleError(error, i18n.t('common.errors.internalServerError'));
        throw error;
      }
    },
    placeholderData: (previousData) => previousData,
  });
}

export function useUpdateGeneralSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdateSystemGeneralSettingsInput) => {
      const data = await graphqlRequest<{ updateSystemGeneralSettings: boolean }>(UPDATE_SYSTEM_GENERAL_SETTINGS_MUTATION, { input });
      return data.updateSystemGeneralSettings;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['generalSettings'] });
      toast.success(i18n.t('common.success.systemUpdated'));
    },
    onError: () => {
      toast.error(i18n.t('common.errors.systemUpdateFailed'));
    },
  });
}

// Proxy Presets
const PROXY_PRESETS_QUERY = `
  query ProxyPresets {
    proxyPresets {
      name
      url
      username
      password
    }
  }
`;

const SAVE_PROXY_PRESET_MUTATION = `
  mutation SaveProxyPreset($input: SaveProxyPresetInput!) {
    saveProxyPreset(input: $input)
  }
`;

const DELETE_PROXY_PRESET_MUTATION = `
  mutation DeleteProxyPreset($url: String!) {
    deleteProxyPreset(url: $url)
  }
`;

export interface ProxyPreset {
  name?: string;
  url: string;
  username?: string;
  password?: string;
}

export interface SaveProxyPresetInput {
  name?: string;
  url: string;
  username?: string;
  password?: string;
}

export function useProxyPresets() {
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['proxyPresets'],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ proxyPresets: ProxyPreset[] }>(PROXY_PRESETS_QUERY);
        return data.proxyPresets;
      } catch (error) {
        handleError(error, i18n.t('common.errors.internalServerError'));
        throw error;
      }
    },
  });
}

export function useSaveProxyPreset() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: SaveProxyPresetInput) => {
      const data = await graphqlRequest<{ saveProxyPreset: boolean }>(SAVE_PROXY_PRESET_MUTATION, { input });
      return data.saveProxyPreset;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['proxyPresets'] });
    },
    onError: () => {
      toast.error(i18n.t('common.errors.systemUpdateFailed'));
    },
  });
}

export function useDeleteProxyPreset() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (url: string) => {
      const data = await graphqlRequest<{ deleteProxyPreset: boolean }>(DELETE_PROXY_PRESET_MUTATION, { url });
      return data.deleteProxyPreset;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['proxyPresets'] });
      toast.success(i18n.t('common.success.systemUpdated'));
    },
    onError: () => {
      toast.error(i18n.t('common.errors.systemUpdateFailed'));
    },
  });
}


// User-Agent Pass-Through Settings
const USER_AGENT_PASS_THROUGH_SETTINGS_QUERY = `
  query UserAgentPassThroughSettings {
    userAgentPassThroughSettings {
      enabled
    }
  }
`;

const UPDATE_USER_AGENT_PASS_THROUGH_SETTINGS_MUTATION = `
  mutation UpdateUserAgentPassThroughSettings($input: UpdateUserAgentPassThroughSettingsInput!) {
    updateUserAgentPassThroughSettings(input: $input)
  }
`;

export interface UserAgentPassThroughSettings {
  enabled: boolean;
}

export interface UpdateUserAgentPassThroughSettingsInput {
  enabled: boolean;
}

export function useUserAgentPassThroughSettings() {
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['userAgentPassThroughSettings'],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ userAgentPassThroughSettings: UserAgentPassThroughSettings }>(USER_AGENT_PASS_THROUGH_SETTINGS_QUERY);
        return data.userAgentPassThroughSettings;
      } catch (error) {
        handleError(error, i18n.t('common.errors.internalServerError'));
        throw error;
      }
    },
  });
}

export function useUpdateUserAgentPassThroughSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdateUserAgentPassThroughSettingsInput) => {
      const data = await graphqlRequest<{ updateUserAgentPassThroughSettings: boolean }>(UPDATE_USER_AGENT_PASS_THROUGH_SETTINGS_MUTATION, { input });
      return data.updateUserAgentPassThroughSettings;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['userAgentPassThroughSettings'] });
      toast.success(i18n.t('common.success.systemUpdated'));
    },
    onError: () => {
      toast.error(i18n.t('common.errors.systemUpdateFailed'));
    },
  });
}

// Pass-Through Settings (request/response body pass-through)
const PASS_THROUGH_SETTINGS_QUERY = `
  query PassThroughSettings {
    passThroughSettings {
      enabled
    }
  }
`;

const UPDATE_PASS_THROUGH_SETTINGS_MUTATION = `
  mutation UpdatePassThroughSettings($input: UpdatePassThroughSettingsInput!) {
    updatePassThroughSettings(input: $input)
  }
`;

export interface PassThroughSettings {
  enabled: boolean;
}

export interface UpdatePassThroughSettingsInput {
  enabled: boolean;
}

export function usePassThroughSettings() {
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['passThroughSettings'],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ passThroughSettings: PassThroughSettings }>(PASS_THROUGH_SETTINGS_QUERY);
        return data.passThroughSettings;
      } catch (error) {
        handleError(error, i18n.t('common.errors.internalServerError'));
        throw error;
      }
    },
  });
}

export function useUpdatePassThroughSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdatePassThroughSettingsInput) => {
      const data = await graphqlRequest<{ updatePassThroughSettings: boolean }>(UPDATE_PASS_THROUGH_SETTINGS_MUTATION, { input });
      return data.updatePassThroughSettings;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['passThroughSettings'] });
      toast.success(i18n.t('common.success.systemUpdated'));
    },
    onError: () => {
      toast.error(i18n.t('common.errors.systemUpdateFailed'));
    },
  });
}

const QUOTA_ENFORCEMENT_SETTINGS_QUERY = `
  query QuotaEnforcementSettings {
    quotaEnforcementSettings {
      enabled
      mode
    }
  }
`;

const UPDATE_QUOTA_ENFORCEMENT_SETTINGS_MUTATION = `
  mutation UpdateQuotaEnforcementSettings($input: UpdateQuotaEnforcementSettingsInput!) {
    updateQuotaEnforcementSettings(input: $input)
  }
`;

export type QuotaEnforcementMode = 'EXHAUSTED_ONLY' | 'DE_PRIORITIZE';

export interface QuotaEnforcementSettings {
  enabled: boolean;
  mode: QuotaEnforcementMode;
}

export interface UpdateQuotaEnforcementSettingsInput {
  enabled?: boolean;
  mode?: QuotaEnforcementMode;
}

export function useQuotaEnforcementSettings() {
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['quotaEnforcementSettings'],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ quotaEnforcementSettings: QuotaEnforcementSettings }>(QUOTA_ENFORCEMENT_SETTINGS_QUERY);
        return data.quotaEnforcementSettings;
      } catch (error) {
        handleError(error, i18n.t('common.errors.internalServerError'));
        throw error;
      }
    },
  });
}

export function useUpdateQuotaEnforcementSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdateQuotaEnforcementSettingsInput) => {
      const data = await graphqlRequest<{ updateQuotaEnforcementSettings: boolean }>(UPDATE_QUOTA_ENFORCEMENT_SETTINGS_MUTATION, { input });
      return data.updateQuotaEnforcementSettings;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['quotaEnforcementSettings'] });
      toast.success(i18n.t('common.success.systemUpdated'));
    },
    onError: () => {
      toast.error(i18n.t('common.errors.systemUpdateFailed'));
    },
  });
}
