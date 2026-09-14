import {
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
  GooseApp,
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

export function SiteApp({ pageSource }: SiteAppProps) {
  const [page, setPage] = useState<AnyPagePayload>();
  const [error, setError] = useState<unknown>();
  const [isNavigating, setIsNavigating] = useState(false);
  const [locale, setLocaleState] = useState<AuthLocale>(detectBrowserLocale);
  const [theme, setTheme] = useState(detectBrowserTheme);
  const themeRef = useRef(theme);
  const activeRequest = useRef<AbortController | null>(null);

  const loadPage = useCallback(
    async (url: URL, historyAction: "push" | "replace" | "none") => {
      activeRequest.current?.abort();
      const request = new AbortController();
      activeRequest.current = request;
      setIsNavigating(true);

      try {
        const nextPage = await pageSource.load(url, request.signal);
        if (request.signal.aborted) return;

        const nextUrl = browserUrlForPage(nextPage.url, url);
        if (historyAction === "push") {
          window.history.pushState(null, "", nextUrl);
        } else if (
          historyAction === "replace" ||
          nextUrl.href !== window.location.href
        ) {
          window.history.replaceState(null, "", nextUrl);
        }

        prepareDocument(nextPage, themeRef.current);
        setError(undefined);
        startTransition(() => setPage(nextPage));

        if (historyAction !== "none") {
          requestAnimationFrame(() => {
            if (url.hash) {
              document
                .getElementById(decodeURIComponent(url.hash.slice(1)))
                ?.scrollIntoView();
            } else {
              window.scrollTo({ top: 0, left: 0, behavior: "auto" });
            }
          });
        }
      } catch (nextError) {
        if (!request.signal.aborted) setError(nextError);
      } finally {
        if (activeRequest.current === request) {
          activeRequest.current = null;
          setIsNavigating(false);
        }
      }
    },
    [pageSource],
  );

  const navigate = useCallback(
    async (href: string, options: GooseNavigateOptions = {}) => {
      const url = new URL(href, window.location.href);
      if (url.origin !== window.location.origin) {
        window.location.assign(url);
        return;
      }
      await loadPage(url, options.replace ? "replace" : "push");
    },
    [loadPage],
  );

  const refresh = useCallback(async () => {
    await loadPage(new URL(window.location.href), "none");
  }, [loadPage]);

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
    void loadPage(new URL(window.location.href), "none");

    function handlePopState() {
      void loadPage(new URL(window.location.href), "none");
    }

    window.addEventListener("popstate", handlePopState);
    return () => {
      window.removeEventListener("popstate", handlePopState);
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
      page?.url,
      pageSource,
      refresh,
      setLocale,
      theme,
      toggleTheme,
    ],
  );

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
          <GooseApp page={page} />
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
