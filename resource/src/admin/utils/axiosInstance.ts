import axios from 'axios'
import type { AxiosResponse, AxiosError } from 'axios'

// 创建 axios 实例
export const axiosInstance = axios.create({
  baseURL: '/', // 使用相对路径，由代理处理
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
axiosInstance.interceptors.request.use(
  (config) => {
    // 添加时间戳防止缓存
    if (config.method === 'get') {
      config.params = {
        ...config.params,
        _t: Date.now()
      }
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
axiosInstance.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data
    if (res === undefined) {
      return response
    }
    // 统一处理响应数据
    const { data } = response

    // 如果后端返回的是标准格式 { code, message, data }
    if (data && typeof data === 'object' && 'code' in data) {
      if (data.code === 0) {
        return data
      } else {
        // 业务错误
        const error = new Error(data.message || '请求失败') as any
        error.code = data.code
        error.response = response
        return Promise.reject(error)
      }
    }
    return data
  },
  (error: AxiosError) => {
    // 处理 HTTP 错误
    if (error.response) {
      const { status, data } = error.response

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
      if (data && typeof data === 'object' && 'message' in data) {
        error.message = (data as any).message
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
)

// 导出常用的请求方法
export const api = {
  get: <T = any>(url: string, params?: any) =>
    axiosInstance.get<T>(url, { params }),

  post: <T = any>(url: string, data?: any) =>
    axiosInstance.post<T>(url, data),

  put: <T = any>(url: string, data?: any) =>
    axiosInstance.put<T>(url, data),

  delete: <T = any>(url: string, params?: any) =>
    axiosInstance.delete<T>(url, { params }),

  patch: <T = any>(url: string, data?: any) =>
    axiosInstance.patch<T>(url, data)
}

export default axiosInstance