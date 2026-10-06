import { afterEach, expect, it } from "vitest";
import { cleanup, render, screen, act } from "@testing-library/react";
import { I18nextProvider } from "react-i18next";
import { createGooseI18n } from "@gooseforum/runtime/i18n";
import { SponsorsPageView } from "../src/site/pages/sponsors-page";

afterEach(cleanup);

it("updates sponsor tier labels when the language changes without replacing page data", async () => {
  const i18n = createGooseI18n("zh");
  render(<I18nextProvider i18n={i18n}><SponsorsPageView page={{
    totalCount: 4,
    content: { title: "Original title", description: "Original description" },
    contact: { title: "Contact", description: "Original contact", buttonText: "Email", buttonLink: "mailto:test@example.com" },
    rules: [],
    sections: ["diamond", "gold", "silver", "supporter"].map(key => ({
      key, tone: key, label: "English payload label",
      sponsors: [{ name: `Sponsor ${key}`, message: "Original message", avatarUrl: "/avatar.webp", link: "" }],
    })),
  }} /></I18nextProvider>);
  for (const [locale, labels] of [
    ["zh", ["钻石合作伙伴", "金牌赞助商", "银牌赞助商", "支持者"]],
    ["en", ["Diamond Partners", "Gold Sponsors", "Silver Sponsors", "Supporters"]],
    ["ja", ["ダイヤモンドパートナー", "ゴールドスポンサー", "シルバースポンサー", "サポーター"]],
    ["it", ["Partner diamante", "Sponsor oro", "Sponsor argento", "Sostenitori"]],
  ] as const) {
    await act(() => i18n.changeLanguage(locale));
    for (const label of labels) expect(screen.getByRole("heading", { name: label })).toBeTruthy();
    expect(screen.getByRole("heading", { name: "Original title" })).toBeTruthy();
    expect(screen.getAllByText("Original message")).toHaveLength(4);
  }
});
