import type { AuthLocale } from "@gooseforum/react/i18n/auth";

const en = {
  title: "Audit log",
  description: "Review administrator operations and affected resources.",
  refresh: "Refresh",
  loading: "Loading…",
  loadFailed: "Failed to load audit records",
  empty: "No audit records",
  operator: "Operator",
  operation: "Operation",
  targetType: "Target type",
  targetId: "Target ID",
  details: "Details",
  time: "Time",
  previous: "Previous",
  next: "Next",
  page: "Page",
  editUser: "User operation",
  editTopic: "Edit topic",
  editCategory: "Edit category",
  system: "System",
  user: "User",
  topic: "Topic",
  docProject: "Document project",
  docVersion: "Document version",
  docContent: "Document content",
  category: "Category",
  unknown: "Unknown",
} as const;
const zh: Record<keyof typeof en, string> = {
  title: "操作记录",
  description: "查看管理员操作及其影响的资源。",
  refresh: "刷新",
  loading: "正在加载…",
  loadFailed: "操作记录加载失败",
  empty: "暂无操作记录",
  operator: "操作人",
  operation: "操作类型",
  targetType: "目标类型",
  targetId: "目标 ID",
  details: "操作详情",
  time: "时间",
  previous: "上一页",
  next: "下一页",
  page: "第",
  editUser: "操作用户",
  editTopic: "编辑主题",
  editCategory: "编辑分类",
  system: "系统",
  user: "用户",
  topic: "主题",
  docProject: "文档项目",
  docVersion: "文档版本",
  docContent: "文档内容",
  category: "分类",
  unknown: "未知",
};
const resources = {
  zh,
  en,
  ja: { ...en, title: "操作履歴" },
  it: {
    ...en,
    title: "Registro operazioni",
    editUser: "Operazione utente",
    editTopic: "Modifica topic",
    editCategory: "Modifica categoria",
    system: "Sistema",
    user: "Utente",
    topic: "Topic",
    docProject: "Progetto documento",
    docVersion: "Versione documento",
    docContent: "Contenuto documento",
    category: "Categoria",
  },
} as const;

const messageKeys: Record<string, MessageKey> = {
  "admin.opt.content.reviewed": "contentReviewed",
  "admin.opt.user.updated": "userUpdated",
  "admin.opt.topic.statusChanged": "topicStatusChanged",
  "admin.opt.topic.pinWeightChanged": "topicPinWeightChanged",
  "admin.opt.topic.categoriesChanged": "topicCategoriesChanged",
  "admin.opt.topic.deleted": "topicDeleted",
  "moderator.opt.topic.statusChanged": "moderatorTopicStatusChanged",
  "admin.opt.category.moderatorAdded": "categoryModeratorAdded",
  "admin.opt.category.moderatorRemoved": "categoryModeratorRemoved",
};

const messages = {
  zh: {
    userUpdated: "更新用户 {userId}：{changedFields}",
    contentReviewed:
      "审核 {type} #{subjectId}，操作 {action}，版本 {version}：{reason}",
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
  en: {
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
  ja: {
    userUpdated: "Updated user {userId}: {changedFields}",
    contentReviewed:
      "{type} #{subjectId} の審査：{action}、バージョン {version}：{reason}",
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
  it: {
    userUpdated: "Utente {userId} aggiornato: {changedFields}",
    contentReviewed:
      "Moderazione {type} #{subjectId}: {action}, versione {version}: {reason}",
    topicStatusChanged: 'Stato del topic "{title}" cambiato in {status}',
    topicPinWeightChanged:
      'Peso di fissaggio del topic "{title}" {oldPinWeight} -> {pinWeight}',
    topicCategoriesChanged:
      'Categorie del topic "{title}" {oldCategoryIds} -> {categoryIds}',
    topicDeleted: 'Topic "{title}" eliminato',
    moderatorTopicStatusChanged:
      'Il moderatore ha cambiato il topic "{title}" in {status}',
    categoryModeratorAdded:
      'Aggiunto il moderatore {username} alla categoria "{categoryName}"',
    categoryModeratorRemoved:
      'Rimosso il moderatore {userId} dalla categoria "{categoryName}"',
  },
} as const;

const statusLabels = {
  zh: { blocked: "封禁", unblocked: "正常" },
  en: { blocked: "blocked", unblocked: "normal" },
  ja: { blocked: "blocked", unblocked: "normal" },
  it: { blocked: "bloccato", unblocked: "normale" },
} as const;
const fieldLabels = {
  zh: { status: "账号状态", activation: "验证状态", role: "角色" },
  en: {
    status: "account status",
    activation: "activation status",
    role: "role",
  },
  ja: {
    status: "account status",
    activation: "activation status",
    role: "role",
  },
  it: {
    status: "stato account",
    activation: "stato attivazione",
    role: "ruolo",
  },
} as const;

type MessageKey = keyof typeof messages.en;
export type AuditTextKey = keyof typeof en;

export function createAuditText(locale: AuthLocale) {
  const dictionary = resources[locale];
  return (key: AuditTextKey) => dictionary[key];
}

export function formatAuditMessage(
  locale: AuthLocale,
  messageCode: string,
  params: Record<string, unknown>,
  fallback: string,
) {
  const key = messageKeys[messageCode];
  if (!key) return fallback || messageCode;
  const values = normalizeParams(locale, params);
  return messages[locale][key].replace(
    /\{(\w+)\}/g,
    (_match, name: string) => values[name] ?? "",
  );
}

function normalizeParams(locale: AuthLocale, params: Record<string, unknown>) {
  const values: Record<string, string> = {};
  for (const [key, value] of Object.entries(params))
    values[key] = Array.isArray(value)
      ? value.join(", ")
      : value == null
        ? ""
        : String(value);
  if (
    typeof params.status === "string" &&
    params.status in statusLabels[locale]
  )
    values.status =
      statusLabels[locale][params.status as keyof typeof statusLabels.en];
  if (Array.isArray(params.changes))
    values.changedFields = params.changes
      .map(
        (field) =>
          fieldLabels[locale][String(field) as keyof typeof fieldLabels.en] ||
          String(field),
      )
      .join(", ");
  return values;
}
