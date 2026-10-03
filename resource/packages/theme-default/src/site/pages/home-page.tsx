import { useEffect, useRef, useState } from "react";
import type { HomeProps, LayoutPayload } from "@gooseforum/client";
import { Bell, ChevronDown, Mail, Plus, UsersRound } from "lucide-react";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@gooseforum/ui/components/collapsible";
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
import { SiteListPanel } from "../layout/site-panel";
import { useAnnouncementDisclosure } from "./use-announcement-disclosure";
import {
  TopicListFooter,
  TopicListToolbar,
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
  const announcement = useAnnouncementDisclosure(announcementVersion);
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
          className="mb-0 rounded-none border-y border-l-2 border-l-primary/45 bg-background px-3 py-1 lg:mb-3 lg:rounded-xl lg:border lg:border-l-2 lg:px-4 lg:py-1.5"
        >
          <Collapsible open={announcement.open} onOpenChange={announcement.setOpen}>
            <div className="flex items-center gap-2 lg:gap-2.5">
              {unread ? (
                <Button
                  variant="ghost"
                  size="icon-sm"
                  className="-mx-1 shrink-0 text-primary hover:bg-primary/10 hover:text-primary"
                  title={t("markRead")}
                  aria-label={t("markRead")}
                  onClick={markRead}
                >
                  <Bell className="announcement-unread-bell size-4" />
                </Button>
              ) : (
                <span className="-mx-1 flex size-7 shrink-0 items-center justify-center text-primary">
                  <Bell aria-hidden="true" className="size-4" />
                </span>
              )}
              <CollapsibleTrigger asChild>
                <button type="button" className="flex min-w-0 flex-1 items-center justify-between gap-2 rounded-sm py-1 text-left text-sm font-medium focus-visible:outline-2 focus-visible:outline-ring"
                  aria-label={announcement.open ? t("collapseAnnouncement") : t("expandAnnouncement")}>
                  {t("announcement")}
                  <ChevronDown className={`size-4 shrink-0 ${announcement.animate ? "transition-transform duration-180 motion-reduce:transition-none" : ""}`} style={{ transform: announcement.open ? "rotate(180deg)" : undefined }} />
                </button>
              </CollapsibleTrigger>
            </div>
            <CollapsibleContent
              className="goose-announcement-content overflow-hidden"
              data-animate={announcement.animate ? "true" : undefined}
              onAnimationEnd={event => {
                if (event.target === event.currentTarget) announcement.finishAnimation();
              }}
            >
              <RenderedContent
                html={page.announcement.html}
                variant="announcement"
                className="min-w-0 pt-1"
              />
            </CollapsibleContent>
          </Collapsible>
        </aside>
      ) : null}
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
