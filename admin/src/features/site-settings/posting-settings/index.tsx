import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2, FileText, Upload, Edit3, Save, Plus, Trash2 } from 'lucide-react'
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
import { Card, CardContent } from '@/components/ui/card'
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
        allowUppercasePosts: true,
      },
      uploadControl: {
        allowAttachments: true,
        authorizedExtensions: ['.jpg', '.jpeg', '.png', '.gif', '.webp'],
        maxAttachmentSizeKb: 5120,
        maxAttachmentsPerPost: 10,
      },
      editControl: {
        editingGracePeriod: 300,
        postEditTimeLimit: 86400,
        allowUsersToDeletePosts: true,
      },
    },
  })

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const response = await axios.get('/api/admin/posting-settings')
        if (response.data.code === 0) {
          form.reset(response.data.result)
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
      headerActions={
        <Button onClick={form.handleSubmit(onSubmit)} disabled={saving}>
          {saving ? (
            <Loader2 className='mr-2 h-4 w-4 animate-spin' />
          ) : (
            <Save className='mr-2 h-4 w-4' />
          )}
          保存配置
        </Button>
      }
    >
        <div className='w-full'>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-6'>
              <Card>
                <CardContent className='space-y-10 pt-6'>
                  <div className='grid gap-10 md:grid-cols-2 lg:grid-cols-3'>
                    
                    {/* 左侧区域：文本控制 & 编辑控制 (占 2/3 宽度) */}
                    <div className='space-y-10 lg:col-span-2'>
                      
                      {/* 文本控制 */}
                      <div className='space-y-6'>
                        <div className='flex items-center gap-2 border-b pb-2 text-lg font-medium'>
                          <FileText className='h-5 w-5 text-muted-foreground' />
                          文本内容控制
                        </div>
                        <div className='space-y-4'>
                          <div className='grid gap-6 md:grid-cols-2'>
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
                          </div>
                          <FormField
                            control={form.control}
                            name='textControl.allowUppercasePosts'
                            render={({ field }) => (
                              <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4 bg-muted/20'>
                                <div className='space-y-0.5'>
                                  <FormLabel className='text-base'>允许全大写内容</FormLabel>
                                  <FormDescription>是否允许用户发布全部由大写字母组成的标题或内容。</FormDescription>
                                </div>
                                <FormControl>
                                  <Switch checked={field.value} onCheckedChange={field.onChange} />
                                </FormControl>
                              </FormItem>
                            )}
                          />
                        </div>
                      </div>

                      {/* 编辑控制 */}
                      <div className='space-y-6'>
                        <div className='flex items-center gap-2 border-b pb-2 text-lg font-medium'>
                          <Edit3 className='h-5 w-5 text-muted-foreground' />
                          编辑与删除控制
                        </div>
                        <div className='space-y-4'>
                          <div className='grid gap-6 md:grid-cols-2'>
                            <FormField
                              control={form.control}
                              name='editControl.editingGracePeriod'
                              render={({ field }) => (
                                <FormItem>
                                  <FormLabel>编辑宽限期 (秒)</FormLabel>
                                  <FormControl>
                                    <Input type='number' {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                  </FormControl>
                                  <FormDescription>发布后多少秒内编辑不显示“已编辑”标记。</FormDescription>
                                  <FormMessage />
                                </FormItem>
                              )}
                            />
                            <FormField
                              control={form.control}
                              name='editControl.postEditTimeLimit'
                              render={({ field }) => (
                                <FormItem>
                                  <FormLabel>可编辑时限 (秒)</FormLabel>
                                  <FormControl>
                                    <Input type='number' {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                  </FormControl>
                                  <FormDescription>发布后多少秒内允许用户编辑。0 为不限制。</FormDescription>
                                  <FormMessage />
                                </FormItem>
                              )}
                            />
                          </div>
                          <FormField
                            control={form.control}
                            name='editControl.allowUsersToDeletePosts'
                            render={({ field }) => (
                              <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4 bg-muted/20'>
                                <div className='space-y-0.5'>
                                  <FormLabel className='text-base'>允许用户删除帖子</FormLabel>
                                  <FormDescription>是否允许普通用户自行删除已发布的帖子。</FormDescription>
                                </div>
                                <FormControl>
                                  <Switch checked={field.value} onCheckedChange={field.onChange} />
                                </FormControl>
                              </FormItem>
                            )}
                          />
                        </div>
                      </div>

                    </div>

                    {/* 右侧区域：上传控制 (占 1/3 宽度) */}
                    <div className='space-y-6 lg:col-span-1'>
                      <div className='flex items-center gap-2 border-b pb-2 text-lg font-medium'>
                        <Upload className='h-5 w-5 text-muted-foreground' />
                        上传控制
                      </div>
                      <div className='space-y-6'>
                        <FormField
                          control={form.control}
                          name='uploadControl.allowAttachments'
                          render={({ field }) => (
                            <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4 bg-muted/20'>
                              <div className='space-y-0.5'>
                                <FormLabel className='text-base'>允许上传附件</FormLabel>
                                <FormDescription>开启后用户可以在帖子中上传文件。</FormDescription>
                              </div>
                              <FormControl>
                                <Switch checked={field.value} onCheckedChange={field.onChange} />
                              </FormControl>
                            </FormItem>
                          )}
                        />
                        <div className='space-y-4'>
                          <FormField
                            control={form.control}
                            name='uploadControl.maxAttachmentSizeKb'
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>最大附件大小 (KB)</FormLabel>
                                <FormControl>
                                  <Input type='number' {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <FormField
                            control={form.control}
                            name='uploadControl.maxAttachmentsPerPost'
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>单帖最大附件数</FormLabel>
                                <FormControl>
                                  <Input type='number' {...field} onChange={e => field.onChange(Number(e.target.value))} />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                        </div>
                        <div className='space-y-4'>
                          <FormLabel>允许的文件扩展名</FormLabel>
                          <div className='flex gap-2'>
                            <Input
                              placeholder='.pdf'
                              value={newExtension}
                              onChange={(e) => setNewExtension(e.target.value)}
                            />
                            <Button type='button' variant='secondary' size='icon' className='shrink-0' onClick={addExtension}>
                              <Plus className='h-4 w-4' />
                            </Button>
                          </div>
                          <div className='flex flex-wrap gap-2'>
                            {form.watch('uploadControl.authorizedExtensions').map((ext: string) => (
                              <Badge key={ext} variant='secondary' className='pl-2 pr-1 py-1 text-xs'>
                                {ext}
                                <Button
                                  type='button'
                                  variant='ghost'
                                  size='icon'
                                  className='h-4 w-4 ml-1 hover:bg-transparent'
                                  onClick={() => removeExtension(ext)}
                                >
                                  <Trash2 className='h-3 w-3 text-destructive' />
                                </Button>
                              </Badge>
                            ))}
                            {form.watch('uploadControl.authorizedExtensions').length === 0 && (
                              <span className='text-xs text-muted-foreground'>未配置扩展名</span>
                            )}
                          </div>
                        </div>
                      </div>
                    </div>

                  </div>
                </CardContent>
              </Card>
            </form>
          </Form>
        </div>
      </ContentLayout>
    )
  }
