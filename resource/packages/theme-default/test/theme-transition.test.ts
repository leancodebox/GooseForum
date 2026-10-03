import { afterEach, describe, expect, it, vi } from "vitest";
import { createThemeTransition } from "../src/app/theme-transition";

const original = Object.getOwnPropertyDescriptor(document, "startViewTransition");
afterEach(() => {
  if (original) Object.defineProperty(document, "startViewTransition", original);
  else Reflect.deleteProperty(document, "startViewTransition");
  document.documentElement.classList.remove("goose-theme-transition");
  vi.unstubAllGlobals();
});

function deferred() {
  let resolve!: () => void;
  let reject!: (reason: Error) => void;
  const promise = new Promise<void>((ok, fail) => { resolve = ok; reject = fail; });
  return { promise, resolve, reject };
}

function native() {
  const transitions: { update: () => void; skipTransition: ReturnType<typeof vi.fn>; done: ReturnType<typeof deferred>; finished: ReturnType<typeof deferred> }[] = [];
  const start = vi.fn((update: () => void) => {
    const done = deferred();
    const finished = deferred();
    const skipTransition = vi.fn();
    transitions.push({ update, skipTransition, done, finished });
    return { updateCallbackDone: done.promise, finished: finished.promise, skipTransition };
  });
  Object.defineProperty(document, "startViewTransition", { configurable: true, value: start });
  vi.stubGlobal("matchMedia", vi.fn(() => ({ matches: false })));
  return { start, transitions };
}

async function settle() { await Promise.resolve(); await Promise.resolve(); await Promise.resolve(); }

describe("theme snapshots", () => {
  it("keeps the latest choice when old callbacks or cleanup arrive late", async () => {
    const { transitions } = native();
    const animation = createThemeTransition();
    const dark = vi.fn();
    const light = vi.fn();
    animation.apply(dark);
    animation.apply(light);
    expect(transitions[0].skipTransition).toHaveBeenCalledOnce();
    transitions[0].update();
    transitions[0].finished.resolve();
    await settle();
    expect(dark).not.toHaveBeenCalled();
    expect(document.documentElement.classList.contains("goose-theme-transition")).toBe(true);
    transitions[1].update();
    transitions[1].done.resolve();
    transitions[1].finished.resolve();
    await settle();
    expect(light).toHaveBeenCalledOnce();
    expect(document.documentElement.classList.contains("goose-theme-transition")).toBe(false);
  });

  it("applies the choice even if snapshot creation fails", async () => {
    const { transitions } = native();
    const update = vi.fn();
    createThemeTransition().apply(update);
    transitions[0].done.reject(new Error("snapshot unavailable"));
    transitions[0].finished.reject(new Error("snapshot unavailable"));
    await settle();
    transitions[0].update();
    expect(update).toHaveBeenCalledOnce();
    expect(document.documentElement.classList.contains("goose-theme-transition")).toBe(false);
  });

  it("does not apply a delayed choice after disposal", async () => {
    const { transitions } = native();
    const animation = createThemeTransition();
    const update = vi.fn();
    animation.apply(update);
    animation.cancel();
    transitions[0].update();
    transitions[0].done.reject(new Error("cancelled"));
    transitions[0].finished.resolve();
    await settle();
    expect(update).not.toHaveBeenCalled();
    expect(document.documentElement.classList.contains("goose-theme-transition")).toBe(false);
  });

  it("updates immediately without native support, with reduced motion, or when animation is disabled", () => {
    Object.defineProperty(document, "startViewTransition", { configurable: true, value: undefined });
    const update = vi.fn();
    const animation = createThemeTransition();
    animation.apply(update);
    const { start } = native();
    vi.stubGlobal("matchMedia", vi.fn(() => ({ matches: true })));
    animation.apply(update);
    vi.stubGlobal("matchMedia", vi.fn(() => ({ matches: false })));
    animation.apply(update, false);
    expect(update).toHaveBeenCalledTimes(3);
    expect(start).not.toHaveBeenCalled();
  });
});
