import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2, MailCheck, Save, Trash2, Plus } from 'lucide-react'
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
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { securitySettingsSchema, type SecuritySettings } from './data/schema'

export default function SecuritySettingsManagement() {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [newAllowedDomain, setNewAllowedDomain] = useState('')

  const form = useForm<SecuritySettings>({
    resolver: zodResolver(securitySettingsSchema),
    defaultValues: {
      enableSignup: true,
      enableEmailVerification: false,
      allowedDomains: [],
    },
  })

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const response = await axios.get('/api/admin/security-settings')
        if (response.data.code === 0) {
          const result = response.data.result
          // 确保 allowedDomains 始终为数组，防止 Zod 校验失败
          if (result && !Array.isArray(result.allowedDomains)) {
            result.allowedDomains = []
          }
          form.reset(result)
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

  const addAllowedDomain = () => {
    const domain = newAllowedDomain.trim().toLowerCase()
    if (!domain) return

    const currentDomains = form.getValues('allowedDomains')
    if (currentDomains.includes(domain)) {
      toast.error('域名已存在')
      return
    }

    form.setValue('allowedDomains', [...currentDomains, domain], {
      shouldDirty: true,
      shouldValidate: true,
    })
    setNewAllowedDomain('')
  }

  const removeAllowedDomain = (domain: string) => {
    const currentDomains = form.getValues('allowedDomains')
    form.setValue(
      'allowedDomains',
      currentDomains.filter((d: string) => d !== domain),
      {
        shouldDirty: true,
        shouldValidate: true,
      }
    )
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
      description='控制用户注册、邮箱验证以及允许的注册邮箱域名。'
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
                  name='enableSignup'
                  render={({ field }) => (
                    <FormItem className='flex flex-row items-center justify-between'>
                      <div className='space-y-0.5'>
                        <FormLabel className='text-base font-medium'>允许新用户注册</FormLabel>
                        <FormDescription>
                          关闭后，前台将不允许创建新账号。
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
                  name='enableEmailVerification'
                  render={({ field }) => (
                    <FormItem className='flex flex-row items-center justify-between'>
                      <div className='space-y-0.5'>
                        <FormLabel className='flex items-center gap-2 text-base font-medium'>
                          <MailCheck className='h-4 w-4 text-muted-foreground' />
                          要求验证邮箱
                        </FormLabel>
                        <FormDescription>
                          新用户完成邮箱验证后才能正常激活和使用账号。
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

                <div className='space-y-4'>
                  <div className='space-y-0.5'>
                    <FormLabel className='text-base font-medium'>允许注册的邮箱域名</FormLabel>
                    <FormDescription>
                      留空表示不限制。配置后，仅允许这些邮箱域名完成注册。
                    </FormDescription>
                  </div>

                  <div className='flex gap-2'>
                    <Input
                      placeholder='例如: gmail.com'
                      value={newAllowedDomain}
                      onChange={(e) => setNewAllowedDomain(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          e.preventDefault()
                          addAllowedDomain()
                        }
                      }}
                      className='max-w-sm'
                    />
                    <Button
                      type='button'
                      variant='secondary'
                      onClick={addAllowedDomain}
                    >
                      <Plus className='mr-2 h-4 w-4' /> 添加
                    </Button>
                  </div>

                  <div className='flex flex-wrap gap-2'>
                      {(() => {
                        const domains = form.watch('allowedDomains')
                        if (!Array.isArray(domains) || domains.length === 0) {
                          return <span className='text-sm text-muted-foreground italic'>未限制域名</span>
                        }
                        return domains.map((domain: string) => (
                          <Badge key={domain} variant='secondary' className='px-3 py-1.5 text-sm font-normal'>
                          {domain}
                          <button
                            type='button'
                            className='ml-2 rounded-full outline-none ring-offset-background focus:ring-2 focus:ring-ring focus:ring-offset-2'
                            onClick={() => removeAllowedDomain(domain)}
                          >
                            <Trash2 className='h-3.5 w-3.5 text-muted-foreground hover:text-destructive transition-colors' />
                          </button>
                        </Badge>
                        ))
                      })()}
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
