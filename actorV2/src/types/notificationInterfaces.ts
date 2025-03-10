export interface NotificationItem {
  id: number;
  type: 'system' | 'comment' | 'like' | 'follow';
  title: string;
  content: string;
  isRead: boolean;
  createTime: string;
  link?: string;
}

export interface NotificationResponse {
  code: number;
  result: {
    total: number;
    list: NotificationItem[];
  };
  message: string;
}