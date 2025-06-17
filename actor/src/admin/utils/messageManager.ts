import { createDiscreteApi } from 'naive-ui';

const { message } = createDiscreteApi(['message']);

/**
 * 显示消息提示
 * @param content 消息内容
 * @param type 消息类型，默认为 'error'
 */
export const enqueueMessage = (content: string, type: 'success' | 'error' | 'info' | 'warning' = 'error') => {
  message[type](content);
};

/**
 * 显示成功消息
 * @param content 消息内容
 */
export const showSuccessMessage = (content: string) => {
  enqueueMessage(content, 'success');
};

/**
 * 显示错误消息
 * @param content 消息内容
 */
export const showErrorMessage = (content: string) => {
  enqueueMessage(content, 'error');
};

/**
 * 显示信息消息
 * @param content 消息内容
 */
export const showInfoMessage = (content: string) => {
  enqueueMessage(content, 'info');
};

/**
 * 显示警告消息
 * @param content 消息内容
 */
export const showWarningMessage = (content: string) => {
  enqueueMessage(content, 'warning');
};
