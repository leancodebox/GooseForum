import { describe, expect, it } from "vitest";
import { supportedLocales } from "@gooseforum/client/i18n/locale";
import { createGooseI18n } from "@gooseforum/runtime/i18n";

describe("migrated content page translations", () => {
  for (const locale of supportedLocales) {
    it(`provides complete page labels for ${locale}`, () => {
      const i18n = createGooseI18n(locale);
      const keys = [
        "site.drafts:title",
        "site.accessGroups:joinTitle",
        "site.moderation:managementTabs.reports",
        "site.publish:toolbar.table",
        "site.topic:reportReasons.spam",
        "site.themePreview:workflowDescription",
        "shared.serverMessages:page.notFound",
      ] as const;
      for (const key of keys) {
        expect(i18n.exists(key)).toBe(true);
        expect(i18n.t(key)).not.toBe(key);
      }
    });
  }
});
