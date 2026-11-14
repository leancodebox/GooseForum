import { lazy, Suspense, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { ImageIcon, Maximize2, ChevronLeft, ChevronRight } from "lucide-react";
import { useTranslation } from "react-i18next";
import { cn } from "@gooseforum/ui/lib/utils";
import { readImageMetadata } from "./image-metadata";

const ImagePreviewDialog = lazy(() => import("./image-preview-dialog"));

let diagramSequence = 0;
let mermaidPromise: Promise<typeof import("mermaid").default> | undefined;
let renderQueue = Promise.resolve();

export function RenderedContent({
  html,
  className,
  variant = "post",
  previewImages = false,
}: {
  html: string;
  className?: string;
  variant?: "post" | "announcement";
  previewImages?: boolean;
}) {
  const [lightbox, setLightbox] = useState<{ images: Array<{ src: string; alt: string }>; index: number } | null>(null);
  const root = useRef<HTMLDivElement>(null);
  const [galleries, setGalleries] = useState<Array<{ gallery: HTMLElement; target: HTMLElement }>>([]);
  const [images, setImages] = useState<Array<{ image: HTMLImageElement; target: HTMLElement }>>([]);
  useEffect(() => {
    if (root.current) void enhanceMermaid(root.current);
  }, [html, variant]);
  useEffect(() => {
    if (!root.current || variant !== "post") {
      setImages([]);
      return;
    }
    setImages(Array.from(root.current.querySelectorAll("img")).flatMap((image) => {
      const existing = image.closest<HTMLElement>(".gf-content-image");
      const existingTarget = existing?.querySelector<HTMLElement>(".gf-content-image-overlay");
      if (existingTarget) return [{ image, target: existingTarget }];
      const link = image.closest("a");
      if (link && (link.textContent?.trim() || link.querySelectorAll("img").length !== 1 || !root.current?.contains(link))) return [];
      const wrapper = document.createElement("span");
      wrapper.className = "gf-content-image";
      const content = link || image;
      content.before(wrapper);
      wrapper.append(content);
      const target = document.createElement("span");
      target.className = "gf-content-image-overlay not-typeset";
      wrapper.append(target);
      return [{ image, target }];
    }));
  }, [html, variant]);
  useEffect(() => {
    if (!root.current) return;
    const entries = Array.from(root.current.querySelectorAll<HTMLElement>('[data-gallery="carousel"]')).map((gallery) => {
      const target = document.createElement("div");
      target.className = "not-typeset";
      gallery.after(target);
      return { gallery, target };
    });
    setGalleries(entries);
    return () => entries.forEach(({ target }) => target.remove());
  }, [html, variant]);
  return (
    <>
      <div
        key={variant}
        ref={root}
        data-content-variant={variant}
        onClick={previewImages ? (event) => {
          if (!(event.target instanceof HTMLImageElement) || !root.current) return;
          event.preventDefault();
          event.stopPropagation();
          const elements = Array.from(root.current.querySelectorAll("img"));
          const index = elements.indexOf(event.target);
          if (index >= 0) setLightbox({ images: elements.map(image => ({ src: image.currentSrc || image.src, alt: image.alt })), index });
        } : undefined}
        className={cn(
          "typeset gf-prose",
          variant === "announcement"
            ? "typeset-announcement gf-prose-announcement"
            : "typeset-forum gf-prose-post",
          className,
        )}
        dangerouslySetInnerHTML={{ __html: html }}
      />
      {previewImages && lightbox ? <Suspense fallback={null}>
        <ImagePreviewDialog images={lightbox.images} index={lightbox.index}
          onIndexChange={index => setLightbox(current => current ? { ...current, index } : null)}
          onClose={() => setLightbox(null)} />
      </Suspense> : null}
      {galleries.map(({ gallery, target }, index) => createPortal(<GalleryControls gallery={gallery} />, target, `gallery-${index}`))}
      {images.map(({ image, target }, index) => createPortal(<ImageInfo image={image} />, target, String(index)))}
    </>
  );
}

function GalleryControls({ gallery }: { gallery: HTMLElement }) {
  const { t } = useTranslation("topic");
  const [index, setIndex] = useState(0);
  const items = Array.from(gallery.children) as HTMLElement[];
  useEffect(() => {
    const update = () => {
      // Each carousel slide has width: 100%, so its index needs no per-item
      // layout reads during touch scrolling.
      setIndex(Math.max(0, Math.min(gallery.children.length - 1, Math.round(gallery.scrollLeft / (gallery.clientWidth || 1)))));
    };
    update();
    gallery.addEventListener("scroll", update, { passive: true });
    return () => gallery.removeEventListener("scroll", update);
  }, [gallery]);
  const move = (next: number) => {
    const item = items[next];
    if (item) gallery.scrollBy({ left: item.getBoundingClientRect().left - gallery.getBoundingClientRect().left, behavior: "auto" });
  };
  if (items.length < 2) return null;
  return <div className="flex items-center justify-center gap-3 py-1 text-sm text-muted-foreground">
    <button type="button" className="rounded-md p-2 hover:bg-muted disabled:opacity-40" aria-label={t("previousImage")} disabled={index === 0} onClick={() => move(index - 1)}><ChevronLeft className="size-4" /></button>
    <span className="tabular-nums" aria-live="polite">{index + 1} / {items.length}</span>
    <button type="button" className="rounded-md p-2 hover:bg-muted disabled:opacity-40" aria-label={t("nextImage")} disabled={index === items.length - 1} onClick={() => move(index + 1)}><ChevronRight className="size-4" /></button>
  </div>;
}

function ImageInfo({ image }: { image: HTMLImageElement }) {
  const { t } = useTranslation("topic");
  const [info, setInfo] = useState(() => readImageMetadata(image));
  useEffect(() => {
    const update = () => setInfo(readImageMetadata(image));
    update();
    image.addEventListener("load", update);
    return () => image.removeEventListener("load", update);
  }, [image]);
  return (
    <button
      type="button"
      className="gf-content-image-info"
      aria-label={t("imageViewer")}
      title={t("imageViewer")}
      onClick={(event) => {
        event.stopPropagation();
        image.click();
      }}
    >
      <ImageIcon className="size-4 shrink-0" aria-hidden="true" />
      <span className="min-w-0 flex-1 truncate text-left" title={info.name}>{info.name || image.alt || t("imageViewer")}</span>
      <span className="shrink-0 text-white/75 tabular-nums">{[info.dimensions, info.size].filter(Boolean).join(" · ")}</span>
      <Maximize2 className="size-4 shrink-0" aria-hidden="true" />
    </button>
  );
}

async function enhanceMermaid(root: HTMLElement) {
  const blocks = Array.from(
    root.querySelectorAll<HTMLElement>("pre > code.language-mermaid"),
  ).filter((block) => block.parentElement?.dataset.gfEnhancement !== "mermaid");
  if (!blocks.length) return;
  if (!mermaidPromise)
    mermaidPromise = import("mermaid").then(({ default: mermaid }) => {
      mermaid.initialize({
        startOnLoad: false,
        securityLevel: "strict",
        suppressErrorRendering: true,
        fontFamily: "inherit",
      });
      return mermaid;
    });
  const mermaid = await mermaidPromise;
  for (const block of blocks) {
    const source = block.parentElement;
    if (!source?.isConnected || source.dataset.gfEnhancement) continue;
    source.dataset.gfEnhancement = "mermaid";
    try {
      const result = renderQueue.then(() =>
        mermaid.render(
          `gf-mermaid-${++diagramSequence}`,
          block.textContent || "",
        ),
      );
      renderQueue = result.then(
        () => undefined,
        () => undefined,
      );
      const { svg, bindFunctions } = await result;
      if (!source.isConnected) continue;
      const diagram = document.createElement("div");
      diagram.className = "gf-content-diagram gf-content-diagram-mermaid";
      diagram.dataset.gfEnhanced = "mermaid";
      diagram.innerHTML = svg;
      bindFunctions?.(diagram);
      source.replaceWith(diagram);
    } catch (error) {
      source.dataset.gfEnhancement = "mermaid-error";
      console.warn(
        "Unable to render Mermaid diagram; preserving its source code.",
        error,
      );
    }
  }
}
