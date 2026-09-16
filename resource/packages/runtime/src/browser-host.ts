import {
  updateDocumentMetadata,
  type PagePayload,
  type ThemePayload,
} from "@gooseforum/client";
import {
  normalizeLocale,
  supportedLocales,
  type Locale,
} from "@gooseforum/client/i18n/locale";
import type { GooseFlashType } from "./runtime";

const flashStorageKey = "goose:flash-messages";
const themeStorageKey = "goose-site-theme";

export function detectBrowserLocale(): Locale {
  const queryLocale = normalizeLocale(
    new URL(window.location.href).searchParams.get("lang"),
  );
  if (queryLocale) return queryLocale;
  const cookieLocale = normalizeLocale(readCookie("lang"));
  if (cookieLocale) return cookieLocale;
  for (const language of navigator.languages || [navigator.language]) {
    const locale = normalizeLocale(language);
    if (locale) return locale;
  }
  return supportedLocales[0];
}

export function applyBrowserLocale(locale: Locale) {
  document.documentElement.lang = locale;
  document.cookie = `lang=${locale}; path=/; max-age=31536000; samesite=lax`;
}

export function queueBrowserFlash(
  message: string,
  type: GooseFlashType = "info",
) {
  const text = message.trim();
  if (!text) return;
  try {
    const stored = window.sessionStorage.getItem(flashStorageKey);
    const parsed: unknown = stored ? JSON.parse(stored) : [];
    const messages = Array.isArray(parsed) ? parsed : [];
    messages.push({ type, message: text });
    window.sessionStorage.setItem(
      flashStorageKey,
      JSON.stringify(messages.slice(-4)),
    );
  } catch {
    // Storage can be unavailable in privacy modes.
  }
}

export function detectBrowserTheme(): ThemePayload["current"] {
  const cookieTheme = readCookie(themeStorageKey);
  if (cookieTheme === "gf-light" || cookieTheme === "gf-dark")
    return cookieTheme;
  try {
    const stored = localStorage.getItem(themeStorageKey);
    if (stored === "gf-light" || stored === "gf-dark") return stored;
  } catch {
    // Fall through to the document theme.
  }
  const documentTheme = document.documentElement.dataset.theme;
  if (documentTheme === "gf-light" || documentTheme === "gf-dark")
    return documentTheme;
  return "gf-light";
}

export function applyBrowserTheme(
  theme: ThemePayload["current"],
  colors?: Record<string, string>,
) {
  document.documentElement.dataset.theme = theme;
  document.documentElement.style.colorScheme =
    theme === "gf-dark" ? "dark" : "light";
  document
    .querySelector('meta[name="theme-color"]')
    ?.setAttribute(
      "content",
      colors?.[theme] || (theme === "gf-dark" ? "#101010" : "#fbfdff"),
    );
  document.cookie = `${themeStorageKey}=${theme}; path=/; max-age=31536000; samesite=lax`;
  try {
    localStorage.setItem(themeStorageKey, theme);
  } catch {
    // Ignore storage failures in restricted browsing modes.
  }
}

export function prepareBrowserDocument(
  payload: PagePayload,
  preferredTheme?: ThemePayload["current"],
) {
  updateDocumentMetadata(payload);
  document.documentElement.lang ||= "zh-CN";
  applyBrowserTheme(
    preferredTheme ?? payload.layout.theme.current,
    payload.layout.theme.colors,
  );
  applyThemeStylesheet(payload.layout.theme);
}

function applyThemeStylesheet(theme: ThemePayload) {
  let link = document.querySelector<HTMLLinkElement>("#goose-site-theme-link");
  if (!theme.enabled || !theme.href) {
    link?.remove();
    return;
  }
  if (!link) {
    link = document.createElement("link");
    link.id = "goose-site-theme-link";
    link.rel = "stylesheet";
    document.head.appendChild(link);
  }
  link.href = theme.href;
}

function readCookie(name: string) {
  const prefix = `${name}=`;
  const value = document.cookie
    .split("; ")
    .find((item) => item.startsWith(prefix))
    ?.slice(prefix.length);
  return value ? decodeURIComponent(value) : undefined;
}
