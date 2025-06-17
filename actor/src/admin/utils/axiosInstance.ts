import axios from 'axios';
import {enqueueMessage} from './messageManager.ts'
import {useUserStore} from "../stores/auth.ts";
import useRouter from "../router/index";


// 创建一个 Axios 实例
const axiosInstance = axios.create({
    baseURL: '/api', // 替换为您的 API 基础 URL
    timeout: 10000, // 请求超时设置
    headers: {
        'Content-Type': 'application/json',
    },
});

const success = 0
const fail = 1

// 添加响应拦截器
axiosInstance.interceptors.response.use(
    (response) => {
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
        return res
    },
    (error) => {
        console.log(error)
        // 处理响应错误
        if (error.response && error.response.status === 401) {
            window.location.href = '/login'
            // useRouter.push({
            //     path: '/login',
            //     query: {redirect: useRouter.currentRoute.value.fullPath} // 记录跳转来源页
            // })
        }
        // 处理响应错误
        return Promise.reject(error);
    }
);

export default axiosInstance;
