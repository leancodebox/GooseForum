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
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { sponsorItemSchema, type SponsorItem } from '../data/schema'

interface SponsorActionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: SponsorItem) => void
  currentRow?: SponsorItem
  title: string
}

export function SponsorActionDialog({
  open,
  onOpenChange,
  onSubmit,
  currentRow,
  title,
}: SponsorActionDialogProps) {
  const form = useForm<SponsorItem>({
    resolver: zodResolver(sponsorItemSchema) as any,
    defaultValues: {
      name: '',
      logo: '',
      info: '',
      url: '',
      tag: [],
    },
  })

  useEffect(() => {
    if (currentRow) {
      form.reset(currentRow)
    } else {
      form.reset({
        name: '',
        logo: '',
        info: '',
        url: '',
        tag: [],
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
      <DialogContent className='sm:max-w-[525px]'>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form
            id='sponsor-form'
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
                  <FormLabel>赞助商名称</FormLabel>
                  <FormControl>
                    <Input placeholder='请输入赞助商名称' {...field} />
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
                  <FormLabel>Logo</FormLabel>
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
                        alt='Logo预览'
                        className='h-16 w-auto rounded border'
                      />
                    </div>
                  )}
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='info'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>描述信息</FormLabel>
                  <FormControl>
                    <Textarea placeholder='请输入赞助商描述' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='url'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>官网链接</FormLabel>
                  <FormControl>
                    <Input placeholder='请输入官网链接' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='tag'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>标签 (用逗号分隔)</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='例如: 技术, 开源, 云服务'
                      value={field.value.join(', ')}
                      onChange={(e) => {
                        const tags = e.target.value
                          .split(',')
                          .map((t) => t.trim())
                          .filter((t) => t !== '')
                        field.onChange(tags)
                      }}
                    />
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
          <Button type='submit' form='sponsor-form'>
            确定
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
