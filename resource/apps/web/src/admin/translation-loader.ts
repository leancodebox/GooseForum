import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";
import { translationLoaders } from "./translation-resources";
import { normalizeAdminPath } from "./navigation";

const dictionaries = new Map<string, unknown>();
const pending = new Map<string, Promise<void>>();
export function loadedAdminNamespaces() {
  return [...new Set([...dictionaries.keys()].map((key) => key.split("/")[1]))];
}
export async function prepareAdminTranslations(
  locale: AuthLocale,
  namespaces: readonly string[] = ["shell"],
) {
  await Promise.all(
    [...new Set(["shell", ...namespaces])].map(async (namespace) => {
      const key = `${locale}/${namespace}`;
      if (dictionaries.has(key)) return;
      if (pending.has(key)) return pending.get(key);
      const loader = translationLoaders[key];
      if (!loader)
        throw new Error(`Unknown admin translation namespace: ${namespace}`);
      const promise = loader()
        .then((module) => {
          dictionaries.set(key, module.default);
        })
        .finally(() => {
          pending.delete(key);
        });
      pending.set(key, promise);
      await promise;
    }),
  );
}
export function getAdminDictionary<Dictionary>(
  namespace: string,
  locale: AuthLocale,
): Dictionary {
  const dictionary = dictionaries.get(`${locale}/${namespace}`);
  if (!dictionary)
    throw new Error(
      `Admin translations were not prepared: ${locale}/${namespace}`,
    );
  return dictionary as Dictionary;
}
export function adminPageNamespaces(path: string) {
  const pages: Record<string, string[]> = {
    "/admin": ["dashboard"],
    "/admin/categories": ["shell", "access-groups"],
    "/admin/access-groups": ["access-groups"],
    "/admin/roles": ["roles"],
    "/admin/users": ["users"],
    "/admin/posts": ["posts"],
    "/admin/links": ["page-config"],
    "/admin/sponsors": ["page-config"],
    "/admin/badges": ["assets"],
    "/admin/files/resources": ["assets"],
    "/admin/opt-records": ["audit", "audit-extra"],
    "/admin/settings/site-info": ["settings"],
    "/admin/settings/site-chrome": ["settings"],
    "/admin/settings/mail": ["system-settings"],
    "/admin/settings/security": ["system-settings"],
    "/admin/settings/posting": ["content-settings"],
    "/admin/settings/announcement": ["content-settings"],
    "/admin/settings/http-notify": ["moderation-settings"],
    "/admin/settings/sensitive-words": ["moderation-settings"],
    "/admin/settings/oauth": ["identity-settings"],
    "/admin/settings/oidc-provider": ["identity-settings"],
  };
  return pages[normalizeAdminPath(path)] || [];
}
