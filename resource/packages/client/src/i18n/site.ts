import zh from "./messages/zh-site.js";
import en from "./messages/en-site.js";
import ja from "./messages/ja-site.js";
import it from "./messages/it-site.js";
import type { Locale } from "./auth.js";

interface SiteMessages {
  shell: {
    more: string;
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
    profile: string;
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
  zh,
  en,
  ja,
  it,
} as const;
