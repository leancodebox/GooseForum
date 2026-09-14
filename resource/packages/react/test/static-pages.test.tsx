import { afterEach, describe, expect, it, vi } from 'vitest'
import { cleanup, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { AnyPagePayload, GooseSiteApi, LayoutPayload } from '@gooseforum/client'
import { GooseApp } from '../src/app/root'
import { GooseI18nProvider } from '../src/i18n'
import { GooseRuntimeProvider, type GooseRuntime } from '../src/runtime'

afterEach(cleanup)

const layout = {
  site: { name: 'GooseForum', description: '', logo: '', favicon: '', brandType: 'default', brandText: '', brandImage: '' },
  viewer: { id: 0, username: '', email: '', avatarUrl: '', isAuthenticated: false, canAccessAdmin: false, isModerator: false, requiresEmailVerification: false, adminPermissions: [] },
  header: [
    { key: 'sponsors', label: 'Sponsors', i18nLabel: 'shell.nav.sponsors', url: '/sponsors' },
    { key: 'links', label: 'Links', i18nLabel: 'shell.nav.links', url: '/links' },
  ],
  sidebar: {
    activeKey: 'links',
    categories: [{ id: 4, label: 'Coding', url: '/c/Coding/4', color: '#8241d6' }],
  },
  footer: { links: [{ name: 'RSS', url: '/rss.xml' }], primary: ['GooseForum © 2024'] },
  unread: { notifications: false, messages: false },
  theme: { enabled: true, current: 'gf-light', themeColor: '#fbfdff' },
} satisfies LayoutPayload

function renderPage(page: AnyPagePayload) {
  const toggleTheme = vi.fn()
  const runtime: GooseRuntime = {
    api: {} as GooseSiteApi,
    currentUrl: page.url,
    isNavigating: false,
    locale: 'zh',
    theme: 'gf-light',
    navigate: vi.fn(),
    queueFlash: vi.fn(),
    redirect: vi.fn(),
    refresh: vi.fn(),
    setLocale: vi.fn(),
    toggleTheme,
  }
  render(
    <GooseI18nProvider locale="zh">
      <GooseRuntimeProvider runtime={runtime}>
        <GooseApp page={page} />
      </GooseRuntimeProvider>
    </GooseI18nProvider>,
  )
  return { toggleTheme, user: userEvent.setup() }
}

function payload(component: 'links.index' | 'sponsors.index' | 'categories.index' | 'members.index', props: unknown): AnyPagePayload {
  return {
    component,
    props,
    layout: { ...layout, sidebar: { ...layout.sidebar, activeKey: component.replace('.index', '') } },
    meta: { title: 'Static page' },
    url: `/${component.replace('.index', '')}`,
    version: '1.0',
  } as AnyPagePayload
}

describe('AppShell and static pages', () => {
  it('renders the Links information hierarchy and safe external links', () => {
    renderPage(payload('links.index', {
      totalCount: 1,
      groups: [{
        name: 'COMMUNITY', emoji: '👥', color: '#64748b',
        links: [{ name: 'Example', desc: 'Community site', url: 'https://example.com', logoUrl: '' }],
      }],
    }))

    expect(screen.getByRole('heading', { level: 1, name: '友情链接' })).toBeTruthy()
    expect(screen.getByRole('heading', { level: 2, name: /COMMUNITY/ })).toBeTruthy()
    const external = screen.getByRole('link', { name: /Example/ })
    expect(external.getAttribute('target')).toBe('_blank')
    expect(external.getAttribute('rel')).toBe('noopener noreferrer')
    expect(screen.getByRole('link', { name: '去发帖申请' }).getAttribute('href')).toBe('/publish')
    expect(screen.getAllByRole('link', { name: '友情链接' }).some((link) => link.getAttribute('aria-current') === 'page')).toBe(true)
  })

  it('supports mobile navigation and the shell theme action', async () => {
    const { toggleTheme, user } = renderPage(payload('links.index', { totalCount: 0, groups: [] }))

    await user.click(screen.getByRole('button', { name: '打开菜单' }))
    const dialog = await screen.findByRole('dialog')
    expect(within(dialog).getByRole('heading', { name: '菜单' })).toBeTruthy()
    expect(within(dialog).getByRole('link', { name: 'Coding' })).toBeTruthy()

    await user.click(within(dialog).getByRole('button', { name: '关闭菜单' }))
    await user.click(screen.getByRole('button', { name: '切换到深色主题' }))
    expect(toggleTheme).toHaveBeenCalledOnce()
  })

  it('renders sponsor tiers, defaults, contact, and rules', () => {
    renderPage(payload('sponsors.index', {
      totalCount: 1,
      content: { title: '感谢支持', description: '支持 GooseForum 的朋友们。' },
      contact: { title: '联系我们', description: '欢迎支持。', buttonText: '发送邮件', buttonLink: 'mailto:test@example.com' },
      rules: [{ content: '内容公开透明。' }],
      sections: [{
        key: 'gold', label: 'Gold', tone: 'gold',
        sponsors: [{ name: 'Alice', message: '', link: 'https://example.com/alice', avatarUrl: '/avatar.webp' }],
      }],
    }))

    expect(screen.getByRole('heading', { level: 1, name: '感谢支持' })).toBeTruthy()
    expect(screen.getByText('感谢支持 GooseForum。')).toBeTruthy()
    const sponsor = screen.getByRole('link', { name: /Alice/ })
    expect(sponsor.getAttribute('rel')).toBe('noopener noreferrer')
    expect(screen.getByRole('link', { name: '发送邮件' }).getAttribute('href')).toBe('mailto:test@example.com')
    expect(screen.getByText('内容公开透明。')).toBeTruthy()
  })

  it('uses official empty states when payload collections are empty', () => {
    renderPage(payload('links.index', { totalCount: 0, groups: [] }))
    expect(screen.getByText('暂无链接')).toBeTruthy()
    expect(screen.getByText('站点还没有配置友情链接。')).toBeTruthy()
  })

  it('preserves the authenticated shell actions and permission-gated admin links', async () => {
    const page = payload('links.index', { totalCount: 0, groups: [] })
    page.layout = {
      ...page.layout,
      viewer: {
        ...page.layout.viewer,
        id: 7,
        username: 'alice',
        avatarUrl: '/alice.webp',
        isAuthenticated: true,
        canAccessAdmin: true,
      },
      unread: { notifications: true, messages: true, moderationReports: false },
    }
    const { user } = renderPage(page)

    await user.click(screen.getByRole('button', { name: 'alice' }))

    expect(await screen.findByRole('menuitem', { name: /发布/ })).toBeTruthy()
    expect(screen.getByRole('menuitem', { name: /设置/ })).toBeTruthy()
    expect(screen.getByRole('menuitem', { name: /访问组/ })).toBeTruthy()
    expect(screen.getByRole('menuitem', { name: /主题预览/ })).toBeTruthy()
    expect(screen.getByRole('menuitem', { name: /管理后台/ })).toBeTruthy()
  })

  it('renders category identity, fallback copy, and compact topic counts', () => {
    renderPage(payload('categories.index', {
      total: 2,
      categories: [
        { id: 1, name: 'Coding', description: 'Development topics', icon: '💻', color: '#8241d6', url: '/c/Coding/1', topicCount: 1_250 },
        { id: 2, name: 'General', description: '', icon: '/general.webp', color: '#22c55e', url: '/c/General/2', topicCount: 3 },
      ],
    }))

    expect(screen.getByRole('heading', { level: 1, name: '全部分类' })).toBeTruthy()
    expect(screen.getByText('2 个分类')).toBeTruthy()
    expect(screen.getByText('1.3k 个主题')).toBeTruthy()
    expect(screen.getByText('这个分类还没有介绍。')).toBeTruthy()
    expect(document.querySelector('a[href="/c/Coding/1"]')).toBeTruthy()
    expect(document.querySelector('img[src="/general.webp"]')).toBeTruthy()
  })

  it('renders member identity, stats, fallbacks, and pagination semantics', () => {
    renderPage(payload('members.index', {
      members: [{
        id: 7, username: 'alice', nickname: 'Alice', avatarUrl: '/alice.webp', bio: '', prestige: 1_250,
        topicCount: 12, replyCount: 34, joinedAt: '2026-01-02', url: '/u/7',
      }],
      previousUrl: '/members?page=1',
      pagination: { page: 2, nextPage: 3, hasNext: true, nextUrl: '/members?page=3' },
    }))

    expect(screen.getByRole('heading', { level: 1, name: '所有成员' })).toBeTruthy()
    expect(screen.getByText('@alice')).toBeTruthy()
    expect(screen.getByText('这位成员还没有填写个人简介。')).toBeTruthy()
    expect(screen.getByText('1.3k')).toBeTruthy()
    expect(screen.getByRole('link', { name: /Alice/ }).getAttribute('href')).toBe('/u/7')
    expect(screen.getByRole('link', { name: '上一页' }).getAttribute('rel')).toBe('prev')
    expect(screen.getByRole('link', { name: '下一页' }).getAttribute('rel')).toBe('next')
  })
})
