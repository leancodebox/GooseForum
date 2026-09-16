import { afterEach, expect, it } from "vitest";
import { cleanup, render } from "@testing-library/react";
import HttpNotifyGuide from "./http-notify-guide";

afterEach(cleanup);

it("renders complete Markdown and switches documentation with the locale", () => {
  const { container, rerender } = render(<HttpNotifyGuide locale="zh" />);
  expect(container.querySelectorAll("h3").length).toBeGreaterThan(4);
  expect(container.textContent).toContain("HMAC_SHA256");
  expect(container.querySelector("pre code.language-json")).not.toBeNull();
  rerender(<HttpNotifyGuide locale="en" />);
  expect(container.textContent).toContain("Signature verification");
  rerender(<HttpNotifyGuide locale="ja" />);
  expect(container.textContent).toContain("リクエスト");
  rerender(<HttpNotifyGuide locale="it" />);
  expect(container.textContent).toContain("Signature verification");
});
