import { StrictMode } from "react";
import { afterEach, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { I18nextProvider } from "react-i18next";
import { createGooseI18n } from "@gooseforum/runtime/i18n";
import { RenderedContent } from "../src/site/content/rendered-content";

afterEach(cleanup);

it("shows loaded image metadata and delegates the accessible control to the existing image click", async () => {
  const click = vi.fn();
  const i18n = createGooseI18n("en");
  const view = render(<StrictMode><I18nextProvider i18n={i18n}>
    <div onClick={event => { if ((event.target as Element).tagName === "IMG") { event.preventDefault(); click(); } }}>
      <RenderedContent html={'<p><a href="/original">\n<img src="/uploads/test%20photo.png" alt="A photo" data-file-size="112640">\n</a></p>'} />
    </div>
  </I18nextProvider></StrictMode>);
  const image = view.container.querySelector("img")!;
  Object.defineProperties(image, { naturalWidth: { value: 540 }, naturalHeight: { value: 804 } });
  fireEvent.load(image);
  expect(screen.getByText("test photo.png")).toBeTruthy();
  expect(screen.getByText("540 × 804 · 110 KB")).toBeTruthy();
  expect(view.container.querySelectorAll(".gf-content-image")).toHaveLength(1);
  expect(view.container.querySelector("a button")).toBeNull();
  await userEvent.setup().click(screen.getByRole("button"));
  expect(click).toHaveBeenCalledOnce();
});

it("replaces image enhancements when HTML changes and leaves announcements alone", () => {
  const i18n = createGooseI18n("en");
  const content = (html: string, variant: "post" | "announcement" = "post") => (
    <I18nextProvider i18n={i18n}><RenderedContent html={html} variant={variant} /></I18nextProvider>
  );
  const view = render(content('<p><img src="/first.png"></p>'));
  expect(screen.getByText("first.png")).toBeTruthy();
  view.rerender(content('<p><img src="/second.png"></p>'));
  expect(screen.queryByText("first.png")).toBeNull();
  expect(screen.getByText("second.png")).toBeTruthy();
  expect(view.container.querySelectorAll(".gf-content-image")).toHaveLength(1);
  view.rerender(content('<p><img src="/second.png"></p>', "announcement"));
  expect(view.container.querySelector(".gf-content-image")).toBeNull();
  expect(screen.queryByRole("button")).toBeNull();
});
