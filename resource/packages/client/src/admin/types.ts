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
  iconUrl: string
  color: string
  level: string
  isEnabled: boolean
  isWearable: boolean
  sortOrder: number
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
