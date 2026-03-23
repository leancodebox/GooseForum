import { z } from 'zod'

export const footerItemSchema = z.object({
  name: z.string().min(1, '名称不能为空'),
  url: z.url('请输入有效的URL'),
})

export const footerGroupSchema = z.object({
  name: z.string().min(1, '分组名称不能为空'),
  children: z.array(footerItemSchema),
})

export const pItemSchema = z.object({
  content: z.string().min(1, '内容不能为空'),
})

export const footerConfigSchema = z.object({
  primary: z.array(pItemSchema),
  list: z.array(footerGroupSchema),
})

export type FooterItem = z.infer<typeof footerItemSchema>
export type FooterGroup = z.infer<typeof footerGroupSchema>
export type PItem = z.infer<typeof pItemSchema>
export type FooterConfig = z.infer<typeof footerConfigSchema>
