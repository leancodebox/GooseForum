import type {
  ModerationLogListResponse,
  ModerationReportListResponse,
  NotificationFilter,
  NotificationListResponse,
  PostWindowPayload,
  SiteThemeConfig,
  UnreadStatusPayload,
  UserCardPayload,
} from '../contracts/payload.js'
import type { GooseApiResult } from '../http/client.js'

export type ToggleAction = 1 | 2

export interface CreatePostInput {
  topicId: number
  content: string
  replyToPostId?: number
}

export interface CreatePostResult {
  id: number
  postNo?: number
  renderedContent: string
}

export interface UpdatePostInput {
  postId: number
  content: string
}

export interface UpdatePostResult {
  id: number
  postNo?: number
  content: string
  renderedContent: string
  updatedAt: string
}

export interface PostWindowInput {
  topicId: number
  anchorPostId?: number
  anchorPostNo?: number
  beforePostNo?: number
  afterPostNo?: number
  limit?: number
}

export interface SubmitTopicInput {
  topicId: number
  title: string
  content: string
  categoryId: number[]
  topicStatus: 0 | 1
}

export interface JoinableAccessGroup {
  id: number
  name: string
  categories: string[]
  status: number
}

export interface ManagedAccessGroup {
  id: number
  name: string
  applications: Array<{ id: number; userId: number; username: string }>
}

export interface ChatMessagePayload {
  id: number
  senderId: number
  content: string
  msgType: number
  isRead: number
  createdAt: string
  isSelf: boolean
}

export interface ChatMessagesResponse {
  list: ChatMessagePayload[]
  hasMoreBefore: boolean
  hasMoreAfter: boolean
  nextBeforeId: number
  latestId: number
}

export interface ChatMessagesInput {
  convId: number
  beforeId?: number
  afterId?: number
  limit?: number
}

export interface SaveUserInfoInput {
  nickname: string
  locale?: string
  bio: string
  signature: string
  website: string
  websiteName: string
  externalInformation: Record<string, { link?: string }>
}

export interface OAuthBindingPayload {
  bound: boolean
  provider?: string
  createdAt?: string
  updatedAt?: string
}

export type OAuthBindingsPayload = Record<string, OAuthBindingPayload>

export interface CaptchaPayload {
  captchaId: string
  captchaImg: string
}

export interface LoginPublicKeyPayload {
  publicKey: string
  serverTs: number
}

export interface LoginInput {
  username: string
  encryptedPassword: string
  captchaId: string
  captchaCode: string
}

export interface RegisterInput {
  username: string
  email: string
  password: string
  captchaId: string
  captchaCode: string
  locale?: string
}

export interface ImageUploadInitResult {
  mode: 'proxy' | 'direct'
  name?: string
  upload?: {
    url: string
    method: string
    fields: Record<string, string>
    expiresAt: string
  }
}

export interface GooseSiteApi {
  accessGroups: {
    list(): Promise<JoinableAccessGroup[]>
    apply(groupId: number): Promise<boolean>
    managed(): Promise<ManagedAccessGroup[]>
    review(groupId: number, memberId: number, approve: boolean): Promise<boolean>
  }
  posts: {
    create(input: CreatePostInput): Promise<CreatePostResult | number | boolean>
    update(input: UpdatePostInput): Promise<UpdatePostResult>
    delete(postId: number): Promise<boolean>
    window(input: PostWindowInput): Promise<PostWindowPayload>
  }
  topics: {
    like(topicId: number, action: ToggleAction): Promise<boolean>
    bookmark(topicId: number, action: ToggleAction): Promise<boolean>
    watch(topicId: number, action: ToggleAction): Promise<boolean>
    setStatus(topicId: number, topicStatus: 0 | 1): Promise<boolean>
    write(input: SubmitTopicInput): Promise<number>
  }
  moderation: {
    setTopicStatus(topicId: number, action: 'ban' | 'unban'): Promise<boolean>
    setPostStatus(postId: number, action: 'ban' | 'unban'): Promise<boolean>
    report(targetType: 'topic' | 'post', targetId: number, reason: string, note: string): Promise<boolean>
    reports(cursor?: number, pageSize?: number, status?: string): Promise<ModerationReportListResponse>
    setReportStatus(id: number, action: 'ban' | 'resolve' | 'reject'): Promise<boolean>
    logs(cursor?: number, pageSize?: number): Promise<ModerationLogListResponse>
  }
  notifications: {
    list(filter: NotificationFilter, cursor?: number, limit?: number): Promise<NotificationListResponse>
    markRead(notificationId: number): Promise<boolean>
    markAllRead(): Promise<boolean>
    unread(): Promise<UnreadStatusPayload>
  }
  users: {
    card(userId: number): Promise<UserCardPayload>
    follow(userId: number, following: boolean): Promise<boolean>
    saveInfo(input: SaveUserInfoInput): Promise<void>
    saveCover(profileCoverUrl: string): Promise<void>
    savePresetAvatar(avatarUrl: string): Promise<{ avatarUrl?: string }>
    wearBadge(badgeCode: string): Promise<void>
    saveEmail(email: string): Promise<void>
    resendActivationEmail(): Promise<GooseApiResult<void>>
    saveUsername(username: string): Promise<void>
    changePassword(oldPassword: string, newPassword: string): Promise<void>
    oauthBindings(): Promise<OAuthBindingsPayload>
    unbindOAuth(provider: string): Promise<void>
  }
  chat: {
    messages(input: ChatMessagesInput): Promise<ChatMessagesResponse>
    send(peerId: number, content: string): Promise<{ convId: number }>
    markRead(convId: number): Promise<boolean>
  }
  auth: {
    captcha(): Promise<CaptchaPayload>
    loginPublicKey(): Promise<LoginPublicKeyPayload>
    login(input: LoginInput): Promise<void>
    register(input: RegisterInput): Promise<GooseApiResult<void>>
    forgotPassword(email: string, captchaId: string, captchaCode: string): Promise<GooseApiResult<void>>
    resetPassword(token: string, newPassword: string): Promise<GooseApiResult<void>>
    logout(): Promise<void>
  }
  themes: {
    save(settings: SiteThemeConfig): Promise<SiteThemeConfig>
    publish(): Promise<SiteThemeConfig>
  }
  uploads: {
    image(file: File): Promise<string>
    avatar(avatar: Blob | Blob[]): Promise<string>
  }
}
