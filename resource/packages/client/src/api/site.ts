import { GooseClientError } from '../http/error.js'
import type { GooseHttpClient } from '../http/client.js'
import type { GooseSiteApi, OIDCConsentDetails, OIDCConsentResult } from './types.js'
import { uploadImage } from './image-upload.js'
import { siteApiRoutes } from './routes.js'
import { protocolRoutes } from './protocol-routes.js'

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
      writeReviewed: (input) => post(http, route('topicsWrite'), { ...input, returnReview: true }),
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
      oidcGrants: () => http.request(route('userOidcGrants')),
      revokeOIDCGrant: (clientId) => post(http, route('userRevokeOidcGrant'), { clientId }),
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
    oidc: {
      consentDetails: (interaction) => rawJSON<OIDCConsentDetails>(
        http,
        protocolRoutes.oidcConsentDetails[1],
        { query: { interaction } },
      ),
      consentDecision: (interaction, decision) => rawJSON<OIDCConsentResult>(
        http,
        protocolRoutes.oidcConsentDecision[1],
        { method: 'POST', json: { interaction, decision } },
      ),
    },
    themes: {
      save: (settings) => post(http, route('themeSave'), { settings }),
      publish: () => post(http, route('themePublish')),
    },
    uploads: {
      image: async (file) => (await uploadImage(http, file, {
        init: route('imageUploadInit'),
        complete: route('imageUploadComplete'),
        abort: route('imageUploadAbort'),
        proxy: route('imageUpload'),
      })).url,
      avatar: (avatar) => uploadAvatar(http, avatar),
    },
  }
}

async function rawJSON<T>(http: GooseHttpClient, path: string, init: import('../http/client.js').GooseRequestInit = {}) {
  const { json, query, headers: inputHeaders, ...requestInit } = init
  const url = query ? http.resolve(`${path}?${new URLSearchParams(Object.entries(query)
    .filter((entry): entry is [string, string | number | boolean] => entry[1] !== undefined && entry[1] !== null)
    .map(([key, value]) => [key, String(value)]))}`) : path
  const headers = new Headers(inputHeaders)
  headers.set('Accept', 'application/json')
  let body = requestInit.body
  if (json !== undefined) {
    headers.set('Content-Type', 'application/json')
    body = JSON.stringify(json)
  }
  const response = await http.fetch(url, { ...requestInit, headers, body })
  const payload = await response.json().catch((cause) => {
    throw new GooseClientError('GooseForum OIDC endpoint returned invalid JSON', { status: response.status, cause })
  }) as T & { error?: string, error_description?: string }
  if (!response.ok) {
    throw new GooseClientError(payload.error_description || payload.error || `GooseForum OIDC request failed with HTTP ${response.status}`, {
      status: response.status,
    })
  }
  return payload
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
