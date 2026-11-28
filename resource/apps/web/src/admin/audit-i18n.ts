import { runtimeText } from "@gooseforum/i18n";
import { adminText, getAdminDictionary, type Dictionary } from "@gooseforum/i18n/admin";
type Messages = Dictionary<"admin.audit">;
type AuditExtra = Dictionary<"admin.audit-extra">;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

type MessageKey = keyof AuditExtra["messages"];

export const adminAuditMessageCodes = {
  "admin.opt.badge.granted": "badgeGranted",
  "admin.opt.badge.revoked": "badgeRevoked",
  "admin.opt.content.reviewed": "contentReviewed",
  "admin.opt.user.updated": "userUpdated",
  "admin.opt.topic.statusChanged": "topicStatusChanged",
  "admin.opt.topic.pinWeightChanged": "topicPinWeightChanged",
  "admin.opt.topic.categoriesChanged": "topicCategoriesChanged",
  "admin.opt.topic.deleted": "topicDeleted",
  "moderator.opt.topic.statusChanged": "moderatorTopicStatusChanged",
  "admin.opt.category.moderatorAdded": "categoryModeratorAdded",
  "admin.opt.category.moderatorRemoved": "categoryModeratorRemoved",
} as const satisfies Record<string, MessageKey>;
export type AuditTextKey = Exclude<keyof Messages, 'authLogs' | 'authLogActions'>;

export function createAuditText(locale: AuthLocale) {
  return (key: AuditTextKey) =>
    adminText(locale, "admin.audit")(key);
}

export function formatAuditMessage(
  locale: AuthLocale,
  messageCode: string,
  params: Record<string, unknown>,
  fallback: string,
) {
  const key = adminAuditMessageCodes[
    messageCode as keyof typeof adminAuditMessageCodes
  ];
  if (!key) return fallback || messageCode;
  const values = normalizeParams(locale, params);
  // API audit messages use heterogeneous parameter sets. Normalize the payload
  // before the runtime formats the checked catalog key.
  return runtimeText(adminText(locale, "admin.audit-extra"), "messages." + key, {
    replace: values,
    defaultValue: fallback || messageCode,
  });
}

function normalizeParams(locale: AuthLocale, params: Record<string, unknown>) {
  const values: Record<string, string> = {};
  for (const [key, value] of Object.entries(params))
    values[key] = Array.isArray(value)
      ? value.map((item: unknown) => stringifyParam(item)).join(", ")
      : value == null
        ? ""
        : stringifyParam(value);
  if (
    typeof params.status === "string" &&
    params.status in
      getAdminDictionary("admin.audit-extra", locale).statusLabels
  )
    values.status = getAdminDictionary("admin.audit-extra", locale).statusLabels[params.status as keyof AuditExtra["statusLabels"]];
  if (typeof params.type === "string")
    values.type = localizedParam(
      locale,
      "contentTypes",
      params.type,
    );
  if (typeof params.action === "string")
    values.action = localizedParam(
      locale,
      "actions",
      params.action,
    );
  if (Array.isArray(params.changes))
    values.changedFields = params.changes
      .map(
        (field) =>
          getAdminDictionary("admin.audit-extra", locale).fieldLabels[
            String(field) as keyof AuditExtra["fieldLabels"]
          ] || String(field),
      )
      .join(", ");
  return values;
}

function stringifyParam(value: unknown): string {
  if (value == null) return "";
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean" || typeof value === "bigint") return String(value);
  try {
    return JSON.stringify(value) ?? "";
  } catch {
    return "";
  }
}

function localizedParam(
  locale: AuthLocale,
  group: "contentTypes" | "actions",
  value: string,
) {
  const labels = getAdminDictionary("admin.audit-extra", locale)[group];
  return value in labels ? labels[value as keyof typeof labels] : value;
}
