import { describe, expect, it } from "vitest";
import { pageDataRequestPath } from "./next-goose-app";

describe("Next SPA page data routing", () => {
  it("requests Go directly when the Next process is managed by Go", () => {
    expect(pageDataRequestPath("/categories?page=2", true)).toBe(
      "/categories?page=2",
    );
  });

  it("uses the Next adapter when running standalone", () => {
    expect(pageDataRequestPath("/categories?page=2")).toBe(
      "/goose-page-data?path=%2Fcategories%3Fpage%3D2",
    );
  });
});
