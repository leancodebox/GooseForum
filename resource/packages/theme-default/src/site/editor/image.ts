import type CompressorType from "compressorjs";

export function validateImage(file: File, maxSize = 10 * 1024 * 1024) {
  return file.type.startsWith("image/") && file.size <= maxSize;
}

export async function optimizeImage(file: File, quality = 0.85) {
  const mimeType = supportsWebP() ? "image/webp" : file.type || "image/jpeg";
  try {
    const { default: Compressor } = (await import("compressorjs")) as {
      default: typeof CompressorType;
    };
    return await new Promise<File>((resolve, reject) => {
      new Compressor(file, {
        quality,
        mimeType,
        checkOrientation: true,
        convertSize: 0,
        success(result) {
          resolve(
            result instanceof File && result.type === mimeType
              ? result
              : new File([result], outputName(file.name, mimeType), {
                  type: mimeType,
                  lastModified: Date.now(),
                }),
          );
        },
        error: reject,
      });
    });
  } catch (error) {
    console.warn(
      "Unable to optimize image; uploading the original file.",
      error,
    );
    return file;
  }
}

function supportsWebP() {
  const canvas = document.createElement("canvas");
  canvas.width = 1;
  canvas.height = 1;
  return canvas.toDataURL("image/webp").startsWith("data:image/webp");
}

function outputName(filename: string, mimeType: string) {
  const extension = mimeType === "image/webp" ? ".webp" : ".jpg";
  return /\.[^/.]+$/.test(filename)
    ? filename.replace(/\.[^/.]+$/, extension)
    : `${filename}${extension}`;
}
