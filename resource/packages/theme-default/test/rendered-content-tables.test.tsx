import { StrictMode } from "react";
import { afterEach, expect, it } from "vitest";
import { cleanup, render } from "@testing-library/react";
import { I18nextProvider } from "react-i18next";
import { createGooseI18n } from "@gooseforum/runtime/i18n";
import { RenderedContent } from "../src/site/content/rendered-content";

afterEach(cleanup);

it("keeps one table wrapper through strict effects and content replacement", () => {
  const i18n = createGooseI18n("en");
  const content = (html: string) => <StrictMode><I18nextProvider i18n={i18n}><RenderedContent html={html} /></I18nextProvider></StrictMode>;
  const view = render(content('<table><thead><tr><th>Number</th><th>Description</th></tr></thead><tbody><tr><td>10</td><td>A longer description</td></tr></tbody></table>'));
  expect(view.container.querySelectorAll('.gf-table-shell')).toHaveLength(1);
  expect(view.container.querySelector('tbody td')?.classList.contains('gf-table-compact')).toBe(true);
  view.rerender(content('<table><tbody><tr><td>Replacement content</td></tr></tbody></table>'));
  expect(view.container.querySelectorAll('.gf-table-shell')).toHaveLength(1);
  expect(view.container.querySelectorAll('table')).toHaveLength(1);
  expect(view.container.textContent).toBe('Replacement content');
  view.rerender(content('<table><tbody><tr><td colspan="2">Group</td></tr><tr><td>1</td><td>2</td></tr></tbody></table>'));
  expect(view.container.querySelector('.gf-table-compact')).toBeNull();
  view.rerender(content('<p>No table remains</p>'));
  expect(view.container.querySelector('.gf-table-shell')).toBeNull();
  expect(view.container.textContent).toBe('No table remains');
});
