import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { ImageIcon, Maximize2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import { cn } from "@gooseforum/ui/lib/utils";
import { readImageMetadata } from "./image-metadata";

let diagramSequence = 0;
let mermaidPromise: Promise<typeof import("mermaid").default> | undefined;
let renderQueue = Promise.resolve();

export function RenderedContent({
  html,
  className,
  variant = "post",
}: {
  html: string;
  className?: string;
  variant?: "post" | "announcement";
}) {
  const root = useRef<HTMLDivElement>(null);
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
  return (
    <>
      <div
        key={variant}
        ref={root}
        data-content-variant={variant}
        className={cn(
          "typeset gf-prose",
          variant === "announcement"
            ? "typeset-announcement gf-prose-announcement"
            : "typeset-forum gf-prose-post",
          className,
        )}
        dangerouslySetInnerHTML={{ __html: html }}
      />
      {images.map(({ image, target }, index) => createPortal(<ImageInfo image={image} />, target, String(index)))}
    </>
  );
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
