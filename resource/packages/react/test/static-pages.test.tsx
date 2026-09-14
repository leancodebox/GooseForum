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
  MessagesPageProps,
  DraftsPageProps,
  ModerationPageProps,
  PublishPageProps,
  TopicDetailProps,
  ThemePreviewProps,
  UserProfileProps,
} from "@gooseforum/client";
import { createEmptySiteThemeTokens } from "@gooseforum/client";
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
    | "notifications.index"
    | "messages.index"
    | "drafts.index"
    | "access-groups.index"
    | "moderation.index"
    | "error.index"
    | "publish.index"
    | "topic.detail"
    | "theme.preview",
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

function messagesProps(): MessagesPageProps {
  return {
    conversations: [
      {
        id: 4,
        peerId: 9,
        peerUsername: "bob",
        peerAvatar: "/bob.webp",
        lastMsg: "Previous message",
        lastMsgTime: "2026-09-14T08:00:00Z",
        unreadCount: 2,
        convId: 4,
        peerUrl: "/u/9",
      },
    ],
    suggestedUsers: [
      {
        id: 10,
        username: "carol",
        nickname: "Carol",
        avatarUrl: "/carol.webp",
        bio: "",
        url: "/u/10",
      },
    ],
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

    const heading = await screen.findByRole("heading", {
      level: 1,
      name: "通知",
    });
    expect(heading.closest("header")?.classList.contains("sm:border-b-0")).toBe(
      true,
    );
    const allTab = screen.getByRole("tab", { name: "全部" });
    expect(allTab.classList.contains("h-8")).toBe(true);
    const tabsList = allTab.closest('[data-slot="tabs-list"]');
    expect(tabsList?.classList.contains("group-data-horizontal/tabs:h-auto")).toBe(
      true,
    );
    expect(tabsList?.classList.contains("group-data-horizontal/tabs:h-8")).toBe(
      false,
    );
    expect(
      allTab.closest("section")?.classList.contains("max-sm:border-t-0"),
    ).toBe(true);
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

  it("loads a selected conversation, marks it read, and sends on Enter", async () => {
    const messages = vi
      .fn()
      .mockResolvedValueOnce({
        list: [
          {
            id: 31,
            senderId: 9,
            content: "Hello back",
            msgType: 1,
            isRead: 1,
            createdAt: "2026-09-14T08:00:00Z",
            isSelf: false,
          },
        ],
        hasMoreBefore: true,
        hasMoreAfter: false,
        nextBeforeId: 31,
        latestId: 31,
      })
      .mockResolvedValueOnce({
        list: [
          {
            id: 30,
            senderId: 7,
            content: "Older message",
            msgType: 1,
            isRead: 1,
            createdAt: "2026-09-13T08:00:00Z",
            isSelf: true,
          },
        ],
        hasMoreBefore: false,
        hasMoreAfter: true,
        nextBeforeId: 0,
        latestId: 31,
      });
    const markRead = vi.fn().mockResolvedValue(true);
    const send = vi.fn().mockResolvedValue({ convId: 4 });
    const page = payload("messages.index", messagesProps());
    page.url = "/messages?userId=9";
    page.layout = {
      ...page.layout,
      viewer: {
        ...page.layout.viewer,
        id: 7,
        username: "alice",
        isAuthenticated: true,
      },
    };
    const { user } = renderPage(page, {
      chat: { messages, markRead, send } as unknown as GooseSiteApi["chat"],
    });

    expect(await screen.findByText("Hello back")).toBeTruthy();
    expect(messages).toHaveBeenCalledWith({ convId: 4, limit: 30 });
    expect(markRead).toHaveBeenCalledWith(4);
    await user.click(screen.getByRole("button", { name: "加载更早消息" }));
    expect(await screen.findByText("Older message")).toBeTruthy();
    expect(messages).toHaveBeenLastCalledWith({
      convId: 4,
      beforeId: 31,
      limit: 30,
    });
    const composer = screen.getByPlaceholderText("输入消息…");
    await user.type(composer, "Hello Bob{enter}");
    expect(send).toHaveBeenCalledWith(9, "Hello Bob");
    expect(
      (await screen.findAllByText("Hello Bob")).length,
    ).toBeGreaterThanOrEqual(2);
  });

  it("starts a new conversation from the shadcn user picker", async () => {
    const page = payload("messages.index", messagesProps());
    page.layout = {
      ...page.layout,
      viewer: {
        ...page.layout.viewer,
        id: 7,
        username: "alice",
        isAuthenticated: true,
      },
    };
    const { user } = renderPage(page, {
      chat: {} as GooseSiteApi["chat"],
    });

    await user.click(
      (await screen.findAllByRole("button", { name: "新私信" }))[0],
    );
    const dialog = await screen.findByRole("dialog");
    await user.type(within(dialog).getByPlaceholderText("搜索用户…"), "carol");
    await user.click(within(dialog).getByRole("button", { name: /Carol/ }));
    expect(await screen.findByText("给 Carol 发第一条消息。")).toBeTruthy();
  });

  it("renders drafts and preserves edit navigation", async () => {
    const props: DraftsPageProps = {
      total: 1,
      drafts: [
        {
          id: 41,
          title: "Unfinished migration",
          description: "Continue later",
          editUrl: "/publish?id=41",
          replyCount: 2,
          viewCount: 9,
          processStatus: 1,
          updatedAt: "2026-09-14T08:00:00Z",
          createdAt: "2026-09-13T08:00:00Z",
          categories: [
            { id: 4, name: "Coding", url: "/c/Coding/4", color: "#8241d6" },
          ],
        },
      ],
      pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" },
    };
    renderPage(payload("drafts.index", props));
    expect(
      await screen.findByRole("heading", { level: 1, name: "草稿箱" }),
    ).toBeTruthy();
    expect(screen.getByText("已封禁")).toBeTruthy();
    expect(
      screen
        .getAllByRole("link", { name: /Unfinished migration|继续编辑/ })
        .some((link) => link.getAttribute("href") === "/publish?id=41"),
    ).toBe(true);
  });

  it("loads access groups in parallel and submits an application", async () => {
    const list = vi
      .fn()
      .mockResolvedValue([
        { id: 5, name: "Core", categories: ["Private"], status: 0 },
      ]);
    const managed = vi.fn().mockResolvedValue([]);
    const apply = vi.fn().mockResolvedValue(true);
    const { user } = renderPage(payload("access-groups.index", {}), {
      accessGroups: {
        list,
        managed,
        apply,
      } as unknown as GooseSiteApi["accessGroups"],
    });
    expect(await screen.findByText("Core")).toBeTruthy();
    expect(list).toHaveBeenCalledOnce();
    expect(managed).toHaveBeenCalledOnce();
    await user.click(screen.getByRole("button", { name: "申请加入" }));
    expect(apply).toHaveBeenCalledWith(5);
    expect(await screen.findByText("等待审核")).toBeTruthy();
  });

  it("loads and handles moderation reports", async () => {
    const props: ModerationPageProps = {
      categoryTabs: [
        {
          key: "all",
          label: "全部",
          url: "/moderation?category=all",
          active: false,
        },
        {
          key: "chat",
          label: "吐槽/脑洞",
          url: "/moderation?category=chat",
          active: true,
        },
      ],
      topics: [],
      pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" },
    };
    const reports = vi.fn().mockResolvedValue({
      items: [
        {
          id: 8,
          targetType: "topic",
          targetId: 21,
          targetUrl: "/p/spam/21",
          title: "Spam topic",
          excerpt: "Bad content",
          reason: "spam",
          note: "",
          status: "open",
          resolution: "",
          reporter: { id: 2, username: "reporter", avatarUrl: "" },
          handler: { id: 0, username: "", avatarUrl: "" },
          categories: [],
          createdAt: "2026-09-14T08:00:00Z",
        },
      ],
      nextCursor: 0,
      hasNext: false,
    });
    const setTopicStatus = vi.fn().mockResolvedValue(true);
    const setReportStatus = vi.fn().mockResolvedValue(true);
    const { user } = renderPage(payload("moderation.index", props), {
      moderation: {
        reports,
        setTopicStatus,
        setReportStatus,
      } as unknown as GooseSiteApi["moderation"],
    });
    expect(
      await screen.findByRole("link", { name: "Spam topic" }),
    ).toBeTruthy();
    const tabsLists = document.querySelectorAll('[data-slot="tabs-list"]');
    expect(tabsLists).toHaveLength(2);
    expect(
      tabsLists[0]?.classList.contains("group-data-horizontal/tabs:h-auto"),
    ).toBe(true);
    expect(tabsLists[0]?.classList.contains("rounded-lg")).toBe(true);
    expect(tabsLists[0]?.classList.contains("bg-transparent")).toBe(true);
    expect(
      document
        .querySelector('[data-slot="moderation-tabs-frame"]')
        ?.classList.contains("bg-muted"),
    ).toBe(true);
    expect(
      tabsLists[1]?.classList.contains("group-data-horizontal/tabs:h-auto"),
    ).toBe(true);
    expect(
      tabsLists[1]?.classList.contains("group-data-horizontal/tabs:h-8"),
    ).toBe(false);
    expect(
      screen.getByRole("tab", { name: "待处理" }).classList.contains("flex-none"),
    ).toBe(true);
    expect(tabsLists[0]?.classList.contains("w-fit")).toBe(true);
    expect(tabsLists[1]?.classList.contains("w-full")).toBe(true);
    await user.click(screen.getByRole("button", { name: "封禁" }));
    expect(setTopicStatus).toHaveBeenCalledWith(21, "ban");
    expect(setReportStatus).toHaveBeenCalledWith(8, "ban");
    await waitFor(() => expect(screen.queryByText("Spam topic")).toBeNull());
    await user.click(screen.getByRole("tab", { name: "封禁记录" }));
    const activeCategory = screen.getByRole("link", { name: "吐槽/脑洞" });
    expect(activeCategory.getAttribute("data-variant")).toBe("default");
    expect(activeCategory.getAttribute("aria-current")).toBe("page");
  });

  it("localizes server-backed error pages", async () => {
    renderPage(
      payload("error.index", {
        code: "404",
        title: "Not found",
        messageCode: "page.notFound",
      }),
    );
    expect(await screen.findByText("404 · 页面不存在")).toBeTruthy();
    expect(screen.getByText("页面不存在，或已经被删除。")).toBeTruthy();
  });

  it("publishes a topic from the React visual/Markdown composer", async () => {
    const props: PublishPageProps = {
      topicId: 0,
      isEditing: false,
      categories: [
        {
          id: 4,
          name: "Coding",
          color: "#8241d6",
          isRestricted: false,
          canCreate: true,
        },
      ],
      topic: { title: "", content: "", categoryIds: [], topicStatus: 0 },
    };
    const writeReviewed = vi
      .fn()
      .mockResolvedValue({
        id: 51,
        moderationStatus: "approved",
        topicStatus: 1,
      });
    const { navigate, user } = renderPage(payload("publish.index", props), {
      topics: { writeReviewed } as unknown as GooseSiteApi["topics"],
      uploads: {} as GooseSiteApi["uploads"],
    });
    expect(
      await screen.findByRole("heading", { level: 1, name: "发布主题" }),
    ).toBeTruthy();
    await user.type(
      screen.getByRole("textbox", { name: "标题" }),
      "New React topic",
    );
    await user.click(screen.getByRole("button", { name: /Coding/ }));
    await user.click(screen.getByRole("radio", { name: "Markdown" }));
    await user.type(screen.getByPlaceholderText(/输入正文/), "Topic body");
    await user.click(screen.getByRole("button", { name: "发布主题" }));
    await waitFor(() =>
      expect(writeReviewed).toHaveBeenCalledWith(
        expect.objectContaining({
          title: "New React topic",
          content: "Topic body",
          categoryId: [4],
          topicStatus: 1,
        }),
      ),
    );
    expect(navigate).toHaveBeenCalledWith("/p/post/51");
  });

  it("supports topic actions, reply windows, and posting replies", async () => {
    const post = {
      id: 61,
      topicId: 60,
      postNo: 1,
      content: "Original body",
      renderedContent: "<p>Original body</p>",
      processStatus: 0,
      isHidden: false,
      canModerate: false,
      author: { id: 7, username: "alice", avatarUrl: "" },
      createdAt: "2026-09-14T08:00:00Z",
      isOwnPost: true,
    };
    const props: TopicDetailProps = {
      topic: {
        id: 60,
        title: "Topic detail",
        description: "Description",
        url: "/p/post/60",
        topicStatus: 1,
        processStatus: 0,
        author: post.author,
        participants: [post.author],
        categories: [],
        replyCount: 0,
        maxPostNo: 2,
        viewCount: 10,
        likeCount: 0,
        isLiked: false,
        isBookmarked: false,
        isWatched: false,
        createdAt: post.createdAt,
        updatedAt: post.createdAt,
      },
      postStream: {
        posts: [post],
        replyTargets: [],
        beforePostNo: 1,
        afterPostNo: 1,
        hasBefore: false,
        hasAfter: true,
        total: 2,
        maxPostNo: 2,
      },
      hotTopics: [],
      permissions: { isOwnTopic: true, canPost: true, canModerateTopic: false },
    };
    const like = vi.fn().mockResolvedValue(true);
    const windowRequest = vi
      .fn()
      .mockResolvedValue({
        posts: [
          {
            ...post,
            id: 62,
            postNo: 2,
            content: "Second",
            renderedContent: "<p>Second</p>",
            isOwnPost: false,
          },
        ],
        replyTargets: [],
        beforePostNo: 2,
        afterPostNo: 2,
        hasBefore: true,
        hasAfter: false,
        total: 2,
        maxPostNo: 2,
      });
    const create = vi
      .fn()
      .mockResolvedValue({
        id: 63,
        postNo: 3,
        renderedContent: "<p>Reply body</p>",
        processStatus: 0,
      });
    const topicPage = payload("topic.detail", props);
    topicPage.layout = {
      ...topicPage.layout,
      viewer: {
        ...topicPage.layout.viewer,
        id: 7,
        username: "alice",
        isAuthenticated: true,
      },
    };
    const { user } = renderPage(topicPage, {
      topics: { like } as unknown as GooseSiteApi["topics"],
      posts: {
        window: windowRequest,
        create,
      } as unknown as GooseSiteApi["posts"],
      users: { card: vi.fn() } as unknown as GooseSiteApi["users"],
    });
    expect(
      await screen.findByRole("heading", { level: 1, name: "Topic detail" }),
    ).toBeTruthy();
    await user.click(screen.getByRole("button", { name: "点赞" }));
    expect(like).toHaveBeenCalledWith(60, 1);
    await user.click(screen.getByRole("button", { name: "加载更多回复" }));
    expect(await screen.findByText("Second")).toBeTruthy();
    await user.click(screen.getByRole("button", { name: "参与讨论" }));
    await user.click(screen.getByRole("radio", { name: "Markdown" }));
    await user.type(screen.getByPlaceholderText(/输入正文/), "Reply body");
    await user.click(screen.getByRole("button", { name: "发布回复" }));
    expect(create).toHaveBeenCalledWith({
      topicId: 60,
      content: "Reply body",
      replyToPostId: 0,
    });
    expect(await screen.findByText("Reply body")).toBeTruthy();
  });

  it("edits and saves the complete site theme draft", async () => {
    const tokens = createEmptySiteThemeTokens();
    Object.assign(tokens, {
      "color-base-100": "#ffffff",
      "color-base-200": "#f8fafc",
      "color-base-300": "#f1f5f9",
      "color-base-content": "#111827",
      "color-icon-muted": "#64748b",
      "color-line": "#e2e8f0",
      "color-primary": "#315ef4",
      "color-primary-content": "#ffffff",
      "color-secondary": "#f0f4f8",
      "color-secondary-content": "#1f2937",
      "color-accent": "#10b981",
      "color-accent-content": "#052e2b",
      "color-neutral": "#1f2937",
      "color-neutral-content": "#ffffff",
      "color-info": "#2563eb",
      "color-info-content": "#eff6ff",
      "color-success": "#16a34a",
      "color-success-content": "#f0fdf4",
      "color-warning": "#f59e0b",
      "color-warning-content": "#fffbeb",
      "color-error": "#dc2626",
      "color-error-content": "#fef2f2",
      "radius-selector": "0.5rem",
      "radius-field": "0.5rem",
      "radius-box": "0.5rem",
      "size-selector": "0.25rem",
      "size-field": "0.25rem",
      border: "1px",
      depth: "1",
    });
    const props: ThemePreviewProps = {
      theme: {
        version: 1,
        enabled: true,
        themes: [
          { name: "gf-light", label: "Light", colorScheme: "light", tokens },
          {
            name: "gf-dark",
            label: "Dark",
            colorScheme: "dark",
            tokens: {
              ...tokens,
              "color-base-100": "#111111",
              "color-base-content": "#eeeeee",
            },
          },
        ],
      },
      defaults: {
        version: 1,
        enabled: true,
        themes: [
          {
            name: "gf-light",
            label: "Light",
            colorScheme: "light",
            tokens: { ...tokens },
          },
          {
            name: "gf-dark",
            label: "Dark",
            colorScheme: "dark",
            tokens: {
              ...tokens,
              "color-base-100": "#111111",
              "color-base-content": "#eeeeee",
            },
          },
        ],
      },
    };
    const save = vi.fn().mockImplementation(async (value) => value);
    const themePage = payload("theme.preview", props);
    themePage.layout = {
      ...themePage.layout,
      viewer: { ...themePage.layout.viewer, adminPermissions: [5] },
    };
    const { user } = renderPage(themePage, {
      themes: { save } as unknown as GooseSiteApi["themes"],
    });
    expect(
      await screen.findByRole("heading", { level: 1, name: "主题预览设置" }),
    ).toBeTruthy();
    await user.click(screen.getByRole("button", { name: /Warm/ }));
    expect(screen.getByText("未保存")).toBeTruthy();
    await user.click(screen.getByRole("button", { name: "保存草稿" }));
    await waitFor(() =>
      expect(save).toHaveBeenCalledWith(
        expect.objectContaining({ enabled: true }),
      ),
    );
    expect(
      await screen.findByText("主题草稿已保存，不会影响全站。"),
    ).toBeTruthy();
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
