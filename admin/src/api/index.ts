import request from '@/lib/request'
import type { 
  PageResult, Category, User, Article, Role, Permission,
  DailyTraffic, SiteSettings, MailSettings, SecuritySettings,
  PostingSettings, FooterConfig, SponsorsConfig, AnnouncementConfig
} from './types'
import { LinkGroup } from '@/features/links-management/data/schema'

export function getFriendLinks() {
  return request.get<LinkGroup[]>('/api/admin/friend-links')
}

export function saveFriendLinks(linksInfo: LinkGroup[]) {
  return request.post<any>('/api/admin/save-friend-links', { linksInfo })
}

export function getCategoryList() {
  return request.post<Category[]>('/api/admin/category-list')
}

export function saveCategory(data: Category & { id: number }) {
  return request.post<any>('/api/admin/category-save', data)
}

export function deleteCategory(id: number) {
  return request.post<any>('/api/admin/category-delete', { id })
}

export interface UserListParams {
  username?: string
  userId?: number
  email?: string
  page?: number
  pageSize?: number
}

export function getUserList(params: UserListParams) {
  return request.post<PageResult<User>>('/api/admin/user-list', params)
}

export function editUser(data: { userId: number; status: number; validate: number; roleId: number }) {
  return request.post<any>('/api/admin/user-edit', data)
}

export function getAllRoleItem() {
  return request.get<any>('/api/admin/get-all-role-item')
}

export interface ArticlesListParams {
  page?: number
  pageSize?: number
  search?: string
  userId?: number
}

export function getArticlesList(params: ArticlesListParams) {
  return request.post<PageResult<Article>>('/api/admin/articles-list', params)
}

export function editArticle(data: { id: number; processStatus: number }) {
  return request.post<any>('/api/admin/article-edit', data)
}

export function getRoleList() {
  return request.post<PageResult<Role>>('/api/admin/role-list')
}

export function saveRole(data: any) {
  return request.post<any>('/api/admin/role-save', data)
}

export function deleteRole(id: number) {
  return request.post<any>('/api/admin/role-delete', { id })
}

export function getPermissionList() {
  return request.post<Permission[]>('/api/admin/get-permission-list')
}

export interface TrafficOverviewParams {
  startDate?: string
  endDate?: string
}

export function getTrafficOverview(params: TrafficOverviewParams) {
  return request.post<DailyTraffic[]>('/api/admin/traffic-overview', params)
}

export function getOptRecordPage(params: { page?: number; pageSize?: number }) {
  return request.post<any>('/api/admin/opt-record-page', params)
}

export function getSiteStatistics() {
  return request.get<any>('/api/forum/get-site-statistics')
}

export function getApplySheetList(params: any) {
  return request.post<any>('/api/admin/apply-sheet-list', params)
}

export function updateApplySheet(data: any) {
  return request.post<any>('/api/admin/apply-sheet-update', data)
}

export function getMailSettings() {
  return request.post<MailSettings>('/api/admin/mail-settings')
}

export function saveMailSettings(data: { settings: MailSettings }) {
  return request.post<any>('/api/admin/save-mail-settings', data)
}

export function testMailConnection(data: { host: string; port: number; username: string; password: string; useSSL: boolean }) {
  return request.post<any>('/api/admin/test-mail-connection', data)
}

export function getSiteSettings() {
  return request.post<SiteSettings>('/api/admin/site-settings')
}

export function saveSiteSettings(data: { settings: SiteSettings }) {
  return request.post<any>('/api/admin/save-site-settings', data)
}

export function getSecuritySettings() {
  return request.post<SecuritySettings>('/api/admin/security-settings')
}

export function saveSecuritySettings(data: { settings: SecuritySettings }) {
  return request.post<any>('/api/admin/save-security-settings', data)
}

export function getPostingSettings() {
  return request.post<PostingSettings>('/api/admin/posting-settings')
}

export function savePostingSettings(data: { settings: PostingSettings }) {
  return request.post<any>('/api/admin/save-posting-settings', data)
}

export function getFooterLinks() {
  return request.post<FooterConfig>('/api/admin/footer-links')
}

export function saveFooterLinks(data: { settings: FooterConfig }) {
  return request.post<any>('/api/admin/save-footer-links', data)
}

export function getSponsors() {
  return request.post<SponsorsConfig>('/api/admin/sponsors')
}

export function saveSponsors(data: { settings: SponsorsConfig }) {
  return request.post<any>('/api/admin/save-sponsors', data)
}

export function getAnnouncement() {
  return request.post<AnnouncementConfig>('/api/admin/announcement')
}

export function saveAnnouncement(data: { settings: AnnouncementConfig }) {
  return request.post<any>('/api/admin/save-announcement', data)
}
