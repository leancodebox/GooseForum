import { useCallback, useEffect, useRef, useState } from "react";
import type { GooseAdminApi, SiteSettings } from "@gooseforum/client";
import { Button } from "@gooseforum/ui/components/button";
import {
  Empty,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import {
  Field,
  FieldGroup,
  FieldLabel,
  FieldSet,
  FieldLegend,
} from "@gooseforum/ui/components/field";
import { Input } from "@gooseforum/ui/components/input";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Textarea } from "@gooseforum/ui/components/textarea";
import { FileText, Globe, Save, Upload } from "lucide-react";
import { toast } from "sonner";
import type { SettingsTextKey } from "../settings-i18n";
type Text = (k: SettingsTextKey) => string;
const empty: SiteSettings = {
  siteName: "",
  siteUrl: "",
  siteLogo: "",
  siteEmail: "",
  siteDescription: "",
  siteKeywords: "",
  externalLinks: "",
};
export function SiteInfoManagementPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text: Text;
}) {
  const [form, setForm] = useState(empty);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const textRef = useRef(text);
  useEffect(() => {
    textRef.current = text;
  }, [text]);
  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setForm({ ...empty, ...(await api.settings.site()) });
    } catch (r) {
      setError(msg(r, textRef.current("loadFailed")));
    } finally {
      setLoading(false);
    }
  }, [api]);
  useEffect(() => {
    void load();
  }, [load]);
  async function save(e?: React.FormEvent) {
    e?.preventDefault();
    setSaving(true);
    try {
      await api.settings.saveSite({ ...form });
      toast.success(text("saved"));
    } catch (r) {
      toast.error(msg(r, text("saveFailed")));
    } finally {
      setSaving(false);
    }
  }
  async function upload(file?: File) {
    if (!file) return;
    try {
      const r = await api.pages.uploadImage(file);
      setForm((current) => ({ ...current, siteLogo: r.url || "" }));
      toast.success(text("saved"));
    } catch (r) {
      toast.error(msg(r, text("saveFailed")));
    }
  }
  return (
    <main className="flex flex-1 flex-col gap-3 px-3 py-3 lg:px-4">
      <header className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">{text("site")}</h2>
          <p className="text-xs text-muted-foreground">{text("siteHint")}</p>
        </div>
        <Button
          size="sm"
          disabled={saving || loading}
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
      {loading ? (
        <Empty className="min-h-64 border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Spinner />
            </EmptyMedia>
            <EmptyTitle>{text("loading")}</EmptyTitle>
          </EmptyHeader>
        </Empty>
      ) : error ? (
        <div className="rounded-lg border border-destructive/30 p-3 text-destructive">
          {error}
        </div>
      ) : (
        <form onSubmit={save}>
          <div className="grid gap-8 md:grid-cols-2">
            <FieldSet>
              <FieldLegend className="flex items-center gap-2 border-b pb-2">
                <Globe className="size-5 text-muted-foreground" />
                {text("basic")}
              </FieldLegend>
              <FieldGroup>
                <Field>
                  <FieldLabel>{text("name")}</FieldLabel>
                  <Input
                    value={form.siteName}
                    onChange={(e) =>
                      setForm({ ...form, siteName: e.target.value })
                    }
                  />
                </Field>
                <Field>
                  <FieldLabel>{text("url")}</FieldLabel>
                  <Input
                    value={form.siteUrl}
                    onChange={(e) =>
                      setForm({ ...form, siteUrl: e.target.value })
                    }
                  />
                </Field>
                <Field>
                  <FieldLabel>{text("email")}</FieldLabel>
                  <Input
                    value={form.siteEmail}
                    onChange={(e) =>
                      setForm({ ...form, siteEmail: e.target.value })
                    }
                  />
                </Field>
                <Field>
                  <FieldLabel>{text("logo")}</FieldLabel>
                  <div className="flex gap-3">
                    <span className="grid size-20 shrink-0 place-items-center overflow-hidden rounded-lg border bg-muted">
                      {form.siteLogo ? (
                        <img
                          src={form.siteLogo}
                          alt="Logo"
                          className="size-full object-cover"
                        />
                      ) : (
                        <Upload className="size-8 text-muted-foreground" />
                      )}
                    </span>
                    <div className="flex flex-1 flex-col gap-2">
                      <Input
                        value={form.siteLogo}
                        onChange={(e) =>
                          setForm({ ...form, siteLogo: e.target.value })
                        }
                      />
                      <Button
                        asChild
                        variant="outline"
                        size="sm"
                        className="self-start"
                      >
                        <label>
                          <input
                            className="sr-only"
                            type="file"
                            accept="image/*"
                            onChange={(e) => void upload(e.target.files?.[0])}
                          />
                          <Upload data-icon="inline-start" />
                          {text("upload")}
                        </label>
                      </Button>
                    </div>
                  </div>
                </Field>
              </FieldGroup>
            </FieldSet>
            <FieldSet>
              <FieldLegend className="flex items-center gap-2 border-b pb-2">
                <FileText className="size-5 text-muted-foreground" />
                {text("content")}
              </FieldLegend>
              <FieldGroup>
                <Field>
                  <FieldLabel>{text("description")}</FieldLabel>
                  <Textarea
                    value={form.siteDescription}
                    onChange={(e) =>
                      setForm({ ...form, siteDescription: e.target.value })
                    }
                  />
                </Field>
                <Field>
                  <FieldLabel>{text("keywords")}</FieldLabel>
                  <Input
                    value={form.siteKeywords}
                    onChange={(e) =>
                      setForm({ ...form, siteKeywords: e.target.value })
                    }
                  />
                </Field>
                <Field>
                  <FieldLabel>{text("external")}</FieldLabel>
                  <Textarea
                    className="min-h-28 font-mono text-xs"
                    value={form.externalLinks || ""}
                    onChange={(e) =>
                      setForm({ ...form, externalLinks: e.target.value })
                    }
                  />
                </Field>
              </FieldGroup>
            </FieldSet>
          </div>
        </form>
      )}
    </main>
  );
}
function msg(r: unknown, f: string) {
  return r instanceof Error && r.message ? r.message : f;
}
