import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2, Mail, Server, Shield, Send, Info, Save } from 'lucide-react'
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
import { mailSettingsSchema, type MailSettings } from './data/schema'

export default function MailSettingsManagement() {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [testEmail, setTestEmail] = useState('')

  const form = useForm<MailSettings>({
    resolver: zodResolver(mailSettingsSchema),
    defaultValues: {
      enableMail: true,
      smtpHost: '',
      smtpPort: 587,
      useSSL: false,
      smtpUsername: '',
      smtpPassword: '',
      fromName: '',
      fromEmail: '',
    },
  })

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const response = await axios.get('/api/admin/mail-settings')
        if (response.data.code === 0) {
          form.reset(response.data.result)
        }
      } catch (error) {
        toast.error('加载邮件设置失败')
      } finally {
        setLoading(false)
      }
    }
    fetchSettings()
  }, [form])

  const onSubmit = async (data: MailSettings) => {
    setSaving(true)
    try {
      const response = await axios.post('/api/admin/save-mail-settings', {
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

  const handleTestConnection = async () => {
    if (!testEmail) {
      toast.error('请输入测试邮箱')
      return
    }

    setTesting(true)
    try {
      const response = await axios.post('/api/admin/test-mail-connection', {
        settings: form.getValues(),
        testEmail: testEmail,
      })
      if (response.data.code === 0 && response.data.result.success) {
        toast.success(response.data.result.message || '测试邮件发送成功')
      } else {
        toast.error(response.data.result?.message || response.data.msg || '测试邮件发送失败')
      }
    } catch (error) {
      toast.error('发送测试邮件失败，请检查配置')
    } finally {
      setTesting(false)
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
    <>
      <ContentLayout
        title='邮件设置'
        description='配置 SMTP 服务器以发送验证邮件、通知等。'
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
              <CardContent className='space-y-10'>
                <div className='grid gap-10 md:grid-cols-2 lg:grid-cols-3'>
                  {/* 左侧：SMTP 服务器设置 */}
                  <div className='space-y-6 lg:col-span-2'>
                    <div className='flex items-center gap-2 border-b pb-2 text-lg font-medium'>
                      <Server className='h-5 w-5 text-muted-foreground' />
                      SMTP 服务器设置
                    </div>
                    <div className='space-y-4'>
                      <FormField
                        control={form.control}
                        name='enableMail'
                        render={({ field }) => (
                          <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4 bg-muted/20'>
                            <div className='space-y-0.5'>
                              <FormLabel className='text-base'>启用邮件服务</FormLabel>
                              <FormDescription>
                                开启后，系统将能够发送验证码、通知等邮件。
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

                      <div className='grid gap-6 md:grid-cols-2'>
                        <FormField
                          control={form.control}
                          name='smtpHost'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>SMTP 主机</FormLabel>
                              <FormControl>
                                <Input placeholder='smtp.example.com' {...field} disabled={!form.watch('enableMail')} />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                        <FormField
                          control={form.control}
                          name='smtpPort'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>SMTP 端口</FormLabel>
                              <FormControl>
                                <Input
                                  type='number'
                                  placeholder='587'
                                  {...field}
                                  onChange={(e) => field.onChange(Number(e.target.value))}
                                  disabled={!form.watch('enableMail')}
                                />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                        <FormField
                          control={form.control}
                          name='smtpUsername'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>用户名 (邮箱)</FormLabel>
                              <FormControl>
                                <Input placeholder='user@example.com' {...field} disabled={!form.watch('enableMail')} />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                        <FormField
                          control={form.control}
                          name='smtpPassword'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>密码 / 授权码</FormLabel>
                              <FormControl>
                                <Input type='password' placeholder='••••••••' {...field} disabled={!form.watch('enableMail')} />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      </div>

                      <FormField
                        control={form.control}
                        name='useSSL'
                        render={({ field }) => (
                          <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4 bg-muted/20'>
                            <div className='space-y-0.5'>
                              <FormLabel className='text-base flex items-center gap-2'>
                                <Shield className='h-4 w-4' />
                                使用 SSL 加密
                              </FormLabel>
                              <FormDescription>
                                通常端口 465 需要开启 SSL，587 通常使用 STARTTLS。
                              </FormDescription>
                            </div>
                            <FormControl>
                              <Switch
                                checked={field.value}
                                onCheckedChange={field.onChange}
                                disabled={!form.watch('enableMail')}
                              />
                            </FormControl>
                          </FormItem>
                        )}
                      />
                    </div>
                  </div>

                  {/* 右侧：发件人信息 & 测试邮件 */}
                  <div className='space-y-10 lg:col-span-1'>
                    {/* 发件人信息 */}
                    <div className='space-y-6'>
                      <div className='flex items-center gap-2 border-b pb-2 text-lg font-medium'>
                        <Mail className='h-5 w-5 text-muted-foreground' />
                        发件人信息
                      </div>
                      <div className='space-y-4'>
                        <FormField
                          control={form.control}
                          name='fromName'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>发件人名称</FormLabel>
                              <FormControl>
                                <Input placeholder='GooseForum' {...field} disabled={!form.watch('enableMail')} />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                        <FormField
                          control={form.control}
                          name='fromEmail'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>发件人邮箱</FormLabel>
                              <FormControl>
                                <Input placeholder='noreply@example.com' {...field} disabled={!form.watch('enableMail')} />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      </div>
                    </div>

                    {/* 测试邮件 */}
                    <div className='space-y-6'>
                      <div className='flex items-center gap-2 border-b pb-2 text-lg font-medium'>
                        <Send className='h-5 w-5 text-muted-foreground' />
                        发送测试邮件
                      </div>
                      <div className='space-y-6'>
                        <div className='space-y-2'>
                          <FormLabel>测试收件箱</FormLabel>
                          <div className='flex gap-2'>
                            <Input
                              type='email'
                              placeholder='test@example.com'
                              value={testEmail}
                              onChange={(e) => setTestEmail(e.target.value)}
                            />
                            <Button
                              type='button'
                              variant='secondary'
                              onClick={handleTestConnection}
                              disabled={testing || !testEmail}
                            >
                              {testing ? <Loader2 className='h-4 w-4 animate-spin' /> : <Send className='h-4 w-4' />}
                              <span className='ml-2'>测试</span>
                            </Button>
                          </div>
                        </div>
                        <div className='rounded-lg border bg-muted/30 p-4 space-y-3'>
                          <h4 className='text-sm font-medium flex items-center gap-2'>
                            <Info className='h-4 w-4 text-blue-500 shrink-0' />
                            常用配置参考
                          </h4>
                          <div className='flex flex-col gap-2 text-xs text-muted-foreground'>
                            <div className='flex items-start justify-between gap-4'>
                              <span className='shrink-0 mt-[2px]'>QQ 邮箱:</span>
                              <span className='font-mono break-all text-right'>smtp.qq.com:587</span>
                            </div>
                            <div className='flex items-start justify-between gap-4'>
                              <span className='shrink-0 mt-[2px]'>163 邮箱:</span>
                              <span className='font-mono break-all text-right'>smtp.163.com:25</span>
                            </div>
                            <div className='flex items-start justify-between gap-4'>
                              <span className='shrink-0 mt-[2px]'>Gmail:</span>
                              <span className='font-mono break-all text-right'>smtp.gmail.com:587</span>
                            </div>
                            <div className='flex items-start justify-between gap-4'>
                              <span className='shrink-0 mt-[2px]'>Outlook:</span>
                              <span className='font-mono break-all text-right'>smtp.office365.com:587</span>
                            </div>
                          </div>
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
    </>
  )
}
