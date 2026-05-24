import { useTranslation } from 'react-i18next';
import ContentSection from '../components/content-section';
import ProfileForm from './profile-form';
import SecurityForm from '../security/security-form';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

export default function SettingsProfile() {
  const { t } = useTranslation();

  return (
    <ContentSection title={t('profile.title')} desc={t('profile.description')}>
      <Tabs defaultValue="profile" className="w-full">
        <TabsList className="mb-4">
          <TabsTrigger value="profile">{t('profile.title')}</TabsTrigger>
          <TabsTrigger value="security">{t('security.title', 'Security')}</TabsTrigger>
        </TabsList>
        <TabsContent value="profile">
          <ProfileForm />
        </TabsContent>
        <TabsContent value="security">
          <SecurityForm />
        </TabsContent>
      </Tabs>
    </ContentSection>
  );
}
