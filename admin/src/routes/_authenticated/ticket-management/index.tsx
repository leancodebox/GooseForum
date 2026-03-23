import { createFileRoute } from '@tanstack/react-router'
import TicketManagement from '@/features/ticket-management'

export const Route = createFileRoute(
  '/_authenticated/ticket-management/'
)({
  component: TicketManagement,
})
