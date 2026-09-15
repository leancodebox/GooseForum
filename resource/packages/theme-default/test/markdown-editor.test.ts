import { describe, expect, it } from "vitest";
import { insertBlock, insertInline, renderMarkdown } from "@gooseforum/markdown";
import { parseEditableVisualMarkdown, serializeVisualMarkdown } from "../src/site/editor/visual-markdown-schema";

describe("React Markdown editor primitives", () => {
  it("round-trips visual Markdown marks, lists, code, and tables", () => {
    const source = [
      "# Heading",
      "",
      "A **bold** and ~~removed~~ line.",
      "",
      "- one",
      "- two",
      "",
      "```ts",
      "const value = 1",
      "```",
      "",
      "| Name | Value |",
      "| --- | --- |",
      "| Goose | Forum |",
    ].join("\n");
    const serialized = serializeVisualMarkdown(parseEditableVisualMarkdown(source));
    expect(serialized).toContain("# Heading");
    expect(serialized).toContain("**bold**");
    expect(serialized).toContain("~~removed~~");
    expect(serialized).toContain("const value = 1");
    expect(serialized).toContain("| Name | Value |");
  });

  it("preserves selection positions for inline and block edits", () => {
    expect(insertInline("hello", 0, 5, "**", "**", "text")).toEqual({
      value: "**hello**",
      start: 2,
      end: 7,
    });
    expect(insertBlock("before", 6, 6, "after").value).toBe(
      "before\n\nafter",
    );
  });

  it("renders safe Markdown without enabling raw HTML", () => {
    const html = renderMarkdown("<script>alert(1)</script>\n\n**safe**");
    expect(html).not.toContain("<script>");
    expect(html).toContain("&lt;script&gt;");
    expect(html).toContain("<strong>safe</strong>");
  });
});
