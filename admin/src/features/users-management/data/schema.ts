import { z } from 'zod'

export const userSchema = z.object({
  userId: z.number(),
  username: z.string(),
  avatarUrl: z.string().nullable().optional(),
  email: z.string(),
  status: z.number(),
  validate: z.number(),
  prestige: z.number(),
  roleList: z.array(z.object({
    name: z.string(),
    value: z.number()
  })).nullable().optional(),
  roleId: z.number().nullable().optional(),
  createTime: z.string(),
  lastActiveTime: z.string().nullable().optional(),
})

export type User = z.infer<typeof userSchema>

export const userListSchema = z.array(userSchema)
