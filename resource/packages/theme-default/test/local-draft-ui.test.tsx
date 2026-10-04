import "fake-indexeddb/auto";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { GooseSiteApi, LayoutPayload, PublishPageProps } from "@gooseforum/client";
import { GooseRuntimeProvider, type GooseRuntime } from "@gooseforum/runtime";
import { GooseI18nProvider } from "@gooseforum/runtime/i18n";
import { PublishPageView } from "../src/site/pages/publish-page";
import {
  activateDraftAccount, clearAccountDrafts, draftChangeEvent, draftFingerprint, draftGeneration, draftOwner,
  listLocalDrafts, writeLocalDraft, type DraftValues, type LocalDraft,
} from "../src/site/drafts/local-draft-store";

vi.mock("../src/site/editor/markdown-composer", () => ({
  MarkdownComposer: ({ value, onChange, sourceVersion }: { value: string; onChange(value: string): void; sourceVersion: number }) => <textarea aria-label="Body" value={value} data-version={sourceVersion} onChange={(event) => onChange(event.target.value)} />,
}));

let userId = 1100;
const page: PublishPageProps = {
  topicId: 40, isEditing: true,
  categories: [{ id: 3, name: "General", color: "#ff0000", isRestricted: false, canCreate: true }],
  topic: { title: "Server title", content: "Server body", categoryIds: [3], topicStatus: 1, sourceVersion: 1 },
};
const initial: DraftValues = { title: page.topic.title, content: page.topic.content, categoryIds: [3], replyTargetId: 0, sourceVersion: 1 };
async function seed(overrides: Partial<LocalDraft> = {}) {
  const owner = draftOwner(userId);
  const draft: LocalDraft = { ...initial, content: "Recovered local body", title: "Recovered title", id: "local-copy", owner, kind: "edit-topic", objectId: 40, topicId: 40, sourceFingerprint: draftFingerprint(initial), revision: 1, updatedAt: Date.now(), bytes: 100, ...overrides };
  await writeLocalDraft(draft, draftGeneration(owner));
  return draft;
}
function show(writeReviewed = vi.fn().mockResolvedValue({ id: 40, moderationStatus: "pending", content: "Canonical body", sourceVersion: 1 }), input = page) {
  const runtime = { api: { topics: { writeReviewed } } as unknown as GooseSiteApi, currentUrl: "/publish?id=40", locale: "zh", theme: "gf-light", isNavigating: false, navigate: vi.fn(), redirect: vi.fn(), refresh: vi.fn(), queueFlash: vi.fn(), setLocale: vi.fn(), toggleTheme: vi.fn() } satisfies GooseRuntime;
  render(<GooseI18nProvider locale="zh"><GooseRuntimeProvider runtime={runtime}><PublishPageView page={input} layout={{ viewer: { id: userId, isAuthenticated: true } } as LayoutPayload} /></GooseRuntimeProvider></GooseI18nProvider>);
  return { runtime, writeReviewed };
}

beforeEach(async () => { userId++; await activateDraftAccount(userId); });
afterEach(async () => { cleanup(); await clearAccountDrafts(draftOwner(userId)); vi.restoreAllMocks(); });

it("offers restoration without overwriting server text and deletes a discarded draft", async () => {
  await seed(); show();
  await screen.findByRole("button", { name: "恢复副本" });
  expect(screen.getByLabelText<HTMLTextAreaElement>("Body").value).toBe("Server body");
  expect(screen.getByRole("dialog", { name: "发现本地草稿" })).toBeTruthy();
  expect(screen.getByText<HTMLButtonElement>("保存草稿").disabled).toBe(true);
  expect(screen.getByText<HTMLButtonElement>("更新主题").disabled).toBe(true);
  await userEvent.click(screen.getByRole("button", { name: "丢弃本地草稿" }));
  await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  await waitFor(() => expect(screen.getByRole<HTMLButtonElement>("button", { name: "更新主题" }).disabled).toBe(false));
  expect(await listLocalDrafts(draftOwner(userId))).toHaveLength(0);
});

it("enables mentions for legacy topic edits without an opt-in checkbox", async () => {
  const { writeReviewed } = show(undefined, { ...page, topic: { ...page.topic, sourceVersion: 0 } });
  const body = screen.getByLabelText<HTMLTextAreaElement>("Body");
  expect(body.dataset.version).toBe("1");
  expect(screen.queryByRole("checkbox", { name: "启用 @提及语法" })).toBeNull();
  await userEvent.clear(body);
  await userEvent.type(body, "Hello @alice");
  await waitFor(() => expect(screen.getByRole<HTMLButtonElement>("button", { name: "更新主题" }).disabled).toBe(false));
  await userEvent.click(screen.getByRole("button", { name: "更新主题" }));
  await waitFor(() => expect(writeReviewed).toHaveBeenCalledWith(expect.objectContaining({ sourceVersion: 1, content: "Hello @alice" })));
});

it("does not display a copy's content when the current category is unavailable", async () => {
  await seed({ categoryIds: [999], content: "Private draft content", title: "Private draft title" }); show();
  const restore = await screen.findByRole("button", { name: "恢复副本" });
  expect((restore as HTMLButtonElement).disabled).toBe(true);
  expect(screen.queryByText("Private draft content")).toBeNull();
  expect(screen.queryByText("Private draft title")).toBeNull();
});

it("keeps publishing available after its own autosave notification", async () => {
  show();
  await waitFor(() => expect(screen.getByRole<HTMLButtonElement>('button', { name: '更新主题' }).disabled).toBe(false));
  await userEvent.type(screen.getByLabelText('Body'), ' updated');
  let draft: LocalDraft | undefined;
  await waitFor(async () => {
    draft = (await listLocalDrafts(draftOwner(userId)))[0];
    expect(draft).toBeTruthy();
  }, { timeout: 3000 });
  act(() => {
    window.dispatchEvent(new CustomEvent(draftChangeEvent, { detail: { owner: draftOwner(userId), action: 'write', sourceId: draft!.id } }));
  });
  expect(screen.getByRole<HTMLButtonElement>('button', { name: '更新主题' }).disabled).toBe(false);
});

it("restores explicitly, submits the restored source and clears only after accepted persistence", async () => {
  await seed(); const { writeReviewed } = show();
  await userEvent.click(await screen.findByRole("button", { name: "恢复副本" }));
  expect(screen.getByLabelText<HTMLTextAreaElement>("Body").value).toBe("Recovered local body");
  await userEvent.click(screen.getByRole("button", { name: "保存草稿" }));
  await waitFor(() => expect(writeReviewed).toHaveBeenCalledWith(expect.objectContaining({ title: "Recovered title", content: "Recovered local body", sourceVersion: 1 })));
  await waitFor(() => expect(screen.getByLabelText<HTMLTextAreaElement>("Body").value).toBe("Canonical body"));
  expect(await listLocalDrafts(draftOwner(userId))).toHaveLength(0);
});

it("retains the copy when publishing is rejected", async () => {
  await seed(); const rejected = vi.fn().mockResolvedValue({ id: 40, moderationStatus: "rejected", content: "Rejected" }); show(rejected);
  await userEvent.click(await screen.findByRole("button", { name: "恢复副本" }));
  await userEvent.click(screen.getByRole("button", { name: "更新主题" }));
  await waitFor(() => expect(rejected).toHaveBeenCalled());
  expect(screen.getByLabelText<HTMLTextAreaElement>("Body").value).toBe("Recovered local body");
  expect((await listLocalDrafts(draftOwner(userId))).some((draft) => draft.content === "Recovered local body")).toBe(true);
});

it("preserves new edits and the local draft after a server request fails", async () => {
  await seed();
  const failed = vi.fn().mockRejectedValue(new Error("offline"));
  show(failed);
  await userEvent.click(await screen.findByRole("button", { name: "恢复副本" }));
  await userEvent.type(screen.getByLabelText("Body"), " with newer input");
  await userEvent.click(screen.getByRole("button", { name: "保存草稿" }));
  await waitFor(() => expect(failed).toHaveBeenCalled());
  await waitFor(() => expect(screen.getByRole<HTMLButtonElement>("button", { name: "保存草稿" }).disabled).toBe(false));
  expect(screen.getByLabelText<HTMLTextAreaElement>("Body").value).toBe("Recovered local body with newer input");
  await waitFor(async () => {
    const drafts = await listLocalDrafts(draftOwner(userId));
    expect(drafts).toHaveLength(1);
    expect(drafts[0].content).toBe("Recovered local body with newer input");
  });
});
