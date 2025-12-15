import axiosInstance from './axiosInstance';

import axios from 'axios';
import type {
    ArticleData,
    ArticleListItem,
    Notifications,
    OAuthBindings,
    PageData,
    QueryList,
    Result,
    UserInfo
} from "./gooseForumInterfaces.ts";

// 获取文章枚举
export const getArticleEnum = async (): Promise<any> => {
    try {
        return await axiosInstance.get('forum/get-articles-enum');
    } catch (error) {
        throw new Error(`获取文章枚举失败: ${error}`);
    }
}

// 获取文章原始数据
export const getArticlesOrigin = async (id: any): Promise<Result<any>> => {
    try {
        return await axiosInstance.post('/forum/get-articles-origin', {
            id: parseInt(id)
        });
    } catch (error) {
        throw new Error(`获取文章原始数据失败: ${error}`);
    }
}

// 提交文章的函数
export const submitArticle = async (article: ArticleData): Promise<Result<any>> => {
    try {
        return await axiosInstance.post('/forum/write-articles', {
            id: article.id,
            content: article.content,
            title: article.title,
            type: article.type,
            categoryId: article.categoryId,
        });
    } catch (error) {
        // 检查是否有响应数据
        if (axios.isAxiosError(error) && error.response) {
            // 从响应中提取错误信息
            const errorMessage = error.response.data?.msg || '提交文章失败';
            alert(`提交文章失败: ${errorMessage}`);
        } else {
            alert(`提交文章失败`);
        }
        throw new Error(`提交文章失败: ${error}`);
    }
};

// Mock 获取用户信息
export const getUserInfo = async (): Promise<Result<UserInfo>> => {
    return axiosInstance.get("/get-user-info")
}


// 获取通知列表
export function getNotificationList(page: any, pageSize: any, unreadOnly: any): Promise<Result<PageData<Notifications>>> {
    return axiosInstance.post('/forum/notification/list', {
        page: page,
        pageSize: pageSize,
        unreadOnly: unreadOnly,
    })
}

// 获取通知列表
export function queryNotificationList(startId: number, pageSize: number, unreadOnly: any): Promise<Result<QueryList<Notifications>>> {
    return axiosInstance.post('/forum/notification/query', {
        startId: startId,
        pageSize: pageSize,
        unreadOnly: unreadOnly,
    })
}

// 获取未读通知数量
export function getUnreadCount(): Promise<Result<any>> {
    return axiosInstance.get('/forum/notification/unread-count')
}

// 标记通知为已读
export function markAsReadById(notificationId: any): Promise<Result<any>> {
    return axiosInstance.post('/forum/notification/mark-read', {
        notificationId: notificationId
    })
}

// 标记所有通知为已读
export function markAllAsRead(): Promise<Result<any>> {
    return axiosInstance.post('/forum/notification/mark-all-read')
}

// 删除通知
export function deleteNotification(notificationId: any): Promise<Result<any>> {
    return axiosInstance.post('/forum/notification/delete', {
        notificationId: notificationId
    })
}

export function uploadAvatar(formData: FormData): Promise<Result<any>> {
    // 直接使用传入的FormData，不再重新创建
    // 确保请求头设置正确
    return axiosInstance.post("/upload-avatar", formData, {
        headers: {
            'Content-Type': 'multipart/form-data'
        }
    });
}


export function saveUserInfo(
    nickname?: String,
    email?: String,
    bio?: String,
    signature?: String,
    website?: String,
    websiteName?: string,
    externalInformation?: any
): Promise<Result<any>> {
    return axiosInstance.post('set-user-info', {
        nickname: nickname,
        email: email,
        bio: bio,
        signature: signature,
        website: website,
        websiteName: websiteName,
        externalInformation: externalInformation
    })
}

export function saveUserEmail(
    email?: String,
): Promise<Result<any>> {
    return axiosInstance.post('set-user-email', {
        email: email,
    })
}


export function saveUserName(
    newUsername?: String,
): Promise<Result<any>> {
    return axiosInstance.post('set-user-name', {
        username: newUsername,
    })
}


export function getUserArticles(page: number,
                                pageSize: number): Promise<Result<PageData<ArticleListItem>>> {
    return axiosInstance.post('forum/get-user-articles', {
        page: page,
        pageSize: pageSize,
    })
}

// 获取用户收藏文章列表
export const getUserBookmarkedArticles = (page: number, pageSize: number): Promise<Result<PageData<ArticleListItem>>> => {
    return axiosInstance.post('forum/get-user-bookmarked-articles', {
        page,
        pageSize
    });
};

export function changePassword(oldPassword: string,
                               newPassword: string): Promise<Result<any>> {
    return axiosInstance.post('change-password', {
        oldPassword: oldPassword,
        newPassword: newPassword,
    })
}

// OAuth相关API函数

// 获取OAuth绑定状态
export function getOAuthBindings(): Promise<Result<OAuthBindings>> {
    return axiosInstance.get('oauth/bindings');
}

// 解绑OAuth
export function unbindOAuth(provider: string): Promise<Result<any>> {
    return axiosInstance.post(`auth/${provider}/unbind`);
}

// 获取验证码
export function getCaptcha(): Promise<Result<any>> {
    return axiosInstance.get('/get-captcha');
}

// 用户登录
export function login(username: string, password: string, captchaId: string, captchaCode: string): Promise<Result<any>> {
    return axiosInstance.post('/login', {
        username,
        password,
        captchaId,
        captchaCode
    });
}

// 用户注册
export function register(username: string, email: string, password: string, captchaId: string, captchaCode: string): Promise<Result<any>> {
    return axiosInstance.post('/register', {
        username,
        email,
        password,
        captchaId,
        captchaCode
    });
}

// 忘记密码
export function forgotPassword(email: string, captchaId: string, captchaCode: string): Promise<Result<string>> {
    return axiosInstance.post('/forgot-password', {
        email,
        captchaId,
        captchaCode
    });
}

// 重置密码
export function resetPassword(token: string, newPassword: string): Promise<Result<any>> {
    return axiosInstance.post('/reset-password', {
        token,
        newPassword
    });
}

// 图片上传
export function uploadImage(formData: FormData): Promise<Result<any>> {
    return axiosInstance.post('/file/img-upload', formData, {
        baseURL: '',
        headers: {
            'Content-Type': 'multipart/form-data'
        }
    });
}
