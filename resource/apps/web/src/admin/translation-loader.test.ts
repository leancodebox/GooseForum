import { beforeEach, describe, expect, it, vi } from "vitest";
import { supportedLocales } from "@gooseforum/client/i18n/locale";
import { adminNavGroups } from "./nav";

beforeEach(() => {
  vi.resetModules();
});
describe("on-demand admin translations", () => {
  it.each(supportedLocales)("prepares attachment permission labels on the trust route in %s", async locale => {
    const loader = await import("./translation-loader");
    const { getAdminDictionary } = await import("@gooseforum/i18n/admin");
    await loader.prepareAdminTranslations(locale, loader.adminPageNamespaces("/admin/settings/trust"));
    const dictionary = getAdminDictionary("admin.trust-settings", locale);
    await loader.prepareAdminTranslations("en", ["admin.trust-settings"]);
    const english = getAdminDictionary("admin.trust-settings", "en");
    expect(Object.keys(dictionary).sort()).toEqual(Object.keys(english).sort());
    expect(dictionary.allowAttachmentUpload).toBe({
      zh: "附件上传",
      en: "Attachments",
      ja: "ファイル添付",
      it: "Allegati",
      ru: "Вложения",
      ko: "파일 첨부",
      fr: "Pièces jointes",
    }[locale]);
    const { createTrustSettingsText } = await import("./trust-settings-i18n");
    expect(createTrustSettingsText(locale)("orderInvalid", { higher: "H", lower: "L", field: "F" })).not.toMatch(/\{\w+\}/);
  });
  it("loads the shell, shared header and dashboard for the home page and keeps other locales unloaded", async () => {
    const loader = await import("./translation-loader");
    const { getAdminDictionary } = await import("@gooseforum/i18n/admin");
    await loader.prepareAdminTranslations(
      "zh",
      loader.adminPageNamespaces("/admin"),
    );
    expect(loader.loadedAdminNamespaces().sort()).toEqual([
      "admin.dashboard",
      "admin.shell",
      "shared.shell",
    ]);
    expect(() => getAdminDictionary("admin.users", "zh")).toThrow(
      "not prepared",
    );
    expect(() => getAdminDictionary("admin.shell", "en")).toThrow(
      "not prepared",
    );
  });
  it("covers every sidebar route and preserves localized labels", async () => {
    const loader = await import("./translation-loader");
    const { getAdminDictionary } = await import("@gooseforum/i18n/admin");
    for (const group of adminNavGroups)
      for (const item of group.items) {
        expect(loader.adminPageNamespaces(item.url).length).toBeGreaterThan(0);
        await loader.prepareAdminTranslations(
          "it",
          loader.adminPageNamespaces(item.url),
        );
      }
    expect(
      getAdminDictionary("admin.users", "it").title,
    ).toBe("Utenti");
    expect(
      getAdminDictionary("admin.users", "it")
        .description,
    ).toBe("Cerca utenti e gestisci lo stato degli account, i ruoli e i distintivi.");
  });
  it("deduplicates concurrent imports, caches them and retries rejected imports", async () => {
    const loader = await import("./translation-loader");
    const { adminLoaders } = await import("@gooseforum/i18n/admin");
    const users = vi
      .spyOn(adminLoaders, "zh/admin.users")
      .mockRejectedValueOnce(new Error("Temporary failure"));
    await expect(
      loader.prepareAdminTranslations("zh", ["admin.users"]),
    ).rejects.toThrow("Temporary failure");
    await Promise.all([
      loader.prepareAdminTranslations("zh", ["admin.users"]),
      loader.prepareAdminTranslations("zh", ["admin.users"]),
    ]);
    await loader.prepareAdminTranslations("zh", ["admin.users"]);
    expect(users).toHaveBeenCalledTimes(2);
  });
  it("prepares the assets dictionary before rendering the badges route", async () => {
    const loader = await import("./translation-loader");
    const { getAdminDictionary } = await import("@gooseforum/i18n/admin");
    expect(loader.adminPageNamespaces("/admin/badges")).toEqual(["admin.assets"]);
    await loader.prepareAdminTranslations(
      "zh",
      loader.adminPageNamespaces("/admin/badges"),
    );
    expect(
      getAdminDictionary("admin.assets", "zh").badges,
    ).toBe("徽章");
  });
});
