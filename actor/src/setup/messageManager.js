import { createDiscreteApi } from 'naive-ui';

const { message } = createDiscreteApi(['message']);

export const enqueueMessage = (content, type = 'error') => {
  message[type](content);
};
