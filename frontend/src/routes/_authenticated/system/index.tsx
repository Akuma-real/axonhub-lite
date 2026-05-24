import { createFileRoute } from '@tanstack/react-router';
import { RouteGuard } from '@/components/route-guard';
import SystemManagement, { type SystemTabKey } from '@/features/system';

const systemTabKeys = [
  'general',
  'brand',
  'storage',
  'retry',
  'webhook',
  'proxy',
  'quota',
  'diagnostics',
  'about',
] as const satisfies readonly SystemTabKey[];

function ProtectedSystem() {
  const search = Route.useSearch();

  return (
    <RouteGuard requiredScopes={['read_system']}>
      <SystemManagement initialTab={search.tab} />
    </RouteGuard>
  );
}

export const Route = createFileRoute('/_authenticated/system/')({
  component: ProtectedSystem,
  validateSearch: (search: Record<string, unknown>): { tab?: SystemTabKey } => {
    if (typeof search.tab === 'string' && systemTabKeys.includes(search.tab as SystemTabKey)) {
      return { tab: search.tab as SystemTabKey };
    }

    return {};
  },
});
