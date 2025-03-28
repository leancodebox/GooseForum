import axiosInstance from './axiosInstance';
import {enqueueMessage} from "@/utils/messageManager.ts";
import axios from 'axios';
import type {Result} from "@/types/articleInterfaces.ts";


// Mock 获取用户信息
export const getUserInfo = async () => {
    return axiosInstance.get("/get-user-info")
}

export function getUserList() {
    return axiosInstance.post("admin/user-list")
}

export function editUser(userId, status, validate, roleId) {
    return axiosInstance.post("admin/user-edit", {
        userId: userId,
        status: status,
        validate: validate,
        roleId: roleId,
    })
}

export function getAllRoleItem() {
    return axiosInstance.post('/admin/get-all-role-item')
}

export function getPermissionList() {
    return axiosInstance.post("admin/get-permission-list")
}

export function getRoleList() {
    return axiosInstance.post("admin/role-list")
}

export function getRoleSave(id, roleName, permission) {
    return axiosInstance.post("admin/role-save", {
        id: id,
        roleName: roleName,
        permissions: permission
    })
}

export function getRoleDel(id) {
    return axiosInstance.post("admin/role-delete", {
        id: id,
    })
}

export function getAdminArticlesList(page = 1, pageSize = 10) {
    return axiosInstance.post("admin/articles-list", {
        page: page,
        pageSize: pageSize,
    })
}

export const getCategoryList = () => {
    return axiosInstance.post('/admin/category-list')
}

export const saveCategory = (data) => {
    return axiosInstance.post('/admin/category-save', data)
}

export const deleteCategory = (id) => {
    return axiosInstance.post('/admin/category-delete', {id})
}


// 文章管理相关接口
export const editArticle = (id, processStatus) => {
    return axiosInstance.post('/admin/article-edit', {
        id,
        processStatus
    })
}
