import { z } from 'zod'

export const permissionSchema = z.object({
  id: z.number(),
  name: z.string(),
})

export type Permission = z.infer<typeof permissionSchema>

export const roleSchema = z.object({
  roleId: z.number(),
  roleName: z.string(),
  effective: z.number(),
  permissions: z.array(permissionSchema),
  createTime: z.string(),
})

export type Role = z.infer<typeof roleSchema>
