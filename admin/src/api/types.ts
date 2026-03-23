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
}

export interface Article {
  id: number
  title: string
  description: string
  type: number
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

export interface FooterItem {
  content: string
}

export interface FooterGroup {
  title: string
  links: { name: string; url: string }[]
}

export interface FooterConfig {
  primary: FooterItem[]
  list: FooterGroup[]
}

export interface SponsorItem {
  name: string
  logo: string
  info: string
  url: string
  tag: string[]
}

export interface UserSponsor {
  name: string
  amount: string
  time: string
}

export interface SponsorsConfig {
  sponsors: {
    level0: SponsorItem[]
    level1: SponsorItem[]
    level2: SponsorItem[]
    level3: SponsorItem[]
  }
  users: UserSponsor[]
}

export interface AnnouncementConfig {
  enabled: boolean
  title: string
  content: string
  link?: string
}
