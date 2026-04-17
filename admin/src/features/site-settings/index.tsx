import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2, Image as ImageIcon, Globe, FileText, Code, PanelBottom } from 'lucide-react'
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
import { Textarea } from '@/components/ui/textarea'
import { siteInfoSchema, type SiteInfo } from './data/schema'
import { FooterEditor } from './components/footer-editor'

export default function SiteInfoManagement() {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const form = useForm<SiteInfo>({
    resolver: zodResolver(siteInfoSchema),
    defaultValues: {
      siteName: '',
      siteUrl: '',
      siteLogo: '',
      siteEmail: '',
      siteDescription: '',
      siteKeywords: '',
      externalLinks: '',
      footerInfo: {
        primary: [],
        list: [],
      },
    },
  })

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const response = await axios.get('/api/admin/site-settings')
        if (response.data.code === 0) {
          const result = response.data.result
          if (!result.footerInfo) {
            result.footerInfo = { primary: [], list: [] }
          } else {
            if (!result.footerInfo.primary) result.footerInfo.primary = []
            if (!result.footerInfo.list) result.footerInfo.list = []
          }
          form.reset(result)
        }
      } catch (error) {
        toast.error('加载站点设置失败')
        console.error(error)
      } finally {
        setLoading(false)
      }
    }
    fetchSettings()
  }, [form])

  const onSubmit = async (data: SiteInfo) => {
    setSaving(true)
    try {
      const response = await axios.post('/api/admin/save-site-settings', {
        settings: data,
      })
      if (response.data.code === 0) {
        toast.success('保存成功')
      } else {
        toast.error(response.data.message || '保存失败')
      }
    } catch (error) {
      toast.error('保存失败')
      console.error(error)
    } finally {
      setSaving(false)
    }
  }

  const handleLogoUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    const formData = new FormData()
    formData.append('file', file)

    try {
      const response = await axios.post('/file/img-upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
      if (response.data.code === 0) {
        form.setValue('siteLogo', response.data.result.url)
        toast.success('上传成功')
      } else {
        toast.error(response.data.message || '上传失败')
      }
    } catch (error) {
      toast.error('上传失败')
      console.error(error)
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
      title='站点信息'
      description='管理论坛的基础信息、SEO 设置和外部资源。'
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
            <Globe className='mr-2 h-4 w-4' />
          )}
          保存设置
        </Button>
      }
    >
      <div className='flex flex-1 flex-col'>
        <div className='faded-bottom h-full w-full overflow-y-auto scroll-smooth pe-4 pb-12'>
          <div className='-mx-1 px-1.5'>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-10'>
                {/* 采用左右分栏的更紧凑布局 */}
                <div className='grid gap-10 md:grid-cols-2'>
                  {/* 左侧：基本信息 & 联系方式 */}
                  <div className='space-y-6'>
                    <div className='flex items-center gap-2 border-b pb-2 text-lg font-medium'>
                      <Globe className='h-5 w-5 text-muted-foreground' />
                      基本信息
                    </div>
                    <div className='space-y-4'>
                      <FormField
                        control={form.control}
                        name='siteName'
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>站点名称</FormLabel>
                            <FormControl>
                              <Input placeholder='GooseForum' {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                      <FormField
                        control={form.control}
                        name='siteUrl'
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>站点 URL</FormLabel>
                            <FormControl>
                              <Input placeholder='https://example.com' {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                      <FormField
                        control={form.control}
                        name='siteLogo'
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>站点 Logo</FormLabel>
                            <div className='flex items-start gap-4'>
                              <div className='h-20 w-20 shrink-0 overflow-hidden rounded-lg border bg-muted flex items-center justify-center'>
                                {field.value ? (
                                  <img src={field.value} alt='Logo' className='h-full w-full object-cover' />
                                ) : (
                                  <ImageIcon className='h-10 w-10 text-muted-foreground/50' />
                                )}
                              </div>
                              <div className='flex flex-col gap-3 w-full'>
                                <FormControl>
                                  <Input placeholder='Logo URL' {...field} />
                                </FormControl>
                                <div className='flex items-center gap-2'>
                                  <Button
                                    type='button'
                                    variant='secondary'
                                    size='sm'
                                    className='w-full sm:w-auto'
                                    onClick={() => document.getElementById('logo-upload')?.click()}
                                  >
                                    上传图片
                                  </Button>
                                  <input
                                    id='logo-upload'
                                    type='file'
                                    accept='image/*'
                                    className='hidden'
                                    onChange={handleLogoUpload}
                                  />
                                </div>
                              </div>
                            </div>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                      <FormField
                        control={form.control}
                        name='siteEmail'
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>站点邮箱</FormLabel>
                            <FormControl>
                              <Input placeholder='admin@example.com' {...field} />
                            </FormControl>
                            <FormDescription>用于接收系统通知和用户反馈</FormDescription>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    </div>
                  </div>

                  {/* 右侧：SEO & 外部资源 */}
                  <div className='space-y-10'>
                    {/* SEO 设置 */}
                    <div className='space-y-6'>
                      <div className='flex items-center gap-2 border-b pb-2 text-lg font-medium'>
                        <FileText className='h-5 w-5 text-muted-foreground' />
                        SEO 设置
                      </div>
                      <div className='space-y-4'>
                        <FormField
                          control={form.control}
                          name='siteKeywords'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>关键词</FormLabel>
                              <FormControl>
                                <Input placeholder='forum, community, discussion' {...field} />
                              </FormControl>
                              <FormDescription>用逗号分隔多个关键词</FormDescription>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                        <FormField
                          control={form.control}
                          name='siteDescription'
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>站点描述</FormLabel>
                              <FormControl>
                                <Textarea
                                  placeholder='一个现代化的论坛系统'
                                  className='resize-none h-[64px]'
                                  {...field}
                                />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      </div>
                    </div>
                    {/* 外部资源区块 */}
                    <div className='space-y-6'>
                      <div className='flex items-center gap-2 border-b pb-2 text-lg font-medium'>
                        <Code className='h-5 w-5 text-muted-foreground' />
                        外部资源链接 / Meta 标签
                      </div>
                      <FormField
                        control={form.control}
                        name='externalLinks'
                        render={({ field }) => (
                          <FormItem>
                            <FormControl>
                              <Textarea
                                placeholder='<link rel="stylesheet" href="...">&#10;<script src="..."></script>'
                                className='min-h-[160px] font-mono text-sm'
                                {...field}
                              />
                            </FormControl>
                            <FormDescription>
                              每行输入一个完整的 HTML 标签。请确保代码片段安全。
                            </FormDescription>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    </div>
                  </div>
                </div>

                {/* Footer 页脚编辑区块 - 占据全宽 */}
                <div className='space-y-6 pt-4'>
                  <div className='flex items-center gap-2 border-b pb-2 text-lg font-medium'>
                    <PanelBottom className='h-5 w-5 text-muted-foreground' />
                    页脚内容 (Footer)
                  </div>
                  <div className='bg-background rounded-lg border p-4'>
                    <FooterEditor control={form.control} />
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
