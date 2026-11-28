import type { AccessGroup } from "@gooseforum/client";
import { adminText, type Dictionary } from "@gooseforum/i18n/admin";
type Messages = Dictionary<"admin.access-groups">;
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

export type AccessGroupTextKey = keyof Messages;

export function createAccessGroupText(locale: AuthLocale) {
  return (key: AccessGroupTextKey) =>
    adminText(locale, "admin.access-groups")(key);
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
