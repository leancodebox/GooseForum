import type { GooseHttpClient } from '../http/client.js'
import type { GooseAdminApi } from './types.js'

const post = <T>(http: GooseHttpClient, path: string, json: unknown = {}) =>
  http.request<T>(path, { method: 'POST', json })

export function createAdminApi(http: GooseHttpClient): GooseAdminApi {
  return {
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
    users: {
      list: (input) => post(http, '/api/admin/user-list', input),
    },
  }
}
