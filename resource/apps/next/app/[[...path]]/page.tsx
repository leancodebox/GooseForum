import type { Metadata } from "next";
import { headers } from "next/headers";
import {
  cachedGooseResources,
  goosePageNamespaces,
  prepareGooseTranslations,
} from "@gooseforum/runtime/i18n/loader";
import { NextGooseApp } from "@/next-goose-app";
import { loadNextPage } from "@/page-adapter";
import { pagePath, resolveLocale } from "@/request-context";

interface RouteProps {
  params: Promise<{ path?: string[] }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

export async function generateMetadata(props: RouteProps): Promise<Metadata> {
  const request = await resolveRequest(props);
  const page = await loadNextPage(
    request.path,
    request.cookie,
    request.acceptLanguage,
  );
  return {
    title: page.meta.title,
    description: page.meta.description,
    robots: page.meta.robots,
    alternates: page.meta.canonical
      ? { canonical: page.meta.canonical }
      : undefined,
    openGraph: page.meta.openGraph
      ? {
          title: page.meta.openGraph.title,
          description: page.meta.openGraph.description,
          type: "website",
          url: page.meta.openGraph.url,
          siteName: page.meta.openGraph.siteName,
          images: page.meta.openGraph.image
            ? [page.meta.openGraph.image]
            : undefined,
        }
      : undefined,
  };
}

export default async function GoosePage(props: RouteProps) {
  const request = await resolveRequest(props);
  const page = await loadNextPage(
    request.path,
    request.cookie,
    request.acceptLanguage,
  );
  const locale = resolveLocale(
    request.path,
    request.cookie,
    request.acceptLanguage,
  );
  await prepareGooseTranslations(locale, goosePageNamespaces(page.component));

  return (
    <NextGooseApp
      page={page}
      locale={locale}
      initialResources={cachedGooseResources()}
    />
  );
}

async function resolveRequest({ params, searchParams }: RouteProps) {
  const [route, search, requestHeaders] = await Promise.all([
    params,
    searchParams,
    headers(),
  ]);
  return {
    path: pagePath(route.path, search),
    cookie: requestHeaders.get("cookie") || "",
    acceptLanguage: requestHeaders.get("accept-language") || "",
  };
}
