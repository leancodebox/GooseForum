import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useAnnouncementDisclosure } from "../src/site/pages/use-announcement-disclosure";

const key = "goose:announcement:disclosure";
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  localStorage.removeItem(key);
});

describe("announcement disclosure", () => {
  it("restores the choice without animation and opens a new announcement", () => {
    const first = renderHook(() => useAnnouncementDisclosure("version-1"));
    expect(first.result.current.open).toBe(true);
    expect(first.result.current.animate).toBe(false);
    act(() => first.result.current.setOpen(false));
    expect(first.result.current.open).toBe(false);
    expect(first.result.current.animate).toBe(true);
    first.unmount();
    const restored = renderHook(({ version }) => useAnnouncementDisclosure(version), {
      initialProps: { version: "version-1" },
    });
    expect(restored.result.current.open).toBe(false);
    expect(restored.result.current.animate).toBe(false);
    restored.rerender({ version: "version-2" });
    expect(restored.result.current.open).toBe(true);
    expect(restored.result.current.animate).toBe(false);
  });

  it("only animates user actions and clears the animation flag after completion", () => {
    localStorage.setItem(key, JSON.stringify({ version: "version-1", open: true }));
    const { result } = renderHook(() => useAnnouncementDisclosure("version-1"));
    expect(result.current.animate).toBe(false);
    act(() => result.current.setOpen(false));
    act(() => result.current.finishAnimation());
    expect(result.current.animate).toBe(false);
    act(() => result.current.setOpen(true));
    expect(result.current.animate).toBe(true);
    expect(JSON.parse(localStorage.getItem(key)!)).toEqual({ version: "version-1", open: true });
  });

  it("ignores corrupt storage and remains usable when storage is blocked", () => {
    localStorage.setItem(key, "invalid-json");
    const { result } = renderHook(() => useAnnouncementDisclosure("version-1"));
    expect(result.current.open).toBe(true);
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => { throw new Error("blocked"); });
    act(() => result.current.setOpen(false));
    expect(result.current.open).toBe(false);
  });
});
