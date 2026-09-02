import type {
  AccessGroupsPageProps,
  CategoryPageProps,
  CategoriesPageProps,
  DraftsPageProps,
  ErrorPageProps,
  HomeProps,
  LinksPageProps,
  LoginPageProps,
  MessagesPageProps,
  MembersPageProps,
  ModerationPageProps,
  NotificationsPageProps,
  PublishPageProps,
  ResetPasswordPageProps,
  SearchPageProps,
  SettingsPageProps,
  SponsorsPageProps,
  ThemePreviewProps,
  TopicDetailProps,
  UserProfileProps,
  PagePayload,
} from './payload.js'

export const pageComponents = [
  'home.index',
  'topic.detail',
  'user.profile',
  'category.index',
  'categories.index',
  'members.index',
  'links.index',
  'sponsors.index',
  'notifications.index',
  'messages.index',
  'drafts.index',
  'moderation.index',
  'settings.index',
  'access-groups.index',
  'theme.preview',
  'publish.index',
  'search.index',
  'auth.login',
  'auth.resetPassword',
  'error.index',
] as const

export interface PagePayloadMap {
  'home.index': HomeProps
  'topic.detail': TopicDetailProps
  'user.profile': UserProfileProps
  'category.index': CategoryPageProps
  'categories.index': CategoriesPageProps
  'members.index': MembersPageProps
  'links.index': LinksPageProps
  'sponsors.index': SponsorsPageProps
  'notifications.index': NotificationsPageProps
  'messages.index': MessagesPageProps
  'drafts.index': DraftsPageProps
  'moderation.index': ModerationPageProps
  'settings.index': SettingsPageProps
  'access-groups.index': AccessGroupsPageProps
  'theme.preview': ThemePreviewProps
  'publish.index': PublishPageProps
  'search.index': SearchPageProps
  'auth.login': LoginPageProps
  'auth.resetPassword': ResetPasswordPageProps
  'error.index': ErrorPageProps
}

/** Known page component names shipped by the core server. */
export type PageComponent = keyof PagePayloadMap & string

export type PageProps<TComponent extends PageComponent> = PagePayloadMap[TComponent]

export type AnyPagePayload = {
  [TComponent in PageComponent]: PagePayload<PagePayloadMap[TComponent], TComponent>
}[PageComponent]
