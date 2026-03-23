import { createFileRoute } from '@tanstack/react-router'
import LinksManagement from '@/features/links-management'

export const Route = createFileRoute('/_authenticated/links-management/')({
  component: LinksManagement,
})
