import { AuthUser } from '@/stores/authStore';

export function getUserScopes(user: AuthUser | null): string[] {
  if (!user) return [];

  const scopes = new Set<string>();

  user.scopes?.forEach((scope) => scopes.add(scope));

  return Array.from(scopes);
}

export function hasScope(user: AuthUser | null, scope: string): boolean {
  if (!user) return false;
  if (user.isOwner) return true;

  const userScopes = getUserScopes(user);
  return userScopes.includes(scope);
}

export function isProjectOwner(user: AuthUser | null): boolean {
  return !!user?.isOwner;
}

export function filterGrantableScopes(currentUser: AuthUser | null, allScopes: string[]): string[] {
  if (!currentUser) return [];

  if (currentUser.isOwner) {
    return allScopes;
  }

  const userScopes = getUserScopes(currentUser);

  return allScopes.filter((scope) => userScopes.includes(scope));
}

export function canGrantScopes(currentUser: AuthUser | null, scopesToGrant: string[]): boolean {
  if (!currentUser) return false;

  if (currentUser.isOwner) {
    return true;
  }

  const userScopes = getUserScopes(currentUser);

  return scopesToGrant.every((scope) => userScopes.includes(scope));
}

export function filterGrantableRoles<T extends { scopes?: string[] }>(
  currentUser: AuthUser | null,
  allRoles: T[]
): T[] {
  if (!currentUser) return [];

  if (currentUser.isOwner) {
    return allRoles;
  }

  const userScopes = getUserScopes(currentUser);

  return allRoles.filter((role) => {
    if (!role.scopes || role.scopes.length === 0) return true;

    return role.scopes.every((scope) => userScopes.includes(scope));
  });
}

export function canGrantRole(currentUser: AuthUser | null, roleScopes: string[]): boolean {
  if (!currentUser) return false;

  if (currentUser.isOwner) {
    return true;
  }

  const userScopes = getUserScopes(currentUser);

  return roleScopes.every((scope) => userScopes.includes(scope));
}

export function canEditUserPermissions(
  currentUser: AuthUser | null,
  targetUserScopes: string[],
  targetUserIsOwner: boolean
): boolean {
  if (!currentUser) return false;

  if (targetUserIsOwner && !currentUser.isOwner) {
    return false;
  }

  if (currentUser.isOwner) {
    return true;
  }

  const userScopes = getUserScopes(currentUser);

  return targetUserScopes.every((scope) => userScopes.includes(scope));
}
