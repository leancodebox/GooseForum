<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { baseKeymap, chainCommands, lift, setBlockType, toggleMark, wrapIn } from 'prosemirror-commands'
import { history, redo, undo } from 'prosemirror-history'
import { inputRules, textblockTypeInputRule, wrappingInputRule } from 'prosemirror-inputrules'
import { keymap } from 'prosemirror-keymap'
import { type Mark, type Node as ProseMirrorNode } from 'prosemirror-model'
import { liftListItem, sinkListItem, splitListItem, wrapInList } from 'prosemirror-schema-list'
import { EditorState, Plugin, TextSelection, type Command, type Transaction } from 'prosemirror-state'
import { goToNextCell, tableEditing } from 'prosemirror-tables'
import { EditorView } from 'prosemirror-view'
import { isEditableBoundaryBlock, parseEditableVisualMarkdown, parseVisualMarkdown, serializeVisualMarkdown, visualMarkdownSchema } from '@/runtime/prosemirror-markdown'

const props = defineProps<{
  modelValue: string
  placeholder: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  paste: [event: ClipboardEvent]
  drop: [event: DragEvent]
  dragover: [event: DragEvent]
  dragleave: [event: DragEvent]
}>()

const root = ref<HTMLElement | null>(null)
let view: EditorView | null = null
let applyingExternalValue = false
const listItem = visualMarkdownSchema.nodes.list_item

const markdownInputRules = inputRules({
  rules: [
    wrappingInputRule(/^\s*>\s$/, visualMarkdownSchema.nodes.blockquote),
    wrappingInputRule(/^\s*([-+*])\s$/, visualMarkdownSchema.nodes.bullet_list),
    wrappingInputRule(/^(\d+)\.\s$/, visualMarkdownSchema.nodes.ordered_list, match => ({ order: Number(match[1]) })),
    textblockTypeInputRule(/^(#{1,6})\s$/, visualMarkdownSchema.nodes.heading, match => ({ level: match[1].length })),
    textblockTypeInputRule(/^```$/, visualMarkdownSchema.nodes.code_block),
  ],
})

function restoreBoundaryParagraphs(transaction: Transaction) {
  const insertionPoints: number[] = []
  let position = 0
  let previousNeedsBoundary = false
  transaction.doc.forEach((node) => {
    const needsBoundary = isEditableBoundaryBlock(node)
    if (previousNeedsBoundary && needsBoundary) insertionPoints.push(position)
    position += node.nodeSize
    previousNeedsBoundary = needsBoundary
  })
  if (transaction.doc.firstChild && isEditableBoundaryBlock(transaction.doc.firstChild)) insertionPoints.push(0)
  if (transaction.doc.lastChild && isEditableBoundaryBlock(transaction.doc.lastChild)) insertionPoints.push(transaction.doc.content.size)
  if (!insertionPoints.length) return null

  for (const insertionPoint of [...new Set(insertionPoints)].sort((a, b) => b - a)) {
    transaction.insert(insertionPoint, visualMarkdownSchema.nodes.paragraph.create())
  }
  return transaction
}

const boundaryParagraphs = new Plugin({
  appendTransaction: (_transactions, _oldState, newState) => restoreBoundaryParagraphs(newState.tr),
})

const commandKeymap = keymap({
  'Mod-z': undo,
  'Shift-Mod-z': redo,
  'Mod-y': redo,
  'Mod-b': toggleMark(visualMarkdownSchema.marks.strong),
  'Mod-i': toggleMark(visualMarkdownSchema.marks.em),
  'Mod-`': toggleMark(visualMarkdownSchema.marks.code),
  Enter: splitListItem(listItem),
  Tab: chainCommands(goToNextCell(1), sinkListItem(listItem)),
  'Shift-Tab': chainCommands(goToNextCell(-1), liftListItem(listItem)),
})

function createState(markdown: string) {
  const doc = parseEditableVisualMarkdown(markdown)
  return EditorState.create({
    doc,
    plugins: [history(), markdownInputRules, boundaryParagraphs, commandKeymap, keymap(baseKeymap), tableEditing()],
  })
}

function currentMarkdown() {
  return view ? serializeVisualMarkdown(view.state.doc) : props.modelValue
}

function run(command: Command) {
  if (!view) return false
  const handled = command(view.state, view.dispatch, view)
  if (handled) view.focus()
  return handled
}

function insertNode(node: ProseMirrorNode) {
  if (!view) return
  const { from, to } = view.state.selection
  view.dispatch(view.state.tr.replaceRangeWith(from, to, node).scrollIntoView())
  view.focus()
}

function insertMarkdown(markdown: string) {
  if (!view) return
  const parsed = parseVisualMarkdown(markdown)
  const { from, to } = view.state.selection
  view.dispatch(view.state.tr.replaceRange(from, to, parsed.slice(0)).scrollIntoView())
  view.focus()
}

function insertText(text: string) {
  if (!view) return
  const { from, to } = view.state.selection
  view.dispatch(view.state.tr.insertText(text, from, to).scrollIntoView())
  view.focus()
}

function activeLink() {
  if (!view) return null
  const { from, to, empty, $from } = view.state.selection
  const linkType = visualMarkdownSchema.marks.link
  if (!empty) {
    let mark: Mark | null = null
    view.state.doc.nodesBetween(from, to, (node) => {
      const nodeMark = linkType.isInSet(node.marks)
      if (nodeMark) mark = nodeMark
      return !mark
    })
    return mark ? { from, to, mark } : null
  }

  const mark = linkType.isInSet($from.marks())
    || linkType.isInSet($from.nodeBefore?.marks || [])
    || linkType.isInSet($from.nodeAfter?.marks || [])
  if (!mark) return null

  const parentStart = $from.start()
  const ranges: Array<{ from: number, to: number }> = []
  $from.parent.forEach((node, offset) => {
    const nodeMark = linkType.isInSet(node.marks)
    if (!node.isText || !nodeMark || !mark.eq(nodeMark)) return
    const nodeFrom = parentStart + offset
    const previous = ranges.at(-1)
    if (previous?.to === nodeFrom) previous.to += node.nodeSize
    else ranges.push({ from: nodeFrom, to: nodeFrom + node.nodeSize })
  })
  const range = ranges.find(item => item.from <= from && item.to >= from)
  return range ? { ...range, mark } : null
}

function activeLinkHref() {
  return activeLink()?.mark.attrs.href as string | undefined
}

function setLink(href: string, placeholder: string) {
  if (!view) return
  const { from, to, empty } = view.state.selection
  const link = visualMarkdownSchema.marks.link.create({ href })
  const existing = activeLink()
  if (existing) {
    view.dispatch(view.state.tr.addMark(existing.from, existing.to, link).scrollIntoView())
    view.focus()
    return
  }
  if (empty) {
    const text = placeholder
    const node = visualMarkdownSchema.text(text, [link])
    insertNode(node)
    const end = from + text.length
    view.dispatch(view.state.tr.setSelection(TextSelection.create(view.state.doc, from, end)))
    return
  }
  view.dispatch(view.state.tr.addMark(from, to, link).scrollIntoView())
  view.focus()
}

function hasAncestor(typeName: string) {
  return Boolean(findAncestor(typeName))
}

function findAncestor(...typeNames: string[]) {
  if (!view) return null
  const { $from } = view.state.selection
  for (let depth = $from.depth; depth > 0; depth -= 1) {
    const node = $from.node(depth)
    if (typeNames.includes(node.type.name)) return { node, position: $from.before(depth) }
  }
  return null
}

function toggleList(typeName: 'bullet_list' | 'ordered_list') {
  const currentList = findAncestor('bullet_list', 'ordered_list')
  if (!currentList) {
    run(wrapInList(visualMarkdownSchema.nodes[typeName]))
    return
  }
  if (currentList.node.type.name === typeName) {
    run(liftListItem(listItem))
    return
  }
  if (!view) return
  const attrs = typeName === 'ordered_list'
    ? { order: 1, tight: currentList.node.attrs.tight }
    : { tight: currentList.node.attrs.tight }
  view.dispatch(view.state.tr.setNodeMarkup(currentList.position, visualMarkdownSchema.nodes[typeName], attrs).scrollIntoView())
  view.focus()
}

function setBlock(block: 'paragraph' | 'code_block' | `heading_${1 | 2 | 3 | 4 | 5 | 6}`) {
  if (block === 'paragraph') run(setBlockType(visualMarkdownSchema.nodes.paragraph))
  else if (block === 'code_block') run(setBlockType(visualMarkdownSchema.nodes.code_block))
  else run(setBlockType(visualMarkdownSchema.nodes.heading, { level: Number(block.slice(-1)) }))
}

function insertTable(rowCount: number, columnCount: number) {
  if (!view) return
  const header = visualMarkdownSchema.nodes.table_header
  const cell = visualMarkdownSchema.nodes.table_cell
  const row = visualMarkdownSchema.nodes.table_row
  const rowsCount = Math.max(1, rowCount)
  const columnsCount = Math.max(1, columnCount)
  const rows = [
    row.create(null, Array.from({ length: columnsCount }, () => header.create())),
    ...Array.from({ length: rowsCount - 1 }, () => row.create(null, Array.from({ length: columnsCount }, () => cell.create()))),
  ]
  const table = visualMarkdownSchema.nodes.table.create(null, rows)
  const { from, to } = view.state.selection
  const currentTable = findAncestor('table')
  const insertionPosition = currentTable ? currentTable.position + currentTable.node.nodeSize : from
  let transaction = currentTable
    ? view.state.tr.insert(insertionPosition, table)
    : view.state.tr.replaceRangeWith(from, to, table)
  const cellSelection = TextSelection.findFrom(transaction.doc.resolve(insertionPosition), 1, true)
  if (cellSelection) transaction = transaction.setSelection(cellSelection)
  view.dispatch(transaction.scrollIntoView())
  view.focus()
}

function applyAction(action: 'bold' | 'italic' | 'strike' | 'inlineCode' | 'quote' | 'code' | 'bulletList' | 'orderedList' | 'horizontalRule' | 'hardBreak') {
  if (action === 'bold') run(toggleMark(visualMarkdownSchema.marks.strong))
  else if (action === 'italic') run(toggleMark(visualMarkdownSchema.marks.em))
  else if (action === 'strike') run(toggleMark(visualMarkdownSchema.marks.strike))
  else if (action === 'inlineCode') run(toggleMark(visualMarkdownSchema.marks.code))
  else if (action === 'quote') run(hasAncestor('blockquote') ? lift : wrapIn(visualMarkdownSchema.nodes.blockquote))
  else if (action === 'code') setBlock(hasAncestor('code_block') ? 'paragraph' : 'code_block')
  else if (action === 'bulletList') toggleList('bullet_list')
  else if (action === 'orderedList') toggleList('ordered_list')
  else if (action === 'horizontalRule') insertNode(visualMarkdownSchema.nodes.horizontal_rule.create())
  else insertNode(visualMarkdownSchema.nodes.hard_break.create())
}

function focus() {
  view?.focus()
}

function restoreEditableBoundaries() {
  if (!view) return
  const transaction = restoreBoundaryParagraphs(view.state.tr)
  if (transaction) view.dispatch(transaction)
}

onMounted(() => {
  if (!root.value) return
  view = new EditorView(root.value, {
    state: createState(props.modelValue),
    dispatchTransaction(transaction) {
      if (!view) return
      const nextState = view.state.apply(transaction)
      view.updateState(nextState)
      if (transaction.docChanged && !applyingExternalValue) emit('update:modelValue', serializeVisualMarkdown(nextState.doc))
    },
    attributes: {
      class: 'gf-prose gf-prose-post min-h-80 max-w-none px-1 py-4 outline-none',
      'data-placeholder': props.placeholder,
      'aria-label': props.placeholder,
    },
    handleDOMEvents: {
      paste(_view, event) {
        emit('paste', event)
        return event.defaultPrevented
      },
      drop(_view, event) {
        emit('drop', event)
        return event.defaultPrevented
      },
      dragover(_view, event) {
        emit('dragover', event)
        return event.defaultPrevented
      },
      dragleave(_view, event) {
        emit('dragleave', event)
        return event.defaultPrevented
      },
    },
  })
})

watch(() => props.modelValue, (markdown) => {
  if (!view || markdown === currentMarkdown()) return
  applyingExternalValue = true
  view.updateState(createState(markdown))
  applyingExternalValue = false
})

watch(() => props.placeholder, (placeholder) => {
  view?.setProps({ attributes: { ...view.props.attributes, 'data-placeholder': placeholder, 'aria-label': placeholder } })
})

onBeforeUnmount(() => {
  view?.destroy()
  view = null
})

defineExpose({
  applyAction,
  focus,
  restoreEditableBoundaries,
  insertMarkdown,
  insertTable,
  insertText,
  activeLinkHref,
  setBlock,
  setLink,
})
</script>

<template>
  <div ref="root" class="visual-markdown-editor" />
</template>

<style scoped>
.visual-markdown-editor :deep(.ProseMirror) {
  position: relative;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.visual-markdown-editor :deep(.ProseMirror:has(> p:only-child > br.ProseMirror-trailingBreak)::before) {
  color: color-mix(in oklch, var(--gf-color-base-content) 45%, transparent);
  content: attr(data-placeholder);
  left: 0.25rem;
  pointer-events: none;
  position: absolute;
  top: 1rem;
}

.visual-markdown-editor :deep(.ProseMirror-selectednode) {
  outline: 2px solid var(--gf-color-primary);
  outline-offset: 2px;
}

.visual-markdown-editor :deep(th),
.visual-markdown-editor :deep(td) {
  position: relative;
}

.visual-markdown-editor :deep(.selectedCell::after) {
  background: color-mix(in oklch, var(--gf-color-primary) 14%, transparent);
  content: '';
  inset: 0;
  pointer-events: none;
  position: absolute;
}
</style>
