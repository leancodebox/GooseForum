import type { Locale } from "./auth.js";

interface SiteMessages {
  shell: {
    openMenu: string;
    closeMenu: string;
    menu: string;
    search: string;
    switchLanguage: string;
    switchLight: string;
    switchDark: string;
    login: string;
    register: string;
    logout: string;
    publish: string;
    settings: string;
    accessGroups: string;
    themePreview: string;
    admin: string;
    resources: string;
    categories: string;
    nav: Record<
      | "topics"
      | "hot"
      | "popular"
      | "categories"
      | "members"
      | "messages"
      | "notifications"
      | "drafts"
      | "moderation"
      | "links"
      | "sponsors",
      string
    >;
  };
  links: {
    title: string;
    subtitle: string;
    emptyTitle: string;
    emptyDescription: string;
    applyTitle: string;
    applyDescription: string;
    applyAction: string;
    principlesTitle: string;
    principles: { healthy: string; relevant: string; stable: string };
  };
  sponsors: {
    defaultMessage: string;
    emptyTitle: string;
    emptyDescription: string;
    rulesTitle: string;
  };
  categories: {
    title: string;
    subtitle: string;
    total: string;
    topicCount: string;
    noDescription: string;
    emptyTitle: string;
    emptyDescription: string;
  };
  members: {
    title: string;
    subtitle: string;
    noBio: string;
    prestige: string;
    topics: string;
    replies: string;
    joinedAt: string;
    pagination: string;
    previous: string;
    next: string;
    emptyTitle: string;
    emptyDescription: string;
  };
  home: {
    latest: string;
    latestReplies: string;
    latestPublished: string;
    hot: string;
    popular: string;
    newTopic: string;
    topic: string;
    users: string;
    replies: string;
    views: string;
    activity: string;
    pinned: string;
    emptyTitle: string;
    emptyDescription: string;
    emptyCategoryDescription: string;
    categoryLabel: string;
    loadMore: string;
    loading: string;
    allShown: string;
    previous: string;
    next: string;
    currentPage: string;
    pagination: string;
    waterfall: string;
    switchMode: string;
    emailTitle: string;
    emailDescription: string;
    emailAction: string;
    announcement: string;
    markRead: string;
    autoLoadFailed: string;
    justNow: string;
    minuteAgo: string;
    hourAgo: string;
    dayAgo: string;
  };
  search: {
    title: string;
    label: string;
    resultCount: string;
    emptyPrompt: string;
    inputPlaceholder: string;
    action: string;
    page: string;
    noResultsTitle: string;
    noResultsDescription: string;
    startTitle: string;
    startDescription: string;
  };
  user: {
    emptyBio: string;
    noBio: string;
    online: string;
    editProfile: string;
    follow: string;
    following: string;
    joinedAt: string;
    lastActive: string;
    emptyTopics: string;
    emptyActivity: string;
    emptyData: string;
    loading: string;
    message: string;
    followFailed: string;
    loadFailed: string;
    profileNavigation: string;
    activityNavigation: string;
    tabs: Record<
      | "summary"
      | "topics"
      | "likes"
      | "activity"
      | "timeline"
      | "badges"
      | "following"
      | "followers",
      string
    >;
    summarySections: Record<
      "recentTopics" | "recentBadges" | "recentActivity",
      string
    >;
    activity: Record<
      "signup" | "post" | "like" | "follow" | "comment" | "default",
      string
    >;
    stats: Record<
      | "topics"
      | "replies"
      | "reputation"
      | "likesReceived"
      | "likesGiven"
      | "followers"
      | "following"
      | "bookmarks",
      string
    >;
  };
  userCard: {
    unavailable: string;
    online: string;
    activeAt: string;
    loading: string;
    stats: Record<"topics" | "replies" | "likes" | "followers", string>;
    joinedAt: string;
    following: string;
    viewProfile: string;
  };
}

export const siteResources: Record<Locale, SiteMessages> = {
  zh: {
    shell: {
      openMenu: "打开菜单",
      closeMenu: "关闭菜单",
      menu: "菜单",
      search: "搜索",
      switchLanguage: "切换语言",
      switchLight: "切换到浅色主题",
      switchDark: "切换到深色主题",
      login: "登录",
      register: "注册",
      logout: "退出登录",
      publish: "发布",
      settings: "设置",
      accessGroups: "访问组",
      themePreview: "主题预览",
      admin: "管理后台",
      resources: "资源",
      categories: "分类",
      nav: {
        topics: "主题",
        hot: "热门",
        popular: "流行",
        categories: "全部分类",
        members: "所有成员",
        messages: "私信",
        notifications: "通知",
        drafts: "草稿箱",
        moderation: "版主管理",
        links: "友情链接",
        sponsors: "赞助商",
      },
    },
    links: {
      title: "友情链接",
      subtitle: "社区伙伴、开源项目和值得访问的友好站点。",
      emptyTitle: "暂无链接",
      emptyDescription: "站点还没有配置友情链接。",
      applyTitle: "申请友链",
      applyDescription:
        "发帖提交网站名称、描述、地址和 Logo，管理员审核后会出现在这里。",
      applyAction: "去发帖申请",
      principlesTitle: "收录原则",
      principles: {
        healthy: "内容健康、长期可访问。",
        relevant: "优先技术、开源、社区相关站点。",
        stable: "站点信息清晰，Logo 可稳定加载。",
      },
    },
    sponsors: {
      defaultMessage: "感谢支持 GooseForum。",
      emptyTitle: "暂无赞助商",
      emptyDescription: "站点还没有配置赞助信息。",
      rulesTitle: "展示规则",
    },
    categories: {
      title: "全部分类",
      subtitle: "浏览社区的讨论空间，找到你感兴趣的话题。",
      total: "{count} 个分类",
      topicCount: "{count} 个主题",
      noDescription: "这个分类还没有介绍。",
      emptyTitle: "暂无分类",
      emptyDescription: "目前没有你可以浏览的分类。",
    },
    members: {
      title: "所有成员",
      subtitle: "认识正在共同建设这个社区的人。",
      noBio: "这位成员还没有填写个人简介。",
      prestige: "声望",
      topics: "主题",
      replies: "回复",
      joinedAt: "{date} 加入",
      pagination: "成员分页",
      previous: "上一页",
      next: "下一页",
      emptyTitle: "暂无成员",
      emptyDescription: "社区还没有可以展示的成员。",
    },
    home: {
      latest: "最新",
      latestReplies: "最新回复",
      latestPublished: "最新发布",
      hot: "热门",
      popular: "流行",
      newTopic: "新建主题",
      topic: "主题",
      users: "用户",
      replies: "回复",
      views: "浏览",
      activity: "活动",
      pinned: "置顶",
      emptyTitle: "暂无主题",
      emptyDescription: "社区还没有可以展示的主题。",
      emptyCategoryDescription: "这个分类还没有公开主题。",
      categoryLabel: "Category",
      loadMore: "加载更多",
      loading: "加载中…",
      allShown: "已显示全部",
      previous: "上一页",
      next: "下一页",
      currentPage: "第 {page} 页",
      pagination: "分页",
      waterfall: "瀑布流",
      switchMode: "切换到{mode}",
      emailTitle: "请验证邮箱",
      emailDescription: "验证后即可使用完整社区功能。",
      emailAction: "前往设置",
      announcement: "公告",
      markRead: "标记公告已读",
      autoLoadFailed: "自动加载失败，请手动重试。",
      justNow: "刚刚",
      minuteAgo: "{count} 分钟前",
      hourAgo: "{count} 小时前",
      dayAgo: "{count} 天前",
    },
    search: {
      title: "搜索",
      label: "Discover",
      resultCount: "{count} 个结果",
      emptyPrompt: "搜索主题、关键词和社区讨论。",
      inputPlaceholder: "搜索主题...",
      action: "搜索",
      page: "第 {page} / {total} 页",
      noResultsTitle: "没有找到结果",
      noResultsDescription: "换个关键词，或者缩短搜索词再试试。",
      startTitle: "开始搜索",
      startDescription: "输入主题标题、关键词或短语来搜索论坛。",
    },
    user: {
      emptyBio: "这个用户还没有留下简介。",
      noBio: "暂无简介",
      online: "在线",
      editProfile: "编辑资料",
      follow: "关注",
      following: "已关注",
      joinedAt: "加入于 {date}",
      lastActive: "最后活跃 {time}",
      emptyTopics: "还没有发布主题。",
      emptyActivity: "暂无动态。",
      emptyData: "暂无数据。",
      loading: "加载中…",
      message: "私信",
      followFailed: "关注操作失败，请稍后重试。",
      loadFailed: "加载失败，请稍后重试。",
      profileNavigation: "用户资料",
      activityNavigation: "用户动态",
      tabs: {
        summary: "总结",
        topics: "主题",
        likes: "赞",
        activity: "动态",
        timeline: "所有",
        badges: "徽章",
        following: "关注",
        followers: "粉丝",
      },
      summarySections: {
        recentTopics: "近期主题",
        recentBadges: "近期徽章",
        recentActivity: "近期动态",
      },
      activity: {
        signup: "加入论坛",
        post: "发布主题",
        like: "点赞内容",
        follow: "关注用户",
        comment: "参与回复",
        default: "活动",
      },
      stats: {
        topics: "主题",
        replies: "回复",
        reputation: "声望",
        likesReceived: "获赞",
        likesGiven: "点赞",
        followers: "粉丝",
        following: "关注",
        bookmarks: "收藏",
      },
    },
    userCard: {
      unavailable: "用户资料暂时不可用",
      online: "在线",
      activeAt: "活跃于 {time}",
      loading: "加载用户资料",
      stats: {
        topics: "主题",
        replies: "回复",
        likes: "获赞",
        followers: "关注者",
      },
      joinedAt: "加入于 {date}",
      following: "已关注",
      viewProfile: "查看主页",
    },
  },
  en: {
    shell: {
      openMenu: "Open menu",
      closeMenu: "Close menu",
      menu: "Menu",
      search: "Search",
      switchLanguage: "Switch language",
      switchLight: "Switch to light theme",
      switchDark: "Switch to dark theme",
      login: "Log in",
      register: "Sign up",
      logout: "Log out",
      publish: "Publish",
      settings: "Settings",
      accessGroups: "Access groups",
      themePreview: "Theme preview",
      admin: "Admin",
      resources: "Resources",
      categories: "Categories",
      nav: {
        topics: "Topics",
        hot: "Hot",
        popular: "Popular",
        categories: "All categories",
        members: "Members",
        messages: "Messages",
        notifications: "Notifications",
        drafts: "Drafts",
        moderation: "Moderation",
        links: "Links",
        sponsors: "Sponsors",
      },
    },
    links: {
      title: "Links",
      subtitle:
        "Community partners, open-source projects, and friendly sites worth visiting.",
      emptyTitle: "No links",
      emptyDescription: "No friendly links have been configured yet.",
      applyTitle: "Apply for a link",
      applyDescription:
        "Post your site name, description, URL, and logo. It will appear here after admin review.",
      applyAction: "Post an application",
      principlesTitle: "Listing principles",
      principles: {
        healthy: "Healthy content and long-term availability.",
        relevant: "Tech, open-source, and community sites are preferred.",
        stable: "Clear site information and a stable logo URL.",
      },
    },
    sponsors: {
      defaultMessage: "Thanks for supporting GooseForum.",
      emptyTitle: "No sponsors yet",
      emptyDescription: "Sponsor information has not been configured yet.",
      rulesTitle: "Display rules",
    },
    categories: {
      title: "All categories",
      subtitle:
        "Browse community spaces and find conversations that interest you.",
      total: "{count} categories",
      topicCount: "{count} topics",
      noDescription: "This category does not have a description yet.",
      emptyTitle: "No categories",
      emptyDescription: "There are no categories available to you right now.",
    },
    members: {
      title: "All members",
      subtitle: "Meet the people building this community together.",
      noBio: "This member has not added a bio yet.",
      prestige: "Prestige",
      topics: "Topics",
      replies: "Replies",
      joinedAt: "Joined {date}",
      pagination: "Member pages",
      previous: "Previous",
      next: "Next",
      emptyTitle: "No members yet",
      emptyDescription: "There are no members to show yet.",
    },
    home: {
      latest: "Latest",
      latestReplies: "Latest replies",
      latestPublished: "Newest posts",
      hot: "Hot",
      popular: "Popular",
      newTopic: "New topic",
      topic: "Topic",
      users: "Users",
      replies: "Replies",
      views: "Views",
      activity: "Activity",
      pinned: "Pinned",
      emptyTitle: "No topics yet",
      emptyDescription: "There are no topics to show yet.",
      emptyCategoryDescription: "This category has no public topics yet.",
      categoryLabel: "Category",
      loadMore: "Load more",
      loading: "Loading…",
      allShown: "All topics shown",
      previous: "Previous",
      next: "Next",
      currentPage: "Page {page}",
      pagination: "Pagination",
      waterfall: "Waterfall",
      switchMode: "Switch to {mode}",
      emailTitle: "Verify your email",
      emailDescription: "Verify it to use all community features.",
      emailAction: "Open settings",
      announcement: "Announcement",
      markRead: "Mark announcement as read",
      autoLoadFailed: "Automatic loading failed. Try again.",
      justNow: "Just now",
      minuteAgo: "{count}m ago",
      hourAgo: "{count}h ago",
      dayAgo: "{count}d ago",
    },
    search: {
      title: "Search",
      label: "Discover",
      resultCount: "{count} results",
      emptyPrompt: "Search topics, keywords, and community discussions.",
      inputPlaceholder: "Search topics...",
      action: "Search",
      page: "Page {page} of {total}",
      noResultsTitle: "No results",
      noResultsDescription: "Try another keyword or shorten your query.",
      startTitle: "Start searching",
      startDescription:
        "Enter a topic title, keyword, or phrase to search the forum.",
    },
    user: {
      emptyBio: "This user has not added a bio yet.",
      noBio: "No bio",
      online: "Online",
      editProfile: "Edit profile",
      follow: "Follow",
      following: "Following",
      joinedAt: "Joined {date}",
      lastActive: "Last active {time}",
      emptyTopics: "No topics posted yet.",
      emptyActivity: "No activity yet.",
      emptyData: "No data.",
      loading: "Loading…",
      message: "Message",
      followFailed: "Could not update follow status. Try again.",
      loadFailed: "Could not load more. Try again.",
      profileNavigation: "User profile",
      activityNavigation: "User activity",
      tabs: {
        summary: "Summary",
        topics: "Topics",
        likes: "Likes",
        activity: "Activity",
        timeline: "All",
        badges: "Badges",
        following: "Following",
        followers: "Followers",
      },
      summarySections: {
        recentTopics: "Recent topics",
        recentBadges: "Recent badges",
        recentActivity: "Recent activity",
      },
      activity: {
        signup: "Joined the forum",
        post: "Published a topic",
        like: "Liked content",
        follow: "Followed a user",
        comment: "Joined the discussion",
        default: "Activity",
      },
      stats: {
        topics: "Topics",
        replies: "Replies",
        reputation: "Reputation",
        likesReceived: "Likes",
        likesGiven: "Liked",
        followers: "Followers",
        following: "Following",
        bookmarks: "Bookmarks",
      },
    },
    userCard: {
      unavailable: "User profile is temporarily unavailable",
      online: "Online",
      activeAt: "Active {time}",
      loading: "Loading user profile",
      stats: {
        topics: "Topics",
        replies: "Replies",
        likes: "Likes",
        followers: "Followers",
      },
      joinedAt: "Joined {date}",
      following: "Following",
      viewProfile: "View profile",
    },
  },
  ja: {
    shell: {
      openMenu: "メニューを開く",
      closeMenu: "メニューを閉じる",
      menu: "メニュー",
      search: "検索",
      switchLanguage: "言語を切り替え",
      switchLight: "ライトテーマに切り替え",
      switchDark: "ダークテーマに切り替え",
      login: "ログイン",
      register: "登録",
      logout: "ログアウト",
      publish: "投稿",
      settings: "設定",
      accessGroups: "アクセスグループ",
      themePreview: "テーマプレビュー",
      admin: "管理画面",
      resources: "リソース",
      categories: "カテゴリー",
      nav: {
        topics: "トピック",
        hot: "人気",
        popular: "注目",
        categories: "すべてのカテゴリー",
        members: "メンバー",
        messages: "メッセージ",
        notifications: "通知",
        drafts: "下書き",
        moderation: "モデレーション",
        links: "リンク",
        sponsors: "スポンサー",
      },
    },
    links: {
      title: "リンク",
      subtitle:
        "コミュニティパートナー、オープンソースプロジェクト、おすすめサイト。",
      emptyTitle: "リンクはありません",
      emptyDescription: "リンクはまだ設定されていません。",
      applyTitle: "リンクを申請",
      applyDescription:
        "サイト名、説明、URL、ロゴを投稿してください。審査後に掲載されます。",
      applyAction: "申請を投稿",
      principlesTitle: "掲載基準",
      principles: {
        healthy: "健全なコンテンツで長期間アクセスできること。",
        relevant: "技術、オープンソース、コミュニティ関連を優先します。",
        stable: "サイト情報が明確でロゴを安定して読み込めること。",
      },
    },
    sponsors: {
      defaultMessage: "GooseForum へのご支援ありがとうございます。",
      emptyTitle: "スポンサーはいません",
      emptyDescription: "スポンサー情報はまだ設定されていません。",
      rulesTitle: "掲載ルール",
    },
    categories: {
      title: "すべてのカテゴリー",
      subtitle: "コミュニティを巡り、興味のある会話を見つけましょう。",
      total: "{count} カテゴリー",
      topicCount: "{count} トピック",
      noDescription: "このカテゴリーにはまだ説明がありません。",
      emptyTitle: "カテゴリーはありません",
      emptyDescription: "現在閲覧できるカテゴリーはありません。",
    },
    members: {
      title: "すべてのメンバー",
      subtitle: "このコミュニティを一緒につくる人たちを知りましょう。",
      noBio: "このメンバーはまだ自己紹介を追加していません。",
      prestige: "評価",
      topics: "トピック",
      replies: "返信",
      joinedAt: "{date} に参加",
      pagination: "メンバー一覧のページ",
      previous: "前へ",
      next: "次へ",
      emptyTitle: "メンバーはまだいません",
      emptyDescription: "表示できるメンバーはいません。",
    },
    home: {
      latest: "最新",
      latestReplies: "最新の返信",
      latestPublished: "新着投稿",
      hot: "人気",
      popular: "注目",
      newTopic: "新規トピック",
      topic: "トピック",
      users: "ユーザー",
      replies: "返信",
      views: "閲覧",
      activity: "アクティビティ",
      pinned: "固定",
      emptyTitle: "トピックはありません",
      emptyDescription: "表示できるトピックはありません。",
      emptyCategoryDescription: "このカテゴリーには公開トピックがありません。",
      categoryLabel: "Category",
      loadMore: "さらに読み込む",
      loading: "読み込み中…",
      allShown: "すべて表示しました",
      previous: "前へ",
      next: "次へ",
      currentPage: "{page} ページ",
      pagination: "ページ送り",
      waterfall: "連続表示",
      switchMode: "{mode} に切り替え",
      emailTitle: "メールを確認してください",
      emailDescription: "すべての機能を利用するには確認が必要です。",
      emailAction: "設定へ",
      announcement: "お知らせ",
      markRead: "既読にする",
      autoLoadFailed: "自動読み込みに失敗しました。",
      justNow: "たった今",
      minuteAgo: "{count} 分前",
      hourAgo: "{count} 時間前",
      dayAgo: "{count} 日前",
    },
    search: {
      title: "検索",
      label: "Discover",
      resultCount: "{count} 件",
      emptyPrompt: "トピック、キーワード、コミュニティの会話を検索します。",
      inputPlaceholder: "トピックを検索...",
      action: "検索",
      page: "{page} / {total} ページ",
      noResultsTitle: "結果がありません",
      noResultsDescription: "別のキーワードまたは短い検索語をお試しください。",
      startTitle: "検索を開始",
      startDescription: "トピック名、キーワード、フレーズを入力してください。",
    },
    user: {
      emptyBio: "このユーザーはまだ自己紹介を書いていません。",
      noBio: "自己紹介なし",
      online: "オンライン",
      editProfile: "プロフィール編集",
      follow: "フォロー",
      following: "フォロー中",
      joinedAt: "{date} に参加",
      lastActive: "最終アクティブ {time}",
      emptyTopics: "まだトピックを投稿していません。",
      emptyActivity: "アクティビティはありません。",
      emptyData: "データはありません。",
      loading: "読み込み中…",
      message: "メッセージ",
      followFailed: "フォロー状態を更新できませんでした。",
      loadFailed: "読み込みに失敗しました。",
      profileNavigation: "ユーザープロフィール",
      activityNavigation: "ユーザーアクティビティ",
      tabs: {
        summary: "概要",
        topics: "トピック",
        likes: "いいね",
        activity: "アクティビティ",
        timeline: "すべて",
        badges: "バッジ",
        following: "フォロー中",
        followers: "フォロワー",
      },
      summarySections: {
        recentTopics: "最近のトピック",
        recentBadges: "最近のバッジ",
        recentActivity: "最近のアクティビティ",
      },
      activity: {
        signup: "フォーラムに参加しました",
        post: "トピックを投稿しました",
        like: "コンテンツにいいねしました",
        follow: "ユーザーをフォローしました",
        comment: "返信に参加しました",
        default: "アクティビティ",
      },
      stats: {
        topics: "トピック",
        replies: "返信",
        reputation: "評判",
        likesReceived: "いいね",
        likesGiven: "いいね済み",
        followers: "フォロワー",
        following: "フォロー中",
        bookmarks: "ブックマーク",
      },
    },
    userCard: {
      unavailable: "ユーザー情報を利用できません",
      online: "オンライン",
      activeAt: "{time} にアクティブ",
      loading: "ユーザー情報を読み込み中",
      stats: {
        topics: "トピック",
        replies: "返信",
        likes: "いいね",
        followers: "フォロワー",
      },
      joinedAt: "{date} に参加",
      following: "フォロー中",
      viewProfile: "プロフィールを見る",
    },
  },
  it: {
    shell: {
      openMenu: "Apri menu",
      closeMenu: "Chiudi menu",
      menu: "Menu",
      search: "Cerca",
      switchLanguage: "Cambia lingua",
      switchLight: "Passa al tema chiaro",
      switchDark: "Passa al tema scuro",
      login: "Accedi",
      register: "Registrati",
      logout: "Esci",
      publish: "Pubblica",
      settings: "Impostazioni",
      accessGroups: "Gruppi di accesso",
      themePreview: "Anteprima tema",
      admin: "Amministrazione",
      resources: "Risorse",
      categories: "Categorie",
      nav: {
        topics: "Discussioni",
        hot: "Popolari",
        popular: "In evidenza",
        categories: "Tutte le categorie",
        members: "Membri",
        messages: "Messaggi",
        notifications: "Notifiche",
        drafts: "Bozze",
        moderation: "Moderazione",
        links: "Link",
        sponsors: "Sponsor",
      },
    },
    links: {
      title: "Link",
      subtitle:
        "Partner della comunità, progetti open source e siti consigliati.",
      emptyTitle: "Nessun link",
      emptyDescription: "Non sono ancora stati configurati link.",
      applyTitle: "Proponi un link",
      applyDescription:
        "Pubblica nome, descrizione, URL e logo del sito. Apparirà dopo la revisione.",
      applyAction: "Invia una proposta",
      principlesTitle: "Criteri di pubblicazione",
      principles: {
        healthy: "Contenuti sicuri e disponibilità a lungo termine.",
        relevant: "Preferenza per tecnologia, open source e comunità.",
        stable: "Informazioni chiare e URL del logo stabile.",
      },
    },
    sponsors: {
      defaultMessage: "Grazie per il supporto a GooseForum.",
      emptyTitle: "Nessuno sponsor",
      emptyDescription:
        "Le informazioni sugli sponsor non sono ancora configurate.",
      rulesTitle: "Regole di visualizzazione",
    },
    categories: {
      title: "Tutte le categorie",
      subtitle:
        "Esplora gli spazi della community e trova le discussioni che ti interessano.",
      total: "{count} categorie",
      topicCount: "{count} argomenti",
      noDescription: "Questa categoria non ha ancora una descrizione.",
      emptyTitle: "Nessuna categoria",
      emptyDescription: "Al momento non ci sono categorie disponibili.",
    },
    members: {
      title: "Tutti i membri",
      subtitle: "Conosci le persone che costruiscono insieme questa community.",
      noBio: "Questo membro non ha ancora aggiunto una biografia.",
      prestige: "Prestigio",
      topics: "Argomenti",
      replies: "Risposte",
      joinedAt: "Iscritto {date}",
      pagination: "Pagine dei membri",
      previous: "Precedente",
      next: "Successivo",
      emptyTitle: "Nessun membro",
      emptyDescription: "Non ci sono ancora membri da mostrare.",
    },
    home: {
      latest: "Recenti",
      latestReplies: "Ultime risposte",
      latestPublished: "Nuove pubblicazioni",
      hot: "Popolari",
      popular: "In evidenza",
      newTopic: "Nuovo argomento",
      topic: "Argomento",
      users: "Utenti",
      replies: "Risposte",
      views: "Visualizzazioni",
      activity: "Attività",
      pinned: "Fissato",
      emptyTitle: "Nessun argomento",
      emptyDescription: "Non ci sono argomenti da mostrare.",
      emptyCategoryDescription:
        "Questa categoria non contiene argomenti pubblici.",
      categoryLabel: "Category",
      loadMore: "Carica altro",
      loading: "Caricamento…",
      allShown: "Tutti gli argomenti visualizzati",
      previous: "Precedente",
      next: "Successivo",
      currentPage: "Pagina {page}",
      pagination: "Paginazione",
      waterfall: "Flusso continuo",
      switchMode: "Passa a {mode}",
      emailTitle: "Verifica la tua email",
      emailDescription: "Verificala per usare tutte le funzioni.",
      emailAction: "Apri impostazioni",
      announcement: "Annuncio",
      markRead: "Segna come letto",
      autoLoadFailed: "Caricamento automatico non riuscito.",
      justNow: "Adesso",
      minuteAgo: "{count} min fa",
      hourAgo: "{count} h fa",
      dayAgo: "{count} g fa",
    },
    search: {
      title: "Cerca",
      label: "Discover",
      resultCount: "{count} risultati",
      emptyPrompt: "Cerca argomenti, parole chiave e discussioni.",
      inputPlaceholder: "Cerca argomenti...",
      action: "Cerca",
      page: "Pagina {page} di {total}",
      noResultsTitle: "Nessun risultato",
      noResultsDescription:
        "Prova un'altra parola chiave o una ricerca più breve.",
      startTitle: "Inizia a cercare",
      startDescription: "Inserisci un titolo, una parola chiave o una frase.",
    },
    user: {
      emptyBio: "Questo utente non ha ancora aggiunto una biografia.",
      noBio: "Nessuna biografia",
      online: "Online",
      editProfile: "Modifica profilo",
      follow: "Segui",
      following: "Segue",
      joinedAt: "Iscritto {date}",
      lastActive: "Ultima attività {time}",
      emptyTopics: "Nessun argomento pubblicato.",
      emptyActivity: "Nessuna attività.",
      emptyData: "Nessun dato.",
      loading: "Caricamento…",
      message: "Messaggio",
      followFailed: "Impossibile aggiornare lo stato. Riprova.",
      loadFailed: "Caricamento non riuscito. Riprova.",
      profileNavigation: "Profilo utente",
      activityNavigation: "Attività utente",
      tabs: {
        summary: "Riepilogo",
        topics: "Argomenti",
        likes: "Mi piace",
        activity: "Attività",
        timeline: "Tutto",
        badges: "Badge",
        following: "Segue",
        followers: "Follower",
      },
      summarySections: {
        recentTopics: "Argomenti recenti",
        recentBadges: "Badge recenti",
        recentActivity: "Attività recente",
      },
      activity: {
        signup: "Si è iscritto al forum",
        post: "Ha pubblicato un argomento",
        like: "Ha messo mi piace a un contenuto",
        follow: "Ha seguito un utente",
        comment: "Ha partecipato alla discussione",
        default: "Attività",
      },
      stats: {
        topics: "Argomenti",
        replies: "Risposte",
        reputation: "Reputazione",
        likesReceived: "Mi piace",
        likesGiven: "Mi piace dati",
        followers: "Follower",
        following: "Segue",
        bookmarks: "Salvati",
      },
    },
    userCard: {
      unavailable: "Profilo utente non disponibile",
      online: "Online",
      activeAt: "Attivo {time}",
      loading: "Caricamento profilo",
      stats: {
        topics: "Argomenti",
        replies: "Risposte",
        likes: "Mi piace",
        followers: "Follower",
      },
      joinedAt: "Iscritto {date}",
      following: "Segue",
      viewProfile: "Vedi profilo",
    },
  },
};
