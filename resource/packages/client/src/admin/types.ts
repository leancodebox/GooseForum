export interface PageResult<T> {
  list: T[]
  total: number
  hasNext?: boolean
  page: number
  pageSize?: number
  size?: number
}

export interface AdminCategoryModerator {
  id: number
  userId: number
  username: string
  avatarUrl?: string | null
  status: number
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

export interface AccessGroupGrant {
  categoryId: number
  level: number
}

export interface AccessGroupMember {
  id: number
  userId: number
  username: string
  avatarUrl?: string | null
  memberRole: 'member' | 'manager'
  status: number
}

export interface AccessGroup {
  id: number
  name: string
  systemKey?: string
  joinMode: 'system' | 'invite_only' | 'application'
  status: number
  members: AccessGroupMember[]
  grants: AccessGroupGrant[]
}

export interface AccessCategory {
  id: number
  name: string
  color: string
  isRestricted: boolean
  multiCategoryTopicCount: number
}

export interface AccessControlOverview {
  groups: AccessGroup[]
  categories: AccessCategory[]
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
  roleList?: { name: string; value: number }[] | null
  createTime: string
  lastActiveTime?: string | null
  badges?: UserBadge[]
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

export interface AdminOptRecord {
  id: number
  optUserId: number
  optType: number
  targetType: number
  targetId: string
  optInfo: string
  optInfoPayload?: { messageCode?: string; params?: Record<string, unknown> }
  createdAt: string
}

export interface SiteSettings { siteName:string;siteUrl:string;siteLogo:string;siteEmail:string;siteDescription:string;siteKeywords:string;externalLinks?:string }
export interface SiteChromeItem { id:string;enabled:boolean;type:'link'|'text'|string;label:string;i18nLabel:string;url:string }
export interface SiteChromeGroup { id:string;title:string;i18nLabel:string;items:SiteChromeItem[] }
export interface SiteChromeConfig { header:SiteChromeItem[];mainMenu:SiteChromeItem[];resources:SiteChromeItem[];sidebarGroups:SiteChromeGroup[];footerInfo?:{primary:{content:string}[];list:{name:string;url:string}[]};brandType?:string;brandText?:string;brandImage?:string }
export interface MailSettings { enableMail:boolean;smtpHost:string;smtpPort:number;useSSL:boolean;smtpUsername:string;smtpPassword:string;fromName:string;fromEmail:string }
export interface SecuritySettings { enableSignup:boolean;enableEmailVerification:boolean;allowedDomains:string[] }
export interface PostingSettings { externalLinks?:{enabled:boolean;whitelist:string[]};textControl:{minPostLength:number;maxPostLength:number;minTitleLength:number;maxTitleLength:number;newUserPostCooldownMinutes:number;maxDailyTopicsPerUser:number};uploadControl:{allowAttachments:boolean;authorizedExtensions:string[];maxAttachmentSizeKb:number;maxDailyUploadsPerUser:number;newUserUploadCooldownMinutes:number} }
export interface AnnouncementConfig { enabled:boolean;content:string }
export interface HttpNotifyEndpoint { id:string;name:string;enabled:boolean;url:string;secret:string;events:string[];timeoutSeconds:number;failureCount:number;lastError:string;abnormalTerminated:boolean }
export interface HttpNotifySettings { enabled:boolean;endpoints:HttpNotifyEndpoint[] }
export interface SensitiveWordSettings { enabled:boolean;mode:'after_review'|'visible_then_review' }
export interface SensitiveWord { id:number;word:string;action:'reject'|'replace'|'record';replacement:string;enabled:boolean }
export interface OAuthProviderSettings { key:string;displayName:string;kind:string;enabled:boolean;clientId:string;clientSecret?:string;clientSecretConfigured:boolean;clearClientSecret?:boolean;callbackUrl:string;discoveryUrl?:string;scopes?:string[] }
export interface OAuthSettings { providers:OAuthProviderSettings[] }
export type OIDCClientAuthMethod='none'|'client_secret_basic'|'client_secret_post'
export interface OIDCProviderStatus { enabled:boolean;available:boolean;issuer?:string;error?:string }
export interface OIDCClient { clientId:string;name:string;redirectUris:string[];scopes:string[];grantTypes:string[];tokenEndpointAuthMethod:OIDCClientAuthMethod;requirePkce:boolean;public:boolean;enabled:boolean }
export type OIDCClientInput=Omit<OIDCClient,'clientId'>
export interface OIDCClientCredentials { client:OIDCClient;clientSecret?:string }
export interface SiteStatistics { userCount:number;userMonthCount:number;topicMaxId:number;topicMonthCount:number;postMaxId:number;linksCount:number }
export interface DailyTraffic { date:string;regCount:number;topicCount:number;replyCount:number }
export interface ServerVersion { version:string;commit:string;buildDate:string;mode:'development'|'snapshot'|'release'|'custom'|string }
export interface GithubRelease { id:number;tag_name:string;published_at:string;body:string;html_url:string;prerelease:boolean;draft:boolean }

export interface UserBadge extends AdminBadge {
  source?: string
  reason?: string
  grantedAt?: string
}

export interface UserBadgeOptions {
  options: AdminBadge[]
  active: UserBadge[]
}

export interface AdminTopic {
  moderationStatus: string
  moderationReason: string
  moderationVersion: number
  moderatedAt?: string
  id: number
  title: string
  description?: string | null
  categoryId: number[]
  userId: number
  username: string
  userAvatarUrl?: string | null
  topicStatus: number
  processStatus: number
  deleted: boolean
  viewCount: number
  replyCount: number
  likeCount: number
  pinWeight: number
  createdAt: string
  updatedAt?: string
}

export interface TopicSource extends AdminTopic {
  content: string
}

export interface ReviewPost {
  id: number
  topicId: number
  topicTitle: string
  userId: number
  postNo: number
  content: string
  moderationStatus: string
  moderationReason: string
  moderationVersion: number
  processStatus: number
  updatedAt: string
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

export type SponsorLevel = 'level0' | 'level1' | 'level2' | 'level3'

export interface SponsorsConfig {
  sponsors: Record<SponsorLevel, SponsorItem[]>
  content: { title: string; description: string }
  contact: { title: string; description: string; buttonText: string; buttonLink: string }
  rules: { content: string }[]
}

export interface AdminImageUploadResult {
  url?: string
  filename?: string
  size?: number
}

export interface AdminPermissionOption {
  name: string
  label?: string
  value: number
}

export interface AdminRole {
  roleId: number
  roleName: string
  effective: number
  permissions: { id: number; name: string }[]
  createTime: string
}

export interface GooseAdminApi {
  dashboard: {
    statistics(): Promise<SiteStatistics>
    traffic(startDate?: string, endDate?: string): Promise<DailyTraffic[]>
    version(): Promise<ServerVersion>
    releases(): Promise<GithubRelease[]>
  }
  settings: {
    site(): Promise<SiteSettings>
    saveSite(settings: SiteSettings): Promise<unknown>
    chrome(): Promise<SiteChromeConfig>
    saveChrome(settings: SiteChromeConfig): Promise<unknown>
    mail(): Promise<MailSettings>
    saveMail(settings: MailSettings): Promise<unknown>
    testMail(settings: MailSettings, testEmail: string): Promise<unknown>
    security(): Promise<SecuritySettings>
    saveSecurity(settings: SecuritySettings): Promise<unknown>
    posting(): Promise<PostingSettings>
    savePosting(settings: PostingSettings): Promise<unknown>
    announcement(): Promise<AnnouncementConfig>
    saveAnnouncement(settings: AnnouncementConfig): Promise<unknown>
    httpNotify(): Promise<HttpNotifySettings>
    saveHttpNotify(settings: HttpNotifySettings): Promise<unknown>
    sensitiveWordSettings(): Promise<SensitiveWordSettings>
    saveSensitiveWordSettings(settings: SensitiveWordSettings): Promise<unknown>
    sensitiveWords(): Promise<SensitiveWord[]>
    saveSensitiveWord(word: SensitiveWord): Promise<SensitiveWord>
    deleteSensitiveWord(id: number): Promise<unknown>
    oauth(): Promise<OAuthSettings>
    saveOAuth(settings: OAuthSettings): Promise<OAuthSettings>
    oidcStatus(): Promise<OIDCProviderStatus>
    saveOIDCStatus(enabled: boolean): Promise<OIDCProviderStatus>
    rotateOIDCSigningKey(): Promise<OIDCProviderStatus>
    oidcClients(): Promise<OIDCClient[]>
    createOIDCClient(input: OIDCClientInput): Promise<OIDCClientCredentials>
    updateOIDCClient(client: OIDCClient): Promise<OIDCClient>
    rotateOIDCClientSecret(clientId: string): Promise<OIDCClientCredentials>
  }
  audit: {
    records(input: { page?: number; pageSize?: number; optUserId?: number; optType?: number; targetType?: number; targetId?: number }): Promise<PageResult<AdminOptRecord>>
  }
  assets: {
    badges(): Promise<AdminBadge[]>
    saveBadge(badge: AdminBadge): Promise<unknown>
    deleteBadge(code: string): Promise<unknown>
    files(input: { page?: number; pageSize?: number }): Promise<PageResult<AdminFileResource>>
  }
  pages: {
    links(): Promise<FriendLinkGroup[]>
    saveLinks(groups: FriendLinkGroup[]): Promise<unknown>
    sponsors(): Promise<SponsorsConfig>
    saveSponsors(config: SponsorsConfig): Promise<unknown>
    uploadImage(file: File): Promise<AdminImageUploadResult>
  }
  topics: {
    list(input: { page?: number; pageSize?: number; search?: string; moderationStatus?: string; categoryId?: number; userId?: number }): Promise<PageResult<AdminTopic>>
    reviewPosts(input: { page: number; pageSize: number; moderationStatus?: string; search?: string }): Promise<PageResult<ReviewPost>>
    source(topicId: number): Promise<TopicSource>
    setProcessStatus(topicId: number, processStatus: number): Promise<unknown>
    delete(topicId: number): Promise<unknown>
    setPin(topicId: number, pinWeight: number): Promise<unknown>
    setCategories(topicId: number, categoryId: number[]): Promise<unknown>
    review(kind: 'topic' | 'post', input: { id: number; version: number; action: 'approve' | 'reject' | 'recheck'; reason: string }): Promise<boolean>
  }
  users: {
    list(input: { page?: number; pageSize?: number; username?: string; userId?: number; email?: string }): Promise<PageResult<AdminUser>>
    edit(input: { userId: number; status: number; validate: number; roleId: number }): Promise<unknown>
    roles(): Promise<{ name: string; value: number }[]>
    badgeOptions(userId: number): Promise<UserBadgeOptions>
    saveBadges(userId: number, badgeCodes: string[]): Promise<unknown>
  }
  roles: {
    list(): Promise<PageResult<AdminRole>>
    permissions(): Promise<AdminPermissionOption[]>
    save(input: { id: number; roleName: string; permissions: number[] }): Promise<unknown>
    delete(id: number): Promise<unknown>
  }
  accessGroups: {
    overview(): Promise<AccessControlOverview>
    save(input: { id?: number; name: string; joinMode: 'invite_only' | 'application'; status: number }): Promise<number>
    delete(id: number): Promise<unknown>
    saveMember(input: { groupId: number; userId?: number; username?: string; memberRole: 'member' | 'manager' }): Promise<number>
    deleteMember(groupId: number, memberId: number): Promise<unknown>
    reviewApplication(groupId: number, memberId: number, approve: boolean): Promise<unknown>
  }
  categories: {
    list(): Promise<AdminCategory[]>
    save(category: AdminCategory): Promise<unknown>
    delete(id: number): Promise<unknown>
    access(): Promise<AccessControlOverview>
    saveAccess(categoryId: number, grants: { accessGroupId: number; level: number }[]): Promise<unknown>
    addModerator(categoryId: number, user: { userId?: number; username?: string }): Promise<unknown>
    deleteModerator(id: number): Promise<unknown>
  }
  moderators: {
    list(): Promise<AdminCategoryModerator[]>
    add(user: { userId?: number; username?: string }): Promise<unknown>
    delete(id: number): Promise<unknown>
  }
}
