export interface RequestPermissions {
  canViewUsers: boolean;
  canViewApiKeys: boolean;
  canViewChannels: boolean;
  canViewRoles: boolean;
}

export function useRequestPermissions(): RequestPermissions {
  return {
    canViewUsers: false,
    canViewApiKeys: true,
    canViewChannels: true,
    canViewRoles: false,
  };
}
