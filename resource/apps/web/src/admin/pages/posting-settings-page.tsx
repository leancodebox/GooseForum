import { useCallback, useEffect, useRef, useState } from "react";
import type { GooseAdminApi, PostingSettings } from "@gooseforum/client";
import { Badge } from "@gooseforum/ui/components/badge";
import { Button } from "@gooseforum/ui/components/button";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@gooseforum/ui/components/field";
import { Input } from "@gooseforum/ui/components/input";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Switch } from "@gooseforum/ui/components/switch";
import { FileText, Plus, Save, Trash2, Upload } from "lucide-react";
import { toast } from "sonner";
import type { ContentSettingsTextKey } from "../content-settings-i18n";
type Text = (k: ContentSettingsTextKey) => string;
const defaults: PostingSettings = {
  textControl: {
    minPostLength: 5,
    maxPostLength: 50000,
    minTitleLength: 5,
    maxTitleLength: 100,
    newUserPostCooldownMinutes: 0,
    maxDailyTopicsPerUser: 10,
  },
  uploadControl: {
    allowAttachments: true,
    authorizedExtensions: [".jpg", ".jpeg", ".png", ".gif", ".webp"],
    maxAttachmentSizeKb: 5120,
    maxDailyUploadsPerUser: 10,
    newUserUploadCooldownMinutes: 0,
  },
};
export function PostingSettingsPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text: Text;
}) {
  const [form, setForm] = useState(defaults);
  const [extension, setExtension] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const tr = useRef(text);
  useEffect(() => {
    tr.current = text;
  }, [text]);
  const load = useCallback(async () => {
    setLoading(true);
    try {
      const v = await api.settings.posting();
      setForm({
        textControl: { ...defaults.textControl, ...v.textControl },
        uploadControl: {
          ...defaults.uploadControl,
          ...v.uploadControl,
          authorizedExtensions: Array.isArray(
            v.uploadControl?.authorizedExtensions,
          )
            ? v.uploadControl.authorizedExtensions
            : defaults.uploadControl.authorizedExtensions,
        },
      });
    } catch (r) {
      toast.error(r instanceof Error ? r.message : tr.current("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [api]);
  useEffect(() => {
    void load();
  }, [load]);
  function number(
    section: "textControl" | "uploadControl",
    key: string,
    value: number,
  ) {
    setForm((current) => ({
      ...current,
      [section]: { ...current[section], [key]: value },
    }));
  }
  function add() {
    const v = extension.trim().toLowerCase();
    if (!v) return;
    if (!v.startsWith(".")) {
      toast.warning(text("invalidExtension"));
      return;
    }
    if (!form.uploadControl.authorizedExtensions.includes(v))
      setForm({
        ...form,
        uploadControl: {
          ...form.uploadControl,
          authorizedExtensions: [...form.uploadControl.authorizedExtensions, v],
        },
      });
    setExtension("");
  }
  async function save() {
    setSaving(true);
    try {
      await api.settings.savePosting(form);
      toast.success(text("saved"));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    } finally {
      setSaving(false);
    }
  }
  return (
    <main className="flex flex-1 flex-col gap-5 px-3 py-3 lg:px-4">
      <header className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">{text("posting")}</h2>
          <p className="text-xs text-muted-foreground">{text("postingHint")}</p>
        </div>
        <Button
          size="sm"
          disabled={loading || saving}
          onClick={() => void save()}
        >
          {saving ? (
            <Spinner data-icon="inline-start" />
          ) : (
            <Save data-icon="inline-start" />
          )}
          {text("save")}
        </Button>
      </header>
      <div className="grid gap-10 lg:grid-cols-2">
        <FieldSet>
          <FieldLegend className="flex items-center gap-2 border-b pb-2">
            <FileText className="size-5 text-muted-foreground" />
            {text("text")}
          </FieldLegend>
          <div className="grid gap-5 sm:grid-cols-2">
            <N
              label={text("minTitle")}
              value={form.textControl.minTitleLength}
              disabled={loading}
              onChange={(v) => number("textControl", "minTitleLength", v)}
            />
            <N
              label={text("maxTitle")}
              value={form.textControl.maxTitleLength}
              disabled={loading}
              onChange={(v) => number("textControl", "maxTitleLength", v)}
            />
            <N
              label={text("minPost")}
              value={form.textControl.minPostLength}
              disabled={loading}
              onChange={(v) => number("textControl", "minPostLength", v)}
            />
            <N
              label={text("maxPost")}
              value={form.textControl.maxPostLength}
              disabled={loading}
              onChange={(v) => number("textControl", "maxPostLength", v)}
            />
            <N
              label={text("postCooldown")}
              value={form.textControl.newUserPostCooldownMinutes}
              disabled={loading}
              onChange={(v) =>
                number("textControl", "newUserPostCooldownMinutes", v)
              }
            />
            <N
              label={text("dailyTopics")}
              value={form.textControl.maxDailyTopicsPerUser}
              disabled={loading}
              onChange={(v) =>
                number("textControl", "maxDailyTopicsPerUser", v)
              }
            />
          </div>
        </FieldSet>
        <FieldSet>
          <FieldLegend className="flex items-center gap-2 border-b pb-2">
            <Upload className="size-5 text-muted-foreground" />
            {text("uploads")}
          </FieldLegend>
          <FieldGroup>
            <Field
              orientation="horizontal"
              className="rounded-lg border bg-muted/10 p-4"
            >
              <div className="flex-1">
                <FieldLabel>{text("allowAttachments")}</FieldLabel>
                <FieldDescription>{text("allowHint")}</FieldDescription>
              </div>
              <Switch
                checked={form.uploadControl.allowAttachments}
                onCheckedChange={(allowAttachments) =>
                  setForm({
                    ...form,
                    uploadControl: { ...form.uploadControl, allowAttachments },
                  })
                }
              />
            </Field>
            <div className="grid gap-5 sm:grid-cols-2">
              <N
                label={text("dailyUploads")}
                value={form.uploadControl.maxDailyUploadsPerUser}
                disabled={!form.uploadControl.allowAttachments}
                onChange={(v) =>
                  number("uploadControl", "maxDailyUploadsPerUser", v)
                }
              />
              <N
                label={text("maxSize")}
                value={form.uploadControl.maxAttachmentSizeKb}
                disabled={!form.uploadControl.allowAttachments}
                onChange={(v) =>
                  number("uploadControl", "maxAttachmentSizeKb", v)
                }
              />
              <N
                label={text("uploadCooldown")}
                value={form.uploadControl.newUserUploadCooldownMinutes}
                disabled={!form.uploadControl.allowAttachments}
                onChange={(v) =>
                  number("uploadControl", "newUserUploadCooldownMinutes", v)
                }
              />
            </div>
            <Field>
              <FieldLabel>{text("extensions")}</FieldLabel>
              <div className="flex gap-2">
                <Input
                  value={extension}
                  onChange={(e) => setExtension(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault();
                      add();
                    }
                  }}
                  placeholder={text("extensionPlaceholder")}
                />
                <Button variant="secondary" onClick={add}>
                  <Plus data-icon="inline-start" />
                  {text("add")}
                </Button>
              </div>
              <div className="flex flex-wrap gap-2">
                {form.uploadControl.authorizedExtensions.map((v) => (
                  <Badge key={v} variant="secondary" className="gap-2">
                    {v}
                    <Button
                      variant="ghost"
                      size="icon-xs"
                      onClick={() =>
                        setForm({
                          ...form,
                          uploadControl: {
                            ...form.uploadControl,
                            authorizedExtensions:
                              form.uploadControl.authorizedExtensions.filter(
                                (x) => x !== v,
                              ),
                          },
                        })
                      }
                    >
                      <Trash2 />
                    </Button>
                  </Badge>
                ))}
              </div>
            </Field>
          </FieldGroup>
        </FieldSet>
      </div>
    </main>
  );
}
function N({
  label,
  value,
  onChange,
  disabled,
}: {
  label: string;
  value: number;
  onChange(v: number): void;
  disabled: boolean;
}) {
  return (
    <Field>
      <FieldLabel>{label}</FieldLabel>
      <Input
        type="number"
        value={value}
        disabled={disabled}
        onChange={(e) => onChange(Number(e.target.value))}
      />
    </Field>
  );
}
