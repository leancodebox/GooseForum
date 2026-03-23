import { z } from 'zod'

export const postSchema = z.object({
  id: z.number(),
  title: z.string(),
  description: z.string().nullable().optional(),
  type: z.number(),
  userId: z.number(),
  username: z.string(),
  userAvatarUrl: z.string().nullable().optional(),
  articleStatus: z.number(), // 文章状态：0 草稿 1 发布
  processStatus: z.number(), // 管理状态：0 正常 1 封禁
  viewCount: z.number(),
  replyCount: z.number(),
  likeCount: z.number(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export type Post = z.infer<typeof postSchema>

export const postListSchema = z.array(postSchema)
