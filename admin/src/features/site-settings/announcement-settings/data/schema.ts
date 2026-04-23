import { z } from 'zod'

export const announcementSettingsSchema = z.object({
  enabled: z.boolean(),
  title: z.string(),
  content: z.string(),
  link: z.string().optional(),
})

export type AnnouncementSettings = z.infer<typeof announcementSettingsSchema>