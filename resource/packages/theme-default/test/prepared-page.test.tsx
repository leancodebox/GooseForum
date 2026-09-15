import { Suspense } from "react";
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { preparedPage, prepareGoosePage } from "@gooseforum/runtime/prepared-page";

afterEach(cleanup);

it("renders a prepared page synchronously without committing a fallback", async () => {
  const loader = vi.fn(async () => ({
    default: ({ title }: { title: string }) => <h1>{title}</h1>,
  }));
  const Page = preparedPage("test.prepared", loader);
  await Promise.all([
    prepareGoosePage("test.prepared"),
    prepareGoosePage("test.prepared"),
  ]);
  render(
    <Suspense fallback={<p>Loading page</p>}>
      <Page title="Ready page" />
    </Suspense>,
  );
  expect(screen.getByRole("heading", { name: "Ready page" })).toBeTruthy();
  expect(screen.queryByText("Loading page")).toBeNull();
  expect(loader).toHaveBeenCalledTimes(1);
});

it("allows retrying a failed module preparation", async () => {
  const loader = vi
    .fn()
    .mockRejectedValueOnce(new Error("Temporary load failure"))
    .mockResolvedValue({ default: () => <h1>Recovered page</h1> });
  const Page = preparedPage("test.retry", loader);
  await expect(prepareGoosePage("test.retry")).rejects.toThrow(
    "Temporary load failure",
  );
  await prepareGoosePage("test.retry");
  render(<Page />);
  expect(screen.getByRole("heading", { name: "Recovered page" })).toBeTruthy();
});
