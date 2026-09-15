import { afterEach, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { ColorPicker } from "@gooseforum/ui/components/color-picker";
import { colorPalette, hexChannels, hslHex, rgbHex } from "@gooseforum/ui/lib/color";

afterEach(() => {
  cleanup();
  localStorage.clear();
});
it("converts hex and HSL channels and builds a varied valid palette", () => {
  expect(hexChannels("#abc")).toEqual([170, 187, 204]);
  expect(hexChannels("#nope")).toBeNull();
  expect(rgbHex([-1, 128.4, 300])).toBe("#0080ff");
  expect(hslHex(0, 100, 50)).toBe("#ff0000");
  expect(hslHex(120, 100, 50)).toBe("#00ff00");
  expect(hslHex(240, 100, 50)).toBe("#0000ff");
  expect(new Set(colorPalette.flat()).size).toBe(130);
  expect(colorPalette.flat().every((color) => !!hexChannels(color))).toBe(true);
});
it("previews choices, validates input, restores the original color and remembers selections without submitting the parent form", async () => {
  const submit = vi.fn((event) => event.preventDefault());
  function Fixture() {
    const [value, setValue] = useState("#123456");
    return (
      <form onSubmit={submit}>
        <ColorPicker
          label="Category color"
          locale="en"
          value={value}
          onChange={setValue}
        />
        <output>{value}</output>
      </form>
    );
  }
  const user = userEvent.setup();
  render(<Fixture />);
  await user.click(screen.getByRole("button", { name: "Category color" }));
  await user.click(screen.getByRole("button", { name: "#000000" }));
  expect(screen.getByRole("status", { hidden: true }).textContent).toBe(
    "#000000",
  );
  await user.click(screen.getByRole("tab", { name: "RGB channels" }));
  fireEvent.change(screen.getByRole("slider", { name: "R" }), {
    target: { value: "255" },
  });
  expect(screen.getByRole("status", { hidden: true }).textContent).toBe(
    "#ff0000",
  );
  const input = screen.getByLabelText("Color value");
  await user.clear(input);
  await user.type(input, "invalid{Enter}");
  expect(screen.getByRole("alert")).toBeTruthy();
  expect(screen.getByRole("status", { hidden: true }).textContent).toBe(
    "#ff0000",
  );
  await user.click(
    screen.getByRole("button", { name: "Restore original color" }),
  );
  expect(screen.getByRole("status", { hidden: true }).textContent).toBe(
    "#123456",
  );
  await user.clear(input);
  await user.type(input, "#abc{Enter}");
  expect(screen.getByRole("status", { hidden: true }).textContent).toBe(
    "#aabbcc",
  );
  await user.click(screen.getByRole("button", { name: "Done" }));
  expect(screen.queryByLabelText("Color value")).toBeNull();
  expect(submit).not.toHaveBeenCalled();
  await user.click(screen.getByRole("button", { name: "Category color" }));
  expect(
    screen.getByRole("button", { name: "Recent colors #aabbcc" }),
  ).toBeTruthy();
});
