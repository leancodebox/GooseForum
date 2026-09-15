import { beforeEach, describe, expect, it, vi } from "vitest";
import { adminNavGroups } from "./nav";

beforeEach(() => {
  vi.resetModules();
});
describe("on-demand admin translations", () => {
  it("loads only shell and dashboard for the home page and keeps other locales unloaded", async () => {
    const loader = await import("./translation-loader");
    await loader.prepareAdminTranslations(
      "zh",
      loader.adminPageNamespaces("/admin"),
    );
    expect(loader.loadedAdminNamespaces().sort()).toEqual([
      "dashboard",
      "shell",
    ]);
    expect(() => loader.getAdminDictionary("users", "zh")).toThrow(
      "not prepared",
    );
    expect(() => loader.getAdminDictionary("shell", "en")).toThrow(
      "not prepared",
    );
  });
  it("covers every sidebar route and preserves translated and fallback labels", async () => {
    const loader = await import("./translation-loader");
    for (const group of adminNavGroups)
      for (const item of group.items) {
        expect(loader.adminPageNamespaces(item.url).length).toBeGreaterThan(0);
        await loader.prepareAdminTranslations(
          "it",
          loader.adminPageNamespaces(item.url),
        );
      }
    expect(
      loader.getAdminDictionary<{ title: string }>("users", "it").title,
    ).toBe("Utenti");
    expect(
      loader.getAdminDictionary<{ description: string }>("users", "it")
        .description,
    ).toBe("Search users and manage account status, roles, and badges.");
  });
  it("deduplicates concurrent imports, caches them and retries rejected imports", async () => {
    const loader = await import("./translation-loader");
    const { translationLoaders } = await import("./translation-resources");
    const users = vi
      .spyOn(translationLoaders, "zh/users")
      .mockRejectedValueOnce(new Error("Temporary failure"));
    await expect(
      loader.prepareAdminTranslations("zh", ["users"]),
    ).rejects.toThrow("Temporary failure");
    await Promise.all([
      loader.prepareAdminTranslations("zh", ["users"]),
      loader.prepareAdminTranslations("zh", ["users"]),
    ]);
    await loader.prepareAdminTranslations("zh", ["users"]);
    expect(users).toHaveBeenCalledTimes(2);
  });
  it("prepares the assets dictionary before rendering the badges route", async () => {
    const loader = await import("./translation-loader");
    expect(loader.adminPageNamespaces("/admin/badges")).toEqual(["assets"]);
    await loader.prepareAdminTranslations(
      "zh",
      loader.adminPageNamespaces("/admin/badges"),
    );
    expect(
      loader.getAdminDictionary<{ badges: string }>("assets", "zh").badges,
    ).toBe("徽章");
  });
});
