import { z } from 'zod'

export const siteInfoSchema = z.object({
  siteName: z.string().min(1, '站点名称不能为空'),
  siteUrl: z.url('请输入有效的 URL').or(z.literal('')),
  siteLogo: z.string().or(z.literal('')),
  siteEmail: z.email('请输入有效的邮箱').or(z.literal('')),
  siteDescription: z.string().max(500, '站点描述不能超过 500 个字符'),
  siteKeywords: z.string().max(200, '关键词不能超过 200 个字符'),
  externalLinks: z.string(),
})

export type SiteInfo = z.infer<typeof siteInfoSchema>
