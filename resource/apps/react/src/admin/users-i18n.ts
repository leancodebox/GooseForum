import type { AuthLocale } from '@gooseforum/react/i18n/auth'

const en = {
  title: 'Users', description: 'Search users and manage account status, roles, and badges.', search: 'Search username…', searchAction: 'Search', clear: 'Clear', refresh: 'Refresh', loading: 'Loading…', loadFailed: 'Failed to load users', retry: 'Retry', empty: 'No matching users', user: 'User', roles: 'Roles', status: 'Status', createdAt: 'Created', lastActive: 'Last active', actions: 'Actions', enabled: 'Active', disabled: 'Disabled', verified: 'Verified', unverified: 'Unverified', noRole: 'No role', never: 'Never', edit: 'Edit user', account: 'Account', emailVerification: 'Email verification', role: 'Role', details: 'Details', prestige: 'Prestige', badges: 'Badges', badgeHint: 'Automatic badges are read-only; manual badges can be selected below.', automaticBadges: 'Automatic badges', manualBadges: 'Manual badges', noBadges: 'No manual badges available', badgeLoading: 'Loading badges…', save: 'Save changes', saving: 'Saving…', cancel: 'Cancel', saved: 'User saved', saveFailed: 'Failed to save user', perPage: 'per page', page: 'Page', previous: 'Previous', next: 'Next',
} as const

const zh: Record<keyof typeof en, string> = {
  title: '用户', description: '搜索用户并管理账号状态、角色和徽章。', search: '搜索用户名…', searchAction: '搜索', clear: '清除', refresh: '刷新', loading: '正在加载…', loadFailed: '用户加载失败', retry: '重试', empty: '没有匹配的用户', user: '用户', roles: '角色', status: '状态', createdAt: '创建时间', lastActive: '最后活跃', actions: '操作', enabled: '正常', disabled: '禁用', verified: '已验证', unverified: '未验证', noRole: '无角色', never: '从未', edit: '编辑用户', account: '账号设置', emailVerification: '邮箱验证', role: '角色', details: '用户信息', prestige: '声望', badges: '徽章', badgeHint: '自动徽章仅供查看；可在下方选择手动徽章。', automaticBadges: '自动徽章', manualBadges: '手动徽章', noBadges: '暂无可用的手动徽章', badgeLoading: '正在加载徽章…', save: '保存修改', saving: '保存中…', cancel: '取消', saved: '用户已保存', saveFailed: '用户保存失败', perPage: '每页', page: '第', previous: '上一页', next: '下一页',
}

const resources = {
  zh,
  en,
  ja: { ...en, title: 'ユーザー', search: 'ユーザー名を検索…', edit: 'ユーザーを編集', save: '変更を保存', cancel: 'キャンセル' },
  it: { ...en, title: 'Utenti', search: 'Cerca nome utente…', edit: 'Modifica utente', save: 'Salva modifiche', cancel: 'Annulla' },
} as const

export type UserTextKey = keyof typeof en
export function createUserText(locale: AuthLocale) {
  const dictionary = resources[locale]
  return (key: UserTextKey) => dictionary[key]
}
