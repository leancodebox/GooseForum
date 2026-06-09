import type { NotificationFilter, NotificationListResponse, ReplyWindowPayload, UserCardPayload, UserHoverCardPayload } from '@/types/payload'
import { i18n } from './i18n'
import { resolveApiMessage } from './api-message'

interface ApiResponse<T> {
  code?: number
  messageCode?: string
  params?: Record<string, unknown>
  result?: T
  data?: T
}

function responseMessage(data: ApiResponse<unknown>, fallback: string) {
  return resolveApiMessage(data, fallback)
}

function t(key: string) {
  return i18n.global.t(key)
}

async function readApiResponse<T>(response: Response, fallback: string): Promise<T> {
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }
  const data = (await response.json()) as ApiResponse<T>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(responseMessage(data, fallback))
  }
  return (data.result ?? data.data) as T
}

async function readApiSuccessMessage(response: Response, successFallback: string, errorFallback: string): Promise<string> {
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }
  const data = (await response.json()) as ApiResponse<unknown>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(responseMessage(data, errorFallback))
  }
  return responseMessage(data, successFallback)
}

export interface PostReplyResult {
  id: number
  replyNo?: number
  renderedContent: string
}

export interface UpdateReplyResult {
  id: number
  replyNo?: number
  content: string
  renderedContent: string
  updatedAt: string
}

export async function postReply(articleId: number, content: string, replyId = 0): Promise<PostReplyResult | number | boolean> {
  const response = await fetch('/api/forum/articles-reply', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      articleId,
      content,
      replyId,
    }),
  })
  return readApiResponse<PostReplyResult | number | boolean>(response, t('api.replyFailed'))
}

export async function updateReply(replyId: number, content: string): Promise<UpdateReplyResult> {
  const response = await fetch('/api/forum/articles-reply-update', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      replyId,
      content,
    }),
  })
  return readApiResponse<UpdateReplyResult>(response, t('api.replyUpdateFailed'))
}

export async function deleteReply(replyId: number): Promise<boolean> {
  const response = await fetch('/api/forum/articles-reply-delete', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      replyId,
    }),
  })
  return readApiResponse<boolean>(response, t('api.replyDeleteFailed'))
}

export interface ReplyWindowInput {
  articleId: number
  anchorReplyId?: number
  anchorReplyNo?: number
  before?: number
  after?: number
  beforeReplyNo?: number
  afterReplyNo?: number
  limit?: number
  tail?: boolean
}

export async function getArticleRepliesWindow(input: ReplyWindowInput): Promise<ReplyWindowPayload> {
  const params = new URLSearchParams({
    articleId: String(input.articleId),
  })
  if (input.anchorReplyId) params.set('anchorReplyId', String(input.anchorReplyId))
  if (input.anchorReplyNo) params.set('anchorReplyNo', String(input.anchorReplyNo))
  if (input.before) params.set('before', String(input.before))
  if (input.after) params.set('after', String(input.after))
  if (input.beforeReplyNo) params.set('beforeReplyNo', String(input.beforeReplyNo))
  if (input.afterReplyNo) params.set('afterReplyNo', String(input.afterReplyNo))
  if (input.limit) params.set('limit', String(input.limit))
  if (input.tail) params.set('tail', 'true')

  const response = await fetch(`/api/forum/article-replies-window?${params.toString()}`, {
    headers: {
      Accept: 'application/json',
    },
  })
  return readApiResponse<ReplyWindowPayload>(response, t('api.repliesLoadFailed'))
}

export async function likeArticle(id: number, action: 1 | 2): Promise<boolean> {
  const response = await fetch('/api/forum/like-articles', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      id,
      action,
    }),
  })
  return readApiResponse<boolean>(response, t('api.likeFailed'))
}

export async function bookmarkArticle(id: number, action: 1 | 2): Promise<boolean> {
  const response = await fetch('/api/forum/bookmark-article', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      id,
      action,
    }),
  })
  return readApiResponse<boolean>(response, t('api.bookmarkFailed'))
}

export async function watchArticle(id: number, action: 1 | 2): Promise<boolean> {
  const response = await fetch('/api/forum/watch-article', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      id,
      action,
    }),
  })
  return readApiResponse<boolean>(response, t('api.watchFailed'))
}

export async function updateArticleStatus(id: number, articleStatus: 0 | 1): Promise<boolean> {
  const response = await fetch('/api/forum/article-status', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      id,
      articleStatus,
    }),
  })
  return readApiResponse<boolean>(response, t('api.articleStatusFailed'))
}

export async function markAllNotificationsRead(): Promise<boolean> {
  const response = await fetch('/api/forum/notification/mark-all-read', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  })
  return readApiResponse<boolean>(response, t('api.markReadFailed'))
}

export async function markNotificationRead(notificationId: number): Promise<boolean> {
  const response = await fetch('/api/forum/notification/mark-read', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ notificationId }),
    keepalive: true,
  })
  return readApiResponse<boolean>(response, t('api.markReadFailed'))
}

export async function fetchNotifications(filter: NotificationFilter, cursor = 0, limit = 20): Promise<NotificationListResponse> {
  const params = new URLSearchParams({
    filter,
    cursor: String(cursor),
    limit: String(limit),
  })
  const response = await fetch(`/api/forum/notifications?${params.toString()}`, {
    headers: {
      Accept: 'application/json',
    },
  })
  return readApiResponse<NotificationListResponse>(response, t('api.notificationsLoadFailed'))
}

export async function getUserCard(userId: number): Promise<UserCardPayload> {
  const response = await fetch(`/api/user-card?userId=${encodeURIComponent(String(userId))}`, {
    headers: {
      Accept: 'application/json',
    },
  })
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = (await response.json()) as ApiResponse<UserCardPayload>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(responseMessage(data, t('api.userLoadFailed')))
  }

  const result = data.result ?? data.data
  if (!result) {
    throw new Error(t('api.userEmpty'))
  }
  return result
}

export async function getUserHoverCard(userId: number): Promise<UserHoverCardPayload> {
  const response = await fetch(`/api/user-hover-card?userId=${encodeURIComponent(String(userId))}`, {
    headers: {
      Accept: 'application/json',
    },
  })
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = (await response.json()) as ApiResponse<UserHoverCardPayload>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(responseMessage(data, t('api.userLoadFailed')))
  }

  const result = data.result ?? data.data
  if (!result) {
    throw new Error(t('api.userEmpty'))
  }
  return result
}

export async function followUser(userId: number, isFollowing: boolean): Promise<boolean> {
  const response = await fetch('/api/forum/follow-user', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      id: userId,
      action: isFollowing ? 2 : 1,
    }),
  })
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = (await response.json()) as ApiResponse<boolean>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(responseMessage(data, t('api.followFailed')))
  }
  return data.result ?? data.data ?? true
}

export interface SubmitArticleInput {
  id: number
  title: string
  content: string
  type: number
  categoryId: number[]
  articleStatus: 0 | 1
}

export async function submitArticle(article: SubmitArticleInput): Promise<number> {
  const response = await fetch('/api/forum/write-articles', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(article),
  })
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = (await response.json()) as ApiResponse<number>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(responseMessage(data, t('api.articleSaveFailed')))
  }
  return data.result ?? data.data ?? article.id
}

export async function uploadImage(file: File): Promise<string> {
  const formData = new FormData()
  formData.append('file', file)
  const response = await fetch('/file/img-upload', {
    method: 'POST',
    body: formData,
  })
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = (await response.json()) as ApiResponse<{ url: string }>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(responseMessage(data, t('api.imageUploadFailed')))
  }
  const result = data.result ?? data.data
  if (!result?.url) {
    throw new Error(t('api.imageUploadEmpty'))
  }
  return result.url
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

export async function getChatMessages(convId: number, page = 1, pageSize = 50): Promise<ChatMessagePayload[]> {
  const response = await fetch('/api/forum/chat/messages', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ convId, page, pageSize }),
  })
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = (await response.json()) as ApiResponse<{ list: ChatMessagePayload[] }>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(responseMessage(data, t('api.messagesLoadFailed')))
  }
  return data.result?.list ?? data.data?.list ?? []
}

export async function sendChatMessage(peerId: number, content: string): Promise<number> {
  const response = await fetch('/api/forum/chat/send', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ peerId, content, msgType: 1 }),
  })
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = (await response.json()) as ApiResponse<{ convId: number }>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(responseMessage(data, t('api.sendFailed')))
  }
  return data.result?.convId ?? data.data?.convId ?? 0
}

export async function markChatRead(convId: number): Promise<boolean> {
  const response = await fetch('/api/forum/chat/mark-read', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ convId }),
  })
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = (await response.json()) as ApiResponse<boolean>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(responseMessage(data, t('api.markReadFailed')))
  }
  return data.result ?? data.data ?? true
}

export interface SaveUserInfoInput {
  nickname: string
  bio: string
  signature: string
  website: string
  websiteName: string
  externalInformation: Record<string, { link?: string }>
}

export async function saveUserInfo(input: SaveUserInfoInput): Promise<boolean> {
  const response = await fetch('/api/set-user-info', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(input),
  })
  await readApiResponse<unknown>(response, t('api.profileSaveFailed'))
  return true
}

export async function saveUserProfileCover(profileCoverUrl: string): Promise<boolean> {
  const response = await fetch('/api/set-user-profile-cover', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ profileCoverUrl }),
  })
  await readApiResponse<unknown>(response, t('api.coverSaveFailed'))
  return true
}

export async function saveUserEmail(email: string): Promise<boolean> {
  const response = await fetch('/api/set-user-email', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email }),
  })
  await readApiResponse<unknown>(response, t('api.emailSaveFailed'))
  return true
}

export async function resendActivationEmail(): Promise<string> {
  const response = await fetch('/api/resend-activation-email', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  })
  return readApiSuccessMessage(response, t('settings.status.activationEmailSent'), t('api.activationEmailSendFailed'))
}

export async function saveUserName(username: string): Promise<boolean> {
  const response = await fetch('/api/set-user-name', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username }),
  })
  await readApiResponse<unknown>(response, t('api.usernameSaveFailed'))
  return true
}

export async function changePassword(oldPassword: string, newPassword: string): Promise<boolean> {
  const response = await fetch('/api/change-password', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ oldPassword, newPassword }),
  })
  await readApiResponse<unknown>(response, t('api.passwordChangeFailed'))
  return true
}

export async function uploadAvatar(avatar: Blob | Blob[]): Promise<string> {
  const formData = new FormData()
  const avatars = Array.isArray(avatar) ? avatar : [avatar]
  const fields = ['avatar', 'avatarMedium']
  const filenames = ['avatar.webp', 'avatar_medium.webp']
  avatars.slice(0, 2).forEach((item, index) => {
    formData.append(fields[index], item, item instanceof File ? item.name : filenames[index])
  })
  const response = await fetch('/api/upload-avatar', {
    method: 'POST',
    body: formData,
  })
  const result = await readApiResponse<string | { avatarUrl?: string; url?: string }>(response, t('api.avatarUploadFailed'))
  if (typeof result === 'string') return result
  const url = result.avatarUrl || result.url
  if (!url) throw new Error(t('api.avatarUploadEmpty'))
  return url
}

export interface OAuthBindingPayload {
  bound: boolean
  provider?: string
  createdAt?: string
  updatedAt?: string
}

export type OAuthBindingsPayload = Record<string, OAuthBindingPayload>

export async function getOAuthBindings(): Promise<OAuthBindingsPayload> {
  const response = await fetch('/api/oauth/bindings', {
    headers: {
      Accept: 'application/json',
    },
  })
  return readApiResponse<OAuthBindingsPayload>(response, t('api.bindingsLoadFailed'))
}

export async function unbindOAuth(provider: string): Promise<boolean> {
  const response = await fetch(`/api/auth/${encodeURIComponent(provider)}/unbind`, {
    method: 'POST',
  })
  await readApiResponse<unknown>(response, t('api.unbindFailed'))
  return true
}

interface CaptchaPayload {
  captchaId: string
  captchaImg: string
}

let publicKeyPromise: Promise<string> | undefined

export async function getCaptcha(): Promise<CaptchaPayload> {
  const response = await fetch('/api/get-captcha', {
    headers: {
      Accept: 'application/json',
    },
  })
  return readApiResponse<CaptchaPayload>(response, t('api.captchaLoadFailed'))
}

export async function login(username: string, password: string, captchaId: string, captchaCode: string): Promise<boolean> {
  const encryptedPassword = await encryptLoginPassword(password)
  const response = await fetch('/api/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      username,
      encryptedPassword,
      captchaId,
      captchaCode,
    }),
  })
  await readApiResponse<unknown>(response, t('api.loginFailed'))
  return true
}

export async function register(
  username: string,
  email: string,
  password: string,
  captchaId: string,
  captchaCode: string,
): Promise<string> {
  const response = await fetch('/api/register', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      userName: username,
      email,
      passWord: password,
      captchaId,
      captchaCode,
    }),
  })
  return readApiSuccessMessage(response, t('auth.validation.registerSuccess'), t('api.registerFailed'))
}

export async function forgotPassword(email: string, captchaId: string, captchaCode: string): Promise<string> {
  const response = await fetch('/api/forgot-password', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      email,
      captchaId,
      captchaCode,
    }),
  })
  return readApiSuccessMessage(response, t('server.auth.passwordReset.mailQueued'), t('api.resetEmailFailed'))
}

export async function resetPassword(token: string, newPassword: string): Promise<string> {
  const response = await fetch('/api/reset-password', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      token,
      newPassword,
    }),
  })
  return readApiSuccessMessage(response, t('server.auth.passwordReset.success'), t('api.passwordResetFailed'))
}

async function encryptLoginPassword(password: string): Promise<string> {
  const publicKey = await getLoginPublicKey()
  const payload = JSON.stringify({
    password,
    ts: Date.now(),
  })
  if (!window.crypto?.subtle) {
    return encryptLoginPasswordWithForge(publicKey, payload)
  }
  try {
    return await encryptLoginPasswordWithWebCrypto(publicKey, payload)
  } catch {
    return encryptLoginPasswordWithForge(publicKey, payload)
  }
}

async function encryptLoginPasswordWithWebCrypto(publicKey: string, payload: string): Promise<string> {
  const key = await window.crypto.subtle.importKey(
    'spki',
    pemToArrayBuffer(publicKey),
    {
      name: 'RSA-OAEP',
      hash: 'SHA-256',
    },
    false,
    ['encrypt'],
  )

  const encrypted = await window.crypto.subtle.encrypt(
    { name: 'RSA-OAEP' },
    key,
    new TextEncoder().encode(payload),
  )

  return arrayBufferToBase64(encrypted)
}

async function encryptLoginPasswordWithForge(publicKey: string, payload: string): Promise<string> {
  const { default: forge } = await import('node-forge')
  const key = forge.pki.publicKeyFromPem(publicKey)
  const encrypted = key.encrypt(payload, 'RSA-OAEP', {
    md: forge.md.sha256.create(),
    mgf1: {
      md: forge.md.sha256.create(),
    },
  })
  return forge.util.encode64(encrypted)
}

async function getLoginPublicKey(): Promise<string> {
  if (!publicKeyPromise) {
    publicKeyPromise = fetch('/api/login-public-key', {
      headers: {
        Accept: 'application/json',
      },
    })
      .then((response) => readApiResponse<{ publicKey: string }>(response, t('api.loginKeyLoadFailed')).then((data) => data.publicKey))
      .catch((error) => {
        publicKeyPromise = undefined
        throw error
      })
  }
  return publicKeyPromise
}

function pemToArrayBuffer(pem: string): ArrayBuffer {
  const base64 = pem
    .replace(/-----BEGIN PUBLIC KEY-----/g, '')
    .replace(/-----END PUBLIC KEY-----/g, '')
    .replace(/\s/g, '')
  const binary = window.atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes.buffer
}

function arrayBufferToBase64(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (let i = 0; i < bytes.byteLength; i += 1) {
    binary += String.fromCharCode(bytes[i])
  }
  return window.btoa(binary)
}
