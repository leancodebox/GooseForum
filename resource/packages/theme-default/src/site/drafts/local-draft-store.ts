import { openDB, type DBSchema } from "idb";

export type DraftKind = "new-topic" | "edit-topic" | "new-reply" | "edit-reply";
export interface DraftValues {
  title: string;
  content: string;
  categoryIds: number[];
  replyTargetId: number;
  sourceVersion: 0 | 1;
}
export interface LocalDraft extends DraftValues {
  id: string;
  owner: string;
  kind: DraftKind;
  objectId: number;
  topicId: number;
  sourceFingerprint: string;
  revision: number;
  updatedAt: number;
  bytes: number;
}
export type DraftIdentity = { id: string; revision: number };
interface DraftDatabase extends DBSchema {
  drafts: { key: string; value: LocalDraft; indexes: { owner: string } };
}
export const draftRetention = 14 * 24 * 60 * 60 * 1000;
export const draftCountLimit = 50;
export const draftByteLimit = 10 * 1024 * 1024;
export const draftChangeEvent = "goose:local-draft-change";
const accountKey = "goose:local-draft-account";
const changeKey = "goose:local-draft-change";
const stoppedOwners = new Set<string>();
const generations = new Map<string, number>();
let database: ReturnType<typeof openDB<DraftDatabase>> | undefined;
let accountActivation = 0;
let idSequence = 0;

export function draftInstanceId() {
  if (typeof globalThis.crypto?.randomUUID === "function") return crypto.randomUUID();
  if (typeof globalThis.crypto?.getRandomValues === "function") return [...crypto.getRandomValues(new Uint32Array(4))].map((value) => value.toString(16)).join("-");
  return `${Date.now().toString(36)}-${(++idSequence).toString(36)}-${Math.random().toString(36).slice(2)}-${Math.random().toString(36).slice(2)}`;
}

export function draftOwner(userId: number, origin = window.location.origin) {
  return `${origin}:${userId}`;
}

function db() {
  if (typeof indexedDB === "undefined") return Promise.reject(new Error("storage"));
  database ||= openDB<DraftDatabase>("gooseforum-local-drafts", 1, {
    upgrade(database) {
      database.createObjectStore("drafts", { keyPath: "id" }).createIndex("owner", "owner");
    },
    blocking() { void database?.then((value) => value.close()); database = undefined; },
  }).catch((error: unknown) => { database = undefined; throw error; });
  return database;
}

export function draftGeneration(owner: string) { return generations.get(owner) || 0; }

export function isEmptyDraft(values: Pick<DraftValues, "title" | "content">) {
  return !values.title.trim() && !values.content.trim();
}

export function sameDraftContext(left: LocalDraft, right: LocalDraft) {
  return left.owner === right.owner && left.kind === right.kind && left.objectId === right.objectId;
}

export async function clearAccountDrafts(owner: string) {
  stoppedOwners.add(owner);
  generations.set(owner, draftGeneration(owner) + 1);
  announce(owner, "clear");
  try {
    const database = await db();
    const tx = database.transaction("drafts", "readwrite");
    for (const key of await tx.store.index("owner").getAllKeys(owner)) await tx.store.delete(key);
    await tx.done;
    announce(owner, "delete");
  } catch { /* Editing and logout stay available when browser storage is denied. */ }
}

export async function activateDraftAccount(userId: number) {
  const activation = ++accountActivation;
  const owner = userId ? draftOwner(userId) : "";
  let previous = "";
  try { previous = localStorage.getItem(accountKey) || ""; } catch {}
  if (previous && previous !== owner) await clearAccountDrafts(previous);
  if (activation !== accountActivation) return;
  if (owner) stoppedOwners.delete(owner);
  try { localStorage.setItem(accountKey, owner); } catch {}
}

export function listenDraftChanges(listener: (owner: string, action: string, sourceId?: string) => void) {
  const receive = (value: { owner?: string; action?: string; sourceId?: string }) => {
    if (!value.owner || !value.action) return;
    if (value.action === "clear") {
      stoppedOwners.add(value.owner);
      generations.set(value.owner, draftGeneration(value.owner) + 1);
    }
    listener(value.owner, value.action, value.sourceId);
  };
  const storage = (event: StorageEvent) => {
    if (event.key !== changeKey || !event.newValue) return;
    try {
      const value: unknown = JSON.parse(event.newValue);
      if (value && typeof value === "object" && "owner" in value && "action" in value && typeof value.owner === "string" && typeof value.action === "string") receive({ owner: value.owner, action: value.action, sourceId: 'sourceId' in value && typeof value.sourceId === 'string' ? value.sourceId : undefined });
    } catch {}
  };
  const local = (event: Event) => {
    const detail = (event as CustomEvent<{ owner: string; action: string; sourceId?: string }>).detail;
    // Same-tab clear has already invalidated the owner before broadcasting.
    listener(detail.owner, detail.action, detail.sourceId);
  };
  window.addEventListener("storage", storage);
  window.addEventListener(draftChangeEvent, local);
  return () => {
    window.removeEventListener("storage", storage);
    window.removeEventListener(draftChangeEvent, local);
  };
}

function announce(owner: string, action: string, sourceId?: string) {
  window.dispatchEvent(new CustomEvent(draftChangeEvent, { detail: { owner, action, sourceId } }));
  try { localStorage.setItem(changeKey, JSON.stringify({ owner, action, sourceId, nonce: draftInstanceId() })); } catch {}
}

export async function listLocalDrafts(owner: string) {
  const database = await db();
  const tx = database.transaction("drafts", "readwrite");
  const drafts = await tx.store.index("owner").getAll(owner);
  const active: LocalDraft[] = [];
  for (const draft of drafts) {
    if (isEmptyDraft(draft) || draft.updatedAt < Date.now() - draftRetention) await tx.store.delete(draft.id);
    else active.push(draft);
  }
  active.sort((a, b) => b.updatedAt - a.updatedAt);
  const contexts = new Set<string>();
  const retained: LocalDraft[] = [];
  for (const draft of active) {
    const context = `${draft.kind}:${draft.objectId}`;
    if (contexts.has(context) || retained.length >= draftCountLimit) await tx.store.delete(draft.id);
    else { contexts.add(context); retained.push(draft); }
  }
  await tx.done;
  return retained;
}

export async function writeLocalDraft(draft: LocalDraft, generation: number, options: {
  expected?: DraftIdentity | null;
  canWrite?: () => boolean;
} = {}) {
  if (stoppedOwners.has(draft.owner) || draftGeneration(draft.owner) !== generation) return false;
  // A blob URL cannot be used after this browser document closes.
  if (!isEmptyDraft(draft) && /\bblob:/i.test(draft.content)) throw new Error("attachments");
  const database = await db();
  if (stoppedOwners.has(draft.owner) || draftGeneration(draft.owner) !== generation || options.canWrite?.() === false) return false;
  const tx = database.transaction("drafts", "readwrite");
  const rows = await tx.store.index("owner").getAll(draft.owner);
  if (!isEmptyDraft(draft) && options.expected !== undefined) {
    const current = rows.filter((row) => sameDraftContext(row, draft) && !isEmptyDraft(row) && row.updatedAt >= Date.now() - draftRetention).sort((a, b) => b.updatedAt - a.updatedAt)[0];
    if (options.expected === null ? Boolean(current) : !current || current.id !== options.expected.id || current.revision !== options.expected.revision) {
      await tx.done;
      announce(draft.owner, "conflict");
      return false;
    }
  }
  const previous = rows.find((row) => row.id === draft.id);
  if (previous && previous.revision >= draft.revision) { await tx.done; return false; }
  const active: LocalDraft[] = [];
  const replaced: LocalDraft[] = [];
  for (const row of rows) {
    if (isEmptyDraft(row) || row.updatedAt < Date.now() - draftRetention) await tx.store.delete(row.id);
    else if (sameDraftContext(row, draft)) replaced.push(row);
    else active.push(row);
  }
  if (stoppedOwners.has(draft.owner) || draftGeneration(draft.owner) !== generation || options.canWrite?.() === false) { await tx.done; return false; }
  if (isEmptyDraft(draft)) {
    await tx.store.delete(draft.id);
    await tx.done;
    announce(draft.owner, "delete");
    return true;
  }
  if (options.expected === undefined && replaced.some((row) => row.id !== draft.id && row.updatedAt > draft.updatedAt)) { await tx.done; return false; }
  if (active.length >= draftCountLimit || active.reduce((sum, row) => sum + row.bytes, 0) + draft.bytes > draftByteLimit) {
    await tx.done;
    throw new Error("capacity");
  }
  await tx.store.put(draft);
  for (const row of replaced) if (row.id !== draft.id) await tx.store.delete(row.id);
  await tx.done;
  announce(draft.owner, "write", draft.id);
  return true;
}

export async function deleteLocalDraft(id: string, owner: string, revision?: number) {
  const database = await db();
  const tx = database.transaction("drafts", "readwrite");
  const current = await tx.store.get(id);
  if (current?.owner === owner && (revision === undefined || current.revision <= revision)) await tx.store.delete(id);
  await tx.done;
  announce(owner, "delete");
}

export function draftFingerprint(values: DraftValues) {
  const text = JSON.stringify(values);
  let hash = 2166136261;
  for (let i = 0; i < text.length; i++) hash = Math.imul(hash ^ text.charCodeAt(i), 16777619);
  return (hash >>> 0).toString(16);
}
