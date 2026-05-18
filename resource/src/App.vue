<script setup lang="ts">
import type {
  ArticleDetailProps,
  CategoryPageProps,
  ErrorPageProps,
  HomeProps,
  LinksPageProps,
  LoginPageProps,
  MessagesPageProps,
  NotificationsPageProps,
  PagePayload,
  PublishPageProps,
  ResetPasswordPageProps,
  SearchPageProps,
  SettingsPageProps,
  SponsorsPageProps,
  UserProfileProps,
} from './types/payload'
import { provideI18n } from './runtime/i18n'
import { asyncPageComponents } from './runtime/page-registry'

const {
  HomePage,
  ArticlePage,
  UserPage,
  CategoryPage,
  LinksPage,
  SponsorsPage,
  NotificationsPage,
  MessagesPage,
  SettingsPage,
  PublishPage,
  SearchPage,
  LoginPage,
  ResetPasswordPage,
  ErrorPage,
} = asyncPageComponents

defineProps<{
  payload: PagePayload
}>()

provideI18n()
</script>

<template>
  <HomePage
    v-if="payload.component === 'home.index'"
    :layout="payload.layout"
    :props="payload.props as HomeProps"
  />
  <ArticlePage
    v-else-if="payload.component === 'article.detail'"
    :layout="payload.layout"
    :props="payload.props as ArticleDetailProps"
  />
  <UserPage
    v-else-if="payload.component === 'user.profile'"
    :layout="payload.layout"
    :props="payload.props as UserProfileProps"
  />
  <CategoryPage
    v-else-if="payload.component === 'category.index'"
    :layout="payload.layout"
    :props="payload.props as CategoryPageProps"
  />
  <LinksPage
    v-else-if="payload.component === 'links.index'"
    :layout="payload.layout"
    :props="payload.props as LinksPageProps"
  />
  <SponsorsPage
    v-else-if="payload.component === 'sponsors.index'"
    :layout="payload.layout"
    :props="payload.props as SponsorsPageProps"
  />
  <NotificationsPage
    v-else-if="payload.component === 'notifications.index'"
    :layout="payload.layout"
    :props="payload.props as NotificationsPageProps"
  />
  <MessagesPage
    v-else-if="payload.component === 'messages.index'"
    :layout="payload.layout"
    :props="payload.props as MessagesPageProps"
  />
  <SettingsPage
    v-else-if="payload.component === 'settings.index'"
    :layout="payload.layout"
    :props="payload.props as SettingsPageProps"
  />
  <PublishPage
    v-else-if="payload.component === 'publish.index'"
    :layout="payload.layout"
    :props="payload.props as PublishPageProps"
  />
  <SearchPage
    v-else-if="payload.component === 'search.index'"
    :layout="payload.layout"
    :props="payload.props as SearchPageProps"
  />
  <LoginPage
    v-else-if="payload.component === 'auth.login'"
    :layout="payload.layout"
    :props="payload.props as LoginPageProps"
  />
  <ResetPasswordPage
    v-else-if="payload.component === 'auth.resetPassword'"
    :layout="payload.layout"
    :props="payload.props as ResetPasswordPageProps"
  />
  <ErrorPage
    v-else-if="payload.component === 'error.notFound'"
    :layout="payload.layout"
    :props="payload.props as ErrorPageProps"
  />
</template>
