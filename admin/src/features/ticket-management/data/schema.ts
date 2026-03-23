import { z } from 'zod'

export const ticketSchema = z.object({
  id: z.number(),
  userId: z.number(),
  applyUserInfo: z.string(),
  type: z.number(),
  status: z.number(),
  title: z.string(),
  content: z.string(),
  reply: z.string().optional(),
  createTime: z.string(),
  updateTime: z.string(),
})

export type Ticket = z.infer<typeof ticketSchema>

export const TICKET_TYPES = {
  1: 'Bug反馈',
  2: '功能建议',
  3: '技术支持',
  4: '账户问题',
  5: '其他',
} as const

export const TICKET_STATUS = {
  1: '待处理',
  2: '处理中',
  3: '已解决',
  4: '已关闭',
} as const

export const TICKET_TYPE_VARIANTS = {
  1: 'destructive',
  2: 'default',
  3: 'warning',
  4: 'secondary',
  5: 'outline',
} as const

export const TICKET_STATUS_VARIANTS = {
  1: 'warning',
  2: 'default',
  3: 'success',
  4: 'secondary',
} as const
