import { afterEach, expect, it, vi } from "vitest";
import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { GooseAdminApi } from "@gooseforum/client";
import { FileResourcesManagementPage } from "./file-resources-management-page";

afterEach(() => { cleanup(); vi.unstubAllGlobals(); });

it("appends grid pages, retries failures, and resets pagination when switching to list", async () => {
  vi.stubGlobal("ResizeObserver", class { observe() {} disconnect() {} });
  let onIntersect: IntersectionObserverCallback;
  vi.stubGlobal("IntersectionObserver", class {
    constructor(callback: IntersectionObserverCallback) { onIntersect = callback; }
    observe() {}
    disconnect() {}
  });
  const file = (id: number) => ({ id, name: `file-${id}`, type: "application/pdf", url: `/files/${id}`, size: 100, userId: 1, createdAt: "" });
  const files = vi.fn()
    .mockResolvedValueOnce({ list: [file(1)], total: 40, page: 1 })
    .mockRejectedValueOnce(new Error("network failed"))
    .mockResolvedValueOnce({ list: [file(2)], total: 40, page: 2 })
    .mockResolvedValueOnce({ list: [file(1)], total: 40, page: 1 });
  render(<FileResourcesManagementPage api={{ assets: { files } } as unknown as GooseAdminApi} text={key => key} />);
  await screen.findByRole("button", { name: "file-1" });
  act(() => onIntersect([{ isIntersecting: true } as IntersectionObserverEntry], {} as IntersectionObserver));
  await screen.findByText("network failed");
  expect(screen.getByRole("button", { name: "file-1" })).toBeTruthy();
  fireEvent.click(screen.getByRole("button", { name: "retry" }));
  await screen.findByRole("button", { name: "file-2" });
  expect(screen.getByRole("button", { name: "file-1" })).toBeTruthy();
  fireEvent.click(screen.getByRole("button", { name: "listView" }));
  await screen.findByRole("table");
  await waitFor(() => expect(files).toHaveBeenLastCalledWith({ page: 1, pageSize: 20 }));
  expect(screen.queryByRole("button", { name: "file-2" })).toBeNull();
});

it("falls back when the preview image cannot be loaded", async () => {
  vi.stubGlobal("ResizeObserver", class { observe() {} disconnect() {} });
  const file = { id: 1, name: "missing.png", type: "image/png", url: "/files/1", size: 100, userId: 1, createdAt: "" };
  const files = vi.fn().mockResolvedValue({ list: [file], total: 1, page: 1 });
  render(<FileResourcesManagementPage api={{ assets: { files } } as unknown as GooseAdminApi} text={key => key} />);
  fireEvent.click(await screen.findByRole("button", { name: "missing.png" }));
  const image = await screen.findByRole("img", { name: "missing.png" });
  fireEvent.error(image);
  expect(screen.queryByRole("img", { name: "missing.png" })).toBeNull();
  expect(screen.getAllByText("missing.png").length).toBeGreaterThan(1);
});

it("falls back when a list thumbnail cannot be loaded", async () => {
  const file = { id: 1, name: "missing.png", type: "image/png", url: "/files/1", size: 100, userId: 1, createdAt: "" };
  const files = vi.fn().mockResolvedValue({ list: [file], total: 1, page: 1 });
  vi.stubGlobal("ResizeObserver", class { observe() {} disconnect() {} });
  render(<FileResourcesManagementPage api={{ assets: { files } } as unknown as GooseAdminApi} text={key => key} />);
  await screen.findByRole("button", { name: "missing.png" });
  fireEvent.click(screen.getByRole("button", { name: "listView" }));
  await screen.findByRole("table");
  const image = document.querySelector("tbody img");
  expect(image).toBeTruthy();
  fireEvent.error(image!);
  expect(document.querySelector("tbody img")).toBeNull();
});

it("ignores a stale response after switching view", async () => {
  vi.stubGlobal("ResizeObserver", class { observe() {} disconnect() {} });
  const file = (id: number) => ({ id, name: `file-${id}`, type: "application/pdf", url: `/files/${id}`, size: 100, userId: 1, createdAt: "" });
  let resolveFirst!: (value: { list: ReturnType<typeof file>[]; total: number; page: number }) => void;
  let resolveSecond!: (value: { list: ReturnType<typeof file>[]; total: number; page: number }) => void;
  const files = vi.fn()
    .mockImplementationOnce(() => new Promise(resolve => { resolveFirst = resolve; }))
    .mockImplementationOnce(() => new Promise(resolve => { resolveSecond = resolve; }));
  render(<FileResourcesManagementPage api={{ assets: { files } } as unknown as GooseAdminApi} text={key => key} />);
  fireEvent.click(await screen.findByRole("button", { name: "listView" }));
  resolveFirst({ list: [file(1)], total: 1, page: 1 });
  await act(async () => {});
  expect(screen.queryByText("file-1")).toBeNull();
  resolveSecond({ list: [file(2)], total: 1, page: 1 });
  expect(await screen.findByText("file-2")).toBeTruthy();
  expect(screen.queryByText("file-1")).toBeNull();
});
