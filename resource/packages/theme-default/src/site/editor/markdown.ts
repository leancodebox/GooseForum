import MarkdownIt from "markdown-it";
import anchor from "markdown-it-anchor";
// The plugin does not publish TypeScript declarations.
// @ts-expect-error markdown-it-task-lists is an untyped MarkdownIt plugin.
import taskLists from "markdown-it-task-lists";
import TurndownService from "turndown";

const renderer = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
})
  .use(anchor, {
    slugify: (value: string) => value.trim().toLowerCase().replace(/\s+/g, "-"),
  })
  .use(taskLists, { enabled: true });
renderer.renderer.rules.s_open = () => "<del>";
renderer.renderer.rules.s_close = () => "</del>";
const turndown = new TurndownService({
  headingStyle: "atx",
  bulletListMarker: "-",
  codeBlockStyle: "fenced",
});
export const renderMarkdown = (source: string) => renderer.render(source || "");
export const markdownFromClipboard = (data: DataTransfer | null) => {
  const html = data?.getData("text/html") || "";
  return html.trim() ? turndown.turndown(html).trim() : "";
};
export const hasUnsupportedVisualMarkdown = (value: string) =>
  /^\s*(?:>\s*)*(?:[-+*]|\d+[.)])\s+\[[ xX]\]\s+/m.test(value);
export function insertInline(
  source: string,
  start: number,
  end: number,
  before: string,
  after: string,
  placeholder: string,
) {
  const selected = source.slice(start, end) || placeholder;
  return {
    value: `${source.slice(0, start)}${before}${selected}${after}${source.slice(end)}`,
    start: start + before.length,
    end: start + before.length + selected.length,
  };
}
export function insertBlock(
  source: string,
  start: number,
  end: number,
  block: string,
) {
  const left = source.slice(0, start);
  const right = source.slice(end);
  const prefix =
    left && !left.endsWith("\n\n") ? (left.endsWith("\n") ? "\n" : "\n\n") : "";
  const suffix =
    right && !right.startsWith("\n\n")
      ? right.startsWith("\n")
        ? "\n"
        : "\n\n"
      : "";
  return {
    value: `${left}${prefix}${block}${suffix}${right}`,
    start: left.length + prefix.length,
    end: left.length + prefix.length + block.length,
  };
}
