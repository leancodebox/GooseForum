import { useState, useEffect } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import axios from 'axios'
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
import { useCategories } from './categories-provider'
import { categorySchema, type Category } from '../data/schema'

const PRESET_COLORS = [
  '#64748b', // Slate
  '#ef4444', // Red
  '#f97316', // Orange
  '#f59e0b', // Amber
  '#22c55e', // Green
  '#10b981', // Emerald
  '#06b6d4', // Cyan
  '#3b82f6', // Blue
  '#6366f1', // Indigo
  '#8b5cf6', // Violet
  '#a855f7', // Purple
  '#ec4899', // Pink
]

interface CategoriesActionDialogProps {
  onSuccess: () => void
}

export function CategoriesActionDialog({ onSuccess }: CategoriesActionDialogProps) {
  const { open, setOpen, currentRow } = useCategories()
  const [loading, setLoading] = useState(false)
  const isEdit = open === 'edit'

  const form = useForm<Category>({
    resolver: zodResolver(categorySchema),
    defaultValues: {
      id: 0,
      category: '',
      desc: '',
      icon: '',
      color: '',
      slug: '',
      sort: 0,
      status: 1,
    },
  })

  useEffect(() => {
    if (open === 'edit' && currentRow) {
      form.reset({
        ...currentRow,
        icon: currentRow.icon || '',
        color: currentRow.color || '',
        slug: currentRow.slug || '',
      })
    } else if (open === 'create') {
      form.reset({
        id: 0,
        category: '',
        desc: '',
        icon: '',
        color: '',
        slug: '',
        sort: 0,
        status: 1,
      })
    }
  }, [open, currentRow, form])

  const onSubmit = async (data: Category) => {
    setLoading(true)
    try {
      const response = await axios.post('/api/admin/category-save', {
        ...data,
        id: isEdit ? data.id : 0,
      })

      if (response.data.code === 0) {
        toast.success(isEdit ? '编辑成功' : '新增成功')
        setOpen(null)
        onSuccess()
      } else {
        toast.error(response.data.message || '操作失败')
      }
    } catch (error) {
      console.error('Failed to save category:', error)
      toast.error('操作失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open === 'create' || open === 'edit'} onOpenChange={(v) => !v && setOpen(null)}>
      <DialogContent className='sm:max-w-[425px]'>
        <DialogHeader>
          <DialogTitle>{isEdit ? '编辑分类' : '新增分类'}</DialogTitle>
          <DialogDescription>
            {isEdit ? '在此修改分类信息。' : '在此填写新分类的信息。'}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4'>
            <FormField
              control={form.control}
              name='category'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>分类名称</FormLabel>
                  <FormControl>
                    <Input placeholder='请输入分类名称' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='slug'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Slug</FormLabel>
                  <FormControl>
                    <Input placeholder='请输入 Slug' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='icon'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>图标 / Emoji</FormLabel>
                  <FormControl>
                    <Input placeholder='请输入图标名或 Emoji' {...field} />
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
                    <div className='space-y-3'>
                      <div className='flex items-center gap-2'>
                        <Popover>
                          <PopoverTrigger asChild>
                            <button
                              type='button'
                              className='h-8 w-8 rounded border border-black/10 transition-transform hover:scale-110 active:scale-95'
                              style={{ backgroundColor: field.value || 'transparent' }}
                            />
                          </PopoverTrigger>
                          <PopoverContent className='w-auto p-3'>
                            <div className='space-y-3'>
                              <div className='text-xs font-medium text-muted-foreground'>选择预设颜色</div>
                              <div className='grid grid-cols-6 gap-2'>
                                {PRESET_COLORS.map((color) => (
                                  <button
                                    key={color}
                                    type='button'
                                    className={cn(
                                      'h-6 w-6 rounded-full border border-black/10 transition-transform hover:scale-110 active:scale-95',
                                      field.value === color && 'ring-2 ring-primary ring-offset-2'
                                    )}
                                    style={{ backgroundColor: color }}
                                    onClick={() => field.onChange(color)}
                                  />
                                ))}
                                <button
                                  type='button'
                                  className={cn(
                                    'h-6 w-6 rounded-full border border-black/10 bg-white text-[10px] font-medium transition-transform hover:scale-110 active:scale-95',
                                    !field.value && 'ring-2 ring-primary ring-offset-2'
                                  )}
                                  onClick={() => field.onChange('')}
                                >
                                  无
                                </button>
                              </div>
                              <div className='pt-2 border-t'>
                                <div className='text-xs font-medium text-muted-foreground mb-2'>自定义颜色</div>
                                <div className='flex items-center gap-2'>
                                  <div className='relative h-8 w-8 overflow-hidden rounded border border-black/10'>
                                    <Input
                                      type='color'
                                      value={field.value || '#000000'}
                                      onChange={(e) => field.onChange(e.target.value)}
                                      className='absolute -inset-2 h-12 w-12 cursor-pointer border-none bg-transparent p-0'
                                    />
                                  </div>
                                  <Input
                                    placeholder='#000000'
                                    value={field.value || ''}
                                    className='h-8 w-24 font-mono text-xs'
                                    onChange={(e) => field.onChange(e.target.value)}
                                  />
                                </div>
                              </div>
                            </div>
                          </PopoverContent>
                        </Popover>
                        <Input
                          placeholder='未设置颜色'
                          value={field.value || ''}
                          className='h-8 flex-1 font-mono text-xs'
                          onChange={(e) => field.onChange(e.target.value)}
                        />
                      </div>
                    </div>
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
                    <Input placeholder='请输入描述' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <Button type='submit' disabled={loading}>
                {loading && <span className='mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent' />}
                保存
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
