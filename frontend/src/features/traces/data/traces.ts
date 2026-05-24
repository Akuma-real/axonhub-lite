import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { graphqlRequest } from '@/gql/graphql';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { Trace, TraceConnection, TraceDetail, traceConnectionSchema, traceDetailSchema } from './schema';

// GraphQL query for traces
function buildTracesQuery() {
  return `
    query GetTraces(
      $first: Int
      $after: Cursor
      $orderBy: TraceOrder
      $where: TraceWhereInput
    ) {
      traces(first: $first, after: $after, orderBy: $orderBy, where: $where) {
        edges {
          node {
            id
            traceID
            firstUserQuery
            createdAt
            updatedAt
            thread {
              id
              threadID
            }
            requests(where: { status: completed }) {
              totalCount
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

// GraphQL query for trace detail
function buildTraceDetailQuery() {
  return `
    query GetTraceDetail($id: ID!) {
      node(id: $id) {
        ... on Trace {
          id
          traceID
          createdAt
          updatedAt
          usageMetadata {
            totalInputTokens
            totalOutputTokens
            totalTokens
            totalCost
            totalCachedTokens
            totalCachedWriteTokens
          }
          thread {
            id
            threadID
          }
          requests(where: { status: completed }) {
            totalCount
          }
        }
      }
    }
  `;
}

// GraphQL query for trace with request traces
function buildTraceWithRequestTracesQuery() {
  return `query GetTraceWithSegments($id: ID!) {
      node(id: $id) {
        ... on Trace {
          id
          traceID
          createdAt
          updatedAt
          usageMetadata {
            totalInputTokens
            totalOutputTokens
            totalTokens
            totalCost
            totalCachedTokens
            totalCachedWriteTokens
          }
          thread {
            id
            threadID
          }
          requests(where: { status: completed }) {
            totalCount
          }
          rawRootSegment
        }
      }
    }
  `;
}

// Query hooks
export function useTraces(variables?: {
  first?: number;
  after?: string;
  orderBy?: { field: 'CREATED_AT'; direction: 'ASC' | 'DESC' };
  where?: {
    threadID?: string;
    traceID?: string;
    [key: string]: any;
  };
}) {
  const { handleError } = useErrorHandler();
  const { t } = useTranslation();

  return useQuery({
    queryKey: ['traces', variables],
    queryFn: async () => {
      try {
        const query = buildTracesQuery();
        const data = await graphqlRequest<{ traces: TraceConnection }>(query, variables);
        return traceConnectionSchema.parse(data?.traces);
      } catch (error) {
        handleError(error, t('common.errors.internalServerError'));
        throw error;
      }
    },
    enabled: true,
  });
}

export function useTrace(id: string) {
  const { handleError } = useErrorHandler();
  const { t } = useTranslation();

  return useQuery({
    queryKey: ['trace', id],
    queryFn: async () => {
      try {
        const query = buildTraceDetailQuery();
        const data = await graphqlRequest<{ node: Trace }>(query, { id });
        if (!data.node) {
          throw new Error('Trace not found');
        }
        return traceDetailSchema.parse(data.node);
      } catch (error) {
        handleError(error, t('common.errors.internalServerError'));
        throw error;
      }
    },
    enabled: !!id,
  });
}

export function useTraceWithSegments(id: string) {
  const { handleError } = useErrorHandler();
  const { t } = useTranslation();

  return useQuery({
    queryKey: ['trace-with-segments', id],
    queryFn: async () => {
      try {
        const query = buildTraceWithRequestTracesQuery();
        const data = await graphqlRequest<{ node: TraceDetail }>(query, { id });
        if (!data.node) {
          throw new Error('Trace not found');
        }
        return traceDetailSchema.parse(data.node);
      } catch (error) {
        handleError(error, t('common.errors.internalServerError'));
        throw error;
      }
    },
    enabled: !!id,
  });
}

// Backward compatibility alias
export const useTraceWithRequestTraces = useTraceWithSegments;
