"use client";

import {
  Activity,
  startTransition,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { AnyPagePayload } from "@gooseforum/client";
import type { Locale } from "@gooseforum/client/i18n/locale";
import type { Resource } from "i18next";
import { GoosePage, isStandalonePage } from "./root";
import { AppShell } from "../site/layout/app-shell";
import { BootstrapError, BootstrapLoading } from "@gooseforum/ui/bootstrap";
import { prepareGoosePage } from "@gooseforum/runtime/prepared-page";
import {
  GooseRuntimeProvider,
  isSiteSpaPath,
  type GooseNavigateOptions,
  type GooseRuntime,
} from "@gooseforum/runtime";
import {
  applyBrowserLocale,
  applyBrowserTheme,
  detectBrowserTheme,
  detectBrowserLocale,
  prepareBrowserDocument as prepareDocument,
  queueBrowserFlash,
} from "@gooseforum/runtime/browser-host";
import type { PageSource } from "@gooseforum/runtime/page-source";
import { GooseI18nProvider, prepareGooseTranslations, goosePageNamespaces, loadedGooseNamespaces } from "@gooseforum/runtime/i18n";

export interface SiteAppProps {
  pageSource: PageSource<AnyPagePayload>;
  initialPage?: AnyPagePayload;
  initialLocale?: Locale;
  initialTheme?: AnyPagePayload["layout"]["theme"]["current"];
  initialResources?: Resource;
  errorDescription?: string;
}

interface CachedPage {
  key: string;
  page: AnyPagePayload;
}

type ScrollAction = "preserve" | "reset" | "restore";

interface PendingScroll {
  page: AnyPagePayload;
  action: Exclude<ScrollAction, "preserve">;
  entryKey: string;
  hash: string;
}

const keepAliveComponents = new Set([
  "home.index",
  "category.index",
  "search.index",
]);
const maxCachedPages = 10;
const historyEntryKey = "gooseEntryKey";
const historyEntryIndex = "gooseEntryIndex";

export function SiteApp({
  pageSource,
  initialPage,
  initialLocale,
  initialTheme,
  initialResources,
  errorDescription,
}: SiteAppProps) {
  const [page, setPage] = useState<AnyPagePayload | undefined>(initialPage);
  const [cachedPages, setCachedPages] = useState<CachedPage[]>(() =>
    initialPage && keepAliveComponents.has(initialPage.component)
      ? [{ key: routeCacheKey(new URL(initialPage.url, browserHref())), page: initialPage }]
      : [],
  );
  const [error, setError] = useState<unknown>();
  const [isNavigating, setIsNavigating] = useState(false);
  const [locale, setLocaleState] = useState<Locale>(() =>
    initialLocale ?? detectBrowserLocale(),
  );
  const localeRef = useRef(locale);
  const localeVersion = useRef(0);
  useEffect(() => () => { localeVersion.current++; }, []);
  const [theme, setTheme] = useState(() =>
    initialTheme ?? detectBrowserTheme(),
  );
  const themeRef = useRef(theme);
  const activeRequest = useRef<AbortController | null>(null);
  const navigationBlocker = useRef<
    ((href: string) => boolean | Promise<boolean>) | null
  >(null);
  const cachedPageRef = useRef(new Map(cachedPages.map(entry => [entry.key, entry.page])));
  const scrollPositions = useRef(
    new Map<string, { left: number; top: number }>(),
  );
  const pendingScroll = useRef<PendingScroll | null>(null);
  const activeHistoryEntry = useRef("");
  const activeHistoryIndex = useRef(0);
  const suppressNextPop = useRef(false);

  const rememberPage = useCallback((nextPage: AnyPagePayload) => {
    if (!keepAliveComponents.has(nextPage.component)) return;
    const key = routeCacheKey(new URL(nextPage.url, window.location.href));
    setCachedPages((current) => {
      const next = [
        ...current.filter((entry) => entry.key !== key),
        { key, page: nextPage },
      ].slice(-maxCachedPages);
      cachedPageRef.current = new Map(
        next.map((entry) => [entry.key, entry.page]),
      );
      return next;
    });
  }, []);

  const handleContentReady = useCallback((committedPage: AnyPagePayload) => {
    const target = pendingScroll.current;
    if (!target || target.page !== committedPage) return;
    pendingScroll.current = null;
    applyScroll(
      target.action,
      target.entryKey,
      target.hash,
      scrollPositions.current,
    );
  }, []);

  const loadPage = useCallback(
    async (
      url: URL,
      historyAction: "push" | "replace" | "none",
      scrollAction: ScrollAction = "preserve",
      targetHistoryEntry?: string,
      targetHistoryIndex?: number,
    ) => {
      activeRequest.current?.abort();
      const request = new AbortController();
      activeRequest.current = request;
      setIsNavigating(true);

      try {
        const nextPage = await pageSource.load(url, request.signal);
        if (request.signal.aborted) return;
        const language = localeRef.current;
        await Promise.all([
          prepareGoosePage(nextPage.component),
          prepareGooseTranslations(language, goosePageNamespaces(nextPage.component)),
        ]);
        if (language !== localeRef.current) await prepareGooseTranslations(localeRef.current, goosePageNamespaces(nextPage.component));
        if (request.signal.aborted) return;

        const nextUrl = browserUrlForPage(nextPage.url, url);
        let nextHistoryEntry =
          targetHistoryEntry ||
          activeHistoryEntry.current ||
          ensureHistoryEntry();
        let nextHistoryIndex = targetHistoryIndex ?? activeHistoryIndex.current;
        if (historyAction === "push") {
          saveScrollPosition(
            scrollPositions.current,
            activeHistoryEntry.current,
          );
          nextHistoryEntry = createHistoryEntryKey();
          nextHistoryIndex = activeHistoryIndex.current + 1;
          window.history.pushState(
            withHistoryEntry(
              window.history.state,
              nextHistoryEntry,
              nextHistoryIndex,
            ),
            "",
            nextUrl,
          );
        } else if (
          historyAction === "replace" ||
          nextUrl.href !== window.location.href
        ) {
          window.history.replaceState(
            withHistoryEntry(
              window.history.state,
              nextHistoryEntry,
              nextHistoryIndex,
            ),
            "",
            nextUrl,
          );
        }
        activeHistoryEntry.current = nextHistoryEntry;
        activeHistoryIndex.current = nextHistoryIndex;

        prepareDocument(nextPage, themeRef.current);
        setError(undefined);
        pendingScroll.current = scrollAction === "preserve"
          ? null
          : {
              page: nextPage,
              action: scrollAction,
              entryKey: nextHistoryEntry,
              hash: url.hash,
            };
        startTransition(() => {
          rememberPage(nextPage);
          setPage(nextPage);
        });
      } catch (nextError) {
        if (!request.signal.aborted) setError(nextError);
      } finally {
        if (activeRequest.current === request) {
          activeRequest.current = null;
          setIsNavigating(false);
        }
      }
    },
    [pageSource, rememberPage],
  );

  const navigate = useCallback(
    async (href: string, options: GooseNavigateOptions = {}) => {
      if (
        navigationBlocker.current &&
        !(await navigationBlocker.current(href))
      ) {
        return;
      }
      const url = new URL(href, window.location.href);
      if (url.origin !== window.location.origin || !isSiteSpaPath(url.pathname)) {
        window.location.assign(url);
        return;
      }
      await loadPage(url, options.replace ? "replace" : "push", "reset");
    },
    [loadPage],
  );

  const refresh = useCallback(async () => {
    await loadPage(new URL(window.location.href), "none");
  }, [loadPage]);

  const registerNavigationBlocker = useCallback(
    (blocker: (href: string) => boolean | Promise<boolean>) => {
      navigationBlocker.current = blocker;
      return () => {
        if (navigationBlocker.current === blocker)
          navigationBlocker.current = null;
      };
    },
    [],
  );

  const setLocale = useCallback(async (nextLocale: Locale) => {
    const version = ++localeVersion.current;
    try {
      await prepareGooseTranslations(nextLocale, loadedGooseNamespaces());
      // A concurrent route load may have added another namespace in the meantime.
      await prepareGooseTranslations(nextLocale, loadedGooseNamespaces());
      if (version !== localeVersion.current) return;
      localeRef.current = nextLocale;
      applyBrowserLocale(nextLocale);
      setLocaleState(nextLocale);
    } catch (reason) {
      if (version === localeVersion.current) queueBrowserFlash(reason instanceof Error ? reason.message : String(reason), "error");
    }
  }, []);

  const toggleTheme = useCallback(() => {
    setTheme((current) => {
      const next = current === "gf-dark" ? "gf-light" : "gf-dark";
      themeRef.current = next;
      applyBrowserTheme(next, page?.layout.theme.colors);
      return next;
    });
  }, [page?.layout.theme.colors]);

  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);

  useEffect(() => {
    const previousScrollRestoration = window.history.scrollRestoration;
    window.history.scrollRestoration = "manual";
    activeHistoryEntry.current = ensureHistoryEntry();
    activeHistoryIndex.current = readHistoryIndex(window.history.state);
    if (initialPage) prepareDocument(initialPage, themeRef.current);
    else void loadPage(new URL(window.location.href), "none");

    async function handlePopState(event: PopStateEvent) {
      if (suppressNextPop.current) {
        suppressNextPop.current = false;
        return;
      }
      const targetIndex = readHistoryIndex(event.state);
      const previousIndex = activeHistoryIndex.current;
      const url = new URL(window.location.href);
      if (
        navigationBlocker.current &&
        !(await navigationBlocker.current(url.href))
      ) {
        const delta = previousIndex - targetIndex;
        if (delta) {
          suppressNextPop.current = true;
          window.history.go(delta);
        }
        return;
      }
      activeRequest.current?.abort();
      saveScrollPosition(scrollPositions.current, activeHistoryEntry.current);
      const entryKey = readHistoryEntry(event.state) || ensureHistoryEntry();
      activeHistoryEntry.current = entryKey;
      activeHistoryIndex.current = targetIndex;
      const cached = cachedPageRef.current.get(routeCacheKey(url));
      if (cached) {
        setIsNavigating(false);
        setError(undefined);
        prepareDocument(cached, themeRef.current);
        pendingScroll.current = {
          page: cached,
          action: "restore",
          entryKey,
          hash: url.hash,
        };
        startTransition(() => {
          rememberPage(cached);
          setPage(cached);
        });
        return;
      }
      void loadPage(url, "none", "restore", entryKey, targetIndex);
    }

    const handlePopStateEvent = (event: PopStateEvent) => {
      void handlePopState(event);
    };
    window.addEventListener("popstate", handlePopStateEvent);
    return () => {
      window.removeEventListener("popstate", handlePopStateEvent);
      window.history.scrollRestoration = previousScrollRestoration;
      activeRequest.current?.abort();
    };
  }, [loadPage, initialPage, rememberPage]);

  const fetchPage = useCallback(
    (href: string, signal?: AbortSignal) =>
      pageSource.load(new URL(href, window.location.href), signal),
    [pageSource],
  );

  const runtime = useMemo<GooseRuntime>(
    () => ({
      api: pageSource.api,
      currentUrl: page?.url ?? browserHref(),
      isNavigating,
      theme,
      locale,
      navigate,
      registerNavigationBlocker,
      fetchPage,
      queueFlash: queueBrowserFlash,
      redirect: (href) => window.location.assign(href),
      refresh,
      setLocale,
      toggleTheme,
    }),
    [
      fetchPage,
      isNavigating,
      locale,
      navigate,
      registerNavigationBlocker,
      page?.url,
      pageSource,
      refresh,
      setLocale,
      theme,
      toggleTheme,
    ],
  );

  const activeCacheKey =
    page && keepAliveComponents.has(page.component)
      ? routeCacheKey(new URL(page.url, browserHref()))
      : "";

  const pageContent = page ? (
    <>
      {cachedPages.map((entry) => (
        <Activity
          key={entry.key}
          mode={!error && entry.key === activeCacheKey ? "visible" : "hidden"}
        >
          <GoosePage page={entry.page} onContentReady={handleContentReady} />
        </Activity>
      ))}
      {error ? <BootstrapError error={error} description={errorDescription} onRetry={() => void refresh()} retrying={isNavigating} /> : !activeCacheKey ? <GoosePage page={page} onContentReady={handleContentReady} /> : null}
    </>
  ) : null;

  return (
    <GooseI18nProvider locale={locale} initialResources={initialResources}>
      <GooseRuntimeProvider runtime={runtime}>
        {!page && error ? (
          <BootstrapError
            error={error}
            description={errorDescription}
            onRetry={() => void refresh()}
            retrying={isNavigating}
          />
        ) : page ? (
          <AppShell layout={page.layout} standalone={isStandalonePage(page)}>
            {pageContent}
          </AppShell>
        ) : (
          <BootstrapLoading />
        )}
      </GooseRuntimeProvider>
    </GooseI18nProvider>
  );
}

function browserHref() {
  return typeof window === "undefined"
    ? "http://gooseforum.local/"
    : window.location.href;
}

function browserUrlForPage(pageUrl: string, requestedUrl: URL) {
  const payloadUrl = new URL(pageUrl, requestedUrl);
  const browserUrl = new URL(
    `${payloadUrl.pathname}${payloadUrl.search}`,
    window.location.origin,
  );
  browserUrl.hash = payloadUrl.hash || requestedUrl.hash;
  return browserUrl;
}

function routeCacheKey(url: URL) {
  return `${url.pathname}${url.search}`;
}

function createHistoryEntryKey() {
  return (
    globalThis.crypto?.randomUUID?.() ||
    `${Date.now()}-${Math.random().toString(36).slice(2)}`
  );
}

function readHistoryEntry(state: unknown) {
  if (!state || typeof state !== "object") return "";
  const value = (state as Record<string, unknown>)[historyEntryKey];
  return typeof value === "string" ? value : "";
}

function readHistoryIndex(state: unknown) {
  if (!state || typeof state !== "object") return 0;
  const value = (state as Record<string, unknown>)[historyEntryIndex];
  return typeof value === "number" && Number.isFinite(value) ? value : 0;
}

function withHistoryEntry(
  state: unknown,
  entryKey: string,
  entryIndex = readHistoryIndex(state),
) {
  return {
    ...(state && typeof state === "object" ? state : {}),
    [historyEntryKey]: entryKey,
    [historyEntryIndex]: entryIndex,
  };
}

function ensureHistoryEntry() {
  const existing = readHistoryEntry(window.history.state);
  if (existing) return existing;
  const entryKey = createHistoryEntryKey();
  window.history.replaceState(
    withHistoryEntry(window.history.state, entryKey),
    "",
    window.location.href,
  );
  return entryKey;
}

function saveScrollPosition(
  positions: Map<string, { left: number; top: number }>,
  entryKey: string,
) {
  if (!entryKey) return;
  positions.set(entryKey, { left: window.scrollX, top: window.scrollY });
}

function applyScroll(
  action: Exclude<ScrollAction, "preserve">,
  entryKey: string,
  hash: string,
  positions: Map<string, { left: number; top: number }>,
) {
  if (hash) {
    document
      .getElementById(decodeURIComponent(hash.slice(1)))
      ?.scrollIntoView({ behavior: "instant" });
    return;
  }
  const saved = action === "restore" ? positions.get(entryKey) : undefined;
  window.scrollTo({
    left: saved?.left || 0,
    top: saved?.top || 0,
    behavior: "instant",
  });
}
