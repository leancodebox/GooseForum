import { expect, test } from "@playwright/test";

const timestamp = "2026-10-03T00:00:00Z";
const post = { id: 61, topicId: 60, postNo: 1, content: "Original topic", renderedContent: "<p>Original topic</p>", sourceVersion: 1, processStatus: 0, isHidden: false, canModerate: false, author: { id: 7, username: "alice", avatarUrl: "" }, createdAt: timestamp, isOwnPost: true };
const target = { ...post, id: 199, postNo: 100, content: "Distant target", renderedContent: "<p>Distant target</p>", author: { id: 8, username: "bob", avatarUrl: "" }, isOwnPost: false };
const stream = { posts: [post], replyTargets: [], beforePostNo: 1, afterPostNo: 1, hasBefore: false, hasAfter: true, total: 100, maxPostNo: 100 };
const layout = {
  site: { name: "GooseForum", description: "", logo: "", favicon: "", brandType: "default", brandText: "", brandImage: "" },
  viewer: { id: 7, username: "alice", email: "", avatarUrl: "", isAuthenticated: true, canAccessAdmin: false, isModerator: false, requiresEmailVerification: false, adminPermissions: [] },
  header: [], sidebar: { activeKey: "topics", categories: [] }, footer: { links: [], primary: [] }, unread: { notifications: false, messages: false }, theme: { enabled: false, current: "gf-light", themeColor: "#fbfdff" },
};
const topicPage = {
  component: "topic.detail", version: "1.0", url: "/p/post/60?localReply=1&lang=en", meta: { title: "Long discussion" }, layout,
  props: { topic: { id: 60, title: "Long discussion", description: "", url: "/p/post/60", topicStatus: 1, processStatus: 0, author: post.author, participants: [], categories: [], replyCount: 99, maxPostNo: 100, viewCount: 1, likeCount: 0, isLiked: false, isBookmarked: false, isWatched: false, createdAt: timestamp, updatedAt: timestamp }, postStream: stream, hotTopics: [], permissions: { isOwnTopic: true, canPost: true, canModerateTopic: false } },
};


test("reply attachments retry, submit and render on desktop and mobile", async ({ page }, testInfo) => {
  const payload = structuredClone(topicPage);
  payload.props.postStream.hasAfter = false;
  await page.route("**/__goose_page/**", route => route.fulfill({ json: payload }));
  await page.route("**/file/attachments/policy", route => route.fulfill({ json: { code: 0, result: { enabled: true, maxFileSizeKb: 5120, maxFilesPerPost: 5, extensions: [".txt", ".pdf", ".zip"] } } }));
  let uploads = 0;
  const file = { id: 401, filename: "notes.txt", size: 12, url: "/file/attachments/401" };
  await page.route("**/file/attachments", async route => {
    uploads++;
    if (uploads === 1) await route.fulfill({ status: 400, json: { code: 1, messageCode: "attachment.invalidFile" } });
    else await route.fulfill({ json: { code: 0, result: file } });
  });
  let submitted: Record<string, unknown> | undefined;
  await page.route("**/api/forum/posts/create", async route => {
    submitted = route.request().postDataJSON();
    await route.fulfill({ json: { code: 0, result: { id: 402, postNo: 2, content: "Reply with attachment", renderedContent: "<p>Reply with attachment</p>", sourceVersion: 1, attachments: [file], version: 1 } } });
  });
  await page.route("**/api/forum/posts/window**", route => route.fulfill({ json: { code: 0, result: { ...stream, hasAfter: false } } }));
  await page.goto("/p/post/60?localReply=1&lang=en");
  const composer = page.locator(".goose-composer-surface").filter({ has: page.getByRole("radio", { name: "Markdown", exact: true }) });
  await composer.getByRole("radio", { name: "Markdown", exact: true }).click();
  await composer.locator("textarea").fill("Reply with attachment");
  await composer.getByLabel("Upload attachment", { exact: true }).setInputFiles({ name: "notes.txt", mimeType: "text/plain", buffer: Buffer.from("hello world!") });
  await expect(composer.getByRole("button", { name: "Retry upload" })).toBeVisible();
  const submit = composer.getByRole("button", { name: "Post reply", exact: true });
  await expect(submit).toBeDisabled();
  await composer.getByRole("button", { name: "Retry upload" }).click();
  await expect(composer.getByRole("link", { name: "notes.txt" })).toBeVisible();
  await expect(submit).toBeEnabled();
  await expect(composer.getByText("Up to 5 files, 5120 KiB each. Types: .txt, .pdf, .zip")).toBeVisible();
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
  await page.screenshot({ path: testInfo.outputPath("attachment-composer.png"), fullPage: true });
  await submit.click();
  await expect.poll(() => submitted?.attachmentIds).toEqual([401]);
  await expect(page.locator('[data-post-no="2"]').getByRole("link", { name: "notes.txt" })).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath("attachment-post.png"), fullPage: true });
});

test("keeps newer reply text without resubmitting attachments already bound to the first reply", async ({ page }) => {
  const payload = structuredClone(topicPage);
  payload.props.postStream.hasAfter = false;
  const file = { id: 401, filename: "notes.txt", size: 12, url: "/file/attachments/401" };
  await page.route("**/__goose_page/**", route => route.fulfill({ json: payload }));
  await page.route("**/file/attachments/policy", route => route.fulfill({ json: { code: 0, result: { enabled: true, maxFileSizeKb: 5120, maxFilesPerPost: 5, extensions: [".txt"] } } }));
  await page.route("**/file/attachments", route => route.fulfill({ json: { code: 0, result: file } }));
  let release!: () => void;
  const pending = new Promise<void>(resolve => { release = resolve; });
  const submissions: Array<{ content: string; attachmentIds: number[] }> = [];
  await page.route("**/api/forum/posts/create", async route => {
    const input = route.request().postDataJSON();
    submissions.push(input);
    const first = submissions.length === 1;
    if (first) await pending;
    await route.fulfill({ json: { code: 0, result: { id: first ? 402 : 403, postNo: first ? 2 : 3, content: input.content, renderedContent: `<p>${input.content}</p>`, sourceVersion: 1, attachments: first ? [file] : [], version: 1 } } });
  });
  await page.route("**/api/forum/posts/window**", route => route.fulfill({ json: { code: 0, result: { ...stream, hasAfter: false } } }));
  await page.goto("/p/post/60?localReply=1&lang=en");
  const composer = page.locator(".goose-composer-surface").filter({ has: page.getByRole("radio", { name: "Markdown", exact: true }) });
  await composer.getByRole("radio", { name: "Markdown", exact: true }).click();
  await composer.locator("textarea").fill("First reply");
  await composer.getByLabel("Upload attachment", { exact: true }).setInputFiles({ name: "notes.txt", mimeType: "text/plain", buffer: Buffer.from("hello world!") });
  await expect(composer.getByRole("link", { name: "notes.txt" })).toBeVisible();
  const submit = composer.getByRole("button", { name: "Post reply", exact: true });
  await submit.click();
  await expect.poll(() => submissions.length).toBe(1);
  expect(submissions[0].attachmentIds).toEqual([401]);
  await composer.locator("textarea").fill("Newer reply text");
  release();
  await expect(submit).toBeEnabled();
  await expect(composer.locator("textarea")).toHaveValue("Newer reply text");
  await expect(composer.getByRole("link", { name: "notes.txt" })).toHaveCount(0);
  await expect(page.locator('[data-post-no="2"]').getByRole("link", { name: "notes.txt" })).toBeVisible();
  await submit.click();
  await expect.poll(() => submissions.length).toBe(2);
  expect(submissions[1]).toMatchObject({ content: "Newer reply text", attachmentIds: [] });
});

test("new topics hide the attachment upload input when permission is denied", async ({ page }) => {
  await page.route("**/__goose_page/**", route => route.fulfill({ json: {
    component: "publish.index", version: "1.0", url: "/publish?lang=en", meta: { title: "New topic" }, layout,
    props: { topicId: 0, isEditing: false, categories: [{ id: 4, name: "Coding", color: "#8241d6", isRestricted: false, canCreate: true }], topic: { title: "", content: "", categoryIds: [], topicStatus: 1, sourceVersion: 1, attachments: [], version: 0 } },
  } }));
  await page.route("**/file/attachments/policy", route => route.fulfill({ status: 403, json: { code: 1, messageCode: "permission.denied" } }));
  await page.goto("/publish?lang=en");
  await expect(page.getByText("Attachment uploads are unavailable for this account.")).toBeVisible();
  await expect(page.getByLabel("Upload attachment", { exact: true })).toHaveCount(0);
});

test("editing a topic removes attachments explicitly and preserves input on conflict", async ({ page }, testInfo) => {
  const file = { id: 401, filename: "original.txt", size: 12, url: "/file/attachments/401" };
  await page.route("**/__goose_page/**", route => route.fulfill({ json: {
    component: "publish.index", version: "1.0", url: "/publish?id=60&lang=en", meta: { title: "Edit topic" }, layout,
    props: { topicId: 60, isEditing: true, categories: [{ id: 4, name: "Coding", color: "#8241d6", isRestricted: false, canCreate: true }], topic: { title: "Topic with attachment", content: "Original topic content", categoryIds: [4], topicStatus: 1, sourceVersion: 1, attachments: [file], version: 7 } },
  } }));
  await page.route("**/file/attachments/policy", route => route.fulfill({ status: 403, json: { code: 1, messageCode: "permission.denied" } }));
  let submitted: Record<string, unknown> | undefined;
  await page.route("**/api/forum/topics/write", async route => {
    submitted = route.request().postDataJSON();
    await route.fulfill({ json: { code: 1, messageCode: "attachment.conflict" } });
  });
  await page.goto("/publish?id=60&lang=en");
  await expect(page.getByRole("link", { name: "original.txt" })).toBeVisible();
  await page.getByRole("button", { name: "Remove original.txt" }).click();
  await expect(page.getByRole("link", { name: "original.txt" })).toHaveCount(0);
  await page.getByRole("button", { name: "Update topic", exact: true }).click();
  await expect.poll(() => submitted?.attachmentIds).toEqual([]);
  expect(submitted?.version).toBe(7);
  await expect(page.getByText("Content changed. Reload before editing again.")).toBeVisible();
  await expect(page.getByRole("textbox", { name: "Title", exact: true })).toHaveValue("Topic with attachment");
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
  await page.screenshot({ path: testInfo.outputPath("attachment-edit-conflict.png"), fullPage: true });
});
