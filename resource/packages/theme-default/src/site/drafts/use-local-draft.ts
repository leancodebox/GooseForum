import { useEffect, useMemo, useRef, useState } from "react";
import {
  deleteLocalDraft, draftFingerprint, draftGeneration, draftInstanceId, draftOwner,
  isEmptyDraft, listenDraftChanges, listLocalDrafts, writeLocalDraft,
  type DraftIdentity, type DraftKind, type DraftValues, type LocalDraft,
} from "./local-draft-store";

export type DraftSaveStatus = "idle" | "saving" | "saved" | "storage" | "capacity" | "attachments";
interface DraftOptions {
  userId: number;
  kind: DraftKind;
  objectId: number;
  topicId: number;
  branch?: number;
  base: DraftValues;
  values: DraftValues;
  enabled?: boolean;
}

export class DraftWriter {
  readonly id: string;
  readonly generation: number;
  revision = 0;
  values: DraftValues;
  baseline: string;
  stopped = false;
  paused = false;
  private debounce?: ReturnType<typeof setTimeout>;
  private maximum?: ReturnType<typeof setTimeout>;
  private writes: Promise<unknown> = Promise.resolve();
  private lastWritten = 0;
  private cleanedRevision = -1;
  private restored?: { id: string; revision: number };
  private expected: DraftIdentity | null = null;
  private committing = false;
  readonly options: Omit<DraftOptions, "values" | "enabled">;
  private notify: (status: DraftSaveStatus) => void;

  constructor(
    options: Omit<DraftOptions, "values" | "enabled">,
    notify: (status: DraftSaveStatus) => void,
  ) {
    this.options = options;
    this.notify = notify;
    this.id = `v1:${draftOwner(options.userId)}:${options.kind}:${options.objectId}:${draftInstanceId()}`;
    this.values = options.base;
    this.baseline = JSON.stringify(options.base);
    this.generation = draftGeneration(draftOwner(options.userId));
  }

  update(values: DraftValues) {
    if (this.stopped || JSON.stringify(values) === JSON.stringify(this.values)) return;
    this.values = values;
    this.revision++;
    if (this.paused) return;
    if (isEmptyDraft(values)) {
      this.setStatus("idle");
      void this.flush();
      return;
    }
    this.setStatus("saving");
    clearTimeout(this.debounce);
    this.debounce = setTimeout(() => { void this.flush(); }, 1000);
    this.maximum ||= setTimeout(() => { void this.flush(); }, 5000);
  }

  private setStatus(status: DraftSaveStatus) { this.notify(status); }
  subscribe(notify: (status: DraftSaveStatus) => void) {
    this.notify = notify;
    return () => { this.notify = () => {}; };
  }

  flush() {
    clearTimeout(this.debounce);
    clearTimeout(this.maximum);
    this.debounce = this.maximum = undefined;
    if (this.stopped || this.paused || this.committing || (!this.revision && !this.restored)) return this.writes;
    const revision = this.revision;
    if (isEmptyDraft(this.values) || JSON.stringify(this.values) === this.baseline) {
      this.writes = this.writes.catch(() => {}).then(async () => {
        if (this.stopped || this.paused) return;
        if (revision <= this.cleanedRevision && !this.restored) return;
        await deleteLocalDraft(this.id, draftOwner(this.options.userId), revision);
        if (this.restored) {
          await deleteLocalDraft(this.restored.id, draftOwner(this.options.userId), this.restored.revision);
          this.restored = undefined;
        }
        this.cleanedRevision = revision;
        this.expected = null;
        if (this.revision === revision) this.setStatus("idle");
      }).catch(() => { this.setStatus("storage"); });
      return this.writes;
    }
    const values = structuredClone(this.values);
    this.writes = this.writes.catch(() => {}).then(async () => {
      if (this.stopped || this.paused || revision <= this.lastWritten) return;
      const draft: LocalDraft = {
        ...values,
        id: this.id,
        owner: draftOwner(this.options.userId),
        kind: this.options.kind,
        objectId: this.options.objectId,
        topicId: this.options.topicId,
        sourceFingerprint: draftFingerprint(this.options.base),
        revision,
        updatedAt: Date.now(),
        bytes: new TextEncoder().encode(JSON.stringify(values)).byteLength,
      };
      try {
        if (await writeLocalDraft(draft, this.generation, { expected: this.expected, canWrite: () => !this.paused && !this.stopped })) {
          this.lastWritten = revision;
          this.expected = { id: this.id, revision };
          if (this.revision === revision) this.setStatus("saved");
        } else if (!this.stopped) { this.pause(); this.setStatus("idle"); }
      } catch (reason) {
        const message = reason instanceof Error ? reason.message : "storage";
        this.setStatus(message === "capacity" || message === "attachments" ? message : "storage");
      }
    });
    return this.writes;
  }

  restore(draft: LocalDraft) {
    this.restored = { id: draft.id, revision: draft.revision };
    this.expected = this.restored;
  }

  expectDraft(draft?: LocalDraft) {
    this.expected = draft ? { id: draft.id, revision: draft.revision } : null;
  }

  pause() {
    this.paused = true;
    clearTimeout(this.debounce); clearTimeout(this.maximum);
    this.debounce = this.maximum = undefined;
  }

  resume() {
    if (this.stopped || !this.paused) return;
    this.paused = false;
    void this.flush();
  }

  capture(values: DraftValues) {
    this.update(values);
    const submission = { revision: this.revision, values: structuredClone(values) };
    void this.flush();
    return submission;
  }

  async submitted(submission: ReturnType<DraftWriter["capture"]>, canonical: DraftValues) {
    this.committing = true;
    const restored = this.restored;
    // Wait for already queued snapshots, then delete only snapshots covered by this submission.
    this.writes = this.writes.catch(() => {}).then(async () => {
      await deleteLocalDraft(this.id, draftOwner(this.options.userId), submission.revision);
      if (restored) {
        await deleteLocalDraft(restored.id, draftOwner(this.options.userId), restored.revision);
        if (this.restored === restored) this.restored = undefined;
      }
      if (this.revision === submission.revision) this.expected = null;
    }).catch(() => {});
    await this.writes;
    this.committing = false;
    this.options.base = structuredClone(canonical);
    this.baseline = JSON.stringify(canonical);
    if (this.revision === submission.revision) {
      clearTimeout(this.debounce); clearTimeout(this.maximum);
      this.debounce = this.maximum = undefined;
      this.values = canonical;
      this.lastWritten = submission.revision;
      this.setStatus("idle");
    } else {
      this.revision++;
      void this.flush();
    }
    return this.revision === submission.revision;
  }

  stop() {
    this.stopped = true;
    clearTimeout(this.debounce); clearTimeout(this.maximum);
  }
}

export function useLocalDraft(options: DraftOptions) {
  const [statusEntry, setStatusEntry] = useState<{ id: string; status: DraftSaveStatus }>({ id: "", status: "idle" });
  const [candidateEntry, setCandidateEntry] = useState<{ id: string; drafts: LocalDraft[]; ready: boolean; failed?: boolean }>({ id: "", drafts: [], ready: false });
  const recoveryEpoch = useRef(0);
  const currentWriter = useRef<DraftWriter | undefined>(undefined);
  const [retryVersion, setRetryVersion] = useState(0);
  const active = Boolean(options.userId && options.enabled !== false);
  const writer = useMemo(() => {
    if (!active) return undefined;
    const instance = new DraftWriter(options, () => {});
    instance.pause();
    return instance;
  },
    // Values change while editing; only context changes create another branch.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [options.userId, options.kind, options.objectId, options.topicId, options.branch, active]);
  const valuesRef = useRef(options.values);
  useEffect(() => {
    if (!writer) return;
    currentWriter.current = writer;
    let disposed = false;
    const setStatus = (status: DraftSaveStatus) => setStatusEntry({ id: writer.id, status });
    const unsubscribe = writer.subscribe(setStatus);
    const owner = draftOwner(options.userId);
    const load = () => {
      writer.pause();
      setCandidateEntry((entry) => ({ id: writer.id, drafts: entry.id === writer.id ? entry.drafts : [], ready: false }));
      const epoch = ++recoveryEpoch.current;
      void listLocalDrafts(owner).then((drafts) => {
        if (disposed || epoch !== recoveryEpoch.current) return;
        const candidates = drafts.filter((draft) => draft.id !== writer.id && draft.kind === options.kind && draft.objectId === options.objectId);
        setCandidateEntry({ id: writer.id, drafts: candidates, ready: true });
        if (candidates.length) { writer.pause(); setStatus("idle"); }
        else { writer.expectDraft(drafts.find((draft) => draft.id === writer.id)); writer.resume(); }
      }).catch(() => {
        if (!disposed && epoch === recoveryEpoch.current) {
          writer.pause();
          setCandidateEntry((entry) => ({ ...entry, id: writer.id, ready: false, failed: true }));
          setStatus("storage");
        }
      });
    };
    load();
    const remove = listenDraftChanges((changedOwner, action, sourceId) => {
      if (owner !== changedOwner) return;
      if (action === 'write' && sourceId === writer.id) return;
      if (action === "clear") { writer.stop(); recoveryEpoch.current++; setCandidateEntry({ id: writer.id, drafts: [], ready: true }); setStatus("idle"); }
      else load();
    });
    const flush = () => { writer.update(valuesRef.current); void writer.flush(); };
    const visibility = () => { if (document.visibilityState === "hidden") flush(); };
    window.addEventListener("pagehide", flush);
    window.addEventListener("blur", flush);
    document.addEventListener("visibilitychange", visibility);
    return () => {
      disposed = true;
      if (currentWriter.current === writer) currentWriter.current = undefined;
      unsubscribe();
      remove();
      window.removeEventListener("pagehide", flush);
      window.removeEventListener("blur", flush);
      document.removeEventListener("visibilitychange", visibility);
      // Its queued writes retain the original account, object and values.
      void writer.flush();
    };
  }, [writer, options.userId, options.kind, options.objectId, retryVersion]);
  useEffect(() => {
    valuesRef.current = options.values;
    writer?.update(options.values);
  }, [writer, options.values]);
  useEffect(() => {
    if (writer && candidateEntry.id === writer.id && candidateEntry.ready && !candidateEntry.drafts.length) writer.resume();
  }, [writer, candidateEntry.id, candidateEntry.ready, candidateEntry.drafts.length]);

  async function discard(draft: LocalDraft) {
    try {
      await deleteLocalDraft(draft.id, draft.owner, draft.revision);
      const remaining = (await listLocalDrafts(draft.owner)).filter((item) => item.id !== writer?.id && item.kind === options.kind && item.objectId === options.objectId);
      if (currentWriter.current !== writer) return;
      recoveryEpoch.current++;
      setCandidateEntry({ id: writer?.id || "", drafts: remaining, ready: true });
      if (!remaining.length) { writer?.expectDraft(); writer?.resume(); }
      else writer?.pause();
    } catch { if (writer && currentWriter.current === writer) setStatusEntry({ id: writer.id, status: "storage" }); }
  }

  async function dismiss() {
    if (!writer || candidateEntry.id !== writer.id || !candidateEntry.ready) return;
    for (const draft of candidateEntry.drafts) await discard(draft);
  }

  const candidates = candidateEntry.id === writer?.id ? candidateEntry.drafts : [];
  const reading = Boolean(writer && (candidateEntry.id !== writer.id || !candidateEntry.ready && !candidateEntry.failed));
  return {
    writer, status: statusEntry.id === writer?.id ? statusEntry.status : "idle" as DraftSaveStatus,
    candidates,
    recoveryPending: reading || candidates.length > 0,
    dismiss,
    discard,
    restore: (draft: LocalDraft) => {
      if (!writer) return;
      recoveryEpoch.current++;
      writer.restore(draft);
      setCandidateEntry({ id: writer.id, drafts: [], ready: true });
    },
    retry: () => {
      writer?.pause();
      if (writer) setCandidateEntry((entry) => ({ ...entry, id: writer.id, ready: false, failed: false }));
      setRetryVersion((version) => version + 1);
    },
  };
}
