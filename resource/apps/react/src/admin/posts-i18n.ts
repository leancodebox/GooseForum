import type { AuthLocale } from '@gooseforum/react/i18n/auth'

const en = {
  title: 'Posts', description: 'Review topics and replies, manage categories, visibility, and pin weight.', topics: 'Topics', replies: 'Replies', allReviews: 'All review states', allCategories: 'All categories', pending: 'Pending', approved: 'Approved', rejected: 'Rejected', none: 'Not reviewed', search: 'Search content…', searchAction: 'Search', clear: 'Clear', refresh: 'Refresh', loading: 'Loading…', loadFailed: 'Failed to load content', empty: 'No matching content', previous: 'Previous', next: 'Next', page: 'Page', topic: 'Topic', author: 'Author', status: 'Status', actions: 'Actions', active: 'Active', inactive: 'Inactive', deleted: 'Deleted', processed: 'Processed', pin: 'Pin', noCategory: 'No category', views: 'Views', repliesCount: 'Replies', likes: 'Likes', openReview: 'View source / review', editCategories: 'Edit categories', editPin: 'Set pin weight', restore: 'Restore', disable: 'Disable', delete: 'Delete', categoryTitle: 'Topic categories', categoryHint: 'Select one to three categories. The first is the main category.', mainCategory: 'Main', categoryRequired: 'Select at least one available category', categoryLimit: 'A topic can have at most three categories', mainCategoryChange: 'Change main category?', mainCategoryChangeHint: 'Changing the first category changes the topic’s primary category.', save: 'Save', saving: 'Saving…', cancel: 'Cancel', saved: 'Saved', pinTitle: 'Pin weight', pinHint: 'Use 0 to remove pinning; higher values rank first.', sourceEmpty: 'No source content', close: 'Close', copy: 'Copy source', copied: 'Source copied', reviewReason: 'Enter a review reason', approve: 'Approve', reject: 'Reject', recheck: 'Recheck', reviewSaved: 'Review saved', actionDisableTitle: 'Disable topic?', actionDisableHint: 'The topic will be hidden from normal listings.', actionRestoreTitle: 'Restore topic?', actionRestoreHint: 'The topic will return to normal processing.', confirm: 'Confirm', deleteTitle: 'Delete topic?', deleteHint: 'This deletes the selected topic.', unknownCategory: 'Unknown category',
} as const

const zh: Record<keyof typeof en, string> = {
  title: '内容', description: '审核主题和回复，并管理分类、状态及置顶权重。', topics: '主题', replies: '回复', allReviews: '全部审核状态', allCategories: '全部分类', pending: '待审核', approved: '已通过', rejected: '已拒绝', none: '未审核', search: '搜索内容…', searchAction: '搜索', clear: '清除', refresh: '刷新', loading: '正在加载…', loadFailed: '内容加载失败', empty: '没有匹配的内容', previous: '上一页', next: '下一页', page: '第', topic: '主题', author: '作者', status: '状态', actions: '操作', active: '正常', inactive: '停用', deleted: '已删除', processed: '已处理', pin: '置顶', noCategory: '无分类', views: '浏览', repliesCount: '回复', likes: '点赞', openReview: '查看来源 / 审核', editCategories: '修改分类', editPin: '设置置顶', restore: '恢复', disable: '停用', delete: '删除', categoryTitle: '主题分类', categoryHint: '选择一至三个分类，第一个分类为主分类。', mainCategory: '主分类', categoryRequired: '请至少选择一个有效分类', categoryLimit: '一个主题最多选择三个分类', mainCategoryChange: '确认修改主分类？', mainCategoryChangeHint: '调整第一个分类会改变主题的主分类。', save: '保存', saving: '保存中…', cancel: '取消', saved: '已保存', pinTitle: '置顶权重', pinHint: '设置为 0 可取消置顶；数值越大排序越靠前。', sourceEmpty: '暂无来源内容', close: '关闭', copy: '复制原文', copied: '原文已复制', reviewReason: '请输入审核原因', approve: '通过', reject: '拒绝', recheck: '重新审核', reviewSaved: '审核结果已保存', actionDisableTitle: '确认停用主题？', actionDisableHint: '主题将从普通列表中隐藏。', actionRestoreTitle: '确认恢复主题？', actionRestoreHint: '主题将恢复正常处理状态。', confirm: '确认', deleteTitle: '确认删除主题？', deleteHint: '此操作将删除选中的主题。', unknownCategory: '未知分类',
}

const resources = {
  zh,
  en,
  ja: { ...en, title: 'コンテンツ', topics: 'トピック', replies: '返信', search: 'コンテンツを検索…', save: '保存', cancel: 'キャンセル' },
  it: { ...en, title: 'Contenuti', topics: 'Discussioni', replies: 'Risposte', search: 'Cerca contenuti…', save: 'Salva', cancel: 'Annulla' },
} as const

export type PostTextKey = keyof typeof en
export function createPostText(locale: AuthLocale) {
  const dictionary = resources[locale]
  return (key: PostTextKey) => dictionary[key]
}
