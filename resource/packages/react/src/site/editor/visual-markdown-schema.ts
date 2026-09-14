import MarkdownIt from "markdown-it";
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

const nodes = baseSchema.spec.nodes.append(
  tableNodes({
    tableGroup: "block",
    cellContent: "inline*",
    cellAttributes: {
      align: {
        default: null,
        getFromDOM: (element) => element.style.textAlign || null,
        setDOMAttr: (value, attrs) => {
          if (value) attrs.style = `text-align: ${value}`;
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
const boundaryTypes = new Set(["code_block", "horizontal_rule", "table"]);
export const isBoundaryBlock = (node: ProseMirrorNode) =>
  boundaryTypes.has(node.type.name);
const tokenizer = new MarkdownIt({
  html: false,
  linkify: false,
  typographer: false,
});
export const visualMarkdownParser = new MarkdownParser(
  visualMarkdownSchema,
  tokenizer,
  {
    ...defaultMarkdownParser.tokens,
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
export const visualMarkdownSerializer = new MarkdownSerializer(
  {
    ...defaultMarkdownSerializer.nodes,
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
export function parseVisualMarkdown(markdown: string) {
  return visualMarkdownParser.parse(markdown || "");
}
export function parseEditableVisualMarkdown(markdown: string) {
  const doc = parseVisualMarkdown(markdown);
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
