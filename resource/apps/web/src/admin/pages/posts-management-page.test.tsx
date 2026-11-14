import { afterEach, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { AdminTopic, GooseAdminApi } from "@gooseforum/client";
import messages from "../messages/zh-posts";
import { PostsManagementPage } from "./posts-management-page";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

function topic(id: number, title: string): AdminTopic {
  return {
    id,
    title,
    moderationStatus: "pending",
    moderationReason: "",
    moderationVersion: 0,
    categoryId: [],
    userId: 7,
    username: "alice",
    userAvatarUrl: "",
    topicStatus: 1,
    processStatus: 0,
    deleted: false,
    viewCount: 0,
    replyCount: 0,
    likeCount: 0,
    pinWeight: 0,
    createdAt: "2026-11-14T00:00:00Z",
  };
}

it("reviews selected topics through the batch contract and refreshes the page", async () => {
  window.history.replaceState({}, "", "/admin/posts");
  const list = vi.fn().mockResolvedValue({ list: [topic(1, "主题一"), topic(2, "主题二")], total: 2, page: 1, pageSize: 10, hasNext: false });
  const reviewBatch = vi.fn().mockResolvedValue({ items: [{ id: 1, success: true }] });
  const api = { topics: { list, reviewBatch }, categories: { list: vi.fn().mockResolvedValue([]) } } as unknown as GooseAdminApi;
  const user = userEvent.setup();

  render(<PostsManagementPage api={api} text={(key) => messages[key]} initialData={{
    query: { kind: "topic", page: 1, pageSize: 10, search: "", moderationStatus: "all", categoryId: "all" },
    result: { list: [topic(1, "主题一"), topic(2, "主题二")], total: 2, page: 1, pageSize: 10, hasNext: false },
    categories: [],
  }} />);

  await user.click(screen.getAllByRole("checkbox", { name: "选择 主题一" })[0]);
  await user.click(screen.getByRole("button", { name: /^通过$/ }));
  const dialog = screen.getByRole("dialog");
  await user.type(within(dialog).getByPlaceholderText(messages.reviewReason), "已确认");
  await user.click(within(dialog).getByRole("button", { name: /^通过$/ }));

  await waitFor(() => expect(reviewBatch).toHaveBeenCalledWith("topic", {
    items: [{ id: 1, version: 0 }], action: "approve", reason: "已确认",
  }));
  await waitFor(() => expect(list).toHaveBeenCalled());
});

it("keeps failed batch items selected for an immediate retry", async () => {
  window.history.replaceState({}, "", "/admin/posts");
  const list = vi.fn().mockResolvedValue({ list: [topic(1, "主题一"), topic(2, "主题二")], total: 2, page: 1, pageSize: 10, hasNext: false });
  const reviewBatch = vi.fn().mockResolvedValue({ items: [{ id: 1, success: false }, { id: 2, success: true }] });
  const api = { topics: { list, reviewBatch }, categories: { list: vi.fn().mockResolvedValue([]) } } as unknown as GooseAdminApi;
  const user = userEvent.setup();

  render(<PostsManagementPage api={api} text={(key) => messages[key]} initialData={{
    query: { kind: "topic", page: 1, pageSize: 10, search: "", moderationStatus: "all", categoryId: "all" },
    result: { list: [topic(1, "主题一"), topic(2, "主题二")], total: 2, page: 1, pageSize: 10, hasNext: false },
    categories: [],
  }} />);

  await user.click(screen.getAllByRole("checkbox", { name: "选择 主题一" })[0]);
  await user.click(screen.getAllByRole("checkbox", { name: "选择 主题二" })[0]);
  await user.click(screen.getByRole("button", { name: /^通过$/ }));
  const dialog = screen.getByRole("dialog");
  await user.type(within(dialog).getByPlaceholderText(messages.reviewReason), "再次确认");
  await user.click(within(dialog).getByRole("button", { name: /^通过$/ }));

  await waitFor(() => expect(reviewBatch).toHaveBeenCalled());
  await waitFor(() => expect(screen.getByText(`${messages.selected} 1 · ${messages.selectedPageOnly}`)).toBeTruthy());
  expect(screen.getAllByRole("checkbox", { name: "选择 主题一" }).every((checkbox) => (checkbox as HTMLButtonElement).getAttribute("aria-checked") === "true")).toBe(true);
  expect(screen.getAllByRole("checkbox", { name: "选择 主题二" }).every((checkbox) => (checkbox as HTMLButtonElement).getAttribute("aria-checked") !== "true")).toBe(true);
});

it("selects the whole topic page from the mobile toolbar", async () => {
  window.history.replaceState({}, "", "/admin/posts");
  const api = { topics: { list: vi.fn(), reviewBatch: vi.fn() }, categories: { list: vi.fn().mockResolvedValue([]) } } as unknown as GooseAdminApi;
  const user = userEvent.setup();

  render(<PostsManagementPage api={api} text={(key) => messages[key]} initialData={{
    query: { kind: "topic", page: 1, pageSize: 10, search: "", moderationStatus: "all", categoryId: "all" },
    result: { list: [topic(1, "主题一"), topic(2, "主题二")], total: 2, page: 1, pageSize: 10, hasNext: false },
    categories: [],
  }} />);

  await user.click(screen.getAllByRole("checkbox", { name: messages.selectAll })[1]);

  expect(screen.getByText(`${messages.selected} 2 · ${messages.selectedPageOnly}`)).toBeTruthy();
});

it("clears selected topics when the list is refreshed", async () => {
  window.history.replaceState({}, "", "/admin/posts");
  const list = vi.fn().mockResolvedValue({ list: [topic(1, "主题一")], total: 1, page: 1, pageSize: 10, hasNext: false });
  const api = { topics: { list, reviewBatch: vi.fn() }, categories: { list: vi.fn().mockResolvedValue([]) } } as unknown as GooseAdminApi;
  const user = userEvent.setup();

  render(<PostsManagementPage api={api} text={(key) => messages[key]} initialData={{
    query: { kind: "topic", page: 1, pageSize: 10, search: "", moderationStatus: "all", categoryId: "all" },
    result: { list: [topic(1, "主题一")], total: 1, page: 1, pageSize: 10, hasNext: false },
    categories: [],
  }} />);

  await user.click(screen.getAllByRole("checkbox", { name: "选择 主题一" })[0]);
  expect(screen.getByText(`${messages.selected} 1 · ${messages.selectedPageOnly}`)).toBeTruthy();
  await user.click(screen.getByRole("button", { name: messages.refresh }));
  await waitFor(() => expect(list).toHaveBeenCalled());
  expect(screen.queryByText(`${messages.selected} 1 · ${messages.selectedPageOnly}`)).toBeNull();
});
