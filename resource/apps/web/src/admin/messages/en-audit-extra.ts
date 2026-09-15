export default {
  messages: {
    userUpdated: "Updated user {userId}: {changedFields}",
    contentReviewed:
      "Reviewed {type} #{subjectId}: {action}, version {version}: {reason}",
    topicStatusChanged: 'Topic "{title}" status changed to {status}',
    topicPinWeightChanged:
      'Topic "{title}" pin weight {oldPinWeight} -> {pinWeight}',
    topicCategoriesChanged:
      'Topic "{title}" categories {oldCategoryIds} -> {categoryIds}',
    topicDeleted: 'Deleted topic "{title}"',
    moderatorTopicStatusChanged:
      'Moderator changed topic "{title}" to {status}',
    categoryModeratorAdded:
      'Added moderator {username} to category "{categoryName}"',
    categoryModeratorRemoved:
      'Removed moderator {userId} from category "{categoryName}"',
  },
  statusLabels: {
    blocked: "blocked",
    unblocked: "normal",
  },
  contentTypes: {
    topic: "topic",
    post: "reply",
  },
  actions: {
    approve: "approve",
    reject: "reject",
    recheck: "recheck",
  },
  fieldLabels: {
    status: "account status",
    activation: "activation status",
    role: "role",
  },
} as const;
