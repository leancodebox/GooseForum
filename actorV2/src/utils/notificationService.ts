import axiosInstance from './axiosInstance';
import type { NotificationResponse } from '@/types/notificationInterfaces';
import { showErrorMessage } from './messageManager';

export const getNotifications = async (page: number = 1, size: number = 10): Promise<NotificationResponse> => {
  try {
    return await axiosInstance.get(`/notifications?page=${page}&size=${size}`);
  } catch (error) {
    showErrorMessage('获取通知失败');
    throw error;
  }
};

export const markAsRead = async (id: number): Promise<void> => {
  try {
    await axiosInstance.post(`/notifications/${id}/read`);
  } catch (error) {
    showErrorMessage('标记已读失败');
    throw error;
  }
};

export const markAllAsRead = async (): Promise<void> => {
  try {
    await axiosInstance.post('/notifications/read-all');
  } catch (error) {
    showErrorMessage('标记全部已读失败');
    throw error;
  }
};