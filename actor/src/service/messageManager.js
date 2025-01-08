import { createDiscreteApi } from 'naive-ui';

const { message } = createDiscreteApi(['message']);

const messageQueue = [];
let isDisplaying = false;

const displayNextMessage = () => {
  if (messageQueue.length === 0) {
    isDisplaying = false;
    return;
  }

  isDisplaying = true;
  const { content, type } = messageQueue.shift();
  message[type](content);

  setTimeout(() => {
    displayNextMessage();
  }, 3000); // 3秒后显示下一个消息
};

export const enqueueMessage = (content, type = 'error') => {
  messageQueue.push({ content, type });
  if (!isDisplaying) {
    displayNextMessage();
  }
}; 