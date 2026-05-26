import type {
  ApiEnvelope,
  AdminArticle,
  AdminBadge,
  AdminCategory,
  AdminOptRecord,
  AdminRole,
  AdminUser,
  AnnouncementConfig,
  ArticleSource,
  DailyTraffic,
  FriendLinkGroup,
  GithubRelease,
  MailSettings,
  PageResult,
  PostingSettings,
  SecuritySettings,
  ServerVersion,
  SiteSettings,
  SiteStatistics,
  SponsorsConfig,
  UserBadgeOptions,
} from '@/admin/types'

function responseMessage(data: ApiEnvelope<unknown>, fallback: string) {
  return String(data.message || data.msg || fallback)
}

async function readApiResponse<T>(response: Response, fallback: string): Promise<T> {
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }
  const data = (await response.json()) as ApiEnvelope<T>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(responseMessage(data, fallback))
  }
  const result = data.result ?? data.data
  return result as T
}

async function getJson<T>(url: string, fallback: string): Promise<T> {
  const response = await fetch(url, { headers: { Accept: 'application/json' } })
  return readApiResponse<T>(response, fallback)
}

async function postJson<T>(url: string, body?: unknown, fallback = '请求失败'): Promise<T> {
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body ?? {}),
  })
  return readApiResponse<T>(response, fallback)
}

async function postEnvelope<T>(url: string, body?: unknown, fallback = '请求失败'): Promise<ApiEnvelope<T>> {
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body ?? {}),
  })
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }
  const data = (await response.json()) as ApiEnvelope<T>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(responseMessage(data, fallback))
  }
  return data
}

export async function getSiteStatistics(): Promise<SiteStatistics> {
  const response = await fetch('/api/forum/get-site-statistics', {
    headers: { Accept: 'application/json' },
  })
  return readApiResponse<SiteStatistics>(response, '获取站点统计失败')
}

export async function getTrafficOverview(startDate?: string, endDate?: string): Promise<DailyTraffic[]> {
  return postJson<DailyTraffic[]>('/api/admin/traffic-overview', { startDate, endDate }, '获取流量概览失败')
}

export async function getServerVersion(): Promise<ServerVersion> {
  const response = await fetch('/api/admin/server-version', {
    headers: { Accept: 'application/json' },
  })
  return readApiResponse<ServerVersion>(response, '获取服务端版本失败')
}

export async function getGithubReleases(): Promise<GithubRelease[]> {
  const response = await fetch('https://api.github.com/repos/leancodebox/GooseForum/releases', {
    headers: { Accept: 'application/vnd.github+json' },
  })
  if (!response.ok) {
    throw new Error('无法获取版本信息')
  }
  return (await response.json()) as GithubRelease[]
}

export function getUserList(params: { page?: number, pageSize?: number, username?: string }) {
  return postJson<PageResult<AdminUser>>('/api/admin/user-list', params, '获取用户列表失败')
}

export function editUser(data: { userId: number, status: number, validate: number, roleId: number }) {
  return postJson<unknown>('/api/admin/user-edit', data, '保存用户失败')
}

export function getAllRoleItem() {
  return getJson<{ name: string, value: number }[]>('/api/admin/get-all-role-item', '获取角色选项失败')
}

export function getUserBadgeOptions(userId: number) {
  return postJson<UserBadgeOptions>('/api/admin/user-badge-options', { userId }, '获取用户徽章失败')
}

export function saveUserBadges(userId: number, badgeCodes: string[]) {
  return postJson<unknown>('/api/admin/save-user-badges', { userId, badgeCodes }, '保存用户徽章失败')
}

export function getRoleList() {
  return postJson<PageResult<AdminRole>>('/api/admin/role-list', {}, '获取角色列表失败')
}

export function saveRole(data: { id: number, roleName: string, permissions: number[] }) {
  return postJson<unknown>('/api/admin/role-save', data, '保存角色失败')
}

export function deleteRole(id: number) {
  return postJson<unknown>('/api/admin/role-delete', { id }, '删除角色失败')
}

export function getCategoryList() {
  return postJson<AdminCategory[]>('/api/admin/category-list', {}, '获取分类列表失败')
}

export function saveCategory(data: AdminCategory & { id: number }) {
  return postJson<unknown>('/api/admin/category-save', data, '保存分类失败')
}

export function deleteCategory(id: number) {
  return postJson<unknown>('/api/admin/category-delete', { id }, '删除分类失败')
}

export function getArticlesList(params: { page?: number, pageSize?: number, search?: string }) {
  return postJson<PageResult<AdminArticle>>('/api/admin/articles-list', params, '获取帖子列表失败')
}

export function getOptRecordList(params: { page?: number, pageSize?: number, optUserId?: number, optType?: number, targetType?: number, targetId?: number }) {
  return postJson<PageResult<AdminOptRecord>>('/api/admin/opt-record-page', params, '获取操作记录失败')
}

export function getArticleSource(id: number) {
  return postJson<ArticleSource>('/api/admin/article-source', { id }, '获取帖子原文失败')
}

export function editArticle(data: { id: number, processStatus: number }) {
  return postJson<unknown>('/api/admin/article-edit', data, '保存帖子状态失败')
}

export function editArticlePin(data: { id: number, pinWeight: number }) {
  return postJson<unknown>('/api/admin/article-pin-edit', data, '保存置顶权重失败')
}

export function editArticleCategories(data: { id: number, categoryId: number[] }) {
  return postJson<unknown>('/api/admin/article-categories-edit', data, '保存帖子分类失败')
}

export function getFriendLinks() {
  return getJson<FriendLinkGroup[]>('/api/admin/friend-links', '获取友情链接失败')
}

export function saveFriendLinks(linksInfo: FriendLinkGroup[]) {
  return postJson<unknown>('/api/admin/save-friend-links', { linksInfo }, '保存友情链接失败')
}

export function getSponsors() {
  return getJson<SponsorsConfig>('/api/admin/sponsors', '获取赞助配置失败')
}

export function saveSponsors(sponsorsInfo: SponsorsConfig) {
  return postJson<unknown>('/api/admin/save-sponsors', { sponsorsInfo }, '保存赞助配置失败')
}

export function getBadges() {
  return getJson<AdminBadge[]>('/api/admin/badges', '获取徽章失败')
}

export function saveBadge(data: AdminBadge) {
  return postJson<unknown>('/api/admin/badge-save', data, '保存徽章失败')
}

export function deleteBadge(code: string) {
  return postJson<unknown>('/api/admin/badge-delete', { code }, '删除徽章失败')
}

export function getSiteSettings() {
  return getJson<SiteSettings>('/api/admin/site-settings', '获取站点设置失败')
}

export function getMailSettings() {
  return getJson<MailSettings>('/api/admin/mail-settings', '获取发信服务失败')
}

export function getSecuritySettings() {
  return getJson<SecuritySettings>('/api/admin/security-settings', '获取安全设置失败')
}

export function getPostingSettings() {
  return getJson<PostingSettings>('/api/admin/posting-settings', '获取发布设置失败')
}

export function getAnnouncement() {
  return getJson<AnnouncementConfig>('/api/admin/announcement', '获取系统公告失败')
}

export function saveSiteSettings(settings: SiteSettings) {
  return postJson<unknown>('/api/admin/save-site-settings', { settings }, '保存站点设置失败')
}

export function saveMailSettings(settings: MailSettings) {
  return postJson<unknown>('/api/admin/save-mail-settings', { settings }, '保存发信服务失败')
}

export function testMailConnection(settings: MailSettings, testEmail: string) {
  return postEnvelope<{ success?: boolean, message?: string }>('/api/admin/test-mail-connection', { settings, testEmail }, '发送测试邮件失败')
}

export function saveSecuritySettings(settings: SecuritySettings) {
  return postJson<unknown>('/api/admin/save-security-settings', { settings }, '保存安全设置失败')
}

export function savePostingSettings(settings: PostingSettings) {
  return postJson<unknown>('/api/admin/save-posting-settings', { settings }, '保存发布设置失败')
}

export function saveAnnouncement(settings: AnnouncementConfig) {
  return postJson<unknown>('/api/admin/save-announcement', { settings }, '保存系统公告失败')
}
