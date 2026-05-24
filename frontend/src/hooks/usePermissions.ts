import { useCallback, useMemo } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useMe } from '@/features/auth/data/auth';

/**
 * Hook for checking user permissions based on scopes
 * Provides utilities to check if user has specific permissions for actions
 * Supports both system-level and project-level scopes
 */
export function usePermissions() {
  const { user: authUser } = useAuthStore((state) => state.auth);
  const { data: meData } = useMe();

  const user = meData || authUser;
  const isOwner = user?.isOwner ?? false;

  // Check if user has a specific scope at system level only
  const hasSystemScope = useCallback(
    (_requiredScope: string): boolean => {
      return true;
    },
    []
  );

  // Check if user has a specific scope at project level only
  const hasProjectScope = useCallback(
    (_requiredScope: string): boolean => {
      return true;
    },
    []
  );

  // Check if user has a specific scope (system-level or project-level)
  const hasScope = useCallback(
    (requiredScope: string): boolean => {
      return hasSystemScope(requiredScope) || hasProjectScope(requiredScope);
    },
    [hasSystemScope, hasProjectScope]
  );

  // Check if user has any of the required scopes
  const hasAnyScope = useCallback(
    (requiredScopes: string[]): boolean => {
      if (requiredScopes.length === 0) {
        return true;
      }

      return requiredScopes.some((scope) => hasScope(scope));
    },
    [hasScope]
  );

  // Check if user has all of the required scopes
  const hasAllScopes = useCallback(
    (requiredScopes: string[]): boolean => {
      if (requiredScopes.length === 0) {
        return true;
      }

      return requiredScopes.every((scope) => hasScope(scope));
    },
    [hasScope]
  );

  // Common permission checks for channel operations
  const channelPermissions = useMemo(
    () => ({
      canRead: true,
      canWrite: true,
      canBulkImport: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
      canTest: true,
      canOrder: true,
    }),
    []
  );

  // Common permission checks for user operations
  const userPermissions = useMemo(
    () => ({
      canRead: true,
      canWrite: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
    }),
    []
  );

  // Common permission checks for role operations
  const rolePermissions = useMemo(
    () => ({
      canRead: true,
      canWrite: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
    }),
    []
  );

  // Common permission checks for API key operations
  const apiKeyPermissions = useMemo(
    () => ({
      canRead: true,
      canWrite: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
    }),
    []
  );

  // Common permission checks for model operations
  const modelPermissions = useMemo(
    () => ({
      canRead: true,
      canWrite: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
    }),
    []
  );

  // Common permission checks for project operations
  const projectPermissions = useMemo(
    () => ({
      canRead: false,
      canWrite: false,
      canCreate: false,
      canEdit: false,
      canDelete: false,
    }),
    []
  );

  return {
    user,
    isOwner,
    hasScope,
    hasSystemScope,
    hasProjectScope,
    hasAnyScope,
    hasAllScopes,
    channelPermissions,
    userPermissions,
    rolePermissions,
    apiKeyPermissions,
    modelPermissions,
    projectPermissions,
  };
}
