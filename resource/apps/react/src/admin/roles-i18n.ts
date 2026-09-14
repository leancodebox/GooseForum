import type { AuthLocale } from '@gooseforum/react/i18n/auth'

const en = {
  title: 'Roles', description: 'Manage administrator roles and permission sets.', create: 'New role', search: 'Search roles…', searchAction: 'Search', all: 'All', enabled: 'Enabled', disabled: 'Disabled', refresh: 'Refresh', loading: 'Loading…', loadFailed: 'Failed to load roles', retry: 'Retry', empty: 'No matching roles', id: 'ID', name: 'Role name', status: 'Status', permissions: 'Permissions', createdAt: 'Created', actions: 'Actions', edit: 'Edit', delete: 'Delete', createTitle: 'Create role', editTitle: 'Edit role', formHint: 'Select the permissions granted by this role.', namePlaceholder: 'Enter a role name', nameRequired: 'Enter a role name', permissionRequired: 'Select at least one permission', save: 'Save', saving: 'Saving…', cancel: 'Cancel', saved: 'Role saved', saveFailed: 'Failed to save role', deleteTitle: 'Delete role', deleteConfirm: 'Delete this role?', deleted: 'Role deleted', deleteFailed: 'Failed to delete role', previous: 'Previous', next: 'Next', page: 'Page', perPage: 'per page',
} as const

const zh: Record<keyof typeof en, string> = {
  title: '角色', description: '管理后台角色及其权限集合。', create: '新建角色', search: '搜索角色…', searchAction: '搜索', all: '全部', enabled: '启用', disabled: '停用', refresh: '刷新', loading: '正在加载…', loadFailed: '角色加载失败', retry: '重试', empty: '没有匹配的角色', id: 'ID', name: '角色名称', status: '状态', permissions: '权限', createdAt: '创建时间', actions: '操作', edit: '编辑', delete: '删除', createTitle: '新建角色', editTitle: '编辑角色', formHint: '选择该角色拥有的后台权限。', namePlaceholder: '请输入角色名称', nameRequired: '请输入角色名称', permissionRequired: '请至少选择一项权限', save: '保存', saving: '保存中…', cancel: '取消', saved: '角色已保存', saveFailed: '角色保存失败', deleteTitle: '删除角色', deleteConfirm: '确定删除这个角色吗？', deleted: '角色已删除', deleteFailed: '角色删除失败', previous: '上一页', next: '下一页', page: '第', perPage: '每页',
}

const resources = {
  zh,
  en,
  ja: { ...en, title: 'ロール', create: '新規ロール', edit: '編集', delete: '削除', save: '保存', cancel: 'キャンセル' },
  it: { ...en, title: 'Ruoli', create: 'Nuovo ruolo', edit: 'Modifica', delete: 'Elimina', save: 'Salva', cancel: 'Annulla' },
} as const

export type RoleTextKey = keyof typeof en
export function createRoleText(locale: AuthLocale) {
  const dictionary = resources[locale]
  return (key: RoleTextKey) => dictionary[key]
}
