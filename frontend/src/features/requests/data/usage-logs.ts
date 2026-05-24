import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { graphqlRequest } from '@/gql/graphql';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { useUsageLogPermissions } from '../../../gql/useUsageLogPermissions';
import { UsageLog, UsageLogConnection, usageLogConnectionSchema, usageLogSchema } from './usage-logs-schema';

// Dynamic GraphQL query builder
function buildUsageLogsQuery(permissions: { canViewChannels: boolean }) {
  const channelFields = permissions.canViewChannels
    ? `
          channel {
            id
            name
            type
          }`
    : '';

  return `
    query GetUsageLogs($first: Int, $after: Cursor, $orderBy: UsageLogOrder, $where: UsageLogWhereInput) {
      usageLogs(first: $first, after: $after, orderBy: $orderBy, where: $where) {
        edges {
          node {
            id
            createdAt
            updatedAt
            requestID${channelFields}
            modelID
            promptTokens
            completionTokens
            totalTokens
            promptAudioTokens
            promptCachedTokens
            promptWriteCachedTokens
            completionAudioTokens
            completionReasoningTokens
            completionAcceptedPredictionTokens
            completionRejectedPredictionTokens
            source
            format
            totalCost
            costItems {
              itemCode
              quantity
              subtotal
            }
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
}

function buildUsageLogDetailQuery(permissions: { canViewChannels: boolean }) {
  const channelFields = permissions.canViewChannels
    ? `
        channel {
          id
          name
          type
        }`
    : '';

  return `
    query GetUsageLog($id: ID!) {
      node(id: $id) {
        ... on UsageLog {
          id
          createdAt
          updatedAt
          requestID${channelFields}
          modelID
          promptTokens
          completionTokens
          totalTokens
          promptAudioTokens
          promptCachedTokens
          promptWriteCachedTokens
          completionAudioTokens
          completionReasoningTokens
          completionAcceptedPredictionTokens
          completionRejectedPredictionTokens
          source
          format
          totalCost
          costItems {
            itemCode
            quantity
            subtotal
          }
        }
      }
    }
  `;
}

// Query hooks
export function useUsageLogs(variables?: {
  first?: number;
  after?: string;
  orderBy?: { field: 'CREATED_AT'; direction: 'ASC' | 'DESC' };
  where?: {
    source?: string;
    modelID?: string;
    channelID?: string;
    requestID?: string;
    [key: string]: any;
  };
}, options?: { enabled?: boolean }) {
  const { handleError } = useErrorHandler();
  const { t } = useTranslation();
  const permissions = useUsageLogPermissions();
  const enabled = options?.enabled ?? true;

  return useQuery({
    queryKey: ['usageLogs', variables, permissions],
    queryFn: async () => {
      try {
        const query = buildUsageLogsQuery(permissions);
        const data = await graphqlRequest<{ usageLogs: UsageLogConnection }>(query, variables);
        return usageLogConnectionSchema.parse(data?.usageLogs);
      } catch (error) {
        handleError(error, t('common.errors.internalServerError'));
        throw error;
      }
    },
    enabled,
  });
}

export function useUsageLog(id: string) {
  const { handleError } = useErrorHandler();
  const { t } = useTranslation();
  const permissions = useUsageLogPermissions();

  return useQuery({
    queryKey: ['usageLog', id, permissions],
    queryFn: async () => {
      try {
        const query = buildUsageLogDetailQuery(permissions);
        const data = await graphqlRequest<{ node: UsageLog }>(query, { id });
        if (!data.node) {
          throw new Error('Usage log not found');
        }
        return usageLogSchema.parse(data.node);
      } catch (error) {
        handleError(error, t('common.errors.internalServerError'));
        throw error;
      }
    },
    enabled: !!id,
  });
}
