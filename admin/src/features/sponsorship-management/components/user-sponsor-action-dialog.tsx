import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Upload } from 'lucide-react'
import { toast } from 'sonner'
import axios from 'axios'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
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
import { userSponsorSchema, type UserSponsor } from '../data/schema'

interface UserSponsorActionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: UserSponsor) => void
  currentRow?: UserSponsor
  title: string
}

export function UserSponsorActionDialog({
  open,
  onOpenChange,
  onSubmit,
  currentRow,
  title,
}: UserSponsorActionDialogProps) {
  const form = useForm<UserSponsor>({
    resolver: zodResolver(userSponsorSchema) as any,
    defaultValues: {
      name: '',
      logo: '',
      amount: '',
      time: new Date().toISOString().split('T')[0],
    },
  })

  useEffect(() => {
    if (currentRow) {
      form.reset(currentRow)
    } else {
      form.reset({
        name: '',
        logo: '',
        amount: '',
        time: new Date().toISOString().split('T')[0],
      })
    }
  }, [currentRow, form, open])

  const handleLogoUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    const formData = new FormData()
    formData.append('file', file)

    try {
      const response = await axios.post('/file/img-upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
      if (response.data.code === 0) {
        form.setValue('logo', response.data.result.url)
        toast.success('上传成功')
      } else {
        toast.error(response.data.message || '上传失败')
      }
    } catch (error) {
      toast.error('上传过程中发生错误')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[425px]'>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form
            id='user-sponsor-form'
            onSubmit={form.handleSubmit((data) => {
              onSubmit(data)
              onOpenChange(false)
            })}
            className='space-y-4'
          >
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>用户名称</FormLabel>
                  <FormControl>
                    <Input placeholder='请输入用户名称' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='logo'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>头像 (Logo)</FormLabel>
                  <div className='flex gap-2'>
                    <FormControl>
                      <Input placeholder='请输入Logo URL或上传图片' {...field} />
                    </FormControl>
                    <div className='relative'>
                      <Input
                        type='file'
                        className='absolute inset-0 opacity-0 cursor-pointer'
                        onChange={handleLogoUpload}
                        accept='image/*'
                      />
                      <Button type='button' variant='outline' size='icon'>
                        <Upload className='h-4 w-4' />
                      </Button>
                    </div>
                  </div>
                  {field.value && (
                    <div className='mt-2'>
                      <img
                        src={field.value}
                        alt='头像预览'
                        className='h-12 w-12 rounded-full border object-cover'
                      />
                    </div>
                  )}
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='amount'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>赞助金额</FormLabel>
                  <FormControl>
                    <Input placeholder='例如: ￥100' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='time'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>赞助时间</FormLabel>
                  <FormControl>
                    <Input type='date' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </form>
        </Form>
        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button type='submit' form='user-sponsor-form'>
            确定
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
