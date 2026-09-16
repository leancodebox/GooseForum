import { afterEach, describe, expect, it, vi } from "vitest";
import { NextRequest } from "next/server";
import { proxyToGooseForum } from "./upstream-proxy";

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
});

describe("Next runtime proxy", () => {
  it("uses the runtime origin and preserves redirects and cookies", async () => {
    vi.stubEnv("GOOSEFORUM_ORIGIN", "http://127.0.0.1:5234");
    const fetch = vi.fn(
      (_input: RequestInfo | URL) =>
        Promise.resolve(
          new Response(null, {
            status: 302,
            headers: [
              ["Location", "http://127.0.0.1:5234/login?next=%2Fadmin"],
              ["Set-Cookie", "session=one; Path=/; HttpOnly"],
              ["Set-Cookie", "lang=zh; Path=/"],
            ],
          }),
        ),
    );
    vi.stubGlobal("fetch", fetch);

    const response = await proxyToGooseForum(
      new NextRequest("http://next.local/api/session?fresh=1", {
        headers: { Cookie: "existing=yes", Connection: "keep-alive" },
      }),
    );

    expect(fetch.mock.calls[0]?.[0]).toEqual(
      new URL("http://127.0.0.1:5234/api/session?fresh=1"),
    );
    const requestHeaders = new Headers(fetch.mock.calls[0]?.[1]?.headers);
    expect(requestHeaders.get("cookie")).toBe("existing=yes");
    expect(requestHeaders.has("connection")).toBe(false);
    expect(response.headers.get("location")).toBe("/login?next=%2Fadmin");
    expect(response.headers.getSetCookie()).toEqual([
      "session=one; Path=/; HttpOnly",
      "lang=zh; Path=/",
    ]);
  });

  it("returns a clear error when no runtime origin is configured", async () => {
    vi.stubEnv("GOOSEFORUM_ORIGIN", "");
    const response = await proxyToGooseForum(
      new NextRequest("http://next.local/static/logo.webp"),
    );
    expect(response.status).toBe(503);
  });
});
