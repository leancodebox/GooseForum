import { expect, test } from "@playwright/test";

test("renders the standalone payload and handles theme navigation", async ({
  page,
  request,
}, testInfo) => {
  const response = await request.get("/?lang=en");
  expect(response.ok()).toBe(true);
  const html = await response.text();
  expect(html).toContain("Standalone Next host");
  expect(html).toContain("No topics yet");

  const errors: string[] = [];
  const documentRequests: string[] = [];
  const pageDataRequests: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  page.on("console", (message) => {
    if (message.type() === "error") errors.push(message.text());
  });
  page.on("request", (request) => {
    if (request.resourceType() === "document")
      documentRequests.push(request.url());
    if (request.url().includes("/goose-page-data"))
      pageDataRequests.push(request.url());
  });
  await page.goto("/?lang=en");
  await expect(page).toHaveTitle("GooseForum Next");
  await expect(page.getByText("Standalone Next host")).toBeVisible();
  await expect(page.getByText("No topics yet")).toBeVisible();

  // A client-only interaction proves hydration is complete before SPA navigation.
  await page.getByRole("button", { name: "Switch to dark theme" }).click();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "gf-dark");

  await page.getByRole("link", { name: "Hot" }).click();
  await expect(page).toHaveURL(/\?sort=hot/);
  await expect(page.getByRole("link", { name: "Hot" })).toHaveAttribute(
    "data-variant",
    "default",
  );
  expect(documentRequests).toHaveLength(1);
  expect(pageDataRequests).toHaveLength(1);
  expect(pageDataRequests[0]).toContain("%3Fsort%3Dhot");

  expect(errors).toEqual([]);
  await testInfo.attach("next-standalone", {
    body: await page.screenshot({ fullPage: false }),
    contentType: "image/png",
  });
});
