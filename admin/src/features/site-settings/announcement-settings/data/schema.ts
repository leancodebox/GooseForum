import { z } from 'zod'

export const announcementSettingsSchema = z.object({
  enabled: z.boolean(),
  content: z.string(),
})

export type AnnouncementSettings = z.infer<typeof announcementSettingsSchema>
