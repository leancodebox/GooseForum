import { afterEach, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { GooseAdminApi } from "@gooseforum/client";
import { toast } from "sonner";
import messages from "../messages/zh-identity-settings";
import { OAuthSettingsPage } from "./oauth-settings-page";

vi.stubGlobal("ResizeObserver", class {
  observe() {}
  unobserve() {}
  disconnect() {}
});

afterEach(cleanup);

it("explains and copies the callback assembled from three segments", async () => {
  const api = { settings: {
    oauth: vi.fn().mockResolvedValue({ siteUrl: "https://forum.example/base/", providers: [] }),
  } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  const writeText = vi.spyOn(navigator.clipboard, "writeText");
  render(<OAuthSettingsPage api={api} text={(key) => messages[key]} />);

  await waitFor(() => expect(api.settings.oauth).toHaveBeenCalled());
  await user.click(screen.getByRole("button", { name: "添加提供方" }));
  await user.type(screen.getByRole("textbox", { name: "提供方标识" }), "Company-SSO");

  const callback = screen.getByLabelText("回调地址");
  expect(callback.textContent).toBe(
    "https://forum.example/base/api/auth/company-sso/callback",
  );

  await user.hover(screen.getByText("/api/auth/"));
  expect(await screen.findByText("GooseForum 固定的 OAuth 回调路径。")).toBeTruthy();

  await user.click(screen.getByRole("button", { name: "复制完整回调地址" }));
  expect(writeText).toHaveBeenCalledWith(
    "https://forum.example/base/api/auth/company-sso/callback",
  );
});

it("shows a localized actionable message when OAuth settings cannot be saved", async () => {
  const error = vi.spyOn(toast, "error").mockImplementation(() => "");
  const api = { settings: {
    oauth: vi.fn().mockResolvedValue({
      siteUrl: "https://forum.example",
      providers: [{
        key: "github",
        displayName: "GitHub",
        kind: "github",
        enabled: false,
        clientId: "",
        clientSecretConfigured: false,
        callbackUrl: "https://forum.example/api/auth/github/callback",
      }],
    }),
    saveOAuth: vi.fn().mockRejectedValue(new Error("OIDC provider requires a discovery URL")),
  } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<OAuthSettingsPage api={api} text={(key) => messages[key]} />);

  await screen.findByRole("button", { name: "保存" });
  await user.click(screen.getByRole("button", { name: "保存" }));
  await waitFor(() => expect(error).toHaveBeenCalledWith("保存失败", {
    description: "无法验证提供方配置，请检查凭据、Discovery URL 和网络连接。",
  }));
  error.mockRestore();
});

it("renders a retryable error state when settings fail to load", async () => {
  const oauth = vi.fn()
    .mockRejectedValueOnce(new Error("internal error"))
    .mockResolvedValueOnce({ siteUrl: "https://forum.example", providers: [] });
  const api = { settings: { oauth } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<OAuthSettingsPage api={api} text={(key) => messages[key]} />);

  expect(await screen.findByText("设置加载失败")).toBeTruthy();
  expect(screen.queryByText("internal error")).toBeNull();
  expect(screen.getByRole("button", { name: "添加提供方" })).toHaveProperty("disabled", true);
  expect(screen.getByRole("button", { name: "保存" })).toHaveProperty("disabled", true);
  await user.click(screen.getByRole("button", { name: "重试" }));
  await waitFor(() => expect(oauth).toHaveBeenCalledTimes(2));
  expect(await screen.findByRole("button", { name: "添加提供方" })).toBeTruthy();
});

it("confirms before removing a custom provider", async () => {
  const api = { settings: {
    oauth: vi.fn().mockResolvedValue({
      siteUrl: "https://forum.example",
      providers: [{
        key: "company-sso",
        displayName: "Company SSO",
        kind: "oidc",
        enabled: false,
        clientId: "",
        clientSecretConfigured: false,
        callbackUrl: "https://forum.example/api/auth/company-sso/callback",
        discoveryUrl: "",
        scopes: ["openid"],
      }],
    }),
  } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<OAuthSettingsPage api={api} text={(key) => messages[key]} />);

  await screen.findByRole("tab", { name: /Company SSO/ });
  await user.click(screen.getByRole("button", { name: "删除提供方" }));
  expect(screen.getByRole("dialog")).toBeTruthy();
  expect(screen.getByText("删除这个 OAuth 提供方？")).toBeTruthy();
  await user.click(screen.getByRole("button", { name: "删除" }));
  expect(screen.queryByText("Company SSO")).toBeNull();
});
