import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";
import { adminTranslations, type AdminNamespace } from "@gooseforum/i18n/admin";
import { normalizeAdminPath } from "./navigation";

export const loadedAdminNamespaces = adminTranslations.loadedNamespaces;
export async function prepareAdminTranslations(locale: AuthLocale, namespaces: readonly AdminNamespace[] = []) {
  await adminTranslations.prepare(locale, ["admin.shell", "shared.shell", ...namespaces]);
}

export function adminPageNamespaces(path: string) {
  const pages: Record<string, readonly AdminNamespace[]> = {
    "/admin": ["admin.dashboard"],
    "/admin/categories": ["admin.shell", "admin.access-groups", "admin.settings"],
    "/admin/access-groups": ["admin.access-groups"],
    "/admin/roles": ["admin.roles"],
    "/admin/users": ["admin.users"],
    "/admin/registration-invitations": ["admin.assets"],
    "/admin/posts": ["admin.posts"],
    "/admin/links": ["admin.page-config"],
    "/admin/sponsors": ["admin.page-config"],
    "/admin/badges": ["admin.assets"],
    "/admin/files/resources": ["admin.assets"],
    "/admin/opt-records": ["admin.audit", "admin.audit-extra"],
    "/admin/auth-logs": ["admin.audit"],
    "/admin/settings/site-info": ["admin.settings"],
    "/admin/settings/site-chrome": ["admin.settings"],
    "/admin/settings/mail": ["admin.system-settings"],
    "/admin/settings/security": ["admin.system-settings"],
    "/admin/settings/posting": ["admin.content-settings"],
    "/admin/settings/agent": ["admin.system-settings"],
    "/admin/settings/trust": ["admin.trust-settings"],
    "/admin/settings/announcement": ["admin.content-settings"],
    "/admin/settings/http-notify": ["admin.moderation-settings"],
    "/admin/settings/sensitive-words": ["admin.moderation-settings"],
    "/admin/settings/oauth": ["admin.identity-settings"],
    "/admin/settings/oidc-provider": ["admin.identity-settings"],
  };
  return pages[normalizeAdminPath(path)] || [];
}
