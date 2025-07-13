/**
 * Toast 提示工具函数
 */

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface ToastOptions {
  duration?: number; // 显示时长（毫秒）
  position?: 'top' | 'bottom' | 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right';
  closable?: boolean; // 是否可关闭
}

/**
 * 显示 Toast 提示
 * @param message 提示消息
 * @param type 提示类型
 * @param options 选项
 */
export const showToast = (
  message: string, 
  type: ToastType = 'info', 
  options: ToastOptions = {}
): void => {
  const {
    duration = 3000,
    position = 'top-right',
    closable = true
  } = options;

  // 创建 toast 容器（如果不存在）
  let container = document.getElementById('toast-container');
  if (!container) {
    container = document.createElement('div');
    container.id = 'toast-container';
    container.className = getContainerClass(position);
    document.body.appendChild(container);
  }

  // 创建 toast 元素
  const toast = document.createElement('div');
  const toastId = `toast-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  toast.id = toastId;
  toast.className = getToastClass(type);
  
  // 设置 toast 内容
  toast.innerHTML = `
    <div class="flex items-center gap-2">
      ${getIcon(type)}
      <span class="flex-1">${escapeHtml(message)}</span>
      ${closable ? '<button class="btn btn-ghost btn-xs" onclick="removeToast(\''+toastId+'\')">&times;</button>' : ''}
    </div>
  `;

  // 添加到容器
  container.appendChild(toast);

  // 添加进入动画
  setTimeout(() => {
    toast.classList.add('toast-enter');
  }, 10);

  // 自动移除
  if (duration > 0) {
    setTimeout(() => {
      removeToast(toastId);
    }, duration);
  }

  // 添加全局移除函数
  if (!(window as any).removeToast) {
    (window as any).removeToast = removeToast;
  }
};

/**
 * 移除指定的 Toast
 * @param toastId Toast ID
 */
const removeToast = (toastId: string): void => {
  const toast = document.getElementById(toastId);
  if (toast) {
    toast.classList.add('toast-exit');
    setTimeout(() => {
      toast.remove();
      
      // 如果容器为空，移除容器
      const container = document.getElementById('toast-container');
      if (container && container.children.length === 0) {
        container.remove();
      }
    }, 300);
  }
};

/**
 * 清除所有 Toast
 */
export const clearAllToasts = (): void => {
  const container = document.getElementById('toast-container');
  if (container) {
    container.remove();
  }
};

/**
 * 获取容器样式类
 */
const getContainerClass = (position: string): string => {
  const baseClass = 'fixed z-50 flex flex-col gap-2 p-4 pointer-events-none';
  
  switch (position) {
    case 'top':
      return `${baseClass} top-4 left-1/2 transform -translate-x-1/2`;
    case 'bottom':
      return `${baseClass} bottom-4 left-1/2 transform -translate-x-1/2`;
    case 'top-left':
      return `${baseClass} top-4 left-4`;
    case 'top-right':
      return `${baseClass} top-4 right-4`;
    case 'bottom-left':
      return `${baseClass} bottom-4 left-4`;
    case 'bottom-right':
      return `${baseClass} bottom-4 right-4`;
    default:
      return `${baseClass} top-4 right-4`;
  }
};

/**
 * 获取 Toast 样式类
 */
const getToastClass = (type: ToastType): string => {
  const baseClass = 'alert max-w-sm pointer-events-auto transform transition-all duration-300 opacity-0 translate-y-2';
  
  switch (type) {
    case 'success':
      return `${baseClass} alert-success`;
    case 'error':
      return `${baseClass} alert-error`;
    case 'warning':
      return `${baseClass} alert-warning`;
    case 'info':
    default:
      return `${baseClass} alert-info`;
  }
};

/**
 * 获取图标
 */
const getIcon = (type: ToastType): string => {
  switch (type) {
    case 'success':
      return '<svg class="w-5 h-5 text-success" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>';
    case 'error':
      return '<svg class="w-5 h-5 text-error" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>';
    case 'warning':
      return '<svg class="w-5 h-5 text-warning" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"></path></svg>';
    case 'info':
    default:
      return '<svg class="w-5 h-5 text-info" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>';
  }
};

/**
 * HTML 转义
 */
const escapeHtml = (text: string): string => {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
};

// 添加 CSS 样式
const addToastStyles = (): void => {
  if (document.getElementById('toast-styles')) return;
  
  const style = document.createElement('style');
  style.id = 'toast-styles';
  style.textContent = `
    .toast-enter {
      opacity: 1 !important;
      transform: translateY(0) !important;
    }
    
    .toast-exit {
      opacity: 0 !important;
      transform: translateY(-8px) !important;
    }
  `;
  
  document.head.appendChild(style);
};

// 初始化样式
addToastStyles();

// 便捷方法
export const showSuccess = (message: string, options?: ToastOptions) => 
  showToast(message, 'success', options);

export const showError = (message: string, options?: ToastOptions) => 
  showToast(message, 'error', options);

export const showWarning = (message: string, options?: ToastOptions) => 
  showToast(message, 'warning', options);

export const showInfo = (message: string, options?: ToastOptions) => 
  showToast(message, 'info', options);