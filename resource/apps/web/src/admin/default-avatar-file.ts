import { canvasWebpFile } from "@gooseforum/ui/lib/canvas-webp";

export async function defaultAvatarFile(file: File) {
  const image = await createImageBitmap(file);
  try {
    const side = Math.min(image.width, image.height);
    const size = Math.min(side, 256);
    const canvas = document.createElement("canvas");
    canvas.width = canvas.height = size;
    const context = canvas.getContext("2d");
    if (!context) throw new Error("Canvas is unavailable.");
    context.imageSmoothingQuality = "high";
    context.drawImage(image, (image.width - side) / 2, (image.height - side) / 2, side, side, 0, 0, size, size);
    return await canvasWebpFile(canvas, file.name, 0.86);
  } finally {
    image.close();
  }
}
