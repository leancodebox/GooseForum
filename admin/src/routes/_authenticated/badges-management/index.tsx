import { createFileRoute } from '@tanstack/react-router'
import BadgesManagement from '@/features/badges-management'

export const Route = createFileRoute('/_authenticated/badges-management/')({
  component: BadgesManagement,
})
