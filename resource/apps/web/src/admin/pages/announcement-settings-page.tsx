import { useCallback, useEffect, useRef, useState } from "react";
import type { AnnouncementConfig, GooseAdminApi } from "@gooseforum/client";
import { Button } from "@gooseforum/ui/components/button";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@gooseforum/ui/components/field";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Switch } from "@gooseforum/ui/components/switch";
import { Code, Save } from "lucide-react";
import { toast } from "sonner";
import type { ContentSettingsTextKey } from "../content-settings-i18n";
import { AnnouncementMarkdownEditor } from "../components/announcement-markdown-editor";
type Text = (k: ContentSettingsTextKey) => string;
export function AnnouncementSettingsPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text: Text;
}) {
  const [form, setForm] = useState<AnnouncementConfig>({
    enabled: false,
    content: "",
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const tr = useRef(text);
  useEffect(() => {
    tr.current = text;
  }, [text]);
  const load = useCallback(async () => {
    setLoading(true);
    try {
      setForm({ ...(await api.settings.announcement()) });
    } catch (r) {
      toast.error(r instanceof Error ? r.message : tr.current("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [api]);
  useEffect(() => {
    void load();
  }, [load]);
  async function save() {
    setSaving(true);
    try {
      await api.settings.saveAnnouncement(form);
      toast.success(text("saved"));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    } finally {
      setSaving(false);
    }
  }
  return (
    <main className="flex flex-1 flex-col gap-6 px-3 py-3 lg:px-4">
      <header className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">{text("announcement")}</h2>
          <p className="text-xs text-muted-foreground">
            {text("announcementHint")}
          </p>
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
      <FieldGroup className="max-w-5xl gap-6">
        <Field orientation="horizontal">
          <div className="flex-1">
            <FieldLabel>{text("announcementEnabled")}</FieldLabel>
            <FieldDescription>
              {text("announcementEnabledHint")}
            </FieldDescription>
          </div>
          <Switch
            checked={form.enabled}
            disabled={loading}
            onCheckedChange={(enabled) =>
              setForm((current) => ({ ...current, enabled }))
            }
          />
        </Field>
        <Field>
          <FieldLabel>{text("announcementContent")}</FieldLabel>
          <FieldDescription>{text("announcementContentHint")}</FieldDescription>
          <AnnouncementMarkdownEditor
            value={form.content}
            disabled={loading}
            text={text}
            onChange={(content) => setForm((current) => ({ ...current, content }))}
            onUploadImage={async (file) => {
              const result = await api.pages.uploadImage(file);
              if (!result.url) throw new Error(text("announcementEditorUploadFailed"));
              return result.url;
            }}
          />
        </Field>
        <Button
          variant="outline"
          className="self-start"
          onClick={() => {
            if (!form.content)
              setForm((current) => ({
                ...current,
                content: text("exampleContent"),
              }));
          }}
        >
          <Code data-icon="inline-start" />
          {text("example")}
        </Button>
      </FieldGroup>
    </main>
  );
}
