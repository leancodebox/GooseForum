// 通知类型枚举
export enum NotificationType {
  SUCCESS = 'success',
  ERROR = 'error',
  WARNING = 'warning',
  INFO = 'info'
}

// 通知配置接口
export interface NotificationConfig {
  message: string;
  type?: NotificationType;
  duration?: number; // 显示时长，0表示不自动关闭
  position?: 'top-right' | 'top-left' | 'top-center' | 'bottom-right' | 'bottom-left' | 'bottom-center';
  closable?: boolean; // 是否显示关闭按钮
  title?: string; // 可选标题
  icon?: boolean; // 是否显示图标
}

// 通知实例接口
export interface NotificationInstance {
  id: string;
  close: () => void;
}

// 内部通知实例接口
interface InternalNotificationInstance {
  id: string;
  element: HTMLElement;
  timer?: number;
}

// 全局通知管理器
class NotificationManager {
  private instances: InternalNotificationInstance[] = [];
  private container: HTMLElement | null = null;
  private idCounter = 0;

  constructor() {
    this.initContainer();
  }

  // 初始化容器
  private initContainer(): void {
    if (!this.container) {
      this.container = document.createElement('div');
      this.container.id = 'global-notification-container';
      // 使用DaisyUI的toast容器样式
      this.container.className = 'toast toast-top toast-end';
      document.body.appendChild(this.container);
    }
  }

  // 显示通知
  show(config: NotificationConfig): string {
    const id = `notification-${++this.idCounter}`;
    const {
      message,
      type = NotificationType.INFO,
      duration = 3000,
      position = 'top-right',
      closable = true
    } = config;

    // 创建通知元素
    const element = this.createElement(id, message, type, closable, config);
    
    // 设置位置
    this.setContainerPosition(position);
    
    // 添加到容器
    if (this.container) {
      this.container.appendChild(element);
    }

    // 添加到实例列表
    const instance: InternalNotificationInstance = { id, element };
    this.instances.push(instance);

    // 显示动画
    setTimeout(() => {
      element.classList.remove('translate-x-full', 'opacity-0');
      element.classList.add('translate-x-0', 'opacity-100');
    }, 10);

    // 自动关闭
    if (duration > 0) {
      instance.timer = window.setTimeout(() => {
        this.close(id);
      }, duration);
    }

    return id;
  }

  // 创建通知元素
  private createElement(id: string, message: string, type: NotificationType, closable: boolean, config?: NotificationConfig): HTMLElement {
    const element = document.createElement('div');
    element.id = id;
    element.setAttribute('role', 'alert');
    
    // 使用DaisyUI的alert组件样式
    const typeClass = this.getTypeClass(type);
    element.className = `alert ${typeClass} mb-3 min-w-80 max-w-96 transform translate-x-full opacity-0 transition-all duration-300 pointer-events-auto relative shadow-lg`;
    
    // 添加图标（如果启用）
    if (config?.icon !== false) {
      const icon = this.createIcon(type);
      element.appendChild(icon);
    }
    
    // 创建文本内容容器
    const textContainer = document.createElement('div');
    textContainer.className = 'flex-1';
    
    // 添加标题（如果有）
    if (config?.title) {
      const titleElement = document.createElement('h3');
      titleElement.textContent = config.title;
      titleElement.className = 'font-bold';
      textContainer.appendChild(titleElement);
    }
    
    // 添加消息文本
    const messageElement = document.createElement(config?.title ? 'div' : 'span');
    messageElement.textContent = message;
    if (config?.title) {
      messageElement.className = 'text-xs';
    }
    textContainer.appendChild(messageElement);
    
    element.appendChild(textContainer);
    
    // 添加关闭按钮
    if (closable) {
      const closeBtn = document.createElement('button');
      closeBtn.innerHTML = '×';
      closeBtn.className = 'btn btn-ghost btn-xs btn-circle absolute top-2 right-2';
      closeBtn.onclick = () => this.close(id);
      element.appendChild(closeBtn);
    }
    
    return element;
  }

  // 设置容器位置
  private setContainerPosition(position: string): void {
    if (!this.container) return;
    
    // 使用DaisyUI的toast位置类
    let toastClasses = 'toast';
    
    switch (position) {
      case 'top-right':
        toastClasses += ' toast-top toast-end top-20';
        break;
      case 'top-left':
        toastClasses += ' toast-top toast-start top-20';
        break;
      case 'top-center':
        toastClasses += ' toast-top toast-center top-20';
        break;
      case 'bottom-right':
        toastClasses += ' toast-end'; // toast-bottom is default
        break;
      case 'bottom-left':
        toastClasses += ' toast-start'; // toast-bottom is default
        break;
      case 'bottom-center':
        toastClasses += ' toast-center'; // toast-bottom is default
        break;
    }
    
    this.container.className = toastClasses;
  }

  // 获取类型对应的DaisyUI类名
  private getTypeClass(type: NotificationType): string {
    switch (type) {
      case NotificationType.SUCCESS:
        return 'alert-success';
      case NotificationType.ERROR:
        return 'alert-error';
      case NotificationType.WARNING:
        return 'alert-warning';
      case NotificationType.INFO:
      default:
        return 'alert-info';
    }
  }

  // 创建DaisyUI风格的SVG图标
  private createIcon(type: NotificationType): HTMLElement {
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('xmlns', 'http://www.w3.org/2000/svg');
    svg.setAttribute('fill', 'none');
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('class', 'h-6 w-6 shrink-0 stroke-current');
    
    const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    path.setAttribute('stroke-linecap', 'round');
    path.setAttribute('stroke-linejoin', 'round');
    path.setAttribute('stroke-width', '2');
    
    switch (type) {
      case NotificationType.SUCCESS:
        path.setAttribute('d', 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z');
        break;
      case NotificationType.ERROR:
        path.setAttribute('d', 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z');
        break;
      case NotificationType.WARNING:
        path.setAttribute('d', 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z');
        break;
      case NotificationType.INFO:
      default:
        path.setAttribute('d', 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z');
        break;
    }
    
    svg.appendChild(path);
    return svg as unknown as HTMLElement;
  }

  // 关闭通知
  close(id: string) {
    const index = this.instances.findIndex(instance => instance.id === id);
    if (index === -1) return;

    const instance = this.instances[index];
    
    // 清除定时器
    if (instance.timer) {
      clearTimeout(instance.timer);
    }

    // 关闭动画
    instance.element.classList.add('translate-x-full', 'opacity-0');

    // 移除元素
    setTimeout(() => {
      if (instance.element.parentNode) {
        instance.element.parentNode.removeChild(instance.element);
      }
      this.instances.splice(index, 1);
    }, 300);
  }

  // 关闭所有通知
  closeAll() {
    this.instances.forEach(instance => {
      this.close(instance.id);
    });
  }
}

// 创建全局实例
const notificationManager = new NotificationManager();

// 导出便捷函数
export const notification = {
  show: (config: NotificationConfig) => notificationManager.show(config),
  success: (message: string, duration?: number) => 
    notificationManager.show({ message, type: NotificationType.SUCCESS, duration }),
  error: (message: string, duration?: number) => 
    notificationManager.show({ message, type: NotificationType.ERROR, duration }),
  warning: (message: string, duration?: number) => 
    notificationManager.show({ message, type: NotificationType.WARNING, duration }),
  info: (message: string, duration?: number) => 
    notificationManager.show({ message, type: NotificationType.INFO, duration }),
  close: (id: string) => notificationManager.close(id),
  closeAll: () => notificationManager.closeAll()
};

// 将通知函数挂载到全局对象，供HTML直接调用
if (typeof window !== 'undefined') {
  (window as any).notification = notification;
  (window as any).NotificationType = NotificationType;
}

// 默认导出
export default notification;