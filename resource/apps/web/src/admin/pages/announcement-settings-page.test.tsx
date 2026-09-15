import { afterEach, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { GooseAdminApi } from "@gooseforum/client";
import messages from "../messages/zh-content-settings";
import { AnnouncementSettingsPage } from "./announcement-settings-page";

afterEach(cleanup);

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

  const editor = await screen.findByRole("textbox", { name: "公告内容" });
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
      content: "**公告**",
    }),
  );
});
