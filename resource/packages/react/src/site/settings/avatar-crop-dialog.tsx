import { useTranslation } from "react-i18next";
import { Button } from "../../components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../../components/ui/dialog";
import { Spinner } from "../../components/ui/spinner";
import type { useAvatarCrop } from "./use-avatar-crop";

export function AvatarCropDialog({
  crop,
}: {
  crop: ReturnType<typeof useAvatarCrop>;
}) {
  const { t } = useTranslation("settings");
  return (
    <Dialog
      open={crop.open}
      onOpenChange={(open) => {
        if (!open) crop.close();
      }}
    >
      <DialogContent
        showCloseButton={false}
        className="flex max-h-[calc(100vh-2rem)] flex-col gap-0 overflow-hidden p-0 sm:max-w-[760px]"
      >
        <DialogHeader className="border-b px-5 py-3 text-left">
          <DialogTitle>{t("avatar.cropTitle")}</DialogTitle>
          <DialogDescription>{t("avatar.cropDescription")}</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 overflow-y-auto p-4 md:grid-cols-[minmax(280px,420px)_180px] md:items-start md:justify-center">
          <div className="avatar-crop-workspace aspect-square w-full max-w-[420px] justify-self-center overflow-hidden rounded-lg border bg-muted">
            <img
              ref={crop.imageRef}
              src={crop.sourceUrl}
              alt={t("avatar.cropAlt")}
              className="block"
            />
          </div>
          <aside className="flex flex-col gap-3">
            <h3 className="text-sm font-semibold">{t("avatar.preview")}</h3>
            <div className="size-32 overflow-hidden rounded-full border bg-muted">
              <img
                src={crop.sourceUrl}
                alt={t("avatar.previewAlt")}
                className="size-full object-cover"
              />
            </div>
            <p className="rounded-lg bg-muted p-3 text-sm leading-6 text-muted-foreground">
              {t("avatar.cropTip")}
            </p>
          </aside>
        </div>
        {crop.cropError ? (
          <p className="border-t border-destructive/20 bg-destructive/10 px-5 py-3 text-sm text-destructive">
            {crop.cropError}
          </p>
        ) : null}
        <DialogFooter className="flex-row justify-end border-t bg-muted/40 px-5 py-3">
          <Button type="button" variant="secondary" onClick={crop.close}>
            {t("cancel")}
          </Button>
          <Button
            type="button"
            disabled={crop.uploading}
            onClick={() => void crop.upload()}
          >
            {crop.uploading ? <Spinner data-icon="inline-start" /> : null}
            {crop.uploading ? t("avatar.uploading") : t("avatar.confirmUpload")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
