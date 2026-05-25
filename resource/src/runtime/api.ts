import type { ReplyWindowPayload, UserCardPayload, UserHoverCardPayload } from '@/types/payload'

interface ApiResponse<T> {
  code?: number
  message?: string
  msg?: string
  result?: T
  data?: T
}

function responseMessage(data: ApiResponse<unknown>, fallback: string) {
  return String(data.message || data.msg || fallback)
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

export interface PostReplyResult {
  id: number
  renderedContent: string
}

export interface UpdateReplyResult {
  id: number
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
  return readApiResponse<PostReplyResult | number | boolean>(response, '回复失败')
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
  return readApiResponse<UpdateReplyResult>(response, '更新回复失败')
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
  return readApiResponse<boolean>(response, '删除回复失败')
}

export interface ReplyWindowInput {
  articleId: number
  anchorReplyId?: number
  before?: number
  after?: number
  limit?: number
}

export async function getArticleRepliesWindow(input: ReplyWindowInput): Promise<ReplyWindowPayload> {
  const params = new URLSearchParams({
    articleId: String(input.articleId),
  })
  if (input.anchorReplyId) params.set('anchorReplyId', String(input.anchorReplyId))
  if (input.before) params.set('before', String(input.before))
  if (input.after) params.set('after', String(input.after))
  if (input.limit) params.set('limit', String(input.limit))

  const response = await fetch(`/api/forum/article-replies-window?${params.toString()}`, {
    headers: {
      Accept: 'application/json',
    },
  })
  return readApiResponse<ReplyWindowPayload>(response, '回复加载失败')
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
  return readApiResponse<boolean>(response, '点赞失败')
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
  return readApiResponse<boolean>(response, '收藏失败')
}

export async function markAllNotificationsRead(): Promise<boolean> {
  const response = await fetch('/api/forum/notification/mark-all-read', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  })
  return readApiResponse<boolean>(response, '标记已读失败')
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
  return readApiResponse<boolean>(response, '标记已读失败')
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
    throw new Error(data.message || data.msg || '用户信息加载失败')
  }

  const result = data.result ?? data.data
  if (!result) {
    throw new Error('用户信息为空')
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
    throw new Error(data.message || data.msg || '用户信息加载失败')
  }

  const result = data.result ?? data.data
  if (!result) {
    throw new Error('用户信息为空')
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
    throw new Error(data.message || data.msg || '关注操作失败')
  }
  return data.result ?? data.data ?? true
}

export interface SubmitArticleInput {
  id: number
  title: string
  content: string
  type: number
  categoryId: number[]
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
    throw new Error(data.message || data.msg || '文章保存失败')
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
    throw new Error(data.message || data.msg || '图片上传失败')
  }
  const result = data.result ?? data.data
  if (!result?.url) {
    throw new Error('图片上传返回为空')
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
    throw new Error(data.message || data.msg || '消息加载失败')
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
    throw new Error(data.message || data.msg || '发送失败')
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
    throw new Error(data.message || data.msg || '标记已读失败')
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
  await readApiResponse<unknown>(response, '资料保存失败')
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
  await readApiResponse<unknown>(response, '邮箱保存失败')
  return true
}

export async function saveUserName(username: string): Promise<boolean> {
  const response = await fetch('/api/set-user-name', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username }),
  })
  await readApiResponse<unknown>(response, '用户名保存失败')
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
  await readApiResponse<unknown>(response, '密码修改失败')
  return true
}

export async function uploadAvatar(avatar: Blob): Promise<string> {
  const formData = new FormData()
  formData.append('avatar', avatar, 'avatar.webp')
  const response = await fetch('/api/upload-avatar', {
    method: 'POST',
    body: formData,
  })
  const result = await readApiResponse<string | { avatarUrl?: string; url?: string }>(response, '头像上传失败')
  if (typeof result === 'string') return result
  const url = result.avatarUrl || result.url
  if (!url) throw new Error('头像上传返回为空')
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
  return readApiResponse<OAuthBindingsPayload>(response, '绑定状态加载失败')
}

export async function unbindOAuth(provider: string): Promise<boolean> {
  const response = await fetch(`/api/auth/${encodeURIComponent(provider)}/unbind`, {
    method: 'POST',
  })
  await readApiResponse<unknown>(response, '解绑失败')
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
  return readApiResponse<CaptchaPayload>(response, '验证码加载失败')
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
  await readApiResponse<unknown>(response, '登录失败')
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
  return readApiResponse<string>(response, '注册失败')
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
  return readApiResponse<string>(response, '重置邮件发送失败')
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
  return readApiResponse<string>(response, '密码重置失败')
}

async function encryptLoginPassword(password: string): Promise<string> {
  if (!window.crypto?.subtle) {
    throw new Error('当前浏览器不支持安全登录加密')
  }

  const publicKey = await getLoginPublicKey()
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

  const payload = JSON.stringify({
    password,
    ts: Date.now(),
  })
  const encrypted = await window.crypto.subtle.encrypt(
    { name: 'RSA-OAEP' },
    key,
    new TextEncoder().encode(payload),
  )

  return arrayBufferToBase64(encrypted)
}

async function getLoginPublicKey(): Promise<string> {
  if (!publicKeyPromise) {
    publicKeyPromise = fetch('/api/login-public-key', {
      headers: {
        Accept: 'application/json',
      },
    }).then((response) => readApiResponse<{ publicKey: string }>(response, '登录密钥加载失败').then((data) => data.publicKey))
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
