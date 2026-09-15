export default {
  messages: {
    userUpdated: "更新用户 {userId}：{changedFields}",
    contentReviewed:
      "审核{type} #{subjectId}，操作{action}，版本 {version}：{reason}",
    topicStatusChanged: "主题「{title}」状态调整为{status}",
    topicPinWeightChanged:
      "主题「{title}」置顶权重 {oldPinWeight} -> {pinWeight}",
    topicCategoriesChanged:
      "主题「{title}」分类 {oldCategoryIds} -> {categoryIds}",
    topicDeleted: "删除主题「{title}」",
    moderatorTopicStatusChanged: "版主将主题「{title}」调整为{status}",
    categoryModeratorAdded: "为分类「{categoryName}」添加版主 {username}",
    categoryModeratorRemoved: "移除分类「{categoryName}」的版主 {userId}",
  },
  statusLabels: {
    blocked: "封禁",
    unblocked: "正常",
  },
  contentTypes: {
    topic: "主题",
    post: "回复",
  },
  actions: {
    approve: "通过",
    reject: "拒绝",
    recheck: "重新检测",
  },
  fieldLabels: {
    status: "账号状态",
    activation: "验证状态",
    role: "角色",
  },
} as const;
