import { act, renderHook } from "@testing-library/react";
import { expect, it } from "vitest";
import { useLatestRequest } from "@gooseforum/runtime";

it("invalidates older requests and requests completing after unmount", () => {
  const { result, unmount } = renderHook(useLatestRequest);
  let first!: () => boolean;
  let second!: () => boolean;
  act(() => {
    first = result.current();
    second = result.current();
  });
  expect(first()).toBe(false);
  expect(second()).toBe(true);
  unmount();
  expect(second()).toBe(false);
});
