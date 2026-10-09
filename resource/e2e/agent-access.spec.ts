import { expect, test } from "@playwright/test";
import { readFile } from "node:fs/promises";

test.beforeEach(async ({ page }) => {
  await page.routeWebSocket(/ws:\/\/127\.0\.0\.1:3011\//, (socket) => {
    socket.send(JSON.stringify({ type: "connected" }));
  });
  await page.route("**/static/pic/**", async (route) => {
    const file = new URL(route.request().url()).pathname.split("/").at(-1)!;
    if (!/^(\d+|default-avatar)\.webp$/.test(file)) { await route.abort(); return; }
    await route.fulfill({ contentType: "image/webp", body: await readFile(new URL(`../static/pic/${file}`, import.meta.url)) });
  });
});

const layout = {
  site: {
    name: "GooseForum",
    description: "",
    logo: "",
    favicon: "",
    brandType: "default",
    brandText: "",
    brandImage: "",
  },
  viewer: {
    id: 7,
    username: "agent-user",
    email: "agent@example.test",
    avatarUrl: "",
    isAuthenticated: true,
    canAccessAdmin: false,
    isModerator: false,
    requiresEmailVerification: false,
    adminPermissions: [],
  },
  header: [],
  sidebar: { activeKey: "topics", categories: [] },
  footer: { links: [], primary: [] },
  unread: { notifications: false, messages: false },
  theme: { enabled: false, current: "gf-light", themeColor: "#fbfdff" },
};

test("admin saves Agent policy settings", async ({ page }, testInfo) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  let settings = { enabled: true, manualTokens: true, readPerMinute: 60, userReadPerMinute: 120, writePerMinute: 10, anonymousReadPerMinute: 30, ipPerMinute: 120 };
  await page.route("**/__goose_page/**", (route) => route.fulfill({ json: {
    component: "admin.shell", props: {}, layout: { ...layout, viewer: { ...layout.viewer, canAccessAdmin: true, adminPermissions: [5] } },
    meta: { title: "Agent access" }, url: "/admin/settings/agent?lang=en", version: "1.0",
  } }));
  await page.route("**/api/admin/agent-settings", (route) => route.fulfill({ json: { code: 0, result: settings } }));
  await page.route("**/api/admin/save-agent-settings", (route) => {
    settings = route.request().postDataJSON().settings;
    return route.fulfill({ json: { code: 0, result: true } });
  });
  await page.goto("/admin/settings/agent?lang=en");
  await expect(page.getByRole("heading", { name: "Agent access", exact: true, level: 2 })).toBeVisible();
  await page.getByRole("switch", { name: "Enable Agent API", exact: true }).uncheck();
  await page.getByRole("switch", { name: "Allow manual tokens", exact: true }).uncheck();
  await page.getByLabel("Per authorization", { exact: true }).fill("17");
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(page.getByText("Settings saved", { exact: true })).toBeVisible();
  expect(settings).toMatchObject({ enabled: false, manualTokens: false, readPerMinute: 17 });
  await page.reload();
  await expect(page.getByRole("switch", { name: "Enable Agent API", exact: true })).not.toBeChecked();
  await expect(page.getByLabel("Per authorization", { exact: true })).toHaveValue("17");
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth + 1)).toBe(false);
  await page.screenshot({ path: `/tmp/gooseforum-agent-admin-${testInfo.project.name}.png`, fullPage: true });
  expect(errors).toEqual([]);
});

test("creates and revokes a fallback token without retaining its secret", async ({
  page,
}, testInfo) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  await page.route("**/__goose_page/**", (route) =>
    route.fulfill({
      json: {
        component: "settings.index",
        version: "1.0",
        url: "/settings?tab=agent-tokens&lang=en",
        meta: { title: "Agent Tokens" },
        layout,
        props: {
          user: {
            id: 7,
            username: "agent-user",
            nickname: "Agent User",
            email: "agent@example.test",
            locale: "en",
            avatarUrl: "",
            profileCoverUrl: "",
            bio: "",
            signature: "",
            websiteName: "",
            website: "",
            prestige: 0,
            createdAt: "2026-01-01T00:00:00Z",
            externalInformation: {},
            badges: [],
            wearableBadges: [],
          },
          stats: {
            topicCount: 0,
            replyCount: 0,
            followerCount: 0,
            followingCount: 0,
            likeReceivedCount: 0,
            likeGivenCount: 0,
            collectionCount: 0,
            createdAt: "2026-01-01T00:00:00Z",
          },
          tabs: [
            { key: "applications", url: "/settings?tab=applications" },
            { key: "agent-tokens", url: "/settings?tab=agent-tokens" },
          ],
          privacy: { showTopics: true, showReplies: true },
        },
      },
    }),
  );
  await page.route("**/api/mfa", (route) =>
    route.fulfill({
      json: {
        code: 0,
        result: { enabled: false, available: true, remainingCodes: 0 },
      },
    }),
  );
  await page.route("**/api/forum/unread-status", (route) =>
    route.fulfill({
      json: { code: 0, result: { notifications: false, messages: false } },
    }),
  );
  let entry: Record<string, unknown> | undefined;
  let request: Record<string, unknown> | undefined;
  await page.route("**/api/agent-tokens", (route) =>
    route.fulfill({ json: { code: 0, result: entry ? [entry] : [] } }),
  );
  await page.route("**/api/agent-tokens/create", (route) => {
    request = route.request().postDataJSON();
    entry = {
      id: "token-1",
      name: request!.name,
      prefix: "gf_agent_example",
      scopes: request!.scopes,
      createdAt: "2026-10-09T00:00:00Z",
      expiresAt: "2026-11-08T00:00:00Z",
      revokedAt: null,
    };
    return route.fulfill({
      json: {
        code: 0,
        result: { entry, token: "gf_agent_example_secret_only_once" },
      },
    });
  });
  await page.route("**/api/agent-tokens/revoke", (route) => {
    expect(route.request().postDataJSON()).toEqual({ id: "token-1" });
    entry = { ...entry, revokedAt: "2026-10-09T01:00:00Z" };
    return route.fulfill({ json: { code: 0, result: true } });
  });
  await page.route("**/api/agent-tokens/delete", (route) => {
    expect(route.request().postDataJSON()).toEqual({ id: "token-1" });
    entry = undefined;
    return route.fulfill({ json: { code: 0, result: true } });
  });
  await page.goto("/settings?tab=agent-tokens&lang=en");
  await expect(page).toHaveTitle(/Agent Tokens/);
  await expect(
    page.getByRole("heading", { name: "Agent Tokens", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Create token", exact: true }),
  ).toBeDisabled();
  await page.getByLabel("Name", { exact: true }).fill("Research agent");
  await page
    .getByLabel("Current password", { exact: true })
    .fill("example-password");
  await page
    .getByRole("checkbox", { name: "Reply to topics", exact: true })
    .check();
  await page.getByRole("checkbox", { name: "Upload images", exact: true }).check();
  await page.getByRole("button", { name: "Create token", exact: true }).click();
  await expect(page.getByLabel("Token", { exact: true })).toHaveValue(
    "gf_agent_example_secret_only_once",
  );
  await page.getByRole("button", { name: "Copy", exact: true }).click();
  await expect(page.getByRole("button", { name: "Copied", exact: true })).toBeVisible();
  expect(await page.evaluate(() => navigator.clipboard.readText())).toBe("gf_agent_example_secret_only_once");
  await expect(
    page.getByLabel("Current password", { exact: true }),
  ).toHaveValue("");
  expect(request).toMatchObject({
    name: "Research agent",
    days: 30,
    scopes: ["forum:read", "posts:create", "images:upload"],
  });
  await page.screenshot({
    path: `/tmp/gooseforum-agent-token-${testInfo.project.name}.png`,
    fullPage: true,
  });
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth > innerWidth + 1,
    ),
  ).toBe(false);
  await page.getByRole("button", { name: "Dismiss", exact: true }).click();
  await expect(page.getByLabel("Token", { exact: true })).toHaveCount(0);
  await page
    .getByRole("button", { name: "Revoke Research agent", exact: true })
    .click();
  await page.getByRole("button", { name: "Revoke", exact: true }).click();
  await expect(page.getByText("Revoked", { exact: true })).toBeVisible();
  await page.getByRole("button", { name: "Delete Research agent", exact: true }).click();
  await page.getByRole("button", { name: "Delete", exact: true }).click();
  await expect(page.getByText("No tokens", { exact: true })).toBeVisible();
  expect(errors).toEqual([]);
});

test("explains forum permissions in browser consent and returns to the client", async ({
  page,
}, testInfo) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  await page.route("**/__goose_page/**", (route) =>
    route.fulfill({
      json: {
        component: "auth.oidcConsent",
        props: { interaction: "agent-consent" },
        layout,
        meta: { title: "Agent authorization" },
        url: "/oauth2/consent?interaction=agent-consent&lang=en",
        version: "1.0",
      },
    }),
  );
  await page.route("**/oauth2/consent/details**", (route) =>
    route.fulfill({
      json: {
        client: { id: "agent-client", name: "Research agent", public: true },
        scopes: [
          "openid",
          "forum:read",
          "topics:create",
          "posts:create",
          "images:upload",
          "offline_access",
        ],
        expires_at: "2099-01-01T00:00:00Z",
      },
    }),
  );
  let decision: unknown;
  await page.route("**/oauth2/consent", (route) => {
    decision = route.request().postDataJSON();
    return route.fulfill({
      json: {
        redirect_url:
          "http://127.0.0.1:3011/agent-callback?code=example&state=verified",
      },
    });
  });
  await page.route("**/agent-callback?**", (route) =>
    route.fulfill({
      contentType: "text/html",
      body: "<title>Authorized</title><p>Authorized</p>",
    }),
  );
  // Vite proxies /oauth2 to Go; render its SPA shell for this isolated UI test.
  await page.route((url) => url.pathname === "/oauth2/consent", async (route) => {
    if (route.request().resourceType() !== "document") { await route.fallback(); return; }
    await route.fulfill({ response: await page.request.get("/") });
  });
  await page.goto("/oauth2/consent?interaction=agent-consent&lang=en");
  await expect(
    page.getByText("Read forum content you can access", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("Publish topics on your behalf", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("Reply to topics on your behalf", { exact: true }),
  ).toBeVisible();
  await page.screenshot({
    path: `/tmp/gooseforum-agent-consent-${testInfo.project.name}.png`,
    fullPage: true,
  });
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth > innerWidth + 1,
    ),
  ).toBe(false);
  await page
    .getByRole("button", { name: "Allow and continue", exact: true })
    .click();
  await expect(page).toHaveURL(/agent-callback\?code=example&state=verified/);
  expect(decision).toEqual({
    interaction: "agent-consent",
    decision: "approve",
  });
  expect(errors).toEqual([]);
});
