"use client";

import {
  Suspense,
  use,
  useEffect,
  useId,
  useLayoutEffect,
  useRef,
  useState,
  useTransition,
  type ComponentType,
  type ReactNode,
  type SVGProps,
} from "react";
import {
  BellIcon,
  ChevronDownIcon,
  FileTextIcon,
  HeartIcon,
  InboxIcon,
  KeyRoundIcon,
  LayoutGridIcon,
  LinkIcon,
  MessageCircleIcon,
  PaletteIcon,
  ScaleIcon,
  ShieldIcon,
  UsersRoundIcon,
} from "lucide-react";
import type { LayoutPayload, NavItemPayload } from "@gooseforum/client";
import { useTranslation } from "react-i18next";
import { HeaderSlot } from "@gooseforum/ui/components/header-host";
import { ShellGrid, ShellMenuTrigger, useShellSidebar } from "@gooseforum/ui/components/shell-layout";
import { SiteHeader } from "@gooseforum/ui/components/site-header";
import { searchUrl } from "@gooseforum/runtime/search-url";
import { Button } from "@gooseforum/ui/components/button";
import { CategoryNavIcon } from "@gooseforum/ui/components/category-icon";
import { PageNavigationIcons, NavigationIconContent, preloadNavigationIcons } from "@gooseforum/ui/components/navigation-icon";
import { authLocales, localeLabels } from "@gooseforum/runtime/i18n/auth";
import { cn } from "@gooseforum/ui/lib/utils";
import { GooseLink, GooseNavigationProvider, useGooseRuntime } from "@gooseforum/runtime";
import { logout } from "@gooseforum/runtime/account";
import { useUnreadStatus } from "@gooseforum/runtime/unread-status";
import {
  emptyShellHeader,
  ShellHeaderContext,
  type ShellHeaderState,
} from "./shell-header";

interface ShellNavItem {
  key: string;
  label: string;
  url: string;
  active: boolean;
  icon?: ComponentType<SVGProps<SVGSVGElement>>;
  iconName?: string;
  category?: { icon?: string; color: string };
  attention?: boolean;
}

export function AppShell({
  layout,
  children,
  standalone = false,
}: {
  layout: LayoutPayload;
  children: ReactNode;
  standalone?: boolean;
}) {
  const { t, i18n } = useTranslation("shell");
  const runtime = useGooseRuntime();
  useEffect(() => {
    void import("@gooseforum/runtime/local-draft-store").then(({ activateDraftAccount }) =>
      activateDraftAccount(layout.viewer.isAuthenticated ? layout.viewer.id : 0),
    );
  }, [layout.viewer.id, layout.viewer.isAuthenticated]);
  const sidebar = useShellSidebar();
  const [moreState, setMoreState] = useState<"initial" | "expanded" | "collapsed">("initial");
  const unread = useUnreadStatus(layout.unread);
  const [shellHeader, setShellHeader] = useState<ShellHeaderState>(
    emptyShellHeader,
  );
  const activeKey = layout.sidebar.activeKey || "topics";
  const hiddenSystemLinks = layout.sidebar.hiddenSystemLinks ?? [];
  const primary = [
    { ...nav("topics", t("nav.topics"), "/", MessageCircleIcon), active: ['topics', 'hot', 'popular'].includes(activeKey) },
    nav("categories", t("nav.categories"), "/categories", LayoutGridIcon),
    ...(!hiddenSystemLinks.includes("members")
      ? [nav("members", t("nav.members"), "/members", UsersRoundIcon)]
      : []),
    ...(layout.viewer.isAuthenticated
      ? [
          nav(
            "messages",
            t("nav.messages"),
            "/messages",
            InboxIcon,
            unread.messages,
          ),
          nav(
            "notifications",
            t("nav.notifications"),
            "/notifications",
            BellIcon,
            unread.notifications,
          ),
          nav("drafts", t("nav.drafts"), "/drafts", FileTextIcon),
          nav("access-groups", t("accessGroups"), "/access-groups", KeyRoundIcon),
        ]
      : []),
    ...(layout.viewer.isModerator
      ? [
          nav(
            "moderation",
            t("nav.moderation"),
            "/moderation",
            ScaleIcon,
            unread.moderationReports,
          ),
        ]
      : []),
    ...serverItems(layout.sidebar.main),
  ];
  const systemResources = [
    nav("links", t("nav.links"), "/links", LinkIcon),
    nav("sponsors", t("nav.sponsors"), "/sponsors", HeartIcon),
  ];
  const resources = [
    ...systemResources.filter((resource) => !hiddenSystemLinks.includes(resource.key)),
    ...(layout.viewer.isAuthenticated && layout.viewer.canAccessAdmin
      ? [nav("theme-preview", t("themePreview"), "/theme-preview", PaletteIcon)]
      : []),
    ...serverItems(layout.sidebar.resources),
  ];
  const categories = (layout.sidebar.categories || []).map((category) => ({
    key: `category_${category.id}`,
    label: category.label,
    url: category.url,
    active: activeKey === `category_${category.id}`,
    category,
  }));
  const groups = (layout.sidebar.groups || [])
    .map((group) => ({
      key: group.key,
      title: displayLabel(group, group.title),
      items: serverItems(group.items),
    }))
    .filter((group) => group.title && group.items.length);
  function nav(
    key: string,
    label: string,
    url: string,
    icon?: ShellNavItem["icon"],
    attention = false,
  ): ShellNavItem {
    return { key, label, url, icon, attention, active: activeKey === key };
  }

  function displayLabel(item: { i18nLabel?: string }, fallback: string) {
    const key = item.i18nLabel?.startsWith("shell.")
      ? item.i18nLabel.slice("shell.".length)
      : "";
    return key && i18n.exists(key, { ns: "shell" }) ? t(key) : fallback;
  }

  function serverItems(items?: NavItemPayload[]): ShellNavItem[] {
    return (items || []).map((item) => ({
      ...nav(item.key, displayLabel(item, item.label), item.url, LinkIcon),
      iconName: item.icon,
    }));
  }

  function renderNavigation() { return (
    <ShellNavigation
      primary={primary}
      resources={resources}
      groups={groups}
      categories={categories}
      footer={layout.footer}
      labels={{ categories: t("categories"), more: t("more"), showLess: t("showLess") }}
      expanded={moreState === "expanded"}
      hasExpanded={moreState !== "initial"}
      onExpandedChange={(expanded) => setMoreState(expanded ? "expanded" : "collapsed")}
      onNavigate={() => sidebar.setMobileOpen(false)}
    />
  ); }

  return (
    <ShellHeaderContext.Provider value={setShellHeader}>
      <div className={cn(standalone ? "min-h-svh" : "min-h-[calc(100svh-4rem)]", "bg-muted text-foreground")}>
        <HeaderSlot>
          <GooseNavigationProvider value={runtime.navigate}>
            <PageNavigationIcons value={layout.icons}>
              <SiteHeader site={layout.site} viewer={layout.viewer} header={layout.header} unread={unread}
                hidden={standalone} Link={GooseLink} t={(key, fallback) => t(key, { defaultValue: fallback ?? key })}
                locale={runtime.locale} locales={authLocales} localeLabels={localeLabels}
                onLocaleChange={runtime.setLocale} themePreference={runtime.themePreference ?? "system"}
                onThemeChange={value => runtime.setThemePreference?.(value)}
                onSearch={query => void runtime.navigate(searchUrl(query))} onLogout={() => logout(runtime.api, layout.viewer.id, () => runtime.redirect(runtime.currentUrl))}
                sidebarTrigger={<ShellMenuTrigger sidebar={sidebar} navigation={renderNavigation()} t={t} />}
                contextTitle={shellHeader.visible ? (
                    <button
                      type="button"
                      className="flex min-w-0 flex-1 flex-col items-start justify-center gap-0.5 self-stretch text-left"
                      onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
                    >
                      <span className="block max-w-full truncate text-lg font-semibold leading-6 hover:text-primary md:text-xl">
                        {shellHeader.title}
                      </span>
                      {shellHeader.tags.length ? (
                        <span className="flex max-w-full items-center gap-2 overflow-hidden text-[11px] font-medium leading-4 text-muted-foreground">
                          {shellHeader.tags.map((tag) => (
                            <span
                              key={tag.id}
                              className="inline-flex min-w-0 shrink-0 items-center gap-1"
                            >
                              <span
                                className="size-1.5 rounded-sm"
                                style={{ backgroundColor: tag.color || "currentColor" }}
                              />
                              <span className="max-w-28 truncate">{tag.name}</span>
                            </span>
                          ))}
                        </span>
                      ) : null}
                    </button>
                  ) : null}
              />
            </PageNavigationIcons>
          </GooseNavigationProvider>
        </HeaderSlot>
        <ShellGrid sidebar={sidebar} standalone={standalone} navigation={renderNavigation()}>
          {layout.viewer.isAuthenticated && layout.viewer.restrictionStatus === "suspended" ? <div role="status" className="flex flex-wrap items-start gap-2 border-b px-4 py-3 text-sm">
            <ShieldIcon className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
            <div className="min-w-0 flex-1 break-words">
              <p className="font-medium">{t("accountSuspended")}</p>
              {layout.viewer.restrictionUntil ? <p className="text-muted-foreground">{t("restrictionUntil", { date: new Date(layout.viewer.restrictionUntil).toLocaleString(i18n.language) })}</p> : <p className="text-muted-foreground">{t("restrictionPermanent")}</p>}
              {layout.viewer.restrictionReason ? <p className="mt-1">{layout.viewer.restrictionReason}</p> : null}
            </div>
            <GooseLink href="/settings?tab=account" className="underline underline-offset-4">{t("accountSecurity")}</GooseLink>
          </div> : null}
          {children}
        </ShellGrid>
      </div>
    </ShellHeaderContext.Provider>
  );
}

function ShellNavigation({
  primary,
  resources,
  groups,
  categories,
  footer,
  labels,
  expanded,
  hasExpanded,
  onExpandedChange,
  onNavigate,
}: {
  primary: ShellNavItem[];
  resources: ShellNavItem[];
  groups: Array<{ key: string; title: string; items: ShellNavItem[] }>;
  categories: ShellNavItem[];
  footer: LayoutPayload["footer"];
  labels: { categories: string; more: string; showLess: string };
  expanded: boolean;
  hasExpanded: boolean;
  onExpandedChange(expanded: boolean): void;
  onNavigate(): void;
}) {
  const pageIcons = use(PageNavigationIcons);
  const listId = useId();
  const [opening, startOpening] = useTransition();
  const focusRevealedLink = useRef(false);
  useLayoutEffect(() => {
    if (!expanded || !focusRevealedLink.current) return;
    focusRevealedLink.current = false;
    document.getElementById(listId)
      ?.querySelector<HTMLAnchorElement>('[data-collapsible] a:not([aria-current="page"])')
      ?.focus({ preventScroll: true });
  }, [expanded, listId]);

  const topicItem = primary.find(item => item.key === 'topics');
  const alwaysVisibleKeys = ['categories', 'members', 'messages', 'notifications'];
  const secondaryItems = [...primary.filter(item => item.key !== 'topics' && !alwaysVisibleKeys.includes(item.key)), ...resources];
  const primaryItems = [
    ...(topicItem ? [topicItem] : []),
    ...primary.filter(item => alwaysVisibleKeys.includes(item.key)),
  ];
  const moreItems = secondaryItems.filter(item => !item.active);
  function preloadMoreIcons() {
    if (pageIcons) return;
    return preloadNavigationIcons(moreItems.map(item => item.iconName));
  }
  return (
    <nav className="flex flex-col gap-2">
      <div className="flex flex-col">
        <NavList
          id={listId}
          items={primaryItems}
          disclosure={{ items: secondaryItems, expanded, hasExpanded }}
          onNavigate={onNavigate}
        />
        {moreItems.length > 0 && (
          <Button
            type="button"
            variant="ghost"
            aria-expanded={expanded}
            aria-controls={listId}
            aria-busy={opening || undefined}
            onPointerEnter={() => !expanded && void preloadMoreIcons()}
            onFocus={() => !expanded && void preloadMoreIcons()}
            onClick={(event) => {
              if (opening) return;
              focusRevealedLink.current = !expanded && event.detail === 0;
              if (expanded || pageIcons) {
                onExpandedChange(!expanded);
                return;
              }
              startOpening(async () => {
                await preloadMoreIcons();
                startOpening(() => onExpandedChange(true));
              });
            }}
            className="group h-8 w-full justify-start gap-2 text-[13px] leading-[18.5714px] text-foreground/75 transition-colors duration-150 aria-expanded:bg-accent aria-expanded:text-foreground"
          >
            <ChevronDownIcon aria-hidden="true" className="group-aria-expanded:rotate-180 motion-safe:transition-transform motion-safe:duration-150" />
            <span>{expanded ? labels.showLess : labels.more}</span>
            {!expanded && moreItems.some(item => item.attention) ? <MenuDot /> : null}
          </Button>
        )}
      </div>
      {groups.map((group) => (
        <NavSection key={group.key} title={group.title}>
          <NavList items={group.items} onNavigate={onNavigate} compact />
        </NavSection>
      ))}
      {categories.length ? (
        <NavSection title={labels.categories}>
          <NavList items={categories} onNavigate={onNavigate} compact />
        </NavSection>
      ) : null}
      <footer className="px-2 text-xs leading-5 text-muted-foreground">
        <div className="flex flex-wrap gap-x-3">
          {(footer.links || []).map((link) => (
            <GooseLink key={`${link.name}-${link.url}`} href={link.url}>
              {link.name}
            </GooseLink>
          ))}
        </div>
        <div className="mt-1 flex flex-wrap gap-x-3">
          {(footer.primary || []).map((item) => (
            <span key={item}>{item}</span>
          ))}
        </div>
      </footer>
    </nav>
  );
}

function NavSection({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <section>
      <h2 className="mb-1 px-2 text-[10px] font-bold uppercase tracking-wide text-muted-foreground">
        {title}
      </h2>
      {children}
    </section>
  );
}

function NavList({
  id,
  items,
  onNavigate,
  compact = false,
  disclosure,
}: {
  id?: string;
  items: ShellNavItem[];
  onNavigate(): void;
  compact?: boolean;
  disclosure?: { items: ShellNavItem[]; expanded: boolean; hasExpanded: boolean };
}) {
  const allItems = disclosure ? [...items, ...disclosure.items] : items;
  return (
    <ul id={id} className={cn("flex flex-col", disclosure ? "goose-sidebar-nav-list" : compact ? "gap-px" : "gap-0.5")}>
      {allItems.map((item, index) => {
        const Icon = item.icon;
        const collapsible = index >= items.length;
        const collapsed = collapsible && !disclosure?.expanded && !item.active;
        return (
          <li
            key={item.key}
            data-collapsible={collapsible || undefined}
            data-collapsed={collapsed || undefined}
            inert={collapsed}
            aria-hidden={collapsed || undefined}
          >
            <Suspense fallback={<span aria-hidden="true" className={cn("block", compact ? "h-7" : "h-8")} />}>
              <Button
                asChild
                variant="ghost"
                className={cn(
                  "w-full justify-start gap-2 text-[13px] leading-[18.5714px] transition-none",
                  compact ? "h-7" : "h-8",
                  "sidebar-nav-link text-foreground/75 hover:bg-accent hover:text-foreground focus-visible:ring-2",
                  collapsible && "focus-visible:ring-inset",
                )}
              >
                <GooseLink
                  href={item.url}
                  aria-current={item.active ? "page" : undefined}
                  onClick={onNavigate}
                >
                  {item.category ? (
                    <CategoryNavIcon icon={item.category.icon} color={item.category.color} suspend />
                  ) : (
                    <NavigationIconContent
                      name={!collapsed || disclosure?.hasExpanded ? item.iconName : undefined}
                      fallback={Icon ? <Icon aria-hidden="true" /> : undefined}
                    />
                  )}
                  <span className="min-w-0 truncate">{item.label}</span>
                  {item.attention ? <MenuDot /> : null}
                </GooseLink>
              </Button>
            </Suspense>
          </li>
        );
      })}
    </ul>
  );
}

function MenuDot() {
  return (
    <span
      className="ml-auto size-2 shrink-0 rounded-full bg-destructive"
      aria-hidden="true"
    />
  );
}
