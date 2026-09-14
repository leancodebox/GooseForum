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
}

export interface GooseAdminApi {
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
  users: {
    list(input: { page?: number; pageSize?: number; username?: string; userId?: number; email?: string }): Promise<PageResult<AdminUser>>
  }
}
