import { useEffect, useMemo, useState } from "react";
import { GooseClientError, type GooseSiteApi, type PostPayload, type ReplyTargetPayload } from "@gooseforum/client";
import type { LocalDraft } from "./local-draft-store";

export type DraftTargetStatus = "available" | "checking" | "unavailable" | "error";
type TargetResult = { status: DraftTargetStatus; post?: PostPayload };

export function useDraftReplyTargets({ candidates, posts, references, api, topicId, context, enabled }: {
  candidates: LocalDraft[];
  posts: PostPayload[];
  references: ReplyTargetPayload[];
  api: GooseSiteApi["posts"];
  topicId: number;
  context: string;
  enabled: boolean;
}) {
  const [results, setResults] = useState<{ context: string; items: Map<number, TargetResult> }>({ context: "", items: new Map() });
  const [retry, setRetry] = useState(0);
  const known = useMemo(() => {
    const items = new Map<number, TargetResult>();
    for (const reference of references) items.set(reference.id, { status: reference.unavailable ? "unavailable" : "available" });
    for (const post of posts) items.set(post.id, { status: post.isHidden ? "unavailable" : "available", post });
    return items;
  }, [posts, references]);
  const missingIds = [...new Set(candidates.map((draft) => draft.replyTargetId).filter((id) => id > 0 && !known.has(id)))].sort((a, b) => a - b).join(",");

  useEffect(() => {
    if (!enabled || !missingIds) return;
    let disposed = false;
    const ids = missingIds.split(",").map(Number);
    // Check the retained branches one at a time to bound background requests.
    void (async () => {
      for (const id of ids) {
        if (disposed) return;
        const publish = (result: TargetResult) => {
          if (disposed) return;
          setResults((previous) => ({ context, items: new Map(previous.context === context ? previous.items : []).set(id, result) }));
        };
        publish({ status: "checking" });
        try {
          const window = await api.window({ topicId, anchorPostId: id, limit: 1 });
          const post = window.posts.find((post) => post.id === id && post.topicId === topicId);
          publish(post && !post.isHidden ? { status: "available", post } : { status: "unavailable" });
        } catch (reason) {
          const unavailable = reason instanceof GooseClientError && (reason.status === 403 || reason.status === 404 || reason.messageCode === "post.notFound" || reason.messageCode === "topic.notFound" || reason.messageCode === "permission.denied");
          publish({ status: unavailable ? "unavailable" : "error" });
        }
      }
    })();
    return () => { disposed = true; };
  }, [api, topicId, context, enabled, missingIds, retry]);

  function target(id: number): TargetResult {
    if (!enabled) return { status: "unavailable" };
    if (!id) return { status: "available" };
    return known.get(id) || (results.context === context ? results.items.get(id) : undefined) || { status: "checking" };
  }
  return { target, retry: () => setRetry((value) => value + 1) };
}
