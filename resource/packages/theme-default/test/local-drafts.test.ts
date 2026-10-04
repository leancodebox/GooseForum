import "fake-indexeddb/auto";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  activateDraftAccount, clearAccountDrafts, deleteLocalDraft,
  draftByteLimit, draftCountLimit, draftGeneration, draftInstanceId, draftOwner,
  draftRetention, listLocalDrafts, writeLocalDraft,
  type DraftValues, type LocalDraft,
} from "../src/site/drafts/local-draft-store";
import { DraftWriter, useLocalDraft } from "../src/site/drafts/use-local-draft";
import { openDB } from "idb";

let userId = 900;
const base: DraftValues = { title: "A topic", content: "Original", categoryIds: [3], replyTargetId: 0, sourceVersion: 1 };
function record(id = draftInstanceId(), revision = 1): LocalDraft {
  return { ...base, id, owner: draftOwner(userId), kind: "edit-topic", objectId: 12, topicId: 12, sourceFingerprint: "original", revision, updatedAt: Date.now(), bytes: 100 };
}
function writer(notify = vi.fn()) {
  return new DraftWriter({ userId, kind: "edit-topic", objectId: 12, topicId: 12, base }, notify);
}
beforeEach(async () => { userId++; await activateDraftAccount(userId); });
afterEach(async () => { cleanup(); vi.useRealTimers(); vi.unstubAllGlobals(); await clearAccountDrafts(draftOwner(userId)); });

describe("browser local draft storage", () => {
  it("keeps distinct branches and accounts, rejects stale revisions", async () => {
    const owner = draftOwner(userId);
    const generation = draftGeneration(owner);
    const first = record();
    const second = { ...record(), objectId: 13 };
    await writeLocalDraft(first, generation);
    await writeLocalDraft(second, generation);
    expect(await writeLocalDraft({ ...first, content: "older", revision: 0 }, generation)).toBe(false);
    expect(await listLocalDrafts(owner)).toHaveLength(2);
    expect(await listLocalDrafts(draftOwner(userId + 100))).toHaveLength(0);
    await deleteLocalDraft(first.id, draftOwner(userId + 100));
    expect(await listLocalDrafts(owner)).toHaveLength(2);
    await writeLocalDraft({ ...first, content: "newer", revision: 2 }, generation);
    await deleteLocalDraft(first.id, owner, 1);
    expect((await listLocalDrafts(owner)).find((item) => item.id === first.id)?.content).toBe("newer");
  });

  it("keeps one latest draft per context and isolates accounts and different objects", async () => {
    const owner = draftOwner(userId), generation = draftGeneration(owner);
    await writeLocalDraft({ ...record(), updatedAt: Date.now() - draftRetention - 1 }, generation);
    expect(await listLocalDrafts(owner)).toHaveLength(0);
    const other = { ...record("other-account"), owner: draftOwner(userId + 100) };
    await writeLocalDraft(other, draftGeneration(other.owner));
    const now = Date.now();
    await writeLocalDraft({ ...record("different-topic"), objectId: 13, updatedAt: now - 2000 }, generation);
    await writeLocalDraft({ ...record("different-kind"), kind: "new-reply", updatedAt: now - 3000 }, generation);
    for (let i = 0; i < 4; i++) await writeLocalDraft({ ...record(`copy-${i}`), updatedAt: now - 1000 + i }, generation);
    expect((await listLocalDrafts(owner)).map((row) => row.id)).toEqual(["copy-3", "different-topic", "different-kind"]);
    expect(await writeLocalDraft({ ...record("late-old-snapshot"), updatedAt: now - 5000 }, generation)).toBe(false);
    expect(await listLocalDrafts(other.owner)).toEqual([other]);
    await clearAccountDrafts(other.owner);
  });

  it("updates a current branch without evicting another retained copy", async () => {
    const owner = draftOwner(userId), generation = draftGeneration(owner);
    const now = Date.now();
    const first = { ...record("first"), updatedAt: now - 300 };
    await writeLocalDraft(first, generation);
    await writeLocalDraft({ ...record("second"), objectId: 13, updatedAt: now - 200 }, generation);
    await writeLocalDraft({ ...record("third"), objectId: 14, updatedAt: now - 100 }, generation);
    await writeLocalDraft({ ...first, revision: 2, content: "Updated", updatedAt: now }, generation);
    expect((await listLocalDrafts(owner)).map((row) => row.id)).toEqual(["first", "third", "second"]);
    expect((await listLocalDrafts(owner))[0].content).toBe("Updated");
    expect(await writeLocalDraft({ ...first, title: "", content: " " }, generation)).toBe(false);
    expect((await listLocalDrafts(owner))[0].content).toBe("Updated");
  });

  it("keeps one latest draft under concurrent writes to the same context", async () => {
    const owner = draftOwner(userId), generation = draftGeneration(owner), now = Date.now();
    const drafts = Array.from({ length: 8 }, (_, i) => ({ ...record(`concurrent-${i}`), updatedAt: now - 1000 + i }));
    await Promise.all(drafts.map((draft) => writeLocalDraft(draft, generation)));
    expect((await listLocalDrafts(owner)).map((row) => row.id)).toEqual(["concurrent-7"]);
  });

  it("compares the expected context owner before replacing an in-flight editor's draft", async () => {
    const owner = draftOwner(userId), generation = draftGeneration(owner);
    const first = record("first-tab");
    await writeLocalDraft(first, generation, { expected: null });
    const second = { ...record("second-tab"), content: "Newer tab", updatedAt: Date.now() };
    expect(await writeLocalDraft(second, generation, { expected: { id: first.id, revision: 1 } })).toBe(true);
    expect(await writeLocalDraft({ ...first, revision: 2, content: "Late first tab", updatedAt: Date.now() + 1 }, generation, { expected: { id: first.id, revision: 1 } })).toBe(false);
    expect(await writeLocalDraft(record("unaware-tab"), generation, { expected: null })).toBe(false);
    expect(await listLocalDrafts(owner)).toEqual([second]);
    expect(await writeLocalDraft({ ...second, revision: 2 }, generation, { expected: { id: second.id, revision: 1 }, canWrite: () => false })).toBe(false);
    expect(await listLocalDrafts(owner)).toEqual([second]);
  });

  it("uses the verified revision when the system clock moves backwards", async () => {
    const owner = draftOwner(userId), generation = draftGeneration(owner);
    const saved = { ...record("future-tab"), updatedAt: Date.now() + 60_000 };
    await writeLocalDraft(saved, generation);
    const resumed = { ...record("resumed-tab"), content: "Edited after clock correction" };
    expect(await writeLocalDraft(resumed, generation, { expected: { id: saved.id, revision: saved.revision } })).toBe(true);
    expect(await listLocalDrafts(owner)).toEqual([resumed]);
  });

  it("preserves different objects when count or byte safety limits reject a write", async () => {
    const owner = draftOwner(userId), generation = draftGeneration(owner);
    for (let i = 0; i < draftCountLimit; i++) await writeLocalDraft({ ...record(`kept-${i}`), objectId: 100 + i }, generation);
    await expect(writeLocalDraft(record(), generation)).rejects.toThrow("capacity");
    await expect(writeLocalDraft({ ...record("kept-0", 2), objectId: 100, bytes: draftByteLimit + 1 }, generation)).rejects.toThrow("capacity");
    expect(await listLocalDrafts(owner)).toHaveLength(draftCountLimit);
  });

  it("rejects whitespace-only drafts, accepts title-only drafts and safely deletes a cleared branch", async () => {
    const owner = draftOwner(userId), generation = draftGeneration(owner);
    const empty = { ...record("empty"), title: " \n", content: "\t ", categoryIds: [99], replyTargetId: 45, sourceVersion: 0 as const };
    await writeLocalDraft(empty, generation);
    expect(await listLocalDrafts(owner)).toHaveLength(0);
    await writeLocalDraft({ ...empty, title: "Only a title", revision: 2 }, generation);
    await writeLocalDraft({ ...record("other-branch"), objectId: 13 }, generation);
    expect(await writeLocalDraft({ ...empty, revision: 1 }, generation)).toBe(false);
    expect(await listLocalDrafts(owner)).toHaveLength(2);
    await writeLocalDraft({ ...empty, revision: 3 }, generation);
    expect((await listLocalDrafts(owner)).map((row) => row.id)).toEqual(["other-branch"]);
  });

  it("cleans historical blank, expired and duplicate drafts when loading", async () => {
    const owner = draftOwner(userId), now = Date.now();
    // Seed records written before the new retention rule directly into the existing database.
    await listLocalDrafts(owner);
    const database = await openDB("gooseforum-local-drafts", 1);
    const tx = database.transaction("drafts", "readwrite");
    for (let i = 0; i < 5; i++) await tx.store.put({ ...record(`old-${i}`), updatedAt: now - 1000 + i });
    await tx.store.put({ ...record("other-object"), objectId: 13, updatedAt: now - 2000 });
    await tx.store.put({ ...record("blank"), title: " ", content: "\n", updatedAt: now });
    await tx.store.put({ ...record("expired"), updatedAt: now - draftRetention - 1 });
    await tx.done;
    expect((await listLocalDrafts(owner)).map((row) => row.id)).toEqual(["old-4", "other-object"]);
    expect(await database.count("drafts")).toBe(2);
    database.close();
  });

  it("invalidates pending writers before clearing and cannot resurrect them after another login", async () => {
    const owner = draftOwner(userId), generation = draftGeneration(owner);
    const pending = writeLocalDraft(record(), generation);
    await clearAccountDrafts(owner);
    await pending;
    expect(await listLocalDrafts(owner)).toHaveLength(0);
    await activateDraftAccount(userId);
    expect(await writeLocalDraft(record(), generation)).toBe(false);
    expect(await writeLocalDraft(record(), draftGeneration(owner))).toBe(true);
  });

  it("uses a branch id on an HTTP deployment without randomUUID", () => {
    vi.stubGlobal("crypto", { getRandomValues: globalThis.crypto.getRandomValues.bind(globalThis.crypto) });
    const ids = new Set(Array.from({ length: 100 }, () => draftInstanceId()));
    expect(ids.size).toBe(100);
    expect(writer().id).toContain(`:edit-topic:12:`);
  });

  it("rejects temporary blob references", async () => {
    await expect(writeLocalDraft({ ...record(), content: "![x](blob:http://localhost/123)" }, draftGeneration(draftOwner(userId)))).rejects.toThrow("attachments");
  });
});

describe("local draft lifecycle", () => {
  it("pauses writes until recovery is decided, then resumes the current input", async () => {
    const saved = record();
    await writeLocalDraft(saved, draftGeneration(saved.owner));
    const draft = writer(); draft.pause();
    draft.update({ ...base, content: "New input" }); await draft.flush();
    expect(await listLocalDrafts(saved.owner)).toEqual([saved]);
    draft.restore(saved);
    draft.resume(); await draft.flush();
    expect(await listLocalDrafts(saved.owner)).toHaveLength(1);
    expect((await listLocalDrafts(saved.owner))[0].content).toBe("New input");
  });

  it("cleans a restored source matching the baseline without deleting a newer revision", async () => {
    const saved = record();
    await writeLocalDraft(saved, draftGeneration(saved.owner));
    const draft = writer(); draft.pause(); draft.restore(saved); draft.resume(); await draft.flush();
    expect(await listLocalDrafts(saved.owner)).toHaveLength(0);
    await writeLocalDraft(saved, draftGeneration(saved.owner));
    const another = writer(); another.pause(); another.restore(saved);
    await writeLocalDraft({ ...saved, revision: 2, content: "Newer" }, draftGeneration(saved.owner));
    another.resume(); await another.flush();
    expect((await listLocalDrafts(saved.owner))[0].content).toBe("Newer");
  });
  it("does not save empty editors when only metadata changes", async () => {
    const empty: DraftValues = { title: "", content: "", categoryIds: [], replyTargetId: 0, sourceVersion: 1 };
    const notify = vi.fn();
    const draft = new DraftWriter({ userId, kind: "new-topic", objectId: 0, topicId: 0, base: empty }, notify);
    for (const values of [{ ...empty, categoryIds: [3] }, { ...empty, replyTargetId: 99 }, { ...empty, sourceVersion: 0 as const }, { ...empty, title: "\t ", content: "\n" }]) {
      draft.update(values); await draft.flush();
      expect(await listLocalDrafts(draftOwner(userId))).toHaveLength(0);
      expect(notify).toHaveBeenLastCalledWith("idle");
    }
  });

  it("removes only its branch on clearing and safely saves text typed immediately afterwards", async () => {
    const notify = vi.fn(), draft = writer(notify);
    await writeLocalDraft({ ...record("separate-branch"), objectId: 13 }, draftGeneration(draftOwner(userId)));
    draft.update({ ...base, content: "Changed" }); await draft.flush();
    draft.update({ ...base, title: "", content: " " }); await draft.flush();
    expect((await listLocalDrafts(draftOwner(userId))).map((row) => row.id)).toEqual(["separate-branch"]);
    expect(notify).toHaveBeenLastCalledWith("idle");
    draft.update({ ...base, title: "Title only", content: "" }); await draft.flush();
    draft.update({ ...base, title: "", content: "" });
    draft.update({ ...base, title: "", content: "New input" }); await draft.flush();
    const copies = await listLocalDrafts(draftOwner(userId));
    expect(copies.find((row) => row.id === draft.id)?.content).toBe("New input");
    expect(copies.some((row) => row.id === "separate-branch")).toBe(true);
  });

  it("debounces writes and saves continuous typing at the maximum wait", async () => {
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });
    const notify = vi.fn(), draft = writer(notify);
    draft.update({ ...base, content: "First" });
    await vi.advanceTimersByTimeAsync(900);
    expect(await listLocalDrafts(draftOwner(userId))).toHaveLength(0);
    for (let i = 0; i < 5; i++) {
      draft.update({ ...base, content: `Typing ${i}` });
      await vi.advanceTimersByTimeAsync(900);
    }
    expect((await listLocalDrafts(draftOwner(userId)))[0].content).toBe("Typing 4");
    expect(notify).toHaveBeenLastCalledWith("saved");
  });

  it("removes a saved branch when edits are undone to the original content", async () => {
    const notify = vi.fn(), draft = writer(notify);
    draft.update({ ...base, content: "Changed" }); await draft.flush();
    expect(await listLocalDrafts(draftOwner(userId))).toHaveLength(1);
    draft.update(base); await draft.flush();
    expect(await listLocalDrafts(draftOwner(userId))).toHaveLength(0);
    expect(notify).toHaveBeenLastCalledWith("idle");
  });

  it("preserves changes typed during a successful submission", async () => {
    const draft = writer();
    const submission = draft.capture({ ...base, content: "Sent" });
    await draft.flush();
    draft.update({ ...base, content: "Typed while sending" }); await draft.flush();
    expect(await draft.submitted(submission, { ...base, content: "Normalized sent" })).toBe(false);
    expect((await listLocalDrafts(draftOwner(userId)))[0].content).toBe("Typed while sending");
  });

  it("clears only a restored revision on success and uses the canonical baseline", async () => {
    const saved = record();
    await writeLocalDraft(saved, draftGeneration(saved.owner));
    const draft = writer(); draft.restore(saved);
    const submission = draft.capture({ ...base, content: "Sent" });
    await draft.flush();
    expect(await writeLocalDraft({ ...saved, revision: 2, content: "Another tab edited", updatedAt: Date.now() }, draftGeneration(saved.owner))).toBe(true);
    expect(await draft.submitted(submission, { ...base, content: "Canonical" })).toBe(true);
    await draft.flush();
    const copies = await listLocalDrafts(saved.owner);
    expect(copies).toHaveLength(1);
    expect(copies[0].content).toBe("Another tab edited");
  });

  it("keeps editing state after storage failure and can retry the same revision", async () => {
    const notify = vi.fn(), draft = writer(notify);
    const database = globalThis.indexedDB;
    vi.stubGlobal("indexedDB", undefined);
    draft.update({ ...base, content: "Still here" }); await draft.flush();
    expect(draft.values.content).toBe("Still here");
    expect(notify).toHaveBeenLastCalledWith("storage");
    vi.stubGlobal("indexedDB", database);
    await draft.flush();
    expect(notify).toHaveBeenLastCalledWith("saved");
    expect((await listLocalDrafts(draftOwner(userId)))[0].content).toBe("Still here");
  });
});

describe("local draft recovery decisions", () => {
  function input(values = base) { return { userId, kind: "edit-topic" as const, objectId: 12, topicId: 12, base, values }; }

  it("keeps autosave paused after a failed load and retries before replacing the old draft", async () => {
    const saved = record(); await writeLocalDraft(saved, draftGeneration(saved.owner));
    const database = globalThis.indexedDB;
    vi.stubGlobal("indexedDB", undefined);
    const { result, rerender } = renderHook((options) => useLocalDraft(options), { initialProps: input() });
    await waitFor(() => expect(result.current.status).toBe("storage"));
    rerender(input({ ...base, content: "New text" }));
    expect(result.current.recoveryPending).toBe(false);
    expect(result.current.writer?.paused).toBe(true);
    vi.stubGlobal("indexedDB", database);
    act(() => result.current.retry());
    await waitFor(() => expect(result.current.candidates).toHaveLength(1));
    expect(await listLocalDrafts(saved.owner)).toEqual([saved]);
    expect(result.current.writer?.paused).toBe(true);
  });

  it("keeps the unique old draft intact before loading and while the prompt is undecided", async () => {
    const saved = record(); await writeLocalDraft(saved, draftGeneration(saved.owner));
    const { result, rerender } = renderHook((options) => useLocalDraft(options), { initialProps: input() });
    expect(result.current.recoveryPending).toBe(true);
    await waitFor(() => expect(result.current.candidates).toHaveLength(1));
    rerender(input({ ...base, content: "Typing before deciding" }));
    await result.current.writer?.flush();
    expect(result.current.writer?.paused).toBe(true);
    expect(await listLocalDrafts(saved.owner)).toEqual([saved]);
    await act(async () => { await result.current.dismiss(); });
    expect(result.current.recoveryPending).toBe(false);
    await result.current.writer?.flush();
    expect((await listLocalDrafts(saved.owner))[0].content).toBe("Typing before deciding");
    expect((await listLocalDrafts(saved.owner))[0].id).not.toBe(saved.id);
  });

  it("restores after React receives the chosen values and resumes automatic saving", async () => {
    const saved = { ...record(), content: "Recovered content" }; await writeLocalDraft(saved, draftGeneration(saved.owner));
    const { result, rerender } = renderHook((options) => useLocalDraft(options), { initialProps: input() });
    await waitFor(() => expect(result.current.candidates).toHaveLength(1));
    act(() => { result.current.restore(saved); rerender(input({ ...base, content: saved.content })); });
    await result.current.writer?.flush();
    const rows = await listLocalDrafts(saved.owner);
    expect(rows).toHaveLength(1);
    expect(rows[0].content).toBe("Recovered content");
    expect(result.current.recoveryPending).toBe(false);
    expect(result.current.writer?.paused).toBe(false);
  });

  it("does not treat recovery-state renders during submitted cleanup as new input", async () => {
    const saved = { ...record(), content: "Recovered content" };
    await writeLocalDraft(saved, draftGeneration(saved.owner));
    const { result, rerender } = renderHook((options) => useLocalDraft(options), { initialProps: input() });
    await waitFor(() => expect(result.current.candidates).toHaveLength(1));
    const recovered = { ...base, content: saved.content };
    act(() => { result.current.restore(saved); rerender(input(recovered)); });
    const draft = result.current.writer!;
    const submission = draft.capture(recovered);
    await draft.flush();
    const canonical = { ...recovered, content: "Canonical body" };
    await act(async () => {
      expect(await draft.submitted(submission, canonical)).toBe(true);
      rerender(input(canonical));
    });
    expect(draft.revision).toBe(submission.revision);
    expect(draft.values.content).toBe("Canonical body");
    expect(await listLocalDrafts(saved.owner)).toHaveLength(0);
  });

  it("discards the selected draft and does not erase a concurrently updated revision", async () => {
    const saved = record(); await writeLocalDraft(saved, draftGeneration(saved.owner));
    const { result } = renderHook(() => useLocalDraft(input()));
    await waitFor(() => expect(result.current.candidates).toHaveLength(1));
    await act(async () => {
      await writeLocalDraft({ ...saved, revision: 2, content: "Updated elsewhere" }, draftGeneration(saved.owner));
      await result.current.discard(saved);
    });
    await waitFor(() => expect(result.current.candidates[0]?.revision).toBe(2));
    expect(result.current.writer?.paused).toBe(true);
    await act(async () => { await result.current.dismiss(); });
    expect(await listLocalDrafts(saved.owner)).toHaveLength(0);
    expect(result.current.recoveryPending).toBe(false);
  });

  it("pauses immediately when another editor saves before checking the new context draft", async () => {
    const { result, rerender } = renderHook((options) => useLocalDraft(options), { initialProps: input() });
    await waitFor(() => expect(result.current.recoveryPending).toBe(false));
    rerender(input({ ...base, content: "Unsubmitted local input" }));
    const remote = { ...record("remote-editor"), content: "Saved in another tab" };
    await act(async () => {
      await writeLocalDraft(remote, draftGeneration(remote.owner));
      expect(result.current.writer?.paused).toBe(true);
      await result.current.writer?.flush();
    });
    await waitFor(() => expect(result.current.candidates[0]?.id).toBe(remote.id));
    expect(result.current.recoveryPending).toBe(true);
    expect(await listLocalDrafts(remote.owner)).toEqual([remote]);
  });
});
