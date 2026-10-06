import type { HomeProps, LayoutPayload } from "@gooseforum/client";
import { Mail, Plus, UsersRound } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "@gooseforum/ui/components/button";
import { ListFilter, ListFilterItem } from "@gooseforum/ui/components/list-filter";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import { GooseLink } from "@gooseforum/runtime";
import { AnnouncementPanel } from "../content/announcement-panel";
import { SiteListPanel } from "../layout/site-panel";
import {
  TopicListFooter,
  TopicListToolbar,
  TopicListModeSwitch,
  TopicTable,
  useTopicList,
} from "../topics/topic-list";

export function HomePageView({
  layout,
  page,
  pageUrl,
}: {
  layout: LayoutPayload;
  page: HomeProps;
  pageUrl: string;
}) {
  const { t } = useTranslation("home");
  const list = useTopicList(page, pageUrl);
  return (
    <div className="pb-12">
      {layout.viewer.requiresEmailVerification ? (
        <aside className="border-y border-warning/30 bg-warning/10 lg:-mt-3 lg:mb-3 lg:rounded-b-lg lg:border-x lg:border-t-0">
          <div className="flex items-center gap-2 px-3 py-2 text-[13px] text-warning lg:px-4 lg:text-sm">
            <Mail className="size-4" />
            <div className="flex-1">
              <strong>{t("emailTitle")}</strong> · {t("emailDescription")}
            </div>
            <GooseLink href="/settings" className="font-semibold">
              {t("emailAction")}
            </GooseLink>
          </div>
        </aside>
      ) : null}
      <AnnouncementPanel announcement={page.announcement} />
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
            <nav>
              {page.tabs.map((tab) => (
                <ListFilterItem
                  key={tab.key}
                  asChild
                >
                  <GooseLink
                    href={tab.url}
                    aria-current={tab.active ? "page" : undefined}
                  >
                    {tab.key === "latest"
                      ? t("latest")
                      : tab.key === "hot"
                        ? t("hot")
                        : tab.key === "popular"
                          ? t("popular")
                          : tab.label || tab.key}
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
          showPinned={page.sort === "" || page.sort === "latest"}
          t={t}
        />
        {!list.topics.length ? (
          <Empty className="min-h-48 border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <UsersRound />
              </EmptyMedia>
              <EmptyTitle>{t("emptyTitle")}</EmptyTitle>
              <EmptyDescription>{t("emptyDescription")}</EmptyDescription>
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
      </SiteListPanel>
    </div>
  );
}
