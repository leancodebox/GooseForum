import { z } from 'zod'

export const securitySettingsSchema = z.object({
  enableSignup: z.boolean(),
  enableEmailVerification: z.boolean(),
  allowedDomains: z.array(z.string()),
})

export type SecuritySettings = z.infer<typeof securitySettingsSchema>
