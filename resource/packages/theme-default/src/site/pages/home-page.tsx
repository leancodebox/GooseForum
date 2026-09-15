import { useEffect, useRef, useState } from "react";
import type { HomeProps, LayoutPayload } from "@gooseforum/client";
import { Bell, Mail, Plus, UsersRound } from "lucide-react";
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
import { RenderedContent } from "../content/rendered-content";
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
  const announcementVersion = `${page.announcement.enabled}:${page.announcement.publishedAt || ""}`;
  const currentAnnouncementVersion = useRef(announcementVersion);

  useEffect(() => {
    if (currentAnnouncementVersion.current === announcementVersion) return;
    currentAnnouncementVersion.current = announcementVersion;
    setUnread(announcementUnread(page));
  }, [announcementVersion, page]);

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
      {page.announcement.enabled ? (
        <aside
          aria-label={t("announcement")}
          className="mb-0 rounded-none border-y border-l-2 border-l-primary/45 bg-background px-3 py-2 lg:mb-3 lg:rounded-xl lg:border lg:border-l-2 lg:px-4 lg:py-2.5"
        >
          <div className="flex items-start gap-2 lg:gap-2.5">
            {unread ? (
              <Button
                variant="ghost"
                size="icon-sm"
                className="-mx-1 shrink-0 text-primary hover:bg-primary/10 hover:text-primary"
                title={t("markRead")}
                aria-label={t("markRead")}
                onClick={markRead}
              >
                <Bell className="announcement-unread-bell" />
              </Button>
            ) : (
              <Bell
                aria-hidden="true"
                className="mt-1 size-4 shrink-0 text-primary"
              />
            )}
            <RenderedContent
              html={page.announcement.html}
              variant="announcement"
              className="min-w-0 flex-1"
            />
          </div>
        </aside>
      ) : null}
      <section className="overflow-hidden border-b bg-background lg:rounded-xl lg:border">
        <div className="flex flex-col gap-3 border-b px-4 py-3 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <nav className="-m-1 flex min-w-0 gap-2 overflow-x-auto p-1">
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
