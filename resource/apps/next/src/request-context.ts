import { normalizeLocale, type Locale } from "@gooseforum/client/i18n/locale";

export function resolveLocale(
  path: string,
  cookieHeader: string,
  acceptLanguage: string,
): Locale {
  const queryLocale = normalizeLocale(new URL(path, "http://next.local").searchParams.get("lang"));
  if (queryLocale) return queryLocale;
  const cookieLocale = normalizeLocale(readCookie(cookieHeader, "lang"));
  if (cookieLocale) return cookieLocale;
  return normalizeLocale(acceptLanguage) || "zh";
}

export function readCookie(header: string, name: string) {
  const prefix = `${name}=`;
  const value = header
    .split(";")
    .map((item) => item.trim())
    .find((item) => item.startsWith(prefix))
    ?.slice(prefix.length);
  return value ? decodeURIComponent(value) : undefined;
}

export function pagePath(
  segments: string[] | undefined,
  search: Record<string, string | string[] | undefined>,
) {
  const pathname = segments?.length
    ? `/${segments.map(encodeURIComponent).join("/")}`
    : "/";
  const query = new URLSearchParams();
  for (const key of Object.keys(search).sort()) {
    const value = search[key];
    for (const item of Array.isArray(value) ? value : [value]) {
      if (item !== undefined) query.append(key, item);
    }
  }
  const encoded = query.toString();
  return encoded ? `${pathname}?${encoded}` : pathname;
}
