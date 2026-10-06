import { afterEach, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { GooseAdminApi, LayoutPayload } from "@gooseforum/client";
import messages from "../messages/zh-settings";
import { SiteChromeManagementPage } from "./site-chrome-management-page";

afterEach(cleanup);

it("folds secondary navigation into More and preserves custom configuration when saving", async () => {
  const item = (id: string, label: string) => ({ id, label, type: "link" as const, enabled: true, url: `/${id}`, i18nLabel: "" });
  const config = {
    header: [],
    mainMenu: [item("custom-main", "自定义主菜单")],
    resources: [item("custom-resource", "自定义资源")],
    sidebarGroups: [{ id: "group", title: "我的分组", i18nLabel: "", items: [item("custom-group", "分组入口")] }],
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
  const menu = within(screen.getByRole("dialog"));
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
  await user.click(editor.getByRole("button", { name: "取消" }));
  await user.click(more);
  await user.click(within(screen.getByRole("dialog")).getByRole("button", { name: "添加" }));
  const addEditor = within(screen.getByRole("dialog", { name: "添加" }));
  expect(addEditor.queryByText("类型")).toBeNull();
  const linkType = addEditor.getByRole("radio", { name: "链接" });
  const textType = addEditor.getByRole("radio", { name: "文本" });
  expect(linkType.getAttribute("data-state")).toBe("on");
  await user.click(textType);
  expect(textType.getAttribute("data-state")).toBe("on");
  expect(linkType.getAttribute("data-state")).toBe("off");
  await user.type(addEditor.getAllByRole("textbox")[0], "新增入口");
  await user.click(addEditor.getByRole("button", { name: "保存" }));
  await user.click(more);
  expect(screen.getByText("新增入口")).toBeTruthy();
  await user.keyboard("{Escape}");
  expect(screen.queryByText("自定义主菜单")).toBeNull();
  await user.click(screen.getByRole("button", { name: "保存" }));
  await waitFor(() => expect(saveChrome).toHaveBeenCalledWith(expect.objectContaining({
    mainMenu: config.mainMenu,
    resources: [...config.resources, expect.objectContaining({ label: "新增入口", type: "text" })],
    sidebarGroups: config.sidebarGroups,
  })));
});
