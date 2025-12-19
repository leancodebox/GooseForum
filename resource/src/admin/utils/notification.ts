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
    
    // 设置样式
    let alertClass = 'alert shadow-lg transform transition-all duration-300 translate-x-full opacity-0 mb-2 min-w-[300px]';
    
    switch (type) {
      case NotificationType.SUCCESS:
        alertClass += ' alert-success';
        break;
      case NotificationType.ERROR:
        alertClass += ' alert-error';
        break;
      case NotificationType.WARNING:
        alertClass += ' alert-warning';
        break;
      case NotificationType.INFO:
        alertClass += ' alert-info';
        break;
    }

    element.className = alertClass;

    // 图标
    let iconHtml = '';
    if (config?.icon !== false) {
      switch (type) {
        case NotificationType.SUCCESS:
          iconHtml = '<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>';
          break;
        case NotificationType.ERROR:
          iconHtml = '<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>';
          break;
        case NotificationType.WARNING:
          iconHtml = '<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>';
          break;
        case NotificationType.INFO:
          iconHtml = '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>';
          break;
      }
    }

    // 标题
    const titleHtml = config?.title ? `<h3 class="font-bold">${config.title}</h3>` : '';
    
    // 关闭按钮
    const closeBtnHtml = closable ? 
      `<button class="btn btn-sm btn-ghost btn-circle absolute right-2 top-2" onclick="document.getElementById('${id}').remove()">✕</button>` : '';

    element.innerHTML = `
      ${iconHtml}
      <div>
        ${titleHtml}
        <div class="text-sm">${message}</div>
      </div>
      ${closeBtnHtml}
    `;

    // 绑定关闭按钮事件
    if (closable) {
      const closeBtn = element.querySelector('button');
      if (closeBtn) {
        closeBtn.onclick = (e) => {
          e.stopPropagation();
          this.close(id);
        };
      }
    }

    return element;
  }

  // 设置容器位置
  private setContainerPosition(position: string): void {
    if (!this.container) return;
    
    // 移除旧的位置类
    this.container.classList.remove(
      'toast-top', 'toast-bottom', 'toast-start', 'toast-center', 'toast-end', 'toast-middle'
    );
    
    // 添加新的位置类
    switch (position) {
      case 'top-right':
        this.container.classList.add('toast', 'toast-top', 'toast-end');
        break;
      case 'top-left':
        this.container.classList.add('toast', 'toast-top', 'toast-start');
        break;
      case 'top-center':
        this.container.classList.add('toast', 'toast-top', 'toast-center');
        break;
      case 'bottom-right':
        this.container.classList.add('toast', 'toast-bottom', 'toast-end');
        break;
      case 'bottom-left':
        this.container.classList.add('toast', 'toast-bottom', 'toast-start');
        break;
      case 'bottom-center':
        this.container.classList.add('toast', 'toast-bottom', 'toast-center');
        break;
      default:
        this.container.classList.add('toast', 'toast-top', 'toast-end');
    }
  }

  // 关闭通知
  close(id: string): void {
    const index = this.instances.findIndex(i => i.id === id);
    if (index !== -1) {
      const { element, timer } = this.instances[index];
      
      // 清除定时器
      if (timer) {
        clearTimeout(timer);
      }
      
      // 移除动画
      element.classList.remove('translate-x-0', 'opacity-100');
      element.classList.add('translate-x-full', 'opacity-0');
      
      // 延迟移除元素
      setTimeout(() => {
        if (element.parentNode) {
          element.parentNode.removeChild(element);
        }
        this.instances.splice(index, 1);
      }, 300);
    }
  }

  // 快捷方法
  success(message: string, title?: string, duration?: number): string {
    return this.show({ message, title, type: NotificationType.SUCCESS, duration });
  }

  error(message: string, title?: string, duration?: number): string {
    return this.show({ message, title, type: NotificationType.ERROR, duration });
  }

  warning(message: string, title?: string, duration?: number): string {
    return this.show({ message, title, type: NotificationType.WARNING, duration });
  }

  info(message: string, title?: string, duration?: number): string {
    return this.show({ message, title, type: NotificationType.INFO, duration });
  }
}

// 导出单例
export const notification = new NotificationManager();
