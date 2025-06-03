// 简单的axios实例配置
const API_BASE_URL = '/api'

// 简单的fetch封装，模拟axios行为
const request = async (url, options = {}) => {
  const config = {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    ...options
  }

  // 添加token
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }

  try {
    const response = await fetch(`${API_BASE_URL}${url}`, config)
    const data = await response.json()
    
    if (data.code === 0) {
      return data
    } else {
      throw new Error(data.message || '请求失败')
    }
  } catch (error) {
    console.error('API请求失败:', error)
    throw error
  }
}

// 获取文章枚举
export const getArticleEnum = async () => {
  try {
    return await request('/bbs/get-articles-enum', {
      method: 'GET'
    })
  } catch (error) {
    throw new Error(`获取文章枚举失败: ${error.message}`)
  }
}

// 获取文章原始数据
export const getArticlesOrigin = async (id) => {
  try {
    return await request('/bbs/get-articles-origin', {
      method: 'POST',
      body: JSON.stringify({
        id: parseInt(id)
      })
    })
  } catch (error) {
    throw new Error(`获取文章原始数据失败: ${error.message}`)
  }
}

// 提交文章
export const submitArticle = async (article) => {
  try {
    return await request('/bbs/write-articles', {
      method: 'POST',
      body: JSON.stringify({
        id: article.id,
        content: article.articleContent,
        title: article.articleTitle,
        type: article.type,
        categoryId: article.categoryId
      })
    })
  } catch (error) {
    throw new Error(`提交文章失败: ${error.message}`)
  }
}

// 获取用户信息
export const getUserInfo = async () => {
  try {
    return await request('/get-user-info', {
      method: 'GET'
    })
  } catch (error) {
    throw new Error(`获取用户信息失败: ${error.message}`)
  }
}

// 获取通知列表
export const getNotificationList = async (page = 1, size = 10) => {
  try {
    return await request('/bbs/notification/list', {
      method: 'POST',
      body: JSON.stringify({ page, size })
    })
  } catch (error) {
    throw new Error(`获取通知列表失败: ${error.message}`)
  }
}

// 标记通知为已读
export const markAsRead = async (id) => {
  try {
    return await request('/bbs/notification/mark-read', {
      method: 'POST',
      body: JSON.stringify({ id })
    })
  } catch (error) {
    throw new Error(`标记通知失败: ${error.message}`)
  }
}

// 标记所有通知为已读
export const markAllAsRead = async () => {
  try {
    return await request('/bbs/notification/mark-all-read', {
      method: 'POST'
    })
  } catch (error) {
    throw new Error(`标记所有通知失败: ${error.message}`)
  }
}

// 获取未读通知数量
export const getUnreadCount = async () => {
  try {
    return await request('/bbs/notification/unread-count', {
      method: 'GET'
    })
  } catch (error) {
    throw new Error(`获取未读通知数量失败: ${error.message}`)
  }
}

// 上传头像
export const uploadAvatar = async (formData) => {
  try {
    const token = localStorage.getItem('access_token')
    const headers = {}
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }

    const response = await fetch(`${API_BASE_URL}/upload-avatar`, {
      method: 'POST',
      headers,
      body: formData
    })
    
    const data = await response.json()
    if (data.code === 0) {
      return data
    } else {
      throw new Error(data.message || '上传失败')
    }
  } catch (error) {
    throw new Error(`上传头像失败: ${error.message}`)
  }
}

// 保存用户信息
export const saveUserInfo = async (userInfo) => {
  try {
    return await request('/set-user-info', {
      method: 'POST',
      body: JSON.stringify(userInfo)
    })
  } catch (error) {
    throw new Error(`保存用户信息失败: ${error.message}`)
  }
}

// 获取用户文章
export const getUserArticles = async (page = 1, size = 10) => {
  try {
    return await request('/bbs/get-user-articles', {
      method: 'POST',
      body: JSON.stringify({ page, size })
    })
  } catch (error) {
    throw new Error(`获取用户文章失败: ${error.message}`)
  }
}

// 修改密码
export const changePassword = async (oldPassword, newPassword) => {
  try {
    return await request('/change-password', {
      method: 'POST',
      body: JSON.stringify({
        oldPassword,
        newPassword
      })
    })
  } catch (error) {
    throw new Error(`修改密码失败: ${error.message}`)
  }
}