import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
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
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { ScrollArea } from '@/components/ui/scroll-area'
import { linkSchema, type Link } from '../data/schema'

interface Props {
  currentRow?: Link | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: Link) => void
}

export function LinksActionDialog({ currentRow, open, onOpenChange, onSubmit }: Props) {
  const isEdit = !!currentRow
  const form = useForm<Link>({
    resolver: zodResolver(linkSchema),
    values: (currentRow ?? {
      name: '',
      desc: '',
      url: '',
      logoUrl: '',
      status: 1,
    }) as Link,
  })

  const handleSubmit = (data: Link) => {
    onSubmit(data)
    onOpenChange(false)
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        onOpenChange(v)
        if (!v) form.reset()
      }}
    >
      <DialogContent className='sm:max-w-[425px]'>
        <DialogHeader>
          <DialogTitle>{isEdit ? '编辑链接' : '添加链接'}</DialogTitle>
          <DialogDescription>
            {isEdit ? '修改友情链接信息。' : '添加一个新的友情链接。'}
          </DialogDescription>
        </DialogHeader>
        <ScrollArea className='max-h-[80vh] px-1'>
          <Form {...form}>
            <form
              id='link-form'
              onSubmit={form.handleSubmit(handleSubmit)}
              className='space-y-4 p-1'
            >
              <FormField
                control={form.control}
                name='name'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>名称</FormLabel>
                    <FormControl>
                      <Input placeholder='链接名称' {...field} />
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
                    <FormLabel>URL</FormLabel>
                    <FormControl>
                      <Input placeholder='https://...' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='desc'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>描述</FormLabel>
                    <FormControl>
                      <Input placeholder='链接描述' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='logoUrl'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Logo URL</FormLabel>
                    <FormControl>
                      <Input placeholder='Logo 图片地址' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='status'
                render={({ field }) => (
                  <FormItem className='flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm'>
                    <div className='space-y-0.5'>
                      <FormLabel>展示状态</FormLabel>
                      <div className='text-[12px] text-muted-foreground'>
                        控制该链接是否在前端页面展示
                      </div>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value === 1}
                        onCheckedChange={(checked) => field.onChange(checked ? 1 : 0)}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
            </form>
          </Form>
        </ScrollArea>
        <DialogFooter>
          <Button type='submit' form='link-form'>
            保存
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
