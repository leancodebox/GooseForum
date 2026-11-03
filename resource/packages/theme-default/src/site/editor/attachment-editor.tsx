import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import type { AttachmentInfo, AttachmentPolicy } from "@gooseforum/client";
import { useGooseRuntime } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";
import { Button } from "@gooseforum/ui/components/button";

export interface AttachmentEditorProps {
  files: AttachmentInfo[];
  onChange(files: AttachmentInfo[]): void;
  onBusy(busy: boolean): void;
  disabled?: boolean;
}
export function AttachmentList({ files }: { files: AttachmentInfo[] }) {
  const { t } = useTranslation("publish");
  if (!files.length) return null;
  return <ul aria-label={t("attachmentsTitle")} className="my-3 grid gap-2 text-sm">
    {files.map((file) => <li key={file.id} className="min-w-0 rounded-lg border px-3 py-2">
      <a href={`/file/attachments/${file.id}`} className="break-all text-primary underline" download>{file.filename}</a>
      <span className="ml-2 whitespace-nowrap text-muted-foreground">{Math.ceil(file.size / 1024)} KiB</span>
    </li>)}
  </ul>;
}
export function AttachmentEditor({ files, onChange, onBusy, disabled }: AttachmentEditorProps) {
  const runtime = useGooseRuntime();
  const { t } = useTranslation("publish");
  const serverError = useServerErrorMessage();
  const [policy, setPolicy] = useState<AttachmentPolicy>();
  const [unavailable, setUnavailable] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [failed, setFailed] = useState<File>();
  const [error, setError] = useState("");
  const latestFiles = useRef(files); latestFiles.current = files;
  const active = useRef(true);
  useEffect(() => {
    active.current = true;
    void runtime.api.attachments.policy().then((value) => { if (active.current) setPolicy(value); }).catch(() => { if (active.current) setUnavailable(true); });
    return () => { active.current = false; };
  }, [runtime]);
  const canUpload = Boolean(policy?.enabled && policy.extensions.length && files.length < policy.maxFilesPerPost);
  async function upload(file: File) {
    if (disabled || uploading || !canUpload) return;
    setUploading(true); setFailed(undefined); setError(""); onBusy(true);
    try {
      const result = await runtime.api.attachments.upload(file);
      if (active.current) { onChange([...latestFiles.current, result]); onBusy(false); }
    } catch (reason) {
      if (active.current) { setFailed(file); setError(serverError(reason, t("attachmentFailed"))); }
    } finally { if (active.current) setUploading(false); }
  }
  return <section className="mt-3 rounded-lg border p-3" aria-label={t("attachmentsTitle")}>
    <div className="flex flex-wrap items-center gap-2">
      <strong className="text-sm">{t("attachmentsTitle")}</strong>
      {policy ? <label className="cursor-pointer text-sm text-primary underline">
        {t("attachmentUpload")}
        <input aria-label={t("attachmentUpload")} type="file" accept={policy.extensions.join(",")} className="sr-only" disabled={disabled || uploading || !!failed || !canUpload} onChange={(event) => { const file = event.target.files?.[0]; event.target.value = ""; if (file) void upload(file); }} />
      </label> : null}
    </div>
    {policy ? <p className="mt-1 text-xs text-muted-foreground">{t("attachmentLimits", { count: policy.maxFilesPerPost, size: policy.maxFileSizeKb, types: policy.extensions.join(", ") })}</p> : null}
    {unavailable ? <p className="text-xs text-muted-foreground">{t("attachmentUnavailable")}</p> : null}
    <AttachmentList files={files} />
    {files.map((file) => <Button key={file.id} type="button" size="sm" variant="ghost" disabled={disabled || uploading} onClick={() => onChange(files.filter((item) => item.id !== file.id))}>{t("attachmentRemove")} <span className="max-w-48 truncate">{file.filename}</span></Button>)}
    {uploading ? <p role="status" className="text-sm">{t("attachmentUploading")}</p> : null}
    {failed ? <div role="alert" className="mt-2 text-sm text-destructive"><p>{error}</p><Button type="button" size="sm" variant="outline" disabled={disabled || uploading || !canUpload} onClick={() => void upload(failed)}>{t("attachmentRetry")}</Button><Button type="button" size="sm" variant="ghost" disabled={disabled || uploading} onClick={() => { setFailed(undefined); setError(""); onBusy(false); }}>{t("attachmentRemove")}</Button></div> : null}
  </section>;
}
