import { createFileRoute } from '@tanstack/react-router'
import { DashboardStats } from '@/features/dashboard-stats'

export const Route = createFileRoute('/_authenticated/')({
  component: DashboardStats,
})
