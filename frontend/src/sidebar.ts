import {
  IconActivity,
  IconAi,
  IconKey,
  IconLayoutDashboard,
  IconPlayerPlay,
  IconRobot,
  IconSettings,
  IconUserCircle,
} from '@tabler/icons-react';
import { Command } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import { useRoutePermissions } from '@/hooks/useRoutePermissions';
import { useMe } from '@/features/auth/data/auth';
import { type SidebarData, type NavGroup, type NavLink } from './components/layout/types';

export function useSidebarData(): SidebarData {
  const { t } = useTranslation();
  const { user: authUser } = useAuthStore((state) => state.auth);
  const { data: meData } = useMe();
  const { filterNavGroups } = useRoutePermissions();

  // Use data from me query if available, otherwise fall back to auth store
  const user = meData || authUser;

  // Generate user initials for avatar
  const getInitials = (firstName?: string, lastName?: string, email?: string) => {
    if (firstName && lastName) {
      return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase();
    }
    if (firstName) {
      return firstName.slice(0, 2).toUpperCase();
    }
    if (email) {
      return email.split('@')[0].slice(0, 2).toUpperCase();
    }
    return 'U';
  };

  // Generate user display name
  const getDisplayName = (firstName?: string, lastName?: string, email?: string) => {
    if (firstName && lastName) {
      return `${firstName} ${lastName}`;
    }
    if (firstName) {
      return firstName;
    }
    if (email) {
      const username = email.split('@')[0];
      return username.charAt(0).toUpperCase() + username.slice(1);
    }
    return 'User';
  };

  const rawNavGroups: NavGroup[] = [
    {
      title: t('sidebar.groups.admin'),
      items: [
        {
          title: t('sidebar.items.dashboard'),
          url: '/',
          icon: IconLayoutDashboard,
        } as NavLink,
        {
          title: t('sidebar.items.channels'),
          url: '/channels',
          icon: IconAi,
        } as NavLink,
        {
          title: t('sidebar.items.models'),
          url: '/models',
          icon: IconRobot,
        } as NavLink,
        {
          title: t('sidebar.items.apiKeys'),
          url: '/api-keys',
          icon: IconKey,
        } as NavLink,
        {
          title: t('sidebar.items.requests'),
          url: '/requests',
          icon: IconActivity,
        } as NavLink,
        {
          title: t('sidebar.items.playground'),
          url: '/playground',
          icon: IconPlayerPlay,
        } as NavLink,
      ],
    },
    {
      title: t('sidebar.groups.settings'),
      items: [
        {
          title: t('sidebar.items.system'),
          url: '/system',
          icon: IconSettings,
        } as NavLink,
        {
          title: t('profile.title'),
          url: '/settings/profile',
          icon: IconUserCircle,
        } as NavLink,
      ],
    },
  ];

  const filteredNavGroups = filterNavGroups(rawNavGroups);

  return {
    user: {
      name: getDisplayName(user?.firstName, user?.lastName, user?.email),
      email: user?.email || 'user@example.com',
      avatar: user?.avatar || getInitials(user?.firstName, user?.lastName, user?.email),
    },
    teams: [
      {
        name: t('sidebar.team.name'),
        logo: Command,
        description: '',
        // DO NOT USE THIS
        // plan: t('sidebar.team.plan'),
      },
    ],
    navGroups: filteredNavGroups,
  };
}
