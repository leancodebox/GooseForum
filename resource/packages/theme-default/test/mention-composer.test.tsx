import { useState } from "react";
import { afterEach, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { GooseSiteApi } from "@gooseforum/client";
import { GooseRuntimeProvider, type GooseRuntime } from "@gooseforum/runtime";
import { GooseI18nProvider } from "@gooseforum/runtime/i18n";
import { MarkdownComposer } from "../src/site/editor/markdown-composer";

afterEach(cleanup);

function show() {
  const mentions = vi.fn().mockResolvedValue([{ id: "12", username: "alice" }]);
  const runtime = {
    api: { users: { mentions } } as unknown as GooseSiteApi,
    currentUrl: "/publish", locale: "en", theme: "gf-light", isNavigating: false,
    navigate: vi.fn(), redirect: vi.fn(), refresh: vi.fn(), queueFlash: vi.fn(),
    setLocale: vi.fn(), toggleTheme: vi.fn(),
  } satisfies GooseRuntime;
  function Editor() {
    const [value, setValue] = useState("");
    const [version, setVersion] = useState<0 | 1>(1);
    return <>
      <button type="button" onClick={() => setVersion(0)}>Legacy syntax</button>
      <MarkdownComposer value={value} onChange={setValue} sourceVersion={version} />
    </>;
  }
  render(<GooseI18nProvider locale="en"><GooseRuntimeProvider runtime={runtime}><Editor /></GooseRuntimeProvider></GooseI18nProvider>);
  return mentions;
}

async function sourceWithCandidate() {
  const mentions = show();
  await userEvent.click(screen.getByRole("radio", { name: "Markdown" }));
  const source = screen.getByRole<HTMLTextAreaElement>("textbox");
  await userEvent.type(source, "@ali");
  await screen.findByRole("option", { name: "@alice" });
  expect(mentions).toHaveBeenCalledWith("ali");
  return source;
}

it("leaves IME confirmation to the input and places a normal mention insertion's caret after the tag", async () => {
  const source = await sourceWithCandidate();
  const composition = new KeyboardEvent("keydown", { key: "Enter", isComposing: true, bubbles: true, cancelable: true });
  fireEvent(source, composition);
  expect(composition.defaultPrevented).toBe(false);
  expect(source.value).toBe("@ali");
  expect(screen.getByRole("option", { name: "@alice" })).toBeTruthy();
  const composingKeyCode = new KeyboardEvent("keydown", { key: "Enter", keyCode: 229, bubbles: true, cancelable: true });
  fireEvent(source, composingKeyCode);
  expect(composingKeyCode.defaultPrevented).toBe(false);
  expect(source.value).toBe("@ali");

  fireEvent.keyDown(source, { key: "Enter" });
  const tag = '[mention user="12"]@alice[/mention]';
  expect(source.value).toBe(tag);
  await waitFor(() => {
    expect(source.selectionStart).toBe(tag.length);
    expect(source.selectionEnd).toBe(tag.length);
  });
});

it("does not insert a hidden candidate after switching to legacy syntax", async () => {
  const source = await sourceWithCandidate();
  await userEvent.click(screen.getByRole("button", { name: "Legacy syntax" }));
  expect(screen.queryByRole("option", { name: "@alice" })).toBeNull();
  const enter = new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true });
  fireEvent(source, enter);
  expect(enter.defaultPrevented).toBe(false);
  expect(source.value).toBe("@ali");
});
