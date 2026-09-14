import { useEffect, useState, type FormEvent } from "react";
import type { SearchPageProps } from "@gooseforum/client";
import { Search, UsersRound } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "../../components/ui/empty";
import { Input } from "../../components/ui/input";
import { GooseLink, useGooseRuntime } from "../../runtime";
import { PageHeader } from "../layout/page-header";
import { compactNumber, TopicTable } from "../topics/topic-list";

export function SearchPageView({ page }: { page: SearchPageProps }) {
  const { t: searchT } = useTranslation("search");
  const { t: topicT } = useTranslation("home");
  const runtime = useGooseRuntime();
  const [query, setQuery] = useState(page.query);
  const hasQuery = Boolean(page.query.trim());

  useEffect(() => setQuery(page.query), [page.query]);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const params = new URLSearchParams();
    if (query.trim()) params.set("q", query.trim());
    void runtime.navigate(`/search${params.size ? `?${params}` : ""}`);
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
            className="w-full sm:w-80 lg:w-96"
            onSubmit={submit}
          >
            <div className="flex h-10 items-center gap-2 rounded-md border bg-background px-3 text-sm text-muted-foreground transition focus-within:border-primary focus-within:ring-4 focus-within:ring-primary/20">
              <Search className="size-4 shrink-0" />
              <label htmlFor="site-search-query" className="sr-only">
                {searchT("inputPlaceholder")}
              </label>
              <Input
                id="site-search-query"
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
      <section className="overflow-hidden rounded-xl border bg-background max-sm:rounded-t-none max-sm:border-t-0">
        {page.topics.length ? (
          <>
            <TopicTable topics={page.topics} t={topicT} />
            {page.totalPages > 1 ? (
              <footer className="flex flex-col gap-3 border-t bg-muted/50 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
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
                        href={`/search?q=${encodeURIComponent(page.query)}&page=${page.pagination.page - 1}`}
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
      </section>
    </main>
  );
}
