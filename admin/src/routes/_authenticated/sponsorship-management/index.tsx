import { createFileRoute } from '@tanstack/react-router'
import SponsorshipManagement from '@/features/sponsorship-management'

export const Route = createFileRoute('/_authenticated/sponsorship-management/')({
  component: SponsorshipManagement,
})
