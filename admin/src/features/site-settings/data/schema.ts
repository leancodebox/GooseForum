import { z } from 'zod'

export const pItemSchema = z.object({
  content: z.string()
})

export const footerItemSchema = z.object({
  name: z.string(),
  url: z.string()
})

export const footerInfoSchema = z.object({
  primary: z.array(pItemSchema),
  list: z.array(footerItemSchema)
})

export const siteInfoSchema = z.object({
  siteName: z.string().min(1, '站点名称不能为空'),
  siteUrl: z.url('请输入有效的 URL').or(z.literal('')),
  siteLogo: z.string().or(z.literal('')),
  siteEmail: z.email('请输入有效的邮箱').or(z.literal('')),
  siteDescription: z.string().max(500, '站点描述不能超过 500 个字符'),
  siteKeywords: z.string().max(200, '关键词不能超过 200 个字符'),
  externalLinks: z.string().or(z.literal('')),
  footerInfo: footerInfoSchema,
  brandType: z.enum(['default', 'text', 'image']),
  brandText: z.string().or(z.literal('')),
  brandImage: z.string().or(z.literal(''))
})

export type SiteInfo = z.infer<typeof siteInfoSchema>
export type FooterInfo = z.infer<typeof footerInfoSchema>
export type PItem = z.infer<typeof pItemSchema>
export type FooterItem = z.infer<typeof footerItemSchema>

