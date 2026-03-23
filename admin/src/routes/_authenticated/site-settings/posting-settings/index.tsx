import { createFileRoute } from '@tanstack/react-router'
import PostingSettingsManagement from '@/features/site-settings/posting-settings'

export const Route = createFileRoute(
  '/_authenticated/site-settings/posting-settings/'
)({
  component: PostingSettingsManagement,
})
