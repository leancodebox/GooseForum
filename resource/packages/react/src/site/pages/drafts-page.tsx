import type { DraftPayload, DraftsPageProps } from "@gooseforum/client";
import { FileText, PenSquare, ShieldAlert } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "../../components/ui/empty";
import { GooseLink, useGooseRuntime } from "../../runtime";
import { PageHeader } from "../layout/page-header";

export function DraftsPageView({ page }: { page: DraftsPageProps }) {
  const { t } = useTranslation("drafts");
  const { locale } = useGooseRuntime();

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
      <section className="overflow-hidden rounded-xl border bg-background max-sm:rounded-t-none max-sm:border-t-0">
        <div className="hidden grid-cols-[minmax(0,1fr)_152px_132px] gap-4 border-b bg-muted/50 px-4 py-2 text-[11px] font-bold uppercase text-muted-foreground md:grid">
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
      </section>
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
    <article className="grid gap-3 px-4 py-3 transition-colors hover:bg-muted/50 md:grid-cols-[minmax(0,1fr)_152px_132px] md:items-center md:gap-4">
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
        className="text-xs font-medium text-muted-foreground md:text-right"
        dateTime={draft.updatedAt}
      >
        {formatDate(draft.updatedAt, locale)}
      </time>
      <div className="flex md:justify-end">
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
