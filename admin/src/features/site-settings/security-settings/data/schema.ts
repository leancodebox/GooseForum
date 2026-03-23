import { z } from 'zod'

export const securitySettingsSchema = z.object({
  enableSignup: z.boolean(),
  enableEmailVerification: z.boolean(),
  mustApproveUsers: z.boolean(),
  minPasswordLength: z.number().min(1).max(100),
  inviteOnly: z.boolean(),
  restrictions: z.object({
    allowedDomains: z.array(z.string()),
    blockedDomains: z.array(z.string()),
    maxRegistrationsPerIp: z.number().min(0),
  }),
})

export type SecuritySettings = z.infer<typeof securitySettingsSchema>
