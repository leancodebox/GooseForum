import { useEffect, useState } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from '@tanstack/react-router'
// import { Link, useNavigate } from '@tanstack/react-router'
import { Loader2, LogIn, RefreshCcw } from 'lucide-react'
import { toast } from 'sonner'
import axios from 'axios'
// import { IconFacebook, IconGithub } from '@/assets/brand-icons'
import { useAuthStore } from '@/stores/auth-store'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { PasswordInput } from '@/components/password-input'
import { encryptLoginPassword } from '@/lib/login-crypto'

const formSchema = z.object({
  email: z.string().min(1, 'Please enter your email/username'),
  password: z
    .string()
    .min(1, 'Please enter your password')
    .min(6, 'Password must be at least 6 characters long'),
  captcha: z.string().min(1, 'Please enter captcha'),
})

interface UserAuthFormProps extends React.HTMLAttributes<HTMLFormElement> {
  redirectTo?: string
}

export function UserAuthForm({
  className,
  redirectTo,
  ...props
}: UserAuthFormProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [captchaData, setCaptchaData] = useState<{ id: string; img: string }>({
    id: '',
    img: '',
  })
  const navigate = useNavigate()
  const { auth } = useAuthStore()

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      email: '',
      password: '',
      captcha: '',
    },
  })

  const fetchCaptcha = async () => {
    try {
      const response = await axios.get('/api/get-captcha')
      if (response.data.code === 0) {
        setCaptchaData({
          id: response.data.result.captchaId,
          img: response.data.result.captchaImg,
        })
      }
    } catch (error) {
      toast.error('Failed to load captcha')
    }
  }

  useEffect(() => {
    fetchCaptcha()
  }, [])

  async function onSubmit(data: z.infer<typeof formSchema>) {
    setIsLoading(true)

    try {
      const encryptedPassword = await encryptLoginPassword(data.password)
      const response = await axios.post('/api/login', {
        username: data.email,
        encryptedPassword,
        captchaId: captchaData.id,
        captchaCode: data.captcha,
      })

      if (response.data.code === 0) {
        // 后端登录接口 Login 返回的是 component.SuccessData("登录成功")
        // 所以 response.data.result 是字符串 "登录成功"，而不是用户信息
        // 我们需要额外调用一次获取用户信息的接口
         try {
           const infoResponse = await axios.get('/api/get-user-info')
           if (infoResponse.data.code === 0) {
             const userData = infoResponse.data.result
            auth.setUser({
              accountNo: userData.userId.toString(),
              email: userData.email,
              role: userData.isAdmin ? ['admin'] : ['user'],
              exp: Date.now() + 24 * 60 * 60 * 1000,
            })
            toast.success(`Welcome back, ${userData.username}!`)
            const targetPath = redirectTo || '/'
            navigate({ to: targetPath, replace: true })
          } else {
            toast.error('Failed to fetch user info after login')
          }
        } catch (error) {
          toast.error('An error occurred while fetching user info')
        }
      } else {
        toast.error(response.data.msg || 'Login failed')
        fetchCaptcha() // 登录失败刷新验证码
        form.setValue('captcha', '') // 清空验证码输入
      }
    } catch (error) {
      toast.error('An error occurred during login')
      fetchCaptcha()
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn('grid gap-3', className)}
        {...props}
      >
        <FormField
          control={form.control}
          name='email'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email / Username</FormLabel>
              <FormControl>
                <Input placeholder='name@example.com' {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='password'
          render={({ field }) => (
            <FormItem className='relative'>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <PasswordInput placeholder='********' {...field} />
              </FormControl>
              <FormMessage />
              {/* <Link
                to='/forgot-password'
                className='absolute end-0 -top-0.5 text-sm font-medium text-muted-foreground hover:opacity-75'
              >
                Forgot password?
              </Link> */}
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='captcha'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Captcha</FormLabel>
              <div className='flex gap-2'>
                <FormControl>
                  <Input placeholder='Enter captcha' {...field} />
                </FormControl>
                <div
                  className='relative flex h-9 w-32 shrink-0 cursor-pointer items-center justify-center overflow-hidden rounded-md border bg-muted'
                  onClick={fetchCaptcha}
                  title='Click to refresh captcha'
                >
                  {captchaData.img ? (
                    <img
                      src={captchaData.img}
                      alt='captcha'
                      className='h-full w-full object-contain'
                    />
                  ) : (
                    <RefreshCcw className='h-4 w-4 animate-spin text-muted-foreground' />
                  )}
                </div>
              </div>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button className='mt-2' disabled={isLoading}>
          {isLoading ? <Loader2 className='animate-spin' /> : <LogIn />}
          Sign in
        </Button>

        {/* <div className='relative my-2'>
          <div className='absolute inset-0 flex items-center'>
            <span className='w-full border-t' />
          </div>
          <div className='relative flex justify-center text-xs uppercase'>
            <span className='bg-background px-2 text-muted-foreground'>
              Or continue with
            </span>
          </div>
        </div>

        <div className='grid grid-cols-2 gap-2'>
          <Button variant='outline' type='button' disabled={isLoading}>
            <IconGithub className='h-4 w-4' /> GitHub
          </Button>
          <Button variant='outline' type='button' disabled={isLoading}>
            <IconFacebook className='h-4 w-4' /> Facebook
          </Button>
        </div> */}
      </form>
    </Form>
  )
}
