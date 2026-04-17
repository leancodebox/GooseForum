import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2, FileText, Upload, Save, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import axios from 'axios'
import { ContentLayout } from '@/components/layout/content-layout'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { postingSettingsSchema, type PostingSettings } from './data/schema'

export default function PostingSettingsManagement() {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [newExtension, setNewExtension] = useState('')

  const form = useForm<PostingSettings>({
    resolver: zodResolver(postingSettingsSchema),
    defaultValues: {
      textControl: {
        minPostLength: 5,
        maxPostLength: 50000,
        minTitleLength: 5,
        maxTitleLength: 100,
        newUserPostCooldownMinutes: 0,
      },
      uploadControl: {
        allowAttachments: true,
        authorizedExtensions: ['.jpg', '.jpeg', '.png', '.gif', '.webp'],
        maxAttachmentSizeKb: 5120,
        maxDailyUploadsPerUser: 10,
        newUserUploadCooldownMinutes: 0,
      },
    },
  })

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const response = await axios.get('/api/admin/posting-settings')
        if (response.data.code === 0) {
          const result = response.data.result
          // 确保 authorizedExtensions 始终为数组，防止 Zod 校验失败
          if (result && result.uploadControl && result.uploadControl.authorizedExtensions === null) {
            result.uploadControl.authorizedExtensions = []
          }
          form.reset(result)
        }
      } catch (error) {
        toast.error('加载发布设置失败')
      } finally {
        setLoading(false)
      }
    }
    fetchSettings()
  }, [form])

  const onSubmit = async (data: PostingSettings) => {
    setSaving(true)
    try {
      const response = await axios.post('/api/admin/save-posting-settings', {
        settings: data,
      })
      if (response.data.code === 0) {
        toast.success('保存成功')
      } else {
        toast.error(response.data.msg || '保存失败')
      }
    } catch (error) {
      toast.error('保存失败')
    } finally {
      setSaving(false)
    }
  }

  const addExtension = () => {
    if (!newExtension) return
    if (!newExtension.startsWith('.')) {
      toast.error('扩展名必须以 . 开头')
      return
    }
    
    const currentExtensions = form.getValues('uploadControl.authorizedExtensions')
    if (currentExtensions.includes(newExtension)) {
      toast.error('该扩展名已存在')
      return
    }

    form.setValue('uploadControl.authorizedExtensions', [...currentExtensions, newExtension])
    setNewExtension('')
  }

  const removeExtension = (ext: string) => {
    const currentExtensions = form.getValues('uploadControl.authorizedExtensions')
    form.setValue('uploadControl.authorizedExtensions', currentExtensions.filter((e: string) => e !== ext))
  }

  if (loading) {
    return (
      <div className='flex h-[400px] items-center justify-center'>
        <Loader2 className='h-8 w-8 animate-spin text-primary' />
      </div>
    )
  }

  return (
    <ContentLayout 
      title='发布内容设置'
      description='控制帖子标题、内容长度及附件上传规则。'
      showSeparator={true}
      headerActions={
        <Button 
          onClick={form.handleSubmit(onSubmit, (errors) => {
            console.error('Form validation errors:', errors)
            toast.error('请检查表单填写是否正确')
          })} 
          disabled={saving}
        >
          {saving ? (
            <Loader2 className='mr-2 h-4 w-4 animate-spin' />
          ) : (
            <Save className='mr-2 h-4 w-4' />
          )}
          保存配置
        </Button>
      }
    >
      <div className='flex flex-1 flex-col'>
        <div className='faded-bottom h-full w-full overflow-y-auto scroll-smooth pe-4 pb-12'>
          <div className='-mx-1 px-1.5'>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit, (errors) => {
                console.error('Form validation errors:', errors)
                toast.error('请检查表单填写是否正确')
              })} className='space-y-12'>
                
                <div className='grid gap-12 lg:grid-cols-2'>
                  {/* 左侧区域：文本内容控制 */}
                  <div className='space-y-6'>
                    <div className='flex items-center gap-2 border-b pb-2 text-lg font-medium'>
                      <FileText className='h-5 w-5 text-muted-foreground' />
                      文本内容控制
                    </div>
                        <div className='grid gap-6 sm:grid-cols-2'>
                          <FormField
                            control={form.control}
                            name='textControl.minTitleLength'
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>最小标题长度</FormLabel>
                                <FormControl>
                                  <Input type='number' {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <FormField
                            control={form.control}
                            name='textControl.maxTitleLength'
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>最大标题长度</FormLabel>
                                <FormControl>
                                  <Input type='number' {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <FormField
                            control={form.control}
                            name='textControl.minPostLength'
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>最小正文长度</FormLabel>
                                <FormControl>
                                  <Input type='number' {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <FormField
                            control={form.control}
                            name='textControl.maxPostLength'
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>最大正文长度</FormLabel>
                                <FormControl>
                                  <Input type='number' {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <FormField
                            control={form.control}
                            name='textControl.newUserPostCooldownMinutes'
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>新用户发帖冷却 (分钟)</FormLabel>
                                <FormControl>
                                  <Input type='number' {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                </FormControl>
                                <FormDescription>注册后需等待多久才能发布第一个帖子。</FormDescription>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                        </div>
                  </div>

                  {/* 右侧区域：附件控制 */}
                  <div className='space-y-6'>
                    <div className='flex items-center gap-2 border-b pb-2 text-lg font-medium'>
                      <Upload className='h-5 w-5 text-muted-foreground' />
                      附件控制
                    </div>
                    <div className='space-y-6'>
                      <FormField
                        control={form.control}
                        name='uploadControl.allowAttachments'
                        render={({ field }) => (
                          <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4 bg-muted/10'>
                            <div className='space-y-0.5'>
                              <FormLabel className='text-base font-medium'>允许上传附件</FormLabel>
                              <FormDescription>开启后，用户可以在帖子中上传图片或文件。</FormDescription>
                            </div>
                            <FormControl>
                              <Switch checked={field.value} onCheckedChange={field.onChange} />
                            </FormControl>
                          </FormItem>
                        )}
                      />
                      
                      <div className='grid gap-6 sm:grid-cols-2'>
                        <FormField
                          control={form.control}
                          name='uploadControl.maxDailyUploadsPerUser'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>单用户每日最大上传数量</FormLabel>
                              <FormControl>
                                <Input type='number' {...field} onChange={e => field.onChange(Number(e.target.value))} disabled={!form.watch('uploadControl.allowAttachments')} />
                              </FormControl>
                              <FormDescription className='min-h-[32px]'>限制每个用户每天上传附件总数。</FormDescription>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                        
                        <FormField
                          control={form.control}
                          name='uploadControl.maxAttachmentSizeKb'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>单个附件大小限制 (KB)</FormLabel>
                              <FormControl>
                                <Input type='number' {...field} onChange={e => field.onChange(Number(e.target.value))} disabled={!form.watch('uploadControl.allowAttachments')} />
                              </FormControl>
                              <FormDescription className='min-h-[32px]'>设置允许上传的单个文件最大体积。</FormDescription>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                        
                        <FormField
                          control={form.control}
                          name='uploadControl.newUserUploadCooldownMinutes'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>新用户上传冷却 (分钟)</FormLabel>
                              <FormControl>
                                <Input type='number' {...field} onChange={e => field.onChange(Number(e.target.value))} disabled={!form.watch('uploadControl.allowAttachments')} />
                              </FormControl>
                              <FormDescription className='min-h-[32px]'>注册后需等待多久才能上传附件。</FormDescription>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      </div>

                      <div className='space-y-4 pt-2'>
                        <FormLabel className='text-sm font-medium'>允许的扩展名</FormLabel>
                        <div className='flex gap-2'>
                          <Input
                            placeholder='例如: .jpg'
                            value={newExtension}
                            onChange={(e) => setNewExtension(e.target.value)}
                            onKeyDown={(e) => {
                              if (e.key === 'Enter') {
                                e.preventDefault()
                                addExtension()
                              }
                            }}
                            className='max-w-[150px]'
                            disabled={!form.watch('uploadControl.allowAttachments')}
                          />
                          <Button
                            type='button'
                            variant='secondary'
                            onClick={addExtension}
                            disabled={!form.watch('uploadControl.allowAttachments')}
                          >
                            <Plus className='mr-2 h-4 w-4' /> 添加
                          </Button>
                        </div>

                        <div className='flex flex-wrap gap-2'>
                          {(form.watch('uploadControl.authorizedExtensions') || []).length === 0 ? (
                            <span className='text-sm text-muted-foreground italic'>无限制或不允许</span>
                          ) : (
                            (form.watch('uploadControl.authorizedExtensions') || []).map((ext: string) => (
                              <Badge key={ext} variant='secondary' className='px-3 py-1.5 text-sm font-normal'>
                                {ext}
                                <button
                                  type='button'
                                  className='ml-2 rounded-full outline-none ring-offset-background focus:ring-2 focus:ring-ring focus:ring-offset-2'
                                  onClick={() => removeExtension(ext)}
                                  disabled={!form.watch('uploadControl.allowAttachments')}
                                >
                                  <Trash2 className='h-3.5 w-3.5 text-muted-foreground hover:text-destructive transition-colors' />
                                </button>
                              </Badge>
                            ))
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </form>
            </Form>
          </div>
        </div>
      </div>
    </ContentLayout>
  )
}

