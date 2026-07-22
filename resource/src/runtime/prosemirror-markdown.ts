import MarkdownIt from 'markdown-it'
import { Fragment, Schema, type Node as ProseMirrorNode } from 'prosemirror-model'
import {
  defaultMarkdownParser,
  defaultMarkdownSerializer,
  MarkdownParser,
  MarkdownSerializer,
  schema as baseSchema,
} from 'prosemirror-markdown'
import { tableNodes } from 'prosemirror-tables'

function renderTableCell(cell: Parameters<typeof defaultMarkdownSerializer.nodes.paragraph>[1]) {
  const paragraph = visualMarkdownSchema.nodes.paragraph.create(null, cell.content)
  const document = visualMarkdownSchema.nodes.doc.create(null, paragraph)
  const value = visualMarkdownSerializer.serialize(document).replace(/\\?\n/g, ' ').replace(/\|/g, '\\|')
  return value || ' '
}

const nodes = baseSchema.spec.nodes.append(tableNodes({
  tableGroup: 'block',
  cellContent: 'inline*',
  cellAttributes: {
    align: {
      default: null,
      getFromDOM: element => element.style.textAlign || null,
      setDOMAttr: (value, attrs) => {
        if (value) attrs.style = `text-align: ${value}`
      },
    },
  },
}))

const marks = baseSchema.spec.marks.addBefore('code', 'strike', {
  parseDOM: [{ tag: 'del' }, { tag: 's' }, { style: 'text-decoration=line-through' }],
  toDOM() {
    return ['del', 0]
  },
})

export const visualMarkdownSchema = new Schema({
  nodes,
  marks,
})

const editableBoundaryBlocks = new Set(['code_block', 'horizontal_rule', 'table'])

export function isEditableBoundaryBlock(node: ProseMirrorNode) {
  return editableBoundaryBlocks.has(node.type.name)
}

const tokenizer = new MarkdownIt({
  html: false,
  linkify: false,
  typographer: false,
})

export const visualMarkdownParser = new MarkdownParser(visualMarkdownSchema, tokenizer, {
  ...defaultMarkdownParser.tokens,
  s: { mark: 'strike' },
  table: { block: 'table' },
  thead: { ignore: true },
  tbody: { ignore: true },
  tr: { block: 'table_row' },
  th: { block: 'table_header', getAttrs: token => ({ align: token.attrGet('style')?.match(/text-align:\s*(left|center|right)/)?.[1] || null }) },
  td: { block: 'table_cell', getAttrs: token => ({ align: token.attrGet('style')?.match(/text-align:\s*(left|center|right)/)?.[1] || null }) },
})

export const visualMarkdownSerializer = new MarkdownSerializer(
  {
    ...defaultMarkdownSerializer.nodes,
    paragraph(state, node, parent, index) {
      if (!node.content.size) {
        const serializerState = state as typeof state & { flushClose: () => void }
        serializerState.flushClose()
        return
      }
      defaultMarkdownSerializer.nodes.paragraph(state, node, parent, index)
    },
    bullet_list(state, node) {
      state.renderList(node, '  ', () => '- ')
    },
    table(state, node) {
      const rows = Array.from({ length: node.childCount }, (_, rowIndex) => node.child(rowIndex))
      if (!rows.length) return
      const renderRow = (row: typeof node) => {
        const cells = Array.from({ length: row.childCount }, (_, cellIndex) => renderTableCell(row.child(cellIndex)))
        return `| ${cells.join(' | ')} |`
      }
      const lines = [
        renderRow(rows[0]),
        `| ${Array.from({ length: rows[0].childCount }, (_, index) => {
          const align = rows[0].child(index).attrs.align
          if (align === 'left') return ':---'
          if (align === 'center') return ':---:'
          if (align === 'right') return '---:'
          return '---'
        }).join(' | ')} |`,
        ...rows.slice(1).map(renderRow),
      ]
      state.write(lines.join('\n'))
      state.closeBlock(node)
    },
    table_row() {},
    table_header() {},
    table_cell() {},
  },
  {
    ...defaultMarkdownSerializer.marks,
    strike: {
      open: '~~',
      close: '~~',
      mixable: true,
      expelEnclosingWhitespace: true,
    },
  },
)

export function parseVisualMarkdown(markdown: string) {
  return visualMarkdownParser.parse(markdown || '')
}

export function parseEditableVisualMarkdown(markdown: string) {
  const doc = parseVisualMarkdown(markdown)
  const blocks: typeof doc[] = []
  let previousNeedsBoundary = false

  doc.forEach((node) => {
    const needsBoundary = isEditableBoundaryBlock(node)
    if (previousNeedsBoundary && needsBoundary) blocks.push(visualMarkdownSchema.nodes.paragraph.create())
    blocks.push(node)
    previousNeedsBoundary = needsBoundary
  })
  if (blocks[0] && isEditableBoundaryBlock(blocks[0])) blocks.unshift(visualMarkdownSchema.nodes.paragraph.create())
  if (blocks.at(-1) && isEditableBoundaryBlock(blocks.at(-1)!)) blocks.push(visualMarkdownSchema.nodes.paragraph.create())

  return doc.copy(Fragment.fromArray(blocks))
}

export function serializeVisualMarkdown(doc: Parameters<typeof visualMarkdownSerializer.serialize>[0]) {
  let content = doc.content
  const isEmptyParagraph = (node: typeof doc) => node.type.name === 'paragraph' && node.content.size === 0
  if (doc.childCount > 1 && isEmptyParagraph(doc.firstChild!) && isEditableBoundaryBlock(doc.child(1))) {
    content = content.cut(doc.firstChild!.nodeSize)
  }
  const normalizedDoc = doc.copy(content)
  if (
    normalizedDoc.childCount > 1
    && isEmptyParagraph(normalizedDoc.lastChild!)
    && isEditableBoundaryBlock(normalizedDoc.child(normalizedDoc.childCount - 2))
  ) {
    content = content.cut(0, content.size - normalizedDoc.lastChild!.nodeSize)
  }
  return visualMarkdownSerializer.serialize(doc.copy(content))
}
