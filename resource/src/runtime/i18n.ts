import { computed, inject, provide, ref, type ComputedRef, type InjectionKey, type Ref } from 'vue'

type Locale = 'zh' | 'en' | 'ja'
type Dict = Record<string, string>

const fallbackLocale: Locale = 'zh'
const supportedLocales: Locale[] = ['zh', 'en', 'ja']

const messages: Record<Locale, Dict> = {
  zh: {
    'common.back': '返回',
    'common.home': '回到首页',
    'common.loading': '处理中...',
    'shell.openMenu': '打开菜单',
    'shell.menu': '菜单',
    'shell.closeMenu': '关闭菜单',
    'shell.search': '搜索',
    'shell.switchLanguage': '切换语言',
    'shell.notifications': '通知',
    'shell.publish': '发布',
    'shell.userMenu': '用户菜单',
    'shell.profile': '个人主页',
    'shell.settings': '设置',
    'shell.admin': '管理后台',
    'shell.logout': '退出登录',
    'shell.login': '登录',
    'shell.register': '注册',
    'shell.resources': 'Resources',
    'shell.categories': 'Categories',
    'notifications.title': '通知',
    'notifications.unread': '{count} 未读',
    'notifications.summary': '共 {total} 条，回复、点赞、关注和系统消息会集中展示在这里。',
    'notifications.markAllRead': '全部已读',
    'notifications.table.notification': 'Notification',
    'notifications.table.time': 'Time',
    'notifications.emptyTitle': '暂无通知',
    'notifications.emptyDescription': '新的回复、点赞和关注会出现在这里。',
    'notifications.fallback': '查看通知详情',
    'notifications.actorFallback': '用户',
    'notifications.verb.like': '点赞了',
    'notifications.verb.follow': '关注了你',
    'notifications.verb.reply': '回复了',
    'notifications.verb.comment': '评论了',
    'notifications.markAllReadFailed': '标记已读失败',
  },
  en: {
    'common.back': 'Back',
    'common.home': 'Home',
    'common.loading': 'Working...',
    'shell.openMenu': 'Open menu',
    'shell.menu': 'Menu',
    'shell.closeMenu': 'Close menu',
    'shell.search': 'Search',
    'shell.switchLanguage': 'Switch language',
    'shell.notifications': 'Notifications',
    'shell.publish': 'Publish',
    'shell.userMenu': 'User menu',
    'shell.profile': 'Profile',
    'shell.settings': 'Settings',
    'shell.admin': 'Admin',
    'shell.logout': 'Log out',
    'shell.login': 'Log in',
    'shell.register': 'Sign up',
    'shell.resources': 'Resources',
    'shell.categories': 'Categories',
    'notifications.title': 'Notifications',
    'notifications.unread': '{count} unread',
    'notifications.summary': '{total} notifications, including replies, likes, follows, and system messages.',
    'notifications.markAllRead': 'Mark all read',
    'notifications.table.notification': 'Notification',
    'notifications.table.time': 'Time',
    'notifications.emptyTitle': 'No notifications',
    'notifications.emptyDescription': 'Replies, likes, and follows will show up here.',
    'notifications.fallback': 'View notification',
    'notifications.actorFallback': 'User',
    'notifications.verb.like': 'liked',
    'notifications.verb.follow': 'followed you',
    'notifications.verb.reply': 'replied to',
    'notifications.verb.comment': 'commented on',
    'notifications.markAllReadFailed': 'Failed to mark notifications read',
  },
  ja: {
    'common.back': '戻る',
    'common.home': 'ホームへ',
    'common.loading': '処理中...',
    'shell.openMenu': 'メニューを開く',
    'shell.menu': 'メニュー',
    'shell.closeMenu': 'メニューを閉じる',
    'shell.search': '検索',
    'shell.switchLanguage': '言語を切り替え',
    'shell.notifications': '通知',
    'shell.publish': '投稿',
    'shell.userMenu': 'ユーザーメニュー',
    'shell.profile': 'プロフィール',
    'shell.settings': '設定',
    'shell.admin': '管理画面',
    'shell.logout': 'ログアウト',
    'shell.login': 'ログイン',
    'shell.register': '登録',
    'shell.resources': 'Resources',
    'shell.categories': 'Categories',
    'notifications.title': '通知',
    'notifications.unread': '{count} 未読',
    'notifications.summary': '全 {total} 件。返信、いいね、フォロー、システム通知をまとめて表示します。',
    'notifications.markAllRead': 'すべて既読',
    'notifications.table.notification': 'Notification',
    'notifications.table.time': 'Time',
    'notifications.emptyTitle': '通知はありません',
    'notifications.emptyDescription': '返信、いいね、フォローがここに表示されます。',
    'notifications.fallback': '通知を見る',
    'notifications.actorFallback': 'ユーザー',
    'notifications.verb.like': 'いいねしました',
    'notifications.verb.follow': 'あなたをフォローしました',
    'notifications.verb.reply': '返信しました',
    'notifications.verb.comment': 'コメントしました',
    'notifications.markAllReadFailed': '既読にできませんでした',
  },
}

interface I18nContext {
  locale: Ref<Locale>
  t: (key: string, params?: Record<string, string | number>) => string
}

const I18nKey: InjectionKey<I18nContext> = Symbol('goose-i18n')

function normalizeLocale(value?: string | null): Locale | undefined {
  const normalized = (value || '').toLowerCase()
  const short = normalized.split('-')[0] as Locale
  return supportedLocales.includes(short) ? short : undefined
}

function readCookie(name: string) {
  if (typeof document === 'undefined') return ''
  return document.cookie
    .split('; ')
    .find((item) => item.startsWith(`${name}=`))
    ?.split('=')
    .slice(1)
    .join('=') || ''
}

function detectLocale(): Locale {
  const cookieLocale = normalizeLocale(decodeURIComponent(readCookie('lang')))
  if (cookieLocale) return cookieLocale
  if (typeof navigator !== 'undefined') return normalizeLocale(navigator.language) || fallbackLocale
  return fallbackLocale
}

export function createI18n() {
  const locale = ref<Locale>(detectLocale())

  function t(key: string, params: Record<string, string | number> = {}) {
    const template = messages[locale.value][key] || messages[fallbackLocale][key] || key
    return Object.entries(params).reduce(
      (text, [name, value]) => text.replace(new RegExp(`\\{${name}\\}`, 'g'), String(value)),
      template,
    )
  }

  return { locale, t }
}

export function provideI18n() {
  const i18n = createI18n()
  provide(I18nKey, i18n)
  return i18n
}

export function useI18n(): I18nContext {
  const i18n = inject(I18nKey)
  if (!i18n) return createI18n()
  return i18n
}

export function useLocaleLabel(): ComputedRef<string> {
  const { locale } = useI18n()
  return computed(() => locale.value)
}
