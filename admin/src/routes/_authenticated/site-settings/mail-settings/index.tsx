import { createFileRoute } from '@tanstack/react-router'
import MailSettingsManagement from '@/features/site-settings/mail-settings'

export const Route = createFileRoute(
  '/_authenticated/site-settings/mail-settings/'
)({
  component: MailSettingsManagement,
})
