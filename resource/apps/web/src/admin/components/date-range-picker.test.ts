import { describe, expect, it } from "vitest";
import { presetDateRange } from "./date-range";

describe("dashboard date range presets", () => {
  it("returns inclusive local calendar ranges", () => {
    const today = new Date(2026, 8, 14, 22, 30);

    expect(presetDateRange(7, today)).toEqual({
      start: "2026-09-08",
      end: "2026-09-14",
    });
    expect(presetDateRange(30, today)).toEqual({
      start: "2026-08-16",
      end: "2026-09-14",
    });
    expect(presetDateRange(90, today)).toEqual({
      start: "2026-06-17",
      end: "2026-09-14",
    });
  });
});
