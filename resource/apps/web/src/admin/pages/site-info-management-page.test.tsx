import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";
import type { GooseAdminApi } from "@gooseforum/client";
import text from "../messages/zh-settings";
import { SiteInfoManagementPage } from "./site-info-management-page";

afterEach(() => { cleanup(); vi.unstubAllGlobals(); });

it("saves the uploaded object key and preserves it while assignment is disabled", async () => {
  vi.stubGlobal("ResizeObserver", class { observe() {} unobserve() {} disconnect() {} });
  const saveSite = vi.fn().mockResolvedValue({ avatarFiles: { "2026/one.png": "avatars/default/v1/one.png" } });
  const uploadImage = vi.fn().mockImplementation(async (file: File) => ({ objectName: "2026/" + file.name }));
  const api = { settings: { site: async () => ({ defaultAvatarFile: "legacy.png" }), saveSite }, pages: { uploadImage } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<SiteInfoManagementPage api={api} text={key => text[key]} />);
  const toggle = await screen.findByRole("switch", { name: text.assignAvatar });
  expect(toggle.getAttribute("aria-checked")).toBe("true");
  expect(screen.getByRole("button", { name: text.builtinAvatars + " 1" }).getAttribute("aria-pressed")).toBe("false");
  await user.upload(screen.getByLabelText(text.uploadAvatars), ["one.png", "two.png"].map(name => new File(["avatar"], name, { type: "image/png" })));
  await screen.findByRole("button", { name: text.customAvatars + " 3" });
  await user.click(toggle);
  await user.click(screen.getByRole("button", { name: text.save }));
  expect(saveSite).toHaveBeenCalledWith(expect.objectContaining({ disableDefaultAvatar: true, avatarPool: expect.arrayContaining([
    { file: "legacy.png", enabled: true }, { file: "2026/one.png", enabled: true }, { file: "2026/two.png", enabled: true },
  ]) }));
  const preview = screen.getByRole("button", { name: text.customAvatars + " 2" }).querySelector("img");
  expect(preview?.getAttribute("src")).toBe("/file/img/avatars/default/v1/one.png");
  await user.click(toggle);
  for (const button of screen.getAllByRole("button", { name: text.removeAvatar })) await user.click(button);
  expect(screen.getByRole("alert").textContent).toBe(text.selectAvatar);
  expect((screen.getByRole("button", { name: text.save }) as HTMLButtonElement).disabled).toBe(true);
  await user.click(screen.getByRole("button", { name: text.enableAll }));
  await user.click(screen.getByRole("button", { name: text.save }));
  expect(saveSite.mock.lastCall?.[0].avatarPool).toHaveLength(12);
});
