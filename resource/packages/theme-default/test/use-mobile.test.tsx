import { act, cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { useIsMobile } from "@gooseforum/ui/hooks/use-mobile";

function Harness() {
  return <output>{useIsMobile() ? "mobile" : "desktop"}</output>;
}

let notifyMediaChange: (() => void) | undefined;

beforeEach(() => {
  Object.defineProperty(window, "innerWidth", { configurable: true, value: 1024, writable: true });
  vi.stubGlobal("matchMedia", vi.fn(() => ({
    addEventListener: (_event: string, listener: () => void) => { notifyMediaChange = listener; },
    removeEventListener: vi.fn(),
  })));
});

afterEach(() => {
  cleanup();
  notifyMediaChange = undefined;
  vi.unstubAllGlobals();
});

it("uses the current viewport snapshot and follows media changes", () => {
  render(<Harness />);
  expect(screen.getByRole("status").textContent).toBe("desktop");

  act(() => {
    window.innerWidth = 640;
    notifyMediaChange?.();
  });
  expect(screen.getByRole("status").textContent).toBe("mobile");
});
