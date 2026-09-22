export default {
  messages: {
    userUpdated: "Обновлён пользователь {userId}: изменены поля {changedFields}",
    contentReviewed:
      "Проверено {type} #{subjectId}: {action}, версия {version}: {reason}",
    topicStatusChanged: 'Статус темы "{title}" изменён на {status}',
    topicPinWeightChanged:
      'Вес закрепления темы "{title}": {oldPinWeight} → {pinWeight}',
    topicCategoriesChanged:
      'Категории темы "{title}": {oldCategoryIds} → {categoryIds}',
    topicDeleted: 'Тема "{title}" удалена',
    moderatorTopicStatusChanged:
      'Модератор изменил статус темы "{title}" на {status}',
    categoryModeratorAdded:
      'Добавлен модератор {username} в категорию "{categoryName}"',
    categoryModeratorRemoved:
      'Удален модератор {userId} из категории "{categoryName}"',
  },
  statusLabels: {
    blocked: "заблокирован",
    unblocked: "обычный",
  },
  contentTypes: {
    topic: "тема",
    post: "ответ",
  },
  actions: {
    approve: "одобрить",
    reject: "отклонить",
    recheck: "отправить на повторную проверку",
  },
  fieldLabels: {
    status: "статус аккаунта",
    activation: "статус активации",
    role: "роль",
  },
} as const;
