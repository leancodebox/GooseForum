import { useEffect } from 'react'
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
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import { linkGroupSchema, type LinkGroup } from '../data/schema'

const PRESET_COLORS = [
  '#3b82f6', // Blue (bg-blue-500)
  '#22c55e', // Green (bg-green-500)
  '#a855f7', // Purple (bg-purple-500)
  '#64748b', // Slate
  '#ef4444', // Red
  '#f97316', // Orange
  '#f59e0b', // Amber
  '#10b981', // Emerald
  '#06b6d4', // Cyan
  '#6366f1', // Indigo
  '#8b5cf6', // Violet
  '#ec4899', // Pink
]

const PRESET_EMOJIS = ['🔗', '👥', '✍️', '📁', '⭐', '🔥', '💡', '🛠️', '🎨', '💬', '📢', '🏠', '🚀', '🎁']

interface Props {
  currentRow?: LinkGroup | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: LinkGroup) => void
}

export function LinksGroupActionDialog({ currentRow, open, onOpenChange, onSubmit }: Props) {
  const isEdit = !!currentRow
  const form = useForm<LinkGroup>({
    resolver: zodResolver(linkGroupSchema),
    defaultValues: currentRow ?? {
      name: '',
      emoji: '',
      color: '#3b82f6',
      links: [],
    },
  })

  useEffect(() => {
    if (currentRow) {
      form.reset(currentRow)
    } else {
      form.reset({
        name: '',
        emoji: '',
        color: '#3b82f6',
        links: [],
      })
    }
  }, [currentRow, form])

  const handleSubmit = (data: LinkGroup) => {
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
          <DialogTitle>{isEdit ? '编辑分组' : '添加分组'}</DialogTitle>
          <DialogDescription>
            {isEdit ? '修改友情链接分组名称。' : '添加一个新的友情链接分组。'}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form
            id='group-form'
            onSubmit={form.handleSubmit(handleSubmit)}
            className='space-y-4'
          >
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>分组名称</FormLabel>
                  <FormControl>
                    <Input placeholder='输入分组名称' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className='grid grid-cols-2 gap-4'>
              <FormField
                control={form.control}
                name='emoji'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Emoji 图标</FormLabel>
                    <FormControl>
                      <div className='flex gap-2'>
                        <Popover>
                          <PopoverTrigger asChild>
                            <Button
                              variant='outline'
                              className='h-9 w-9 p-0'
                              type='button'
                            >
                              {field.value || '🔗'}
                            </Button>
                          </PopoverTrigger>
                          <PopoverContent className='w-auto p-2'>
                            <div className='grid grid-cols-4 gap-2'>
                              {PRESET_EMOJIS.map((emoji) => (
                                <button
                                  key={emoji}
                                  type='button'
                                  className={cn(
                                    'h-8 w-8 rounded text-lg transition-all hover:bg-accent hover:scale-110 active:scale-95',
                                    field.value === emoji && 'bg-accent ring-1 ring-primary'
                                  )}
                                  onClick={() => field.onChange(emoji)}
                                >
                                  {emoji}
                                </button>
                              ))}
                            </div>
                          </PopoverContent>
                        </Popover>
                        <Input placeholder='Emoji' {...field} className='flex-1' />
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='color'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>颜色</FormLabel>
                    <FormControl>
                      <div className='flex gap-2'>
                        <Popover>
                          <PopoverTrigger asChild>
                            <button
                              type='button'
                              className='h-9 w-9 rounded border border-input transition-all hover:scale-110 active:scale-95'
                              style={{ backgroundColor: field.value || '#64748b' }}
                            />
                          </PopoverTrigger>
                          <PopoverContent className='w-auto p-3'>
                            <div className='space-y-3'>
                              <div className='text-xs font-medium text-muted-foreground'>选择预设颜色</div>
                              <div className='grid grid-cols-4 gap-2'>
                                {PRESET_COLORS.map((color) => (
                                  <button
                                    key={color}
                                    type='button'
                                    className={cn(
                                      'h-7 w-7 rounded-full border border-black/10 transition-transform hover:scale-110 active:scale-95',
                                      field.value === color && 'ring-2 ring-primary ring-offset-2'
                                    )}
                                    style={{ backgroundColor: color }}
                                    onClick={() => field.onChange(color)}
                                  />
                                ))}
                                <button
                                  type='button'
                                  className={cn(
                                    'h-7 w-7 rounded-full border border-dashed border-input bg-white text-[10px] font-medium transition-transform hover:scale-110 active:scale-95',
                                    !field.value && 'ring-2 ring-primary ring-offset-2'
                                  )}
                                  onClick={() => field.onChange('')}
                                >
                                  清除
                                </button>
                              </div>
                            </div>
                          </PopoverContent>
                        </Popover>
                        <Input placeholder='#hex' {...field} className='flex-1' />
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </form>
        </Form>
        <DialogFooter>
          <Button type='submit' form='group-form'>
            保存
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
