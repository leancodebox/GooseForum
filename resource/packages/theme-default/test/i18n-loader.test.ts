import { beforeEach, describe, expect, it, vi } from "vitest";

beforeEach(() => {
  vi.resetModules();
});
describe("on-demand C translations", () => {
  it.each(["zh", "en", "ja", "it"] as const)("loads attachment errors and publishing labels in %s", async locale => {
    const loader = await import("@gooseforum/runtime/i18n/loader");
    await loader.prepareGooseTranslations(locale, ["publish"]);
    const resources = loader.cachedGooseResources()[locale];
    expect(resources.serverMessages).toHaveProperty("attachment.invalidFile");
    expect(resources.publish).toHaveProperty("attachmentUpload");
    const expected = {
      zh: { invalidFile: "文件类型不支持，或文件名、内容不合法。", upload: "上传附件" },
      en: { invalidFile: "Unsupported file type or invalid filename/content.", upload: "Upload attachment" },
      ja: { invalidFile: "未対応のファイル形式、または無効なファイル名・内容です。", upload: "ファイルを添付" },
      it: { invalidFile: "Tipo di file non supportato o nome/contenuto non valido.", upload: "Carica allegato" },
    }[locale];
    expect(resources.serverMessages).toMatchObject({ "attachment.invalidFile": expected.invalidFile });
    expect(resources.publish).toMatchObject({ attachmentUpload: expected.upload });
  });
  it.each(["zh", "en", "ja", "it"] as const)("preserves moderation action dictionaries in %s", async (locale) => {
    const loader = await import("@gooseforum/runtime/i18n/loader");
    const { resourceLoaders } = await import("@gooseforum/runtime/i18n/resource-loaders");
    await loader.prepareGooseTranslations(locale, ["moderation"]);
    const canonical = (await resourceLoaders[`${locale}/moderation`]()).default;
    expect(loader.cachedGooseResources()[locale].moderation).toHaveProperty("blocked.restore", canonical.blocked.restore);
    expect(loader.cachedGooseResources()[locale].moderation).toHaveProperty("blocked.emptyTitle", canonical.blocked.emptyTitle);
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
    expect(resources.zh.topic).toBeUndefined();
    expect(resources.zh.settings).toBeUndefined();
    // The language picker must remain usable without loading every language's auth dictionary.
    const { localeLabels } = await import('@gooseforum/client/i18n/locale');
    expect(localeLabels.en.label).toBe('English');
    expect(localeLabels.ja.label).toBe('日本語');
  });
  it("deduplicates concurrent namespace loads and reuses completed dictionaries", async () => {
    const loader = await import("@gooseforum/runtime/i18n/loader");
    const { resourceLoaders } = await import("@gooseforum/runtime/i18n/resource-loaders");
    const topic = vi.spyOn(resourceLoaders, "zh/topic");
    await Promise.all([
      loader.prepareGooseTranslations("zh", ["topic"]),
      loader.prepareGooseTranslations("zh", ["topic"]),
    ]);
    await loader.prepareGooseTranslations("zh", ["topic"]);
    expect(topic).toHaveBeenCalledTimes(1);
    expect(loader.cachedGooseResources().zh.topic).toHaveProperty(
      "replyTargetUnavailable",
      "原回复不可见",
    );
  });
  it("loads Chinese fallback alongside a selected language, and can retry a failed import", async () => {
    const loader = await import("@gooseforum/runtime/i18n/loader");
    const { resourceLoaders } = await import("@gooseforum/runtime/i18n/resource-loaders");
    const settings = vi
      .spyOn(resourceLoaders, "en/settings")
      .mockRejectedValueOnce(new Error("Temporary failure"));
    await expect(
      loader.prepareGooseTranslations("en", ["settings"]),
    ).rejects.toThrow("Temporary failure");
    await loader.prepareGooseTranslations("en", ["settings"]);
    expect(settings).toHaveBeenCalledTimes(2);
    expect(Object.keys(loader.cachedGooseResources()).sort()).toEqual([
      "en",
      "zh",
    ]);
    expect(loader.cachedGooseResources().en.settings).toBeTruthy();
    expect(loader.cachedGooseResources().zh.settings).toBeTruthy();
  });
  it("notifies mounted providers before preparation resolves and unsubscribes cleanly", async () => {
    const loader = await import("@gooseforum/runtime/i18n/loader");
    const listener = vi.fn();
    const unsubscribe = loader.subscribeGooseResources(listener);
    await loader.prepareGooseTranslations("zh");
    expect(listener).toHaveBeenCalledWith("zh", "home", expect.any(Object));
    unsubscribe();
    listener.mockClear();
    await loader.prepareGooseTranslations("zh", ["search"]);
    expect(listener).not.toHaveBeenCalled();
  });
});
