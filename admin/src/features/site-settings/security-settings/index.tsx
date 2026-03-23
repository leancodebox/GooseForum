import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2, UserPlus, Lock, Globe, Save, Trash2, Plus } from 'lucide-react'
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
import { Badge } from '@/components/ui/badge'
import { securitySettingsSchema, type SecuritySettings } from './data/schema'

export default function SecuritySettingsManagement() {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [newAllowedDomain, setNewAllowedDomain] = useState('')
  const [newBlockedDomain, setNewBlockedDomain] = useState('')

  const form = useForm<SecuritySettings>({
    resolver: zodResolver(securitySettingsSchema),
    defaultValues: {
      enableSignup: true,
      enableEmailVerification: false,
      mustApproveUsers: false,
      minPasswordLength: 6,
      inviteOnly: false,
      restrictions: {
        allowedDomains: [],
        blockedDomains: [],
        maxRegistrationsPerIp: 10,
      },
    },
  })

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const response = await axios.get('/api/admin/security-settings')
        if (response.data.code === 0) {
          form.reset(response.data.result)
        }
      } catch (error) {
        toast.error('加载安全设置失败')
      } finally {
        setLoading(false)
      }
    }
    fetchSettings()
  }, [form])

  const onSubmit = async (data: SecuritySettings) => {
    setSaving(true)
    try {
      const response = await axios.post('/api/admin/save-security-settings', {
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

  const addDomain = (type: 'allowed' | 'blocked') => {
    const domain = type === 'allowed' ? newAllowedDomain : newBlockedDomain
    if (!domain) return
    
    const currentDomains = form.getValues(`restrictions.${type}Domains`)
    if (currentDomains.includes(domain)) {
      toast.error('域名已存在')
      return
    }

    form.setValue(`restrictions.${type}Domains`, [...currentDomains, domain])
    if (type === 'allowed') setNewAllowedDomain('')
    else setNewBlockedDomain('')
  }

  const removeDomain = (type: 'allowed' | 'blocked', domain: string) => {
    const currentDomains = form.getValues(`restrictions.${type}Domains`)
    form.setValue(`restrictions.${type}Domains`, currentDomains.filter((d: string) => d !== domain))
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
      title='安全与注册'
      description='管理用户注册流程、密码强度及访问限制。'
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
              {/* 注册设置 */}
              <Card className='md:col-span-2'>
                <CardHeader>
                  <CardTitle className='flex items-center gap-2'>
                    <UserPlus className='h-5 w-5' />
                    注册设置
                  </CardTitle>
                  <CardDescription>控制新用户的加入方式和验证流程。</CardDescription>
                </CardHeader>
                <CardContent className='space-y-4'>
                  <FormField
                    control={form.control}
                    name='enableSignup'
                    render={({ field }) => (
                      <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                        <div className='space-y-0.5'>
                          <FormLabel className='text-base'>允许新用户注册</FormLabel>
                          <FormDescription>
                            关闭后，普通用户将无法通过注册页面加入论坛。
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
                      name='enableEmailVerification'
                      render={({ field }) => (
                        <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                          <div className='space-y-0.5'>
                            <FormLabel className='text-base'>要求邮件验证</FormLabel>
                            <FormDescription>注册后必须通过邮件验证才能激活账号。</FormDescription>
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
                      name='mustApproveUsers'
                      render={({ field }) => (
                        <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                          <div className='space-y-0.5'>
                            <FormLabel className='text-base'>管理员人工审核</FormLabel>
                            <FormDescription>新注册用户必须经管理员手动审核后方可使用。</FormDescription>
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
                  </div>

                  <div className='grid gap-4 md:grid-cols-2'>
                    <FormField
                      control={form.control}
                      name='inviteOnly'
                      render={({ field }) => (
                        <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                          <div className='space-y-0.5'>
                            <FormLabel className='text-base'>仅限邀请注册</FormLabel>
                            <FormDescription>开启后，必须持有邀请码方可注册。</FormDescription>
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
                      name='minPasswordLength'
                      render={({ field }) => (
                        <FormItem className='rounded-lg border p-4'>
                          <div className='space-y-0.5 mb-2'>
                            <FormLabel className='text-base flex items-center gap-2'>
                              <Lock className='h-4 w-4' />
                              最小密码长度
                            </FormLabel>
                          </div>
                          <FormControl>
                            <Input
                              type='number'
                              {...field}
                              onChange={(e) => field.onChange(Number(e.target.value))}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                </CardContent>
              </Card>

              {/* 访问与 IP 限制 */}
              <Card className='md:col-span-2'>
                <CardHeader>
                  <CardTitle className='flex items-center gap-2'>
                    <Globe className='h-5 w-5' />
                    访问与限制
                  </CardTitle>
                  <CardDescription>配置域名黑白名单及 IP 注册限制。</CardDescription>
                </CardHeader>
                <CardContent className='space-y-6'>
                  <FormField
                    control={form.control}
                    name='restrictions.maxRegistrationsPerIp'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>单 IP 最大注册数</FormLabel>
                        <FormControl>
                          <Input
                            type='number'
                            {...field}
                            onChange={(e) => field.onChange(Number(e.target.value))}
                          />
                        </FormControl>
                        <FormDescription>24 小时内同一 IP 允许的最大注册用户数，0 为不限制。</FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <div className='grid gap-6 md:grid-cols-2'>
                    <div className='space-y-4'>
                      <FormLabel>允许注册的邮箱域名</FormLabel>
                      <div className='flex gap-2'>
                        <Input
                          placeholder='example.com'
                          value={newAllowedDomain}
                          onChange={(e) => setNewAllowedDomain(e.target.value)}
                        />
                        <Button type='button' variant='outline' size='icon' onClick={() => addDomain('allowed')}>
                          <Plus className='h-4 w-4' />
                        </Button>
                      </div>
                      <div className='flex flex-wrap gap-2'>
                        {form.watch('restrictions.allowedDomains').map((domain: string) => (
                          <Badge key={domain} variant='secondary' className='pl-2 pr-1 py-1'>
                            {domain}
                            <Button
                              type='button'
                              variant='ghost'
                              size='icon'
                              className='h-4 w-4 ml-1 hover:bg-transparent'
                              onClick={() => removeDomain('allowed', domain)}
                            >
                              <Trash2 className='h-3 w-3 text-destructive' />
                            </Button>
                          </Badge>
                        ))}
                      </div>
                    </div>

                    <div className='space-y-4'>
                      <FormLabel>禁止注册的邮箱域名</FormLabel>
                      <div className='flex gap-2'>
                        <Input
                          placeholder='spam.com'
                          value={newBlockedDomain}
                          onChange={(e) => setNewBlockedDomain(e.target.value)}
                        />
                        <Button type='button' variant='outline' size='icon' onClick={() => addDomain('blocked')}>
                          <Plus className='h-4 w-4' />
                        </Button>
                      </div>
                      <div className='flex flex-wrap gap-2'>
                        {form.watch('restrictions.blockedDomains').map((domain: string) => (
                          <Badge key={domain} variant='secondary' className='pl-2 pr-1 py-1'>
                            {domain}
                            <Button
                              type='button'
                              variant='ghost'
                              size='icon'
                              className='h-4 w-4 ml-1 hover:bg-transparent'
                              onClick={() => removeDomain('blocked', domain)}
                            >
                              <Trash2 className='h-3 w-3 text-destructive' />
                            </Button>
                          </Badge>
                        ))}
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </form>
        </Form>
      </div>
    </ContentLayout>
  )
}
