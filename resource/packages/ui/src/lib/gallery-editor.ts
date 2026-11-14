import { createGalleryDnd } from "./gallery-dnd";
import { closeHistory } from "prosemirror-history";
import { Fragment, type Node as PMNode } from "prosemirror-model";
import { NodeSelection, Plugin, PluginKey, Selection, type Command } from "prosemirror-state";
import { Decoration, DecorationSet, type EditorView, type NodeView } from "prosemirror-view";

// Let ProseMirror own temporary styles on its paragraphs as well as their content.
const dragDecorations = new PluginKey<DecorationSet>("gallery-drag");
export const galleryDragPlugin = new Plugin<DecorationSet>({
  key: dragDecorations,
  state: {
    init: () => DecorationSet.empty,
    apply: (tr, previous) => (tr.getMeta(dragDecorations) as DecorationSet | undefined) ?? (tr.docChanged ? DecorationSet.empty : previous),
  },
  props: { decorations: state => dragDecorations.getState(state) },
});

export interface GalleryLabels { grid: string; carousel: string; ungroup: string; moveBefore: string; moveAfter: string; previous: string; next: string; moveOut: string }

export function galleryNodeView(node: PMNode, view: EditorView, getPos: () => number | undefined, labels: () => GalleryLabels): NodeView {
  let current = node;
  let destroyed = false;
  let frame = 0;
  const dom = document.createElement("div");
  dom.className = "gallery-editor-block";
  const toolbar = document.createElement("div");
  toolbar.className = "gallery-editor-toolbar";
  toolbar.contentEditable = "false";
  const contentDOM = document.createElement("div");
  contentDOM.tabIndex = 0;
  let selected = -1;
  let dragOriginal: PMNode | null = null;
  let dragIndex = -1;
  const itemIndex = (target: EventTarget | null) => {
    const element = target instanceof Element ? target.closest("p") : null;
    return element?.parentElement === contentDOM ? Array.from(contentDOM.children).indexOf(element) : -1;
  };
  function move(from: number, to: number) {
    const pos = getPos();
    if (pos === undefined || !view.editable || from < 0 || to < 0 || from >= current.childCount || to >= current.childCount || from === to) return;
    const items = Array.from({ length: current.childCount }, (_, i) => current.child(i));
    const [item] = items.splice(from, 1);
    items.splice(to, 0, item);
    selected = to;
    const tr = closeHistory(view.state.tr).replaceWith(pos + 1, pos + current.nodeSize - 1, Fragment.fromArray(items));
    const target = pos + 2 + items.slice(0, to).reduce((size, item) => size + item.nodeSize, 0);
    view.dispatch(tr.setSelection(Selection.near(tr.doc.resolve(target))));
    view.focus();
  }
  function moveOut(index: number, before = false) {
    const pos = getPos();
    if (pos === undefined || !view.editable || index < 0 || index >= current.childCount) return;
    const items = Array.from({ length: current.childCount }, (_, i) => current.child(i));
    const [item] = items.splice(index, 1);
    const remaining = items.length > 1 ? [current.copy(Fragment.fromArray(items))] : items;
    const replacement = before ? [item, ...remaining] : [...remaining, item];
    selected = -1;
    const tr = closeHistory(view.state.tr).replaceWith(pos, pos + current.nodeSize, Fragment.fromArray(replacement));
    const target = pos + 1 + (before ? 0 : remaining.reduce((size, item) => size + item.nodeSize, 0));
    view.dispatch(tr.setSelection(Selection.near(tr.doc.resolve(target))).scrollIntoView());
    view.focus();
  }
  const moveOutButton = document.createElement("button");
  moveOutButton.type = "button";
  moveOutButton.textContent = labels().moveOut;
  moveOutButton.addEventListener("mousedown", event => event.preventDefault());
  moveOutButton.addEventListener("click", () => moveOut(selected));
  const moveButtons = (["moveBefore", "moveAfter"] as const).map((action, i) => {
    const button = document.createElement("button");
    button.type = "button";
    button.textContent = labels()[action];
    button.addEventListener("mousedown", event => event.preventDefault());
    button.addEventListener("click", () => move(selected, selected + (i === 0 ? -1 : 1)));
    return { button, action };
  });
  function updateSelection() {
    moveOutButton.disabled = !view.editable || selected < 0 || selected >= current.childCount;
    dom.classList.toggle("is-selected", selected >= 0);
    moveButtons[0].button.disabled = !view.editable || selected <= 0;
    moveButtons[1].button.disabled = !view.editable || selected < 0 || selected >= current.childCount - 1;
  }
  const outside = (event: Event) => {
    if (!dom.contains(event.target as globalThis.Node)) { selected = -1; dom.classList.remove("is-selected"); updateSelection(); }
  };
  document.addEventListener("pointerdown", outside);
  contentDOM.addEventListener("focus", () => {
    if (selected < 0 && current.childCount) selected = 0;
    updateSelection();
  });
  contentDOM.addEventListener("click", event => {
    selected = itemIndex(event.target);
    dom.classList.toggle("is-selected", selected >= 0);
    updateSelection();
  });
  contentDOM.addEventListener("pointerdown", event => {
    const index = itemIndex(event.target);
    if (event.pointerType !== "mouse" || event.button !== 0 || index < 0 || !view.editable || !(event.target instanceof HTMLImageElement)) return;
    // Prevent native double/triple-click text or whole-block selection from
    // taking over the gesture. Select the clicked image in ProseMirror instead.
    event.preventDefault();
    const pos = getPos();
    if (pos === undefined) return;
    let imagePos = pos + 2;
    for (let i = 0; i < index; i++) imagePos += current.child(i).nodeSize;
    const paragraph = current.child(index);
    for (let i = 0; i < paragraph.childCount && paragraph.child(i).type.name !== "image"; i++) imagePos += paragraph.child(i).nodeSize;
    selected = index;
    view.dispatch(view.state.tr.setSelection(NodeSelection.create(view.state.doc, imagePos)));
    view.focus();
    updateSelection();
  });
  // Keep sorting inside this group instead of starting ProseMirror's image move.
  dom.addEventListener("dragstart", event => { event.preventDefault(); });
  contentDOM.addEventListener("dblclick", event => { if (event.target instanceof HTMLImageElement) { event.preventDefault(); event.stopPropagation(); } });
  function decorateDrag(target: number) {
    if (destroyed) return;
    const pos = getPos();
    if (pos === undefined) return;
    const decorations: Decoration[] = [];
    let offset = pos + 1;
    current.forEach((item, _, index) => {
      const style = `${index === dragIndex ? "opacity: 0.25;" : ""}${index === target ? "outline: 3px solid var(--color-primary); outline-offset: -3px;" : ""}`;
      if (style) decorations.push(Decoration.node(offset, offset + item.nodeSize, { style }));
      offset += item.nodeSize;
    });
    view.dispatch(view.state.tr.setMeta(dragDecorations, DecorationSet.create(view.state.doc, decorations)));
  }
  function clearDrag() {
    if (destroyed) return;
    dragIndex = -1;
    view.dispatch(view.state.tr.setMeta(dragDecorations, DecorationSet.empty));
  }
  const dnd = createGalleryDnd({
    container: contentDOM, group: dom, editable: () => view.editable,
    start: index => { dragOriginal = current; dragIndex = index; decorateDrag(index); },
    over: decorateDrag,
    cancel: () => { dragOriginal = null; clearDrag(); },
    finish: (index, target, point) => {
      if (destroyed) return;
      const original = dragOriginal;
      dragOriginal = null;
      clearDrag();
      if (!original?.eq(current) || !view.editable) return;
      if (target >= 0) move(index, target);
      else if (point) {
        const bounds = contentDOM.getBoundingClientRect();
        const element = document.elementFromPoint(point.x, point.y);
        if (element && view.dom.contains(element) && !dom.contains(element) && !element.closest("[data-gallery]") && (point.y < bounds.top || point.y > bounds.bottom)) moveOut(index, point.y < bounds.top);
      }
    },
  });
  const navigation = document.createElement("div");
  navigation.contentEditable = "false";
  navigation.className = "not-typeset flex items-center justify-center gap-3 py-1 text-sm text-muted-foreground";
  const count = document.createElement("span");
  count.className = "tabular-nums";
  let visible = 0;
  const navButtons = (["previous", "next"] as const).map((action, i) => {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "rounded-md p-2 hover:bg-muted disabled:opacity-40";
    button.textContent = i === 0 ? "‹" : "›";
    button.setAttribute("aria-label", labels()[action]);
    button.addEventListener("mousedown", event => event.preventDefault());
    button.addEventListener("click", () => {
      const item = contentDOM.children[visible + (i === 0 ? -1 : 1)];
      if (item) contentDOM.scrollBy({ left: item.getBoundingClientRect().left - contentDOM.getBoundingClientRect().left });
    });
    return button;
  });
  function updateNavigation() {
    if (current.attrs.mode !== "carousel") return;
    // Carousel items occupy exactly one viewport; avoid measuring every image
    // on every scroll event (including dnd-kit's automatic scrolling).
    visible = Math.max(0, Math.min(current.childCount - 1, Math.round(contentDOM.scrollLeft / (contentDOM.clientWidth || 1))));
    count.textContent = `${visible + 1} / ${current.childCount}`;
    navButtons[0].disabled = visible === 0;
    navButtons[1].disabled = visible >= current.childCount - 1;
  }
  navigation.append(navButtons[0], count, navButtons[1]);
  contentDOM.addEventListener("scroll", updateNavigation, { passive: true });
  const buttons = (["grid", "carousel", "ungroup"] as const).map(action => {
    const button = document.createElement("button");
    button.type = "button";
    button.textContent = labels()[action];
    button.addEventListener("mousedown", event => event.preventDefault());
    button.addEventListener("click", () => {
      const pos = getPos();
      if (pos === undefined || !view.editable) return;
      const tr = closeHistory(view.state.tr);
      if (action === "ungroup") tr.replaceWith(pos, pos + current.nodeSize, current.content);
      else tr.setNodeMarkup(pos, undefined, { ...current.attrs, mode: action });
      view.dispatch(tr.scrollIntoView());
      view.focus();
    });
    toolbar.append(button);
    return { button, action };
  });
  function update(next: PMNode) {
    if (next.type !== node.type) return false;
    current = next;
    contentDOM.className = `image-gallery image-gallery--${next.attrs.mode}`;
    contentDOM.dataset.gallery = next.attrs.mode as string;
    buttons.forEach(({ button, action }) => {
      button.textContent = labels()[action];
      if (action !== "ungroup") button.setAttribute("aria-pressed", String(next.attrs.mode === action));
      button.disabled = !view.editable;
    });
    navigation.style.display = next.attrs.mode !== "carousel" || next.childCount < 2 ? "none" : "";
    cancelAnimationFrame(frame);
    frame = requestAnimationFrame(() => { updateNavigation(); updateSelection(); dnd.update(); });
    updateSelection();
    return true;
  }
  toolbar.append(...moveButtons.map(({ button }) => button), moveOutButton);
  dom.append(toolbar, contentDOM, navigation);
  update(node);

  return { dom, contentDOM, update,
    stopEvent: event => toolbar.contains(event.target as globalThis.Node) || navigation.contains(event.target as globalThis.Node),
    destroy: () => { destroyed = true; cancelAnimationFrame(frame); dnd.destroy(); document.removeEventListener("pointerdown", outside); },
    ignoreMutation: mutation => mutation.type !== "selection" && (
      !contentDOM.contains(mutation.target) ||
      (mutation.type === "attributes" && mutation.target === contentDOM)
    ),
  };
}

// Group only neighbouring image paragraphs in the same parent. Text separates groups.
export function groupAdjacentImages(mode: "grid" | "carousel"): Command {
  return (state, dispatch) => {
    const { $from } = state.selection;
    let depth = $from.depth;
    if ($from.parent.type.name === "paragraph") depth--;
    if (depth < 0) return false;
    const parent = $from.node(depth);
    if (parent.type.name === "gallery") {
      if (dispatch) dispatch(closeHistory(state.tr).setNodeMarkup($from.before(depth), undefined, { mode }));
      return true;
    }
    const paragraphs = (node: PMNode) => node.type.name === "paragraph" && node.childCount > 0 &&
      Array.from({ length: node.childCount }, (_, i) => node.child(i)).every(child => child.type.name === "image" || child.type.name === "hard_break" || (child.isText && !child.text?.trim()));
    let start = $from.index(depth);
    if (start >= parent.childCount || !paragraphs(parent.child(start))) {
      // A cursor in the empty paragraph immediately after the images is convenient.
      if (start > 0 && (start === parent.childCount || parent.child(start).content.size === 0) && paragraphs(parent.child(start - 1))) start--;
      else return false;
    }
    let end = start + 1;
    while (start > 0 && paragraphs(parent.child(start - 1))) start--;
    while (end < parent.childCount && paragraphs(parent.child(end))) end++;
    let from = $from.start(depth);
    for (let i = 0; i < start; i++) from += parent.child(i).nodeSize;
    let to = from;
    const items: PMNode[] = [];
    for (let i = start; i < end; i++) {
      const paragraph = parent.child(i);
      to += paragraph.nodeSize;
      paragraph.forEach(child => { if (child.type.name === "image") items.push(state.schema.nodes.paragraph.create(null, child)); });
    }
    if (items.length < 2 || !parent.canReplaceWith(start, end, state.schema.nodes.gallery)) return false;
    if (dispatch) dispatch(closeHistory(state.tr).replaceWith(from, to, state.schema.nodes.gallery.create({ mode }, Fragment.fromArray(items))).scrollIntoView());
    return true;
  };
}
