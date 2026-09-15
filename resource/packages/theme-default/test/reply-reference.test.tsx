import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ReplyReference } from "../src/site/topics/reply-reference";
import { GooseRuntimeProvider, type GooseRuntime } from "@gooseforum/runtime";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});
const target = {
  id: 2,
  postNo: 2,
  author: { id: 1, username: "alice", avatarUrl: "" },
  renderedContent:
    '<p>Quoted <strong>body</strong> <a href="https://example.com">source</a></p>',
};
const t = (key: string) => key;
function view(value = target) {
  return (
    <GooseRuntimeProvider
      runtime={{ navigate: vi.fn() } as unknown as GooseRuntime}
    >
      <ReplyReference topicId={60} target={value} t={t} />
    </GooseRuntimeProvider>
  );
}
describe("reply reference", () => {
  it("uses body typography and a proper source permalink without nesting content links", () => {
    const { container } = render(view());
    expect(container.querySelector(".typeset-forum strong")?.textContent).toBe(
      "body",
    );
    expect(screen.getByRole("link", { name: "#2" }).getAttribute("href")).toBe(
      "/p/post/60/2#post-2",
    );
    expect(container.querySelector("a a")).toBeNull();
    expect(screen.queryByRole("button")).toBeNull();
  });
  it("expands overflowing quotes, collapses them, resets for changed content and cleans up its observer", async () => {
    vi.spyOn(HTMLElement.prototype, "scrollHeight", "get").mockReturnValue(240);
    const disconnect = vi.fn();
    vi.stubGlobal(
      "ResizeObserver",
      class {
        observe() {}
        disconnect = disconnect;
      },
    );
    const user = userEvent.setup();
    const { container, rerender, unmount } = render(view());
    const body = () => container.querySelector(".typeset-forum")!;
    expect(body().classList.contains("overflow-hidden")).toBe(true);
    await user.click(screen.getByRole("button", { name: "expandReply" }));
    expect(screen.getByRole("button").getAttribute("aria-expanded")).toBe(
      "true",
    );
    expect(body().classList.contains("overflow-hidden")).toBe(false);
    await user.click(screen.getByRole("button", { name: "collapseReply" }));
    expect(body().classList.contains("overflow-hidden")).toBe(true);
    await user.click(screen.getByRole("button", { name: "expandReply" }));
    rerender(view({ ...target, renderedContent: "<p>Changed quote</p>" }));
    expect(screen.getByRole("button").getAttribute("aria-expanded")).toBe(
      "false",
    );
    unmount();
    expect(disconnect).toHaveBeenCalledTimes(2);
  });
  it("handles unavailable targets without showing their content", () => {
    render(
      <GooseRuntimeProvider
        runtime={{ navigate: vi.fn() } as unknown as GooseRuntime}
      >
        <ReplyReference
          topicId={60}
          target={{ ...target, unavailable: true }}
          t={t}
        />
      </GooseRuntimeProvider>,
    );
    expect(screen.getByText("replyTargetUnavailable")).toBeTruthy();
    expect(screen.queryByText("body")).toBeNull();
  });
});
