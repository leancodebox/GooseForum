import { adminText } from '@/admin/runtime/i18n-text'
import { resolveApiMessage } from '@/runtime/api-message'
import type {
  ApiEnvelope,
  AdminArticle,
  AdminBadge,
  AdminCategory,
  AdminOptRecord,
  AdminPermissionOption,
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
  return resolveApiMessage(data, fallback)
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

async function postJson<T>(url: string, body?: unknown, fallback = adminText('k000l')): Promise<T> {
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

async function postEnvelope<T>(url: string, body?: unknown, fallback = adminText('k000l')): Promise<ApiEnvelope<T>> {
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
  return readApiResponse<SiteStatistics>(response, adminText('k000m'))
}

export async function getTrafficOverview(startDate?: string, endDate?: string): Promise<DailyTraffic[]> {
  return postJson<DailyTraffic[]>('/api/admin/traffic-overview', { startDate, endDate }, adminText('k000n'))
}

export async function getServerVersion(): Promise<ServerVersion> {
  const response = await fetch('/api/admin/server-version', {
    headers: { Accept: 'application/json' },
  })
  return readApiResponse<ServerVersion>(response, adminText('k000o'))
}

export async function getGithubReleases(): Promise<GithubRelease[]> {
  const response = await fetch('https://api.github.com/repos/leancodebox/GooseForum/releases', {
    headers: { Accept: 'application/vnd.github+json' },
  })
  if (!response.ok) {
    throw new Error(adminText('k000p'))
  }
  return (await response.json()) as GithubRelease[]
}

export function getUserList(params: { page?: number, pageSize?: number, username?: string, userId?: number, email?: string }) {
  return postJson<PageResult<AdminUser>>('/api/admin/user-list', params, adminText('k000q'))
}

export function editUser(data: { userId: number, status: number, validate: number, roleId: number }) {
  return postJson<unknown>('/api/admin/user-edit', data, adminText('k000r'))
}

export function getAllRoleItem() {
  return getJson<{ name: string, value: number }[]>('/api/admin/get-all-role-item', adminText('k000s'))
}

export function getUserBadgeOptions(userId: number) {
  return postJson<UserBadgeOptions>('/api/admin/user-badge-options', { userId }, adminText('k000t'))
}

export function saveUserBadges(userId: number, badgeCodes: string[]) {
  return postJson<unknown>('/api/admin/save-user-badges', { userId, badgeCodes }, adminText('k000u'))
}

export function getRoleList() {
  return postJson<PageResult<AdminRole>>('/api/admin/role-list', {}, adminText('k000v'))
}

export function getPermissionList() {
  return postJson<AdminPermissionOption[]>('/api/admin/get-permission-list', {}, adminText('k000w'))
}

export function saveRole(data: { id: number, roleName: string, permissions: number[] }) {
  return postJson<unknown>('/api/admin/role-save', data, adminText('k000x'))
}

export function deleteRole(id: number) {
  return postJson<unknown>('/api/admin/role-delete', { id }, adminText('k000y'))
}

export function getCategoryList() {
  return postJson<AdminCategory[]>('/api/admin/category-list', {}, adminText('k000z'))
}

export function saveCategory(data: AdminCategory & { id: number }) {
  return postJson<unknown>('/api/admin/category-save', data, adminText('k0010'))
}

export function deleteCategory(id: number) {
  return postJson<unknown>('/api/admin/category-delete', { id }, adminText('k0011'))
}

export function addCategoryModerator(data: { categoryId: number, userId?: number, username?: string }) {
  return postJson<unknown>('/api/admin/category-moderator-add', data, '添加版主失败')
}

export function deleteCategoryModerator(id: number) {
  return postJson<unknown>('/api/admin/category-moderator-delete', { id }, '移除版主失败')
}

export function getArticlesList(params: { page?: number, pageSize?: number, search?: string }) {
  return postJson<PageResult<AdminArticle>>('/api/admin/articles-list', params, adminText('k0012'))
}

export function getOptRecordList(params: { page?: number, pageSize?: number, optUserId?: number, optType?: number, targetType?: number, targetId?: number }) {
  return postJson<PageResult<AdminOptRecord>>('/api/admin/opt-record-page', params, adminText('k0013'))
}

export function getArticleSource(id: number) {
  return postJson<ArticleSource>('/api/admin/article-source', { id }, adminText('k0014'))
}

export function editArticle(data: { id: number, processStatus: number }) {
  return postJson<unknown>('/api/admin/article-edit', data, adminText('k0015'))
}

export function deleteArticle(id: number) {
  return postJson<unknown>('/api/admin/article-delete', { id }, adminText('k00cd'))
}

export function editArticlePin(data: { id: number, pinWeight: number }) {
  return postJson<unknown>('/api/admin/article-pin-edit', data, adminText('k0016'))
}

export function editArticleCategories(data: { id: number, categoryId: number[] }) {
  return postJson<unknown>('/api/admin/article-categories-edit', data, adminText('k0017'))
}

export function getFriendLinks() {
  return getJson<FriendLinkGroup[]>('/api/admin/friend-links', adminText('k0018'))
}

export function saveFriendLinks(linksInfo: FriendLinkGroup[]) {
  return postJson<unknown>('/api/admin/save-friend-links', { linksInfo }, adminText('k0019'))
}

export function getSponsors() {
  return getJson<SponsorsConfig>('/api/admin/sponsors', adminText('k001a'))
}

export function saveSponsors(sponsorsInfo: SponsorsConfig) {
  return postJson<unknown>('/api/admin/save-sponsors', { sponsorsInfo }, adminText('k001b'))
}

export function getBadges() {
  return getJson<AdminBadge[]>('/api/admin/badges', adminText('k001c'))
}

export function saveBadge(data: AdminBadge) {
  return postJson<unknown>('/api/admin/badge-save', data, adminText('k001d'))
}

export function deleteBadge(code: string) {
  return postJson<unknown>('/api/admin/badge-delete', { code }, adminText('k001e'))
}

export function getSiteSettings() {
  return getJson<SiteSettings>('/api/admin/site-settings', adminText('k001f'))
}

export function getMailSettings() {
  return getJson<MailSettings>('/api/admin/mail-settings', adminText('k001g'))
}

export function getSecuritySettings() {
  return getJson<SecuritySettings>('/api/admin/security-settings', adminText('k001h'))
}

export function getPostingSettings() {
  return getJson<PostingSettings>('/api/admin/posting-settings', adminText('k001i'))
}

export function getAnnouncement() {
  return getJson<AnnouncementConfig>('/api/admin/announcement', adminText('k001j'))
}

export function saveSiteSettings(settings: SiteSettings) {
  return postJson<unknown>('/api/admin/save-site-settings', { settings }, adminText('k001k'))
}

export function saveMailSettings(settings: MailSettings) {
  return postJson<unknown>('/api/admin/save-mail-settings', { settings }, adminText('k001l'))
}

export function testMailConnection(settings: MailSettings, testEmail: string) {
  return postEnvelope<{ success?: boolean, messageCode?: string, params?: Record<string, unknown> }>('/api/admin/test-mail-connection', { settings, testEmail }, adminText('k000i'))
}

export function saveSecuritySettings(settings: SecuritySettings) {
  return postJson<unknown>('/api/admin/save-security-settings', { settings }, adminText('k001m'))
}

export function savePostingSettings(settings: PostingSettings) {
  return postJson<unknown>('/api/admin/save-posting-settings', { settings }, adminText('k001n'))
}

export function saveAnnouncement(settings: AnnouncementConfig) {
  return postJson<unknown>('/api/admin/save-announcement', { settings }, adminText('k001o'))
}
