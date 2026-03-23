import { createFileRoute } from '@tanstack/react-router'
import UsersManagement from '@/features/users-management'

export const Route = createFileRoute('/_authenticated/users-management/')({
  component: UsersManagement,
})
