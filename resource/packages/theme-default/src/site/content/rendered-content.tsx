import { useEffect, useRef } from "react";
import { cn } from "@gooseforum/ui/lib/utils";

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
  useEffect(() => {
    if (root.current) void enhanceMermaid(root.current);
  }, [html]);
  return (
    <div
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
