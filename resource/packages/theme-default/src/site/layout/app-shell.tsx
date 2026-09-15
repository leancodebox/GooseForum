"use client";

import {
  useEffect,
  useRef,
  useState,
  type ComponentType,
  type ReactNode,
  type SVGProps,
} from "react";
import {
  BellIcon,
  EllipsisVerticalIcon,
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
  MoonIcon,
  PaletteIcon,
  PenSquareIcon,
  SearchIcon,
  SettingsIcon,
  ShieldIcon,
  ScaleIcon,
  SunIcon,
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
import { unreadStatusEvent } from "@gooseforum/runtime/unread-status";
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
  color?: string;
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
  const [mobileOpen, setMobileOpen] = useState(false);
  const [unread, setUnread] = useState(layout.unread);
  const [languageMenuOpen, setLanguageMenuOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const [shellHeader, setShellHeader] = useState<ShellHeaderState>(
    emptyShellHeader,
  );
  const menuCloseTimers = useRef<
    Record<"language" | "user", number | undefined>
  >({
    language: undefined,
    user: undefined,
  });
  const activeKey = layout.sidebar.activeKey || "topics";
  const primary = [
    { ...nav("topics", t("nav.topics"), "/", MessageCircleIcon), active: ['topics', 'hot', 'popular'].includes(activeKey) },
    nav("categories", t("nav.categories"), "/categories", LayoutGridIcon),
    nav("members", t("nav.members"), "/members", UsersRoundIcon),
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
  const resources = [
    nav("links", t("nav.links"), "/links", LinkIcon),
    nav("sponsors", t("nav.sponsors"), "/sponsors", HeartIcon),
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
    color: category.color,
  }));
  const groups = (layout.sidebar.groups || [])
    .map((group) => ({
      key: group.key,
      title: displayLabel(group, group.title),
      items: serverItems(group.items),
    }))
    .filter((group) => group.title && group.items.length);
  const headerItems = layout.header?.length
    ? serverItems(layout.header)
    : resources.slice(0, 2);

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
    return (items || []).map((item) =>
      nav(item.key, displayLabel(item, item.label), item.url, LinkIcon),
    );
  }

  async function logout() {
    await runtime.api.auth.logout();
    runtime.redirect(runtime.currentUrl);
  }

  function setHeaderMenu(menu: "language" | "user", open: boolean) {
    window.clearTimeout(menuCloseTimers.current[menu]);
    menuCloseTimers.current[menu] = undefined;
    if (open) {
      const other = menu === "language" ? "user" : "language";
      window.clearTimeout(menuCloseTimers.current[other]);
      menuCloseTimers.current[other] = undefined;
      if (other === "language") setLanguageMenuOpen(false);
      else setUserMenuOpen(false);
    }
    if (menu === "language") setLanguageMenuOpen(open);
    else setUserMenuOpen(open);
  }

  function closeHeaderMenuSoon(menu: "language" | "user") {
    window.clearTimeout(menuCloseTimers.current[menu]);
    menuCloseTimers.current[menu] = window.setTimeout(
      () => setHeaderMenu(menu, false),
      120,
    );
  }

  useEffect(() => {
    setUnread(layout.unread);
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

  useEffect(() => {
    if (!layout.viewer.isAuthenticated) return;
    const refreshUnread = () => {
      void runtime.api.notifications
        .unread()
        .then(setUnread)
        .catch(() => undefined);
    };
    const timer = window.setInterval(refreshUnread, 30_000);
    return () => window.clearInterval(timer);
  }, [layout.viewer.isAuthenticated, runtime.api.notifications]);

  useEffect(
    () => () => {
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
      labels={{ categories: t("categories"), more: t("more") }}
      onNavigate={() => setMobileOpen(false)}
    />
  ); }

  return (
    <ShellHeaderContext.Provider value={setShellHeader}>
      <div className="min-h-svh bg-muted text-foreground">
      {runtime.isNavigating ? (
        <div className="fixed inset-x-0 top-0 z-50 h-0.5 animate-pulse bg-primary" />
      ) : null}
      <header hidden={standalone} className="sticky top-0 z-40 border-b bg-background/95 backdrop-blur-sm">
        <div className="mx-auto flex h-16 w-full max-w-[1600px] items-center gap-2 px-3 lg:gap-8 lg:px-8">
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
                className="rounded-md px-2 py-1.5 text-sm font-medium text-foreground/75 hover:bg-accent"
              >
                {item.label}
              </GooseLink>
            ))}
          </nav>

          <div
            className={cn(
              "ml-auto items-center gap-0.5 lg:gap-1",
              shellHeader.visible ? "hidden md:flex" : "flex",
            )}
          >
            <Button
              asChild
              variant="ghost"
              size="icon-lg"
              className="[&_svg]:size-5"
              aria-label={t("search")}
            >
              <GooseLink href="/search">
                <SearchIcon className="size-5" />
              </GooseLink>
            </Button>
            <Button
              variant="ghost"
              size="icon-lg"
              className="[&_svg]:size-5"
              onClick={runtime.toggleTheme}
              aria-label={
                runtime.theme === "gf-dark" ? t("switchLight") : t("switchDark")
              }
            >
              {runtime.theme === "gf-dark" ? (
                <SunIcon className="size-5" />
              ) : (
                <MoonIcon className="size-5" />
              )}
            </Button>
            <DropdownMenu
              modal={false}
              open={languageMenuOpen}
              onOpenChange={(open) => setHeaderMenu("language", open)}
            >
              <div
                className="relative"
                onMouseEnter={() => setHeaderMenu("language", true)}
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
                onMouseEnter={() => setHeaderMenu("language", true)}
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
                  onMouseEnter={() => setHeaderMenu("user", true)}
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
                        <AvatarFallback>
                          {layout.viewer.username.slice(0, 2).toUpperCase()}
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
                  onMouseEnter={() => setHeaderMenu("user", true)}
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
      </header>

      <main className={standalone ? "w-full" : "mx-auto grid w-full max-w-[1600px] grid-cols-1 lg:grid-cols-[210px_minmax(0,1fr)] lg:gap-3 lg:px-8 lg:py-3 xl:grid-cols-[224px_minmax(0,1fr)]"}>
        <aside
          className={standalone ? "hidden" : "sticky top-16 -my-3 hidden h-[calc(100vh-4rem)] min-w-0 self-start overflow-y-auto py-3 pl-1 pr-3 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden lg:block"}
          aria-label="Sidebar"
        >
          {renderNavigation()}
        </aside>
        <section className="min-w-0">{children}</section>
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
      <>
        Goose<span className="text-foreground">Forum</span>
      </>
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
  onNavigate,
}: {
  primary: ShellNavItem[];
  resources: ShellNavItem[];
  groups: Array<{ key: string; title: string; items: ShellNavItem[] }>;
  categories: ShellNavItem[];
  footer: LayoutPayload["footer"];
  labels: { categories: string; more: string };
  onNavigate(): void;
}) {
  const topicItem = primary.find(item => item.key === 'topics');
  const primaryKeys = ['topics', 'categories', 'members', 'messages', 'notifications'];
  const secondaryItems = [...primary.filter(item => !primaryKeys.includes(item.key)), ...resources];
  const promotedItem = secondaryItems.find(item => item.active);
  const visibleItems = [
    ...(topicItem ? [topicItem] : []),
    ...primary.filter(item => ['categories', 'members', 'messages', 'notifications'].includes(item.key)),
    ...(promotedItem ? [promotedItem] : []),
  ];
  const moreItems = [
    ...secondaryItems.filter(item => item.key !== promotedItem?.key),
  ];
  return (
    <nav className="flex flex-col gap-2">
      <div className="flex flex-col gap-0.5">
        <NavList items={visibleItems} onNavigate={onNavigate} />
        <NavMoreMenu label={labels.more} items={moreItems} onNavigate={onNavigate} />
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

function NavMoreMenu({ label, items, onNavigate }: { label: string; items: ShellNavItem[]; onNavigate(): void }) {
  return <DropdownMenu modal={false}>
    <DropdownMenuTrigger asChild>
      <Button variant="ghost" aria-label={label} className="site-more-trigger h-8 w-full justify-start gap-2 text-[13px] leading-[18.5714px] text-foreground/75 data-[state=open]:text-foreground">
        <EllipsisVerticalIcon data-icon="inline-start" aria-hidden="true" />
        <span>{label}</span>
        {items.some(item => item.attention) ? <MenuDot /> : <span className="ml-auto" />}
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent side="bottom" align="start" sideOffset={2} collisionPadding={8} className="site-more-menu min-w-0 max-w-[calc(100vw-2rem)] max-h-[min(70dvh,32rem)] overflow-y-auto rounded-md p-0.5 ring-0 duration-100 [--tw-enter-scale:1]! [--tw-exit-scale:1]! [--tw-enter-translate-y:0]!">
      <DropdownMenuGroup>
        {items.map(item => { const Icon = item.icon; return <DropdownMenuItem key={item.key} asChild className="h-8 gap-2 rounded-md px-2 text-[13px] font-medium leading-[18.5714px] text-foreground/75 focus:bg-accent focus:text-foreground">
          <GooseLink href={item.url} onClick={onNavigate}>
            {Icon ? <Icon data-icon="inline-start" aria-hidden="true" /> : null}
            <span>{item.label}</span>
            {item.attention ? <MenuDot /> : null}
          </GooseLink>
        </DropdownMenuItem> })}
      </DropdownMenuGroup>
    </DropdownMenuContent>
  </DropdownMenu>;
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
  items,
  onNavigate,
  compact = false,
}: {
  items: ShellNavItem[];
  onNavigate(): void;
  compact?: boolean;
}) {
  return (
    <ul className={cn("flex flex-col", compact ? "gap-px" : "gap-0.5")}>
      {items.map((item) => {
        const Icon = item.icon;
        return (
          <li key={item.key}>
            <Button
              asChild
              variant="ghost"
              className={cn(
                "w-full justify-start gap-2 text-[13px] leading-[18.5714px] transition-none",
                compact ? "h-7" : "h-8",
                "focus-visible:ring-2",
                item.active
                  ? "bg-primary/10 text-primary hover:bg-primary/15 hover:text-primary"
                  : "text-foreground/75 hover:bg-accent hover:text-foreground",
              )}
            >
              <GooseLink
                href={item.url}
                aria-current={item.active ? "page" : undefined}
                onClick={onNavigate}
              >
                {Icon ? (
                  <Icon aria-hidden="true" />
                ) : (
                  <span
                    className="size-2 rounded-sm"
                    style={{ backgroundColor: item.color }}
                  />
                )}
                <span className="min-w-0 truncate">{item.label}</span>
                {item.attention ? <MenuDot /> : null}
              </GooseLink>
            </Button>
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
