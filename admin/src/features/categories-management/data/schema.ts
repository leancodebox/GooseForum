import { z } from 'zod'

export const categorySchema = z.object({
  id: z.number(),
  category: z.string().min(1, '分类名称不能为空'),
  desc: z.string().optional(),
  icon: z.string().optional(),
  color: z.string().optional(),
  slug: z.string().optional(),
  sort: z.number().optional(),
  status: z.number().optional(),
})

export type Category = z.infer<typeof categorySchema>
