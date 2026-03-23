import { createFileRoute } from '@tanstack/react-router'
import SiteInfo from '@/features/site-settings'

export const Route = createFileRoute('/_authenticated/site-settings/site-info/')({
  component: SiteInfo,
})
