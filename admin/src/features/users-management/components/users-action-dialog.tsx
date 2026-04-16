'use client'

import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect, useState } from 'react'
import axios from 'axios'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
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
    }
  }, [open])

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

  return (
    <Dialog
      open={open}
      onOpenChange={(state) => {
        form.reset()
        onOpenChange(state)
      }}
    >
      <DialogContent className='sm:max-w-lg'>
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
