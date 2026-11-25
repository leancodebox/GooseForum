"use client";
import { avatarInitial } from "@gooseforum/ui/lib/avatar-initial";
import { DefaultBrand } from "./default-brand";

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
  LanguagesIcon,
  LayoutGridIcon,
  LinkIcon,
  LogOutIcon,
  MenuIcon,
  MessageCircleIcon,
  PaletteIcon,
  PenSquareIcon,
  SettingsIcon,
  ShieldIcon,
  ScaleIcon,
  UserRoundIcon,
  UsersRoundIcon,
  XIcon,
} from "lucide-react";
import type { LayoutPayload, NavItemPayload } from "@gooseforum/client";
import { useTranslation } from "react-i18next";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@gooseforum/ui/components/avatar";
import { Button } from "@gooseforum/ui/components/button";
import { CategoryNavIcon } from "@gooseforum/ui/components/category-icon";
import { PageNavigationIcons, NavigationIconContent, preloadNavigationIcons } from "@gooseforum/ui/components/navigation-icon";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@gooseforum/ui/components/dropdown-menu";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@gooseforum/ui/components/sheet";
import { authLocales, localeLabels } from "@gooseforum/runtime/i18n/auth";
import { cn } from "@gooseforum/ui/lib/utils";
import { GooseLink, useGooseRuntime } from "@gooseforum/runtime";
import { ThemeMenuOptions, ThemeMenuTrigger } from "./theme-menu";
import { announceUnreadStatus, unreadStatusEvent } from "@gooseforum/runtime/unread-status";
import { HeaderSearch } from "./header-search";
import {
  emptyShellHeader,
  ShellHeaderContext,
  type ShellHeaderState,
} from "./shell-header";

const headerMenuKeys = ["theme", "language", "user"] as const;
type HeaderMenu = (typeof headerMenuKeys)[number];

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
    void import("../drafts/local-draft-store").then(({ activateDraftAccount }) =>
      activateDraftAccount(layout.viewer.isAuthenticated ? layout.viewer.id : 0),
    );
  }, [layout.viewer.id, layout.viewer.isAuthenticated]);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [moreState, setMoreState] = useState<"initial" | "expanded" | "collapsed">("initial");
  const [sidebarCollapsed, setSidebarCollapsed] = useState(() => {
    try { return localStorage.getItem("goose:shell-sidebar-collapsed") === "true"; }
    catch { return false; }
  });
  const [sidebarAnimate, setSidebarAnimate] = useState(false);
  function toggleSidebar() {
    const collapsed = !sidebarCollapsed;
    setSidebarCollapsed(collapsed);
    setSidebarAnimate(true);
    try { localStorage.setItem("goose:shell-sidebar-collapsed", String(collapsed)); } catch {}
  }
  const [unread, setUnread] = useState(layout.unread);
  const [themeMenuOpen, setThemeMenuOpen] = useState(false);
  const [languageMenuOpen, setLanguageMenuOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const [shellHeader, setShellHeader] = useState<ShellHeaderState>(
    emptyShellHeader,
  );
  const menuCloseTimers = useRef<
    Record<HeaderMenu, number | undefined>
  >({
    theme: undefined,
    language: undefined,
    user: undefined,
  });
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
  const headerItems = layout.header ? serverItems(layout.header) : systemResources;
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

  async function logout() {
    if (!window.dispatchEvent(new Event("goose:before-logout", { cancelable: true }))) return;
    await runtime.api.auth.logout();
    const { clearAccountDrafts, draftOwner } = await import("../drafts/local-draft-store");
    await clearAccountDrafts(draftOwner(layout.viewer.id));
    runtime.redirect(runtime.currentUrl);
  }

  function setHeaderMenu(menu: HeaderMenu, open: boolean) {
    const setters = {
      theme: setThemeMenuOpen,
      language: setLanguageMenuOpen,
      user: setUserMenuOpen,
    };
    window.clearTimeout(menuCloseTimers.current[menu]);
    menuCloseTimers.current[menu] = undefined;
    if (open) {
      for (const other of headerMenuKeys) {
        if (other === menu) continue;
        window.clearTimeout(menuCloseTimers.current[other]);
        menuCloseTimers.current[other] = undefined;
        setters[other](false);
      }
    }
    setters[menu](open);
  }

  function openHeaderMenuOnHover(menu: HeaderMenu) {
    if (window.matchMedia?.("(hover: none)").matches) return;
    setHeaderMenu(menu, true);
  }

  function closeHeaderMenuSoon(menu: HeaderMenu) {
    if (window.matchMedia?.("(hover: none)").matches) return;
    window.clearTimeout(menuCloseTimers.current[menu]);
    menuCloseTimers.current[menu] = window.setTimeout(
      () => setHeaderMenu(menu, false),
      120,
    );
  }

  useEffect(() => {
    setUnread(layout.unread);
    announceUnreadStatus(layout.unread);
  }, [layout.unread]);

  useEffect(() => {
    const updateUnread = (event: Event) => {
      const detail = (event as CustomEvent<Partial<LayoutPayload["unread"]>>)
        .detail;
      setUnread((current) => ({ ...current, ...detail }));
    };
    window.addEventListener(unreadStatusEvent, updateUnread);
    return () => window.removeEventListener(unreadStatusEvent, updateUnread);
  }, []);

  useEffect(
    () => () => {
      window.clearTimeout(menuCloseTimers.current.theme);
      window.clearTimeout(menuCloseTimers.current.language);
      window.clearTimeout(menuCloseTimers.current.user);
    },
    [],
  );

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
      onNavigate={() => setMobileOpen(false)}
    />
  ); }

  return (
    <ShellHeaderContext.Provider value={setShellHeader}>
      <div className="min-h-svh bg-muted text-foreground">
      <header hidden={standalone} className="sticky top-0 z-40 border-b bg-background/95 backdrop-blur-sm">
        <div className="mx-auto flex h-16 w-full max-w-[1600px] items-center gap-2 px-3 lg:gap-3 lg:px-6">
          <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
            <SheetTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="[&_svg]:size-5 lg:hidden"
                aria-label={t("openMenu")}
              >
                <MenuIcon className="size-5" />
              </Button>
            </SheetTrigger>
            <SheetContent
              side="left"
              className="w-72 p-0"
              showCloseButton={false}
            >
              <SheetHeader className="relative border-b">
                <SheetTitle>{t("menu")}</SheetTitle>
                <SheetDescription className="sr-only">
                  {t("menu")}
                </SheetDescription>
                <SheetClose asChild>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    className="absolute right-3 top-3"
                    aria-label={t("closeMenu")}
                  >
                    <XIcon />
                  </Button>
                </SheetClose>
              </SheetHeader>
              <div className="min-h-0 flex-1 overflow-y-auto px-3 pb-5 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
                {renderNavigation()}
              </div>
            </SheetContent>
          </Sheet>

          <Button variant="ghost" size="icon" className="hidden shrink-0 lg:inline-flex"
            aria-label={sidebarCollapsed ? t("expandSidebar") : t("collapseSidebar")}
            aria-expanded={!sidebarCollapsed} aria-controls="goose-desktop-sidebar" onClick={toggleSidebar}>
            <MenuIcon className="size-5.5" />
          </Button>

          <div className={cn(shellHeader.visible && "hidden md:block")}>
            <SiteBrand layout={layout} />
          </div>
          {shellHeader.visible ? (
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
          {headerItems.length > 0 && (
            <Suspense fallback={null}>
              <nav
                className={cn(
                  "hidden items-center gap-1 lg:flex",
                  shellHeader.visible && "lg:hidden",
                )}
                aria-label="Header navigation"
              >
                {headerItems.map((item) => (
                  <GooseLink
                    key={item.key}
                    href={item.url}
                    className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm font-medium text-foreground/75 hover:bg-accent"
                  >
                    {item.iconName && item.key !== "links" && item.key !== "sponsors" && (
                      <span className="size-4 shrink-0"><NavigationIconContent name={item.iconName} /></span>
                    )}
                    {item.label}
                  </GooseLink>
                ))}
              </nav>
            </Suspense>
          )}

          <div className="ml-auto flex min-w-0 items-center justify-end gap-0.5 lg:gap-1">
            <HeaderSearch />
            <div
              className={cn(
                "shrink-0 items-center gap-0.5 lg:gap-1",
                shellHeader.visible ? "hidden md:flex" : "flex",
              )}
            >
            <DropdownMenu
              modal={false}
              open={themeMenuOpen}
              onOpenChange={(open) => setHeaderMenu("theme", open)}
            >
              <div
                className="relative"
                onMouseEnter={() => openHeaderMenuOnHover("theme")}
                onMouseLeave={() => closeHeaderMenuSoon("theme")}
              >
                <ThemeMenuTrigger size="icon-lg" />
              </div>
              <DropdownMenuContent
                className="w-max min-w-32 whitespace-nowrap"
                align="end"
                sideOffset={8}
                onMouseEnter={() => openHeaderMenuOnHover("theme")}
                onMouseLeave={() => closeHeaderMenuSoon("theme")}
                onCloseAutoFocus={(event) => event.preventDefault()}
              >
                <ThemeMenuOptions />
              </DropdownMenuContent>
            </DropdownMenu>
            <DropdownMenu
              modal={false}
              open={languageMenuOpen}
              onOpenChange={(open) => setHeaderMenu("language", open)}
            >
              <div
                className="relative"
                onMouseEnter={() => openHeaderMenuOnHover("language")}
                onMouseLeave={() => closeHeaderMenuSoon("language")}
              >
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon-lg"
                    className="[&_svg]:size-5"
                    aria-label={t("switchLanguage")}
                  >
                    <LanguagesIcon className="size-5" />
                  </Button>
                </DropdownMenuTrigger>
              </div>
              <DropdownMenuContent
                align="end"
                sideOffset={8}
                onMouseEnter={() => openHeaderMenuOnHover("language")}
                onMouseLeave={() => closeHeaderMenuSoon("language")}
                onCloseAutoFocus={(event) => event.preventDefault()}
              >
                <DropdownMenuRadioGroup
                  value={runtime.locale}
                  onValueChange={(value) =>
                    void runtime.setLocale(value as typeof runtime.locale)
                  }
                >
                  {authLocales.map((locale) => (
                    <DropdownMenuRadioItem key={locale} value={locale}>
                      {localeLabels[locale].label}
                    </DropdownMenuRadioItem>
                  ))}
                </DropdownMenuRadioGroup>
              </DropdownMenuContent>
            </DropdownMenu>

            {layout.viewer.isAuthenticated ? (
              <DropdownMenu
                modal={false}
                open={userMenuOpen}
                onOpenChange={(open) => setHeaderMenu("user", open)}
              >
                <div
                  className="relative ml-1 inline-flex size-10 shrink-0 items-center justify-center rounded-full"
                  onMouseEnter={() => openHeaderMenuOnHover("user")}
                  onMouseLeave={() => closeHeaderMenuSoon("user")}
                >
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon-lg"
                      className="relative grid size-10 place-items-center rounded-full p-0"
                      aria-label={layout.viewer.username}
                    >
                      <Avatar className="size-9">
                        <AvatarImage
                          src={layout.viewer.avatarUrl}
                          alt={layout.viewer.username}
                        />
                        <AvatarFallback userId={layout.viewer.id}>
                          {avatarInitial(layout.viewer.nickname, layout.viewer.username)}
                        </AvatarFallback>
                      </Avatar>
                      {unread.messages || unread.notifications ? (
                        <span
                          className="absolute right-0.5 top-0.5 size-2.5 rounded-full bg-destructive ring-2 ring-background"
                          aria-hidden="true"
                        />
                      ) : null}
                    </Button>
                  </DropdownMenuTrigger>
                </div>
                <DropdownMenuContent
                  align="end"
                  sideOffset={8}
                  className="w-48"
                  onMouseEnter={() => openHeaderMenuOnHover("user")}
                  onMouseLeave={() => closeHeaderMenuSoon("user")}
                  onCloseAutoFocus={(event) => event.preventDefault()}
                >
                  <DropdownMenuGroup>
                    <DropdownMenuItem asChild>
                      <GooseLink href={`/u/${layout.viewer.id}`}>
                        <UserRoundIcon />
                        {t("profile")}
                      </GooseLink>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <GooseLink href="/messages">
                        <InboxIcon />
                        {t("nav.messages")}
                        {unread.messages ? <MenuDot /> : null}
                      </GooseLink>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <GooseLink href="/notifications">
                        <BellIcon />
                        {t("nav.notifications")}
                        {unread.notifications ? <MenuDot /> : null}
                      </GooseLink>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <GooseLink href="/drafts">
                        <FileTextIcon />
                        {t("nav.drafts")}
                      </GooseLink>
                    </DropdownMenuItem>
                  </DropdownMenuGroup>
                  <DropdownMenuSeparator />
                  <DropdownMenuGroup>
                    <DropdownMenuItem asChild>
                      <GooseLink href="/publish">
                        <PenSquareIcon />
                        {t("publish")}
                      </GooseLink>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <GooseLink href="/settings">
                        <SettingsIcon />
                        {t("settings")}
                      </GooseLink>
                    </DropdownMenuItem>
                    {layout.viewer.canAccessAdmin ? (
                      <DropdownMenuItem asChild>
                        <GooseLink href="/admin">
                          <ShieldIcon />
                          {t("admin")}
                        </GooseLink>
                      </DropdownMenuItem>
                    ) : null}
                  </DropdownMenuGroup>
                  <DropdownMenuSeparator />
                  <DropdownMenuGroup>
                    <DropdownMenuItem
                      variant="destructive"
                      onSelect={() => void logout()}
                    >
                      <LogOutIcon />
                      {t("logout")}
                    </DropdownMenuItem>
                  </DropdownMenuGroup>
                </DropdownMenuContent>
              </DropdownMenu>
            ) : (
              <>
                <Button asChild variant="ghost">
                  <GooseLink href="/login">{t("login")}</GooseLink>
                </Button>
                <Button asChild className="hidden lg:inline-flex">
                  <GooseLink href="/login?register=true">
                    {t("register")}
                  </GooseLink>
                </Button>
              </>
            )}
            </div>
          </div>
        </div>
      </header>

      <main data-sidebar-collapsed={sidebarCollapsed} data-sidebar-animate={sidebarAnimate}
        className={standalone ? "w-full" : "goose-shell-grid mx-auto grid w-full max-w-[1600px] grid-cols-1 lg:px-6 lg:py-3"}>
        <aside
          id="goose-desktop-sidebar"
          inert={sidebarCollapsed || standalone} aria-hidden={sidebarCollapsed || standalone}
          className={cn(standalone ? "hidden" : "sticky top-16 -my-3 hidden h-[calc(100vh-4rem)] min-w-0 self-start overflow-x-hidden overflow-y-auto py-3 pl-1 pr-3 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden lg:block", sidebarCollapsed && "invisible")}
          aria-label="Sidebar"
        >
          <div className="w-[194px] xl:w-[208px]">{renderNavigation()}</div>
        </aside>
        <section data-slot="goose-page-content" className="min-w-0">
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
        </section>
      </main>
      </div>
    </ShellHeaderContext.Provider>
  );
}

function SiteBrand({ layout }: { layout: LayoutPayload }) {
  const site = layout.site;
  const content =
    site.brandType === "image" && site.brandImage ? (
      <img
        src={site.brandImage}
        alt={site.name}
        className="h-8 w-auto max-w-40 object-contain"
      />
    ) : site.brandType === "text" ? (
      site.brandText || site.name
    ) : (
      <DefaultBrand />
    );
  return (
    <GooseLink
      href="/"
      className="shrink-0 text-xl font-semibold tracking-tight text-primary lg:text-2xl"
    >
      {content}
    </GooseLink>
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
