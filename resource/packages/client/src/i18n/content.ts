import type { Locale } from "./auth.js";

const en = {
  drafts: {
    title: "Drafts",
    total: "{count} drafts",
    summary:
      "Unpublished topics stay here so you can continue editing anytime.",
    newDraft: "New draft",
    edit: "Continue editing",
    untitled: "Untitled draft",
    blocked: "Blocked",
    emptyTitle: "No drafts yet",
    emptyHint: "Save unfinished topics as drafts and come back later.",
    emptyDescription:
      "This draft has no summary yet. Continue editing to generate one.",
    table: { draft: "Draft", updatedAt: "Updated", action: "Action" },
    meta: { createdAt: "Created", views: "Views", replies: "Replies" },
  },
  accessGroups: {
    joinTitle: "Apply to access groups",
    joinDescription:
      "Join a group to access the restricted categories granted to it.",
    noJoinableGroups: "No groups accept applications",
    noJoinableGroupsHint:
      "Invitation-only groups require an administrator to add you.",
    unlocksCategories: "Provides access to: {categories}",
    noCategoryGrants: "This group has no category grants yet",
    joined: "Joined",
    pending: "Awaiting review",
    apply: "Apply to join",
    approve: "Approve",
    reject: "Reject",
    applicationsToReview: "Applications to review",
    managerReviewHint: "You manage these access groups.",
    noPendingApplications: "No pending applications",
    loadFailed: "Failed to load access control",
    memberSaveFailed: "Failed to save member",
    applicationFailed: "Failed to submit the application",
  },
  error: {
    notFoundTitle: "Page not found",
    fallbackMessage: "The page could not be loaded.",
    back: "Back",
    home: "Home",
    messages: {
      "page.notFound": "This page does not exist or has been removed.",
      "route.notFound":
        "The route is not defined. Check the URL and request method.",
    },
  },
  common: { loading: "Loading…", saving: "Saving…" },
} as const;

type Widen<T> = T extends string ? string : { [K in keyof T]: Widen<T[K]> };
type ContentMessages = Widen<typeof en>;

const zh: ContentMessages = {
  drafts: {
    title: "草稿箱",
    total: "{count} 条草稿",
    summary: "这里保存你尚未发布的主题，可以随时继续编辑。",
    newDraft: "新建草稿",
    edit: "继续编辑",
    untitled: "未命名草稿",
    blocked: "已封禁",
    emptyTitle: "还没有草稿",
    emptyHint: "写到一半的主题可以先存成草稿，之后再回来继续。",
    emptyDescription: "这篇草稿还没有摘要，继续编辑后会自动生成。",
    table: { draft: "草稿", updatedAt: "最近更新", action: "操作" },
    meta: { createdAt: "创建于", views: "浏览", replies: "回复" },
  },
  accessGroups: {
    joinTitle: "申请加入访问组",
    joinDescription: "加入后可以访问该组获授权的受限分类。",
    noJoinableGroups: "暂无可申请的访问组",
    noJoinableGroupsHint: "仅邀请访问组需要由管理员添加成员。",
    unlocksCategories: "可访问分类：{categories}",
    noCategoryGrants: "该组暂未关联分类",
    joined: "已加入",
    pending: "等待审核",
    apply: "申请加入",
    approve: "批准",
    reject: "拒绝",
    applicationsToReview: "待审核申请",
    managerReviewHint: "你是以下访问组的管理员。",
    noPendingApplications: "暂无待审核申请",
    loadFailed: "加载访问控制失败",
    memberSaveFailed: "保存成员失败",
    applicationFailed: "提交加入申请失败",
  },
  error: {
    notFoundTitle: "页面不存在",
    fallbackMessage: "页面加载失败。",
    back: "返回",
    home: "回到首页",
    messages: {
      "page.notFound": "页面不存在，或已经被删除。",
      "route.notFound": "路由未定义，请确认 URL 和请求方法是否正确。",
    },
  },
  common: { loading: "加载中…", saving: "保存中…" },
};

export const contentResources: Record<Locale, ContentMessages> = {
  zh,
  en,
  ja: {
    ...en,
    drafts: {
      ...en.drafts,
      title: "下書き",
      total: "下書き {count} 件",
      newDraft: "新しい下書き",
      edit: "編集を続ける",
      untitled: "無題の下書き",
      emptyTitle: "下書きはありません",
    },
    accessGroups: {
      ...en.accessGroups,
      joinTitle: "アクセスグループへの参加申請",
      joinDescription:
        "グループに参加すると、許可された制限付きカテゴリにアクセスできます。",
      noJoinableGroups: "申請可能なグループはありません",
      joined: "参加済み",
      pending: "審査待ち",
      apply: "参加を申請",
      approve: "承認",
      reject: "却下",
    },
    error: {
      ...en.error,
      notFoundTitle: "ページが見つかりません",
      back: "戻る",
      home: "ホーム",
    },
    common: { loading: "読み込み中…", saving: "保存中…" },
  },
  it: {
    ...en,
    drafts: {
      ...en.drafts,
      title: "Bozze",
      total: "{count} bozze",
      newDraft: "Nuova bozza",
      edit: "Continua a modificare",
      untitled: "Bozza senza titolo",
      emptyTitle: "Nessuna bozza",
    },
    accessGroups: {
      ...en.accessGroups,
      joinTitle: "Richiedi accesso ai gruppi",
      joinDescription:
        "Unisciti a un gruppo per accedere alle categorie riservate autorizzate.",
      noJoinableGroups: "Nessun gruppo accetta richieste",
      joined: "Iscritto",
      pending: "In attesa di revisione",
      apply: "Richiedi accesso",
      approve: "Approva",
      reject: "Rifiuta",
    },
    error: {
      ...en.error,
      notFoundTitle: "Pagina non trovata",
      back: "Indietro",
      home: "Home",
    },
    common: { loading: "Caricamento…", saving: "Salvataggio…" },
  },
};
