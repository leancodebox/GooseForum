import axios, { AxiosRequestConfig } from 'axios'
import { toast } from 'sonner'

export interface ApiResponse<T = any> {
  code: number
  msg: string
  result: T
}

const axiosInstance = axios.create({
  baseURL: '',
  timeout: 10000,
  headers: {
    'X-Goose-Request': 'true',
  },
})

axiosInstance.interceptors.request.use(
  (config) => config,
  (error) => Promise.reject(error)
)

axiosInstance.interceptors.response.use(
  (response) => response,
  (error) => {
    let message = '请求失败'
    if (error.response) {
      switch (error.response.status) {
        case 401:
          message = '未授权，请重新登录'
          break
        case 403:
          message = '拒绝访问'
          break
        case 404:
          message = '请求资源不存在'
          break
        case 500:
          message = '服务器错误'
          break
        default:
          message = `请求错误 ${error.response.status}`
      }
    } else if (error.message.includes('timeout')) {
      message = '请求超时'
    } else if (error.message.includes('Network Error')) {
      message = '网络错误'
    }

    toast.error(message)
    return Promise.reject(error)
  }
)

export const request = {
  get: <T = any>(url: string, config?: AxiosRequestConfig) =>
    axiosInstance.get<ApiResponse<T>>(url, config).then(res => res.data),
  
  post: <T = any>(url: string, data?: any, config?: AxiosRequestConfig) =>
    axiosInstance.post<ApiResponse<T>>(url, data, config).then(res => res.data),
  
  put: <T = any>(url: string, data?: any, config?: AxiosRequestConfig) =>
    axiosInstance.put<ApiResponse<T>>(url, data, config).then(res => res.data),
  
  delete: <T = any>(url: string, config?: AxiosRequestConfig) =>
    axiosInstance.delete<ApiResponse<T>>(url, config).then(res => res.data),
}

export default request
