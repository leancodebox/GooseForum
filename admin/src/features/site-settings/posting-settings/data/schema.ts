import { z } from 'zod'

export const postingSettingsSchema = z.object({
  textControl: z.object({
    minPostLength: z.number().min(1),
    maxPostLength: z.number().min(1),
    minTitleLength: z.number().min(1),
    maxTitleLength: z.number().min(1),
    newUserPostCooldownMinutes: z.number().min(0),
  }),
  uploadControl: z.object({
    allowAttachments: z.boolean(),
    authorizedExtensions: z.array(z.string()),
    maxAttachmentSizeKb: z.number().min(1),
    maxDailyUploadsPerUser: z.number().min(0),
    newUserUploadCooldownMinutes: z.number().min(0),
  }),
})

export type PostingSettings = z.infer<typeof postingSettingsSchema>
