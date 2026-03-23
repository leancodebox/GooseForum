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
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
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
        <div className='max-w-4xl'>
          <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-6'>
            <div className='grid gap-6 md:grid-cols-2'>
              {/* SMTP服务器设置 */}
              <Card className='md:col-span-2'>
                <CardHeader>
                  <CardTitle className='flex items-center gap-2'>
                    <Server className='h-5 w-5' />
                    SMTP 服务器设置
                  </CardTitle>
                  <CardDescription>配置邮件推送服务的服务器连接信息。</CardDescription>
                </CardHeader>
                <CardContent className='space-y-4'>
                  <FormField
                    control={form.control}
                    name='enableMail'
                    render={({ field }) => (
                      <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
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

                  <div className='grid gap-4 md:grid-cols-2'>
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
                      <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
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
                </CardContent>
              </Card>

              {/* 发件人信息 */}
              <Card>
                <CardHeader>
                  <CardTitle className='flex items-center gap-2'>
                    <Mail className='h-5 w-5' />
                    发件人信息
                  </CardTitle>
                  <CardDescription>设置邮件中显示的发件人名称和地址。</CardDescription>
                </CardHeader>
                <CardContent className='space-y-4'>
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
                </CardContent>
              </Card>

              {/* 测试邮件 */}
              <Card>
                <CardHeader>
                  <CardTitle className='flex items-center gap-2'>
                    <Send className='h-5 w-5' />
                    发送测试邮件
                  </CardTitle>
                  <CardDescription>在保存后，可以发送一封测试邮件来验证配置是否正确。</CardDescription>
                </CardHeader>
                <CardContent className='space-y-4'>
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
                        variant='outline'
                        onClick={handleTestConnection}
                        disabled={testing || !testEmail}
                      >
                        {testing ? <Loader2 className='h-4 w-4 animate-spin' /> : <Send className='h-4 w-4' />}
                        <span className='ml-2'>测试</span>
                      </Button>
                    </div>
                  </div>
                  <Separator />
                  <div className='space-y-2'>
                    <h4 className='text-sm font-medium flex items-center gap-2'>
                      <Info className='h-4 w-4' />
                      常用配置参考
                    </h4>
                    <div className='grid grid-cols-2 gap-x-4 gap-y-1 text-xs text-muted-foreground'>
                      <span>QQ 邮箱:</span> <span>smtp.qq.com:587</span>
                      <span>163 邮箱:</span> <span>smtp.163.com:25</span>
                      <span>Gmail:</span> <span>smtp.gmail.com:587</span>
                      <span>Outlook:</span> <span>smtp.office365.com:587</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </form>
        </Form>
      </div>
    </ContentLayout>
    </>
  )
}
