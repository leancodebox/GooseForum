import type { Resource, ResourceKey } from "i18next";
import type { Locale } from "@gooseforum/client/i18n/locale";
import { resourceLoaders } from "./resource-loaders";

const namespaceFiles: Record<string, string[]> = {
  auth: ["auth"],
  oidcConsent: ["oidc-consent"],
  ...Object.fromEntries(
    [
      "shell",
      "links",
      "sponsors",
      "categories",
      "members",
      "home",
      "search",
      "user",
      "userCard",
    ].map((name) => [name, [`site-${name}`]]),
  ),
  settings: ["settings"],
  notifications: ["notifications"],
  messages: ["messages"],
  drafts: ["legacy-site-pages-drafts"],
  accessGroups: ["legacy-site-pages-accessGroups"],
  error: ["content-error"],
  contentCommon: ["content-common"],
  serverMessages: ["server-messages"],
  moderation: ["moderation", "legacy-site-pages-moderation"],
  publish: ["publish", "legacy-site-pages-publish"],
  topic: ["topic", "legacy-site-pages-topic"],
  themePreview: ["theme-preview", "legacy-site-pages-themePreview"],
};
export const gooseNamespaces = Object.keys(namespaceFiles);
export const commonNamespaces = [
  "auth",
  "shell",
  "home",
  "userCard",
  "contentCommon",
  "serverMessages",
];
const resources: Resource = {};
const pending = new Map<string, Promise<void>>();
type Listener = (
  locale: string,
  namespace: string,
  dictionary: ResourceKey,
) => void;
const listeners = new Set<Listener>();

export function cachedGooseResources(): Resource {
  return structuredClone(resources);
}
export function loadedGooseNamespaces() {
  return [
    ...new Set(Object.values(resources).flatMap((value) => Object.keys(value))),
  ];
}
export function subscribeGooseResources(listener: Listener) {
  listeners.add(listener);
  for (const [locale, namespaces] of Object.entries(resources))
    for (const [namespace, dictionary] of Object.entries(namespaces))
      listener(locale, namespace, dictionary);
  return () => {
    listeners.delete(listener);
  };
}

export async function prepareGooseTranslations(
  locale: Locale,
  namespaces: readonly string[] = commonNamespaces,
) {
  const languages = locale === "zh" ? ["zh"] : [locale, "zh"];
  await Promise.all(
    languages.flatMap((language) =>
      [...new Set([...commonNamespaces, ...namespaces])].map((namespace) =>
        loadNamespace(language, namespace),
      ),
    ),
  );
}
function loadNamespace(locale: string, namespace: string): Promise<void> {
  if (resources[locale]?.[namespace]) return Promise.resolve();
  const key = `${locale}/${namespace}`;
  const existing = pending.get(key);
  if (existing) return existing;
  const files = namespaceFiles[namespace];
  if (!files)
    return Promise.reject(
      new Error(`Unknown translation namespace: ${namespace}`),
    );
  const promise = Promise.all(
    files.map((file) => resourceLoaders[`${locale}/${file}`]()),
  )
    .then((modules) => {
      const [base, legacy] = modules.map((module) => module.default);
      let dictionary = base;
      if (legacy) {
        dictionary = { ...base, ...legacy };
        if (namespace === "topic")
          dictionary.reportReasons = {
            ...base.reportReasons,
            ...legacy.reportReasons,
          };
        if (namespace === "moderation") {
          dictionary.tabs = { ...base.tabs, ...legacy.managementTabs };
          dictionary.reports = { ...base.reports, ...legacy.reports };
          dictionary.logs = { ...base.logs, ...legacy.logs };
        }
        if (namespace === "themePreview")
          dictionary.presets = {
            ...base.presets,
            ...Object.fromEntries(
              Object.entries(legacy.presets).map(([key, value]) => [
                key,
                (value as { label: string }).label,
              ]),
            ),
          };
      }
      resources[locale] ??= {};
      resources[locale][namespace] = dictionary;
      for (const listener of listeners) listener(locale, namespace, dictionary);
    })
    .finally(() => {
      pending.delete(key);
    });
  pending.set(key, promise);
  return promise;
}

export function goosePageNamespaces(component: string) {
  const pages: Record<string, string[]> = {
    "auth.oidcConsent": ["oidcConsent"],
    "settings.index": ["settings", "user", "publish"],
    "notifications.index": ["notifications"],
    "messages.index": ["messages"],
    "drafts.index": ["drafts"],
    "access-groups.index": ["accessGroups"],
    "error.index": ["error"],
    "moderation.index": ["moderation"],
    "publish.index": ["publish"],
    "topic.detail": ["topic", "publish"],
    "theme.preview": ["themePreview", "topic", "publish"],
    "links.index": ["links"],
    "sponsors.index": ["sponsors"],
    "categories.index": ["categories"],
    "members.index": ["members"],
    "search.index": ["search"],
    "user.profile": ["user"],
  };
  return pages[component] || [];
}
