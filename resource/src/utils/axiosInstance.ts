import axios, {type AxiosError, type AxiosResponse} from 'axios';


const success = 0
const fail = 1

// 创建一个 Axios 实例
const axiosInstance = axios.create({
    baseURL: '/api', // 替换为您的 API 基础 URL
    timeout: 10000, // 请求超时设置
    headers: {
        'Content-Type': 'application/json',
    },
});

// 添加请求拦截器
axiosInstance.interceptors.request.use(
    (config) => {
        // 在发送请求之前做些什么，例如添加 token
        const token = localStorage.getItem('access_token'); // 假设 token 存储在 localStorage
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        // 处理请求错误
        return Promise.reject(error);
    }
);


// 添加响应拦截器
axiosInstance.interceptors.response.use(
    (response: AxiosResponse) => {
        const res = response.data
        if (res === undefined) {
            return response
        }
        // 统一处理响应数据
        const {data} = response

        // 如果后端返回的是标准格式 { code, message, data }
        if (data && typeof data === 'object' && 'code' in data) {
            if (data.code === 0) {
                return data
            } else {
                throw new Error(data?.msg ?? '请求异常')
            }
        }
        return data
    },
    (error: AxiosError) => {
        // 处理 HTTP 错误
        if (error.response) {
            const {status, data} = error.response

            switch (status) {
                case 401:
                    // 未授权，清除 token 并跳转到登录页
                    localStorage.removeItem('admin_token')
                    if (window.location.pathname !== '/admin/login') {
                        window.location.href = '/admin/login'
                    }
                    break
                case 403:
                    console.error('权限不足')
                    break
                case 404:
                    console.error('请求的资源不存在')
                    break
                case 500:
                    console.error('服务器内部错误')
                    break
                default:
                    console.error(`请求错误: ${status}`)
            }

            // 如果后端返回了错误信息，使用后端的错误信息
            if (data && typeof data === 'object' && 'msg' in data) {
                error.message = (data as any).msg
            }
        } else if (error.request) {
            // 网络错误
            error.message = '网络连接失败，请检查网络设置'
        } else {
            // 其他错误
            error.message = error.message || '请求失败'
        }

        return Promise.reject(error)
    }
);

export default axiosInstance;
