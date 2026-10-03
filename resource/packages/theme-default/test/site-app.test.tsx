import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type {
  AnyPagePayload,
  GooseSiteApi,
  LayoutPayload,
} from "@gooseforum/client";
import type { PageSource } from "@gooseforum/runtime/page-source";
import { SiteApp } from "../src/app/site-app";

afterEach(cleanup);
beforeEach(() => {
  window.scrollTo = vi.fn();
});

const layout: LayoutPayload = {
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
    id: 7,
    username: "alice",
    email: "",
    avatarUrl: "",
    isAuthenticated: true,
    canAccessAdmin: false,
    isModerator: false,
    requiresEmailVerification: false,
    adminPermissions: [],
  },
  header: [],
  sidebar: { activeKey: "", categories: [] },
  footer: { links: [], primary: [] },
  unread: { notifications: false, messages: false },
  theme: { enabled: false, current: "gf-light", themeColor: "#fff" },
};

function page(
  component: AnyPagePayload["component"],
  props: unknown,
  url: string,
): AnyPagePayload {
  return {
    component,
    props,
    layout,
    meta: { title: "Test" },
    url,
    version: "1",
  } as AnyPagePayload;
}

describe("SiteApp system theme", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.removeItem("goose-site-theme");
    document.cookie = "goose-site-theme=; path=/; max-age=0";
  });

  it("follows system changes, supports manual overrides, and removes the listener", async () => {
    localStorage.clear();
    document.cookie = "goose-site-theme=; path=/; max-age=0";
    const media = new EventTarget() as EventTarget & { matches: boolean };
    media.matches = true;
    const removeListener = vi.spyOn(media, "removeEventListener");
    vi.stubGlobal("matchMedia", vi.fn(() => media));
    const initial = page("categories.index", { categories: [], total: 0 }, "/categories");
    const source: PageSource<AnyPagePayload> = { api: {} as GooseSiteApi, load: vi.fn() };
    const user = userEvent.setup();
    const mounted = render(<SiteApp pageSource={source} initialPage={initial} />);
    await screen.findByRole("heading", { level: 1 });
    expect(document.documentElement.dataset.theme).toBe("gf-dark");
    expect(localStorage.getItem("goose-site-theme")).toBeNull();
    act(() => { media.matches = false; media.dispatchEvent(new Event("change")); });
    expect(document.documentElement.dataset.theme).toBe("gf-light");
    await user.hover(screen.getByRole("button", { name: /选择主题|Choose theme/ }));
    await user.click(screen.getByRole("menuitemradio", { name: /深色|^Dark$/ }));
    expect(localStorage.getItem("goose-site-theme")).toBe("gf-dark");
    act(() => { media.dispatchEvent(new Event("change")); });
    expect(document.documentElement.dataset.theme).toBe("gf-dark");
    await user.hover(screen.getByRole("button", { name: /选择主题|Choose theme/ }));
    await user.click(screen.getByRole("menuitemradio", { name: /跟随系统|^System$/ }));
    expect(document.documentElement.dataset.theme).toBe("gf-light");
    expect(localStorage.getItem("goose-site-theme")).toBe("system");
    mounted.unmount();
    expect(removeListener).toHaveBeenCalledWith("change", expect.any(Function));
  });
});

describe("SiteApp navigation lifecycle", () => {
  it("keeps cached page DOM when visiting and leaving a standalone page", async () => {
    window.history.replaceState(null, "", "/");
    const home = { ...page("home.index", {
      sort: "latest", tabs: [], topics: [],
      pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" },
      announcement: { enabled: false, html: "" },
    }, "/"), layout: { ...layout, viewer: { ...layout.viewer, isAuthenticated: false } } } as AnyPagePayload;
    const login = { ...page("auth.login", { initialMode: "login", redirectUrl: "/", oauthProviders: [] }, "/login"), layout: home.layout } as AnyPagePayload;
    const source: PageSource<AnyPagePayload> = {
      api: { auth: { captcha: vi.fn().mockResolvedValue({ captchaId: "id", captchaImg: "" }) } } as unknown as GooseSiteApi,
      load: vi.fn(async url => url.pathname === "/login" ? login : home),
    };
    const user = userEvent.setup();
    render(<SiteApp pageSource={source} initialPage={home} />);
    const original = await screen.findByText(/暂无主题|No topics/);
    await user.click(document.querySelector('a[href="/login"]')!);
    await screen.findByRole("textbox", { name: /用户名|Username/ });
    expect(original.isConnected).toBe(true);
    await user.click(screen.getByRole("link", { name: "GooseForum" }));
    expect(await screen.findByText(/暂无主题|No topics/)).toBe(original);
  });
  it("renders an embedded initial page without fetching it again", async () => {
    const initialPage = page("categories.index", { categories: [], total: 0 }, "/categories");
    const source: PageSource<AnyPagePayload> = {
      api: {} as GooseSiteApi,
      load: vi.fn(),
    };
    render(<SiteApp pageSource={source} initialPage={initialPage} />);
    expect(await screen.findByRole("heading", { level: 1 })).toBeTruthy();
    expect(source.load).not.toHaveBeenCalled();
  });
  it("keeps the current content and URL until the next payload is ready", async () => {
    window.history.replaceState(null, "", "/");
    const home = page("home.index", {
      sort: "latest",
      tabs: [],
      topics: [],
      pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" },
      announcement: { enabled: false, html: "" },
    }, "/");
    const categories = page(
      "categories.index",
      { categories: [], total: 0 },
      "/categories",
    );
    let resolvePage!: (value: AnyPagePayload) => void;
    const pendingPage = new Promise<AnyPagePayload>((resolve) => {
      resolvePage = resolve;
    });
    const source: PageSource<AnyPagePayload> = {
      api: {} as GooseSiteApi,
      load: vi.fn(() => pendingPage),
    };
    const user = userEvent.setup();
    render(<SiteApp pageSource={source} initialPage={home} />);

    const currentContent = await screen.findByText(/暂无主题|No topics/);
    const sidebar = screen.getByRole("complementary", { name: "Sidebar" });
    await user.click(
      sidebar.querySelector<HTMLAnchorElement>('a[href="/categories"]')!,
    );

    expect(currentContent.isConnected).toBe(true);
    expect(window.location.pathname).toBe("/");
    expect(source.load).toHaveBeenCalledOnce();
    expect(document.querySelector(".goose-navigation-progress")).toBeNull();
    expect(window.scrollTo).not.toHaveBeenCalled();

    vi.mocked(window.scrollTo).mockImplementation(() => {
      expect(screen.queryByText(/暂无分类|No categories/)).not.toBeNull();
    });
    resolvePage(categories);
    expect(await screen.findByText(/暂无分类|No categories/)).toBeTruthy();
    expect(window.location.pathname).toBe("/categories");
    expect(window.scrollTo).toHaveBeenCalledWith({
      left: 0,
      top: 0,
      behavior: "instant",
    });
  });
  it("keeps one shell and updates navigation together with page content", async () => {
    window.history.replaceState(null, "", "/");
    const home = {
      ...page(
        "home.index",
        {
          sort: "latest",
          tabs: [],
          topics: [],
          pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" },
          announcement: { enabled: false, html: "" },
        },
        "/",
      ),
      layout: {
        ...layout,
        sidebar: { ...layout.sidebar, activeKey: "topics" },
      },
    } as AnyPagePayload;
    const categories = {
      ...page("categories.index", { categories: [], total: 0 }, "/categories"),
      layout: {
        ...layout,
        sidebar: { ...layout.sidebar, activeKey: "categories" },
      },
    } as AnyPagePayload;
    const source: PageSource<AnyPagePayload> = {
      api: {} as GooseSiteApi,
      load: vi.fn(async (url: URL) =>
        url.pathname === "/categories" ? categories : home,
      ),
    };
    const user = userEvent.setup();
    render(<SiteApp pageSource={source} />);

    const sidebar = await screen.findByRole("complementary", { name: "Sidebar" });
    const categoriesLink = sidebar.querySelector<HTMLAnchorElement>('a[href="/categories"]');
    expect(categoriesLink?.className).toContain("hover:bg-accent");
    await user.click(categoriesLink!);

    await waitFor(() => {
      const currentSidebar = screen.getByRole("complementary", { name: "Sidebar" });
      expect(document.querySelectorAll('aside[aria-label="Sidebar"]')).toHaveLength(1);
      expect(
        currentSidebar.querySelector('a[href="/categories"]')?.getAttribute("aria-current"),
      ).toBe("page");
    });
    expect(await screen.findByText(/暂无分类|No categories/)).toBeTruthy();
  });

  it("blocks dirty SPA navigation until the editor decision is resolved", async () => {
    window.history.replaceState(null, "", "/publish");
    const publish = page(
      "publish.index",
      {
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
      },
      "/publish",
    );
    const home = page(
      "home.index",
      {
        sort: "latest",
        tabs: [],
        topics: [],
        pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" },
        announcement: { enabled: false, html: "" },
      },
      "/",
    );
    const load = vi.fn(async (url: URL) =>
      url.pathname === "/publish" ? publish : home,
    );
    const source: PageSource<AnyPagePayload> = {
      api: {} as GooseSiteApi,
      load,
    };
    const user = userEvent.setup();
    render(<SiteApp pageSource={source} />);
    await user.type(
      await screen.findByRole("textbox", { name: /标题|Title/ }),
      "Unsaved",
    );
    await user.click(screen.getByRole("link", { name: /取消|Cancel/ }));
    expect(
      await screen.findByRole("dialog", {
        name: /保存未完成的编辑|Save unfinished edits/,
      }),
    ).toBeTruthy();
    expect(window.location.pathname).toBe("/publish");
    expect(load).toHaveBeenCalledTimes(1);
    await user.click(
      screen.getByRole("button", { name: /不保存离开|Leave without saving/ }),
    );
    await waitFor(() => expect(window.location.pathname).toBe("/"));
    expect(await screen.findByText(/暂无主题|No topics/)).toBeTruthy();
  });
});
