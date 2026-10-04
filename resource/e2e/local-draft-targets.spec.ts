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

test("restores a local reply to a distant post after checking its visibility", async ({ page }, testInfo) => {
  await page.addInitScript(() => {
    localStorage.setItem("goose:local-draft-account", `${location.origin}:7`);
    const request = indexedDB.open("gooseforum-local-drafts", 1);
    request.onupgradeneeded = () => request.result.createObjectStore("drafts", { keyPath: "id" }).createIndex("owner", "owner");
    request.onsuccess = () => {
      const database = request.result;
      const transaction = database.transaction("drafts", "readwrite");
      transaction.objectStore("drafts").put({ id: "long-reply-copy", owner: `${location.origin}:7`, kind: "new-reply", objectId: 60, topicId: 60, title: "", content: "Recovered reply to post 100", categoryIds: [], replyTargetId: 199, sourceVersion: 1, sourceFingerprint: "previous", revision: 1, updatedAt: Date.now(), bytes: 100 });
      transaction.oncomplete = () => database.close();
    };
  });
  await page.route("**/__goose_page/**", (route) => route.fulfill({ json: topicPage }));
  let complete!: () => void;
  let checks = 0;
  await page.route("**/api/forum/posts/window**", async (route) => {
    if (new URL(route.request().url()).searchParams.get("anchorPostId") === "199") {
      checks++;
      await new Promise<void>((resolve) => { complete = resolve; });
      await route.fulfill({ json: { code: 0, result: { ...stream, posts: [target], beforePostNo: 100, afterPostNo: 100, hasBefore: true, hasAfter: false } } });
    } else await route.fulfill({ json: { code: 0, result: { ...stream, hasAfter: false } } });
  });
  await page.goto("/p/post/60?localReply=1&lang=en");
  const panel = page.locator(".goose-composer-surface section");
  const prompt = page.getByRole("dialog", { name: "Local draft found" });
  const restore = prompt.getByRole("button", { name: "Restore copy", exact: true });
  await expect(restore).toBeVisible();
  await expect(restore).toBeDisabled();
  await expect(prompt.getByText(/^Loading more replies/)).toBeVisible();
  await expect(prompt.getByText("Permissions or the reply target changed; this copy cannot be restored", { exact: true })).toHaveCount(0);
  await page.screenshot({ path: testInfo.outputPath("draft-recovery-prompt.png"), fullPage: true });
  await expect.poll(() => checks).toBe(1);
  complete();
  await expect(restore).toBeEnabled();
  await restore.click();
  await expect(prompt).not.toBeVisible();
  await panel.getByRole("radio", { name: "Markdown", exact: true }).click();
  await expect(panel.locator("textarea")).toHaveValue("Recovered reply to post 100");
  await expect(panel.getByText("Reply to bob · #100", { exact: true })).toBeVisible();
  await expect(page.locator('[data-post-no="1"]')).toBeVisible();
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
  await page.screenshot({ path: testInfo.outputPath("long-reply-draft.png"), fullPage: true });
});
