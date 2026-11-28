import { beforeEach, describe, expect, it, vi } from "vitest";

beforeEach(() => {
  vi.resetModules();
});
describe("on-demand C translations", () => {
  it.each(["zh", "en", "ja", "it", "ru"] as const)("loads attachment errors and publishing labels in %s", async locale => {
    const loader = await import("@gooseforum/runtime/i18n/loader");
    await loader.prepareGooseTranslations(locale, ["site.publish"]);
    const resources = loader.cachedGooseResources()[locale];
    expect(resources["shared.serverMessages"]).toHaveProperty("attachment.invalidFile");
    expect(resources["site.publish"]).toHaveProperty("attachmentUpload");
    const expected = {
      zh: { invalidFile: "文件类型不支持，或文件名、内容不合法。", upload: "上传附件" },
      en: { invalidFile: "Unsupported file type or invalid filename/content.", upload: "Upload attachment" },
      ja: { invalidFile: "未対応のファイル形式、または無効なファイル名・内容です。", upload: "ファイルを添付" },
      ru: { invalidFile: "Неподдерживаемый тип файла, неверное имя или содержимое.", upload: "Загрузить вложение" },
      it: { invalidFile: "Tipo di file non supportato o nome/contenuto non valido.", upload: "Carica allegato" },
    }[locale];
    expect(resources["shared.serverMessages"]).toMatchObject({ "attachment.invalidFile": expected.invalidFile });
    expect(resources["site.publish"]).toMatchObject({ attachmentUpload: expected.upload });
  });
  it.each(["zh", "en", "ja", "it", "ru"] as const)("preserves moderation action dictionaries in %s", async (locale) => {
    const loader = await import("@gooseforum/runtime/i18n/loader");
    const { siteLoaders } = await import("@gooseforum/i18n/site");
    await loader.prepareGooseTranslations(locale, ["site.moderation"]);
    const canonical = (await siteLoaders[`${locale}/site.moderation`]()).default;
    expect(loader.cachedGooseResources()[locale]["site.moderation"]).toEqual(canonical);
    const { createGooseI18n } = await import("@gooseforum/runtime/i18n");
    expect(createGooseI18n(locale).t("site.moderation:blocked.restore")).not.toBe("blocked.restore");
  });
  it("loads only the requested language and page namespaces on the home page", async () => {
    const loader = await import("@gooseforum/runtime/i18n/loader");
    expect(loader.cachedGooseResources()).toEqual({});
    await loader.prepareGooseTranslations(
      "zh",
      loader.goosePageNamespaces("home.index"),
    );
    const resources = loader.cachedGooseResources();
    expect(Object.keys(resources)).toEqual(["zh"]);
    expect(Object.keys(resources.zh).sort()).toEqual(
      [...loader.commonNamespaces].sort(),
    );
    expect(resources.zh["site.topic"]).toBeUndefined();
    expect(resources.zh["site.settings"]).toBeUndefined();
    // The language picker must remain usable without loading every language's auth dictionary.
    const { localeLabels } = await import('@gooseforum/client/i18n/locale');
    expect(localeLabels.en.label).toBe('English');
    expect(localeLabels.ja.label).toBe('日本語');
  });
  it("deduplicates concurrent namespace loads and reuses completed dictionaries", async () => {
    const loader = await import("@gooseforum/runtime/i18n/loader");
    const { siteLoaders } = await import("@gooseforum/i18n/site");
    const topic = vi.spyOn(siteLoaders, "zh/site.topic");
    await Promise.all([
      loader.prepareGooseTranslations("zh", ["site.topic"]),
      loader.prepareGooseTranslations("zh", ["site.topic"]),
    ]);
    await loader.prepareGooseTranslations("zh", ["site.topic"]);
    expect(topic).toHaveBeenCalledTimes(1);
    expect(loader.cachedGooseResources().zh["site.topic"]).toHaveProperty(
      "replyTargetUnavailable",
      "原回复不可见",
    );
  });
  it("loads Chinese fallback alongside a selected language, and can retry a failed import", async () => {
    const loader = await import("@gooseforum/runtime/i18n/loader");
    const { siteLoaders } = await import("@gooseforum/i18n/site");
    const settings = vi
      .spyOn(siteLoaders, "en/site.settings")
      .mockRejectedValueOnce(new Error("Temporary failure"));
    await expect(
      loader.prepareGooseTranslations("en", ["site.settings"]),
    ).rejects.toThrow("Temporary failure");
    await loader.prepareGooseTranslations("en", ["site.settings"]);
    expect(settings).toHaveBeenCalledTimes(2);
    expect(Object.keys(loader.cachedGooseResources()).sort()).toEqual([
      "en",
      "zh",
    ]);
    expect(loader.cachedGooseResources().en["site.settings"]).toBeTruthy();
    expect(loader.cachedGooseResources().zh["site.settings"]).toBeTruthy();
  });
  it("formats an English ICU fallback using English plural rules on a Japanese page", async () => {
    const { createGooseI18n } = await import("@gooseforum/runtime/i18n");
    const { prepareIcuFormat } = await import("../../i18n/src/format");
    await prepareIcuFormat();
    const i18n = createGooseI18n("ja", {
      en: { "site.home": { topic: "{count, plural, one {# topic} other {# topics}}" } },
    });
    expect(i18n.t("site.home:topic", { count: 1 })).toBe("1 topic");
    expect(i18n.t("site.home:topic", { count: 2 })).toBe("2 topics");
  });
  it("notifies mounted providers before preparation resolves and unsubscribes cleanly", async () => {
    const loader = await import("@gooseforum/runtime/i18n/loader");
    const listener = vi.fn();
    const unsubscribe = loader.subscribeGooseResources(listener);
    await loader.prepareGooseTranslations("zh");
    expect(listener).toHaveBeenCalledWith("zh", "site.home", expect.any(Object));
    unsubscribe();
    listener.mockClear();
    await loader.prepareGooseTranslations("zh", ["site.search"]);
    expect(listener).not.toHaveBeenCalled();
  });
});
