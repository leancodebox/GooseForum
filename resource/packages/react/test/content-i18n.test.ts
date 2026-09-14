import { describe, expect, it } from "vitest";
import { supportedLocales } from "@gooseforum/client/i18n";
import { createGooseI18n } from "../src/i18n";

describe("migrated content page translations", () => {
  for (const locale of supportedLocales) {
    it(`provides complete page labels for ${locale}`, () => {
      const i18n = createGooseI18n(locale);
      const keys = [
        "drafts:title",
        "accessGroups:joinTitle",
        "moderation:managementTabs.reports",
        "publish:toolbar.table",
        "topic:reportReasons.spam",
        "themePreview:workflowDescription",
        "serverMessages:page.notFound",
      ];
      for (const key of keys) {
        expect(i18n.exists(key)).toBe(true);
        expect(i18n.t(key)).not.toBe(key);
      }
    });
  }
});
