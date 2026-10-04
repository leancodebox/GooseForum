import { Check, RotateCcw, Save, Trash2 } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@gooseforum/ui/components/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@gooseforum/ui/components/dialog";
import { GooseLink, useGooseRuntime } from "@gooseforum/runtime";
import type { LocalDraft } from "./local-draft-store";
import type { DraftSaveStatus } from "./use-local-draft";

export function LocalDraftPanel({ status, candidates, onRestore, onDiscard, onRetry, available, conflicts, checking, failed, onRetryTarget, compact = false }: {
  status: DraftSaveStatus;
  candidates: LocalDraft[];
  onRestore(draft: LocalDraft): boolean;
  onDiscard(draft: LocalDraft): Promise<void>;
  onRetry(): void;
  available(draft: LocalDraft): boolean;
  conflicts(draft: LocalDraft): boolean;
  checking?(draft: LocalDraft): boolean;
  failed?(draft: LocalDraft): boolean;
  onRetryTarget?(): void;
  compact?: boolean;
}) {
  const { t } = useTranslation("publish");
  const { locale } = useGooseRuntime();
  const [discarding, setDiscarding] = useState(false);
  const draft = candidates[0];
  const saveFailed = status === "storage" || status === "capacity" || status === "attachments";
  const statusText = status === "idle" ? "" : t(`localDraft.${status}`);
  async function discard() {
    if (!draft || discarding) return;
    setDiscarding(true);
    try { await onDiscard(draft); } finally { setDiscarding(false); }
  }
  return <>
    {statusText ? <div data-slot="local-draft-status" className={`flex shrink-0 items-center gap-1 text-xs ${compact ? "" : "flex-wrap px-4 py-2"}`}>
      {compact ? <span role="status" title={statusText} className={`flex size-7 items-center justify-center ${saveFailed ? "text-destructive" : "text-muted-foreground"}`}>
        {status === "saved" ? <Check className="size-4" /> : <Save className="size-4" />}
        <span className="sr-only">{statusText}</span>
      </span> : <span role="status" className={saveFailed ? "text-destructive" : "text-muted-foreground"}>{statusText}</span>}
      {saveFailed ? <Button size="icon-sm" variant="ghost" title={t("localDraft.retry")} aria-label={t("localDraft.retry")} onClick={onRetry}><RotateCcw /></Button> : null}
      {!compact && status === "capacity" ? <GooseLink className="text-primary underline" href="/drafts">{t("localDraft.manage")}</GooseLink> : null}
    </div> : null}
    <Dialog open={Boolean(draft)}>
      <DialogContent showCloseButton={false} className="rounded-lg" onInteractOutside={(event) => event.preventDefault()} onEscapeKeyDown={(event) => event.preventDefault()}>
        <DialogHeader>
          <DialogTitle>{t("localDraft.recoveryTitle")}</DialogTitle>
          <DialogDescription>{t("localDraft.recoveryDescription")}</DialogDescription>
        </DialogHeader>
        {draft ? <div className="min-w-0 text-sm">
          <time className="text-xs text-muted-foreground" dateTime={new Date(draft.updatedAt).toISOString()}>{new Intl.DateTimeFormat(locale, { dateStyle: "short", timeStyle: "short" }).format(draft.updatedAt)}</time>
          {available(draft) ? <p className="mt-2 line-clamp-3 break-words">{draft.title || draft.content.slice(0, 160)}</p> : checking?.(draft) ? <p role="status" className="mt-2 text-muted-foreground">{t("loadingMoreReplies", { ns: "topic" })}</p> : <p className="mt-2 text-destructive">{failed?.(draft) ? t("loadFailed", { ns: "topic" }) : t("localDraft.unavailable")}</p>}
          {available(draft) && conflicts(draft) ? <p className="mt-2 text-destructive">{t("localDraft.conflict")}</p> : null}
          {saveFailed ? <p role="status" className="mt-2 text-destructive">{statusText}</p> : null}
        </div> : null}
        <DialogFooter className="flex-wrap rounded-b-lg">
          {draft ? <>
            {failed?.(draft) ? <Button variant="ghost" disabled={discarding} onClick={onRetryTarget}><RotateCcw />{t("localDraft.retry")}</Button> : null}
            <Button variant="outline" disabled={discarding} onClick={() => void discard()}><Trash2 />{t("localDraft.discard")}</Button>
            <Button disabled={discarding || !available(draft)} onClick={() => onRestore(draft)}><RotateCcw />{t("localDraft.restore")}</Button>
          </> : null}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </>;
}
