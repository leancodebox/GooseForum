import { createFileRoute } from '@tanstack/react-router'
import RolesManagement from '@/features/roles-management'

export const Route = createFileRoute('/_authenticated/roles-management/')({
  component: RolesManagement,
})
