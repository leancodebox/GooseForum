import { afterEach, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { GooseAdminApi } from "@gooseforum/client";
import messages from "../messages/zh-content-settings";
import { AnnouncementSettingsPage } from "./announcement-settings-page";

afterEach(() => { cleanup(); vi.restoreAllMocks(); });

it("formats, previews, and saves announcement Markdown", async () => {
  const saveAnnouncement = vi.fn().mockResolvedValue(undefined);
  const api = {
    settings: {
      announcement: vi.fn().mockResolvedValue({ enabled: true, content: "公告" }),
      saveAnnouncement,
    },
    pages: { uploadImage: vi.fn() },
  } as unknown as GooseAdminApi;
  const user = userEvent.setup();

  render(
    <AnnouncementSettingsPage
      api={api}
      text={(key) => messages[key]}
    />,
  );

  await screen.findByRole("radio", { name: "Markdown" });
  await user.click(screen.getByRole("radio", { name: "Markdown" }));
  const editor = screen.getByRole("textbox", { name: "公告内容" });
  editor.focus();
  (editor as HTMLTextAreaElement).setSelectionRange(0, 2);
  await user.click(screen.getByRole("button", { name: "加粗" }));
  await waitFor(() =>
    expect((editor as HTMLTextAreaElement).value).toBe("**公告**"),
  );

  await user.click(screen.getByRole("radio", { name: "预览" }));
  const preview = document.querySelector('[data-slot="announcement-markdown-preview"]');
  expect(preview?.innerHTML).toContain("<strong>公告</strong>");

  await user.click(screen.getByRole("button", { name: "保存" }));
  await waitFor(() =>
    expect(saveAnnouncement).toHaveBeenCalledWith({
      enabled: true,
      content: "",
      items: [{ id: "legacy", title: "", content: "**公告**", enabled: true }],
    }),
  );
});

it("uses timestamp IDs without collisions when adding in the same millisecond", async () => {
  vi.spyOn(Date, "now").mockReturnValue(1791021600000);
  const saveAnnouncement = vi.fn().mockResolvedValue(undefined);
  const api = { settings: {
    announcement: vi.fn().mockResolvedValue({ enabled: true, content: "", items: [] }),
    saveAnnouncement,
  }, pages: { uploadImage: vi.fn() } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<AnnouncementSettingsPage api={api} text={key => messages[key]} />);
  await waitFor(() => expect(screen.getByRole("button", { name: "保存" }).hasAttribute("disabled")).toBe(false));
  await user.click(screen.getByRole("button", { name: "添加" }));
  await user.click(screen.getByRole("switch", { name: "启用公告 1" }));
  await user.click(screen.getByRole("button", { name: "添加" }));
  await user.click(screen.getByRole("switch", { name: "启用公告 2" }));
  await user.click(screen.getByRole("button", { name: "保存" }));
  await waitFor(() => expect(saveAnnouncement).toHaveBeenCalledOnce());
  const items = (saveAnnouncement.mock.calls[0][0] as { items: Array<{ id: string }> }).items;
  expect(items).toHaveLength(2);
  expect(items[0].id).toBe("1791021600000");
  expect(items[1].id).toBe("1791021600001");
});

it("saves an authoritative empty list after deleting a legacy announcement", async () => {
  const saveAnnouncement = vi.fn().mockResolvedValue(undefined);
  const api = { settings: {
    announcement: vi.fn().mockResolvedValue({ enabled: true, content: "旧公告" }),
    saveAnnouncement,
  }, pages: { uploadImage: vi.fn() } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<AnnouncementSettingsPage api={api} text={key => messages[key]} />);
  await user.click(await screen.findByRole("button", { name: messages.remove }));
  await user.click(screen.getByRole("button", { name: "保存" }));
  await waitFor(() => expect(saveAnnouncement).toHaveBeenCalledWith({ enabled: true, content: "", items: [] }));
});
