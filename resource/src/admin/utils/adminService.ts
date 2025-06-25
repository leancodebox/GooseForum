import axiosInstance from './axiosInstance.ts';
import type {
    ApplySheet,
    Articles,
    Category,
    FriendLinksGroup,
    Label,
    Result,
    User,
    UserRole
} from './adminInterfaces.ts';

export interface PageData<T> {
    list: T[];
    page: number
    size: number
    total: number
}

// Mock 获取用户信息
export const getUserInfo = async (): Promise<Result<any>> => {
    return axiosInstance.get('api/get-user-info')
}


export function getUserList(): Promise<Result<PageData<User>>> {
    return axiosInstance.post('api/admin/user-list')
}

export function editUser(userId: any, status: any, validate: any, roleId: any): Promise<Result<any>> {
    return axiosInstance.post('api/admin/user-edit', {
        userId: userId,
        status: status,
        validate: validate,
        roleId: roleId,
    })
}

export function getAllRoleItem(): Promise<Result<Label[]>> {
    return axiosInstance.post('api/admin/get-all-role-item')
}

export function getPermissionList(): Promise<Result<Label[]>> {
    return axiosInstance.post('api/admin/get-permission-list')
}

export function getRoleList(): Promise<Result<PageData<UserRole>>> {
    return axiosInstance.post('api/admin/role-list')
}

export function getRoleSave(id: any, roleName: any, permission: any): Promise<Result<any>> {
    return axiosInstance.post('api/admin/role-save', {
        id: id,
        roleName: roleName,
        permissions: permission
    })
}

export function getRoleDel(id: any): Promise<Result<any>> {
    return axiosInstance.post('api/admin/role-delete', {
        id: id,
    })
}

export function getAdminArticlesList(page = 1, pageSize = 10): Promise<Result<PageData<Articles>>> {
    return axiosInstance.post('api/admin/articles-list', {
        page: page,
        pageSize: pageSize,
    })
}

export const getCategoryList = (): Promise<Result<Category[]>> => {
    return axiosInstance.post('api/admin/category-list')
}

export const saveCategory = (data: any): Promise<Result<any>> => {
    return axiosInstance.post('api/admin/category-save', data)
}

export const deleteCategory = (id: any): Promise<Result<any>> => {
    return axiosInstance.post('api/admin/category-delete', {id})
}

export const editArticle = (id: any, processStatus: any): Promise<Result<any>> => {
    return axiosInstance.post('api/admin/article-edit', {
        id,
        processStatus
    })
}

export function applySheetList(page = 1, pageSize = 10): Promise<Result<PageData<ApplySheet>>> {
    return axiosInstance.post('api/admin/apply-sheet-list', {
        page: page,
        pageSize: pageSize,
    })
}


export function getFriendLinks(): Promise<Result<FriendLinksGroup[]>> {
    return axiosInstance.get('api/admin/friend-links')
}

export function saveFriendLinks(params:any): Promise<Result<FriendLinksGroup[]>> {
    return axiosInstance.post('api/admin/save-friend-links',{
        linksInfo:params
    })
}

export function getSiteStatistics(): Promise<Result<any>> {
    return axiosInstance.get('/api/forum/get-site-statistics')
}
