import { useMutation, useQuery, useQueryClient, keepPreviousData } from '@tanstack/react-query';
import { graphqlRequest } from '@/gql/graphql';

import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { useErrorHandler } from '@/hooks/use-error-handler';
import type {
  ApiKey,
  ApiKeyConnection,
  ApiKeyProfileQuotaUsage,
  ApiKeyTokenUsageStats,
  CreateApiKeyInput,
  UpdateApiKeyInput,
  UpdateApiKeyProfilesInput,
} from './schema';
import { apiKeyConnectionSchema, apiKeyProfileQuotaUsageSchema, apiKeySchema, apiKeyTokenUsageStatsSchema } from './schema';

const APIKEYS_QUERY = `
    query GetApiKeys($first: Int, $after: Cursor, $orderBy: APIKeyOrder, $where: APIKeyWhereInput) {
      apiKeys(first: $first, after: $after, orderBy: $orderBy, where: $where) {
        edges {
          node {
            id
            createdAt
            updatedAt
            key
            name
            status
          }
          cursor
        }
        pageInfo {
          hasNextPage
          hasPreviousPage
          startCursor
          endCursor
        }
        totalCount
      }
    }
  `;

const APIKEY_QUERY = `
    query GetApiKey($id: ID!) {
      node(id: $id) {
        ... on APIKey {
        id
        createdAt
        updatedAt
        key
        name
        status
        profiles {
          activeProfile
          profiles {
            name
            modelMappings { from to }
            channelIDs
            channelTags
            channelTagsMatchMode
            modelIDs
            loadBalanceStrategy
            quota {
              requests
              totalTokens
              cost
              period {
                type
                pastDuration { value unit }
                calendarDuration { unit }
              }
            }
          }
        }
      }
    }
}
  `;

const CREATE_APIKEY_MUTATION = `
    mutation CreateAPIKey($input: CreateAPIKeyInput!) {
      createAPIKey(input: $input) {
        id
        createdAt
        updatedAt
        key
        name
        status
      }
    }
  `;

const UPDATE_APIKEY_MUTATION = `
    mutation UpdateAPIKey($id: ID!, $input: UpdateAPIKeyInput!) {
      updateAPIKey(id: $id, input: $input) {
        id
        createdAt
        updatedAt
        key
        name
        status
      }
    }
  `;

const UPDATE_APIKEY_STATUS_MUTATION = `
  mutation UpdateAPIKeyStatus($id: ID!, $status: APIKeyStatus!) {
    updateAPIKeyStatus(id: $id, status: $status) {
      id
      status
    }
  }
`;

const UPDATE_APIKEY_PROFILES_MUTATION = `
  mutation UpdateAPIKeyProfiles($id: ID!, $input: UpdateAPIKeyProfilesInput!) {
    updateAPIKeyProfiles(id: $id, input: $input) {
      id
      name
      status
      profiles {
        activeProfile
        profiles {
          name
          modelMappings {
            from
            to
          }
          channelIDs
          channelTags
          channelTagsMatchMode
          modelIDs
          loadBalanceStrategy
          quota {
            requests
            totalTokens
            cost
            period {
              type
              pastDuration { value unit }
              calendarDuration { unit }
            }
          }
        }
      }
    }
  }
`;

const BULK_DISABLE_APIKEYS_MUTATION = `
  mutation BulkDisableAPIKeys($ids: [ID!]!) {
    bulkDisableAPIKeys(ids: $ids)
  }
`;

const BULK_ENABLE_APIKEYS_MUTATION = `
  mutation BulkEnableAPIKeys($ids: [ID!]!) {
    bulkEnableAPIKeys(ids: $ids)
  }
`;

const BULK_ARCHIVE_APIKEYS_MUTATION = `
  mutation BulkArchiveAPIKeys($ids: [ID!]!) {
    bulkArchiveAPIKeys(ids: $ids)
  }
`;

const ROTATE_APIKEY_MUTATION = `
  mutation RotateAPIKey($id: ID!) {
    rotateAPIKey(id: $id) {
      id
      key
      name
      status
      createdAt
      updatedAt
    }
  }
`;

const APIKEY_QUOTA_USAGES_QUERY = `
  query APIKeyQuotaUsages($apiKeyId: ID!) {
    apiKeyQuotaUsages(apiKeyId: $apiKeyId) {
      profileName
      quota {
        requests
        totalTokens
        cost
        period {
          type
          pastDuration { value unit }
          calendarDuration { unit }
        }
      }
      window { start end }
      usage { requestCount totalTokens totalCost }
    }
  }
`;

const APIKEY_TOKEN_USAGE_STATS_QUERY = `
  query APIKeyTokenUsageStats($input: APIKeyTokenUsageStatsInput) {
    apiKeyTokenUsageStats(input: $input) {
      apiKeyId
      inputTokens
      outputTokens
      cachedTokens
      reasoningTokens
      topModels {
        modelId
        inputTokens
        outputTokens
        cachedTokens
        reasoningTokens
      }
    }
  }
`;

// React Query hooks
export function useApiKeys(
  variables?: {
    first?: number;
    after?: string;
    orderBy?: { field: 'CREATED_AT'; direction: 'ASC' | 'DESC' };
    where?: {
      nameContainsFold?: string;
      status?: string;
      [key: string]: unknown;
    };
  },
  options?: {
    disableAutoFetch?: boolean;
  }
) {
  const { t } = useTranslation();
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['apiKeys', variables],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ apiKeys: ApiKeyConnection }>(APIKEYS_QUERY, variables);
        return apiKeyConnectionSchema.parse(data?.apiKeys);
      } catch (error) {
        handleError(error, t('common.errors.internalServerError'));
        throw error;
      }
    },
    enabled: !options?.disableAutoFetch,
  });
}

export function useApiKey(id: string) {
  const { t } = useTranslation();
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['apiKey', id],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ node: ApiKey }>(APIKEY_QUERY, { id });
        return apiKeySchema.parse(data.node);
      } catch (error) {
        handleError(error, t('common.errors.internalServerError'));
        throw error;
      }
    },
    enabled: !!id,
  });
}

export function useApiKeyQuotaUsages(
  apiKeyId: string,
  options?: {
    enabled?: boolean;
    refetchInterval?: number;
  }
) {
  const { t } = useTranslation();
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['apiKeyQuotaUsages', apiKeyId],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ apiKeyQuotaUsages: ApiKeyProfileQuotaUsage[] }>(
          APIKEY_QUOTA_USAGES_QUERY,
          { apiKeyId }
        );
        return apiKeyProfileQuotaUsageSchema.array().parse(data.apiKeyQuotaUsages);
      } catch (error) {
        handleError(error, t('common.errors.internalServerError'));
        throw error;
      }
    },
    enabled: !!apiKeyId && (options?.enabled ?? true),
    refetchInterval: options?.refetchInterval,
  });
}

export function useApiKeyTokenUsageStats(
  variables?: {
    apiKeyIds?: string[];
    createdAtGTE?: string;
    createdAtLTE?: string;
  },
  options?: {
    enabled?: boolean;
  }
) {
  const { t } = useTranslation();
  const { handleError } = useErrorHandler();

  return useQuery({
    queryKey: ['apiKeyTokenUsageStats', variables],
    queryFn: async () => {
      try {
        const data = await graphqlRequest<{ apiKeyTokenUsageStats: ApiKeyTokenUsageStats[] }>(
          APIKEY_TOKEN_USAGE_STATS_QUERY,
          { input: variables && Object.keys(variables).length > 0 ? variables : undefined }
        );
        return apiKeyTokenUsageStatsSchema.array().parse(data.apiKeyTokenUsageStats);
      } catch (error) {
        handleError(error, t('common.errors.internalServerError'));
        throw error;
      }
    },
    enabled: options?.enabled ?? true,
    placeholderData: keepPreviousData,
    staleTime: 30000, // Consider data fresh for 30 seconds
  });
}

export function useCreateApiKey() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { handleError } = useErrorHandler();

  return useMutation({
    mutationFn: (input: CreateApiKeyInput) => {
      return graphqlRequest<{ createAPIKey: ApiKey }>(CREATE_APIKEY_MUTATION, { input });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
      toast.success(t('apikeys.messages.createSuccess'));
    },
    onError: (error) => {
      handleError(error, { context: t('apikeys.dialogs.create.title') });
    },
  });
}

export function useUpdateApiKey() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { handleError } = useErrorHandler();

  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateApiKeyInput }) => {
      return graphqlRequest<{ updateAPIKey: ApiKey }>(UPDATE_APIKEY_MUTATION, { id, input });
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
      queryClient.invalidateQueries({ queryKey: ['apiKey', variables.id] });
      toast.success(t('apikeys.messages.updateSuccess'));
    },
    onError: (error) => {
      handleError(error, { context: t('apikeys.dialogs.edit.title') });
    },
  });
}

export function useUpdateApiKeyStatus() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: 'enabled' | 'disabled' | 'archived' }) => {
      return graphqlRequest<{ updateAPIKeyStatus: ApiKey }>(UPDATE_APIKEY_STATUS_MUTATION, { id, status });
    },
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
      queryClient.invalidateQueries({ queryKey: ['apiKey', variables.id] });
      const statusText =
        data.updateAPIKeyStatus.status === 'enabled'
          ? t('apikeys.status.enabled')
          : data.updateAPIKeyStatus.status === 'disabled'
            ? t('apikeys.status.disabled')
            : t('apikeys.status.archived');
      toast.success(t('apikeys.messages.statusUpdateSuccess', { status: statusText }));
    },
    onError: () => {
      toast.error(t('common.errors.internalServerError'));
    },
  });
}

export function useUpdateApiKeyProfiles() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateApiKeyProfilesInput }) => {
      return graphqlRequest<{ updateAPIKeyProfiles: ApiKey }>(UPDATE_APIKEY_PROFILES_MUTATION, { id, input });
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
      queryClient.invalidateQueries({ queryKey: ['apiKey', variables.id] });
      toast.success(t('apikeys.messages.profilesUpdateSuccess'));
    },
    onError: () => {
      toast.error(t('common.errors.internalServerError'));
    },
  });
}

export function useBulkDisableApiKeys() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (ids: string[]) => {
      const data = await graphqlRequest<{ bulkDisableAPIKeys: boolean }>(BULK_DISABLE_APIKEYS_MUTATION, { ids });
      return data.bulkDisableAPIKeys;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
      toast.success(t('apikeys.messages.bulkDisableSuccess', { count: variables.length }));
    },
    onError: () => {
      toast.error(t('common.errors.internalServerError'));
    },
  });
}

export function useBulkEnableApiKeys() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (ids: string[]) => {
      const data = await graphqlRequest<{ bulkEnableAPIKeys: boolean }>(BULK_ENABLE_APIKEYS_MUTATION, { ids });
      return data.bulkEnableAPIKeys;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
      toast.success(t('apikeys.messages.bulkEnableSuccess', { count: variables.length }));
    },
    onError: () => {
      toast.error(t('common.errors.internalServerError'));
    },
  });
}

export function useBulkArchiveApiKeys() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (ids: string[]) => {
      const data = await graphqlRequest<{ bulkArchiveAPIKeys: boolean }>(BULK_ARCHIVE_APIKEYS_MUTATION, { ids });
      return data.bulkArchiveAPIKeys;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
      toast.success(t('apikeys.messages.bulkArchiveSuccess', { count: variables.length }));
    },
    onError: () => {
      toast.error(t('common.errors.internalServerError'));
    },
  });
}

export function useRotateApiKey() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { handleError } = useErrorHandler();

  return useMutation({
    mutationFn: async (id: string) => {
      const data = await graphqlRequest<{ rotateAPIKey: ApiKey }>(ROTATE_APIKEY_MUTATION, { id });
      return apiKeySchema.parse(data.rotateAPIKey);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
      queryClient.invalidateQueries({ queryKey: ['apiKey', variables] });
      toast.success(t('apikeys.messages.rotateSuccess'));
    },
    onError: (error) => {
      handleError(error, { context: t('apikeys.dialogs.rotate.title') });
    },
  });
}
