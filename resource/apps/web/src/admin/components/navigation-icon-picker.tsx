import { useEffect, useMemo, useRef, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@gooseforum/ui/components/button";
import { Input } from "@gooseforum/ui/components/input";
import DynamicNavigationIcon, { navigationIconNames } from "@gooseforum/ui/components/navigation-icon-dynamic";
import type { SettingsTextKey } from "../settings-i18n";

const commonIcons = [
  "link", "globe", "book-open", "message-circle", "users", "house", "code", "git-fork",
  "star", "heart", "external-link", "rss", "file-text", "circle-help", "mail", "bookmark",
];
const searchKey = (value: string) => value.toLowerCase().replace(/[\s_-]+/g, "");
const iconNames = new Set<string>(navigationIconNames);
const catalog = [
  ...commonIcons.filter(name => iconNames.has(name)),
  ...[...iconNames].sort().filter(name => !commonIcons.includes(name)),
].map(name => ({ name, key: searchKey(name) }));
const pageSize = 40;

export default function NavigationIconPicker({ value, onSelect, text }: {
  value?: string;
  onSelect(name?: string): void;
  text(key: SettingsTextKey): string;
}) {
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(0);
  const searchInput = useRef<HTMLInputElement>(null);
  const grid = useRef<HTMLDivElement>(null);
  const search = searchKey(query);
  useEffect(() => {
    searchInput.current?.focus();
  }, []);
  useEffect(() => {
    if (grid.current) grid.current.scrollTop = 0;
  }, [search, page]);
  const matches = useMemo(() => search ? catalog.filter(icon => icon.key.includes(search)) : catalog, [search]);
  const pageCount = Math.ceil(matches.length / pageSize);

  return (
    <div className="flex min-h-0 min-w-0 flex-col gap-3">
      <Input
        ref={searchInput}
        type="search"
        aria-label={text("searchMenuIcons")}
        placeholder={text("searchMenuIcons")}
        value={query}
        onChange={event => {
          setQuery(event.target.value);
          setPage(0);
        }}
      />
      <Button type="button" size="sm" variant="ghost" aria-pressed={!value}
        className="self-start" onClick={() => onSelect()}>
        {text("defaultMenuIcon")}
      </Button>
      {matches.length ? (
        <div ref={grid} className="grid min-h-0 grid-cols-4 content-start gap-1 overflow-y-auto p-1 sm:grid-cols-8">
          {matches.slice(page * pageSize, (page + 1) * pageSize).map(({ name }) => (
            <Button
              key={name}
              type="button"
              variant={value === name ? "secondary" : "ghost"}
              className="h-14 min-w-0 flex-col gap-1 px-1 text-[11px] [&_svg]:size-5"
              aria-pressed={value === name}
              title={name}
              onClick={() => onSelect(name)}
            >
              <DynamicNavigationIcon name={name} />
              <span className="max-w-full truncate">{name}</span>
            </Button>
          ))}
        </div>
      ) : <p role="status" className="px-2 py-1 text-muted-foreground">{text("noMenuIcons")}</p>}
      {pageCount > 1 && (
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <Button type="button" variant="ghost" size="icon-sm" aria-label={text("previousIconPage")}
            disabled={page === 0} onClick={() => setPage(page - 1)}><ChevronLeft /></Button>
          <span>{page + 1} / {pageCount}</span>
          <Button type="button" variant="ghost" size="icon-sm" aria-label={text("nextIconPage")}
            disabled={page + 1 >= pageCount} onClick={() => setPage(page + 1)}><ChevronRight /></Button>
        </div>
      )}
    </div>
  );
}
