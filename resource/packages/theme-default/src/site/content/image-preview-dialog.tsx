import { Dialog, DialogContent, DialogTitle } from "@gooseforum/ui/components/dialog";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useTranslation } from "react-i18next";

// Composer-only viewer: load on demand rather than adding a dialog to every post.
export default function ImagePreviewDialog({ images, index, onIndexChange, onClose }: {
  images: Array<{ src: string; alt: string }>;
  index: number;
  onIndexChange: (index: number) => void;
  onClose: () => void;
}) {
  const { t } = useTranslation("topic");
  const step = (direction: number) => onIndexChange(Math.max(0, Math.min(images.length - 1, index + direction)));
  return <Dialog open onOpenChange={open => { if (!open) onClose(); }}>
    <DialogContent className="sm:max-w-4xl" onKeyDown={event => {
      if (event.key === "ArrowLeft" || event.key === "ArrowRight") { event.preventDefault(); step(event.key === "ArrowLeft" ? -1 : 1); }
    }}>
      <DialogTitle>{t("imageViewer")}</DialogTitle>
      <img src={images[index].src} alt={images[index].alt} className="mx-auto max-h-[75vh] max-w-full object-contain" />
      <div className="flex items-center justify-center gap-4">
        <button type="button" aria-label={t("previousImage")} disabled={index === 0} onClick={() => step(-1)}><ChevronLeft /></button>
        <span>{index + 1} / {images.length}</span>
        <button type="button" aria-label={t("nextImage")} disabled={index === images.length - 1} onClick={() => step(1)}><ChevronRight /></button>
      </div>
    </DialogContent>
  </Dialog>;
}
