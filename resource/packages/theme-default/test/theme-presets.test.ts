import { describe, expect, it } from "vitest";
import { themePresets } from "../src/site/theme/theme-presets";

describe("theme presets", () => {
  it("keeps the default React canvas aligned with the browser theme color", () => {
    const goose = themePresets.find((preset) => preset.key === "goose");

    expect(goose?.themes["gf-light"]["color-base-100"]).toBe("#fbfdff");
  });
});
