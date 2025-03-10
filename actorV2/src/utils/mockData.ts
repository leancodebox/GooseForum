import type { NotificationItem } from '@/types/notificationInterfaces';

// 生成模拟通知数据
const generateMockNotifications = (count: number): NotificationItem[] => {
  return Array.from({ length: count }, (_, index) => ({
    id: index + 1,
    type: ['system', 'comment', 'like', 'follow'][Math.floor(Math.random() * 4)] as NotificationItem['type'],
    title: `测试通知 ${index + 1}`,
    content: `这是一条测试通知消息的内容。消息 ID: ${index + 1}`,
    isRead: Math.random() > 0.5,
    createTime: new Date(Date.now() - Math.random() * 10 * 24 * 60 * 60 * 1000).toLocaleString(),
    link: Math.random() > 0.5 ? '/post/' + (index + 1) : undefined
  }));
};

// 模拟通知数据存储
let mockNotifications = generateMockNotifications(50);

export const mockNotificationService = {
  // 获取通知列表
  getNotifications: async (page: number = 1, size: number = 10, unreadOnly: boolean = false) => {
    let filteredNotifications = mockNotifications;
    if (unreadOnly) {
      filteredNotifications = mockNotifications.filter(n => !n.isRead);
    }
    
    const start = (page - 1) * size;
    const end = start + size;
    const list = filteredNotifications.slice(start, end);
    
    return {
      code: 0,
      result: {
        total: filteredNotifications.length,
        list
      },
      message: 'success'
    };
  },

  // 标记单个通知为已读
  markAsRead: async (id: number) => {
    mockNotifications = mockNotifications.map(notification =>
      notification.id === id ? { ...notification, isRead: true } : notification
    );
    return { code: 0, message: 'success' };
  },

  // 标记所有通知为已读
  markAllAsRead: async () => {
    mockNotifications = mockNotifications.map(notification => ({
      ...notification,
      isRead: true
    }));
    return { code: 0, message: 'success' };
  }
};