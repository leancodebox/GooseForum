import { afterEach, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { GooseAdminApi } from "@gooseforum/client";
import { RegistrationInvitationsPage } from "./registration-invitations-page";

afterEach(() => cleanup());

it("loads invitations and shows a newly created code only once", async () => {
  const registrationInvitations = vi.fn().mockResolvedValue({ list: [], total: 0, page: 1, size: 20 });
  const createRegistrationInvitation = vi.fn().mockResolvedValue({ code: "gf_inv_secret", entry: { id: 1, prefix: "gf_inv_secret".slice(0, 16), maxUses: 1, usedCount: 0, createdBy: 1, createdAt: "", updatedAt: "" } });
  const api = { users: { registrationInvitations, createRegistrationInvitation } } as unknown as GooseAdminApi;
  render(<RegistrationInvitationsPage api={api} text={key => key} />);
  await screen.findByText("empty");
  fireEvent.click(screen.getByRole("button", { name: "newInvitation" }));
  fireEvent.click(screen.getByRole("button", { name: "createInvitation" }));
  await screen.findByDisplayValue("gf_inv_secret");
  expect(createRegistrationInvitation).toHaveBeenCalledWith({ maxUses: 1, expiresAt: null });
  await waitFor(() => expect(registrationInvitations).toHaveBeenCalledTimes(2));
  expect(registrationInvitations).toHaveBeenLastCalledWith({ page: 1, pageSize: 20 });
});

it("copies a registration link for the newly created code", async () => {
  const writeText = vi.fn().mockResolvedValue(undefined);
  Object.assign(navigator, { clipboard: { writeText } });
  const registrationInvitations = vi.fn().mockResolvedValue({ list: [], total: 0, page: 1, size: 20 });
  const createRegistrationInvitation = vi.fn().mockResolvedValue({ code: "gf_inv_a/b", entry: { id: 1, prefix: "gf_inv_a/b", maxUses: 1, usedCount: 0, createdBy: 1, createdAt: "", updatedAt: "" } });
  const api = { users: { registrationInvitations, createRegistrationInvitation } } as unknown as GooseAdminApi;
  render(<RegistrationInvitationsPage api={api} text={key => key} />);
  await screen.findByText("empty");
  fireEvent.click(screen.getByRole("button", { name: "newInvitation" }));
  fireEvent.click(screen.getByRole("button", { name: "createInvitation" }));
  await screen.findByDisplayValue("gf_inv_a/b");
  fireEvent.click(screen.getByRole("button", { name: "copyInvitationLink" }));
  expect(writeText).toHaveBeenCalledWith(`${window.location.origin}/login?register=true&invitation=gf_inv_a%2Fb`);
});

it("passes an optional recipient email when creating an invitation", async () => {
  const registrationInvitations = vi.fn().mockResolvedValue({ list: [], total: 0, page: 1, size: 20 });
  const createRegistrationInvitation = vi.fn().mockResolvedValue({ code: "gf_inv_secret", entry: { id: 1, prefix: "gf_inv_secret", maxUses: 1, usedCount: 0, createdBy: 1, createdAt: "", updatedAt: "" }, emailQueued: true });
  const api = { users: { registrationInvitations, createRegistrationInvitation } } as unknown as GooseAdminApi;
  render(<RegistrationInvitationsPage api={api} text={key => key} />);
  await screen.findByText("empty");
  fireEvent.click(screen.getByRole("button", { name: "newInvitation" }));
  fireEvent.change(screen.getByLabelText("invitationEmail"), { target: { value: "invitee@example.com" } });
  fireEvent.click(screen.getByRole("button", { name: "createInvitation" }));
  await screen.findByDisplayValue("gf_inv_secret");
  expect(createRegistrationInvitation).toHaveBeenCalledWith({ maxUses: 1, expiresAt: null, email: "invitee@example.com" });
});

it("locks pagination while an invitation is being saved", async () => {
  let resolveCreate!: (value: { code: string; entry: { id: number; prefix: string; maxUses: number; usedCount: number; createdBy: number; createdAt: string; updatedAt: string } }) => void;
  const createRegistrationInvitation = vi.fn().mockImplementation(() => new Promise((resolve) => { resolveCreate = resolve; }));
  const registrationInvitations = vi.fn().mockResolvedValue({ list: [], total: 21, page: 1, size: 20 });
  const api = { users: { registrationInvitations, createRegistrationInvitation } } as unknown as GooseAdminApi;
  render(<RegistrationInvitationsPage api={api} text={key => key} />);
  await screen.findByText("empty");
  fireEvent.click(screen.getByRole("button", { name: "newInvitation" }));
  fireEvent.click(screen.getByRole("button", { name: "createInvitation" }));

  expect(screen.getByRole("button", { name: "next", hidden: true })).toHaveProperty("disabled", true);
  resolveCreate({ code: "gf_inv_secret", entry: { id: 1, prefix: "gf_inv_secret", maxUses: 1, usedCount: 0, createdBy: 1, createdAt: "", updatedAt: "" } });
  await screen.findByDisplayValue("gf_inv_secret");
});
