import { describe, expect, it } from "vitest";
import { hasUnsupportedVisualMarkdown, renderMarkdown } from "@gooseforum/markdown";
import { parseEditableVisualMarkdown, serializeVisualMarkdown } from "../src/site/editor/visual-markdown-schema";

describe("image galleries", () => {
  it("starts an empty editor with one paragraph rather than an empty gallery", () => {
    const doc = parseEditableVisualMarkdown("");
    expect(doc.childCount).toBe(1);
    expect(doc.firstChild?.type.name).toBe("paragraph");
    expect(doc.firstChild?.content.size).toBe(0);
    expect(serializeVisualMarkdown(doc)).toBe("");
    expect(doc.type.createAndFill()?.firstChild?.type.name).toBe("paragraph");
  });
  for (const mode of ["grid", "carousel"]) it(`round-trips ${mode} with image metadata`, () => {
    const source = `[gallery mode="${mode}"]\n![First](/one.jpg "Caption")\n![Second](/two.jpg)\n[/gallery]`;
    expect(hasUnsupportedVisualMarkdown(source)).toBe(false);
    const serialized = serializeVisualMarkdown(parseEditableVisualMarkdown(source));
    const html = renderMarkdown(serialized);
    expect(html).toContain(`data-gallery="${mode}"`);
    expect(html.match(/<img /g)).toHaveLength(2);
    expect(html).toContain('title="Caption"');
    expect(serializeVisualMarkdown(parseEditableVisualMarkdown(serialized))).toBe(serialized);
    expect(renderMarkdown(source, 0)).not.toContain("data-gallery");
  });
  it("keeps inline formatting in gallery items", () => {
    const source = '[gallery mode="grid"]\n![One](/one.jpg) [mention user="12"]@alice[/mention]\n[/gallery]';
    expect(renderMarkdown(serializeVisualMarkdown(parseEditableVisualMarkdown(source)))).toContain('data-mention-user="12"');
  });
  it("separates merged gallery images without losing captions or URLs", () => {
    const source = '[gallery mode="carousel"]\n![First](/one.jpg) caption ![Second](/two.jpg)![Third](/three.jpg)\n[/gallery]';
    const doc = parseEditableVisualMarkdown(source);
    const group = doc.child(1);
    expect(group.type.name).toBe("gallery");
    expect(group.childCount).toBe(3);
    const serialized = serializeVisualMarkdown(doc);
    const lines = serialized.split("\n").filter(line => line.includes("!["));
    expect(lines).toHaveLength(3);
    expect(lines.every(line => line.match(/!\[/g)?.length === 1)).toBe(true);
    expect(serialized).toContain("caption");
    expect(serializeVisualMarkdown(parseEditableVisualMarkdown(serialized))).toBe(serialized);
  });
  it("preserves malformed, code and unknown extensions literally", () => {
    for (const source of [
      '[gallery mode="invalid"]\n![x](/x.png)\n[/gallery]',
      '```\n[gallery mode="grid"]\n![x](/x.png)\n[/gallery]\n```',
      '[future]\n[gallery mode="grid"]\n![x](/x.png)\n[/gallery]\n[/future]',
    ]) expect(renderMarkdown(source)).not.toContain("data-gallery");
  });
});
