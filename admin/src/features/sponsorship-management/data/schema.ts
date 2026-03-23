import { z } from 'zod'

export const sponsorItemSchema = z.object({
  name: z.string().min(1, '赞助商名称不能为空'),
  logo: z.string().min(1, 'Logo不能为空'),
  info: z.string().min(1, '描述信息不能为空'),
  url: z.url('请输入有效的URL'),
  tag: z.array(z.string()),
})

export type SponsorItem = z.infer<typeof sponsorItemSchema>

export const userSponsorSchema = z.object({
  name: z.string().min(1, '用户名称不能为空'),
  logo: z.string(),
  amount: z.string(),
  time: z.string(),
})

export type UserSponsor = z.infer<typeof userSponsorSchema>

export const sponsorsSchema = z.object({
  level0: z.array(sponsorItemSchema),
  level1: z.array(sponsorItemSchema),
  level2: z.array(sponsorItemSchema),
  level3: z.array(sponsorItemSchema),
})

export type Sponsors = z.infer<typeof sponsorsSchema>

export const sponsorsConfigSchema = z.object({
  sponsors: sponsorsSchema,
  users: z.array(userSponsorSchema),
})

export type SponsorsConfig = z.infer<typeof sponsorsConfigSchema>
