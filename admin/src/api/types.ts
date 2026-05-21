export interface ApiResponse<T = any> {
  code: number
  msg: string
  result: T
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

export interface Category {
  id: number
  category: string
  desc?: string
  icon?: string
  color?: string
  slug?: string
  sort?: number
  status?: number
}

export interface User {
  userId: number
  username: string
  avatarUrl: string
  email: string
  status: number
  validate: number
  prestige: number
  roleList: { name: string; value: number }[]
  roleId: number
  createTime: string
  lastActiveTime: string
  badges?: UserBadge[]
}

export interface BadgeItem {
  code: string
  type: 'system' | 'custom'
  grantMode: 'auto' | 'manual'
  name: string
  description: string
  iconType: string
  iconKey: string
  iconUrl: string
  color: string
  level: string
  isEnabled: boolean
  sortOrder: number
  isSystem?: boolean
  canDelete?: boolean
}

export interface UserBadge extends BadgeItem {
  source: string
  reason: string
  grantedAt: string
}

export interface UserBadgeOptions {
  options: BadgeItem[]
  active: UserBadge[]
}

export interface Article {
  id: number
  title: string
  description: string
  type: number
  categoryId: number[]
  userId: number
  username: string
  userAvatarUrl: string
  articleStatus: number
  processStatus: number
  viewCount: number
  replyCount: number
  likeCount: number
  createdAt: string
  updatedAt: string
}

export interface ArticleSource extends Article {
  content: string
}

export interface Role {
  roleId: number
  roleName: string
  effective: number
  permissions: { id: number; name: string }[]
  createTime: string
}

export interface Permission {
  id: number
  name: string
}

export interface DailyTraffic {
  date: string
  regCount: number
  articleCount: number
  replyCount: number
}

export interface SiteSettings {
  siteName: string
  siteLogo: string
  siteDescription: string
  siteKeywords: string
  siteUrl: string
  siteEmail: string
  externalLinks?: string
  footerInfo?: FooterInfo
  brandType?: 'default' | 'text' | 'image'
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
  mustApproveUsers: boolean
  minPasswordLength: number
  inviteOnly: boolean
  restrictions: {
    allowedDomains: string[]
    blockedDomains: string[]
    maxRegistrationsPerIp: number
  }
}

export interface PostingSettings {
  textControl: {
    minPostLength: number
    maxPostLength: number
    minTitleLength: number
    maxTitleLength: number
    allowUppercasePosts: boolean
  }
  uploadControl: {
    allowAttachments: boolean
    authorizedExtensions: string[]
    maxAttachmentSizeKb: number
    maxAttachmentsPerPost: number
  }
  editControl: {
    editingGracePeriod: number
    postEditTimeLimit: number
    allowUsersToDeletePosts: boolean
  }
}

export interface PItem {
  content: string
}

export interface FooterItem {
  name: string
  url: string
}

export interface FooterInfo {
  primary: PItem[]
  list: FooterItem[]
}

export interface SponsorItem {
  name: string
  avatarUrl: string
  message: string
  link: string
}

export interface SponsorsConfig {
  sponsors: {
    level0: SponsorItem[]
    level1: SponsorItem[]
    level2: SponsorItem[]
    level3: SponsorItem[]
  }
  content: SponsorsPageIntro
  contact: SponsorsContact
  rules: SponsorsRule[]
}

export interface SponsorsPageIntro {
  title: string
  description: string
}

export interface SponsorsContact {
  title: string
  description: string
  buttonText: string
  buttonLink: string
}

export interface SponsorsRule {
  content: string
}

export interface AnnouncementConfig {
  enabled: boolean
  content: string
}
