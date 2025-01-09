import axios from "axios"
import {createDiscreteApi,} from "naive-ui";
import router from '@/route/router'
import { useUserStore } from '@/modules/user'

const {message} = createDiscreteApi(
    ["message"],
);

const instanceAxios = axios.create({
    baseURL: import.meta.env.VITE_DEV_API_HOST,
    timeout: 10 * 1000,
    headers: {}
})

// 创建请求拦截器
instanceAxios.interceptors.request.use(config => {
    const userStore = useUserStore()
    config.headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + userStore.token,
        ...config.headers
    }
    return config;
});

const success = 0
const fail = 1

instanceAxios.interceptors.response.use(response => {
    const userStore = useUserStore()
    if (response.headers['new-token'] !== undefined) {
        userStore.updateToken(response.headers['new-token'])
    }
    const res = response.data
    if (res === undefined) {
        return response
    }
    switch (res.code) {
        case success:
            return res;
        case fail:
            message.error(res.msg ? res.msg : "响应异常")
            throw new Error(res.msg ? res.msg : "响应异常")
    }
    return response
}, error => {
    const userStore = useUserStore()
    const res = error.response?.data

    // 处理未授权的情况（token失效或未登录）
    if (error.response?.status === 401) {
        message.error('登录已过期，请重新登录')
        userStore.clearUserInfo()

        // 保存当前路由，以便登录后返回
        const currentPath = router.currentRoute.value.fullPath
        if (currentPath !== '/home/regOrLogin') {
            router.push({
                path: '/home/regOrLogin',
                query: { redirect: currentPath }
            })
        }
        return Promise.reject(error)
    }

    if (res === undefined || res.code === undefined) {
        console.error(error)
        return Promise.reject(error)
    }

    if (res.code === fail) {
        message.error(res.msg ? res.msg : "响应异常")
        return Promise.reject(error)
    }

    console.error(error)
    return Promise.reject(error)
})

export function getUserInfo() {
    return instanceAxios.get("get-user-info")
}

export function login(username, password, captchaId, captchaCode) {
    return instanceAxios.post("/login", {
        username: username,
        password: password,
        captchaId: captchaId,
        captchaCode: captchaCode
    })
}

export function reg(email, username, password, captchaId, captchaCode) {
    return instanceAxios.post("/reg", {
        email: email,
        username: username,
        password: password,
        captchaId: captchaId,
        captchaCode: captchaCode
    })
}

export function gtSiteStatistics() {
    return instanceAxios.get('bbs/get-site-statistics')
}

export function getArticleCategory() {
    return instanceAxios.get('bbs/get-articles-category')
}

export function getArticlesPageApi(page = 1, pageSize = 20, search = "") {
    return instanceAxios.post('bbs/get-articles-page', {
        page: page,
        pageSize: pageSize,
        search: search
    })
}


export function getArticlesDetailApi(id, maxCommentId) {
    return instanceAxios.post('bbs/get-articles-detail', {
        maxCommentId: maxCommentId,
        id: parseInt(id),
        pageSize: 10,
    })
}

export function writeArticles(data) {
    return instanceAxios.post('bbs/write-articles', {
        id:data.id,
        content: data.content,
        title: data.title,
        type: data.type,
        categoryId: data.categoryId,
    })
}

export function articlesReply(articleId, content, replyId) {
    return instanceAxios.post("bbs/articles-reply", {
        articleId: articleId,
        content: content,
        replyId: replyId,
    })
}

export function getUserList() {
    return instanceAxios.post("admin/user-list")
}
export function editUser(userId,status,validate,roleId) {
    return instanceAxios.post("admin/user-edit",{
        userId:userId,
        status:status,
        validate:validate,
        roleId:roleId,
    })
}

export function getAllRoleItem() {
    return instanceAxios.post('/admin/get-all-role-item')
}

export function getPermissionList() {
    return instanceAxios.post("admin/get-permission-list")
}

export function getRoleList() {
    return instanceAxios.post("admin/role-list")
}

export function getRoleSave(id, roleName, permission) {
    return instanceAxios.post("admin/role-save", {
        id: id,
        roleName: roleName,
        permissions: permission
    })
}

export function getRoleDel(id) {
    return instanceAxios.post("admin/role-delete", {
        id: id,
    })
}

export function getAdminArticlesList(page = 1, pageSize = 10) {
    return instanceAxios.post("admin/articles-list", {
        page: page,
        pageSize: pageSize,
    })
}

export function getCaptcha() {
    return instanceAxios.get("/get-captcha")
}

export function getUserArticles(page = 1, pageSize = 10) {
    return instanceAxios.post('bbs/get-user-articles', {
        page,
        pageSize
    })
}

export const getUserInfoShow = (userId) => {
    return instanceAxios.post('/get-user-info-show',
         {
            userId
        })
}

export const getArticlesOrigin = (id) => {
    return instanceAxios.post('/bbs/get-articles-origin', {
        id: parseInt(id)
    })
}

export const getCategoryList = () => {
    return instanceAxios.post('/admin/category-list')
}

export const saveCategory = (data) => {
    return instanceAxios.post('/admin/category-save', data)
}

export const deleteCategory = (id) => {
    return instanceAxios.post('/admin/category-delete', { id })
}

export function getUserProfile() {
    return instanceAxios.get("/get-user-info")
}

export function updateUserProfile(data) {
    return instanceAxios.post("/set-user-info", data)
}

export function uploadAvatar(file) {
    const formData = new FormData();
    // 如果是 Blob 对象，需要创建 File 对象
    if (file instanceof Blob) {
        formData.append('avatar', new File([file], 'avatar.png', { type: file.type }));
    } else {
        formData.append('avatar', file);
    }

    return instanceAxios.post("/upload-avatar", formData, {
        headers: {
            'Content-Type': 'multipart/form-data'
        }
    });
}

// 获取通知列表
export function getNotificationList(params) {
  return instanceAxios.post('/bbs/notification/list', params)
}

// 获取未读通知数量
export function getUnreadCount() {
  return instanceAxios.get('/bbs/notification/unread-count')
}

// 标记通知为已读
export function markAsRead(params) {
  return instanceAxios.post('/bbs/notification/mark-read', params)
}

// 标记所有通知为已读
export function markAllAsRead() {
  return instanceAxios.post('/bbs/notification/mark-all-read')
}

// 删除通知
export function deleteNotification(params) {
  return instanceAxios.post('/bbs/notification/delete', params)
}

// 获取通知类型
export function getNotificationTypes() {
  return instanceAxios.get('/bbs/notification/types')
}

// 文章管理相关接口
export const editArticle = (id, processStatus) => {
  return instanceAxios.post('/admin/article-edit', {
    id,
    processStatus
  })
}
export {
    instanceAxios,
    // ... 其他导出保持不变 ...
}

