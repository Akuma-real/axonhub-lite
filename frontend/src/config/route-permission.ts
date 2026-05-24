// 路由权限配置
export type ScopeLevel = 'system' | 'project' | 'any';

export interface RouteConfig {
  path: string;
  requiredScopes?: string[];
  scopeLevel?: ScopeLevel; // 权限级别：system 只检查系统级权限，project 只检查项目级权限，any 检查两者
  mode?: 'hidden' | 'disabled'; // 当没有权限时的处理方式
  children?: RouteConfig[];
}

export interface RouteGroup {
  title: string;
  scopeLevel?: ScopeLevel; // 路由组的默认权限级别
  routes: RouteConfig[];
}

// 定义所有路由的权限配置
export const routeConfigs: RouteGroup[] = [
  {
    title: 'Admin',
    routes: [
      { path: '/' },
      { path: '/channels' },
      { path: '/models' },
      { path: '/api-keys' },
      { path: '/requests' },
      { path: '/traces' },
      { path: '/threads' },
      { path: '/playground' },
    ],
  },
  {
    title: 'Settings',
    routes: [
      { path: '/system' },
      { path: '/settings' },
      { path: '/settings/profile' },
      { path: '/settings/appearance' },
      { path: '/settings/notifications' },
    ],
  },
];

// 获取路由配置的辅助函数
export function getRouteConfig(path: string): RouteConfig | undefined {
  for (const group of routeConfigs) {
    for (const route of group.routes) {
      if (route.path === path) {
        return route;
      }
      if (route.children) {
        const childConfig = route.children.find((child) => child.path === path);
        if (childConfig) return childConfig;
      }
    }
  }
  return undefined;
}

// 检查用户是否有访问路由的权限
export function hasRouteAccess(userScopes: string[], routeConfig: RouteConfig): boolean {
  if (!routeConfig.requiredScopes || routeConfig.requiredScopes.length === 0) {
    return true;
  }

  // 如果用户有通配符权限，则拥有所有权限
  if (userScopes.includes('*')) {
    return true;
  }

  // 检查用户是否拥有所需的任一权限
  return routeConfig.requiredScopes.some((scope) => userScopes.includes(scope));
}

// 检查用户是否有访问路由组的权限
export function hasGroupAccess(userScopes: string[], group: RouteGroup): boolean {
  return group.routes.some((route) => hasRouteAccess(userScopes, route));
}
