import axios from "axios"
import { enqueueMessage } from './messageManager'
import router from '@/route/router'
import { useUserStore } from '@/modules/user'

const instanceAxios = axios.create({
    baseURL: import.meta.env.VITE_DEV_API_HOST,
    timeout: 10 * 1000,
    headers: {}
})

// 创建请求拦截器
instanceAxios.interceptors.request.use(config => {
    return config;
});

const success = 0
const fail = 1

instanceAxios.interceptors.response.use(response => {
    const res = response.data
    if (res === undefined) {
        return response
    }
    switch (res.code) {
        case success:
            return res;
        case fail:
            enqueueMessage(res.msg ? res.msg : "响应异常", 'error');
            throw new Error(res.msg ? res.msg : "响应异常");
    }
    return response
}, error => {
    const res = error.response?.data
    if (res === undefined || res.code === undefined) {
        console.error(error)
        return Promise.reject(error)
    }

    if (res.code === fail) {
        enqueueMessage(res.msg ? res.msg : "响应异常", 'error');
        return Promise.reject(error)
    }

    console.error(error)
    return Promise.reject(error)
})

// 获取初始化状态
export function getSetupStatus() {
  return instanceAxios.get('/setup/status')
}

// 提交初始化设置
export function submitSetup(data) {
  return instanceAxios.post('/setup/init', data)
}
