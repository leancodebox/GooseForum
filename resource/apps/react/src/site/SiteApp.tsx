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
import {
  BootstrapError,
  BootstrapLoading,
  AppShell,
  GoosePage,
  isStandalonePage,
} from "@gooseforum/react/app";
import {
  GooseRuntimeProvider,
  type GooseNavigateOptions,
  type GooseRuntime,
} from "@gooseforum/react/runtime";
import {
  applyBrowserLocale,
  applyBrowserTheme,
  detectBrowserTheme,
  detectBrowserLocale,
  prepareDocument,
  queueBrowserFlash,
  type PageSource,
} from "../browser-runtime";
import type { AuthLocale } from "@gooseforum/react/i18n/auth";
import { GooseI18nProvider } from "@gooseforum/react/i18n";

export interface SiteAppProps {
  pageSource: PageSource<AnyPagePayload>;
}

interface CachedPage {
  key: string;
  page: AnyPagePayload;
}

type ScrollAction = "preserve" | "reset" | "restore";

const keepAliveComponents = new Set([
  "home.index",
  "category.index",
  "search.index",
]);
const maxCachedPages = 10;
const historyEntryKey = "gooseEntryKey";
const historyEntryIndex = "gooseEntryIndex";

export function SiteApp({ pageSource }: SiteAppProps) {
  const [page, setPage] = useState<AnyPagePayload>();
  const [cachedPages, setCachedPages] = useState<CachedPage[]>([]);
  const [error, setError] = useState<unknown>();
  const [isNavigating, setIsNavigating] = useState(false);
  const [locale, setLocaleState] = useState<AuthLocale>(detectBrowserLocale);
  const [theme, setTheme] = useState(detectBrowserTheme);
  const themeRef = useRef(theme);
  const activeRequest = useRef<AbortController | null>(null);
  const navigationBlocker = useRef<
    ((href: string) => boolean | Promise<boolean>) | null
  >(null);
  const cachedPageRef = useRef(new Map<string, AnyPagePayload>());
  const scrollPositions = useRef(
    new Map<string, { left: number; top: number }>(),
  );
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
        startTransition(() => {
          rememberPage(nextPage);
          setPage(nextPage);
        });
        scheduleScroll(
          scrollAction,
          nextHistoryEntry,
          url.hash,
          scrollPositions.current,
        );
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
      if (url.origin !== window.location.origin) {
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

  const setLocale = useCallback((nextLocale: AuthLocale) => {
    applyBrowserLocale(nextLocale);
    setLocaleState(nextLocale);
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
    void loadPage(new URL(window.location.href), "none");

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
        startTransition(() => {
          rememberPage(cached);
          setPage(cached);
        });
        scheduleScroll("restore", entryKey, url.hash, scrollPositions.current);
        return;
      }
      void loadPage(url, "none", "restore", entryKey, targetIndex);
    }

    window.addEventListener("popstate", handlePopState);
    return () => {
      window.removeEventListener("popstate", handlePopState);
      window.history.scrollRestoration = previousScrollRestoration;
      activeRequest.current?.abort();
    };
  }, [loadPage]);

  const runtime = useMemo<GooseRuntime>(
    () => ({
      api: pageSource.api,
      currentUrl: page?.url ?? window.location.href,
      isNavigating,
      theme,
      locale,
      navigate,
      registerNavigationBlocker,
      fetchPage: (href, signal) =>
        pageSource.load(new URL(href, window.location.href), signal),
      queueFlash: queueBrowserFlash,
      redirect: (href) => window.location.assign(href),
      refresh,
      setLocale,
      toggleTheme,
    }),
    [
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
      ? routeCacheKey(new URL(page.url, window.location.href))
      : "";

  const pageContent = page ? (
    <>
      {cachedPages.map((entry) => (
        <Activity
          key={entry.key}
          mode={entry.key === activeCacheKey ? "visible" : "hidden"}
        >
          <GoosePage page={entry.page} />
        </Activity>
      ))}
      {!activeCacheKey ? <GoosePage page={page} /> : null}
    </>
  ) : null;

  return (
    <GooseI18nProvider locale={locale}>
      <GooseRuntimeProvider runtime={runtime}>
        {error ? (
          <BootstrapError
            error={error}
            onRetry={() => void refresh()}
            retrying={isNavigating}
          />
        ) : page ? (
          isStandalonePage(page) ? pageContent : (
            <AppShell layout={page.layout}>{pageContent}</AppShell>
          )
        ) : (
          <BootstrapLoading />
        )}
      </GooseRuntimeProvider>
    </GooseI18nProvider>
  );
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

function scheduleScroll(
  action: ScrollAction,
  entryKey: string,
  hash: string,
  positions: Map<string, { left: number; top: number }>,
) {
  if (action === "preserve") return;
  requestAnimationFrame(() =>
    requestAnimationFrame(() => {
      if (hash) {
        document
          .getElementById(decodeURIComponent(hash.slice(1)))
          ?.scrollIntoView();
        return;
      }
      const saved = action === "restore" ? positions.get(entryKey) : undefined;
      window.scrollTo({
        left: saved?.left || 0,
        top: saved?.top || 0,
        behavior: "auto",
      });
    }),
  );
}
