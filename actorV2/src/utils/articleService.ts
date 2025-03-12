import axiosInstance from './axiosInstance';
import {enqueueMessage} from "@/utils/messageManager.ts";
import axios from 'axios';

// 获取文章枚举
export const getArticleEnum = async (): Promise<any> => {
    try {
        return await axiosInstance.get('bbs/get-articles-enum');
    } catch (error) {
        throw new Error(`获取文章枚举失败: ${error}`);
    }
}

// 获取文章原始数据
export const getArticlesOrigin = async (id: any): Promise<any> => {
    try {
        return await axiosInstance.post('/bbs/get-articles-origin', {
            id: parseInt(id)
        });
    } catch (error) {
        throw new Error(`获取文章原始数据失败: ${error}`);
    }
}

// 提交文章的函数
export const submitArticle = async <T>(article: any): Promise<T> => {
    try {
        return await axiosInstance.post('/bbs/write-articles', {
            id: article.id,
            content: article.articleContent,
            title: article.articleTitle,
            type: article.type,
            categoryId: article.categoryId,
        });
    } catch (error) {
        // 检查是否有响应数据
        if (axios.isAxiosError(error) && error.response) {
            // 从响应中提取错误信息
            const errorMessage = error.response.data?.msg || '提交文章失败';
            enqueueMessage(`提交文章失败: ${errorMessage}`);
        } else {
            enqueueMessage(`提交文章失败`);
        }
        throw new Error(`提交文章失败: ${error}`);
    }
};

// Mock 获取用户信息
export const getUserInfo = async () => {
    try {
        return axiosInstance.get("/get-user-info")
    } catch (error) {
    }
};


// 获取通知列表
export function getNotificationList(params: any) {
    return axiosInstance.post('/bbs/notification/list', params)
}

// 获取未读通知数量
export function getUnreadCount() {
    return axiosInstance.get('/bbs/notification/unread-count')
}

// 标记通知为已读
export function markAsRead(params: any) {
    return axiosInstance.post('/bbs/notification/mark-read', params)
}

// 标记所有通知为已读
export function markAllAsRead() {
    return axiosInstance.post('/bbs/notification/mark-all-read')
}
