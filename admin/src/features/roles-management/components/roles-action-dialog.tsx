import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import axios from 'axios'
import { toast } from 'sonner'
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
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useRoles } from './roles-provider'

const roleFormSchema = z.object({
  id: z.number().optional(),
  roleName: z.string().min(1, '角色名称不能为空'),
  permissions: z.array(z.number()).min(1, '请至少选择一个权限'),
})

type RoleFormValues = z.infer<typeof roleFormSchema>

interface PermissionOption {
  id: number
  name: string
}

export function RolesActionDialog() {
  const { open, setOpen, currentRow, setCurrentRow } = useRoles()

  const form = useForm<RoleFormValues>({
    resolver: zodResolver(roleFormSchema),
    defaultValues: {
      roleName: '',
      permissions: [],
    },
  })

  // 临时：根据后端 permission.go 定义权限
  const permissionOptions: PermissionOption[] = [
    { id: 0, name: '管理员' },
    { id: 1, name: '用户管理' },
    { id: 2, name: '文章管理' },
    { id: 3, name: '页面管理' },
    { id: 4, name: '角色管理' },
    { id: 5, name: '站点管理' },
  ]

  useEffect(() => {
    if (currentRow && open === 'edit') {
      form.reset({
        id: currentRow.roleId,
        roleName: currentRow.roleName,
        permissions: currentRow.permissions.map((p) => p.id),
      })
    } else {
      form.reset({
        roleName: '',
        permissions: [],
      })
    }
  }, [currentRow, open, form])

  const onSubmit = async (data: RoleFormValues) => {
    try {
      const response = await axios.post('/api/admin/role-save', {
        id: data.id || 0,
        roleName: data.roleName,
        permissions: data.permissions,
      })

      if (response.data.code === 0) {
        toast.success(open === 'add' ? '添加成功' : '编辑成功')
        setOpen(null)
        setCurrentRow(null)
        // 刷新列表的逻辑在父组件
        window.location.reload()
      } else {
        toast.error(response.data.message || '操作失败')
      }
    } catch (error) {
      toast.error('请求失败')
    }
  }

  return (
    <Dialog
      open={open === 'add' || open === 'edit'}
      onOpenChange={(val) => !val && setOpen(null)}
    >
      <DialogContent className='sm:max-w-[500px]'>
        <DialogHeader>
          <DialogTitle>{open === 'add' ? '添加角色' : '编辑角色'}</DialogTitle>
          <DialogDescription>
            填写角色名称并分配权限。
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4'>
            <FormField
              control={form.control}
              name='roleName'
              render={({ field }) => (
                <FormItem className='grid grid-cols-6 items-center gap-x-4 gap-y-1 space-y-0'>
                  <FormLabel className='col-span-2 text-right'>角色名称</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      placeholder='输入角色名称'
                      className='col-span-4'
                    />
                  </FormControl>
                  <FormMessage className='col-span-4 col-start-3' />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='permissions'
              render={() => (
                <FormItem className='grid grid-cols-6 items-start gap-x-4 gap-y-1 space-y-0'>
                  <FormLabel className='col-span-2 mt-2 text-right'>
                    权限分配
                  </FormLabel>
                  <div className='col-span-4'>
                    <ScrollArea className='h-[200px] w-full rounded-md border p-4'>
                      <div className='grid grid-cols-2 gap-4'>
                        {permissionOptions.map((item) => (
                          <FormField
                            key={item.id}
                            control={form.control}
                            name='permissions'
                            render={({ field }) => {
                              return (
                                <FormItem
                                  key={item.id}
                                  className='flex flex-row items-start space-x-3 space-y-0'
                                >
                                  <FormControl>
                                    <Checkbox
                                      checked={field.value?.includes(item.id)}
                                      onCheckedChange={(checked) => {
                                        return checked
                                          ? field.onChange([
                                              ...field.value,
                                              item.id,
                                            ])
                                          : field.onChange(
                                              field.value?.filter(
                                                (value) => value !== item.id
                                              )
                                            )
                                      }}
                                    />
                                  </FormControl>
                                  <FormLabel className='text-sm font-normal'>
                                    {item.name}
                                  </FormLabel>
                                </FormItem>
                              )
                            }}
                          />
                        ))}
                      </div>
                    </ScrollArea>
                    <FormMessage className='mt-1' />
                  </div>
                </FormItem>
              )}
            />
            <DialogFooter>
              <Button type='submit'>保存</Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
