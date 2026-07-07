import type { LayoutPayload } from '@/types/payload'

export interface AdminPayload<TProps = unknown> {
  component: string
  props: TProps
  meta: {
    title: string
    robots?: string
  }
  layout: LayoutPayload
  url: string
  version: string
}

export type ManageHomeProps = Record<string, never>

export interface ApiEnvelope<T> {
  code?: number
  messageCode?: string
  params?: Record<string, unknown>
  result?: T
  data?: T
}

export interface PageResult<T> {
  list: T[]
  total: number
  hasNext?: boolean
  page: number
  pageSize?: number
  size?: number
}

export interface AdminCategory {
  id: number
  category: string
  desc?: string
  icon?: string
  color?: string
  slug?: string
  sort?: number
  moderators?: AdminCategoryModerator[]
}

export interface AdminCategoryModerator {
  id: number
  userId: number
  username: string
  avatarUrl?: string
  status: number
}

export interface AdminUser {
  userId: number
  username: string
  avatarUrl?: string | null
  email: string
  status: number
  validate: number
  prestige: number
  roleId?: number | null
  roleList?: { name: string, value: number }[] | null
  createTime: string
  lastActiveTime?: string | null
  badges?: UserBadge[]
}

export interface UserBadge extends AdminBadge {
  source?: string
  reason?: string
  grantedAt?: string
}

export interface UserBadgeOptions {
  options: AdminBadge[]
  active: UserBadge[]
}

export interface AdminRole {
  roleId: number
  roleName: string
  effective: number
  permissions: { id: number, name: string }[]
  createTime: string
}

export interface AdminPermissionOption {
  name: string
  label?: string
  value: number
}

export interface AdminTopic {
  id: number
  title: string
  description?: string | null
  type: number
  categoryId: number[]
  userId: number
  username: string
  userAvatarUrl?: string | null
  topicStatus: number
  processStatus: number
  viewCount: number
  replyCount: number
  likeCount: number
  pinWeight: number
  createdAt: string
  updatedAt?: string
}

export interface AdminOptRecord {
  id: number
  optUserId: number
  optType: number
  targetType: number
  targetId: string
  optInfo: string
  optInfoPayload?: {
    messageCode?: string
    params?: Record<string, unknown>
  }
  createdAt: string
}

export interface AdminFileResource {
  id: number
  name: string
  type: string
  size: number
  userId: number
  uploaderUsername?: string
  createdAt: string
  url: string
}

export interface TopicSource extends AdminTopic {
  content: string
}

export interface AdminBadge {
  code: string
  type: 'system' | 'custom' | string
  grantMode: 'auto' | 'manual' | string
  name: string
  description: string
  iconType?: string
  iconKey?: string
  iconUrl: string
  color: string
  level: string
  isEnabled: boolean
  isWearable: boolean
  sortOrder: number
  isSystem?: boolean
  canDelete?: boolean
}

export interface FriendLink {
  name: string
  url: string
  desc?: string
  logoUrl?: string
  status?: number
}

export interface FriendLinkGroup {
  name: string
  emoji?: string
  color?: string
  links: FriendLink[]
}

export interface SponsorItem {
  name: string
  avatarUrl: string
  message: string
  link: string
}

export interface SponsorsConfig {
  sponsors: Record<'level0' | 'level1' | 'level2' | 'level3', SponsorItem[]>
  content: { title: string, description: string }
  contact: { title: string, description: string, buttonText: string, buttonLink: string }
  rules: { content: string }[]
}

export interface SiteSettings {
  siteName: string
  siteUrl: string
  siteLogo: string
  siteEmail: string
  siteDescription: string
  siteKeywords: string
  externalLinks?: string
}

export interface SiteChromeItem {
  id: string
  enabled: boolean
  type: 'link' | 'text' | string
  label: string
  i18nLabel: string
  url: string
}

export interface SiteChromeGroup {
  id: string
  title: string
  i18nLabel: string
  items: SiteChromeItem[]
}

export interface SiteChromeConfig {
  header: SiteChromeItem[]
  mainMenu: SiteChromeItem[]
  resources: SiteChromeItem[]
  sidebarGroups: SiteChromeGroup[]
  footerInfo?: {
    primary: { content: string }[]
    list: { name: string, url: string }[]
  }
  brandType?: string
  brandText?: string
  brandImage?: string
}

export interface MailSettings {
  enableMail: boolean
  smtpHost: string
  smtpPort: number
  useSSL: boolean
  smtpUsername: string
  smtpPassword: string
  fromName: string
  fromEmail: string
}

export interface SecuritySettings {
  enableSignup: boolean
  enableEmailVerification: boolean
  allowedDomains: string[]
}

export interface PostingSettings {
  textControl: {
    minPostLength: number
    maxPostLength: number
    minTitleLength: number
    maxTitleLength: number
    newUserPostCooldownMinutes: number
  }
  uploadControl: {
    allowAttachments: boolean
    authorizedExtensions: string[]
    maxAttachmentSizeKb: number
    maxDailyUploadsPerUser: number
    newUserUploadCooldownMinutes: number
  }
}

export interface HttpNotifyEndpoint {
  id: string
  name: string
  enabled: boolean
  url: string
  secret: string
  events: string[]
  timeoutSeconds: number
  failureCount: number
  lastError: string
  abnormalTerminated: boolean
}

export interface HttpNotifySettings {
  enabled: boolean
  endpoints: HttpNotifyEndpoint[]
}

export interface AnnouncementConfig {
  enabled: boolean
  content: string
}

export interface SiteStatistics {
  userCount: number
  userMonthCount: number
  articleCount: number
  articleMonthCount: number
  reply: number
  linksCount: number
}

export interface DailyTraffic {
  date: string
  regCount: number
  articleCount: number
  replyCount: number
}

export interface ServerVersion {
  version: string
  commit: string
  buildDate: string
  mode: 'development' | 'snapshot' | 'release' | 'custom' | string
}

export interface GithubRelease {
  id: number
  tag_name: string
  published_at: string
  body: string
  html_url: string
  prerelease: boolean
  draft: boolean
}
