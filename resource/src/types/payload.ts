export interface PagePayload<TProps = unknown> {
  component: string
  props: TProps
  meta: PageMeta
  layout: LayoutPayload
  url: string
  version: string
}

export interface PageMeta {
  title: string
  description?: string
  canonical?: string
  robots?: string
  openGraph?: {
    title?: string
    description?: string
    type?: string
    url?: string
    siteName?: string
    image?: string
    publishedTime?: string
    modifiedTime?: string
    author?: string
    section?: string
    tags?: string[]
  }
  twitter?: {
    card?: string
    title?: string
    description?: string
    image?: string
  }
  jsonLd?: unknown
}

export interface ErrorPageProps {
  code: string
  title: string
  messageCode?: string
  params?: Record<string, unknown>
}

export interface LoginPageProps {
  initialMode: 'login' | 'register' | 'forgot'
  redirectUrl: string
  githubUrl: string
  googleReady: boolean
}

export interface ResetPasswordPageProps {
  token: string
}

export interface LayoutPayload {
  site: SitePayload
  viewer: ViewerPayload
  sidebar: SidebarPayload
  footer: FooterPayload
  unread: UnreadStatusPayload
}

export interface UnreadStatusPayload {
  notifications: boolean
  messages: boolean
  latestNotificationType?: string
}

export interface SitePayload {
  name: string
  description: string
  logo: string
  favicon: string
  externalLinks?: string
  brandType: string
  brandText: string
  brandImage: string
}

export interface ViewerPayload {
  id: number
  username: string
  email: string
  avatarUrl: string
  isAuthenticated: boolean
  isAdmin: boolean
  requiresEmailVerification: boolean
  adminPermissions: number[]
}

export interface CategoryNavPayload {
  id: number
  label: string
  url: string
  color: string
}

export interface NavItemPayload {
  key: string
  label: string
  icon?: string
  url: string
}

export interface SidebarPayload {
  main?: NavItemPayload[]
  resources?: NavItemPayload[]
  categories: CategoryNavPayload[]
  activeKey: string
}

export interface FooterPayload {
  links: Array<{ name: string; url: string }>
  primary: string[]
}

export interface HomeProps {
  sort: string
  tabs: Array<{ key: string; label?: string; url: string; active: boolean }>
  topics: TopicPayload[]
  pagination: {
    page: number
    nextPage: number
    hasNext: boolean
    nextUrl: string
  }
  announcement: {
    enabled: boolean
    html: string
  }
}

export interface ArticleDetailProps {
  article: ArticlePayload
  replies: ReplyPayload[]
  hotTopics: TopicPayload[]
  permissions: {
    isOwnArticle: boolean
    canReply: boolean
  }
}

export interface ArticlePayload {
  id: number
  title: string
  description: string
  url: string
  html: string
  articleStatus: number
  author: {
    id: number
    username: string
    avatarUrl: string
  }
  participants: Array<{ id: number; username: string; avatarUrl: string }>
  categories: Array<{ id: number; name: string; url: string; color: string }>
  replyCount: number
  viewCount: number
  likeCount: number
  isLiked: boolean
  isBookmarked: boolean
  isWatched: boolean
  createdAt: string
  updatedAt: string
}

export interface ReplyPayload {
  id: number
  articleId: number
  content: string
  renderedContent: string
  author: {
    id: number
    username: string
    avatarUrl: string
  }
  createdAt: string
  replyToId?: number
  replyToUserId?: number
  replyToUsername?: string
  isOwnReply: boolean
  updatedAt?: string
}

export interface ReplyWindowPayload {
  replies: ReplyPayload[]
  anchorReplyId?: number
  beforeCursor?: number
  afterCursor?: number
  hasBefore: boolean
  hasAfter: boolean
  total: number
}

export interface TopicPayload {
  id: number
  title: string
  description: string
  url: string
  author: {
    id: number
    username: string
    avatarUrl: string
  }
  participants: Array<{ id: number; username: string; avatarUrl: string }>
  categories: Array<{ id: number; name: string; url: string; color: string }>
  replyCount: number
  viewCount: number
  pinWeight: number
  activityText: string
  lastUpdateTime: string
}

export interface UserCardPayload {
  userId: number
  username: string
  nickname: string
  avatarUrl: string
  profileCoverUrl: string
  bio: string
  signature: string
  websiteName: string
  website: string
  externalInformation: Record<string, { link?: string }>
  isAdmin: boolean
  articleCount: number
  replyCount: number
  likeReceivedCount: number
  likeGivenCount: number
  followerCount: number
  followingCount: number
  collectionCount: number
  isOnline: boolean
  isFollowing: boolean
  isSelf: boolean
  lastActiveTime: string
  createdAt: string
}

export interface UserHoverCardPayload {
  userId: number
  username: string
  nickname: string
  avatarUrl: string
  profileCoverUrl: string
  bio: string
  signature: string
  websiteName: string
  website: string
  externalInformation: Record<string, { link?: string }>
  isAdmin: boolean
  articleCount: number
  replyCount: number
  likeReceivedCount: number
  followerCount: number
  isOnline: boolean
  isFollowing: boolean
  badges: UserBadgePayload[]
  lastActiveTime: string
  createdAt: string
}

export interface UserProfileProps {
  user: UserCardPayload
  badges: UserBadgePayload[]
  topics: TopicPayload[]
  activities: UserActivityPayload[]
  following: UserConnectionPayload[]
  followers: UserConnectionPayload[]
  isOwnProfile: boolean
  canMessage: boolean
  canFollow: boolean
  messageUrl: string
  settingsUrl: string
}

export interface BadgePayload {
  code: string
  type: string
  grantMode: string
  name: string
  description: string
  iconType: string
  iconKey: string
  iconUrl: string
  color: string
  level: string
  isEnabled: boolean
  sortOrder: number
}

export interface UserBadgePayload extends BadgePayload {
  source: string
  reason: string
  grantedAt: string
}

export interface UserActivityPayload {
  id: number
  action: number
  subjectType: string
  subjectId: number
  contentPreview: string
  url: string
  label: string
  createdAt: string
}

export interface UserConnectionPayload {
  id: number
  username: string
  nickname: string
  avatarUrl: string
  bio: string
  url: string
}

export interface CategoryPageProps {
  category: {
    id: number
    name: string
    description: string
    icon: string
    color: string
    url: string
  }
  sort: string
  tabs: Array<{ key: string; label?: string; url: string; active: boolean }>
  topics: TopicPayload[]
  pagination: {
    page: number
    nextPage: number
    hasNext: boolean
    nextUrl: string
  }
}

export interface LinksPageProps {
  groups: LinkGroupPayload[]
  totalCount: number
}

export interface LinkGroupPayload {
  name: string
  emoji: string
  color: string
  links: FriendLinkPayload[]
}

export interface FriendLinkPayload {
  name: string
  desc: string
  url: string
  logoUrl: string
}

export interface SponsorsPageProps {
  sections: SponsorSectionPayload[]
  totalCount: number
  content: SponsorsPageIntroPayload
  contact: SponsorsContactPayload
  rules: SponsorsRulePayload[]
}

export interface SponsorSectionPayload {
  key: string
  label: string
  tone: string
  sponsors: SponsorPayload[]
}

export interface SponsorPayload {
  name: string
  message: string
  link: string
  avatarUrl: string
}

export interface SponsorsPageIntroPayload {
  title: string
  description: string
}

export interface SponsorsContactPayload {
  title: string
  description: string
  buttonText: string
  buttonLink: string
}

export interface SponsorsRulePayload {
  content: string
}

export interface NotificationsPageProps {
  total: number
  unreadCount: number
  notifications: NotificationPayload[]
  pagination: {
    page: number
    nextPage: number
    hasNext: boolean
    nextUrl: string
  }
}

export type NotificationFilter = 'all' | 'unread'

export interface NotificationListResponse {
  items: NotificationPayload[]
  nextCursor: number
  hasNext: boolean
  unreadCount: number
}

export type NotificationTemplateKey =
  | 'notifications.templates.comment'
  | 'notifications.templates.reply'
  | 'notifications.templates.articleComment'
  | 'notifications.templates.follow'
  | 'notifications.templates.badge'

export interface DraftsPageProps {
  total: number
  drafts: DraftPayload[]
  pagination: {
    page: number
    nextPage: number
    hasNext: boolean
    nextUrl: string
  }
}

export interface DraftPayload {
  id: number
  title: string
  description: string
  editUrl: string
  replyCount: number
  viewCount: number
  processStatus: number
  updatedAt: string
  createdAt: string
  categories: Array<{ id: number; name: string; url: string; color: string }>
}

export interface NotificationPayload {
  id: number
  eventType: string
  isRead: boolean
  createdAt: string
  title: string
  content: string
  actor: {
    id: number
    username: string
    avatarUrl?: string
  }
  article?: {
    id: number
    title: string
    url: string
  }
  payload: {
    title?: string
    content?: string
    templateKey?: NotificationTemplateKey
    templateParams?: NotificationTemplateParams
    actorId: number
    actorName?: string
    articleId?: number
    articleTitle?: string
    metadata?: {
      followerName?: string
      badgeCode?: string
      badgeName?: string
      badgeIconUrl?: string
      profileUrl?: string
    }
  }
}

export interface NotificationTemplateParams {
  preview?: string
  followerName?: string
  badgeCode?: string
  badgeName?: string
}

export interface MessagesPageProps {
  conversations: MessageConversationPayload[]
  suggestedUsers: UserConnectionPayload[]
}

export interface MessageConversationPayload {
  id: number
  peerId: number
  peerUsername: string
  peerAvatar: string
  lastMsg: string
  lastMsgTime: string
  unreadCount: number
  convId: number
  peerUrl: string
}

export interface SettingsPageProps {
  user: SettingsUserPayload
  stats: {
    articleCount: number
    replyCount: number
    followerCount: number
    followingCount: number
    likeReceivedCount: number
    likeGivenCount: number
    collectionCount: number
    createdAt: string
  }
  tabs: Array<{ key: string; label?: string; url: string; active: boolean }>
}

export interface SettingsUserPayload {
  id: number
  username: string
  email: string
  nickname: string
  avatarUrl: string
  profileCoverUrl: string
  bio: string
  signature: string
  websiteName: string
  website: string
  prestige: number
  createdAt: string
  externalInformation: Record<string, { link?: string }>
}

export interface PublishPageProps {
  articleId: number
  isEditing: boolean
  categories: PublishCategoryPayload[]
  types: PublishTypePayload[]
  article: {
    title: string
    content: string
    type: number
    categoryIds: number[]
    articleStatus: number
  }
}

export interface PublishCategoryPayload {
  id: number
  name: string
  color: string
}

export interface PublishTypePayload {
  name: string
  value: number
}

export interface SearchPageProps {
  query: string
  topics: TopicPayload[]
  total: number
  totalPages: number
  pagination: {
    page: number
    nextPage: number
    hasNext: boolean
    nextUrl: string
  }
}
