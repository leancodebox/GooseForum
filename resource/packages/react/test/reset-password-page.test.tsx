import { afterEach, describe, expect, it, vi } from 'vitest'
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { GooseSiteApi, LayoutPayload, ResetPasswordPageProps } from '@gooseforum/client'
import { GooseI18nProvider } from '../src/i18n'
import { GooseRuntimeProvider, type GooseRuntime } from '../src/runtime'
import { ResetPasswordPageView } from '../src/site/auth/reset-password-page'

afterEach(cleanup)

const layout = {
  site: {
    name: 'GooseForum',
    description: '',
    logo: '',
    favicon: '',
    brandType: 'default',
    brandText: '',
    brandImage: '',
  },
} as LayoutPayload

function renderReset(page: ResetPasswordPageProps = { token: 'reset-token' }) {
  const resetPassword = vi.fn().mockResolvedValue({ value: undefined, message: '密码已更新' })
  const runtime: GooseRuntime = {
    api: { auth: { resetPassword } } as unknown as GooseSiteApi,
    currentUrl: '/reset-password',
    isNavigating: false, theme: 'gf-light',
    locale: 'zh',
    navigate: vi.fn(),
    queueFlash: vi.fn(),
    redirect: vi.fn(),
    refresh: vi.fn(),
    setLocale: vi.fn(), toggleTheme: vi.fn(),
  }

  render(
    <GooseI18nProvider locale="zh">
      <GooseRuntimeProvider runtime={runtime}>
        <ResetPasswordPageView layout={layout} page={page} />
      </GooseRuntimeProvider>
    </GooseI18nProvider>,
  )
  return { resetPassword, user: userEvent.setup() }
}

describe('ResetPasswordPageView', () => {
  it('explains a missing token and prevents submission', () => {
    const { resetPassword } = renderReset({ token: '' })

    expect(screen.getByText('重置链接缺少 token，请重新从邮件打开。')).toBeTruthy()
    expect((screen.getByRole('button', { name: '保存新密码' }) as HTMLButtonElement).disabled).toBe(true)
    expect(resetPassword).not.toHaveBeenCalled()
  })

  it('validates password length and equality before calling the API', async () => {
    const { resetPassword, user } = renderReset()
    const password = screen.getByLabelText('新密码')
    const confirmation = screen.getByLabelText('确认密码')

    await user.type(password, '123')
    await user.type(confirmation, '123')
    await user.click(screen.getByRole('button', { name: '保存新密码' }))
    expect(await screen.findByText('密码长度至少 6 位')).toBeTruthy()

    await user.clear(password)
    await user.clear(confirmation)
    await user.type(password, '123456')
    await user.type(confirmation, 'abcdef')
    await user.click(screen.getByRole('button', { name: '保存新密码' }))
    expect(await screen.findByText('两次输入的密码不一致')).toBeTruthy()
    expect(resetPassword).not.toHaveBeenCalled()
  })

  it('submits a valid token and clears passwords after success', async () => {
    const { resetPassword, user } = renderReset()
    const password = screen.getByLabelText('新密码')
    const confirmation = screen.getByLabelText('确认密码')

    await user.type(password, 'secure123')
    await user.type(confirmation, 'secure123')
    await user.click(screen.getByRole('button', { name: '保存新密码' }))

    expect(await screen.findByText('密码已更新')).toBeTruthy()
    expect(resetPassword).toHaveBeenCalledWith('reset-token', 'secure123')
    expect((password as HTMLInputElement).value).toBe('')
    expect((confirmation as HTMLInputElement).value).toBe('')
  })

  it('keeps the API error visible and allows returning to login', async () => {
    const { resetPassword, user } = renderReset()
    resetPassword.mockRejectedValueOnce(new Error('重置链接已过期或无效'))

    await user.type(screen.getByLabelText('新密码'), 'secure123')
    await user.type(screen.getByLabelText('确认密码'), 'secure123')
    await user.click(screen.getByRole('button', { name: '保存新密码' }))

    expect(await screen.findByText('重置链接已过期或无效')).toBeTruthy()
    expect(screen.getByRole('link', { name: '返回登录' }).getAttribute('href')).toBe('/login')
  })
})
