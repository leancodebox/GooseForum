import {
  ExternalLinkIcon,
  LinkIcon,
  SendIcon,
  ShieldCheckIcon,
} from "lucide-react";
import type { LinksPageProps } from "@gooseforum/client";
import { useTranslation } from "react-i18next";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "../../components/ui/empty";
import { GooseLink } from "../../runtime";
import { InfoPanel } from "../layout/info-panel";
import { PageHeader } from "../layout/page-header";

export function LinksPageView({ page }: { page: LinksPageProps }) {
  const { t } = useTranslation("links");
  return (
    <div className="pb-12">
      <PageHeader
        compact
        title={t("title")}
        description={t("subtitle")}
        badge={<Badge variant="secondary">{page.totalCount}</Badge>}
      />
      <div className="grid gap-5 px-2 pt-4 lg:px-0 xl:grid-cols-[minmax(0,1fr)_260px]">
        <div className="flex min-w-0 flex-col gap-6">
          {page.groups.map((group) => (
            <section key={group.name} className="flex flex-col gap-2.5">
              <div className="flex items-center justify-between gap-3 px-1 lg:px-0">
                <h2 className="flex min-w-0 items-center gap-2 text-base font-bold">
                  <span
                    className="flex size-7 shrink-0 items-center justify-center rounded-lg bg-accent text-sm"
                    style={{ color: group.color || undefined }}
                  >
                    {group.emoji || "↗"}
                  </span>
                  <span className="truncate">{group.name}</span>
                </h2>
                <Badge variant="secondary">{group.links.length}</Badge>
              </div>
              <div className="grid grid-cols-2 gap-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
                {group.links.map((link) => (
                  <a
                    key={`${group.name}-${link.url}`}
                    href={link.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="group min-w-0 rounded-xl border bg-background p-2 transition-colors hover:border-primary/30 hover:bg-primary/5"
                  >
                    <div className="flex items-center gap-2">
                      <div className="flex size-8 shrink-0 items-center justify-center overflow-hidden rounded-lg border bg-muted">
                        {link.logoUrl ? (
                          <img
                            src={link.logoUrl}
                            alt={link.name}
                            className="size-full object-cover"
                            loading="lazy"
                          />
                        ) : (
                          <LinkIcon
                            width={16}
                            height={16}
                            className="text-muted-foreground"
                            aria-hidden="true"
                          />
                        )}
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="flex min-w-0 items-center gap-1.5">
                          <h3 className="truncate text-[13px] font-semibold group-hover:text-primary">
                            {link.name}
                          </h3>
                          <ExternalLinkIcon
                            width={12}
                            height={12}
                            className="shrink-0 text-muted-foreground group-hover:text-primary"
                            aria-hidden="true"
                          />
                        </div>
                        <p className="mt-0.5 truncate text-[11px] leading-4 text-muted-foreground">
                          {link.desc || link.url}
                        </p>
                      </div>
                    </div>
                  </a>
                ))}
              </div>
            </section>
          ))}
          {!page.groups.length ? (
            <Empty className="min-h-56 border bg-background">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <LinkIcon />
                </EmptyMedia>
                <EmptyTitle>{t("emptyTitle")}</EmptyTitle>
                <EmptyDescription>{t("emptyDescription")}</EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : null}
        </div>

        <aside className="flex flex-col gap-3">
          <InfoPanel
            title={t("applyTitle")}
            action={
              <Button asChild>
                <GooseLink href="/publish">
                  <SendIcon data-icon="inline-start" />
                  {t("applyAction")}
                </GooseLink>
              </Button>
            }
          >
            {t("applyDescription")}
          </InfoPanel>
          <InfoPanel title={t("principlesTitle")}>
            <ul className="flex flex-col gap-2 text-foreground/75">
              {(["healthy", "relevant", "stable"] as const).map((key) => (
                <li key={key} className="flex gap-2">
                  <ShieldCheckIcon
                    width={16}
                    height={16}
                    className="mt-0.5 shrink-0 text-success"
                    aria-hidden="true"
                  />
                  <span>{t(`principles.${key}`)}</span>
                </li>
              ))}
            </ul>
          </InfoPanel>
        </aside>
      </div>
    </div>
  );
}
