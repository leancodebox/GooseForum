import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import { GooseLink, GooseRuntimeProvider, isSiteSpaPath, type GooseRuntime } from "@gooseforum/runtime";

afterEach(cleanup);
describe("independent admin entry", () => {
  it.each(["/admin", "/admin/", "/admin/users?page=2", "http://localhost:3000/admin/settings#mail"])("leaves %s to native navigation", href => {
    const navigate = vi.fn();
    render(<GooseRuntimeProvider runtime={{ navigate } as unknown as GooseRuntime}><GooseLink href={href}>Admin</GooseLink></GooseRuntimeProvider>);
    const link = screen.getByRole("link", { name: "Admin" });
    const event = new MouseEvent("click", { bubbles: true, cancelable: true });
    // Observe whether React claimed the navigation, without navigating jsdom.
    let claimed = false;
    const stopNative = (e: Event) => { claimed = e.defaultPrevented; e.preventDefault(); };
    document.addEventListener("click", stopNative);
    try { link.dispatchEvent(event); } finally { document.removeEventListener("click", stopNative); }
    expect(claimed).toBe(false);
    expect(navigate).not.toHaveBeenCalled();
  });
  it("keeps similarly named site paths inside the site SPA", () => {
    expect(isSiteSpaPath("/administrator")).toBe(true);
    const navigate = vi.fn();
    render(<GooseRuntimeProvider runtime={{ navigate } as unknown as GooseRuntime}><GooseLink href="/search">Search</GooseLink></GooseRuntimeProvider>);
    const event = new MouseEvent("click", { bubbles: true, cancelable: true });
    screen.getByRole("link", { name: "Search" }).dispatchEvent(event);
    expect(event.defaultPrevented).toBe(true);
    expect(navigate).toHaveBeenCalledWith("/search", { replace: undefined });
  });
});
