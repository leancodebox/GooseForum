export function canvasWebpFile(canvas: HTMLCanvasElement, filename: string, quality: number) {
  return new Promise<File>((resolve, reject) => {
    canvas.toBlob(blob => {
      if (blob?.type !== "image/webp") {
        reject(new Error("WebP encoding is unavailable."));
        return;
      }
      const stem = filename.replace(/\.[^.]*$/, "") || "avatar";
      resolve(new File([blob], stem + ".webp", { type: blob.type }));
    }, "image/webp", quality);
  });
}
