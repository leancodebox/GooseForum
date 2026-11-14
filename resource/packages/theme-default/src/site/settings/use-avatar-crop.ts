import { useEffect, useRef, useState, type Dispatch, type RefObject, type SetStateAction } from "react";
import { useGooseRuntime } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";

export type AvatarCropController = {
  avatarUrl: string;
  setAvatarUrl: Dispatch<SetStateAction<string>>;
  inputRef: RefObject<HTMLInputElement | null>;
  imageRef: (element: HTMLImageElement | null) => void;
  sourceUrl: string;
  previewUrl: string;
  previewAspectRatio: number;
  open: boolean;
  uploading: boolean;
  ready: boolean;
  cropError: string;
  choose(): void;
  selectFile(file?: File): void;
  close(): void;
  upload(): Promise<void>;
};

export function useAvatarCrop({
  initialUrl,
  onSuccess,
  onError,
  coverUpload,
}: {
  initialUrl: string;
  onSuccess(url: string): void;
  onError(message: string): void;
  coverUpload?(file: File): Promise<string>;
}) {
  const runtime = useGooseRuntime();
  const isCover = Boolean(coverUpload);
  const aspectRatio = isCover ? 4 / 3 : 1;
  const [previewAspectRatio, setPreviewAspectRatio] = useState(aspectRatio);
  const serverError = useServerErrorMessage();
  const inputRef = useRef<HTMLInputElement>(null);
  const [imageElement, imageRef] = useState<HTMLImageElement | null>(null);
  const cropperRef = useRef<import("cropperjs").default | undefined>(undefined);
  const sourceFile = useRef<File | undefined>(undefined);
  const previewFrame = useRef(0);
  const [avatarUrl, setAvatarUrl] = useState(initialUrl);
  const [sourceUrl, setSourceUrl] = useState("");
  const [previewUrl, setPreviewUrl] = useState("");
  const [open, setOpen] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [cropError, setCropError] = useState("");
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (!open || !sourceUrl || !imageElement) return;
    let active = true;
    let cleanupCropper: (() => void) | undefined;
    void import("cropperjs")
      .then(async ({ default: Cropper }) => {
        if (!active || !imageElement.isConnected) return;
        const cropper = new Cropper(imageElement, {
        template: `<cropper-canvas background><cropper-image translatable scalable rotatable></cropper-image><cropper-shade hidden></cropper-shade><cropper-handle action="select" plain></cropper-handle><cropper-selection ${isCover ? `initial-aspect-ratio="${aspectRatio}"` : `aspect-ratio="${aspectRatio}"`} movable resizable zoomable outlined><cropper-grid role="grid" bordered covered></cropper-grid><cropper-crosshair centered></cropper-crosshair><cropper-handle action="move" theme-color="rgba(37,99,235,.35)"></cropper-handle><cropper-handle action="n-resize"></cropper-handle><cropper-handle action="e-resize"></cropper-handle><cropper-handle action="s-resize"></cropper-handle><cropper-handle action="w-resize"></cropper-handle><cropper-handle action="ne-resize"></cropper-handle><cropper-handle action="nw-resize"></cropper-handle><cropper-handle action="se-resize"></cropper-handle><cropper-handle action="sw-resize"></cropper-handle></cropper-selection></cropper-canvas>`,
      });
        cropperRef.current = cropper;
        const container = cropper.container as HTMLElement;
        const schedulePreview = () => {
          cancelAnimationFrame(previewFrame.current);
          previewFrame.current = requestAnimationFrame(() => {
            const currentSelection = cropper.getCropperSelection();
            if (isCover && currentSelection && currentSelection.height > 0) {
              setPreviewAspectRatio(currentSelection.width / currentSelection.height);
            }
            void cropPreview(cropper).then((url) => {
              if (active) setPreviewUrl(url);
            });
          });
        };
        const selection = cropper.getCropperSelection();
        const onSelectionChange = (event: Event) => {
          if (!isCover) return;
          const { width, height } = (event as CustomEvent<{ width: number; height: number }>).detail;
          const ratio = width / height;
          if (!Number.isFinite(ratio) || ratio < 1 || ratio > 2) {
            event.preventDefault();
            return;
          }
        };
        selection?.addEventListener("change", onSelectionChange);
        container.addEventListener("pointerup", schedulePreview);
        container.addEventListener("wheel", schedulePreview, { passive: true });
        container.addEventListener("keyup", schedulePreview);
        cleanupCropper = () => {
          cancelAnimationFrame(previewFrame.current);
          selection?.removeEventListener("change", onSelectionChange);
          container.removeEventListener("pointerup", schedulePreview);
          container.removeEventListener("wheel", schedulePreview);
          container.removeEventListener("keyup", schedulePreview);
          cropper.destroy();
          if (cropperRef.current === cropper) cropperRef.current = undefined;
        };
        await centerSelection(cropper, () => active, aspectRatio);
        if (active) {
          setReady(true);
          schedulePreview();
        }
      })
      .catch(() => {
        if (active)
          setCropError(
            "裁剪组件加载失败，请刷新页面后重试。",
          );
      });
    return () => {
      active = false;
      cancelAnimationFrame(previewFrame.current);
      if (cleanupCropper) cleanupCropper();
      else {
        cropperRef.current?.destroy();
        cropperRef.current = undefined;
      }
    };
  }, [open, sourceUrl, imageElement, aspectRatio, isCover]);

  useEffect(
    () => () => {
      if (sourceUrl) URL.revokeObjectURL(sourceUrl);
    },
    [sourceUrl],
  );

  function choose() {
    inputRef.current?.click();
  }

  function selectFile(file?: File) {
    if (!file) return;
    if (!file.type.startsWith("image/")) return onError("请选择图片文件。");
    if (file.size > 5 * 1024 * 1024) return onError("图片不能超过 5 MB。");
    if (sourceUrl) URL.revokeObjectURL(sourceUrl);
    sourceFile.current = file;
    setSourceUrl(URL.createObjectURL(file));
    setPreviewUrl("");
    setPreviewAspectRatio(aspectRatio);
    setCropError("");
    setReady(false);
    setOpen(true);
    if (inputRef.current) inputRef.current.value = "";
  }

  function close() {
    if (uploading) return;
    setOpen(false);
    sourceFile.current = undefined;
    setCropError("");
    setPreviewUrl("");
    setSourceUrl((current) => {
      if (current) URL.revokeObjectURL(current);
      return "";
    });
  }

  async function upload() {
    const selection = cropperRef.current?.getCropperSelection();
    const file = sourceFile.current;
    if (!ready || uploading || !selection || !file) return;
    setUploading(true);
    setCropError("");
    try {
      const large = await selection.$toCanvas({
        width: isCover ? 1200 : 300,
        ...(isCover ? {} : { height: 300 }),
        beforeDraw(context) {
          context.imageSmoothingEnabled = true;
          context.imageSmoothingQuality = "high";
        },
      });
      const result = coverUpload
        ? await coverUpload(await canvasFile(large, file.name, 0.86))
        : await runtime.api.uploads.avatar([
            await canvasFile(large, file.name, 0.86),
            await canvasFile(resizeCanvas(large, 96), "avatar_medium.webp", 0.9),
          ]);
      setAvatarUrl(result);
      onSuccess(result);
      setOpen(false);
      setSourceUrl((current) => {
        if (current) URL.revokeObjectURL(current);
        return "";
      });
      sourceFile.current = undefined;
      setPreviewUrl("");
      setReady(false);
    } catch (reason) {
      const message = serverError(reason, isCover ? "背景上传失败。" : "头像上传失败。");
      setCropError(message);
      onError(message);
    } finally {
      setUploading(false);
    }
  }

  const controller: AvatarCropController = {
    avatarUrl,
    setAvatarUrl,
    inputRef,
    imageRef,
    sourceUrl,
    previewUrl,
    previewAspectRatio,
    open,
    uploading,
    ready,
    cropError,
    choose,
    selectFile,
    close,
    upload,
  };
  return controller;
}

async function centerSelection(
  cropper: import("cropperjs").default,
  isActive: () => boolean,
  aspectRatio: number,
) {
  const image = cropper.getCropperImage();
  const canvas = cropper.getCropperCanvas();
  const selection = cropper.getCropperSelection();
  if (!image || !canvas || !selection) throw new Error("无法初始化头像裁切，请重新选择图片。");
  await image.$ready();
  // Wait for portal layout before centering the image and creating the selection.
  for (let attempt = 0; attempt < 20 && isActive(); attempt++) {
    await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
    if (!isActive()) return;
    if (!canvas.clientWidth || !canvas.clientHeight) continue;
    image.$center("contain");
    const canvasRect = canvas.getBoundingClientRect();
    const imageRect = image.getBoundingClientRect();
    const width = Math.min(imageRect.width, imageRect.height * aspectRatio);
    const height = width / aspectRatio;
    if (width <= 0 || height <= 0) continue;
    selection.$change(
      imageRect.left - canvasRect.left + (imageRect.width - width) / 2,
      imageRect.top - canvasRect.top + (imageRect.height - height) / 2,
      width,
      height,
      aspectRatio,
      true,
    );
    return;
  }
  if (isActive()) throw new Error("裁切区域未能加载，请关闭后重新选择图片。");
}

async function cropPreview(cropper: import("cropperjs").default) {
  const selection = cropper.getCropperSelection();
  if (!selection) return "";
  try {
    const canvas = await selection.$toCanvas({
      width: 160,
      beforeDraw(context) {
        context.imageSmoothingEnabled = true;
        context.imageSmoothingQuality = "high";
      },
    });
    return canvas.toDataURL("image/webp", 0.82);
  } catch {
    return "";
  }
}

function resizeCanvas(source: HTMLCanvasElement, size: number) {
  const canvas = document.createElement("canvas");
  canvas.width = size;
  canvas.height = size;
  const context = canvas.getContext("2d");
  if (!context) throw new Error("无法处理头像。");
  context.imageSmoothingEnabled = true;
  context.imageSmoothingQuality = "high";
  context.drawImage(source, 0, 0, size, size);
  return canvas;
}

function canvasFile(
  canvas: HTMLCanvasElement,
  filename: string,
  quality: number,
) {
  return new Promise<File>((resolve, reject) => {
    canvas.toBlob(
      (blob) => {
        if (!blob) return reject(new Error("无法处理头像。"));
        resolve(
          new File([blob], filename.replace(/\.[^.]+$/, ".webp"), {
            type: blob.type,
          }),
        );
      },
      "image/webp",
      quality,
    );
  });
}
