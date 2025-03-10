import axiosInstance from './axiosInstance';
import type { NotificationResponse } from '@/types/notificationInterfaces';
import { showErrorMessage } from './messageManager';
import { mockNotificationService } from './mockData';

export const getNotifications = async (page: number = 1, size: number = 10): Promise<NotificationResponse> => {
  try {
    // 使用 mock 服务
    return await mockNotificationService.getNotifications(page, size);
  } catch (error) {
    showErrorMessage('获取通知失败');
    throw error;
  }
};

export const markAsRead = async (id: number): Promise<void> => {
  try {
    // 使用 mock 服务
    await mockNotificationService.markAsRead(id);
  } catch (error) {
    showErrorMessage('标记已读失败');
    throw error;
  }
};

export const markAllAsRead = async (): Promise<void> => {
  try {
    // 使用 mock 服务
    await mockNotificationService.markAllAsRead();
  } catch (error) {
    showErrorMessage('标记全部已读失败');
    throw error;
  }
};