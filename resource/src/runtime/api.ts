import {
  createGooseClient,
  GooseClientError,
  type CaptchaPayload,
  type ChatMessagePayload,
  type ChatMessagesInput,
  type ChatMessagesResponse,
  type CreatePostResult,
  type JoinableAccessGroup,
  type ManagedAccessGroup,
  type ModerationLogListResponse,
  type ModerationReportListResponse,
  type NotificationFilter,
  type NotificationListResponse,
  type OAuthBindingsPayload,
  type PostWindowInput,
  type PostWindowPayload,
  type SaveUserInfoInput,
  type SubmitTopicInput,
  type UpdatePostResult,
  type UserCardPayload,
} from '@gooseforum/client'
import { i18n } from './i18n'
import { resolveApiMessage } from './api-message'

export type {
  ChatMessagePayload,
  ChatMessagesInput,
  ChatMessagesResponse,
  CreatePostResult,
  JoinableAccessGroup,
  ManagedAccessGroup,
  OAuthBindingsPayload,
  PostWindowInput,
  SaveUserInfoInput,
  SubmitTopicInput,
  UpdatePostResult,
} from '@gooseforum/client'

const client = createGooseClient()

class ApiResponseError extends Error {
  readonly messageCode?: string
  readonly status?: number

  constructor(message: string, messageCode?: string, status?: number) {
    super(message)
    this.name = 'ApiResponseError'
    this.messageCode = messageCode
    this.status = status
  }
}

function t(key: string) {
  return i18n.global.t(key)
}

async function localized<T>(request: Promise<T>, fallback: string): Promise<T> {
  try {
    return await request
  } catch (error) {
    if (error instanceof GooseClientError && error.messageCode) {
      throw new ApiResponseError(resolveApiMessage(error, fallback), error.messageCode, error.status)
    }
    throw error
  }
}

export function getJoinableAccessGroups(): Promise<JoinableAccessGroup[]> {
  return localized(client.api.accessGroups.list(), t('accessGroups.loadFailed'))
}

export function applyToAccessGroup(groupId: number): Promise<boolean> {
  return localized(client.api.accessGroups.apply(groupId), t('accessGroups.applicationFailed'))
}

export function getManagedAccessGroups(): Promise<ManagedAccessGroup[]> {
  return localized(client.api.accessGroups.managed(), t('accessGroups.loadFailed'))
}

export function reviewManagedAccessGroupApplication(groupId: number, memberId: number, approve: boolean): Promise<boolean> {
  return localized(client.api.accessGroups.review(groupId, memberId, approve), t('accessGroups.memberSaveFailed'))
}

export function createPost(topicId: number, content: string, replyToPostId = 0): Promise<CreatePostResult | number | boolean> {
  return localized(client.api.posts.create({ topicId, content, replyToPostId }), t('api.replyFailed'))
}

export function updatePost(postId: number, content: string): Promise<UpdatePostResult> {
  return localized(client.api.posts.update({ postId, content }), t('api.replyUpdateFailed'))
}

export function deletePost(postId: number): Promise<boolean> {
  return localized(client.api.posts.delete(postId), t('api.replyDeleteFailed'))
}

export function getPostWindow(input: PostWindowInput): Promise<PostWindowPayload> {
  return localized(client.api.posts.window(input), t('api.repliesLoadFailed'))
}

export function likeTopic(id: number, action: 1 | 2): Promise<boolean> {
  return localized(client.api.topics.like(id, action), t('api.likeFailed'))
}

export function bookmarkTopic(id: number, action: 1 | 2): Promise<boolean> {
  return localized(client.api.topics.bookmark(id, action), t('api.bookmarkFailed'))
}

export function watchTopic(id: number, action: 1 | 2): Promise<boolean> {
  return localized(client.api.topics.watch(id, action), t('api.watchFailed'))
}

export function updateTopicStatus(id: number, topicStatus: 0 | 1): Promise<boolean> {
  return localized(client.api.topics.setStatus(id, topicStatus), t('api.topicStatusFailed'))
}

export function updateModerationTopicStatus(id: number, action: 'ban' | 'unban'): Promise<boolean> {
  return localized(client.api.moderation.setTopicStatus(id, action), t('api.moderationActionFailed'))
}

export function submitReport(targetType: 'topic' | 'post', targetId: number, reason: string, note: string): Promise<boolean> {
  return localized(client.api.moderation.report(targetType, targetId, reason, note), t('api.reportFailed'))
}

export function updateModerationPostStatus(id: number, action: 'ban' | 'unban'): Promise<boolean> {
  return localized(client.api.moderation.setPostStatus(id, action), t('api.moderationActionFailed'))
}

export function fetchModerationReports(cursor = 0, pageSize = 20, status = 'open'): Promise<ModerationReportListResponse> {
  return localized(client.api.moderation.reports(cursor, pageSize, status), t('api.moderationReportsFailed'))
}

export function updateModerationReportStatus(id: number, action: 'ban' | 'resolve' | 'reject'): Promise<boolean> {
  return localized(client.api.moderation.setReportStatus(id, action), t('api.moderationActionFailed'))
}

export function fetchModerationLogs(cursor = 0, pageSize = 20): Promise<ModerationLogListResponse> {
  return localized(client.api.moderation.logs(cursor, pageSize), t('api.moderationLogsFailed'))
}

export function markAllNotificationsRead(): Promise<boolean> {
  return localized(client.api.notifications.markAllRead(), t('api.markReadFailed'))
}

export function markNotificationRead(notificationId: number): Promise<boolean> {
  return localized(client.api.notifications.markRead(notificationId), t('api.markReadFailed'))
}

export function fetchNotifications(filter: NotificationFilter, cursor = 0, limit = 20): Promise<NotificationListResponse> {
  return localized(client.api.notifications.list(filter, cursor, limit), t('api.notificationsLoadFailed'))
}

export async function getUserCard(userId: number): Promise<UserCardPayload> {
  const result = await localized(client.api.users.card(userId), t('api.userLoadFailed'))
  if (!result) throw new Error(t('api.userEmpty'))
  return result
}

export async function followUser(userId: number, isFollowing: boolean): Promise<boolean> {
  return (await localized(client.api.users.follow(userId, isFollowing), t('api.followFailed'))) ?? true
}

export async function submitTopic(topic: SubmitTopicInput): Promise<number> {
  return (await localized(client.api.topics.write(topic), t('api.topicSaveFailed'))) ?? topic.topicId
}

export function uploadImage(file: File): Promise<string> {
  return localized(client.api.uploads.image(file), t('api.imageUploadFailed'))
}

export async function getChatMessages(input: ChatMessagesInput): Promise<ChatMessagesResponse> {
  const result = await localized(client.api.chat.messages(input), t('api.messagesLoadFailed'))
  return {
    list: result?.list ?? [],
    hasMoreBefore: Boolean(result?.hasMoreBefore),
    hasMoreAfter: Boolean(result?.hasMoreAfter),
    nextBeforeId: result?.nextBeforeId ?? 0,
    latestId: result?.latestId ?? 0,
  }
}

export async function sendChatMessage(peerId: number, content: string): Promise<number> {
  const result = await localized(client.api.chat.send(peerId, content), t('api.sendFailed'))
  return result?.convId ?? 0
}

export async function markChatRead(convId: number): Promise<boolean> {
  return (await localized(client.api.chat.markRead(convId), t('api.markReadFailed'))) ?? true
}

export async function saveUserInfo(input: SaveUserInfoInput): Promise<boolean> {
  await localized(client.api.users.saveInfo(input), t('api.profileSaveFailed'))
  return true
}

export async function saveUserProfileCover(profileCoverUrl: string): Promise<boolean> {
  await localized(client.api.users.saveCover(profileCoverUrl), t('api.coverSaveFailed'))
  return true
}

export async function savePresetAvatar(avatarUrl: string): Promise<string> {
  const result = await localized(client.api.users.savePresetAvatar(avatarUrl), t('api.avatarPresetFailed'))
  if (!result.avatarUrl) throw new Error(t('api.avatarPresetEmpty'))
  return result.avatarUrl
}

export async function wearBadge(badgeCode: string): Promise<boolean> {
  await localized(client.api.users.wearBadge(badgeCode), t('api.badgeWearFailed'))
  return true
}

export async function saveUserEmail(email: string): Promise<boolean> {
  await localized(client.api.users.saveEmail(email), t('api.emailSaveFailed'))
  return true
}

export async function resendActivationEmail(): Promise<string> {
  const result = await localized(client.api.users.resendActivationEmail(), t('api.activationEmailSendFailed'))
  return resolveApiMessage(result, t('settings.status.activationEmailSent'))
}

export async function saveUserName(username: string): Promise<boolean> {
  await localized(client.api.users.saveUsername(username), t('api.usernameSaveFailed'))
  return true
}

export async function changePassword(oldPassword: string, newPassword: string): Promise<boolean> {
  await localized(client.api.users.changePassword(oldPassword, newPassword), t('api.passwordChangeFailed'))
  return true
}

export function uploadAvatar(avatar: Blob | Blob[]): Promise<string> {
  return localized(client.api.uploads.avatar(avatar), t('api.avatarUploadFailed'))
}

export function getOAuthBindings(): Promise<OAuthBindingsPayload> {
  return localized(client.api.users.oauthBindings(), t('api.bindingsLoadFailed'))
}

export async function unbindOAuth(provider: string): Promise<boolean> {
  await localized(client.api.users.unbindOAuth(provider), t('api.unbindFailed'))
  return true
}

const loginInvalidRequestCode = 'auth.login.invalidRequest'
let publicKeyPromise: ReturnType<typeof client.api.auth.loginPublicKey> | undefined

export function getCaptcha(): Promise<CaptchaPayload> {
  return localized(client.api.auth.captcha(), t('api.captchaLoadFailed'))
}

export async function login(username: string, password: string, captchaId: string, captchaCode: string): Promise<boolean> {
  try {
    await submitLogin(username, password, captchaId, captchaCode, true)
  } catch (error) {
    if (!(error instanceof ApiResponseError) || error.messageCode !== loginInvalidRequestCode) throw error
    clearLoginPublicKey()
    await submitLogin(username, password, captchaId, captchaCode, true)
  }
  return true
}

async function submitLogin(username: string, password: string, captchaId: string, captchaCode: string, refreshKey = false) {
  const encryptedPassword = await encryptLoginPassword(password, refreshKey)
  await localized(client.api.auth.login({ username, encryptedPassword, captchaId, captchaCode }), t('api.loginFailed'))
}

export async function register(
  username: string,
  email: string,
  password: string,
  captchaId: string,
  captchaCode: string,
  locale?: string,
): Promise<string> {
  const result = await localized(client.api.auth.register({ username, email, password, captchaId, captchaCode, locale }), t('api.registerFailed'))
  return resolveApiMessage(result, t('auth.validation.registerSuccess'))
}

export async function forgotPassword(email: string, captchaId: string, captchaCode: string): Promise<string> {
  const result = await localized(client.api.auth.forgotPassword(email, captchaId, captchaCode), t('api.resetEmailFailed'))
  return resolveApiMessage(result, t('server.auth.passwordReset.mailQueued'))
}

export async function resetPassword(token: string, newPassword: string): Promise<string> {
  const result = await localized(client.api.auth.resetPassword(token, newPassword), t('api.passwordResetFailed'))
  return resolveApiMessage(result, t('server.auth.passwordReset.success'))
}

async function encryptLoginPassword(password: string, refreshKey = false): Promise<string> {
  const key = await getLoginPublicKey(refreshKey)
  const payload = JSON.stringify({ password, ts: key.serverTs })
  if (!window.crypto?.subtle) return encryptLoginPasswordWithForge(key.publicKey, payload)
  try {
    const importedKey = await window.crypto.subtle.importKey(
      'spki',
      pemToArrayBuffer(key.publicKey),
      { name: 'RSA-OAEP', hash: 'SHA-256' },
      false,
      ['encrypt'],
    )
    const encrypted = await window.crypto.subtle.encrypt(
      { name: 'RSA-OAEP' },
      importedKey,
      new TextEncoder().encode(payload),
    )
    return arrayBufferToBase64(encrypted)
  } catch {
    return encryptLoginPasswordWithForge(key.publicKey, payload)
  }
}

async function encryptLoginPasswordWithForge(publicKey: string, payload: string): Promise<string> {
  const { default: forge } = await import('node-forge')
  const key = forge.pki.publicKeyFromPem(publicKey)
  const encrypted = key.encrypt(forge.util.encodeUtf8(payload), 'RSA-OAEP', {
    md: forge.md.sha256.create(),
    mgf1: { md: forge.md.sha256.create() },
  })
  return forge.util.encode64(encrypted)
}

async function getLoginPublicKey(refresh = false) {
  if (refresh) clearLoginPublicKey()
  if (!publicKeyPromise) {
    publicKeyPromise = localized(client.api.auth.loginPublicKey(), t('api.loginKeyLoadFailed')).catch((error) => {
      publicKeyPromise = undefined
      throw error
    })
  }
  return publicKeyPromise
}

function clearLoginPublicKey() {
  publicKeyPromise = undefined
}

function pemToArrayBuffer(pem: string): ArrayBuffer {
  const base64 = pem
    .replace(/-----BEGIN PUBLIC KEY-----/g, '')
    .replace(/-----END PUBLIC KEY-----/g, '')
    .replace(/\s/g, '')
  const binary = window.atob(base64)
  return Uint8Array.from(binary, (char) => char.charCodeAt(0)).buffer
}

function arrayBufferToBase64(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return window.btoa(binary)
}
