"use client";

import { useMemo } from "react";
import {
  createGooseClient,
  type AnyPagePayload,
} from "@gooseforum/client";
import type { Locale } from "@gooseforum/client/i18n/locale";
import { SiteApp } from "@gooseforum/theme-default/app/site-app";
import type { PageSource } from "@gooseforum/runtime/page-source";
import { cachedGooseResources } from "@gooseforum/runtime/i18n/loader";

type InitialResources = ReturnType<typeof cachedGooseResources>;

export function NextGooseApp({
  page,
  locale,
  initialResources,
}: {
  page: AnyPagePayload;
  locale: Locale;
  initialResources: InitialResources;
}) {
  const pageSource = useMemo<PageSource<AnyPagePayload>>(() => {
    const client = createGooseClient<AnyPagePayload>();
    return {
      api: client.api,
      load(url, signal) {
        const path = `${url.pathname}${url.search}`;
        return client.pages.fetch(
          `/goose-page-data?path=${encodeURIComponent(path)}`,
          { signal },
        );
      },
    };
  }, []);

  return (
    <SiteApp
      pageSource={pageSource}
      initialPage={page}
      initialLocale={locale}
      initialTheme={page.layout.theme.current}
      initialResources={initialResources}
    />
  );
}
