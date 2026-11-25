import { afterEach, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { GooseAdminApi, LayoutPayload } from "@gooseforum/client";
import messages from "../messages/zh-settings";
import { SiteChromeManagementPage } from "./site-chrome-management-page";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

it("folds secondary navigation into More and preserves custom configuration when saving", async () => {
  const fetch = globalThis.fetch;
  vi.stubGlobal("fetch", (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === "string" ? input : "href" in input ? input.href : input.url;
    if (url.endsWith("/lucide.json")) {
      return Promise.resolve(Response.json({ icons: { "book-open": { name: "book-open", size: 24, node: [] } } }));
    }
    return fetch(input, init);
  });
  const item = (id: string, label: string) => ({ id, label, type: "link" as const, enabled: true, url: `/${id}`, i18nLabel: "" });
  const config = {
    header: [],
    mainMenu: [{ ...item("custom-main", "自定义主菜单"), i18nLabel: "legacy.menu" }],
    resources: [item("custom-resource", "自定义资源")],
    sidebarGroups: [{ id: "group", title: "我的分组", i18nLabel: "legacy.group", items: [{ ...item("custom-group", "分组入口"), i18nLabel: "legacy.item" }] }],
  };
  const saveChrome = vi.fn().mockResolvedValue(undefined);
  const api = { settings: { chrome: vi.fn().mockResolvedValue(config), saveChrome } } as unknown as GooseAdminApi;
  const layout = { site: { name: "GooseForum" }, sidebar: { categories: [] } } as unknown as LayoutPayload;
  const user = userEvent.setup();
  render(<SiteChromeManagementPage api={api} layout={layout} text={key => messages[key]} />);

  const more = await screen.findByRole("button", { name: "更多" });
  const sidebar = within(document.querySelector("aside")!);
  expect(sidebar.getByText("成员")).toBeTruthy();
  expect(sidebar.getByText("分类", { selector: "span" })).toBeTruthy();
  expect(sidebar.queryByText("热门")).toBeNull();
  expect(screen.queryByText("自定义主菜单")).toBeNull();
  expect(screen.getByText("分组入口")).toBeTruthy();

  await user.click(more);
  expect(sidebar.getByRole("button", { name: "收起" })).toBe(more);
  expect(screen.queryByRole("dialog")).toBeNull();
  const menu = within(document.getElementById(more.getAttribute("aria-controls")!)!);
  expect(menu.getAllByRole("button", { name: "添加" })).toHaveLength(1);
  expect(menu.queryByText("主菜单")).toBeNull();
  expect(menu.queryByText("资源")).toBeNull();
  expect(screen.getByText("草稿")).toBeTruthy();
  expect(screen.getByText("访问组")).toBeTruthy();
  expect(screen.getByText("主题预览")).toBeTruthy();
  expect(screen.getByText("自定义主菜单")).toBeTruthy();
  expect(screen.getByText("自定义资源")).toBeTruthy();
  await user.click(screen.getByText("自定义主菜单").closest("div")!.querySelector("button")!);
  const editor = within(screen.getByRole("dialog", { name: "编辑" }));
  expect(editor.getByDisplayValue("自定义主菜单")).toBeTruthy();
  expect(editor.queryByText("i18n 键")).toBeNull();
  await user.click(editor.getByRole("button", { name: "取消" }));
  await user.click(menu.getByRole("button", { name: "添加" }));
  const addEditor = within(screen.getByRole("dialog", { name: "添加" }));
  expect(addEditor.queryByText("类型")).toBeNull();
  const linkType = addEditor.getByRole("radio", { name: "链接" });
  const textType = addEditor.getByRole("radio", { name: "文本" });
  expect(linkType.getAttribute("data-state")).toBe("on");
  await user.click(textType);
  expect(textType.getAttribute("data-state")).toBe("on");
  expect(linkType.getAttribute("data-state")).toBe("off");
  await user.type(addEditor.getAllByRole("textbox")[0], "新增入口");
  await user.click(addEditor.getByRole("button", { name: "图标" }));
  const picker = within(await screen.findByRole("dialog", { name: "图标" }));
  await user.type(await picker.findByRole("searchbox", { name: "搜索图标…" }), "book open");
  await user.click(picker.getByRole("button", { name: "book-open" }));
  expect(screen.queryByRole("dialog", { name: "图标" })).toBeNull();
  expect(addEditor.getByRole("button", { name: "图标" }).textContent).toBe("book-open");
  await user.click(addEditor.getByRole("button", { name: "保存" }));
  expect(screen.getByText("新增入口")).toBeTruthy();
  await user.click(more);
  expect(screen.queryByText("自定义主菜单")).toBeNull();

  const linksToggle = screen.getByRole("button", { name: "在顶部导航中显示友情链接" });
  const sponsorsToggle = screen.getByRole("button", { name: "在顶部导航中显示赞助者" });
  const membersToggle = sidebar.getByRole("button", { name: "在侧栏中显示成员" });
  expect(linksToggle.getAttribute("aria-pressed")).toBe("true");
  await user.click(linksToggle);
  expect(linksToggle.getAttribute("aria-pressed")).toBe("false");
  await user.click(membersToggle);
  expect(membersToggle.getAttribute("aria-pressed")).toBe("false");
  await user.click(more);
  const resources = within(document.getElementById(more.getAttribute("aria-controls")!)!);
  const sidebarLinksToggle = resources.getByRole("button", { name: "在侧栏中显示友情链接" });
  const sidebarSponsorsToggle = resources.getByRole("button", { name: "在侧栏中显示赞助者" });
  expect(screen.getAllByRole("button", { name: /^在侧栏中显示/ })).toHaveLength(3);
  expect(sidebarLinksToggle.getAttribute("aria-pressed")).toBe("true");
  await user.click(sidebarSponsorsToggle);
  expect(sidebarSponsorsToggle.getAttribute("aria-pressed")).toBe("false");
  expect(sponsorsToggle.getAttribute("aria-pressed")).toBe("true");
  expect(resources.getByText("自定义资源")).toBeTruthy();
  await user.click(more);

  await user.click(screen.getByRole("button", { name: "保存" }));
  await waitFor(() => expect(saveChrome).toHaveBeenCalledWith(expect.objectContaining({
    header: expect.arrayContaining([
      expect.objectContaining({ id: "links", enabled: false, i18nLabel: "" }),
      expect.objectContaining({ id: "sponsors", enabled: true, i18nLabel: "" }),
    ]),
    hiddenSidebarSystemLinks: ["members", "sponsors"],
    mainMenu: [item("custom-main", "自定义主菜单")],
    resources: [...config.resources, expect.objectContaining({ label: "新增入口", type: "text", icon: "book-open" })],
    sidebarGroups: [{ id: "group", title: "我的分组", i18nLabel: "", items: [item("custom-group", "分组入口")] }],
  })));
});
