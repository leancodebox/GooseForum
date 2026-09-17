import { afterEach, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { GooseAdminApi } from "@gooseforum/client";
import { toast } from "sonner";
import messages from "../messages/zh-identity-settings";
import { OAuthSettingsPage } from "./oauth-settings-page";

afterEach(cleanup);

it("computes a new provider callback from the site URL and edited key", async () => {
  const api = { settings: {
    oauth: vi.fn().mockResolvedValue({ siteUrl: "https://forum.example/base/", providers: [] }),
  } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<OAuthSettingsPage api={api} text={(key) => messages[key]} />);

  await waitFor(() => expect(api.settings.oauth).toHaveBeenCalled());
  await user.click(screen.getByRole("button", { name: "添加提供方" }));
  await user.type(screen.getByRole("textbox", { name: "提供方标识" }), "Company-SSO");

  expect(screen.getByRole("textbox", { name: "回调地址" })).toHaveProperty(
    "value", "https://forum.example/base/api/auth/company-sso/callback",
  );
});

it("shows the server's reason when OAuth settings cannot be saved", async () => {
  const error = vi.spyOn(toast, "error").mockImplementation(() => "");
  const api = { settings: {
    oauth: vi.fn().mockResolvedValue({ siteUrl: "https://forum.example", providers: [] }),
    saveOAuth: vi.fn().mockRejectedValue(new Error("OIDC provider requires a discovery URL")),
  } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<OAuthSettingsPage api={api} text={(key) => messages[key]} />);

  await screen.findByRole("button", { name: "保存" });
  await user.click(screen.getByRole("button", { name: "保存" }));
  await waitFor(() => expect(error).toHaveBeenCalledWith("OIDC provider requires a discovery URL"));
  error.mockRestore();
});
