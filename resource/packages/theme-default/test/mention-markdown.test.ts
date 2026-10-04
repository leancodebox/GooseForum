import { describe, expect, it } from "vitest";
import { hasUnsupportedVisualMarkdown, renderMarkdown, parseExtensionHeader } from "@gooseforum/markdown";
import { parseVisualMarkdown, serializeVisualMarkdown } from "../src/site/editor/visual-markdown-schema";
import fixtures from "../../../../testdata/markdown-extensions/mentions.json";

describe("forum mention syntax", () => {
  for(const fixture of fixtures)it(`shared fixture: ${fixture.name}`,()=>{
    const html=renderMarkdown(fixture.source,fixture.version as 0|1);
    expect(html.match(/data-mention-user=/g)?.length??0).toBe(fixture.mentions);
    for(const unexpected of fixture.notContains??[])expect(html).not.toContain(unexpected);
  });
  const mention='[mention user="9007199254740993"]@alice[/mention]';
  it("keeps stable IDs through visual editing", () => {
    const doc=parseVisualMarkdown(`你好 **${mention}**。`);
    expect(serializeVisualMarkdown(doc)).toContain(mention);
    expect(renderMarkdown(mention)).toContain('href="/u/9007199254740993"');
  });
  it("does not interpret legacy source or protected Markdown contexts", () => {
    for(const source of [mention,`\`${mention}\``,`\\${mention}`,`[${mention}](/url)`,`<span>${mention}</span>`]) {
      const version=source===mention?0:1;
      expect(renderMarkdown(source,version)).not.toContain('data-mention-user=');
    }
  });
  it("keeps unsupported tags in source mode", () => {
    expect(hasUnsupportedVisualMarkdown('[future key="123"]unknown[/future]')).toBe(true);
    expect(hasUnsupportedVisualMarkdown('[mention user="123"]@alice[/mention]')).toBe(false);
    expect(hasUnsupportedVisualMarkdown('[mention user="123"]@alice[/mention]',0)).toBe(true);
    expect(hasUnsupportedVisualMarkdown('[mention user="bad"]@alice[/mention]')).toBe(true);
  });
  it("validates the common attribute grammar",()=>{
    expect(parseExtensionHeader('[note text="a\\"b\\]c"]')?.attributes.text).toBe('a"b]c');
    expect(parseExtensionHeader('[note text="a" text="b"]')).toBeNull();
    expect(renderMarkdown('[mention user="18446744073709551616"]@alice[/mention]')).not.toContain('data-mention-user=');
  });
});
