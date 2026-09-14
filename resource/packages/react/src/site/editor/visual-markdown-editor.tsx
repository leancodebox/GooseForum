import { forwardRef, useEffect, useImperativeHandle, useRef } from "react";
import {
  baseKeymap,
  chainCommands,
  lift,
  setBlockType,
  toggleMark,
  wrapIn,
} from "prosemirror-commands";
import { history, redo, undo } from "prosemirror-history";
import {
  inputRules,
  textblockTypeInputRule,
  wrappingInputRule,
} from "prosemirror-inputrules";
import { keymap } from "prosemirror-keymap";
import type { Node as ProseMirrorNode } from "prosemirror-model";
import {
  liftListItem,
  sinkListItem,
  splitListItem,
  wrapInList,
} from "prosemirror-schema-list";
import {
  EditorState,
  Plugin,
  TextSelection,
  type Command,
  type Transaction,
} from "prosemirror-state";
import { goToNextCell, tableEditing } from "prosemirror-tables";
import { EditorView } from "prosemirror-view";
import {
  isBoundaryBlock,
  parseEditableVisualMarkdown,
  parseVisualMarkdown,
  serializeVisualMarkdown,
  visualMarkdownSchema,
} from "./visual-markdown-schema";

function restoreBoundaryParagraphs(transaction: Transaction) {
  const insertionPoints: number[] = [];
  let position = 0;
  let previousBoundary = false;
  transaction.doc.forEach((node) => {
    const boundary = isBoundaryBlock(node);
    if (previousBoundary && boundary) insertionPoints.push(position);
    position += node.nodeSize;
    previousBoundary = boundary;
  });
  if (transaction.doc.firstChild && isBoundaryBlock(transaction.doc.firstChild))
    insertionPoints.push(0);
  if (transaction.doc.lastChild && isBoundaryBlock(transaction.doc.lastChild))
    insertionPoints.push(transaction.doc.content.size);
  if (!insertionPoints.length) return null;
  for (const point of [...new Set(insertionPoints)].sort(
    (left, right) => right - left,
  ))
    transaction.insert(point, visualMarkdownSchema.nodes.paragraph.create());
  return transaction;
}

const boundaryParagraphs = new Plugin({
  appendTransaction: (_transactions, _oldState, nextState) =>
    restoreBoundaryParagraphs(nextState.tr),
});

export type EditorAction =
  | "bold"
  | "italic"
  | "strike"
  | "inlineCode"
  | "quote"
  | "code"
  | "bulletList"
  | "orderedList"
  | "horizontalRule"
  | "hardBreak";
export interface VisualMarkdownEditorHandle {
  focus(): void;
  applyAction(action: EditorAction): void;
  insertMarkdown(markdown: string): void;
  insertText(text: string): void;
  setBlock(
    block: "paragraph" | "code_block" | `heading_${1 | 2 | 3 | 4 | 5 | 6}`,
  ): void;
  insertTable(rows: number, columns: number): void;
}

export const VisualMarkdownEditor = forwardRef<
  VisualMarkdownEditorHandle,
  {
    value: string;
    placeholder: string;
    onChange(value: string): void;
    onPaste?(event: ClipboardEvent): void;
    onDrop?(event: DragEvent): void;
  }
>(function VisualMarkdownEditor(
  { value, placeholder, onChange, onPaste, onDrop },
  ref,
) {
  const root = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const valueRef = useRef(value);
  const changeRef = useRef(onChange);
  const pasteRef = useRef(onPaste);
  const dropRef = useRef(onDrop);
  valueRef.current = value;
  changeRef.current = onChange;
  pasteRef.current = onPaste;
  dropRef.current = onDrop;
  function createState(markdown: string) {
    const listItem = visualMarkdownSchema.nodes.list_item;
    return EditorState.create({
      doc: parseEditableVisualMarkdown(markdown),
      plugins: [
        history(),
        inputRules({
          rules: [
            wrappingInputRule(
              /^\s*>\s$/,
              visualMarkdownSchema.nodes.blockquote,
            ),
            wrappingInputRule(
              /^\s*([-+*])\s$/,
              visualMarkdownSchema.nodes.bullet_list,
            ),
            wrappingInputRule(
              /^(\d+)\.\s$/,
              visualMarkdownSchema.nodes.ordered_list,
              (match) => ({ order: Number(match[1]) }),
            ),
            textblockTypeInputRule(
              /^(#{1,6})\s$/,
              visualMarkdownSchema.nodes.heading,
              (match) => ({ level: match[1]!.length }),
            ),
            textblockTypeInputRule(
              /^```$/,
              visualMarkdownSchema.nodes.code_block,
            ),
          ],
        }),
        keymap({
          "Mod-z": undo,
          "Shift-Mod-z": redo,
          "Mod-y": redo,
          "Mod-b": toggleMark(visualMarkdownSchema.marks.strong),
          "Mod-i": toggleMark(visualMarkdownSchema.marks.em),
          Enter: splitListItem(listItem),
          Tab: chainCommands(goToNextCell(1), sinkListItem(listItem)),
          "Shift-Tab": chainCommands(goToNextCell(-1), liftListItem(listItem)),
        }),
        keymap(baseKeymap),
        boundaryParagraphs,
        tableEditing(),
      ],
    });
  }
  useEffect(() => {
    if (!root.current) return;
    const view = new EditorView(root.current, {
      state: createState(valueRef.current),
      dispatchTransaction(transaction) {
        const next = view.state.apply(transaction);
        view.updateState(next);
        if (transaction.docChanged)
          changeRef.current(serializeVisualMarkdown(next.doc));
      },
      attributes: {
        class:
          "gf-prose gf-prose-post min-h-80 max-w-none px-1 py-4 outline-none",
        "data-placeholder": placeholder,
        "aria-label": placeholder,
        "aria-multiline": "true",
        role: "textbox",
      },
      handleDOMEvents: {
        paste(_view, event) {
          pasteRef.current?.(event);
          return event.defaultPrevented;
        },
        drop(_view, event) {
          dropRef.current?.(event);
          return event.defaultPrevented;
        },
      },
    });
    viewRef.current = view;
    return () => {
      view.destroy();
      viewRef.current = null;
    };
  }, []);
  useEffect(() => {
    const view = viewRef.current;
    if (!view || serializeVisualMarkdown(view.state.doc) === value) return;
    view.updateState(createState(value));
  }, [value]);
  useEffect(() => {
    viewRef.current?.setProps({
      attributes: {
        ...viewRef.current.props.attributes,
        "data-placeholder": placeholder,
        "aria-label": placeholder,
        "aria-multiline": "true",
        role: "textbox",
      },
    });
  }, [placeholder]);
  useImperativeHandle(
    ref,
    () => ({
      focus: () => viewRef.current?.focus(),
      applyAction,
      insertMarkdown,
      insertText,
      setBlock,
      insertTable,
    }),
    [],
  );
  function run(command: Command) {
    const view = viewRef.current;
    if (!view) return;
    if (command(view.state, view.dispatch, view)) view.focus();
  }
  function ancestor(...names: string[]) {
    const view = viewRef.current;
    if (!view) return null;
    const { $from } = view.state.selection;
    for (let depth = $from.depth; depth > 0; depth--) {
      const node = $from.node(depth);
      if (names.includes(node.type.name))
        return { node, pos: $from.before(depth) };
    }
    return null;
  }
  function insertNode(node: ProseMirrorNode) {
    const view = viewRef.current;
    if (!view) return;
    const { from, to } = view.state.selection;
    view.dispatch(
      view.state.tr.replaceRangeWith(from, to, node).scrollIntoView(),
    );
    view.focus();
  }
  function insertMarkdown(markdown: string) {
    const view = viewRef.current;
    if (!view) return;
    const { from, to } = view.state.selection;
    view.dispatch(
      view.state.tr
        .replaceRange(from, to, parseVisualMarkdown(markdown).slice(0))
        .scrollIntoView(),
    );
    view.focus();
  }
  function insertText(text: string) {
    const view = viewRef.current;
    if (!view) return;
    const { from, to } = view.state.selection;
    view.dispatch(view.state.tr.insertText(text, from, to).scrollIntoView());
    view.focus();
  }
  function setBlock(
    block: "paragraph" | "code_block" | `heading_${1 | 2 | 3 | 4 | 5 | 6}`,
  ) {
    if (block === "paragraph" || block === "code_block")
      run(setBlockType(visualMarkdownSchema.nodes[block]));
    else
      run(
        setBlockType(visualMarkdownSchema.nodes.heading, {
          level: Number(block.slice(-1)),
        }),
      );
  }
  function toggleList(name: "bullet_list" | "ordered_list") {
    const current = ancestor("bullet_list", "ordered_list");
    const item = visualMarkdownSchema.nodes.list_item;
    if (!current) {
      run(wrapInList(visualMarkdownSchema.nodes[name]));
      return;
    }
    if (current.node.type.name === name) {
      run(liftListItem(item));
      return;
    }
    const view = viewRef.current;
    if (!view) return;
    view.dispatch(
      view.state.tr.setNodeMarkup(
        current.pos,
        visualMarkdownSchema.nodes[name],
        name === "ordered_list" ? { order: 1 } : null,
      ),
    );
    view.focus();
  }
  function applyAction(action: EditorAction) {
    if (action === "bold") run(toggleMark(visualMarkdownSchema.marks.strong));
    else if (action === "italic")
      run(toggleMark(visualMarkdownSchema.marks.em));
    else if (action === "strike")
      run(toggleMark(visualMarkdownSchema.marks.strike));
    else if (action === "inlineCode")
      run(toggleMark(visualMarkdownSchema.marks.code));
    else if (action === "quote")
      run(
        ancestor("blockquote")
          ? lift
          : wrapIn(visualMarkdownSchema.nodes.blockquote),
      );
    else if (action === "code")
      setBlock(ancestor("code_block") ? "paragraph" : "code_block");
    else if (action === "bulletList") toggleList("bullet_list");
    else if (action === "orderedList") toggleList("ordered_list");
    else if (action === "horizontalRule")
      insertNode(visualMarkdownSchema.nodes.horizontal_rule.create());
    else insertNode(visualMarkdownSchema.nodes.hard_break.create());
  }
  function insertTable(rowsCount: number, columnsCount: number) {
    const view = viewRef.current;
    if (!view) return;
    const header = visualMarkdownSchema.nodes.table_header;
    const cell = visualMarkdownSchema.nodes.table_cell;
    const row = visualMarkdownSchema.nodes.table_row;
    const rows = [
      row.create(
        null,
        Array.from({ length: columnsCount }, () => header.create()),
      ),
      ...Array.from({ length: Math.max(0, rowsCount - 1) }, () =>
        row.create(
          null,
          Array.from({ length: columnsCount }, () => cell.create()),
        ),
      ),
    ];
    const table = visualMarkdownSchema.nodes.table.create(null, rows);
    const { from, to } = view.state.selection;
    let transaction = view.state.tr.replaceRangeWith(from, to, table);
    const selection = TextSelection.findFrom(
      transaction.doc.resolve(from),
      1,
      true,
    );
    if (selection) transaction = transaction.setSelection(selection);
    view.dispatch(transaction.scrollIntoView());
    view.focus();
  }
  return <div ref={root} className="visual-markdown-editor" />;
});
