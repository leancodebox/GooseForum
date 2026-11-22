import { useRef, useState } from "react";
import { flushSync } from "react-dom";
import { SearchIcon, XIcon } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useGooseRuntime } from "@gooseforum/runtime";
import { Button } from "@gooseforum/ui/components/button";
import { Input } from "@gooseforum/ui/components/input";
import { Popover, PopoverContent, PopoverTrigger } from "@gooseforum/ui/components/popover";
import { searchUrl } from "../search-url";

const innerButtonFocus = "focus-visible:ring-inset focus-visible:ring-muted-foreground/40";

export function HeaderSearch() {
  const { t } = useTranslation("shell");
  const { navigate } = useGooseRuntime();
  const [open, setOpen] = useState(false);
  const input = useRef<HTMLInputElement>(null);

  function changeOpen(nextOpen: boolean) {
    if (!nextOpen) {
      setOpen(false);
      return;
    }
    // Focus in the same user gesture so mobile browsers can open the keyboard.
    flushSync(() => setOpen(true));
    input.current?.focus({ preventScroll: true });
  }

  return (
    <Popover open={open} onOpenChange={changeOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon-lg"
          className="shrink-0 [&_svg]:size-5"
          aria-label={t("openSearch")}
        >
          <SearchIcon />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        align="end"
        sideOffset={8}
        collisionPadding={12}
        aria-label={t("search")}
        className="w-[min(20rem,calc(100vw-1.5rem))] p-2"
        onEscapeKeyDown={(event) => {
          if (event.isComposing) event.preventDefault();
        }}
      >
        <form
          action="/search"
          method="get"
          role="search"
          className="flex h-9 min-w-0 items-center overflow-hidden rounded-lg border border-input bg-background px-0.5 focus-within:border-muted-foreground/40"
          onSubmit={(event) => {
            event.preventDefault();
            const query = input.current?.value.trim();
            if (!query) {
              input.current?.focus();
              return;
            }
            setOpen(false);
            void navigate(searchUrl(query));
          }}
        >
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
            onClick={() => setOpen(false)}
          >
            <XIcon />
          </Button>
          <Button
            type="submit"
            variant="ghost"
            size="icon-sm"
            className={innerButtonFocus}
            aria-label={t("search")}
          >
            <SearchIcon />
          </Button>
        </form>
      </PopoverContent>
    </Popover>
  );
}
