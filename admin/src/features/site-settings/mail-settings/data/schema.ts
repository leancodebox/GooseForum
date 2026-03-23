import { z } from 'zod'

export const mailSettingsSchema = z.object({
  enableMail: z.boolean(),
  smtpHost: z.string().min(1, 'SMTP主机不能为空'),
  smtpPort: z.number().min(1).max(65535),
  useSSL: z.boolean(),
  smtpUsername: z.email('请输入有效的邮箱地址'),
  smtpPassword: z.string().min(1, '密码/授权码不能为空'),
  fromName: z.string().min(1, '发件人名称不能为空'),
  fromEmail: z.email('请输入有效的发件人邮箱'),
})

export type MailSettings = z.infer<typeof mailSettingsSchema>
