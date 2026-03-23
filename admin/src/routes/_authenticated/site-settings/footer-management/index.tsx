import { createFileRoute } from '@tanstack/react-router'
import FooterManagement from '@/features/site-settings/footer-management'

export const Route = createFileRoute(
  '/_authenticated/site-settings/footer-management/',
)({
  component: FooterManagement,
})
