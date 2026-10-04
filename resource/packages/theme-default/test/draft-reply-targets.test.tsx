import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { GooseClientError, type GooseSiteApi, type PostPayload, type PostWindowPayload } from "@gooseforum/client";
import { useDraftReplyTargets } from "../src/site/drafts/use-draft-reply-targets";
import type { LocalDraft } from "../src/site/drafts/local-draft-store";

afterEach(cleanup);
const distantPost: PostPayload = { id: 199, topicId: 60, postNo: 100, content: "Distant reply", renderedContent: "<p>Distant reply</p>", processStatus: 0, isHidden: false, canModerate: false, author: { id: 8, username: "bob", avatarUrl: "" }, createdAt: "2026-10-03T00:00:00Z", isOwnPost: false };
const initial = { ...distantPost, id: 100, postNo: 1 };
const draft = { id: "copy", replyTargetId: 199 } as LocalDraft;
function window(post = distantPost): PostWindowPayload { return { posts: [post], replyTargets: [], hasBefore: true, hasAfter: false, total: 100, maxPostNo: 100 }; }
function options(fetch: GooseSiteApi["posts"]["window"]) { return { candidates: [draft, { ...draft, id: "other-copy" }], posts: [initial], references: [], api: { window: fetch } as GooseSiteApi["posts"], topicId: 60, context: "7:60:1", enabled: true }; }

it("checks a distant target once without replacing the visible post window", async () => {
  let resolve!: (window: PostWindowPayload) => void;
  const fetch = vi.fn(() => new Promise<PostWindowPayload>((done) => { resolve = done; }));
  const input = options(fetch);
  const { result } = renderHook(() => useDraftReplyTargets(input));
  expect(result.current.target(199).status).toBe("checking");
  expect(fetch).toHaveBeenCalledExactlyOnceWith({ topicId: 60, anchorPostId: 199, limit: 1 });
  await act(async () => { resolve(window()); await Promise.resolve(); });
  expect(result.current.target(199)).toEqual({ status: "available", post: distantPost });
  expect(input.posts).toEqual([initial]);
});

it("keeps hidden, missing and denied targets unavailable", async () => {
  for (const response of [window({ ...distantPost, isHidden: true }), { ...window(), posts: [] }, new GooseClientError("missing", { messageCode: "post.notFound" }), new GooseClientError("forbidden", { status: 403 })]) {
    const fetch = response instanceof Error ? vi.fn().mockRejectedValue(response) : vi.fn().mockResolvedValue(response);
    const input = options(fetch);
    const { result, unmount } = renderHook(() => useDraftReplyTargets(input));
    await waitFor(() => expect(result.current.target(199).status).toBe("unavailable"));
    expect(result.current.target(199).post).toBeUndefined();
    unmount();
  }
});

it("offers a retry after network failure and does not misreport a deleted target", async () => {
  const fetch = vi.fn().mockRejectedValueOnce(new Error("offline")).mockResolvedValue(window());
  const input = options(fetch);
  const { result } = renderHook(() => useDraftReplyTargets(input));
  await waitFor(() => expect(result.current.target(199).status).toBe("error"));
  act(() => result.current.retry());
  await waitFor(() => expect(result.current.target(199).status).toBe("available"));
  expect(fetch).toHaveBeenCalledTimes(2);
});

it("ignores an outstanding check after account or composer context changes", async () => {
  let resolve!: (window: PostWindowPayload) => void;
  const fetch = vi.fn(() => new Promise<PostWindowPayload>((done) => { resolve = done; }));
  const input = options(fetch);
  const { result, rerender } = renderHook((props) => useDraftReplyTargets(props), { initialProps: input });
  rerender({ ...input, context: "8:60:2", enabled: false });
  await act(async () => { resolve(window()); await Promise.resolve(); });
  expect(result.current.target(199).status).toBe("unavailable");
  expect(result.current.target(199).post).toBeUndefined();
});
