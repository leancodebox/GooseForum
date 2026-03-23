import { z } from 'zod'

export const postingSettingsSchema = z.object({
  textControl: z.object({
    minPostLength: z.number().min(1),
    maxPostLength: z.number().min(1),
    minTitleLength: z.number().min(1),
    maxTitleLength: z.number().min(1),
    allowUppercasePosts: z.boolean(),
  }),
  uploadControl: z.object({
    allowAttachments: z.boolean(),
    authorizedExtensions: z.array(z.string()),
    maxAttachmentSizeKb: z.number().min(1),
    maxAttachmentsPerPost: z.number().min(1),
  }),
  editControl: z.object({
    editingGracePeriod: z.number().min(0),
    postEditTimeLimit: z.number().min(0),
    allowUsersToDeletePosts: z.boolean(),
  }),
})

export type PostingSettings = z.infer<typeof postingSettingsSchema>
