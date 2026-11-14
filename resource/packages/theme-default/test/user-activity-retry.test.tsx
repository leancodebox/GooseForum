import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { AnyPagePayload, GooseSiteApi, UserProfileProps } from "@gooseforum/client";
import {
  GooseRuntimeProvider,
  type GooseRuntime,
} from "@gooseforum/runtime";
import { GooseI18nProvider } from "@gooseforum/runtime/i18n";
import { UserActivity } from "../src/site/users/user-activity";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("user activity waterfall", () => {
  it("merges a successful page and stops when the server has no next page", async () => {
    vi.stubGlobal("IntersectionObserver", class {
      constructor(private readonly callback: IntersectionObserverCallback) {}
      observe() { this.callback([{ isIntersecting: true } as IntersectionObserverEntry], this as unknown as IntersectionObserver); }
      disconnect() {}
    });
    const timeline = activityPage();
    const fetchPage = vi.fn().mockResolvedValue({
      component: "user.profile",
      props: {
        ...timeline,
        activities: [{ ...timeline.activities[0], id: 2, contentPreview: "Second activity" }],
        pagination: { page: 2, nextPage: 0, hasNext: false, nextUrl: "" },
      },
      layout: {} as AnyPagePayload["layout"],
      meta: { title: "Test" },
      url: "/u/7?section=activity&page=2",
      version: "1",
    } satisfies AnyPagePayload);

    render(
      <GooseI18nProvider locale="zh">
        <GooseRuntimeProvider runtime={createRuntime(fetchPage)}>
          <UserActivity page={timeline} />
        </GooseRuntimeProvider>
      </GooseI18nProvider>,
    );

    await waitFor(() => expect(fetchPage).toHaveBeenCalledWith("/u/7?section=activity&page=2"));
    expect(await screen.findByText(/Second activity/)).toBeTruthy();
    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(fetchPage).toHaveBeenCalledTimes(1);
  });

  it("does not automatically retry a failed load", async () => {
    let notify: IntersectionObserverCallback | undefined;
    vi.stubGlobal(
      "IntersectionObserver",
      class {
        constructor(callback: IntersectionObserverCallback) {
          notify = callback;
        }
        observe() {
          notify?.(
            [{ isIntersecting: true } as IntersectionObserverEntry],
            this as unknown as IntersectionObserver,
          );
        }
        disconnect() {}
      },
    );
    const fetchPage = vi.fn().mockRejectedValue(new Error("temporary activity failure"));

    render(
      <GooseI18nProvider locale="zh">
        <GooseRuntimeProvider runtime={createRuntime(fetchPage)}>
          <UserActivity page={activityPage()} />
        </GooseRuntimeProvider>
      </GooseI18nProvider>,
    );

    await waitFor(() => expect(fetchPage).toHaveBeenCalledTimes(1));
    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(fetchPage).toHaveBeenCalledTimes(1);
  });

  it("drops an in-flight response after switching activity tabs", async () => {
    vi.stubGlobal("IntersectionObserver", class {
      constructor(private readonly callback: IntersectionObserverCallback) {}
      observe() { this.callback([{ isIntersecting: true } as IntersectionObserverEntry], this as unknown as IntersectionObserver); }
      disconnect() {}
    });
    let resolvePage!: (value: AnyPagePayload) => void;
    const fetchPage = vi.fn(() => new Promise<AnyPagePayload>((resolve) => { resolvePage = resolve; }));
    const timeline = activityPage();
    const likes = { ...timeline, activityTab: "likes" as const, likes: [{ id: 2, topicId: 2, title: "New like", url: "/t/new", likedAt: "2026-01-04T08:00:00Z" }] };
    const view = (page: UserProfileProps) => (
      <GooseI18nProvider locale="zh">
        <GooseRuntimeProvider runtime={createRuntime(fetchPage)}>
          <UserActivity page={page} />
        </GooseRuntimeProvider>
      </GooseI18nProvider>
    );
    const rendered = render(view(timeline));
    rendered.rerender(view(likes));
    resolvePage({ component: "user.profile", props: { ...timeline, activities: [{ ...timeline.activities[0], id: 99 }] }, layout: {} as AnyPagePayload["layout"], meta: { title: "Test" }, url: "/u/7?section=activity", version: "1" });
    await waitFor(() => expect(screen.getByText("New like")).toBeTruthy());
    expect(screen.queryByText("First activity")).toBeNull();
  });

  it("drops an in-flight failure after switching activity tabs", async () => {
    vi.stubGlobal("IntersectionObserver", class {
      constructor(private readonly callback: IntersectionObserverCallback) {}
      observe() { this.callback([{ isIntersecting: true } as IntersectionObserverEntry], this as unknown as IntersectionObserver); }
      disconnect() {}
    });
    let rejectPage!: (reason?: unknown) => void;
    const fetchPage = vi.fn(() => new Promise<AnyPagePayload>((_resolve, reject) => { rejectPage = reject; }));
    const timeline = activityPage();
    const likes = { ...timeline, activityTab: "likes" as const, likes: [{ id: 2, topicId: 2, title: "New like", url: "/t/new", likedAt: "2026-01-04T08:00:00Z" }] };
    const view = (page: UserProfileProps) => (
      <GooseI18nProvider locale="zh">
        <GooseRuntimeProvider runtime={createRuntime(fetchPage)}>
          <UserActivity page={page} />
        </GooseRuntimeProvider>
      </GooseI18nProvider>
    );
    const rendered = render(view(timeline));
    rendered.rerender(view(likes));
    rejectPage(new Error("stale activity failure"));
    await waitFor(() => expect(screen.getByText("New like")).toBeTruthy());
    expect(screen.queryByText(/失败|failed/i)).toBeNull();
  });
});

function createRuntime(
  fetchPage: NonNullable<GooseRuntime["fetchPage"]>,
): GooseRuntime {
  return {
    api: {} as GooseSiteApi,
    currentUrl: "/u/7?section=activity",
    isNavigating: false,
    theme: "gf-light",
    locale: "zh",
    navigate: vi.fn(async () => {}),
    fetchPage,
    queueFlash: vi.fn(),
    redirect: vi.fn(),
    refresh: vi.fn(async () => {}),
    setLocale: vi.fn(),
    toggleTheme: vi.fn(),
  };
}

function activityPage(): UserProfileProps {
  return {
    user: {
      userId: 7,
      username: "alice",
      nickname: "Alice",
      avatarUrl: "",
      profileCoverUrl: "",
      bio: "",
      signature: "",
      websiteName: "",
      website: "",
      prestige: 0,
      externalInformation: {},
      isAdmin: false,
      topicCount: 0,
      replyCount: 0,
      likeReceivedCount: 0,
      likeGivenCount: 0,
      followerCount: 0,
      followingCount: 0,
      collectionCount: 0,
      isOnline: false,
      isFollowing: false,
      isSelf: false,
      badges: [],
      wornBadge: null,
      lastActiveTime: "",
      createdAt: "2025-01-02T08:00:00Z",
    },
    section: "activity",
    activityTab: "timeline",
    tabs: [],
    activityTabs: [
      { key: "timeline", url: "/u/7?section=activity", active: true },
    ],
    pagination: { page: 1, nextPage: 2, hasNext: true, nextUrl: "/u/7?section=activity&page=2" },
    badges: [],
    topics: [],
    activities: [
      {
        id: 1,
        action: 2,
        subjectType: "topic",
        subjectId: 1,
        contentPreview: "First activity",
        url: "/p/first/1",
        label: "post",
        createdAt: "2026-01-03T08:00:00Z",
      },
    ],
    likes: [],
    following: [],
    followers: [],
    isOwnProfile: false,
    canMessage: false,
    canFollow: false,
    messageUrl: "",
    settingsUrl: "",
  };
}
