import MarkdownIt from "markdown-it";
import { forumExtensions } from "@gooseforum/markdown";
import {
  Fragment,
  Schema,
  type Node as ProseMirrorNode,
} from "prosemirror-model";
import {
  defaultMarkdownParser,
  defaultMarkdownSerializer,
  MarkdownParser,
  MarkdownSerializer,
  schema as baseSchema,
} from "prosemirror-markdown";
import { tableNodes } from "prosemirror-tables";

// Keep paragraph first: ProseMirror uses the first block type to fill empty documents.
const nodes = baseSchema.spec.nodes.addBefore("text", "gallery", {
  group: "block", content: "paragraph*", isolating: true,
  attrs: { mode: { default: "grid" } },
  toDOM: (node) => ["div", { class: `image-gallery image-gallery--${node.attrs.mode}`, "data-gallery": node.attrs.mode as string, tabindex: "0" }, 0],
  parseDOM: [{ tag: "div[data-gallery]", getAttrs: (element) => ({ mode: element.getAttribute("data-gallery") === "carousel" ? "carousel" : "grid" }) }],
}).addBefore("text", "mention", {
  inline: true, group: "inline", atom: true,
  attrs: { user: {}, label: {} },
  toDOM: (node) => ["span", { "data-mention-user": node.attrs.user as string, class: "mention text-primary" }, node.attrs.label as string],
  parseDOM: [{ tag: "span[data-mention-user]", getAttrs: (element) => ({ user: element.getAttribute("data-mention-user"), label: element.textContent }) }],
}).append(
  tableNodes({
    tableGroup: "block",
    cellContent: "inline*",
    cellAttributes: {
      align: {
        default: null,
        getFromDOM: (element) => element.style.textAlign || null,
        setDOMAttr: (value, attrs) => {
          if (typeof value === "string" && value) {
            attrs.style = `text-align: ${value}`;
          }
        },
      },
    },
  }),
);
const marks = baseSchema.spec.marks.addBefore("code", "strike", {
  parseDOM: [{ tag: "del" }, { tag: "s" }],
  toDOM: () => ["del", 0],
});
export const visualMarkdownSchema = new Schema({ nodes, marks });
const boundaryTypes = new Set(["code_block", "horizontal_rule", "table", "gallery"]);
export const isBoundaryBlock = (node: ProseMirrorNode) =>
  boundaryTypes.has(node.type.name);
const tokenizer = new MarkdownIt({
  html: false,
  linkify: false,
  typographer: false,
});
tokenizer.use(forumExtensions);
export const visualMarkdownParser = new MarkdownParser(
  visualMarkdownSchema,
  tokenizer,
  {
    ...defaultMarkdownParser.tokens,
    gallery: { block: "gallery", getAttrs: (token) => ({ mode: token.attrGet("data-gallery") }) },
    mention: { node: "mention", getAttrs: (token) => ({ user: (token.meta as { id: string }).id, label: token.content }) },
    s: { mark: "strike" },
    table: { block: "table" },
    thead: { ignore: true },
    tbody: { ignore: true },
    tr: { block: "table_row" },
    th: { block: "table_header" },
    td: { block: "table_cell" },
  },
);
function renderCell(cell: ProseMirrorNode) {
  const paragraph = visualMarkdownSchema.nodes.paragraph.create(
    null,
    cell.content,
  );
  const document = visualMarkdownSchema.nodes.doc.create(null, paragraph);
  return (
    visualMarkdownSerializer
      .serialize(document)
      .replace(/\\?\n/g, " ")
      .replace(/\|/g, "\\|") || " "
  );
}
// Inline images can land in the same paragraph after a native drop or paste.
// Gallery items must remain separate blocks, including after source-mode edits.
export function normalizeGalleryItems(node: ProseMirrorNode) {
  const paragraphs: ProseMirrorNode[] = [];
  let changed = false;
  node.forEach(paragraph => {
    let imageCount = 0;
    paragraph.forEach(child => { if (child.type.name === "image") imageCount++; });
    if (imageCount < 2) { paragraphs.push(paragraph); return; }
    changed = true;
    let children: ProseMirrorNode[] = [];
    let hasImage = false;
    const flush = () => {
      while (children.at(-1)?.type.name === "hard_break" || (children.at(-1)?.isText && !children.at(-1)?.text?.trim())) children.pop();
      const last = children.at(-1);
      if (last?.isText) children[children.length - 1] = last.cut(0, last.text!.trimEnd().length);
      if (children.length) paragraphs.push(paragraph.copy(Fragment.fromArray(children)));
      children = [];
    };
    paragraph.forEach(child => {
      if (child.type.name === "image") {
        if (hasImage) flush();
        hasImage = true;
      }
      children.push(child);
    });
    flush();
  });
  return changed ? node.copy(Fragment.fromArray(paragraphs)) : node;
}

function normalizeGalleries(node: ProseMirrorNode): ProseMirrorNode {
  if (node.type.name === "gallery") return normalizeGalleryItems(node);
  if (node.isLeaf) return node;
  const children: ProseMirrorNode[] = [];
  node.forEach(child => children.push(normalizeGalleries(child)));
  return node.copy(Fragment.fromArray(children));
}

export const visualMarkdownSerializer = new MarkdownSerializer(
  {
    ...defaultMarkdownSerializer.nodes,
    gallery(state, node) {
      state.write(`[gallery mode="${node.attrs.mode}"]\n`);
      state.renderContent(normalizeGalleryItems(node));
      state.write("[/gallery]");
      state.closeBlock(node);
    },
    mention(state, node) { state.write(`[mention user="${node.attrs.user}"]${node.attrs.label}[/mention]`); },
    paragraph(state, node, parent, index) {
      if (!node.content.size) {
        (state as typeof state & { flushClose(): void }).flushClose();
        return;
      }
      defaultMarkdownSerializer.nodes.paragraph(state, node, parent, index);
    },
    bullet_list(state, node) {
      state.renderList(node, "  ", () => "- ");
    },
    table(state, node) {
      const rows = Array.from({ length: node.childCount }, (_, index) =>
        node.child(index),
      );
      if (!rows.length) return;
      const line = (row: ProseMirrorNode) =>
        `| ${Array.from({ length: row.childCount }, (_, index) => renderCell(row.child(index))).join(" | ")} |`;
      state.write(
        [
          line(rows[0]),
          `| ${Array.from({ length: rows[0].childCount }, () => "---").join(" | ")} |`,
          ...rows.slice(1).map(line),
        ].join("\n"),
      );
      state.closeBlock(node);
    },
    table_row() {},
    table_header() {},
    table_cell() {},
  },
  {
    ...defaultMarkdownSerializer.marks,
    strike: {
      open: "~~",
      close: "~~",
      mixable: true,
      expelEnclosingWhitespace: true,
    },
  },
);
export function parseVisualMarkdown(markdown: string, sourceVersion: 0 | 1 = 1) {
  return normalizeGalleries(visualMarkdownParser.parse(markdown || "", { sourceVersion }));
}
export function parseEditableVisualMarkdown(markdown: string, sourceVersion: 0 | 1 = 1) {
  const doc = parseVisualMarkdown(markdown,sourceVersion);
  const blocks: ProseMirrorNode[] = [];
  let previousBoundary = false;
  doc.forEach((node) => {
    const boundary = isBoundaryBlock(node);
    if (previousBoundary && boundary)
      blocks.push(visualMarkdownSchema.nodes.paragraph.create());
    blocks.push(node);
    previousBoundary = boundary;
  });
  if (blocks[0] && isBoundaryBlock(blocks[0]))
    blocks.unshift(visualMarkdownSchema.nodes.paragraph.create());
  if (blocks.at(-1) && isBoundaryBlock(blocks.at(-1)!))
    blocks.push(visualMarkdownSchema.nodes.paragraph.create());
  return doc.copy(Fragment.fromArray(blocks));
}
export function serializeVisualMarkdown(doc: ProseMirrorNode) {
  let content = doc.content;
  const emptyParagraph = (node: ProseMirrorNode) =>
    node.type.name === "paragraph" && node.content.size === 0;
  if (
    doc.childCount > 1 &&
    emptyParagraph(doc.firstChild!) &&
    isBoundaryBlock(doc.child(1))
  ) {
    content = content.cut(doc.firstChild!.nodeSize);
  }
  const normalized = doc.copy(content);
  if (
    normalized.childCount > 1 &&
    emptyParagraph(normalized.lastChild!) &&
    isBoundaryBlock(normalized.child(normalized.childCount - 2))
  ) {
    content = content.cut(0, content.size - normalized.lastChild!.nodeSize);
  }
  return visualMarkdownSerializer.serialize(doc.copy(content));
}
