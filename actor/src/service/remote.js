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
const remoteService = {}
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

instanceAxios.interceptors.response.use(response => {
    if (response.headers['new-token'] !== undefined) {
        userStore.updateToken(response.headers['new-token'])
    }
    const res = response.data
    if (res === undefined) {
        return response
    }
    if (res.code === fail) {
        message.error(res.msg ? res.msg : "响应异常")
    }
    return response
})

export function getUserInfo() {
    return instanceAxios.get("get-user-info-v4")
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

export function getArticlesPageApi(page=1,pageSize=10,search="") {
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

export function getSysInfo() {
    return instanceAxios.get("/sys-info")
}

export {remoteService}