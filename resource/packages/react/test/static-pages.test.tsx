import { afterEach, describe, expect, it, vi } from "vitest";
import {
  cleanup,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type {
  AnyPagePayload,
  GooseSiteApi,
  LayoutPayload,
  SettingsPageProps,
  NotificationsPageProps,
  UserProfileProps,
} from "@gooseforum/client";
import { GooseApp } from "../src/app/root";
import { GooseI18nProvider } from "../src/i18n";
import { GooseRuntimeProvider, type GooseRuntime } from "../src/runtime";

afterEach(cleanup);

const layout = {
  site: {
    name: "GooseForum",
    description: "",
    logo: "",
    favicon: "",
    brandType: "default",
    brandText: "",
    brandImage: "",
  },
  viewer: {
    id: 0,
    username: "",
    email: "",
    avatarUrl: "",
    isAuthenticated: false,
    canAccessAdmin: false,
    isModerator: false,
    requiresEmailVerification: false,
    adminPermissions: [],
  },
  header: [
    {
      key: "sponsors",
      label: "Sponsors",
      i18nLabel: "shell.nav.sponsors",
      url: "/sponsors",
    },
    {
      key: "links",
      label: "Links",
      i18nLabel: "shell.nav.links",
      url: "/links",
    },
  ],
  sidebar: {
    activeKey: "links",
    categories: [
      { id: 4, label: "Coding", url: "/c/Coding/4", color: "#8241d6" },
    ],
  },
  footer: {
    links: [{ name: "RSS", url: "/rss.xml" }],
    primary: ["GooseForum © 2024"],
  },
  unread: { notifications: false, messages: false },
  theme: { enabled: true, current: "gf-light", themeColor: "#fbfdff" },
} satisfies LayoutPayload;

function renderPage(page: AnyPagePayload, api: Partial<GooseSiteApi> = {}) {
  const toggleTheme = vi.fn();
  const navigate = vi.fn();
  const runtime: GooseRuntime = {
    api: api as GooseSiteApi,
    currentUrl: page.url,
    isNavigating: false,
    locale: "zh",
    theme: "gf-light",
    navigate,
    queueFlash: vi.fn(),
    redirect: vi.fn(),
    refresh: vi.fn(),
    setLocale: vi.fn(),
    toggleTheme,
  };
  render(
    <GooseI18nProvider locale="zh">
      <GooseRuntimeProvider runtime={runtime}>
        <GooseApp page={page} />
      </GooseRuntimeProvider>
    </GooseI18nProvider>,
  );
  return { navigate, toggleTheme, user: userEvent.setup() };
}

function payload(
  component:
    | "home.index"
    | "links.index"
    | "sponsors.index"
    | "categories.index"
    | "members.index"
    | "category.index"
    | "search.index"
    | "user.profile"
    | "settings.index"
    | "notifications.index",
  props: unknown,
): AnyPagePayload {
  return {
    component,
    props,
    layout: {
      ...layout,
      sidebar: {
        ...layout.sidebar,
        activeKey: component.replace(".index", ""),
      },
    },
    meta: { title: "Static page" },
    url: `/${component.replace(".index", "")}`,
    version: "1.0",
  } as AnyPagePayload;
}

function userProfileProps(): UserProfileProps {
  const topic: UserProfileProps["topics"][number] = {
    id: 20,
    title: "Profile topic",
    description: "A recent discussion",
    url: "/p/profile-topic/20",
    author: { id: 7, username: "alice", avatarUrl: "" },
    participants: [{ id: 7, username: "alice", avatarUrl: "" }],
    categories: [
      { id: 1, name: "Coding", url: "/c/Coding/1", color: "#8241d6" },
    ],
    replyCount: 8,
    viewCount: 120,
    pinWeight: 0,
    processStatus: 0,
    activityText: "",
    lastUpdateTime: new Date().toISOString(),
    unseen: false,
  };
  return {
    user: {
      userId: 7,
      username: "alice",
      nickname: "Alice",
      avatarUrl: "/alice.webp",
      profileCoverUrl: "/cover.webp",
      bio: "Builds the forum.",
      signature: "",
      websiteName: "Alice's site",
      website: "https://example.com",
      prestige: 1_250,
      externalInformation: { github: { link: "https://github.com/alice" } },
      isAdmin: true,
      topicCount: 12,
      replyCount: 34,
      likeReceivedCount: 56,
      likeGivenCount: 21,
      followerCount: 9,
      followingCount: 4,
      collectionCount: 6,
      isOnline: true,
      isFollowing: false,
      isSelf: false,
      badges: [],
      wornBadge: null,
      lastActiveTime: new Date().toISOString(),
      createdAt: "2025-01-02T08:00:00Z",
    },
    section: "summary",
    activityTab: "timeline",
    tabs: [
      { key: "summary", url: "/u/7", active: true },
      { key: "activity", url: "/u/7?section=activity", active: false },
      { key: "badges", url: "/u/7?section=badges", active: false },
    ],
    activityTabs: [
      { key: "timeline", url: "/u/7?section=activity", active: true },
      { key: "topics", url: "/u/7?section=activity&tab=topics", active: false },
      { key: "likes", url: "/u/7?section=activity&tab=likes", active: false },
      {
        key: "following",
        url: "/u/7?section=activity&tab=following",
        active: false,
      },
      {
        key: "followers",
        url: "/u/7?section=activity&tab=followers",
        active: false,
      },
    ],
    pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" },
    badges: [
      {
        code: "contributor",
        type: "system",
        grantMode: "manual",
        name: "Contributor",
        description: "Contributed to the community",
        iconType: "preset",
        iconKey: "contributor",
        iconUrl: "/badge.svg",
        color: "blue",
        level: "special",
        isEnabled: true,
        isWearable: true,
        sortOrder: 1,
        source: "manual",
        reason: "Thanks",
        grantedAt: "2026-01-01",
      },
    ],
    topics: [topic],
    activities: [
      {
        id: 1,
        action: 2,
        subjectType: "topic",
        subjectId: 20,
        contentPreview: "Profile topic",
        url: "/p/profile-topic/20",
        label: "post",
        createdAt: "2026-01-03T08:00:00Z",
      },
    ],
    likes: [],
    following: [],
    followers: [],
    isOwnProfile: false,
    canMessage: true,
    canFollow: true,
    messageUrl: "/messages?user=7",
    settingsUrl: "/settings",
  };
}

function settingsProps(): SettingsPageProps {
  return {
    user: {
      id: 7,
      username: "alice",
      email: "alice@example.com",
      nickname: "Alice",
      locale: "zh",
      avatarUrl: "/alice.webp",
      profileCoverUrl: "/cover.webp",
      bio: "Builds the forum.",
      signature: "",
      websiteName: "Alice's site",
      website: "https://example.com",
      prestige: 1250,
      createdAt: "2025-01-02",
      externalInformation: { github: { link: "https://github.com/alice" } },
      wornBadgeCode: "",
      badges: [],
      wearableBadges: [],
      wornBadge: null,
    },
    stats: {
      topicCount: 12,
      replyCount: 34,
      followerCount: 9,
      followingCount: 4,
      likeReceivedCount: 56,
      likeGivenCount: 21,
      collectionCount: 6,
      createdAt: "2025-01-02",
    },
    tabs: [
      { key: "profile", url: "/settings", active: true },
      { key: "account", url: "/settings?tab=account", active: false },
      { key: "privacy", url: "/settings?tab=privacy", active: false },
      { key: "binding", url: "/settings?tab=binding", active: false },
      { key: "applications", url: "/settings?tab=applications", active: false },
    ],
  };
}

function notificationsProps(): NotificationsPageProps {
  return {
    total: 1,
    unreadCount: 1,
    notifications: [
      {
        id: 91,
        eventType: "comment",
        isRead: false,
        createdAt: "2026-09-14T08:00:00Z",
        title: "New comment",
        content: "Useful reply",
        actor: { id: 8, username: "bob", avatarUrl: "/bob.webp" },
        topic: { id: 20, title: "React migration", url: "/p/react/20" },
        payload: {
          actorId: 8,
          actorName: "bob",
          templateKey: "notifications.templates.comment",
          topicId: 20,
          topicTitle: "React migration",
        },
      },
    ],
    pagination: {
      page: 1,
      nextPage: 2,
      hasNext: false,
      nextUrl: "",
    },
  };
}

describe("AppShell and static pages", () => {
  it("renders the home topic hierarchy and pagination controls", () => {
    renderPage(
      payload("home.index", {
        sort: "",
        tabs: [
          { key: "latest", url: "/", active: true },
          { key: "hot", url: "/?sort=hot", active: false },
        ],
        topics: [
          {
            id: 9,
            title: "React migration",
            description: "Shared topic list",
            url: "/p/post/9",
            author: { id: 7, username: "alice", avatarUrl: "" },
            participants: [{ id: 7, username: "alice", avatarUrl: "" }],
            categories: [
              { id: 1, name: "Coding", url: "/c/Coding/1", color: "#8241d6" },
            ],
            replyCount: 12,
            viewCount: 650,
            pinWeight: 1,
            processStatus: 0,
            activityText: "",
            lastUpdateTime: new Date().toISOString(),
            unseen: true,
          },
        ],
        pagination: {
          page: 1,
          nextPage: 2,
          hasNext: true,
          nextUrl: "/?page=2",
        },
        announcement: { enabled: false, html: "" },
      }),
    );
    expect(screen.getByRole("link", { name: "React migration" })).toBeTruthy();
    expect(document.querySelector('a[href="/c/Coding/1"]')).toBeTruthy();
    expect(screen.getByText("hot")).toBeTruthy();
    expect(screen.getByRole("button", { name: "加载更多" })).toBeTruthy();
  });

  it("renders the complete user summary hierarchy", () => {
    renderPage(payload("user.profile", userProfileProps()));

    expect(
      screen.getByRole("heading", { level: 1, name: "Alice" }),
    ).toBeTruthy();
    expect(screen.getByText("@alice")).toBeTruthy();
    expect(screen.getByText("Admin")).toBeTruthy();
    expect(screen.getByText("在线")).toBeTruthy();
    expect(screen.getByText("1.3k")).toBeTruthy();
    expect(
      document.querySelector('a[href="/p/profile-topic/20"]'),
    ).toBeTruthy();
    expect(screen.getByText("Contributor")).toBeTruthy();
    expect(
      screen.getByRole("link", { name: "GitHub" }).getAttribute("rel"),
    ).toBe("noopener noreferrer ugc");
  });

  it("updates follow state through the shared site API", async () => {
    const follow = vi.fn().mockResolvedValue(true);
    const { user } = renderPage(payload("user.profile", userProfileProps()), {
      users: { follow } as unknown as GooseSiteApi["users"],
    });

    await user.click(screen.getByRole("button", { name: "关注" }));
    expect(follow).toHaveBeenCalledWith(7, false);
    expect(await screen.findByRole("button", { name: "已关注" })).toBeTruthy();
  });

  it("renders profile topic activity with the shared topic table", () => {
    const props = userProfileProps();
    props.section = "activity";
    props.activityTab = "topics";
    props.tabs = props.tabs.map((tab) => ({
      ...tab,
      active: tab.key === "activity",
    }));
    props.activityTabs = props.activityTabs.map((tab) => ({
      ...tab,
      active: tab.key === "topics",
    }));
    renderPage(payload("user.profile", props));

    expect(screen.getByRole("navigation", { name: "用户动态" })).toBeTruthy();
    expect(screen.getByRole("link", { name: "Profile topic" })).toBeTruthy();
    expect(
      within(screen.getByRole("navigation", { name: "用户动态" }))
        .getByRole("link", { name: "主题" })
        .getAttribute("aria-current"),
    ).toBe("page");
  });

  it("edits and saves the React settings profile", async () => {
    const saveInfo = vi.fn().mockResolvedValue(undefined);
    const { user } = renderPage(payload("settings.index", settingsProps()), {
      users: { saveInfo } as unknown as GooseSiteApi["users"],
    });

    expect(
      await screen.findByRole("heading", { level: 1, name: "Alice" }),
    ).toBeTruthy();
    expect(
      screen.getAllByRole("button", { name: "选择预设头像" }),
    ).toHaveLength(12);
    const nickname = screen.getByRole("textbox", { name: "显示名称" });
    await user.clear(nickname);
    await user.type(nickname, "Builder");
    await user.click(screen.getByRole("button", { name: "保存资料" }));
    expect(saveInfo).toHaveBeenCalledWith(
      expect.objectContaining({ nickname: "Builder", locale: "zh" }),
    );
    expect(await screen.findByText("资料已保存。")).toBeTruthy();
  });

  it("loads bindings and authorized apps only when their settings surface is opened", async () => {
    const oauthBindings = vi
      .fn()
      .mockResolvedValue([
        { key: "github", displayName: "GitHub", enabled: true, bound: true },
      ]);
    const oidcGrants = vi.fn().mockResolvedValue([
      {
        clientId: "docs",
        name: "Docs",
        scopes: ["openid", "profile"],
        grantedAt: "2026-01-02",
        enabled: true,
      },
    ]);
    const unbindOAuth = vi.fn().mockResolvedValue(undefined);
    const { user } = renderPage(payload("settings.index", settingsProps()), {
      users: {
        oauthBindings,
        oidcGrants,
        unbindOAuth,
      } as unknown as GooseSiteApi["users"],
    });

    expect(oauthBindings).not.toHaveBeenCalled();
    await user.click(await screen.findByRole("button", { name: "绑定" }));
    expect(await screen.findByText("GitHub")).toBeTruthy();
    expect(oidcGrants).toHaveBeenCalledOnce();
    await user.click(screen.getByRole("button", { name: "解除绑定" }));
    expect(unbindOAuth).toHaveBeenCalledWith("github");
    await user.click(screen.getByRole("button", { name: "授权应用" }));
    expect(await screen.findByText("Docs")).toBeTruthy();
    expect(screen.getByText("openid")).toBeTruthy();
  });

  it("marks an individual notification read optimistically", async () => {
    const markRead = vi.fn().mockResolvedValue(true);
    const { user } = renderPage(
      payload("notifications.index", notificationsProps()),
      {
        notifications: { markRead } as unknown as GooseSiteApi["notifications"],
      },
    );

    expect(
      await screen.findByRole("heading", { level: 1, name: "通知" }),
    ).toBeTruthy();
    expect(screen.getByRole("link", { name: "React migration" })).toBeTruthy();
    await user.click(screen.getByRole("button", { name: "标为已读" }));
    expect(markRead).toHaveBeenCalledWith(91);
    expect(screen.queryByText("1 未读")).toBeNull();
  });

  it("loads the unread notification filter on demand", async () => {
    const list = vi.fn().mockResolvedValue({
      items: [{ ...notificationsProps().notifications[0], id: 92 }],
      nextCursor: 0,
      hasNext: false,
      unreadCount: 1,
    });
    const { user } = renderPage(
      payload("notifications.index", notificationsProps()),
      {
        notifications: { list } as unknown as GooseSiteApi["notifications"],
      },
    );

    await user.click(await screen.findByRole("tab", { name: /未读/ }));
    await waitFor(() => expect(list).toHaveBeenCalledWith("unread", 0, 20));
    expect(
      await screen.findByRole("link", { name: "React migration" }),
    ).toBeTruthy();
  });

  it("loads a user card only after a topic participant avatar is clicked", async () => {
    const card = vi.fn().mockResolvedValue({
      ...userProfileProps().user,
      userId: 27,
      username: "hover-user",
      nickname: "Hover User",
      avatarUrl: "/hover.webp",
      badges: [],
    });
    const home = payload("home.index", {
      sort: "latest",
      tabs: [{ key: "latest", url: "/", active: true }],
      topics: [
        {
          ...userProfileProps().topics[0],
          participants: [
            { id: 27, username: "hover-user", avatarUrl: "/hover.webp" },
          ],
        },
      ],
      pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" },
      announcement: { enabled: false, html: "" },
    });
    const { navigate, user } = renderPage(home, {
      users: { card } as unknown as GooseSiteApi["users"],
    });

    const avatar = document.querySelector('a[title="hover-user"]');
    expect(avatar).toBeTruthy();
    expect(avatar?.classList.contains("hover:z-10")).toBe(true);
    expect(avatar?.classList.contains("hover:scale-110")).toBe(true);
    await user.hover(avatar as Element);
    expect(card).not.toHaveBeenCalled();
    await user.click(avatar as Element);
    expect(await screen.findByLabelText("Hover User")).toBeTruthy();
    expect(card).toHaveBeenCalledOnce();
    expect(navigate).not.toHaveBeenCalled();
    expect(screen.getByText("查看主页")).toBeTruthy();
  });

  it("reuses the topic list for a category without category chips or hot markers", () => {
    renderPage(
      payload("category.index", {
        category: {
          id: 4,
          name: "Coding",
          description: "Development topics",
          icon: "💻",
          color: "#8241d6",
          url: "/c/Coding/4",
        },
        sort: "latest",
        tabs: [{ key: "latest", url: "/c/Coding/4", active: true }],
        topics: [
          {
            id: 10,
            title: "Category topic",
            description: "Shared row",
            url: "/p/category-topic/10",
            author: { id: 7, username: "alice", avatarUrl: "" },
            participants: [{ id: 7, username: "alice", avatarUrl: "" }],
            categories: [
              {
                id: 9,
                name: "Hidden category",
                url: "/c/hidden/9",
                color: "#000",
              },
            ],
            replyCount: 3,
            viewCount: 900,
            pinWeight: 0,
            processStatus: 0,
            activityText: "",
            lastUpdateTime: new Date().toISOString(),
            unseen: false,
          },
        ],
        pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" },
      }),
    );

    expect(
      screen.getByRole("heading", { level: 1, name: "Coding" }),
    ).toBeTruthy();
    expect(screen.getByRole("link", { name: "最新回复" })).toBeTruthy();
    expect(screen.getByRole("link", { name: "Category topic" })).toBeTruthy();
    expect(screen.queryByText("Hidden category")).toBeNull();
    expect(screen.queryByText("hot")).toBeNull();
  });

  it("renders search results and submits through the SPA runtime", async () => {
    const { navigate, user } = renderPage(
      payload("search.index", {
        query: "react",
        total: 1,
        totalPages: 2,
        topics: [
          {
            id: 11,
            title: "Search result",
            description: "Found topic",
            url: "/p/search-result/11",
            author: { id: 7, username: "alice", avatarUrl: "" },
            participants: [{ id: 7, username: "alice", avatarUrl: "" }],
            categories: [],
            replyCount: 0,
            viewCount: 5,
            pinWeight: 0,
            processStatus: 0,
            activityText: "",
            lastUpdateTime: new Date().toISOString(),
            unseen: false,
          },
        ],
        pagination: {
          page: 1,
          nextPage: 2,
          hasNext: true,
          nextUrl: "/search?q=react&page=2",
        },
      }),
    );

    expect(screen.getByRole("link", { name: "Search result" })).toBeTruthy();
    expect(screen.getByText("react · 1 个结果")).toBeTruthy();
    const input = screen.getByRole("textbox", { name: "搜索主题..." });
    await user.clear(input);
    await user.type(input, "goose forum");
    await user.click(screen.getByRole("button", { name: "搜索" }));
    expect(navigate).toHaveBeenCalledWith("/search?q=goose+forum");
  });

  it("renders the Links information hierarchy and safe external links", () => {
    renderPage(
      payload("links.index", {
        totalCount: 1,
        groups: [
          {
            name: "COMMUNITY",
            emoji: "👥",
            color: "#64748b",
            links: [
              {
                name: "Example",
                desc: "Community site",
                url: "https://example.com",
                logoUrl: "",
              },
            ],
          },
        ],
      }),
    );

    expect(
      screen.getByRole("heading", { level: 1, name: "友情链接" }),
    ).toBeTruthy();
    expect(
      screen.getByRole("heading", { level: 2, name: /COMMUNITY/ }),
    ).toBeTruthy();
    const external = screen.getByRole("link", { name: /Example/ });
    expect(external.getAttribute("target")).toBe("_blank");
    expect(external.getAttribute("rel")).toBe("noopener noreferrer");
    expect(
      screen.getByRole("link", { name: "去发帖申请" }).getAttribute("href"),
    ).toBe("/publish");
    expect(
      screen
        .getAllByRole("link", { name: "友情链接" })
        .some((link) => link.getAttribute("aria-current") === "page"),
    ).toBe(true);
  });

  it("supports mobile navigation and the shell theme action", async () => {
    const { toggleTheme, user } = renderPage(
      payload("links.index", { totalCount: 0, groups: [] }),
    );

    await user.click(screen.getByRole("button", { name: "打开菜单" }));
    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByRole("heading", { name: "菜单" })).toBeTruthy();
    expect(within(dialog).getByRole("link", { name: "Coding" })).toBeTruthy();

    await user.click(within(dialog).getByRole("button", { name: "关闭菜单" }));
    await user.click(screen.getByRole("button", { name: "切换到深色主题" }));
    expect(toggleTheme).toHaveBeenCalledOnce();
  });

  it("renders sponsor tiers, defaults, contact, and rules", () => {
    renderPage(
      payload("sponsors.index", {
        totalCount: 1,
        content: {
          title: "感谢支持",
          description: "支持 GooseForum 的朋友们。",
        },
        contact: {
          title: "联系我们",
          description: "欢迎支持。",
          buttonText: "发送邮件",
          buttonLink: "mailto:test@example.com",
        },
        rules: [{ content: "内容公开透明。" }],
        sections: [
          {
            key: "gold",
            label: "Gold",
            tone: "gold",
            sponsors: [
              {
                name: "Alice",
                message: "",
                link: "https://example.com/alice",
                avatarUrl: "/avatar.webp",
              },
            ],
          },
        ],
      }),
    );

    expect(
      screen.getByRole("heading", { level: 1, name: "感谢支持" }),
    ).toBeTruthy();
    expect(screen.getByText("感谢支持 GooseForum。")).toBeTruthy();
    const sponsor = screen.getByRole("link", { name: /Alice/ });
    expect(sponsor.getAttribute("rel")).toBe("noopener noreferrer");
    expect(
      screen.getByRole("link", { name: "发送邮件" }).getAttribute("href"),
    ).toBe("mailto:test@example.com");
    expect(screen.getByText("内容公开透明。")).toBeTruthy();
  });

  it("uses official empty states when payload collections are empty", () => {
    renderPage(payload("links.index", { totalCount: 0, groups: [] }));
    expect(screen.getByText("暂无链接")).toBeTruthy();
    expect(screen.getByText("站点还没有配置友情链接。")).toBeTruthy();
  });

  it("preserves the authenticated shell actions and permission-gated admin links", async () => {
    const page = payload("links.index", { totalCount: 0, groups: [] });
    page.layout = {
      ...page.layout,
      viewer: {
        ...page.layout.viewer,
        id: 7,
        username: "alice",
        avatarUrl: "/alice.webp",
        isAuthenticated: true,
        canAccessAdmin: true,
      },
      unread: { notifications: true, messages: true, moderationReports: false },
    };
    const { user } = renderPage(page);

    await user.hover(screen.getByRole("button", { name: "切换语言" }));
    expect(
      await screen.findByRole("menuitemradio", { name: "English" }),
    ).toBeTruthy();

    await user.hover(screen.getByRole("button", { name: "alice" }));

    expect(await screen.findByRole("menuitem", { name: /发布/ })).toBeTruthy();
    expect(screen.getByRole("menuitem", { name: /设置/ })).toBeTruthy();
    expect(screen.getByRole("menuitem", { name: /访问组/ })).toBeTruthy();
    expect(screen.getByRole("menuitem", { name: /主题预览/ })).toBeTruthy();
    expect(screen.getByRole("menuitem", { name: /管理后台/ })).toBeTruthy();
  });

  it("renders category identity, fallback copy, and compact topic counts", () => {
    renderPage(
      payload("categories.index", {
        total: 2,
        categories: [
          {
            id: 1,
            name: "Coding",
            description: "Development topics",
            icon: "💻",
            color: "#8241d6",
            url: "/c/Coding/1",
            topicCount: 1_250,
          },
          {
            id: 2,
            name: "General",
            description: "",
            icon: "/general.webp",
            color: "#22c55e",
            url: "/c/General/2",
            topicCount: 3,
          },
        ],
      }),
    );

    expect(
      screen.getByRole("heading", { level: 1, name: "全部分类" }),
    ).toBeTruthy();
    expect(screen.getByText("2 个分类")).toBeTruthy();
    expect(screen.getByText("1.3k 个主题")).toBeTruthy();
    expect(screen.getByText("这个分类还没有介绍。")).toBeTruthy();
    expect(document.querySelector('a[href="/c/Coding/1"]')).toBeTruthy();
    expect(document.querySelector('img[src="/general.webp"]')).toBeTruthy();
  });

  it("renders member identity, stats, fallbacks, and pagination semantics", () => {
    renderPage(
      payload("members.index", {
        members: [
          {
            id: 7,
            username: "alice",
            nickname: "Alice",
            avatarUrl: "/alice.webp",
            bio: "",
            prestige: 1_250,
            topicCount: 12,
            replyCount: 34,
            joinedAt: "2026-01-02",
            url: "/u/7",
          },
        ],
        previousUrl: "/members?page=1",
        pagination: {
          page: 2,
          nextPage: 3,
          hasNext: true,
          nextUrl: "/members?page=3",
        },
      }),
    );

    expect(
      screen.getByRole("heading", { level: 1, name: "所有成员" }),
    ).toBeTruthy();
    expect(screen.getByText("@alice")).toBeTruthy();
    expect(screen.getByText("这位成员还没有填写个人简介。")).toBeTruthy();
    expect(screen.getByText("1.3k")).toBeTruthy();
    expect(
      screen.getByRole("link", { name: /Alice/ }).getAttribute("href"),
    ).toBe("/u/7");
    expect(
      screen.getByRole("link", { name: "上一页" }).getAttribute("rel"),
    ).toBe("prev");
    expect(
      screen.getByRole("link", { name: "下一页" }).getAttribute("rel"),
    ).toBe("next");
  });
});
