import { useEffect, useState } from "react";
import type { DraftPayload, DraftsPageProps, LayoutPayload } from "@gooseforum/client";
import { FileText, PenSquare, ShieldAlert, Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Badge } from "@gooseforum/ui/components/badge";
import { Button } from "@gooseforum/ui/components/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import { GooseLink, useGooseRuntime } from "@gooseforum/runtime";
import { PageHeader } from "../layout/page-header";
import { SiteListPanel } from "../layout/site-panel";
import { deleteLocalDraft, draftOwner, listenDraftChanges, listLocalDrafts, type LocalDraft } from "../drafts/local-draft-store";

export function DraftsPageView({ page, layout }: { page: DraftsPageProps; layout?: LayoutPayload }) {
  const { t } = useTranslation("drafts");
  const { t: publishT } = useTranslation("publish");
  const { locale } = useGooseRuntime();
  const [localEntry, setLocalEntry] = useState<{ owner: string; copies: LocalDraft[]; error: boolean }>({ owner: "", copies: [], error: false });
  const userId = layout?.viewer.id || 0;
  const owner = userId ? draftOwner(userId) : "";
  const localCopies = localEntry.owner === owner ? localEntry.copies : [];
  const localError = localEntry.owner === owner && localEntry.error;
  useEffect(() => {
    if (!owner) return;
    let disposed = false;
    let epoch = 0;
    const load = () => {
      const request = ++epoch;
      void listLocalDrafts(owner).then((copies) => {
        if (!disposed && request === epoch) setLocalEntry({ owner, copies, error: false });
      }).catch(() => {
        if (!disposed && request === epoch) setLocalEntry((entry) => ({ owner, copies: entry.owner === owner ? entry.copies : [], error: true }));
      });
    };
    load();
    const remove = listenDraftChanges((changedOwner, action) => {
      if (changedOwner !== owner) return;
      if (action === "clear") { epoch++; setLocalEntry({ owner, copies: [], error: false }); }
      else load();
    });
    return () => { disposed = true; remove(); };
  }, [owner]);
  async function discardLocal(draft: LocalDraft) {
    // The change listener reloads the retained revision, including concurrent edits.
    try { await deleteLocalDraft(draft.id, draft.owner, draft.revision); }
    catch { setLocalEntry((entry) => entry.owner === draft.owner ? { ...entry, error: true } : entry); }
  }

  return (
    <main className="min-w-0 pb-8">
      <PageHeader
        compact
        title={t("title")}
        description={t("summary")}
        badge={
          <Badge variant="secondary">{t("total", { count: page.total })}</Badge>
        }
        actions={
          <Button asChild variant="outline" size="sm">
            <GooseLink href="/publish">
              <PenSquare data-icon="inline-start" />
              {t("newDraft")}
            </GooseLink>
          </Button>
        }
      />
      <SiteListPanel>
        <div className="hidden grid-cols-[minmax(0,1fr)_152px_132px] gap-4 border-b bg-muted/50 px-4 py-2 text-[11px] font-bold uppercase text-muted-foreground lg:grid">
          <span>{t("table.draft")}</span>
          <span className="text-right">{t("table.updatedAt")}</span>
          <span className="text-right">{t("table.action")}</span>
        </div>
        {page.drafts.length ? (
          <div className="divide-y">
            {page.drafts.map((draft) => (
              <DraftRow key={draft.id} draft={draft} locale={locale} t={t} />
            ))}
          </div>
        ) : (
          <Empty className="min-h-56 border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <FileText />
              </EmptyMedia>
              <EmptyTitle>{t("emptyTitle")}</EmptyTitle>
              <EmptyDescription>{t("emptyHint")}</EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Button asChild>
                <GooseLink href="/publish">{t("newDraft")}</GooseLink>
              </Button>
            </EmptyContent>
          </Empty>
        )}
      </SiteListPanel>
      {userId ? <section className="mt-6 border-t">
        <h2 className="px-4 py-3 text-sm font-semibold">{publishT("localDraft.localTitle")}</h2>
        {localError ? <p role="status" className="px-4 py-2 text-sm text-destructive">{publishT("localDraft.storage")}</p> : null}
        {localCopies.length ? <div className="divide-y">
          {localCopies.map((draft) => <div key={draft.id} className="flex flex-wrap items-center justify-between gap-3 px-4 py-3">
            <div className="min-w-0"><span className="text-sm font-medium">{publishT(`localDraft.${draft.kind === "new-topic" ? "newTopic" : draft.kind === "edit-topic" ? "editTopic" : draft.kind === "new-reply" ? "newReply" : "editReply"}`)}</span><time className="ml-3 text-xs text-muted-foreground" dateTime={new Date(draft.updatedAt).toISOString()}>{formatDate(new Date(draft.updatedAt).toISOString(), locale)}</time></div>
            <div className="flex items-center gap-1"><Button asChild size="sm" variant="outline"><GooseLink href={draft.kind === "new-topic" ? "/publish" : draft.kind === "edit-topic" ? `/publish?id=${draft.objectId}` : `/p/post/${draft.topicId}?${draft.kind === "edit-reply" ? `localEditPost=${draft.objectId}` : "localReply=1"}`}><PenSquare />{t("edit")}</GooseLink></Button><Button size="icon-sm" variant="ghost" aria-label={publishT("localDraft.delete")} onClick={() => void discardLocal(draft)}><Trash2 /></Button></div>
          </div>)}
        </div> : <p className="px-4 py-3 text-sm text-muted-foreground">{publishT("localDraft.empty")}</p>}
      </section> : null}
    </main>
  );
}

function DraftRow({
  draft,
  locale,
  t,
}: {
  draft: DraftPayload;
  locale: string;
  t: ReturnType<typeof useTranslation>["t"];
}) {
  return (
    <article className="grid gap-3 px-4 py-3 transition-colors hover:bg-muted/50 lg:grid-cols-[minmax(0,1fr)_152px_132px] lg:items-center lg:gap-4">
      <div className="min-w-0">
        <div className="flex min-w-0 flex-wrap items-center gap-2">
          <GooseLink
            href={draft.editUrl}
            className="min-w-0 max-w-full truncate font-semibold hover:text-primary"
          >
            {draft.title || t("untitled")}
          </GooseLink>
          {draft.categories.map((category) => (
            <Badge
              key={category.id}
              variant="outline"
              className="gap-1 font-medium"
            >
              <span
                className="size-1.5 rounded-full"
                style={{ backgroundColor: category.color }}
              />
              {category.name}
            </Badge>
          ))}
          {draft.processStatus === 1 ? (
            <Badge variant="destructive">
              <ShieldAlert data-icon="inline-start" />
              {t("blocked")}
            </Badge>
          ) : null}
        </div>
        <p className="mt-1.5 line-clamp-1 text-sm leading-6 text-muted-foreground">
          {draft.description || t("emptyDescription")}
        </p>
        <div className="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
          <span>
            {t("meta.createdAt")} {formatDate(draft.createdAt, locale)}
          </span>
          <span>
            {t("meta.views")} {draft.viewCount}
          </span>
          <span>
            {t("meta.replies")} {draft.replyCount}
          </span>
        </div>
      </div>
      <time
        className="text-xs font-medium text-muted-foreground lg:text-right"
        dateTime={draft.updatedAt}
      >
        {formatDate(draft.updatedAt, locale)}
      </time>
      <div className="flex lg:justify-end">
        <Button asChild size="sm">
          <GooseLink href={draft.editUrl}>
            <FileText data-icon="inline-start" />
            {t("edit")}
          </GooseLink>
        </Button>
      </div>
    </article>
  );
}

function formatDate(value: string, locale: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : new Intl.DateTimeFormat(locale, {
        dateStyle: "medium",
        timeStyle: "short",
      }).format(date);
}
