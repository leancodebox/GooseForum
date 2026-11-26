import type { AccessGroup } from "@gooseforum/client";
import { getAdminDictionary } from "./translation-loader";
type EnglishDictionary = typeof import("./messages/en-access-groups").default;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type AccessGroupTextKey = keyof EnglishDictionary;

export function createAccessGroupText(locale: AuthLocale) {
  return (key: AccessGroupTextKey) =>
    getAdminDictionary<EnglishDictionary>("access-groups", locale)[key];
}

export function accessGroupName(group: Pick<AccessGroup, "name" | "systemKey">, text: (key: AccessGroupTextKey) => string) {
  switch (group.systemKey) {
    case "everyone": return text("systemEveryone");
    case "registered": return text("systemRegistered");
    case "trust:100": return text("memberAndAbove");
    case "trust:200": return text("regularAndAbove");
    case "trust:1000": return text("communityLeaderAndAbove");
    default: return group.name;
  }
}
