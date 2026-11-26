import { useEffect, useEffectEvent, useState, type ReactNode } from "react";
import { MenuIcon, XIcon } from "lucide-react";
import { Button } from "./button";
import { cn } from "../lib/utils";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "./sheet";

export function useShellSidebar({ keyboardShortcut = false } = {}) {
  const [mobileOpen, setMobileOpen] = useState(false);
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
  const toggleFromKeyboard = useEffectEvent(() => {
    if (window.matchMedia("(min-width: 64rem)").matches) toggleSidebar();
    else setMobileOpen(open => !open);
  });
  useEffect(() => {
    if (!keyboardShortcut) return;
    function handleKey(event: KeyboardEvent) {
      if (event.key === "b" && (event.metaKey || event.ctrlKey)) {
        event.preventDefault();
        toggleFromKeyboard();
      }
    }
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, [keyboardShortcut]);

  return { mobileOpen, setMobileOpen, sidebarCollapsed, sidebarAnimate, toggleSidebar };
}

type SidebarState = ReturnType<typeof useShellSidebar>;

export function ShellMenuTrigger({ sidebar, navigation, t }: {
  sidebar: SidebarState;
  navigation: ReactNode;
  t(key: string): string;
}) {
  const { mobileOpen, setMobileOpen, sidebarCollapsed, toggleSidebar } = sidebar;
  return (
    <>
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

      <Button variant="ghost" size="icon" className="hidden shrink-0 lg:inline-flex"
        aria-label={sidebarCollapsed ? t("expandSidebar") : t("collapseSidebar")}
        aria-expanded={!sidebarCollapsed} aria-controls="goose-desktop-sidebar" onClick={toggleSidebar}>
        <MenuIcon className="size-5.5" />
      </Button>
    </>
  );
}

export function ShellGrid({ as: Container = "main", sidebar, standalone = false, navigation, children }: {
  sidebar: SidebarState;
  as?: "main" | "div";
  standalone?: boolean;
  navigation: ReactNode;
  children: ReactNode;
}) {
  const { sidebarCollapsed, sidebarAnimate } = sidebar;
  return (
    <Container data-sidebar-collapsed={sidebarCollapsed} data-sidebar-animate={sidebarAnimate}
      className={standalone ? "w-full" : "goose-shell-grid mx-auto grid w-full max-w-[1600px] grid-cols-1 lg:px-6 lg:py-3"}>
      <aside
        id="goose-desktop-sidebar"
        inert={sidebarCollapsed || standalone} aria-hidden={sidebarCollapsed || standalone}
        className={cn(standalone ? "hidden" : "sticky top-16 -my-3 hidden h-[calc(100vh-4rem)] min-w-0 self-start overflow-x-hidden overflow-y-auto py-3 pl-1 pr-3 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden lg:block", sidebarCollapsed && "invisible")}
        aria-label="Sidebar"
      >
        <div className="w-[194px] xl:w-[208px]">{navigation}</div>
      </aside>
      <section data-slot="goose-page-content" className="min-w-0">
        {children}
      </section>
    </Container>
  );
}
