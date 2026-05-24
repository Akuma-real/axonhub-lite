import { useAuthStore } from '@/stores/authStore';
import { type NavGroup, type NavItem } from '@/components/layout/types';
import { useMe } from '@/features/auth/data/auth';

export function useRoutePermissions() {
  const { user: authUser } = useAuthStore((state) => state.auth);
  const { data: meData } = useMe();

  const user = meData || authUser;
  const isOwner = user?.isOwner ?? false;

  const checkRouteAccess = (_path: string): { hasAccess: boolean; mode?: 'hidden' | 'disabled' } => ({ hasAccess: true });
  const checkGroupAccess = (_group: unknown): boolean => true;
  const filterNavItems = (items: NavItem[]): NavItem[] => items;
  const filterNavGroups = (groups: NavGroup[]): NavGroup[] => groups;

  return {
    userScopes: ['*'],
    isOwner,
    checkRouteAccess,
    checkGroupAccess,
    filterNavItems,
    filterNavGroups,
  };
}
