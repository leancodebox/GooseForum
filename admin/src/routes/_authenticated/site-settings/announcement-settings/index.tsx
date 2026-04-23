import { createFileRoute } from '@tanstack/react-router'
import AnnouncementSettingsManagement from '@/features/site-settings/announcement-settings'

export const Route = createFileRoute(
  '/_authenticated/site-settings/announcement-settings/'
)({
  component: AnnouncementSettingsManagement,
})