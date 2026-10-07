export function readImageMetadata(image: HTMLImageElement) {
  let name = image.alt || "";
  try {
    const url = new URL(image.currentSrc || image.src, document.baseURI);
    if (url.protocol === "http:" || url.protocol === "https:") {
      name = decodeURIComponent(url.pathname.split("/").pop() || "") || name;
    }
  } catch {
    // Keep the accessible description for malformed or non-URL sources.
  }
  const providedBytes = Number(image.getAttribute("data-file-size"));
  const resource = performance.getEntriesByName?.(image.currentSrc || image.src)
    .filter(entry => entry.entryType === "resource").at(-1) as PerformanceResourceTiming | undefined;
  const bytes = Number.isFinite(providedBytes) && providedBytes > 0
    ? providedBytes : resource?.decodedBodySize || 0;
  return {
    name,
    dimensions: image.naturalWidth && image.naturalHeight
      ? `${image.naturalWidth} × ${image.naturalHeight}`
      : "",
    size: Number.isFinite(bytes) && bytes > 0 ? formatBytes(bytes) : "",
  };
}

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
