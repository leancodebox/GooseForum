import { afterEach, describe, expect, it, vi } from "vitest";
import { createNextPageAdapter } from "./page-adapter";
import { createDemoPage } from "./demo-payload";

afterEach(() => vi.unstubAllGlobals());

describe("Next page adapter", () => {
  it("runs without an upstream in demo mode", async () => {
    const adapter = createNextPageAdapter("");
    const page = await adapter.load({ path: "/?lang=en" });

    expect(adapter.mode).toBe("demo");
    expect(page.component).toBe("home.index");
    expect(page.url).toBe("/?lang=en");
  });

  it("forwards page protocol headers to an optional upstream", async () => {
    const fetch = vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
      const headers = new Headers(init?.headers);
      expect(headers.get("X-Goose-Page")).toBe("true");
      expect(headers.get("Cookie")).toBe("session=test");
      expect(headers.get("Accept-Language")).toBe("en-US");
      return Promise.resolve(Response.json(createDemoPage("/categories")));
    });
    vi.stubGlobal("fetch", fetch);

    const adapter = createNextPageAdapter("http://127.0.0.1:5234/base");
    const page = await adapter.load({
      path: "/categories",
      cookie: "session=test",
      acceptLanguage: "en-US",
    });

    expect(adapter.mode).toBe("upstream");
    expect(page.url).toBe("/categories");
    expect(fetch.mock.calls[0]?.[0]).toBe(
      "http://127.0.0.1:5234/categories",
    );
  });
});
