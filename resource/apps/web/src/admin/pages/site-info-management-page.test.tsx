import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";
import type { GooseAdminApi } from "@gooseforum/client";
import text from "../messages/zh-settings";
import { SiteInfoManagementPage } from "./site-info-management-page";

afterEach(cleanup);

it("saves the uploaded object key and preserves it while assignment is disabled", async () => {
  const saveSite = vi.fn().mockResolvedValue(undefined);
  const uploadImage = vi.fn().mockResolvedValue({ url: "https://cdn.example.org/example.png", objectName: "2026/example.png" });
  const api = { settings: { site: async () => ({}), saveSite }, pages: { uploadImage } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<SiteInfoManagementPage api={api} text={key => text[key]} />);
  const toggle = await screen.findByRole("checkbox", { name: text.assignAvatar });
  expect((toggle as HTMLInputElement).checked).toBe(true);
  await user.upload(screen.getByLabelText(text.defaultAvatar), new File(["avatar"], "avatar.png", { type: "image/png" }));
  await screen.findByRole("button", { name: text.removeAvatar });
  await user.click(toggle);
  await user.click(screen.getByRole("button", { name: text.save }));
  expect(saveSite).toHaveBeenCalledWith(expect.objectContaining({ disableDefaultAvatar: true, defaultAvatarFile: "2026/example.png" }));
  await user.click(toggle);
  await user.click(screen.getByRole("button", { name: text.removeAvatar }));
  await user.click(screen.getByRole("button", { name: text.save }));
  expect(saveSite).toHaveBeenLastCalledWith(expect.objectContaining({ disableDefaultAvatar: false, defaultAvatarFile: "" }));
});
