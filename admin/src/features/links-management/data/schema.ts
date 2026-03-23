import { z } from 'zod'

export const linkSchema = z.object({
  name: z.string().min(1, '链接名称不能为空'),
  desc: z.string(),
  url: z.url('请输入有效的 URL'),
  logoUrl: z.string(),
  status: z.number(),
})

export type Link = z.infer<typeof linkSchema>

export const linkGroupSchema = z.object({
  name: z.string().min(1, '分组名称不能为空'),
  emoji: z.string().optional(),
  color: z.string().optional(),
  links: z.array(linkSchema),
})

export type LinkGroup = z.infer<typeof linkGroupSchema>

export interface LinkItem extends Link {}

export interface FriendLinkGroup {
  name: string
  emoji?: string
  color?: string
  links: LinkItem[]
}