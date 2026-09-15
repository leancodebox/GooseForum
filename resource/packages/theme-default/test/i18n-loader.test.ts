import { beforeEach, describe, expect, it, vi } from "vitest";

beforeEach(() => {
  vi.resetModules();
});
describe("on-demand C translations", () => {
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
