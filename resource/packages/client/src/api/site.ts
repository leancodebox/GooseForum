import { GooseClientError } from '../http/error.js'
import type { GooseHttpClient } from '../http/client.js'
import type { GooseSiteApi, ImageUploadInitResult } from './types.js'
import { siteApiRoutes } from './routes.js'

type SiteApiRouteName = keyof typeof siteApiRoutes
const route = (name: SiteApiRouteName) => siteApiRoutes[name][1]

const post = <T>(http: GooseHttpClient, path: string, json?: unknown, init: RequestInit = {}) =>
  http.request<T>(path, { ...init, method: 'POST', json })

const postWithMeta = <T>(http: GooseHttpClient, path: string, json?: unknown) =>
  http.requestWithMeta<T>(path, { method: 'POST', json })

export function createSiteApi(http: GooseHttpClient): GooseSiteApi {
  return {
    accessGroups: {
      list: () => http.request(route('accessGroupsList')),
      apply: (groupId) => post(http, route('accessGroupsApply'), { groupId }),
      managed: () => http.request(route('accessGroupsManaged')),
      review: (groupId, memberId, approve) => post(http, route('accessGroupsReview'), { groupId, memberId, approve }),
    },
    posts: {
      create: (input) => post(http, route('postsCreate'), { replyToPostId: 0, ...input }),
      update: (input) => post(http, route('postsUpdate'), input),
      delete: (postId) => post(http, route('postsDelete'), { postId }),
      window: (input) => http.request(route('postsWindow'), { query: {
        topicId: input.topicId,
        anchorPostId: input.anchorPostId,
        anchorPostNo: input.anchorPostNo,
        beforePostNo: input.beforePostNo,
        afterPostNo: input.afterPostNo,
        limit: input.limit,
      } }),
    },
    topics: {
      like: (topicId, action) => post(http, route('topicsLike'), { topicId, action }),
      bookmark: (topicId, action) => post(http, route('topicsBookmark'), { topicId, action }),
      watch: (topicId, action) => post(http, route('topicsWatch'), { topicId, action }),
      setStatus: (topicId, topicStatus) => post(http, route('topicsStatus'), { topicId, topicStatus }),
      write: (input) => post(http, route('topicsWrite'), input),
    },
    moderation: {
      setTopicStatus: (topicId, action) => post(http, route('moderationTopicStatus'), { topicId, action }),
      setPostStatus: (postId, action) => post(http, route('moderationPostStatus'), { postId, action }),
      report: (targetType, targetId, reason, note) => post(http, route('moderationReport'), { targetType, targetId, reason, note }),
      reports: (cursor = 0, pageSize = 20, status = 'open') => post(http, route('moderationReports'), { cursor, pageSize, status }),
      setReportStatus: (id, action) => post(http, route('moderationReportStatus'), { id, action }),
      logs: (cursor = 0, pageSize = 20) => post(http, route('moderationLogs'), { cursor, pageSize }),
    },
    notifications: {
      list: (filter, cursor = 0, limit = 20) => http.request(route('notificationsList'), { query: { filter, cursor, limit } }),
      markRead: (notificationId) => post(http, route('notificationsMarkRead'), { notificationId }, { keepalive: true }),
      markAllRead: () => post(http, route('notificationsMarkAllRead')),
      unread: () => http.request(route('notificationsUnread')),
    },
    users: {
      card: (userId) => http.request(route('userCard'), { query: { userId } }),
      follow: (userId, following) => post(http, route('userFollow'), { id: userId, action: following ? 2 : 1 }),
      saveInfo: (input) => post(http, route('userSaveInfo'), input),
      saveCover: (profileCoverUrl) => post(http, route('userSaveCover'), { profileCoverUrl }),
      savePresetAvatar: (avatarUrl) => post(http, route('userSavePresetAvatar'), { avatarUrl }),
      wearBadge: (badgeCode) => post(http, route('userWearBadge'), { badgeCode }),
      saveEmail: (email) => post(http, route('userSaveEmail'), { email }),
      resendActivationEmail: () => postWithMeta(http, route('userResendActivationEmail')),
      saveUsername: (username) => post(http, route('userSaveUsername'), { username }),
      changePassword: (oldPassword, newPassword) => post(http, route('userChangePassword'), { oldPassword, newPassword }),
      oauthBindings: () => http.request(route('userOauthBindings')),
      unbindOAuth: (provider) => post(http, route('userUnbindOauth').replace(':provider', encodeURIComponent(provider))),
    },
    chat: {
      messages: (input) => post(http, route('chatMessages'), {
        convId: input.convId,
        beforeId: input.beforeId || 0,
        afterId: input.afterId || 0,
        limit: input.limit || 30,
      }),
      send: (peerId, content) => post(http, route('chatSend'), { peerId, content, msgType: 1 }),
      markRead: (convId) => post(http, route('chatMarkRead'), { convId }),
    },
    auth: {
      captcha: () => http.request(route('authCaptcha')),
      loginPublicKey: () => http.request(route('authLoginPublicKey')),
      login: (input) => post(http, route('authLogin'), input),
      register: (input) => postWithMeta(http, route('authRegister'), {
        userName: input.username,
        email: input.email,
        passWord: input.password,
        locale: input.locale,
        captchaId: input.captchaId,
        captchaCode: input.captchaCode,
      }),
      forgotPassword: (email, captchaId, captchaCode) => postWithMeta(http, route('authForgotPassword'), { email, captchaId, captchaCode }),
      resetPassword: (token, newPassword) => postWithMeta(http, route('authResetPassword'), { token, newPassword }),
      logout: () => post(http, route('authLogout')),
    },
    themes: {
      save: (settings) => post(http, route('themeSave'), { settings }),
      publish: () => post(http, route('themePublish')),
    },
    uploads: {
      image: (file) => uploadImage(http, file),
      avatar: (avatar) => uploadAvatar(http, avatar),
    },
  }
}

async function uploadImage(http: GooseHttpClient, file: File) {
  const init = await post<ImageUploadInitResult>(http, route('imageUploadInit'), {
    filename: file.name,
    contentType: file.type,
    size: file.size,
  })
  if (init.mode === 'proxy') return uploadImageThroughServer(http, file)
  if (init.mode !== 'direct' || !init.name || !init.upload?.url || init.upload.method !== 'POST') {
    if (init.name) await abortDirectImageUpload(http, init.name)
    throw new GooseClientError('GooseForum image upload initialization returned incomplete data')
  }

  const formData = new FormData()
  for (const [key, value] of Object.entries(init.upload.fields || {})) formData.append(key, value)
  formData.append('file', file, file.name)
  let response: Response
  try {
    response = await http.fetch(init.upload.url, { method: 'POST', body: formData })
  } catch (error) {
    try {
      return await completeDirectImageUpload(http, init.name)
    } catch {
      throw error
    }
  }
  if (!response.ok) {
    await abortDirectImageUpload(http, init.name)
    throw new GooseClientError(`Image object upload failed with HTTP ${response.status}`, { status: response.status })
  }
  try {
    return await completeDirectImageUpload(http, init.name)
  } catch (error) {
    const transient = error instanceof TypeError
      || (error instanceof GooseClientError && (error.status || 0) >= 500)
    if (!transient) throw error
    return completeDirectImageUpload(http, init.name)
  }
}

async function uploadImageThroughServer(http: GooseHttpClient, file: File) {
  const formData = new FormData()
  formData.append('file', file)
  const response = await http.fetch(route('imageUpload'), { method: 'POST', body: formData })
  if (!response.ok) throw new GooseClientError(`Image upload failed with HTTP ${response.status}`, { status: response.status })
  const envelope = await response.json() as { code?: number; message?: string; messageCode?: string; result?: { url?: string }; data?: { url?: string } }
  if (envelope.code !== undefined && envelope.code !== 0) {
    throw new GooseClientError(envelope.message || envelope.messageCode || 'Image upload failed', {
      status: response.status,
      code: envelope.code,
      messageCode: envelope.messageCode,
    })
  }
  const url = (envelope.result ?? envelope.data)?.url
  if (!url) throw new GooseClientError('GooseForum image upload returned no URL')
  return url
}

async function completeDirectImageUpload(http: GooseHttpClient, name: string) {
  const result = await post<{ url?: string }>(http, route('imageUploadComplete'), { name })
  if (!result.url) throw new GooseClientError('GooseForum image upload returned no URL')
  return result.url
}

async function abortDirectImageUpload(http: GooseHttpClient, name: string) {
  try {
    await post(http, route('imageUploadAbort'), { name })
  } catch {
    // Pending uploads also expire server-side; abort remains best effort.
  }
}

async function uploadAvatar(http: GooseHttpClient, avatar: Blob | Blob[]) {
  const formData = new FormData()
  const avatars = Array.isArray(avatar) ? avatar : [avatar]
  const fields = ['avatar', 'avatarMedium']
  const filenames = ['avatar.webp', 'avatar_medium.webp']
  avatars.slice(0, 2).forEach((item, index) => {
    const filename = typeof File !== 'undefined' && item instanceof File ? item.name : filenames[index]
    formData.append(fields[index]!, item, filename)
  })
  const response = await http.fetch(route('avatarUpload'), { method: 'POST', body: formData })
  if (!response.ok) throw new GooseClientError(`Avatar upload failed with HTTP ${response.status}`, { status: response.status })
  const envelope = await response.json() as {
    code?: number
    message?: string
    messageCode?: string
    result?: string | { avatarUrl?: string; url?: string }
    data?: string | { avatarUrl?: string; url?: string }
  }
  if (envelope.code !== undefined && envelope.code !== 0) {
    throw new GooseClientError(envelope.message || envelope.messageCode || 'Avatar upload failed', {
      status: response.status,
      code: envelope.code,
      messageCode: envelope.messageCode,
    })
  }
  const result = envelope.result ?? envelope.data
  const url = typeof result === 'string' ? result : result?.avatarUrl || result?.url
  if (!url) throw new GooseClientError('GooseForum avatar upload returned no URL')
  return url
}
