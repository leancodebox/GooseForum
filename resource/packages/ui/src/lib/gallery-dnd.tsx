import { useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { createRoot } from "react-dom/client";
import { closestCenter, DndContext, DragOverlay, PointerSensor, useDraggable, useDroppable, useSensor, useSensors, type CollisionDetection } from "@dnd-kit/core";

interface Options {
  container: HTMLElement;
  group: HTMLElement;
  editable: () => boolean;
  start: (index: number) => void;
  over: (index: number) => void;
  finish: (from: number, to: number, point: { x: number; y: number } | null) => void;
  cancel: () => void;
}

// Register ProseMirror-owned elements without React rendering their contentDOM.
function Item({ element, index, editable }: { element: HTMLElement; index: number; editable: Options["editable"] }) {
  const { setNodeRef, listeners } = useDraggable({ id: index });
  const { setNodeRef: setDropRef } = useDroppable({ id: index });
  useLayoutEffect(() => {
    setNodeRef(element);
    setDropRef(element);
    const activate = (event: PointerEvent) => {
      if (editable() && event.pointerType === "mouse" && event.target instanceof HTMLImageElement) {
        // The React sensor's activator only consumes nativeEvent. Run before the
        // contentDOM selection handler prevents native browser text selection.
        listeners?.onPointerDown?.({ nativeEvent: event });
      }
    };
    element.addEventListener("pointerdown", activate);
    return () => { element.removeEventListener("pointerdown", activate); setNodeRef(null); setDropRef(null); };
  }, [element, editable, listeners, setNodeRef, setDropRef]);
  return null;
}

function Preview({ source, mode }: { source: HTMLElement; mode: string }) {
  const ref = useRef<HTMLDivElement>(null);
  useLayoutEffect(() => {
    const copy = source.cloneNode(true) as HTMLElement;
    copy.removeAttribute("style");
    copy.classList.remove("gallery-image-selected");
    copy.querySelectorAll(".ProseMirror-separator, .ProseMirror-trailingBreak").forEach(element => element.remove());
    ref.current?.replaceChildren(copy);
  }, [source]);
  return <div ref={ref} aria-hidden="true" className={`image-gallery image-gallery--${mode} gallery-drag-preview`} />;
}

function GalleryDnd({ options, items }: { options: Options; items: HTMLElement[] }) {
  const { container, group } = options;
  const sensorActive = useRef(false);
  const sensors = useSensors(useSensor(PointerSensor, {
    activationConstraint: { distance: 5 },
    onActivation: () => { sensorActive.current = true; },
  }));
  useLayoutEffect(() => {
    const cancel = () => {
      if (sensorActive.current) container.ownerDocument.dispatchEvent(new PointerEvent("pointercancel"));
    };
    const window = container.ownerDocument.defaultView;
    window?.addEventListener("blur", cancel);
    return () => {
      window?.removeEventListener("blur", cancel);
      // Sensors listen on document: removing the NodeView alone does not end
      // a pending gesture. Cancel it before releasing this React root.
      cancel();
    };
  }, [container]);
  const [active, setActive] = useState<number | null>(null);
  const point = useRef<{ x: number; y: number } | null>(null);
  const initialScroll = useRef(0);
  const collisionDetection: CollisionDetection = args => {
    point.current = args.pointerCoordinates;
    const p = args.pointerCoordinates;
    const rect = container.getBoundingClientRect();
    if (!p || p.x < rect.left || p.x > rect.right || p.y < rect.top || p.y > rect.bottom) return [];
    return closestCenter({ ...args, collisionRect: { left: p.x, right: p.x, top: p.y, bottom: p.y, width: 0, height: 0 } });
  };
  const clear = () => { sensorActive.current = false; group.classList.remove("is-dragging"); setActive(null); };
  return <DndContext sensors={sensors} collisionDetection={collisionDetection}
    autoScroll={{ canScroll: element => element === container, threshold: { x: 0.12, y: 0 }, acceleration: 10 }}
    onDragAbort={() => { sensorActive.current = false; }}
    onDragStart={({ active }) => {
      initialScroll.current = container.scrollLeft;
      group.classList.add("is-dragging");
      setActive(Number(active.id));
      options.start(Number(active.id));
    }}
    onDragOver={({ over }) => options.over(over ? Number(over.id) : -1)}
    onDragCancel={() => { clear(); container.scrollTo({ left: initialScroll.current, behavior: "instant" }); options.cancel(); }}
    onDragEnd={({ active, over }) => { clear(); options.finish(Number(active.id), over ? Number(over.id) : -1, point.current); }}>
    {items.map((element, index) => <Item key={index} element={element} index={index} editable={options.editable} />)}
    {createPortal(<DragOverlay dropAnimation={null} style={{ pointerEvents: "none" }}>
      {active !== null && items[active] ? <Preview source={items[active]} mode={container.dataset.gallery ?? "grid"} /> : null}
    </DragOverlay>, document.body)}
  </DndContext>;
}

export function createGalleryDnd(options: Options) {
  const host = document.createElement("div");
  host.contentEditable = "false";
  options.group.append(host);
  const root = createRoot(host);
  let registered: HTMLElement[] = [];
  return {
    update: () => {
      const items = Array.from(options.container.children) as HTMLElement[];
      // Selection/drag decorations do not change registration. Only reconcile
      // React bridges when ProseMirror actually replaces or reorders elements.
      if (items.length === registered.length && items.every((item, index) => item === registered[index])) return;
      registered = items;
      root.render(<GalleryDnd options={options} items={items} />);
    },
    destroy: () => {
      host.remove();
      // Editor teardown can itself run in a React commit. Release this nested
      // root after that commit rather than synchronously unmounting within it.
      queueMicrotask(() => root.unmount());
    },
  };
}
