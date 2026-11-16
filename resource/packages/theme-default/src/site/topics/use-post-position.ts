import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";
import type { RefObject } from "react";
import type { PostPayload } from "@gooseforum/client";

type Position = { postNo: number; progress: number };
type Geometry = { postNo: number; top: number; bottom: number; anchor: number };
const clamp = (value: number) => Math.max(0, Math.min(1, value));
const readingLine = () => Math.min(window.innerHeight * 0.38, 340);

export function usePostPosition({ root, posts, max, hasAfter, current, load, cancelLoad }: {
  root: RefObject<HTMLDivElement | null>;
  posts: PostPayload[];
  max: number;
  hasAfter: boolean;
  current: number;
  load(value: number): Promise<number | undefined>;
  cancelLoad(): void;
}) {
  const [position, setPosition] = useState<Position>({ postNo: current, progress: (current - 1) / Math.max(1, max - 1) });
  const [target, setTarget] = useState<{ postNo: number; ready: boolean }>();
  const jumping = useRef(false);
  const sequence = useRef(0);
  const animation = useRef(0);
  const measure = useRef(() => {});
  const geometry = useRef<Geometry[]>([]);
  const previousCurrent = useRef(current);

  useLayoutEffect(() => {
    if (previousCurrent.current !== current) {
      previousCurrent.current = current;
      setPosition({ postNo: current, progress: (current - 1) / Math.max(1, max - 1) });
    }
  }, [current, max]);

  useLayoutEffect(() => {
    const nodes = Array.from(root.current?.querySelectorAll<HTMLElement>('[data-post-no]') || []);
    let frame = 0;
    let dirty = true;
    const update = () => {
      frame = 0;
      if (dirty) {
        geometry.current = nodes.map(node => {
          const rect = node.getBoundingClientRect();
          return { postNo: Number(node.dataset.postNo), top: rect.top + scrollY, bottom: rect.bottom + scrollY,
            anchor: Math.max(0, Math.min(document.documentElement.scrollHeight - innerHeight, rect.top + scrollY - readingLine())) };
        }).filter(item => item.bottom > item.top);
        dirty = false;
      }
      const items = geometry.current;
      if (jumping.current || !items.length) return;
      const line = scrollY;
      // Binary search the cached anchors. Normal scroll frames do not measure every post.
      let low = 0;
      let high = items.length - 1;
      while (low < high) {
        const middle = Math.ceil((low + high) / 2);
        if (items[middle].anchor <= line + 1) low = middle;
        else high = middle - 1;
      }
      const last = items[items.length - 1];
      const atEnd = !hasAfter && low === items.length - 1 && line >= last.anchor - 1;
      const item = items[low];
      const next = items[low + 1];
      const fraction = clamp((line - item.anchor) / Math.max(1, (next?.anchor ?? item.anchor) - item.anchor));
      const nextNo = next === last && !hasAfter ? max : next?.postNo ?? item.postNo;
      const virtualPost = item.postNo + fraction * (nextNo - item.postNo);
      const progress = atEnd ? 1 : clamp((virtualPost - 1) / Math.max(1, max - 1));
      setPosition(previous => previous.postNo === item.postNo && Math.abs(previous.progress - progress) < 0.0001 ? previous : { postNo: item.postNo, progress });
    };
    const schedule = () => { if (!frame) frame = requestAnimationFrame(update); };
    const invalidate = () => { dirty = true; schedule(); };
    measure.current = invalidate;
    const observer = typeof ResizeObserver === 'undefined' ? undefined : new ResizeObserver(invalidate);
    nodes.forEach(node => observer?.observe(node));
    if (root.current) observer?.observe(root.current);
    observer?.observe(document.body);
    window.addEventListener('scroll', schedule, { passive: true });
    window.addEventListener('resize', invalidate);
    // A parent layout effect may restore a visible post after prepending replies.
    // Read the settled viewport before paint, not the intermediate inserted layout.
    schedule();
    return () => {
      observer?.disconnect();
      window.removeEventListener('scroll', schedule);
      window.removeEventListener('resize', invalidate);
      cancelAnimationFrame(frame);
    };
  }, [root, posts, max, hasAfter]);

  const cancel = useCallback(() => {
    if (!jumping.current) return;
    sequence.current++;
    jumping.current = false;
    cancelAnimationFrame(animation.current);
    cancelLoad();
    setTarget(undefined);
    measure.current();
  }, [cancelLoad]);

  useEffect(() => {
    const key = (event: KeyboardEvent) => {
      const node = event.target as HTMLElement | null;
      if (node?.closest('input, textarea, select, [contenteditable="true"], [role="slider"]')) return;
      if (['ArrowUp', 'ArrowDown', 'PageUp', 'PageDown', 'Home', 'End', ' '].includes(event.key)) cancel();
    };
    window.addEventListener('wheel', cancel, { passive: true });
    window.addEventListener('touchstart', cancel, { passive: true });
    window.addEventListener('keydown', key);
    return () => {
      window.removeEventListener('wheel', cancel);
      window.removeEventListener('touchstart', cancel);
      window.removeEventListener('keydown', key);
    };
  }, [cancel]);

  const select = async (postNo: number) => {
    const version = ++sequence.current;
    cancelAnimationFrame(animation.current);
    jumping.current = true;
    setTarget({ postNo, ready: false });
    const loaded = posts.some(post => post.postNo === postNo);
    if (loaded) cancelLoad();
    const resolved = loaded ? postNo : await load(postNo);
    if (version !== sequence.current) return;
    if (resolved === undefined) { cancel(); return; }
    setTarget({ postNo: resolved, ready: true });
  };

  useLayoutEffect(() => {
    if (!target?.ready) return;
    const node = root.current?.querySelector<HTMLElement>(`[data-post-no="${target.postNo}"]`);
    if (!node) return;
    const start = scrollY;
    const destination = Math.max(0, Math.min(document.documentElement.scrollHeight - innerHeight, node.getBoundingClientRect().top + start - readingLine() + 1));
    const distance = destination - start;
    const reducedMotion = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches;
    const duration = reducedMotion || Math.abs(distance) > innerHeight * 2 ? 0 : Math.min(360, Math.abs(distance) * 0.35);
    const began = performance.now();
    const frame = (now: number) => {
      const fraction = duration ? clamp((now - began) / duration) : 1;
      const eased = 1 - Math.pow(1 - fraction, 3);
      window.scrollTo({ top: start + distance * eased, behavior: 'instant' });
      if (fraction < 1) animation.current = requestAnimationFrame(frame);
      else {
        jumping.current = false;
        setPosition({ postNo: target.postNo, progress: (target.postNo - 1) / Math.max(1, max - 1) });
        setTarget(undefined);
        measure.current();
      }
    };
    animation.current = requestAnimationFrame(frame);
    return () => cancelAnimationFrame(animation.current);
  }, [target, posts, root, max]);

  useEffect(() => {
    // Activity preserves state but restarts effects. An interrupted jump must not resume on Back.
    // eslint-disable-next-line react-hooks/set-state-in-effect -- Activity reactivation must discard its preserved, interrupted jump.
    setTarget(undefined);
    return () => {
      // eslint-disable-next-line react-hooks/exhaustive-deps -- This is a request generation, not a captured DOM ref; invalidate the latest generation.
      sequence.current++;
      jumping.current = false;
      cancelAnimationFrame(animation.current);
    };
  }, []);

  return { ...position, target: target?.postNo, select, cancel };
}
