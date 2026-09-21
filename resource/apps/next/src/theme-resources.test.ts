import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { NextThemeResources } from "./theme-resources";

describe("Next theme resources", () => {
  it("renders the configured stylesheet and browser theme color", () => {
    const html = renderToStaticMarkup(
      createElement(NextThemeResources, {
        theme: {
          enabled: true,
          href: "/site-theme.css?v=theme-version",
          current: "gf-light",
          themeColor: "#fbfdff",
        },
      }),
    );

    expect(html).toContain('name="theme-color"');
    expect(html).toContain('content="#fbfdff"');
    expect(html).toContain('id="goose-site-theme-link"');
    expect(html).toContain('href="/site-theme.css?v=theme-version"');
    expect(html).toContain('data-precedence="goose-theme"');
  });

  it("omits the stylesheet when custom themes are disabled", () => {
    const html = renderToStaticMarkup(
      createElement(NextThemeResources, {
        theme: {
          enabled: false,
          current: "gf-dark",
          themeColor: "#1b1b1b",
        },
      }),
    );

    expect(html).toContain('content="#1b1b1b"');
    expect(html).not.toContain("goose-site-theme-link");
  });
});
