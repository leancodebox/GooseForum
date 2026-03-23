import { createFileRoute } from '@tanstack/react-router'
import CategoriesManagement from '@/features/categories-management'

export const Route = createFileRoute('/_authenticated/categories-management/')({
  component: CategoriesManagement,
})
