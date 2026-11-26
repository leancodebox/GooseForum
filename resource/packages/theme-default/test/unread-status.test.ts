import { act, renderHook } from "@testing-library/react";
import { expect, it } from "vitest";
import { announceUnreadStatus, useUnreadStatus } from "@gooseforum/runtime/unread-status";

it("keeps the unread snapshot on identical polls and merges partial changes", () => {
  const initial = { notifications: true, messages: false };
  const { result } = renderHook(() => useUnreadStatus(initial));
  const first = result.current;

  act(() => announceUnreadStatus({ ...initial }));
  expect(result.current).toBe(first);

  act(() => announceUnreadStatus({ messages: true }));
  expect(result.current).toEqual({ notifications: true, messages: true });
});
