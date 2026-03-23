import { createFileRoute } from '@tanstack/react-router'
import SecuritySettingsManagement from '@/features/site-settings/security-settings'

export const Route = createFileRoute(
  '/_authenticated/site-settings/security-settings/'
)({
  component: SecuritySettingsManagement,
})
