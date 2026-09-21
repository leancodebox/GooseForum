import { afterEach, describe, expect, it, vi } from "vitest";
import { GET } from "../app/goose-internal/health/route";

describe("managed Next health route", () => {
  afterEach(() => vi.unstubAllEnvs());

  it("returns the process health token without caching", () => {
    vi.stubEnv("GOOSEFORUM_HEALTH_TOKEN", "launch-token");

    const response = GET();

    expect(response.status).toBe(204);
    expect(response.headers.get("X-Goose-Health-Token")).toBe("launch-token");
    expect(response.headers.get("Cache-Control")).toBe("private, no-store");
  });

  it("is unavailable outside the managed process", () => {
    vi.stubEnv("GOOSEFORUM_HEALTH_TOKEN", "");

    expect(GET().status).toBe(404);
  });
});
