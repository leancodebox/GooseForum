'use client'

import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect, useState } from 'react'
import axios from 'axios'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { SelectDropdown } from '@/components/select-dropdown'
import { getUserBadgeOptions, saveUserBadges } from '@/api'
import type { BadgeItem, UserBadge } from '@/api/types'
import { type User } from '../data/schema'

const formSchema = z.object({
  status: z.string(),
  validate: z.string(),
  roleId: z.string(),
})

type UserForm = z.infer<typeof formSchema>

type UserActionDialogProps = {
  currentRow?: User | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function UsersActionDialog({
  currentRow,
  open,
  onOpenChange,
  onSuccess,
}: UserActionDialogProps) {
  const [roles, setRoles] = useState<{ label: string; value: string }[]>([])
  const [badgeOptions, setBadgeOptions] = useState<BadgeItem[]>([])
  const [activeBadges, setActiveBadges] = useState<UserBadge[]>([])
  const [selectedBadgeCodes, setSelectedBadgeCodes] = useState<string[]>([])
  const [loading, setLoading] = useState(false)

  const form = useForm<UserForm>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      status: '0',
      validate: '0',
      roleId: '0',
    },
  })

  useEffect(() => {
    if (open) {
      const fetchRoles = async () => {
        try {
          const response = await axios.get('/api/admin/get-all-role-item')
          if (response.data.code === 0) {
            const roleOptions = response.data.result.map((r: any) => ({
              label: r.name,
              value: r.value.toString(),
            }))
            // 添加“无”选项，对应值为 0
            setRoles([{ label: '无', value: '0' }, ...roleOptions])
          }
        } catch (error) {
          console.error('Failed to fetch roles:', error)
        }
      }
      fetchRoles()
      const fetchBadges = async () => {
        if (!currentRow?.userId) return
        try {
          const response = await getUserBadgeOptions(currentRow.userId)
          if (response.code === 0) {
            const options = response.result.options || []
            const active = response.result.active || []
            setBadgeOptions(options)
            setActiveBadges(active)
            const optionCodes = new Set(options.map((item) => item.code))
            setSelectedBadgeCodes(active.filter((item) => item.source === 'manual' && optionCodes.has(item.code)).map((item) => item.code))
          }
        } catch (error) {
          console.error('Failed to fetch user badges:', error)
        }
      }
      void fetchBadges()
    }
  }, [open, currentRow?.userId])

  useEffect(() => {
    if (currentRow) {
      form.reset({
        status: currentRow.status.toString(),
        validate: currentRow.validate.toString(),
        roleId: currentRow.roleId?.toString() || '0',
      })
    }
  }, [currentRow, form])

  const onSubmit = async (values: UserForm) => {
    if (!currentRow) return

    setLoading(true)
    try {
      const response = await axios.post('/api/admin/user-edit', {
        userId: currentRow.userId,
        status: parseInt(values.status),
        validate: parseInt(values.validate),
        roleId: parseInt(values.roleId),
      })

      if (response.data.code === 0) {
        await saveUserBadges(currentRow.userId, selectedBadgeCodes)
        toast.success('更新成功')
        onOpenChange(false)
        onSuccess?.()
      } else {
        toast.error(response.data.msg || '更新失败')
      }
    } catch (error) {
      console.error('Failed to update user:', error)
      toast.error('网络错误，请稍后再试')
    } finally {
      setLoading(false)
    }
  }

  const toggleBadge = (code: string, checked: boolean) => {
    setSelectedBadgeCodes((current) => {
      if (checked) return Array.from(new Set([...current, code]))
      return current.filter((item) => item !== code)
    })
  }

  const badgeIconURL = (badge: BadgeItem | UserBadge) => {
    return badge.iconUrl || '/static/badges/contributor.svg'
  }

  const badgeToneClass = (badge: BadgeItem | UserBadge) => {
    if (badge.color === 'blue') return 'bg-blue-100 text-blue-700 ring-blue-200'
    if (badge.color === 'emerald') return 'bg-emerald-100 text-emerald-700 ring-emerald-200'
    if (badge.color === 'teal') return 'bg-teal-100 text-teal-700 ring-teal-200'
    if (badge.color === 'sky') return 'bg-sky-100 text-sky-700 ring-sky-200'
    if (badge.color === 'cyan') return 'bg-cyan-100 text-cyan-700 ring-cyan-200'
    if (badge.color === 'rose') return 'bg-rose-100 text-rose-700 ring-rose-200'
    if (badge.color === 'violet') return 'bg-violet-100 text-violet-700 ring-violet-200'
    if (badge.color === 'purple') return 'bg-purple-100 text-purple-700 ring-purple-200'
    if (badge.color === 'fuchsia') return 'bg-fuchsia-100 text-fuchsia-700 ring-fuchsia-200'
    if (badge.color === 'indigo') return 'bg-indigo-100 text-indigo-700 ring-indigo-200'
    if (badge.color === 'amber') return 'bg-amber-100 text-amber-700 ring-amber-200'
    if (badge.color === 'orange') return 'bg-orange-100 text-orange-700 ring-orange-200'
    if (badge.color === 'yellow') return 'bg-yellow-100 text-yellow-700 ring-yellow-200'
    if (badge.color === 'slate') return 'bg-slate-100 text-slate-700 ring-slate-200'
    if (badge.level === 'gold') return 'bg-amber-100 text-amber-700 ring-amber-200'
    if (badge.level === 'special') return 'bg-indigo-100 text-indigo-700 ring-indigo-200'
    return 'bg-blue-100 text-blue-700 ring-blue-200'
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(state) => {
        form.reset()
        onOpenChange(state)
      }}
    >
      <DialogContent className='sm:max-w-4xl'>
        <DialogHeader className='text-start'>
          <DialogTitle>编辑用户</DialogTitle>
          <DialogDescription>
            修改用户状态、验证状态或角色。完成后点击保存。
          </DialogDescription>
        </DialogHeader>
        <div className='max-h-[60vh] overflow-y-auto py-1 pe-3'>
          <Form {...form}>
            <form
              id='user-form'
              onSubmit={form.handleSubmit(onSubmit)}
              className='space-y-4 px-0.5'
            >
              <FormField
                control={form.control}
                name='status'
                render={({ field }) => (
                  <FormItem className='grid grid-cols-6 items-center gap-x-4 gap-y-1 space-y-0'>
                    <FormLabel className='col-span-2 text-right'>账号状态</FormLabel>
                    <div className='col-span-4'>
                      <SelectDropdown
                        className='w-full'
                        placeholder='选择状态'
                        items={[
                          { label: '正常', value: '0' },
                          { label: '已封禁', value: '1' },
                        ]}
                        defaultValue={field.value}
                        onValueChange={field.onChange}
                      />
                    </div>
                    <FormMessage className='col-span-4 col-start-3' />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='validate'
                render={({ field }) => (
                  <FormItem className='grid grid-cols-6 items-center gap-x-4 gap-y-1 space-y-0'>
                    <FormLabel className='col-span-2 text-right'>验证状态</FormLabel>
                    <div className='col-span-4'>
                      <SelectDropdown
                        className='w-full'
                        placeholder='选择验证状态'
                        items={[
                          { label: '未验证', value: '0' },
                          { label: '已验证', value: '1' },
                        ]}
                        defaultValue={field.value}
                        onValueChange={field.onChange}
                      />
                    </div>
                    <FormMessage className='col-span-4 col-start-3' />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='roleId'
                render={({ field }) => (
                  <FormItem className='grid grid-cols-6 items-center gap-x-4 gap-y-1 space-y-0'>
                    <FormLabel className='col-span-2 text-right'>用户角色</FormLabel>
                    <div className='col-span-4'>
                      <SelectDropdown
                        className='w-full'
                        placeholder='选择角色'
                        items={roles}
                        defaultValue={field.value}
                        onValueChange={field.onChange}
                      />
                    </div>
                    <FormMessage className='col-span-4 col-start-3' />
                  </FormItem>
                )}
              />
              <div className='grid grid-cols-6 gap-x-4 gap-y-2'>
                <FormLabel className='col-span-2 pt-2 text-right'>手动徽章</FormLabel>
                <div className='col-span-4 space-y-3'>
                  {activeBadges.some((badge) => badge.source !== 'manual') && (
                    <div className='flex flex-wrap gap-1.5'>
                      {activeBadges.filter((badge) => badge.source !== 'manual').map((badge) => (
                        <Badge key={badge.code} variant='secondary'>{badge.name}</Badge>
                      ))}
                    </div>
                  )}
                  <div className='grid grid-cols-3 gap-2 sm:grid-cols-4 lg:grid-cols-6'>
                    {badgeOptions.map((badge) => (
                      <label
                        key={badge.code}
                        title={badge.description || badge.code}
                        className={`group relative flex aspect-square cursor-pointer flex-col items-center justify-center gap-1 rounded-md p-1.5 text-center transition-colors hover:bg-muted/50 ${
                          selectedBadgeCodes.includes(badge.code) ? 'bg-primary/5 ring-1 ring-primary/40' : ''
                        }`}
                      >
                        <div
                          className={`flex h-11 w-11 shrink-0 items-center justify-center ring-1 ring-inset transition-transform group-hover:scale-105 ${badgeToneClass(badge)}`}
                          style={{ clipPath: 'polygon(25% 5%, 75% 5%, 100% 50%, 75% 95%, 25% 95%, 0 50%)' }}
                        >
                          <img src={badgeIconURL(badge)} alt={badge.name} className='h-6 w-6 object-contain' />
                        </div>
                        <span className='block max-w-full truncate text-xs font-medium leading-5'>{badge.name}</span>
                        <Checkbox
                          className='absolute right-1 top-1 h-3.5 w-3.5 shrink-0 rounded-full bg-background'
                          checked={selectedBadgeCodes.includes(badge.code)}
                          onCheckedChange={(checked) => toggleBadge(badge.code, checked === true)}
                        />
                      </label>
                    ))}
                    {!badgeOptions.length && (
                      <div className='rounded-md border border-dashed p-3 text-sm text-muted-foreground'>暂无可手动下发的徽章。</div>
                    )}
                  </div>
                </div>
              </div>
            </form>
          </Form>
        </div>
        <DialogFooter>
          <Button type='submit' form='user-form' disabled={loading}>
            {loading ? '保存中...' : '保存更改'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
