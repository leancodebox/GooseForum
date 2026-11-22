import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";
import type { GooseAdminApi } from "@gooseforum/client";
import text from "../messages/zh-settings";
import { SiteInfoManagementPage } from "./site-info-management-page";

afterEach(() => { cleanup(); vi.restoreAllMocks(); vi.unstubAllGlobals(); });

it("saves the uploaded object key and preserves it while assignment is disabled", async () => {
  vi.stubGlobal("ResizeObserver", class { observe() {} unobserve() {} disconnect() {} });
  const bitmap = { width: 800, height: 400, close: vi.fn() };
  vi.stubGlobal("createImageBitmap", vi.fn().mockResolvedValue(bitmap));
  const drawImage = vi.fn();
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue({ drawImage } as unknown as CanvasRenderingContext2D);
  vi.spyOn(HTMLCanvasElement.prototype, "toBlob").mockImplementation(function (this: HTMLCanvasElement, callback, type) {
    expect([this.width, this.height, type]).toEqual([256, 256, "image/webp"]);
    callback(new Blob(["encoded"], { type }));
  });
  const saveSite = vi.fn().mockResolvedValue("success");
  const uploadImage = vi.fn().mockImplementation(async (file: File) => ({ objectName: "2026/" + file.name }));
  const api = { settings: { site: async () => ({ builtinAvatarFiles: ["/static/pic/1.webp", "/static/pic/2.webp"], defaultAvatarFile: "legacy.png" }), saveSite }, pages: { uploadImage } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<SiteInfoManagementPage api={api} text={key => text[key]} />);
  const toggle = await screen.findByRole("switch", { name: text.assignAvatar });
  await user.selectOptions(screen.getByLabelText(text.defaultLocale), "ja");
  const manage = screen.getByRole("button", { name: text.manageAvatars });
  expect(manage.getAttribute("aria-expanded")).toBe("false");
  expect(screen.queryByLabelText(text.uploadAvatars)).toBeNull();
  manage.focus();
  await user.keyboard("{Enter}");
  expect(manage.getAttribute("aria-expanded")).toBe("true");
  expect(toggle.getAttribute("aria-checked")).toBe("true");
  expect(screen.getByRole("button", { name: text.builtinAvatars + " 1" }).getAttribute("aria-pressed")).toBe("false");
  await user.upload(screen.getByLabelText(text.uploadAvatars), ["one.png", "two.png"].map(name => new File(["avatar"], name, { type: "image/png" })));
  await screen.findByRole("button", { name: text.customAvatars + " 3" });
  expect(drawImage).toHaveBeenCalledWith(bitmap, 200, 0, 400, 400, 0, 0, 256, 256);
  await user.click(manage);
  expect(screen.getByText("已启用 3 张")).toBeTruthy();
  await user.click(manage);
  await user.click(toggle);
  await user.click(screen.getByRole("button", { name: text.save }));
  expect(saveSite).toHaveBeenCalledWith(expect.objectContaining({ defaultLocale: "ja", disableDefaultAvatar: true, avatarPool: expect.arrayContaining([
    { file: "legacy.png", enabled: true }, { file: "2026/one.webp", enabled: true }, { file: "2026/two.webp", enabled: true },
  ]) }));
  const preview = screen.getByRole("button", { name: text.customAvatars + " 2" }).querySelector("img");
  expect(preview?.getAttribute("src")).toBe("/file/img/2026/one.webp");
  await user.click(toggle);
  for (const button of screen.getAllByRole("button", { name: new RegExp(text.removeAvatar) })) await user.click(button);
  expect(screen.getByRole("alert").textContent).toBe(text.selectAvatar);
  expect((screen.getByRole("button", { name: text.save }) as HTMLButtonElement).disabled).toBe(true);
  await user.click(screen.getByRole("button", { name: text.enableAll }));
  await user.click(screen.getByRole("button", { name: text.save }));
  expect(saveSite.mock.lastCall?.[0].avatarPool).toHaveLength(2);
});
