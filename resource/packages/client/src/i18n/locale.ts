export const supportedLocales = ["zh", "en", "ja", "it"] as const;
export type Locale = (typeof supportedLocales)[number];
export const localeLabels = {
  zh: { label: '简体中文', short: '中' },
  en: { label: 'English', short: 'EN' },
  ja: { label: '日本語', short: '日' },
  it: { label: 'Italiano', short: 'IT' },
} as const;

export function normalizeLocale(value?: string | null): Locale | undefined {
  const short = (value || "")
    .trim()
    .toLowerCase()
    .split(/[-_,;]/)[0] as Locale;
  return supportedLocales.includes(short) ? short : undefined;
}
