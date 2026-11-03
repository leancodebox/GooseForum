import { afterEach, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { GooseAdminApi } from "@gooseforum/client";
import messages from "../messages/zh-content-settings";
import { PostingSettingsPage } from "./posting-settings-page";

afterEach(cleanup);

it("loads legacy settings and saves the external link policy alongside existing settings", async () => {
  const savePosting = vi.fn().mockResolvedValue(undefined);
  const api = { settings: {
    posting: vi.fn().mockResolvedValue({ textControl: { minTitleLength: 7 }, uploadControl: {} }),
    savePosting,
  } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<PostingSettingsPage api={api} text={key => messages[key]} />);
  const toggle = screen.getByRole("switch", { name: "开启离站提醒" });
  await waitFor(() => expect(toggle.hasAttribute("disabled")).toBe(false));
  await user.click(toggle);
  const section = screen.getByText("帖子外链安全跳转").closest("fieldset")!;
  await user.click(within(section).getByRole("button", { name: "添加" }));
  await user.type(screen.getByRole("textbox", { name: "域名白名单 1" }), "example.com");
  await user.click(within(section).getByRole("button", { name: "添加" }));
  await user.type(screen.getByRole("textbox", { name: "域名白名单 2" }), "sub.example.com");
  await user.click(screen.getByRole("button", { name: "保存" }));
  await waitFor(() => expect(savePosting).toHaveBeenCalledWith(expect.objectContaining({
    externalLinks: { enabled: true, whitelist: ["example.com", "sub.example.com"] },
    textControl: expect.objectContaining({ minTitleLength: 7 }),
  })));
});

it("edits independent attachment limits without changing image settings", async () => {
  const savePosting = vi.fn().mockResolvedValue(undefined);
  const api = { settings: { posting: vi.fn().mockResolvedValue({ textControl: {}, uploadControl: { allowAttachments: false, maxAttachmentSizeKb: 123 } }), savePosting } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<PostingSettingsPage api={api} text={key => messages[key]} />);
  const limit = screen.getByRole("spinbutton", { name: "每篇内容附件数量上限" });
  await waitFor(() => expect(limit.hasAttribute("disabled")).toBe(false));
  await user.clear(limit);
  await user.type(limit, "3");
  await user.click(screen.getByRole("button", { name: "保存" }));
  await waitFor(() => expect(savePosting).toHaveBeenCalledWith(expect.objectContaining({
    attachments: expect.objectContaining({ minTrustLevel: 100, maxFilesPerPost: 3, extensions: [".pdf", ".txt", ".zip"] }),
    uploadControl: expect.objectContaining({ allowAttachments: false, maxAttachmentSizeKb: 123 }),
  })));
});
