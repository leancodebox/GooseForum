import { useEffect, useRef } from "react";
import { flushSync } from "react-dom";
import { SearchIcon, XIcon } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useGooseRuntime } from "@gooseforum/runtime";
import { Button } from "@gooseforum/ui/components/button";
import { Input } from "@gooseforum/ui/components/input";
import { cn } from "@gooseforum/ui/lib/utils";
import { searchUrl } from "../search-url";

const innerButtonFocus = "focus-visible:ring-inset focus-visible:ring-muted-foreground/40";

export function HeaderSearch({ open, onOpenChange }: {
  open: boolean;
  onOpenChange(open: boolean): void;
}) {
  const { t } = useTranslation("shell");
  const { navigate } = useGooseRuntime();
  const form = useRef<HTMLFormElement>(null);
  const input = useRef<HTMLInputElement>(null);
  const searchButton = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (!open) return;
    function dismissOutside(event: PointerEvent) {
      if (event.target instanceof Node && !form.current?.contains(event.target)) {
        onOpenChange(false);
      }
    }
    document.addEventListener("pointerdown", dismissOutside);
    return () => document.removeEventListener("pointerdown", dismissOutside);
  }, [open, onOpenChange]);

  function openSearch() {
    // Focus in the same user gesture so mobile browsers can open the keyboard.
    flushSync(() => onOpenChange(true));
    input.current?.focus({ preventScroll: true });
  }

  function closeSearch() {
    searchButton.current?.focus({ preventScroll: true });
    onOpenChange(false);
  }

  return (
    <form
      ref={form}
      action="/search"
      method="get"
      role={open ? "search" : undefined}
      aria-label={open ? t("search") : undefined}
      className={cn(
        "flex h-9 min-w-0 items-center rounded-lg",
        open
          ? "w-80 overflow-hidden border border-input bg-background px-0.5 transition-[width] duration-200 focus-within:border-muted-foreground/40 motion-reduce:transition-none"
          : "w-9 shrink-0",
      )}
      onSubmit={(event) => {
        event.preventDefault();
        const query = input.current?.value.trim();
        if (!query) {
          input.current?.focus();
          return;
        }
        closeSearch();
        void navigate(searchUrl(query));
      }}
      onKeyDown={(event) => {
        if (open && event.key === "Escape" && !event.nativeEvent.isComposing) {
          event.preventDefault();
          event.stopPropagation();
          closeSearch();
        }
      }}
      onBlur={(event) => {
        if (open && event.relatedTarget && !event.currentTarget.contains(event.relatedTarget)) {
          onOpenChange(false);
        }
      }}
    >
      {open && (
        <>
          <Input
            ref={input}
            type="search"
            name="q"
            aria-label={t("searchPlaceholder")}
            placeholder={t("searchPlaceholder")}
            autoComplete="off"
            enterKeyHint="search"
            className="h-full flex-1 rounded-none border-0 bg-transparent pl-2.5 pr-0 shadow-none focus-visible:ring-0 dark:bg-transparent [&::-webkit-search-cancel-button]:appearance-none"
          />
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className={innerButtonFocus}
            aria-label={t("closeSearch")}
            onClick={closeSearch}
          >
            <XIcon />
          </Button>
        </>
      )}
      <Button
        ref={searchButton}
        type={open ? "submit" : "button"}
        variant="ghost"
        size={open ? "icon-sm" : "icon-lg"}
        className={cn("[&_svg]:size-5", open && innerButtonFocus)}
        aria-label={t(open ? "search" : "openSearch")}
        aria-expanded={open ? undefined : false}
        onClick={open ? undefined : openSearch}
      >
        <SearchIcon />
      </Button>
    </form>
  );
}
