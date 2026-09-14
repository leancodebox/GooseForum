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
  FileTextIcon,
  FlameIcon,
  HeartIcon,
  InboxIcon,
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
  SunIcon,
  TrendingUpIcon,
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
} from "../../components/ui/avatar";
import { Button } from "../../components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../components/ui/dropdown-menu";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "../../components/ui/sheet";
import { authLocales } from "../../i18n/auth";
import { cn } from "../../lib/utils";
import { GooseLink, useGooseRuntime } from "../../runtime";
import { unreadStatusEvent } from "../../runtime/unread-status";

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
}: {
  layout: LayoutPayload;
  children: ReactNode;
}) {
  const { t, i18n } = useTranslation("shell");
  const runtime = useGooseRuntime();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [unread, setUnread] = useState(layout.unread);
  const [languageMenuOpen, setLanguageMenuOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const menuCloseTimers = useRef<
    Record<"language" | "user", number | undefined>
  >({
    language: undefined,
    user: undefined,
  });
  const activeKey = layout.sidebar.activeKey || "topics";
  const primary = [
    nav("topics", t("nav.topics"), "/", MessageCircleIcon),
    nav("hot", t("nav.hot"), "/?sort=hot", FlameIcon),
    nav("popular", t("nav.popular"), "/?sort=popular", TrendingUpIcon),
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
        ]
      : []),
    ...(layout.viewer.isModerator
      ? [
          nav(
            "moderation",
            t("nav.moderation"),
            "/moderation",
            BellIcon,
            unread.moderationReports,
          ),
        ]
      : []),
    ...serverItems(layout.sidebar.main),
  ];
  const resources = [
    nav("links", t("nav.links"), "/links", LinkIcon),
    nav("sponsors", t("nav.sponsors"), "/sponsors", HeartIcon),
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

  const navigation = (
    <ShellNavigation
      primary={primary}
      resources={resources}
      groups={groups}
      categories={categories}
      footer={layout.footer}
      labels={{ resources: t("resources"), categories: t("categories") }}
      onNavigate={() => setMobileOpen(false)}
    />
  );

  return (
    <div className="min-h-svh bg-muted text-foreground">
      {runtime.isNavigating ? (
        <div className="fixed inset-x-0 top-0 z-50 h-0.5 animate-pulse bg-primary" />
      ) : null}
      <header className="sticky top-0 z-40 border-b bg-background/95 backdrop-blur-sm">
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
                {navigation}
              </div>
            </SheetContent>
          </Sheet>

          <SiteBrand layout={layout} />
          <nav
            className="hidden items-center gap-1 lg:flex"
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

          <div className="ml-auto flex items-center gap-0.5 lg:gap-1">
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
                    runtime.setLocale(value as typeof runtime.locale)
                  }
                >
                  {authLocales.map((locale) => (
                    <DropdownMenuRadioItem key={locale} value={locale}>
                      {i18n.getFixedT(locale, "auth")("locale.label")}
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
                  className="relative"
                  onMouseEnter={() => setHeaderMenu("user", true)}
                  onMouseLeave={() => closeHeaderMenuSoon("user")}
                >
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon-lg"
                      className="relative"
                      aria-label={layout.viewer.username}
                    >
                      <Avatar>
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
                          className="absolute right-0 top-0 size-2.5 rounded-full bg-destructive ring-2 ring-background"
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
                        {layout.viewer.username}
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
                    <DropdownMenuItem asChild>
                      <GooseLink href="/access-groups">
                        <UsersRoundIcon />
                        {t("accessGroups")}
                      </GooseLink>
                    </DropdownMenuItem>
                    {layout.viewer.canAccessAdmin ? (
                      <DropdownMenuItem asChild>
                        <GooseLink href="/theme-preview">
                          <PaletteIcon />
                          {t("themePreview")}
                        </GooseLink>
                      </DropdownMenuItem>
                    ) : null}
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

      <main className="mx-auto grid w-full max-w-[1600px] grid-cols-1 lg:grid-cols-[210px_minmax(0,1fr)] lg:gap-3 lg:px-8 lg:py-3 xl:grid-cols-[224px_minmax(0,1fr)]">
        <aside
          className="sticky top-16 -my-3 hidden h-[calc(100vh-4rem)] min-w-0 self-start overflow-y-auto py-3 pr-3 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden lg:block"
          aria-label="Sidebar"
        >
          {navigation}
        </aside>
        <section className="min-w-0">{children}</section>
      </main>
    </div>
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
  labels: { resources: string; categories: string };
  onNavigate(): void;
}) {
  return (
    <nav className="flex flex-col gap-3">
      <NavList items={primary} onNavigate={onNavigate} />
      <NavSection title={labels.resources}>
        <NavList items={resources} onNavigate={onNavigate} />
      </NavSection>
      {groups.map((group) => (
        <NavSection key={group.key} title={group.title}>
          <NavList items={group.items} onNavigate={onNavigate} />
        </NavSection>
      ))}
      {categories.length ? (
        <NavSection title={labels.categories}>
          <NavList items={categories} onNavigate={onNavigate} />
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
  items,
  onNavigate,
}: {
  items: ShellNavItem[];
  onNavigate(): void;
}) {
  return (
    <ul className="flex flex-col gap-0.5">
      {items.map((item) => {
        const Icon = item.icon;
        return (
          <li key={item.key}>
            <Button
              asChild
              variant="ghost"
              className={cn(
                "h-8 w-full justify-start text-[13px] leading-[18.5714px]",
                item.active && "bg-primary/10 text-primary",
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
