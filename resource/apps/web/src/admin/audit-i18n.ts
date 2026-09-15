import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary = typeof import("./messages/en-audit").default;
type AuditExtra = typeof import("./messages/en-audit-extra").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

type MessageKey = keyof AuditExtra["messages"];

export const adminAuditMessageCodes = {
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
export type AuditTextKey = keyof EnglishDictionary;

export function createAuditText(locale: AuthLocale) {
  return (key: AuditTextKey) =>
    getAdminDictionary<EnglishDictionary>("audit", locale)[key];
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
  return getAdminDictionary<AuditExtra>("audit-extra", locale).messages[
    key
  ].replace(/\{(\w+)\}/g, (_match, name: string) => values[name] ?? "");
}

function normalizeParams(locale: AuthLocale, params: Record<string, unknown>) {
  const values: Record<string, string> = {};
  for (const [key, value] of Object.entries(params))
    values[key] = Array.isArray(value)
      ? value.join(", ")
      : value == null
        ? ""
        : String(value);
  if (
    typeof params.status === "string" &&
    params.status in
      getAdminDictionary<AuditExtra>("audit-extra", locale).statusLabels
  )
    values.status = getAdminDictionary<AuditExtra>(
      "audit-extra",
      locale,
    ).statusLabels[params.status as keyof AuditExtra["statusLabels"]];
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
          getAdminDictionary<AuditExtra>("audit-extra", locale).fieldLabels[
            String(field) as keyof AuditExtra["fieldLabels"]
          ] || String(field),
      )
      .join(", ");
  return values;
}

function localizedParam(
  locale: AuthLocale,
  group: "contentTypes" | "actions",
  value: string,
) {
  const labels = getAdminDictionary<AuditExtra>("audit-extra", locale)[group];
  return value in labels ? labels[value as keyof typeof labels] : value;
}
