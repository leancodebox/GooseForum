import axios from "axios"
import {useUserStore} from "@/modules/user";
import {createDiscreteApi,} from "naive-ui";
import router from '@/route/router'

const {message} = createDiscreteApi(
    ["message"],
);


const instanceAxios = axios.create({
    baseURL: import.meta.env.VITE_DEV_API_HOST,
    timeout: 10 * 1000,
    headers: {}
})

const userStore = useUserStore()

instanceAxios.interceptors.request.use(config => {
    config.headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + userStore.token,
        ...config.headers
    }
    return config;
});

const success = 0
const fail = 1

/**
 * response
 * {
 *  "code": 1,
 *  "result":{},
 *  "msg":""
 * }
 */
instanceAxios.interceptors.response.use(response => {
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
                query: { redirect: currentPath }  // 保存重定向信息
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

