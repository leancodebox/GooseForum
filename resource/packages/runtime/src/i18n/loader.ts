import type { Locale } from "@gooseforum/client/i18n/locale";
import {
  siteNamespaces as gooseNamespaces,
  siteTranslations,
  type SiteNamespace,
} from "@gooseforum/i18n/site";

export { gooseNamespaces };
export const commonNamespaces = [
  "site.auth", "shared.shell", "site.home", "site.userCard",
  "site.contentCommon", "shared.serverMessages",
] as const;

export const cachedGooseResources = siteTranslations.snapshot;
export const loadedGooseNamespaces = siteTranslations.loadedNamespaces;
export const subscribeGooseResources = siteTranslations.subscribe;

export async function prepareGooseTranslations(
  locale: Locale,
  namespaces: readonly SiteNamespace[] = commonNamespaces,
) {
  return siteTranslations.prepare(locale, [...commonNamespaces, ...namespaces]);
}

export function goosePageNamespaces(component: string) {
  const pages: Record<string, readonly SiteNamespace[]> = {
    "auth.oidcConsent": ["site.oidcConsent"],
    "settings.index": ["site.settings", "site.user", "site.publish"],
    "notifications.index": ["site.notifications"],
    "messages.index": ["site.messages"],
    "drafts.index": ["site.drafts", "site.publish"],
    "access-groups.index": ["site.accessGroups"],
    "error.index": ["site.error"],
    "moderation.index": ["site.moderation"],
    "publish.index": ["site.publish", "site.topic"],
    "topic.detail": ["site.topic", "site.publish"],
    "theme.preview": ["site.themePreview", "site.topic", "site.publish"],
    "links.index": ["site.links"],
    "sponsors.index": ["site.sponsors"],
    "categories.index": ["site.categories"],
    "members.index": ["site.members"],
    "search.index": ["site.search"],
    "user.profile": ["site.user"],
  };
  return pages[component] || [];
}
