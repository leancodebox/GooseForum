export default {
  messages: {
    userUpdated: "ユーザー {userId} を更新：{changedFields}",
    contentReviewed:
      "{type} #{subjectId} を審査：{action}、バージョン {version}：{reason}",
    topicStatusChanged: "トピック「{title}」の状態を{status}に変更",
    topicPinWeightChanged:
      "トピック「{title}」の固定優先度 {oldPinWeight} → {pinWeight}",
    topicCategoriesChanged:
      "トピック「{title}」のカテゴリ {oldCategoryIds} → {categoryIds}",
    topicDeleted: "トピック「{title}」を削除",
    moderatorTopicStatusChanged:
      "モデレーターがトピック「{title}」を{status}に変更",
    categoryModeratorAdded:
      "カテゴリ「{categoryName}」にモデレーター {username} を追加",
    categoryModeratorRemoved:
      "カテゴリ「{categoryName}」からモデレーター {userId} を削除",
  },
  statusLabels: {
    blocked: "ブロック済み",
    unblocked: "通常",
  },
  contentTypes: {
    topic: "トピック",
    post: "返信",
  },
  actions: {
    approve: "承認",
    reject: "拒否",
    recheck: "再チェック",
  },
  fieldLabels: {
    status: "アカウント状態",
    activation: "認証状態",
    role: "ロール",
  },
} as const;
