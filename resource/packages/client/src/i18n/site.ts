import type { Locale } from './auth.js'

interface SiteMessages {
  shell: {
    openMenu: string
    closeMenu: string
    menu: string
    search: string
    switchLanguage: string
    switchLight: string
    switchDark: string
    login: string
    register: string
    logout: string
    publish: string
    settings: string
    accessGroups: string
    themePreview: string
    admin: string
    resources: string
    categories: string
    nav: Record<'topics' | 'hot' | 'popular' | 'categories' | 'members' | 'messages' | 'notifications' | 'drafts' | 'moderation' | 'links' | 'sponsors', string>
  }
  links: {
    title: string
    subtitle: string
    emptyTitle: string
    emptyDescription: string
    applyTitle: string
    applyDescription: string
    applyAction: string
    principlesTitle: string
    principles: { healthy: string, relevant: string, stable: string }
  }
  sponsors: { defaultMessage: string, emptyTitle: string, emptyDescription: string, rulesTitle: string }
  categories: { title: string, subtitle: string, total: string, topicCount: string, noDescription: string, emptyTitle: string, emptyDescription: string }
  members: { title: string, subtitle: string, noBio: string, prestige: string, topics: string, replies: string, joinedAt: string, pagination: string, previous: string, next: string, emptyTitle: string, emptyDescription: string }
}

export const siteResources: Record<Locale, SiteMessages> = {
  zh: {
    shell: {
      openMenu: '打开菜单', closeMenu: '关闭菜单', menu: '菜单', search: '搜索', switchLanguage: '切换语言', switchLight: '切换到浅色主题', switchDark: '切换到深色主题',
      login: '登录', register: '注册', logout: '退出登录', publish: '发布', settings: '设置', accessGroups: '访问组', themePreview: '主题预览', admin: '管理后台', resources: '资源', categories: '分类',
      nav: { topics: '主题', hot: '热门', popular: '流行', categories: '全部分类', members: '所有成员', messages: '私信', notifications: '通知', drafts: '草稿箱', moderation: '版主管理', links: '友情链接', sponsors: '赞助商' },
    },
    links: {
      title: '友情链接', subtitle: '社区伙伴、开源项目和值得访问的友好站点。', emptyTitle: '暂无链接', emptyDescription: '站点还没有配置友情链接。',
      applyTitle: '申请友链', applyDescription: '发帖提交网站名称、描述、地址和 Logo，管理员审核后会出现在这里。', applyAction: '去发帖申请', principlesTitle: '收录原则',
      principles: { healthy: '内容健康、长期可访问。', relevant: '优先技术、开源、社区相关站点。', stable: '站点信息清晰，Logo 可稳定加载。' },
    },
    sponsors: { defaultMessage: '感谢支持 GooseForum。', emptyTitle: '暂无赞助商', emptyDescription: '站点还没有配置赞助信息。', rulesTitle: '展示规则' },
    categories: { title: '全部分类', subtitle: '浏览社区的讨论空间，找到你感兴趣的话题。', total: '{count} 个分类', topicCount: '{count} 个主题', noDescription: '这个分类还没有介绍。', emptyTitle: '暂无分类', emptyDescription: '目前没有你可以浏览的分类。' },
    members: { title: '所有成员', subtitle: '认识正在共同建设这个社区的人。', noBio: '这位成员还没有填写个人简介。', prestige: '声望', topics: '主题', replies: '回复', joinedAt: '{date} 加入', pagination: '成员分页', previous: '上一页', next: '下一页', emptyTitle: '暂无成员', emptyDescription: '社区还没有可以展示的成员。' },
  },
  en: {
    shell: {
      openMenu: 'Open menu', closeMenu: 'Close menu', menu: 'Menu', search: 'Search', switchLanguage: 'Switch language', switchLight: 'Switch to light theme', switchDark: 'Switch to dark theme',
      login: 'Log in', register: 'Sign up', logout: 'Log out', publish: 'Publish', settings: 'Settings', accessGroups: 'Access groups', themePreview: 'Theme preview', admin: 'Admin', resources: 'Resources', categories: 'Categories',
      nav: { topics: 'Topics', hot: 'Hot', popular: 'Popular', categories: 'All categories', members: 'Members', messages: 'Messages', notifications: 'Notifications', drafts: 'Drafts', moderation: 'Moderation', links: 'Links', sponsors: 'Sponsors' },
    },
    links: {
      title: 'Links', subtitle: 'Community partners, open-source projects, and friendly sites worth visiting.', emptyTitle: 'No links', emptyDescription: 'No friendly links have been configured yet.',
      applyTitle: 'Apply for a link', applyDescription: 'Post your site name, description, URL, and logo. It will appear here after admin review.', applyAction: 'Post an application', principlesTitle: 'Listing principles',
      principles: { healthy: 'Healthy content and long-term availability.', relevant: 'Tech, open-source, and community sites are preferred.', stable: 'Clear site information and a stable logo URL.' },
    },
    sponsors: { defaultMessage: 'Thanks for supporting GooseForum.', emptyTitle: 'No sponsors yet', emptyDescription: 'Sponsor information has not been configured yet.', rulesTitle: 'Display rules' },
    categories: { title: 'All categories', subtitle: 'Browse community spaces and find conversations that interest you.', total: '{count} categories', topicCount: '{count} topics', noDescription: 'This category does not have a description yet.', emptyTitle: 'No categories', emptyDescription: 'There are no categories available to you right now.' },
    members: { title: 'All members', subtitle: 'Meet the people building this community together.', noBio: 'This member has not added a bio yet.', prestige: 'Prestige', topics: 'Topics', replies: 'Replies', joinedAt: 'Joined {date}', pagination: 'Member pages', previous: 'Previous', next: 'Next', emptyTitle: 'No members yet', emptyDescription: 'There are no members to show yet.' },
  },
  ja: {
    shell: {
      openMenu: 'メニューを開く', closeMenu: 'メニューを閉じる', menu: 'メニュー', search: '検索', switchLanguage: '言語を切り替え', switchLight: 'ライトテーマに切り替え', switchDark: 'ダークテーマに切り替え',
      login: 'ログイン', register: '登録', logout: 'ログアウト', publish: '投稿', settings: '設定', accessGroups: 'アクセスグループ', themePreview: 'テーマプレビュー', admin: '管理画面', resources: 'リソース', categories: 'カテゴリー',
      nav: { topics: 'トピック', hot: '人気', popular: '注目', categories: 'すべてのカテゴリー', members: 'メンバー', messages: 'メッセージ', notifications: '通知', drafts: '下書き', moderation: 'モデレーション', links: 'リンク', sponsors: 'スポンサー' },
    },
    links: {
      title: 'リンク', subtitle: 'コミュニティパートナー、オープンソースプロジェクト、おすすめサイト。', emptyTitle: 'リンクはありません', emptyDescription: 'リンクはまだ設定されていません。',
      applyTitle: 'リンクを申請', applyDescription: 'サイト名、説明、URL、ロゴを投稿してください。審査後に掲載されます。', applyAction: '申請を投稿', principlesTitle: '掲載基準',
      principles: { healthy: '健全なコンテンツで長期間アクセスできること。', relevant: '技術、オープンソース、コミュニティ関連を優先します。', stable: 'サイト情報が明確でロゴを安定して読み込めること。' },
    },
    sponsors: { defaultMessage: 'GooseForum へのご支援ありがとうございます。', emptyTitle: 'スポンサーはいません', emptyDescription: 'スポンサー情報はまだ設定されていません。', rulesTitle: '掲載ルール' },
    categories: { title: 'すべてのカテゴリー', subtitle: 'コミュニティを巡り、興味のある会話を見つけましょう。', total: '{count} カテゴリー', topicCount: '{count} トピック', noDescription: 'このカテゴリーにはまだ説明がありません。', emptyTitle: 'カテゴリーはありません', emptyDescription: '現在閲覧できるカテゴリーはありません。' },
    members: { title: 'すべてのメンバー', subtitle: 'このコミュニティを一緒につくる人たちを知りましょう。', noBio: 'このメンバーはまだ自己紹介を追加していません。', prestige: '評価', topics: 'トピック', replies: '返信', joinedAt: '{date} に参加', pagination: 'メンバー一覧のページ', previous: '前へ', next: '次へ', emptyTitle: 'メンバーはまだいません', emptyDescription: '表示できるメンバーはいません。' },
  },
  it: {
    shell: {
      openMenu: 'Apri menu', closeMenu: 'Chiudi menu', menu: 'Menu', search: 'Cerca', switchLanguage: 'Cambia lingua', switchLight: 'Passa al tema chiaro', switchDark: 'Passa al tema scuro',
      login: 'Accedi', register: 'Registrati', logout: 'Esci', publish: 'Pubblica', settings: 'Impostazioni', accessGroups: 'Gruppi di accesso', themePreview: 'Anteprima tema', admin: 'Amministrazione', resources: 'Risorse', categories: 'Categorie',
      nav: { topics: 'Discussioni', hot: 'Popolari', popular: 'In evidenza', categories: 'Tutte le categorie', members: 'Membri', messages: 'Messaggi', notifications: 'Notifiche', drafts: 'Bozze', moderation: 'Moderazione', links: 'Link', sponsors: 'Sponsor' },
    },
    links: {
      title: 'Link', subtitle: 'Partner della comunità, progetti open source e siti consigliati.', emptyTitle: 'Nessun link', emptyDescription: 'Non sono ancora stati configurati link.',
      applyTitle: 'Proponi un link', applyDescription: 'Pubblica nome, descrizione, URL e logo del sito. Apparirà dopo la revisione.', applyAction: 'Invia una proposta', principlesTitle: 'Criteri di pubblicazione',
      principles: { healthy: 'Contenuti sicuri e disponibilità a lungo termine.', relevant: 'Preferenza per tecnologia, open source e comunità.', stable: 'Informazioni chiare e URL del logo stabile.' },
    },
    sponsors: { defaultMessage: 'Grazie per il supporto a GooseForum.', emptyTitle: 'Nessuno sponsor', emptyDescription: 'Le informazioni sugli sponsor non sono ancora configurate.', rulesTitle: 'Regole di visualizzazione' },
    categories: { title: 'Tutte le categorie', subtitle: 'Esplora gli spazi della community e trova le discussioni che ti interessano.', total: '{count} categorie', topicCount: '{count} argomenti', noDescription: 'Questa categoria non ha ancora una descrizione.', emptyTitle: 'Nessuna categoria', emptyDescription: 'Al momento non ci sono categorie disponibili.' },
    members: { title: 'Tutti i membri', subtitle: 'Conosci le persone che costruiscono insieme questa community.', noBio: 'Questo membro non ha ancora aggiunto una biografia.', prestige: 'Prestigio', topics: 'Argomenti', replies: 'Risposte', joinedAt: 'Iscritto {date}', pagination: 'Pagine dei membri', previous: 'Precedente', next: 'Successivo', emptyTitle: 'Nessun membro', emptyDescription: 'Non ci sono ancora membri da mostrare.' },
  },
}
