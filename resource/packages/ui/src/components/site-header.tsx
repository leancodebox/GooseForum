import type { SimpleKey } from "@gooseforum/i18n";
import { useEffect, useRef, useState, Suspense, type ComponentType, type ReactNode, type AnchorHTMLAttributes } from "react";
import { BellIcon, FileTextIcon, InboxIcon, LanguagesIcon, LogOutIcon, PenSquareIcon, SettingsIcon, ShieldIcon, UserRoundIcon } from "lucide-react";
import { DefaultBrand } from "./default-brand";
import { avatarInitial } from "../lib/avatar-initial";
import { cn } from "../lib/utils";
import { NavigationIconContent } from "./navigation-icon";
import { Button } from "./button";
import { Avatar, AvatarImage, AvatarFallback } from "./avatar";
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuRadioGroup, DropdownMenuRadioItem, DropdownMenuSeparator, DropdownMenuTrigger } from "./dropdown-menu";
import { ThemeMenuOptions, ThemeMenuTrigger, type ThemePreference } from "./theme-menu";
import { HeaderSearch } from "./header-search";

type HeaderItem = { key: string; label: string; url: string; icon?: string; i18nLabel?: string };
interface SiteHeaderProps<L extends string> {
  site: { name: string; brandType: string; brandText: string; brandImage: string };
  viewer: { id: number; username: string; nickname?: string; avatarUrl: string; isAuthenticated: boolean; canAccessAdmin: boolean };
  header?: HeaderItem[];
  unread: { messages: boolean; notifications: boolean };
  sidebarTrigger: ReactNode;
  contextTitle?: ReactNode;
  hidden?: boolean;
  homeLabel?: string;
  Link: ComponentType<AnchorHTMLAttributes<HTMLAnchorElement> & { href: string }>;
  t(key: SimpleKey<"shared.shell">): string;
  navLabel(key: string, fallback: string): string;
  locale: L;
  locales: readonly L[];
  localeLabels: Record<L, { label: string }>;
  onLocaleChange(locale: L): void;
  themePreference: ThemePreference;
  onThemeChange(value: ThemePreference): void;
  onSearch(query: string): void;
  onLogout(): void | Promise<void>;
}
const headerMenuKeys = ["theme", "language", "user"] as const;
type HeaderMenu = (typeof headerMenuKeys)[number];

export function SiteHeader<L extends string>({ site, viewer, header, unread, sidebarTrigger, contextTitle, hidden, homeLabel, Link, t, navLabel, locale, locales, localeLabels, onLocaleChange, themePreference, onThemeChange, onSearch, onLogout }: SiteHeaderProps<L>) {
  const headerItems = header ?? [
    { key: "links", url: "/links", label: t("nav.links") },
    { key: "sponsors", url: "/sponsors", label: t("nav.sponsors") },
  ];
  const [themeMenuOpen, setThemeMenuOpen] = useState(false);
  const [languageMenuOpen, setLanguageMenuOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const menuCloseTimers = useRef<
    Record<HeaderMenu, number | undefined>
  >({
    theme: undefined,
    language: undefined,
    user: undefined,
  });
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

  useEffect(
    () => () => {
      window.clearTimeout(menuCloseTimers.current.theme);
      window.clearTimeout(menuCloseTimers.current.language);
      window.clearTimeout(menuCloseTimers.current.user);
    },
    [],
  );

  const brand =
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
    <header hidden={hidden} className="sticky top-0 z-40 border-b bg-background/95 backdrop-blur-sm">
      <div className="mx-auto flex h-16 w-full max-w-[1600px] items-center gap-2 px-3 lg:gap-3 lg:px-6">
        {sidebarTrigger}
        <div className={cn(contextTitle && "hidden md:block")}>
          <Link href="/" aria-label={homeLabel} className="shrink-0 text-xl font-semibold tracking-tight text-primary lg:text-2xl">{brand}</Link>
        </div>
        {contextTitle}
        {headerItems.length > 0 && (
          <Suspense fallback={null}>
            <nav
              className={cn(
                "hidden items-center gap-1 lg:flex",
                Boolean(contextTitle) && "lg:hidden",
              )}
              aria-label="Header navigation"
            >
              {headerItems.map((item) => (
                <Link
                  key={item.key}
                  href={item.url}
                  className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm font-medium text-foreground/75 hover:bg-accent"
                >
                  {item.icon && item.key !== "links" && item.key !== "sponsors" && (
                    <span className="size-4 shrink-0"><NavigationIconContent name={item.icon} /></span>
                  )}
                  {item.i18nLabel?.startsWith("shell.") ? navLabel(item.i18nLabel.slice(6), item.label) : item.label}
                </Link>
              ))}
            </nav>
          </Suspense>
        )}

        <div className="ml-auto flex min-w-0 items-center justify-end gap-0.5 lg:gap-1">
          <HeaderSearch t={t} onSearch={onSearch} />
          <div
            className={cn(
              "shrink-0 items-center gap-0.5 lg:gap-1",
              contextTitle ? "hidden md:flex" : "flex",
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
                <ThemeMenuTrigger size="icon-lg" preference={themePreference} label={t("chooseTheme")} />
              </div>
              <DropdownMenuContent
                className="w-max min-w-32 whitespace-nowrap"
                align="end"
                sideOffset={8}
                onMouseEnter={() => openHeaderMenuOnHover("theme")}
                onMouseLeave={() => closeHeaderMenuSoon("theme")}
                onCloseAutoFocus={(event) => event.preventDefault()}
              >
                <ThemeMenuOptions preference={themePreference} onChange={onThemeChange} t={t} />
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
                  value={locale}
                  onValueChange={(value) =>
                    onLocaleChange(value as L)
                  }
                >
                  {locales.map((locale) => (
                    <DropdownMenuRadioItem key={locale} value={locale}>
                      {localeLabels[locale].label}
                    </DropdownMenuRadioItem>
                  ))}
                </DropdownMenuRadioGroup>
              </DropdownMenuContent>
            </DropdownMenu>

            {viewer.isAuthenticated ? (
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
                      aria-label={viewer.username}
                    >
                      <Avatar className="size-9">
                        <AvatarImage
                          src={viewer.avatarUrl}
                          alt={viewer.username}
                        />
                        <AvatarFallback userId={viewer.id}>
                          {avatarInitial(viewer.nickname, viewer.username)}
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
                      <Link href={`/u/${viewer.id}`}>
                        <UserRoundIcon />
                        {t("profile")}
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <Link href="/messages">
                        <InboxIcon />
                        {t("nav.messages")}
                        {unread.messages ? <MenuDot /> : null}
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <Link href="/notifications">
                        <BellIcon />
                        {t("nav.notifications")}
                        {unread.notifications ? <MenuDot /> : null}
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <Link href="/drafts">
                        <FileTextIcon />
                        {t("nav.drafts")}
                      </Link>
                    </DropdownMenuItem>
                  </DropdownMenuGroup>
                  <DropdownMenuSeparator />
                  <DropdownMenuGroup>
                    <DropdownMenuItem asChild>
                      <Link href="/publish">
                        <PenSquareIcon />
                        {t("publish")}
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <Link href="/settings">
                        <SettingsIcon />
                        {t("settings")}
                      </Link>
                    </DropdownMenuItem>
                    {viewer.canAccessAdmin ? (
                      <DropdownMenuItem asChild>
                        <Link href="/admin">
                          <ShieldIcon />
                          {t("admin")}
                        </Link>
                      </DropdownMenuItem>
                    ) : null}
                  </DropdownMenuGroup>
                  <DropdownMenuSeparator />
                  <DropdownMenuGroup>
                    <DropdownMenuItem
                      variant="destructive"
                      onSelect={() => void onLogout()}
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
                  <Link href="/login">{t("login")}</Link>
                </Button>
                <Button asChild className="hidden lg:inline-flex">
                  <Link href="/login?register=true">
                    {t("register")}
                  </Link>
                </Button>
              </>
            )}
          </div>
        </div>
      </div>
    </header>
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
