import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";
import type { GooseAdminApi } from "@gooseforum/client";
import { UsersManagementPage } from "./pages/users-management-page";

afterEach(cleanup);

it("does not overwrite a newer user search with an older response", async () => {
  let finishOld!: (value: unknown) => void;
  const record = (username: string) => ({
    userId: 1,
    username,
    roleList: [],
    status: 0,
  });
  const list = vi
    .fn()
    .mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          finishOld = resolve;
        }),
    )
    .mockResolvedValue({ list: [record("new-user")], total: 1 });
  const api = { users: { list } } as unknown as GooseAdminApi;
  const user = userEvent.setup();
  render(<UsersManagementPage api={api} text={(key) => key} />);
  await user.type(screen.getByPlaceholderText("search"), "new");
  await user.click(screen.getByRole("button", { name: "searchAction" }));
  await waitFor(() => expect(list).toHaveBeenCalledTimes(2));
  await screen.findAllByText("new-user");
  await act(async () => {
    finishOld({ list: [record("old-user")], total: 1 });
  });
  expect(screen.queryByText("old-user")).toBeNull();
  expect(screen.getAllByText("new-user").length).toBeGreaterThan(0);
});
