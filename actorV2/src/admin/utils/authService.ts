import axiosInstance from './axiosInstance';
import {enqueueMessage} from "./messageManager.ts";
import axios from 'axios';
import type {Articles, Category, Label, Result, User} from "../types/adminInterfaces.ts";

export interface PageData<T> {
    list: T[];
    page: number
    size: number
    total: number
}

// Mock 获取用户信息
export const getUserInfo = async () => {
    return axiosInstance.get("/get-user-info")
}


export function getUserList():Promise<Result<PageData<User>>> {
    return axiosInstance.post("admin/user-list")
}

export function editUser(userId:any, status:any, validate:any, roleId:any) {
    return axiosInstance.post("admin/user-edit", {
        userId: userId,
        status: status,
        validate: validate,
        roleId: roleId,
    })
}

export function getAllRoleItem():Promise<Result<Label[]>>  {
    return axiosInstance.post('/admin/get-all-role-item')
}

export function getPermissionList() {
    return axiosInstance.post("admin/get-permission-list")
}

export function getRoleList() {
    return axiosInstance.post("admin/role-list")
}

export function getRoleSave(id:any, roleName:any, permission:any) {
    return axiosInstance.post("admin/role-save", {
        id: id,
        roleName: roleName,
        permissions: permission
    })
}

export function getRoleDel(id:any) {
    return axiosInstance.post("admin/role-delete", {
        id: id,
    })
}

export function getAdminArticlesList(page = 1, pageSize = 10):Promise<Result<PageData<Articles>>>  {
    return axiosInstance.post("admin/articles-list", {
        page: page,
        pageSize: pageSize,
    })
}

export const getCategoryList = ():Promise<Result<Category[]>> => {
    return axiosInstance.post('/admin/category-list')
}

export const saveCategory = (data:any) => {
    return axiosInstance.post('/admin/category-save', data)
}

export const deleteCategory = (id:any) => {
    return axiosInstance.post('/admin/category-delete', {id})
}


// 文章管理相关接口
export const editArticle = (id:any, processStatus:any) => {
    return axiosInstance.post('/admin/article-edit', {
        id,
        processStatus
    })
}
