// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  presetsTitle: "预设",
  presetApplied: "已应用 {name} 预设，保存草稿后保留",
  saveDraft: "保存草稿",
  publishSite: "发布到全站",
  workflowTitle: "使用说明：",
  workflowDescription:
    "当前编辑的是主题草稿。保存草稿不会影响全站；发布到全站后，只有左侧启用开关打开时，自定义主题才会接管站点。",
  draftSaved: "主题草稿已保存，不会影响全站。",
  published: "主题已发布到全站。",
  saveFailed: "主题草稿保存失败",
  publishFailed: "主题发布失败",
  restoredToSaved: "已还原到已保存配置",
  themeRestoredDefault: "{name} 已恢复内置默认",
  allRestoredDefault: "已恢复为内置默认主题，保存预发布后可发布到全站",
  cssCopied: "CSS 已复制",
  cssCopyFailed: "CSS 复制失败",
  pageTitle: "主题预览设置",
  resetDefault: "默认",
  restoreToSavedTitle: "还原到已保存配置",
  enableLabel: "启用",
  tabLatest: "最新",
  tabHot: "热门",
  tabFeatured: "精华",
  samplePublish: "发布主题",
  sampleTopic1: "主题系统重构讨论：颜色、圆角和组件状态",
  sampleTopic2: "Markdown 正文在深色模式下的可读性",
  sampleTopic3: "新用户引导和消息通知的视觉检查",
  sampleTopicExcerpt: "观察弱文本、分隔线、标签和悬停背景在当前主题下的层级。",
  sampleTopicTitle: "一篇主题帖的标题",
  sampleTopicBody:
    "这里模拟正文、链接、引用和代码块。主题变量应该让内容在明暗模式下都清晰，尤其是正文、边框和弱文本。",
  sampleTopicQuote: "颜色层级应该安静，但不能含糊。",
  sampleMessageIncoming: "这个背景还舒服吗？",
  sampleMessageOutgoing: "对比度需要稳。",
  sampleTextarea: "主题变量覆盖输入、焦点和正文。",
  presets: {
    goose: {
      label: "Goose",
      description: "当前默认，干净稳妥",
    },
    clean: {
      label: "Clean",
      description: "偏产品后台，低饱和蓝绿",
    },
    warm: {
      label: "Warm",
      description: "柔和暖色，社区感更强",
    },
    deep: {
      label: "Deep",
      description: "黑为主，强调更深邃",
    },
    cupcake: {
      label: "Cupcake",
      description: "粉色甜点感，柔和轻盈",
    },
    retro: {
      label: "Retro",
      description: "复古暖调，适合轻松社区",
    },
    synthwave: {
      label: "Synthwave",
      description: "霓虹紫蓝，更有视觉张力",
    },
    aqua: {
      label: "Aqua",
      description: "清爽水色，明快通透",
    },
    forest: {
      label: "Forest",
      description: "墨绿层级，适合知识社区",
    },
  },
} as const;
