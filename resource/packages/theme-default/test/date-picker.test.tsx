import { useState, type FormEvent } from "react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DatePicker } from "@gooseforum/ui/components/date-picker";

beforeEach(() => { vi.useFakeTimers({ toFake: ["Date"] }); vi.setSystemTime(new Date(2026, 9, 4)); });
afterEach(() => { cleanup(); vi.useRealTimers(); });

function Harness({ initial = "", withTime = false, onChange = vi.fn(), locale = "en" }: { initial?: string; withTime?: boolean; onChange?(value: string): void; locale?: string }) {
  const [value, setValue] = useState(initial);
  const timeProps = withTime ? { withTime: true as const, timeLabel: "Time", doneLabel: "Done" } : {};
  return <><label htmlFor="date">Date</label><DatePicker id="date" value={value} locale={locale} {...timeProps} placeholder="Choose date" clearLabel="Clear date" onChange={(next) => { setValue(next); onChange(next); }} /></>;
}

it("selects and clears a local date without a UTC date conversion", async () => {
  const change = vi.fn();
  render(<Harness onChange={change} />);
  await userEvent.click(screen.getByLabelText("Date"));
  expect(screen.queryByRole("combobox")).toBeNull();
  await userEvent.click(document.querySelector<HTMLButtonElement>('[data-day="10/6/2026"]')!);
  expect(change).toHaveBeenLastCalledWith("2026-10-06");
  expect(document.querySelector('[data-slot="date-picker-content"]')).toBeNull();
  expect(screen.getByLabelText("Date").textContent).toContain("Oct 6, 2026");
  await userEvent.click(screen.getByLabelText("Date"));
  await userEvent.click(screen.getByRole("button", { name: "Clear date" }));
  expect(change).toHaveBeenLastCalledWith("");
  expect(screen.getByLabelText("Date").textContent).toContain("Choose date");
});

it("keeps the chosen time when changing dates and never submits an outer form", async () => {
  const submit = vi.fn((event: FormEvent<HTMLFormElement>) => event.preventDefault()), change = vi.fn();
  render(<form onSubmit={submit}><Harness initial="2026-10-03T11:23" withTime onChange={change} /></form>);
  await userEvent.click(screen.getByLabelText("Date"));
  expect(screen.getByLabelText<HTMLInputElement>("Time").value).toBe("11:23");
  await userEvent.click(document.querySelector<HTMLButtonElement>('[data-day="10/6/2026"]')!);
  expect(change).toHaveBeenLastCalledWith("2026-10-06T11:23");
  fireEvent.change(screen.getByLabelText("Time"), { target: { value: "10:00" } });
  expect(change).toHaveBeenLastCalledWith("2026-10-06T10:00");
  expect([...document.querySelectorAll<HTMLButtonElement>('[data-slot="date-picker-content"] button')].every((button) => button.type === "button")).toBe(true);
  await userEvent.click(screen.getByRole("button", { name: "Done" }));
  expect(document.querySelector('[data-slot="date-picker-content"]')).toBeNull();
  expect(submit).not.toHaveBeenCalled();
});

it("does not accept rolled-over calendar dates", async () => {
  render(<Harness initial="2026-02-30" />);
  expect(screen.getByLabelText("Date").textContent).toContain("Choose date");
  await userEvent.click(screen.getByLabelText("Date"));
  expect(document.querySelector('[data-selected-single="true"]')).toBeNull();
});

it("requires a selected date before setting a time and clears back to no deadline", async () => {
  const change = vi.fn();
  render(<Harness withTime onChange={change} />);
  await userEvent.click(screen.getByLabelText("Date"));
  expect(screen.getByLabelText<HTMLInputElement>("Time").disabled).toBe(true);
  await userEvent.click(document.querySelector<HTMLButtonElement>('[data-day="10/6/2026"]')!);
  expect(change).toHaveBeenLastCalledWith("2026-10-06T00:00");
  expect(screen.getByLabelText<HTMLInputElement>("Time").disabled).toBe(false);
  await userEvent.click(screen.getByRole("button", { name: "Clear date" }));
  expect(change).toHaveBeenLastCalledWith("");
  expect(screen.getByLabelText("Date").textContent).toContain("Choose date");
});

it.each([['zh', '2026/10/6'], ['ja', '2026/10/6'], ['it', '06/10/2026']])("uses the %s calendar locale", async (locale, day) => {
  const change = vi.fn();
  render(<Harness locale={locale} onChange={change} />);
  await userEvent.click(screen.getByLabelText("Date"));
  await userEvent.click(document.querySelector<HTMLButtonElement>(`[data-day="${day}"]`)!);
  expect(change).toHaveBeenLastCalledWith("2026-10-06");
});
