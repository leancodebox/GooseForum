import type { CategoryPageProps } from "@gooseforum/client";
import { Plus, UsersRound } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "@gooseforum/ui/components/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import { GooseLink } from "@gooseforum/runtime";
import { PageHeader } from "../layout/page-header";
import {
  TopicListFooter,
  TopicListModeSwitch,
  TopicTable,
  useTopicList,
} from "../topics/topic-list";

export function CategoryPageView({
  page,
  pageUrl,
}: {
  page: CategoryPageProps;
  pageUrl: string;
}) {
  const { t } = useTranslation("home");
  const list = useTopicList(page, pageUrl);

  function tabLabel(key: string, fallback?: string) {
    if (key === "latest") return t("latestReplies");
    if (key === "new") return t("latestPublished");
    if (key === "hot") return t("hot");
    if (key === "popular") return t("popular");
    return fallback || key;
  }

  return (
    <div className="pb-12">
      <PageHeader
        compact
        title={page.category.name}
        description={page.category.description}
        badge={
          <span className="inline-flex items-center gap-2">
            <span
              className="size-1.5 rounded-full"
              style={{
                backgroundColor: page.category.color || "var(--primary)",
              }}
            />
            <span className="text-xs font-semibold uppercase text-muted-foreground">
              {t("categoryLabel")}
            </span>
          </span>
        }
      />
      <section className="overflow-hidden border-b bg-background lg:rounded-xl lg:border">
        <div className="flex flex-col gap-3 border-b px-4 py-3 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <nav
              className="flex min-w-0 gap-2 overflow-x-auto"
              aria-label={t("topic")}
            >
              {page.tabs.map((tab) => (
                <Button
                  key={tab.key}
                  asChild
                  variant={tab.active ? "default" : "secondary"}
                  size="sm"
                >
                  <GooseLink
                    href={tab.url}
                    aria-current={tab.active ? "page" : undefined}
                  >
                    {tabLabel(tab.key, tab.label)}
                  </GooseLink>
                </Button>
              ))}
            </nav>
            <TopicListModeSwitch
              mode={list.mode}
              t={t}
              onClick={list.switchMode}
            />
          </div>
          <Button asChild className="shrink-0">
            <GooseLink href="/publish">
              <Plus data-icon="inline-start" />
              {t("newTopic")}
            </GooseLink>
          </Button>
        </div>
        <TopicTable
          topics={list.topics}
          showCategories={false}
          showHot={false}
          t={t}
        />
        {!list.topics.length ? (
          <Empty className="min-h-48 border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <UsersRound />
              </EmptyMedia>
              <EmptyTitle>{t("emptyTitle")}</EmptyTitle>
              <EmptyDescription>
                {t("emptyCategoryDescription")}
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : null}
        <TopicListFooter
          mode={list.mode}
          pagination={list.pagination}
          loading={list.loading}
          error={list.error}
          t={t}
          onLoad={list.loadMore}
        />
        <div ref={list.sentinel} />
      </section>
    </div>
  );
}
