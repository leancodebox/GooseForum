import type { GooseHttpClient } from '../http/client.js'
import type { GooseAdminApi } from './types.js'

const post = <T>(http: GooseHttpClient, path: string, json: unknown = {}) =>
  http.request<T>(path, { method: 'POST', json })

export function createAdminApi(http: GooseHttpClient): GooseAdminApi {
  return {
    settings: {
      site: () => http.request('/api/admin/site-settings'),
      saveSite: (settings) => post(http, '/api/admin/save-site-settings', { settings }),
      chrome: () => http.request('/api/admin/site-chrome'),
      saveChrome: (settings) => post(http, '/api/admin/save-site-chrome', { settings }),
    },
    audit: {
      records: (input) => post(http, '/api/admin/opt-record-page', input),
    },
    assets: {
      badges: () => http.request('/api/admin/badges'),
      saveBadge: (badge) => post(http, '/api/admin/badge-save', badge),
      deleteBadge: (code) => post(http, '/api/admin/badge-delete', { code }),
      files: (input) => post(http, '/api/admin/file-resources', input),
    },
    pages: {
      links: () => http.request('/api/admin/friend-links'),
      saveLinks: (groups) => post(http, '/api/admin/save-friend-links', { linksInfo: groups }),
      sponsors: () => http.request('/api/admin/sponsors'),
      saveSponsors: (config) => post(http, '/api/admin/save-sponsors', { sponsorsInfo: config }),
      uploadImage: (file) => {
        const body = new FormData()
        body.append('file', file)
        return http.request('/api/admin/img-upload', { method: 'POST', body })
      },
    },
    topics: {
      list: (input) => post(http, '/api/admin/topics/list', input),
      reviewPosts: (input) => post(http, '/api/admin/posts/review-list', input),
      source: (topicId) => post(http, '/api/admin/topics/source', { topicId }),
      setProcessStatus: (topicId, processStatus) => post(http, '/api/admin/topics/edit', { topicId, processStatus }),
      delete: (topicId) => post(http, '/api/admin/topics/delete', { topicId }),
      setPin: (topicId, pinWeight) => post(http, '/api/admin/topics/pin-edit', { topicId, pinWeight }),
      setCategories: (topicId, categoryId) => post(http, '/api/admin/topics/categories-edit', { topicId, categoryId }),
      review: (kind, input) => post(http, `/api/admin/${kind === 'topic' ? 'topics' : 'posts'}/review`, input),
    },
    users: {
      list: (input) => post(http, '/api/admin/user-list', input),
      edit: (input) => post(http, '/api/admin/user-edit', input),
      roles: () => http.request('/api/admin/get-all-role-item'),
      badgeOptions: (userId) => post(http, '/api/admin/user-badge-options', { userId }),
      saveBadges: (userId, badgeCodes) => post(http, '/api/admin/save-user-badges', { userId, badgeCodes }),
    },
    roles: {
      list: () => post(http, '/api/admin/role-list'),
      permissions: () => post(http, '/api/admin/get-permission-list'),
      save: (input) => post(http, '/api/admin/role-save', input),
      delete: (id) => post(http, '/api/admin/role-delete', { id }),
    },
    accessGroups: {
      overview: () => post(http, '/api/admin/access-control/overview'),
      save: (input) => post(http, '/api/admin/access-group/save', input),
      delete: (id) => post(http, '/api/admin/access-group/delete', { id }),
      saveMember: (input) => post(http, '/api/admin/access-group/member-save', input),
      deleteMember: (groupId, memberId) => post(http, '/api/admin/access-group/member-delete', { groupId, memberId }),
      reviewApplication: (groupId, memberId, approve) => post(http, '/api/admin/access-group/application-review', { groupId, memberId, approve }),
    },
    categories: {
      list: () => post(http, '/api/admin/category-list'),
      save: (category) => post(http, '/api/admin/category-save', category),
      delete: (id) => post(http, '/api/admin/category-delete', { id }),
      access: () => post(http, '/api/admin/access-control/overview'),
      saveAccess: (categoryId, grants) => post(http, '/api/admin/category-access/save', { categoryId, grants }),
      addModerator: (categoryId, user) => post(http, '/api/admin/category-moderator-add', { categoryId, ...user }),
      deleteModerator: (id) => post(http, '/api/admin/category-moderator-delete', { id }),
    },
    moderators: {
      list: () => post(http, '/api/admin/global-moderator-list'),
      add: (user) => post(http, '/api/admin/global-moderator-add', user),
      delete: (id) => post(http, '/api/admin/global-moderator-delete', { id }),
    },
  }
}
