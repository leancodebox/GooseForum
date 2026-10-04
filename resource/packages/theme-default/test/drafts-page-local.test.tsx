import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import type { DraftsPageProps, GooseSiteApi, LayoutPayload } from "@gooseforum/client";
import { GooseRuntimeProvider, type GooseRuntime } from "@gooseforum/runtime";
import { GooseI18nProvider } from "@gooseforum/runtime/i18n";
import { DraftsPageView } from "../src/site/pages/drafts-page";
import { draftOwner, type LocalDraft } from "../src/site/drafts/local-draft-store";

const storage = vi.hoisted(() => ({ list: vi.fn(), remove: vi.fn(), listen: vi.fn() }));
vi.mock("../src/site/drafts/local-draft-store", async (original) => ({
  ...await original<typeof import("../src/site/drafts/local-draft-store")>(),
  listLocalDrafts: storage.list,
  deleteLocalDraft: storage.remove,
  listenDraftChanges: storage.listen,
}));

let change: (owner: string, action: string) => void;
const page: DraftsPageProps = { total: 0, drafts: [], pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" } };
const runtime = { api: {} as GooseSiteApi, currentUrl: "/drafts", locale: "zh", theme: "gf-light", isNavigating: false, navigate: vi.fn(), redirect: vi.fn(), refresh: vi.fn(), queueFlash: vi.fn(), setLocale: vi.fn(), toggleTheme: vi.fn() } satisfies GooseRuntime;
function tree(userId = 7) {
  return <GooseI18nProvider locale="zh"><GooseRuntimeProvider runtime={runtime}><DraftsPageView page={page} layout={{ viewer: { id: userId } } as LayoutPayload} /></GooseRuntimeProvider></GooseI18nProvider>;
}
function draft(overrides: Partial<LocalDraft> = {}): LocalDraft {
  return { id: "local-copy", owner: draftOwner(7), kind: "edit-topic", objectId: 40, topicId: 40, title: "Local topic", content: "Local content", categoryIds: [], replyTargetId: 0, sourceVersion: 1, sourceFingerprint: "base", revision: 1, updatedAt: Date.now(), bytes: 100, ...overrides };
}
function deferred() {
  let resolve!: (copies: LocalDraft[]) => void;
  const promise = new Promise<LocalDraft[]>((done) => { resolve = done; });
  return { promise, resolve };
}
const localLink = () => screen.queryByRole("link", { name: "继续编辑" });

beforeEach(() => {
  storage.list.mockReset(); storage.remove.mockReset(); storage.listen.mockReset();
  storage.listen.mockImplementation((listener) => { change = listener; return () => {}; });
});
afterEach(cleanup);

it("hides the previous account's drafts before the next account finishes loading", async () => {
  const nextAccount = deferred();
  storage.list.mockResolvedValueOnce([draft()]).mockReturnValueOnce(nextAccount.promise);
  const { rerender } = render(tree());
  await screen.findByRole("link", { name: "继续编辑" });
  rerender(tree(8));
  expect(localLink()).toBeNull();
  await act(async () => { nextAccount.resolve([]); await nextAccount.promise; });
  expect(localLink()).toBeNull();
});

it("does not re-display cleared account drafts from an outstanding load", async () => {
  const pending = deferred();
  storage.list.mockResolvedValueOnce([draft()]).mockReturnValueOnce(pending.promise);
  render(tree());
  await screen.findByRole("link", { name: "继续编辑" });
  act(() => change(draftOwner(7), "write"));
  act(() => change(draftOwner(7), "clear"));
  expect(localLink()).toBeNull();
  await act(async () => { pending.resolve([draft()]); await pending.promise; });
  expect(localLink()).toBeNull();
});

it("keeps the newer load when change notifications overlap", async () => {
  const older = deferred(), newer = deferred();
  storage.list.mockResolvedValueOnce([draft()]).mockReturnValueOnce(older.promise).mockReturnValueOnce(newer.promise);
  render(tree());
  await screen.findByRole("link", { name: "继续编辑" });
  act(() => { change(draftOwner(7), "write"); change(draftOwner(7), "write"); });
  await act(async () => { newer.resolve([draft({ objectId: 42 })]); await newer.promise; });
  expect(localLink()?.getAttribute("href")).toBe("/publish?id=42");
  await act(async () => { older.resolve([draft({ objectId: 41 })]); await older.promise; });
  expect(localLink()?.getAttribute("href")).toBe("/publish?id=42");
});

it("retains a concurrently updated revision after discarding the displayed revision", async () => {
  const saved = draft(), newer = { ...saved, revision: 2, updatedAt: saved.updatedAt + 60_000 };
  storage.list.mockResolvedValueOnce([saved]).mockResolvedValue([newer]);
  storage.remove.mockImplementation(() => { change(saved.owner, "delete"); return Promise.resolve(); });
  const { container } = render(tree());
  await screen.findByRole("link", { name: "继续编辑" });
  await userEvent.click(screen.getByRole("button", { name: "删除副本" }));
  expect(storage.remove).toHaveBeenCalledWith(saved.id, saved.owner, saved.revision);
  await waitFor(() => expect(container.querySelector("time")?.dateTime).toBe(new Date(newer.updatedAt).toISOString()));
  expect(localLink()).toBeTruthy();
});
