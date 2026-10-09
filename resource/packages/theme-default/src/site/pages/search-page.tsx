import { useEffect, useId, useRef, useState, type FormEvent } from "react";
import type { SearchPageProps } from "@gooseforum/client";
import { Search, UsersRound } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Badge } from "@gooseforum/ui/components/badge";
import { Button } from "@gooseforum/ui/components/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import { Input } from "@gooseforum/ui/components/input";
import { ListFilter, ListFilterItem } from "@gooseforum/ui/components/list-filter";
import { GooseLink, useGooseRuntime } from "@gooseforum/runtime";
import { PageHeader } from "../layout/page-header";
import { SiteListPanel } from "../layout/site-panel";
import { compactNumber, TopicListToolbar, TopicTable } from "../topics/topic-list";

export function SearchPageView({ page }: { page: SearchPageProps }) {
  const queryInputId = useId();
  const { t: searchT } = useTranslation("search");
  const { t: topicT } = useTranslation("home");
  const runtime = useGooseRuntime();
  const [query, setQuery] = useState(page.query);
  const payloadQuery = useRef(page.query);
  const hasQuery = Boolean(page.query.trim());
  const sort = page.sort;

  function searchUrl(query: string, sort: string, page = 1) {
    const params = new URLSearchParams();
    if (query.trim()) params.set("q", query.trim());
    if (sort !== "relevance") params.set("sort", sort);
    if (page > 1) params.set("page", String(page));
    return `/search${params.size ? `?${params}` : ""}`;
  }

  useEffect(() => {
    if (payloadQuery.current === page.query) return;
    payloadQuery.current = page.query;
    setQuery(page.query);
  }, [page.query]);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void runtime.navigate(searchUrl(query, sort));
  }

  const description = hasQuery
    ? `${page.query} · ${searchT("resultCount", { count: compactNumber(page.total) })}`
    : searchT("emptyPrompt");

  return (
    <main className="min-w-0 pb-8">
      <PageHeader
        compact
        title={searchT("title")}
        description={description}
        badge={
          <Badge variant="secondary" className="h-5 text-[11px] uppercase">
            {searchT("label")}
          </Badge>
        }
        actions={
          <form
            action="/search"
            method="get"
            className="w-full lg:w-96"
            onSubmit={submit}
          >
            <input type="hidden" name="sort" value={sort} />
            <div className="flex h-10 items-center gap-2 rounded-md border bg-background px-3 text-sm text-muted-foreground transition focus-within:border-primary focus-within:ring-4 focus-within:ring-primary/20">
              <Search className="size-4 shrink-0" />
              <label htmlFor={queryInputId} className="sr-only">
                {searchT("inputPlaceholder")}
              </label>
              <Input
                id={queryInputId}
                name="q"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                className="h-auto min-w-0 flex-1 rounded-none border-0 bg-transparent px-0 py-0 text-foreground shadow-none placeholder:text-muted-foreground focus-visible:ring-0"
                placeholder={searchT("inputPlaceholder")}
              />
              <Button
                type="submit"
                variant="secondary"
                size="sm"
                className="shrink-0"
              >
                {searchT("action")}
              </Button>
            </div>
          </form>
        }
      />
      <SiteListPanel>
        <TopicListToolbar>
          <ListFilter asChild>
            <nav aria-label={searchT("sort")}>
              {(["relevance", "new", "updated"] as const).map((key) => (
                <ListFilterItem key={key} asChild>
                  <GooseLink
                    href={searchUrl(page.query, key)}
                    aria-current={sort === key ? "page" : undefined}
                  >
                    {searchT(key)}
                  </GooseLink>
                </ListFilterItem>
              ))}
            </nav>
          </ListFilter>
        </TopicListToolbar>
        {page.topics.length ? (
          <>
            <TopicTable topics={page.topics} t={topicT} />
            {page.totalPages > 1 ? (
              <footer className="flex flex-col gap-3 border-t bg-muted/50 px-4 py-3 lg:flex-row lg:items-center lg:justify-between">
                <div className="text-sm text-muted-foreground">
                  {searchT("page", {
                    page: page.pagination.page,
                    total: page.totalPages,
                  })}
                </div>
                <div className="flex items-center gap-2">
                  {page.pagination.page > 1 ? (
                    <Button asChild variant="outline" size="sm">
                      <GooseLink
                        href={searchUrl(page.query, sort, page.pagination.page - 1)}
                        rel="prev"
                      >
                        {topicT("previous")}
                      </GooseLink>
                    </Button>
                  ) : null}
                  {page.pagination.hasNext ? (
                    <Button asChild variant="outline" size="sm">
                      <GooseLink href={page.pagination.nextUrl} rel="next">
                        {topicT("next")}
                      </GooseLink>
                    </Button>
                  ) : null}
                </div>
              </footer>
            ) : null}
          </>
        ) : (
          <Empty className="min-h-56 border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                {hasQuery ? <UsersRound /> : <Search />}
              </EmptyMedia>
              <EmptyTitle>
                {searchT(hasQuery ? "noResultsTitle" : "startTitle")}
              </EmptyTitle>
              <EmptyDescription>
                {searchT(
                  hasQuery ? "noResultsDescription" : "startDescription",
                )}
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </SiteListPanel>
    </main>
  );
}
