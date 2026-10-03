import { act, cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { GooseI18nProvider } from "@gooseforum/runtime/i18n";
import { AnnouncementPanel } from "../src/site/content/announcement-panel";

const announcement = {
  enabled: true, html: "", publishedAt: Date.parse("2026-10-03T10:00:00Z"),
  items: [{ id: "a", title: "First", html: "<p>First body</p>" }, { id: "b", title: "Second", html: "<p>Second body</p>" }],
};
function panel(value = announcement) {
  return <GooseI18nProvider locale="en"><AnnouncementPanel announcement={value} /></GooseI18nProvider>;
}
beforeEach(() => {
  localStorage.clear();
  vi.useFakeTimers();
  vi.setSystemTime(new Date("2026-10-03T11:00:00Z"));
  vi.spyOn(document, "visibilityState", "get").mockReturnValue("visible");
});
afterEach(() => { cleanup(); vi.useRealTimers(); vi.restoreAllMocks(); localStorage.clear(); });

it("keeps the current announcement until a reader explicitly switches it", () => {
  render(panel());
  act(() => { vi.advanceTimersByTime(60000); });
  expect(screen.getByText("First body")).toBeTruthy();
  expect(screen.queryByText("Second body")).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "Announcement 2: Second" }));
  act(() => { vi.advanceTimersByTime(60000); });
  expect(screen.getByText("Second body")).toBeTruthy();
  expect(screen.queryByText("First body")).toBeNull();
});

it("preserves the selection when folded and resets selection and read status after a precise total timestamp change", () => {
  const view = render(panel());
  fireEvent.click(screen.getByRole("button", { name: "Mark announcement as read" }));
  expect(localStorage.getItem("goose:announcement:last-read-published-at")).toBe(String(announcement.publishedAt));
  fireEvent.click(screen.getByRole("button", { name: "Announcement 2: Second" }));
  fireEvent.click(screen.getByRole("button", { name: "Collapse announcement" }));
  act(() => { vi.advanceTimersByTime(12000); });
  fireEvent.click(screen.getByRole("button", { name: "Expand announcement" }));
  expect(screen.getByText("Second body")).toBeTruthy();
  fireEvent.click(screen.getByRole("button", { name: "Collapse announcement" }));
  view.rerender(panel({ ...announcement, publishedAt: announcement.publishedAt + 1 }));
  expect(screen.getByText("First body")).toBeTruthy();
  expect(screen.getByRole("button", { name: "Mark announcement as read" })).toBeTruthy();
  expect(document.querySelector('[data-animate="true"]')).toBeNull();
});

it("treats an empty item list as authoritative over legacy HTML", () => {
  render(panel({ ...announcement, html: "<p>Old notice</p>", items: [] }));
  expect(screen.queryByRole("complementary")).toBeNull();
});

it("preserves old date-string read markers when the payload uses milliseconds", () => {
  localStorage.setItem("goose:announcement:last-read-published-at", "2026-10-03T10:00:00Z");
  const view = render(panel());
  expect(screen.queryByRole("button", { name: "Mark announcement as read" })).toBeNull();
  view.rerender(panel({ ...announcement, publishedAt: announcement.publishedAt + 1 }));
  expect(screen.getByRole("button", { name: "Mark announcement as read" })).toBeTruthy();
});
