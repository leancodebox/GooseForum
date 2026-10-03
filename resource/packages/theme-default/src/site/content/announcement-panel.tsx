import { useMemo, useState } from "react";
import type { HomeProps } from "@gooseforum/client";
import { Bell, ChevronDown } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "@gooseforum/ui/components/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@gooseforum/ui/components/collapsible";
import { useAnnouncementDisclosure } from "../pages/use-announcement-disclosure";
import { RenderedContent } from "./rendered-content";

const readKey = "goose:announcement:last-read-published-at";

function isUnread(publishedAt?: number | string) {
  const time = typeof publishedAt === "number" ? publishedAt : Date.parse((publishedAt || "").replace(" ", "T"));
  const age = Date.now() - time;
  if (!Number.isFinite(time) || age < 0 || age > 604800000) return false;
  try {
    const saved = localStorage.getItem(readKey);
    if (!saved) return true;
    if (typeof publishedAt === "number") {
      const readTime = /^\d+$/.test(saved) ? Number(saved) : Date.parse(saved.replace(" ", "T"));
      return !Number.isFinite(readTime) || readTime < time;
    }
    // Compatibility with payloads and read markers from earlier versions.
    return /^\d+$/.test(saved) ? Number(saved) < time : saved !== publishedAt;
  } catch { return true; }
}

export function AnnouncementPanel({ announcement }: { announcement: HomeProps["announcement"] }) {
  const { t } = useTranslation("home");
  const items = useMemo(() => announcement.items ?? (announcement.html.trim()
    ? [{ id: "legacy", title: "", html: announcement.html }] : []), [announcement.items, announcement.html]);
  const version = `${announcement.enabled}:${announcement.publishedAt || ""}`;
  const disclosure = useAnnouncementDisclosure(version);
  const [selection, setSelection] = useState({ version, index: 0 });
  const [read, setRead] = useState(() => ({ version, unread: isUnread(announcement.publishedAt) }));
  if (selection.version !== version) setSelection({ version, index: 0 });
  if (read.version !== version) setRead({ version, unread: isUnread(announcement.publishedAt) });
  const index = Math.min(selection.index, Math.max(0, items.length - 1));
  const active = items[index];
  const multiple = items.length > 1;
  if (!announcement.enabled || !active) return null;
  function select(next: number) { setSelection({ version, index: (next + items.length) % items.length }); }
  function markRead() {
    setRead({ version, unread: false });
    if (announcement.publishedAt) {
      try { localStorage.setItem(readKey, String(announcement.publishedAt)); } catch {}
    }
  }
  return <aside aria-label={t("announcement")}
    className="mb-0 rounded-none border-y border-l-2 border-l-primary/45 bg-background px-3 py-1 lg:mb-3 lg:rounded-xl lg:border lg:border-l-2 lg:px-4 lg:py-1.5">
    <Collapsible open={disclosure.open} onOpenChange={disclosure.setOpen}>
      <div className="relative flex items-center gap-2 lg:gap-2.5">
        {read.unread ? <Button variant="ghost" size="icon-sm" className="-mx-1 shrink-0 text-primary hover:bg-primary/10 hover:text-primary"
          title={t("markRead")} aria-label={t("markRead")} onClick={markRead}><Bell className="announcement-unread-bell size-4" /></Button>
          : <span className="-mx-1 flex size-7 shrink-0 items-center justify-center text-primary"><Bell aria-hidden="true" className="size-4" /></span>}
        <CollapsibleTrigger asChild>
          <button type="button" className={`flex min-w-0 flex-1 items-center gap-2 rounded-sm py-1 text-left text-sm focus-visible:outline-2 focus-visible:outline-ring ${multiple && disclosure.open ? "" : "pr-6"}`}
            aria-label={disclosure.open ? t("collapseAnnouncement") : t("expandAnnouncement")}>
            <span className="shrink-0 font-medium">{t("announcement")}</span>
            {active.title ? <span className="truncate text-xs text-muted-foreground">{active.title}</span> : null}
            <ChevronDown className={`absolute right-0 top-1/2 size-4 -translate-y-1/2 ${disclosure.animate ? "transition-transform duration-180 motion-reduce:transition-none" : ""}`}
              style={{ transform: disclosure.open ? "rotate(180deg)" : undefined }} />
          </button>
        </CollapsibleTrigger>
        {multiple && disclosure.open ? <div className="mr-6 flex max-w-24 shrink-0 gap-0.5 overflow-x-auto sm:max-w-40" aria-label={t("announcementSelect")}>
            {items.map((item, i) => <button type="button" key={item.id} aria-label={t("announcementItem", { index: i + 1, title: item.title || t("announcement") })}
              aria-pressed={i === index} onClick={() => select(i)} className="flex size-6 shrink-0 items-center justify-center rounded-sm focus-visible:outline-2 focus-visible:outline-ring">
              <span className={`h-1.5 rounded-full ${i === index ? "w-4 bg-primary" : "w-1.5 bg-muted-foreground/40"}`} />
            </button>)}
          </div> : null}
      </div>
      <CollapsibleContent className="goose-announcement-content overflow-hidden" data-animate={disclosure.animate ? "true" : undefined}
        onAnimationEnd={event => { if (event.target === event.currentTarget) disclosure.finishAnimation(); }}>
        <div key={active.id}>
          <RenderedContent html={active.html} variant="announcement" className="min-w-0 pt-1" />
        </div>
      </CollapsibleContent>
    </Collapsible>
  </aside>;
}
