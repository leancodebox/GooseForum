import { createFileRoute } from '@tanstack/react-router'
import PostsManagement from '@/features/posts-management'

export const Route = createFileRoute('/_authenticated/posts-management/')({
  component: PostsManagement,
})
