import { Suspense } from "react";
import { afterEach, expect, it } from "vitest";
import { act, cleanup, render, screen } from "@testing-library/react";
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";
import HttpNotifyGuide from "./http-notify-guide";

afterEach(cleanup);

function guide(locale: AuthLocale) {
  return <Suspense fallback="Loading"><HttpNotifyGuide locale={locale} /></Suspense>;
}

it("renders complete Markdown and switches documentation with the locale", async () => {
  const { container, rerender } = render(null);
  for (const [locale, title] of [
    ["zh", "HTTP 通知使用说明"],
    ["en", "Signature verification"],
    ["ja", "署名検証"],
    ["it", "Verifica della firma"],
    ["ko", "서명 확인"],
    ["fr", "Vérification de la signature"],
  ] as const) {
    await act(() => Promise.resolve(rerender(guide(locale))));
    expect(await screen.findByRole("heading", { name: title })).toBeTruthy();
    expect(container.textContent).toContain("HMAC_SHA256");
    expect(container.querySelectorAll("h3").length).toBeGreaterThan(4);
    expect(container.querySelector("pre code.language-json")).not.toBeNull();
  }
});
