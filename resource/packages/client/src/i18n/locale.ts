export const supportedLocales = ["zh", "en", "ja", "it", "ru", "ko", "fr"] as const;
export type Locale = (typeof supportedLocales)[number];
export const localeLabels = {
  zh: { label: '简体中文', short: '中' },
  en: { label: 'English', short: 'EN' },
  ja: { label: '日本語', short: '日' },
  ko: { label: '한국어', short: '한' },
  fr: { label: 'Français', short: 'FR' },
  it: { label: 'Italiano', short: 'IT' },
  // Original Russian translations contributed by @oofsdfs in GooseForum PR #19.
  ru: { label: "Русский", short: "RU" },
} as const satisfies Record<Locale, { label: string; short: string }>;

export const localeMenuOrder: readonly Locale[] = Object.keys(localeLabels) as Locale[];

export function normalizeLocale(value?: string | null): Locale | undefined {
  const short = (value || "")
    .trim()
    .toLowerCase()
    .split(/[-_,;]/)[0] as Locale;
  return supportedLocales.includes(short) ? short : undefined;
}
