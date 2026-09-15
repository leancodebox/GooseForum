// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  title: "モデレーター管理",
  description: "担当範囲内のコミュニティ管理タスクを処理します。",
  notice:
    "大きな権限には大きな責任が伴います。封禁する前に、投稿がコミュニティルールに違反しているか確認し、必要に応じて先に対話してください。",
  guidanceTitle: "管理のリマインダー",
  guidanceDescription:
    "モデレーターの操作はコミュニティの空気に影響します。事実、ルール、影響を確認してから対応してください。",
  guidanceItems: {
    rule: {
      title: "まずルールを確認",
      description:
        "意見の違いだけで対応せず、投稿がコミュニティルールに明確に違反しているか確認してください。",
    },
    context: {
      title: "文脈を読む",
      description:
        "判断する前に、タイトル、本文、カテゴリ、議論の背景をできるだけ確認してください。",
    },
    restraint: {
      title: "権限は控えめに使う",
      description:
        "注意や対話で解決できる場合はそちらを優先し、封禁は明確で必要な場合に限ります。",
    },
  },
  allTitle: "すべての記事",
  allDescription: "現在のカテゴリ内の記事を確認し、状態に応じて処理します。",
  blockedTitle: "封禁済みの記事",
  blockedDescription:
    "現在のカテゴリ内の封禁済み記事を確認し、必要に応じて解除します。",
  allEmptyTitle: "記事はありません",
  blockedEmptyTitle: "封禁済みの記事はありません",
  banTitle: "封禁できる記事",
  banDescription: "担当カテゴリ内の通常記事を確認し、必要に応じて封禁します。",
  unbanTitle: "封禁済みの記事",
  unbanDescription:
    "担当カテゴリ内の封禁済み記事を確認し、必要に応じて解除します。",
  banEmptyTitle: "封禁できる記事はありません",
  unbanEmptyTitle: "封禁済みの記事はありません",
  banAction: "封禁",
  unbanAction: "解除",
  total: "{count} 件",
  knownTotal: "{count} 件",
  knownTotalMore: "{count}+ 件",
  blocked: "封禁済み",
  emptyDescription: "現在の範囲には処理する内容がありません。",
  managementTabs: {
    reports: "通報",
    ban: "封禁済み",
    logs: "操作ログ",
    guidance: "リマインダー",
    posts: "モデレーション",
    members: "メンバー管理",
    settings: "ルール設定",
  },
  reports: {
    loading: "通報を読み込み中",
    loadMore: "さらに読み込む",
    noMore: "これ以上の通報はありません",
    emptyTitle: "未処理の通報はありません",
    emptyDescription: "ユーザーからの通報がここに表示されます。",
    statusTabs: {
      open: "未処理",
      closed: "処理済み",
    },
    table: {
      report: "通報",
      reason: "通報情報",
      people: "担当",
      time: "時間",
      action: "操作",
    },
    ban: "封禁",
    hide: "封禁",
    reject: "無視",
    resolve: "処理済み",
    reasonLabel: "理由",
    statusLabel: "状態",
    submittedAtLabel: "通報",
    handledAtLabel: "処理",
    reporterLabel: "通報者",
    handlerLabel: "処理者",
    noExcerpt: "抜粋はありません",
    targetTypes: {
      topic: "トピック",
      post: "投稿",
    },
    reasons: {
      spam: "スパム",
      abuse: "攻撃・嫌がらせ",
      illegal: "違法コンテンツ",
      irrelevant: "無関係な内容",
      other: "その他",
    },
    resolutions: {
      banned: "封禁済み",
      ignored: "無視済み",
      resolved: "処理済み",
    },
  },
  logs: {
    title: "操作ログ",
    description:
      "モデレーションコンソールの操作を時系列で確認し、後から判断を追跡できます。",
    loading: "ログを読み込み中",
    loadMore: "さらに読み込む",
    noMore: "これ以上のログはありません",
    emptyTitle: "操作ログはありません",
    emptyDescription:
      "封禁、解除、モデレーター範囲の変更がここに記録されます。",
    table: {
      operation: "操作",
      time: "時間",
    },
    actions: {
      topicBlocked: "封禁しました",
      topicUnblocked: "解除しました",
      replyBlocked: "非表示にしました",
      replyUnblocked: "復元しました",
      postBlocked: "非表示にしました",
      postUnblocked: "復元しました",
      reportResolved: "通報を処理しました",
      reportRejected: "通報を無視しました",
      categoryModeratorAdded: "モデレーター範囲を追加しました",
      categoryModeratorRemoved: "モデレーター範囲を削除しました",
      operation: "操作しました",
    },
  },
  tabs: {
    all: "すべて",
    blocked: "封禁済み",
    ban: "すべて",
    unban: "封禁済み",
  },
  table: {
    topic: "記事",
    updatedAt: "更新日時",
    author: "作成者",
    activity: "データ",
    action: "操作",
  },
  meta: {
    views: "閲覧",
    replies: "返信",
  },
} as const;
