import { useState } from "react";
import { afterEach, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { AttachmentInfo, GooseSiteApi } from "@gooseforum/client";
import { GooseRuntimeProvider, type GooseRuntime } from "@gooseforum/runtime";
import { GooseI18nProvider } from "@gooseforum/runtime/i18n";
import { AttachmentEditor } from "../src/site/editor/attachment-editor";

afterEach(cleanup);
const attachment = { id: 9, filename: "notes.txt", size: 42, url: "/file/attachments/9" };
function show({ enabled = true, files = [] as AttachmentInfo[], upload = vi.fn().mockResolvedValue(attachment) } = {}) {
  const runtime = {
    api: { attachments: { policy: vi.fn().mockResolvedValue({ enabled, maxFilesPerPost: 5, maxFileSizeKb: 5120, extensions: [".txt"] }), upload } } as unknown as GooseSiteApi,
    currentUrl: "/publish", locale: "en", theme: "gf-light", isNavigating: false,
    navigate: vi.fn(), redirect: vi.fn(), refresh: vi.fn(), queueFlash: vi.fn(), setLocale: vi.fn(), toggleTheme: vi.fn(),
  } satisfies GooseRuntime;
  function Editor() {
    const [value, setValue] = useState(files);
    const [busy, setBusy] = useState(false);
    const [disabled, setDisabled] = useState(false);
    return <><button onClick={() => setDisabled(!disabled)}>Toggle disabled</button><button disabled={busy}>Submit</button><AttachmentEditor files={value} onChange={setValue} onBusy={setBusy} disabled={disabled} /></>;
  }
  render(<GooseI18nProvider locale="en"><GooseRuntimeProvider runtime={runtime}><Editor /></GooseRuntimeProvider></GooseI18nProvider>);
  return upload;
}

it("disables new uploads under a disabled policy while retaining removal of existing files", async () => {
  const upload = show({ enabled: false, files: [attachment] });
  const input = await screen.findByLabelText<HTMLInputElement>("Upload attachment");
  expect(input.disabled).toBe(true);
  await userEvent.upload(input, new File(["hello"], "notes.txt", { type: "text/plain" }));
  expect(upload).not.toHaveBeenCalled();
  await userEvent.click(screen.getByRole("button", { name: /Remove.*notes.txt/ }));
  expect(screen.queryByRole("link", { name: "notes.txt" })).toBeNull();
});

it("blocks submission after upload failure, honors disabled controls, and retries the original file", async () => {
  const upload = show({ upload: vi.fn().mockRejectedValueOnce(new Error("offline")).mockResolvedValue(attachment) });
  const file = new File(["hello"], "notes.txt", { type: "text/plain" });
  await userEvent.upload(await screen.findByLabelText("Upload attachment"), file);
  const retry = await screen.findByRole<HTMLButtonElement>("button", { name: "Retry upload" });
  expect(screen.getByRole<HTMLButtonElement>("button", { name: "Submit" }).disabled).toBe(true);
  await userEvent.click(screen.getByText("Toggle disabled"));
  expect(retry.disabled).toBe(true);
  expect(screen.getByRole<HTMLButtonElement>("button", { name: "Remove" }).disabled).toBe(true);
  await userEvent.click(retry);
  expect(upload).toHaveBeenCalledTimes(1);
  await userEvent.click(screen.getByText("Toggle disabled"));
  await userEvent.click(retry);
  await screen.findByRole("link", { name: "notes.txt" });
  expect(upload).toHaveBeenLastCalledWith(file);
  await waitFor(() => expect(screen.getByRole<HTMLButtonElement>("button", { name: "Submit" }).disabled).toBe(false));
});

it("allows removing an unresolved upload failure to unblock submission", async () => {
  show({ upload: vi.fn().mockRejectedValue(new Error("offline")) });
  await userEvent.upload(await screen.findByLabelText("Upload attachment"), new File(["hello"], "notes.txt", { type: "text/plain" }));
  await screen.findByRole("alert");
  await userEvent.click(screen.getByRole("button", { name: "Remove" }));
  expect(screen.queryByRole("alert")).toBeNull();
  expect(screen.getByRole<HTMLButtonElement>("button", { name: "Submit" }).disabled).toBe(false);
});
