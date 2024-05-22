import axios from "axios"
import {useUserStore} from "@/modules/user";
import {createDiscreteApi,} from "naive-ui";

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
    const res = error.response.data
    if (res === undefined || res.code === undefined) {
        console.error(error)
        return
    }
    if (res.code === fail) {
        message.error(res.msg ? res.msg : "响应异常")
        return
    }
    console.error(error)
})

export function getUserInfo() {
    return instanceAxios.get("get-user-info")
}

export function login(username, password) {
    return instanceAxios.post("/login", {
        username: username,
        password: password
    })
}

export function reg(email, username, password) {
    return instanceAxios.post("/reg", {
        email: email,
        username: username,
        password: password
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
        content: data.content,
        title: data.title,
        type: data.type,
        categoryId: data.categoryId,
    })
}

export function getUserList() {
    return instanceAxios.post("admin/user-list")
}

export function getRoleList() {
    return instanceAxios.post("admin/role-list")
}

