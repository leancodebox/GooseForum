import type { NotificationItem } from '@/types/notificationInterfaces';

const mockUsers = ['张三', '李四', '王五', '赵六', '系统'];
const mockActions = {
  system: ['系统维护通知', '账号安全提醒', '新功能上线通知'],
  comment: ['评论了你的文章', '回复了你的评论'],
  like: ['赞了你的文章', '赞了你的评论'],
  follow: ['关注了你']
};

const generateMockNotifications = (count: number): NotificationItem[] => {
  return Array.from({ length: count }, (_, index) => {
    const type = ['system', 'comment', 'like', 'follow'][Math.floor(Math.random() * 4)] as NotificationItem['type'];
    const user = type === 'system' ? '系统' : mockUsers[Math.floor(Math.random() * (mockUsers.length - 1))];
    const action = mockActions[type][Math.floor(Math.random() * mockActions[type].length)];
    const target = type === 'system' ? '' : '《Vue3 开发实战》';
    
    return {
      id: index + 1,
      type,
      title: type === 'system' ? action : `${user} ${action}`,
      content: type === 'system' ? '为了提供更好的服务体验，系统将于今晚进行例行维护。' : 
        `${target}${type === 'comment' ? '："这篇文章写得很好！"' : ''}`,
      isRead: Math.random() > 0.5,
      createTime: new Date(Date.now() - Math.random() * 10 * 24 * 60 * 60 * 1000).toLocaleString(),
      link: Math.random() > 0.5 ? '/post/' + (index + 1) : undefined
    };
  });
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