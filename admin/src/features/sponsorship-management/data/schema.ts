import { z } from 'zod'

export const sponsorItemSchema = z.object({
  name: z.string().min(1, '赞助商名称不能为空'),
  avatarUrl: z.string().min(1, 'Logo不能为空'),
  message: z.string().min(1, '描述信息不能为空'),
  link: z.string(),
})

export type SponsorItem = z.infer<typeof sponsorItemSchema>

export const sponsorsSchema = z.object({
  level0: z.array(sponsorItemSchema),
  level1: z.array(sponsorItemSchema),
  level2: z.array(sponsorItemSchema),
  level3: z.array(sponsorItemSchema),
})

export type Sponsors = z.infer<typeof sponsorsSchema>

export const sponsorsContentSchema = z.object({
  title: z.string(),
  description: z.string(),
})

export const sponsorsContactSchema = z.object({
  title: z.string(),
  description: z.string(),
  buttonText: z.string(),
  buttonLink: z.string(),
})

export const sponsorsRuleSchema = z.object({
  content: z.string(),
})

export const sponsorsConfigSchema = z.object({
  sponsors: sponsorsSchema,
  content: sponsorsContentSchema,
  contact: sponsorsContactSchema,
  rules: z.array(sponsorsRuleSchema),
})

export type SponsorsConfig = z.infer<typeof sponsorsConfigSchema>
export type SponsorsRule = z.infer<typeof sponsorsRuleSchema>
