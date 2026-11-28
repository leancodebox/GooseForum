import { Suspense } from "react";
import type { CategoryPageProps } from "@gooseforum/client";
import { Plus, UsersRound } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "@gooseforum/ui/components/button";
import { CategoryIcon } from "@gooseforum/ui/components/category-icon";
import { ListFilter, ListFilterItem } from "@gooseforum/ui/components/list-filter";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import { GooseLink } from "@gooseforum/runtime";
import { PageHeader } from "../layout/page-header";
import { SiteListPanel } from "../layout/site-panel";
import {
  TopicListFooter,
  TopicListToolbar,
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
  const { t } = useTranslation("site.home");
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
          <Suspense fallback={null}>
            <span className="inline-flex items-center gap-2">
              {page.category.icon ? (
                <CategoryIcon icon={page.category.icon} color={page.category.color || undefined} className="max-w-5 overflow-hidden" suspend />
              ) : (
                <span
                  className="size-1.5 rounded-full"
                  style={{ backgroundColor: page.category.color || "var(--primary)" }}
                  aria-hidden="true"
                />
              )}
              <span className="text-xs font-semibold uppercase text-muted-foreground">
                {t("categoryLabel")}
              </span>
            </span>
          </Suspense>
        }
      />
      <SiteListPanel>
        <TopicListToolbar
          action={
            <Button asChild className="shrink-0">
              <GooseLink href="/publish">
                <Plus data-icon="inline-start" />
                {t("newTopic")}
              </GooseLink>
            </Button>
          }
        >
            <ListFilter asChild>
            <nav
              aria-label={t("topic")}
            >
              {page.tabs.map((tab) => (
                <ListFilterItem
                  key={tab.key}
                  asChild
                >
                  <GooseLink
                    href={tab.url}
                    aria-current={tab.active ? "page" : undefined}
                  >
                    {tabLabel(tab.key, tab.label)}
                  </GooseLink>
                </ListFilterItem>
              ))}
            </nav>
            </ListFilter>
            <TopicListModeSwitch
              mode={list.mode}
              t={t}
              onClick={list.switchMode}
            />
        </TopicListToolbar>
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
          pageUrl={pageUrl}
        />
        <div ref={list.sentinel} />
      </SiteListPanel>
    </div>
  );
}
