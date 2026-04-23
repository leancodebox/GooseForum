import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2, Save, Megaphone } from 'lucide-react'
import { toast } from 'sonner'
import { ContentLayout } from '@/components/layout/content-layout'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { getAnnouncement, saveAnnouncement } from '@/api'
import { announcementSettingsSchema, type AnnouncementSettings } from './data/schema'

export default function AnnouncementSettingsManagement() {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const form = useForm<AnnouncementSettings>({
    resolver: zodResolver(announcementSettingsSchema),
    defaultValues: {
      enabled: false,
      title: '',
      content: '',
      link: '',
    },
  })

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const response = await getAnnouncement()
        if (response.code === 0) {
          const result = response.result
          form.reset({
            enabled: result.enabled ?? false,
            title: result.title ?? '',
            content: result.content ?? '',
            link: result.link ?? '',
          })
        }
      } catch (error) {
        toast.error('加载公告设置失败')
      } finally {
        setLoading(false)
      }
    }
    fetchSettings()
  }, [form])

  const onSubmit = async (data: AnnouncementSettings) => {
    setSaving(true)
    try {
      const response = await saveAnnouncement({ settings: data })
      if (response.code === 0) {
        toast.success('保存成功')
      } else {
        toast.error(response.msg || '保存失败')
      }
    } catch (error) {
      toast.error('保存失败')
    } finally {
      setSaving(false)
    }
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
      title='系统公告'
      description='配置显示在页面顶部的系统公告横幅。'
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
          <div className='-mx-1 px-1.5 lg:max-w-2xl'>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-8'>
                <FormField
                  control={form.control}
                  name='enabled'
                  render={({ field }) => (
                    <FormItem className='flex flex-row items-center justify-between'>
                      <div className='space-y-0.5'>
                        <FormLabel className='text-base font-medium'>启用公告</FormLabel>
                        <FormDescription>
                          开启后，系统公告将显示在页面顶部。
                        </FormDescription>
                      </div>
                      <FormControl>
                        <Switch
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name='title'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className='text-base font-medium'>公告标题</FormLabel>
                      <FormDescription>
                        输入公告的标题文本。
                      </FormDescription>
                      <FormControl>
                        <Input
                          placeholder='例如：网站维护通知'
                          {...field}
                          className='max-w-sm'
                        />
                      </FormControl>
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name='content'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className='text-base font-medium'>公告内容</FormLabel>
                      <FormDescription>
                        输入公告的详细内容，支持多行文本。
                      </FormDescription>
                      <FormControl>
                        <Textarea
                          placeholder='例如：网站将于今晚22:00-24:00进行系统维护，届时可能无法访问。'
                          {...field}
                          rows={4}
                          className='max-w-sm resize-none'
                        />
                      </FormControl>
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name='link'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className='text-base font-medium flex items-center gap-2'>
                        <Megaphone className='h-4 w-4 text-muted-foreground' />
                        公告链接（可选）
                      </FormLabel>
                      <FormDescription>
                        输入点击公告后跳转的链接，留空则不可点击。
                      </FormDescription>
                      <FormControl>
                        <Input
                          placeholder='例如：https://example.com/announcement'
                          {...field}
                          className='max-w-sm'
                        />
                      </FormControl>
                    </FormItem>
                  )}
                />
              </form>
            </Form>
          </div>
        </div>
      </div>
    </ContentLayout>
  )
}