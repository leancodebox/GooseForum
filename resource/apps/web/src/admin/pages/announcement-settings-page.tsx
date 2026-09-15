import { useCallback, useEffect, useRef, useState } from "react";
import type { AnnouncementConfig, GooseAdminApi } from "@gooseforum/client";
import { Button } from "@gooseforum/ui/components/button";
import {
  Field,
  FieldDescription,
  FieldLabel,
} from "@gooseforum/ui/components/field";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Switch } from "@gooseforum/ui/components/switch";
import { Textarea } from "@gooseforum/ui/components/textarea";
import { Code, Save } from "lucide-react";
import { toast } from "sonner";
import type { ContentSettingsTextKey } from "../content-settings-i18n";
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
      <div className="flex max-w-3xl flex-col gap-6">
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
            onCheckedChange={(enabled) => setForm({ ...form, enabled })}
          />
        </Field>
        <Field>
          <FieldLabel>{text("announcementContent")}</FieldLabel>
          <FieldDescription>{text("announcementContentHint")}</FieldDescription>
          <Textarea
            className="min-h-64 resize-y font-mono text-sm"
            value={form.content}
            disabled={loading}
            onChange={(e) => setForm({ ...form, content: e.target.value })}
          />
        </Field>
        <Button
          variant="outline"
          className="self-start"
          onClick={() => {
            if (!form.content)
              setForm({ ...form, content: text("exampleContent") });
          }}
        >
          <Code data-icon="inline-start" />
          {text("example")}
        </Button>
      </div>
    </main>
  );
}
