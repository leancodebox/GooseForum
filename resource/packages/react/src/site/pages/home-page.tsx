import { useEffect, useState } from "react";
import type { HomeProps, LayoutPayload } from "@gooseforum/client";
import { Bell, Mail, Plus, UsersRound } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "../../components/ui/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "../../components/ui/empty";
import { GooseLink } from "../../runtime";
import {
  TopicListFooter,
  TopicListModeSwitch,
  TopicTable,
  useTopicList,
} from "../topics/topic-list";

const readKey = "goose:announcement:last-read-published-at";

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
  const [unread, setUnread] = useState(() => announcementUnread(page));

  useEffect(() => setUnread(announcementUnread(page)), [page, pageUrl]);

  function markRead() {
    setUnread(false);
    const time = parseTime(page.announcement.publishedAt);
    if (Number.isFinite(time)) {
      try {
        localStorage.setItem(readKey, String(time));
      } catch {}
    }
  }

  return (
    <div className="pb-12">
      {layout.viewer.requiresEmailVerification ? (
        <aside className="border-y border-warning/30 bg-warning/10 sm:-mt-3 sm:mb-3 sm:rounded-b-lg sm:border-x sm:border-t-0">
          <div className="flex items-center gap-2 px-3 py-2 text-[13px] text-warning sm:px-4 sm:text-sm">
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
      {page.announcement.enabled ? (
        <aside className="border-l-2 border-l-primary/45 bg-background px-3 py-2 sm:mb-3 sm:px-4 sm:py-2.5">
          <div className="flex items-start gap-2">
            {unread ? (
              <Button
                variant="ghost"
                size="icon-sm"
                className="-mx-1 text-primary"
                title={t("markRead")}
                onClick={markRead}
              >
                <Bell />
              </Button>
            ) : (
              <Bell className="mt-1 size-4 text-primary" />
            )}
            <div
              className="min-w-0 flex-1 text-sm leading-6"
              dangerouslySetInnerHTML={{ __html: page.announcement.html }}
            />
          </div>
        </aside>
      ) : null}
      <section className="overflow-hidden rounded-xl border bg-background">
        <div className="flex flex-col gap-3 border-b px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <nav className="flex min-w-0 gap-2 overflow-x-auto">
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
                    {tab.key === "latest"
                      ? t("latest")
                      : tab.key === "hot"
                        ? t("hot")
                        : tab.key === "popular"
                          ? t("popular")
                          : tab.label || tab.key}
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
      </section>
    </div>
  );
}

function parseTime(value?: string) {
  return Date.parse((value || "").replace(" ", "T"));
}

function announcementUnread(page: HomeProps) {
  if (!page.announcement.enabled) return false;
  const time = parseTime(page.announcement.publishedAt);
  const age = Date.now() - time;
  if (!Number.isFinite(time) || age < 0 || age > 604800000) return false;
  try {
    return Number(localStorage.getItem(readKey) || 0) < time;
  } catch {
    return true;
  }
}
