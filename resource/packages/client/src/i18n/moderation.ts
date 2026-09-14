import type { Locale } from "./auth.js";

const en = {
  title: "Moderator workspace",
  description: "Handle community management tasks within your scope.",
  notice:
    "With great power comes great responsibility. Check the rules and full context before acting.",
  tabs: {
    reports: "Reports",
    ban: "Blocked posts",
    logs: "Activity log",
    guidance: "Reminders",
  },
  reports: {
    loading: "Loading reports…",
    loadMore: "Load more",
    noMore: "No more reports",
    emptyTitle: "No reports",
    emptyDescription: "User reports will appear here.",
    statusTabs: { open: "Pending", closed: "Handled" },
    targetTypes: { topic: "Topic", post: "Post" },
    reasons: {
      spam: "Spam",
      abuse: "Harassment",
      illegal: "Illegal content",
      irrelevant: "Off topic",
      other: "Other",
    },
    resolutions: { banned: "Blocked", ignored: "Ignored", resolved: "Handled" },
    reason: "Reason",
    status: "Status",
    reporter: "Reporter",
    handler: "Handler",
    submitted: "Submitted",
    handled: "Handled",
    noExcerpt: "No excerpt",
    block: "Block",
    ignore: "Ignore",
    loadFailed: "Failed to load reports.",
    actionFailed: "Moderation action failed.",
  },
  blocked: {
    emptyTitle: "No blocked posts",
    emptyDescription: "There is nothing to handle in your scope right now.",
    restore: "Restore",
    action: "Action",
    loadFailed: "Failed to update the post.",
    next: "Next page",
  },
  logs: {
    loading: "Loading logs…",
    loadMore: "Load more",
    noMore: "No more logs",
    emptyTitle: "No activity yet",
    emptyDescription: "Blocks, restores, and scope changes will appear here.",
    operation: "Operation",
    time: "Time",
    loadFailed: "Failed to load activity.",
    actions: {
      topicBlocked: "blocked",
      topicUnblocked: "restored",
      replyBlocked: "hid",
      replyUnblocked: "restored",
      postBlocked: "hid",
      postUnblocked: "restored",
      reportResolved: "resolved report",
      reportRejected: "ignored report",
      categoryModeratorAdded: "added moderator scope",
      categoryModeratorRemoved: "removed moderator scope",
      operation: "updated",
    },
  },
  guidance: {
    rule: {
      title: "Check the rules first",
      description:
        "Do not act on disagreement alone. Confirm a clear rules violation.",
    },
    context: {
      title: "Read the full context",
      description:
        "Review the title, body, category, and discussion background.",
    },
    restraint: {
      title: "Use power with restraint",
      description: "Prefer reminders and conversation when they can work.",
    },
  },
} as const;

type Widen<T> = T extends string ? string : { [K in keyof T]: Widen<T[K]> };
type Messages = Widen<typeof en>;

const zh: Messages = {
  title: "版主工作台",
  description: "集中处理你负责范围内的社区管理任务。",
  notice:
    "权力越大责任越大。处理前请对照规则、查看完整上下文，并克制使用权限。",
  tabs: {
    reports: "举报",
    ban: "封禁记录",
    logs: "操作日志",
    guidance: "管理提醒",
  },
  reports: {
    loading: "正在加载举报…",
    loadMore: "加载更多",
    noMore: "没有更多举报",
    emptyTitle: "暂无举报",
    emptyDescription: "普通用户提交举报后，会出现在这里。",
    statusTabs: { open: "待处理", closed: "已处理" },
    targetTypes: { topic: "主题", post: "帖子" },
    reasons: {
      spam: "垃圾广告",
      abuse: "攻击辱骂",
      illegal: "违法违规",
      irrelevant: "无关内容",
      other: "其他",
    },
    resolutions: { banned: "已封禁", ignored: "已忽略", resolved: "已处理" },
    reason: "原因",
    status: "状态",
    reporter: "提交人",
    handler: "处理人",
    submitted: "提交",
    handled: "处理",
    noExcerpt: "暂无摘要",
    block: "封禁",
    ignore: "忽略",
    loadFailed: "加载举报失败。",
    actionFailed: "管理操作失败。",
  },
  blocked: {
    emptyTitle: "暂无已封禁帖子",
    emptyDescription: "当前范围内没有需要处理的内容。",
    restore: "解封",
    action: "操作",
    loadFailed: "更新帖子失败。",
    next: "下一页",
  },
  logs: {
    loading: "正在加载日志…",
    loadMore: "加载更多",
    noMore: "没有更多日志",
    emptyTitle: "暂无操作日志",
    emptyDescription: "封禁、解封或版主范围调整后，会在这里留下记录。",
    operation: "操作",
    time: "时间",
    loadFailed: "加载操作日志失败。",
    actions: {
      topicBlocked: "封禁了",
      topicUnblocked: "解封了",
      replyBlocked: "隐藏了",
      replyUnblocked: "恢复了",
      postBlocked: "隐藏了",
      postUnblocked: "恢复了",
      reportResolved: "处理了举报",
      reportRejected: "忽略了举报",
      categoryModeratorAdded: "新增版主范围",
      categoryModeratorRemoved: "移除了版主范围",
      operation: "操作了",
    },
  },
  guidance: {
    rule: {
      title: "先对照规则",
      description: "不要因为观点不同而处理帖子，先确认是否明确违反社区规则。",
    },
    context: {
      title: "看完整上下文",
      description: "处理前查看标题、正文、分类和讨论背景，避免只凭片段决定。",
    },
    restraint: {
      title: "克制使用权限",
      description: "能提醒就先提醒，能沟通就先沟通。",
    },
  },
};

export const moderationResources: Record<Locale, Messages> = {
  zh,
  en,
  ja: {
    ...en,
    title: "モデレーターワークスペース",
    tabs: {
      reports: "報告",
      ban: "ブロック済み",
      logs: "操作ログ",
      guidance: "注意事項",
    },
  },
  it: {
    ...en,
    title: "Area moderazione",
    tabs: {
      reports: "Segnalazioni",
      ban: "Contenuti bloccati",
      logs: "Registro attività",
      guidance: "Promemoria",
    },
  },
};
