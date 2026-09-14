import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type {
  AnyPagePayload,
  GooseAdminApi,
  GooseSiteApi,
  LayoutPayload,
} from "@gooseforum/client";
import type { PageSource } from "../browser-runtime";
import { SiteApp } from "./SiteApp";

afterEach(cleanup);

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

describe("SiteApp navigation lifecycle", () => {
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
      admin: {} as GooseAdminApi,
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
