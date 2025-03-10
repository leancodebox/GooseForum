import axios from 'axios';

// 创建一个 Axios 实例
const axiosInstance = axios.create({
  baseURL: 'https://your-api-url.com/api', // 替换为您的 API 基础 URL
  timeout: 10000, // 请求超时设置
  headers: {
    'Content-Type': 'application/json',
  },
});

// 添加请求拦截器
axiosInstance.interceptors.request.use(
  (config) => {
    // 在发送请求之前做些什么，例如添加 token
    const token = localStorage.getItem('token'); // 假设 token 存储在 localStorage
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
  (response) => {
    // 处理响应数据
    return response.data;
  },
  (error) => {
    // 处理响应错误
    return Promise.reject(error);
  }
);

export default axiosInstance; 