import { defaultAvatarFile } from "../default-avatar-file";
import { errorMessage } from "../error-message";
import { AdminPage } from '../components/admin-page'
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
import { Switch } from "@gooseforum/ui/components/switch";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Textarea } from "@gooseforum/ui/components/textarea";
import { ChevronDown, CircleMinus, FileText, Globe, Save, Upload } from "lucide-react";
import { toast } from "@gooseforum/ui/components/sonner";
import type { SettingsTextKey } from "../settings-i18n";
type Text = (k: SettingsTextKey) => string;
const avatarImageClass = "size-full rounded-full bg-muted object-cover";
type AvatarOption = { file: string; enabled: boolean };

function avatarOptions(settings: SiteSettings, builtins: string[]): AvatarOption[] {
  const pool = settings.avatarPool ?? (settings.defaultAvatarFile
    ? [{ file: settings.defaultAvatarFile, enabled: true }]
    : builtins.map(file => ({ file, enabled: true })));
  return [
    ...builtins.map(file => ({ file, enabled: pool.some(option => option.file === file && option.enabled) })),
    ...pool.filter(option => !builtins.includes(option.file)),
  ];
}
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
  const [builtins, setBuiltins] = useState<string[]>([]);
  const [form, setForm] = useState(empty);
  const [loading, setLoading] = useState(true);
  const [uploadingAvatar, setUploadingAvatar] = useState(false);
  const [avatarsExpanded, setAvatarsExpanded] = useState(false);
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
      const { builtinAvatarFiles, ...settings } = await api.settings.site();
      if (!Array.isArray(builtinAvatarFiles)) throw new Error(textRef.current("loadFailed"));
      setBuiltins(builtinAvatarFiles);
      setForm({ ...empty, ...settings });
    } catch (r) {
      setError(errorMessage(r, textRef.current("loadFailed")));
    } finally {
      setLoading(false);
    }
  }, [api]);
  useEffect(() => {
    void load();
  }, [load]);
  const pool = avatarOptions(form, builtins);
  const enabledAvatars = pool.filter(option => option.enabled);
  const missingSelection = !form.disableDefaultAvatar && enabledAvatars.length === 0;
  function updatePool(update: (options: AvatarOption[]) => AvatarOption[]) {
    setForm(current => ({ ...current, defaultAvatarFile: "", avatarPool: update(avatarOptions(current, builtins)) }));
  }
  async function save(e?: React.FormEvent) {
    e?.preventDefault();
    if (uploadingAvatar || saving || missingSelection) return;
    setSaving(true);
    try {
      await api.settings.saveSite({ ...form, defaultAvatarFile: "", avatarPool: pool });
      toast.success(text("saved"));
    } catch (r) {
      toast.error(errorMessage(r, text("saveFailed")));
    } finally {
      setSaving(false);
    }
  }
  async function uploadAvatars(files: File[]) {
    setUploadingAvatar(true);
    try {
      for (const file of files) {
        try {
          const { objectName } = await api.pages.uploadImage(await defaultAvatarFile(file));
          if (!objectName) throw new Error(text("saveFailed"));
          updatePool(options => options.some(option => option.file === objectName)
            ? options : [...options, { file: objectName, enabled: true }]);
        } catch (reason) {
          toast.error(file.name + " · " + errorMessage(reason, text("saveFailed")));
        }
      }
    } finally {
      setUploadingAvatar(false);
    }
  }
  async function upload(file?: File) {
    if (!file) return;
    try {
      const r = await api.pages.uploadImage(file);
      setForm((current) => ({ ...current, siteLogo: r.url || "" }));
      toast.success(text("saved"));
    } catch (r) {
      toast.error(errorMessage(r, text("saveFailed")));
    }
  }
  return (
    <AdminPage>
      <header className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">{text("site")}</h2>
          <p className="text-xs text-muted-foreground">{text("siteHint")}</p>
        </div>
        <Button
          size="sm"
          disabled={saving || loading || uploadingAvatar || missingSelection}
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
            <section className="space-y-3 border-t pt-5 md:col-span-2">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <h3 className="text-base font-semibold">{text("avatarPoolTitle")}</h3>
                <div className="flex items-center gap-2 text-sm font-medium">
                  <label htmlFor="avatar-assignment">{text("assignAvatar")}</label>
                  <Switch id="avatar-assignment" checked={!form.disableDefaultAvatar} disabled={saving} onCheckedChange={enabled => setForm(current => ({ ...current, disableDefaultAvatar: !enabled }))} />
                </div>
              </div>
              <p className="text-sm text-muted-foreground">{text(form.disableDefaultAvatar ? "avatarOff" : "avatarRandom")}</p>
              <div className="flex flex-wrap items-center gap-3">
                {enabledAvatars.length > 0 && <div aria-hidden="true" className="flex -space-x-1.5">
                  {enabledAvatars.slice(0, 3).map(({ file }) => <img key={file} src={builtins.includes(file) ? file : `/file/img/${file}`} alt="" className="size-7 rounded-full border-2 border-background bg-muted object-cover" />)}
                </div>}
                <span className="text-sm text-muted-foreground">{text("avatarEnabledCount").replace("{count}", String(enabledAvatars.length))}</span>
                <Button type="button" variant="ghost" size="sm" aria-expanded={avatarsExpanded} aria-controls="avatar-options" onClick={() => setAvatarsExpanded(value => !value)}>
                  {text("manageAvatars")}<ChevronDown aria-hidden="true" className={avatarsExpanded ? "rotate-180" : ""} />
                </Button>
              </div>
              <div id="avatar-options" hidden={!avatarsExpanded} className="space-y-5">
                {avatarsExpanded && [true, false].map(builtin => {
                  const options = pool.filter(option => builtins.includes(option.file) === builtin);
                  const enabled = options.filter(option => option.enabled).length;
                  return <fieldset key={String(builtin)} aria-labelledby={builtin ? "builtin-avatar-title" : "custom-avatar-title"} disabled={saving} className="min-w-0 space-y-3 disabled:opacity-60">
                    <div className="flex min-h-8 flex-wrap items-center gap-2">
                      <div className="grid grid-cols-[minmax(5em,auto)_auto] items-baseline gap-x-3">
                        <h4 id={builtin ? "builtin-avatar-title" : "custom-avatar-title"} className="font-medium">{text(builtin ? "builtinAvatars" : "customAvatars")}</h4>
                        {options.length > 0 && <span className="text-sm text-muted-foreground">{text("avatarEnabled")} {enabled} / {options.length}</span>}
                      </div>
                      {builtin ? <Button type="button" size="sm" variant="outline" onClick={() => updatePool(items => items.map(item => builtins.includes(item.file) ? { ...item, enabled: enabled !== options.length } : item))}>
                        {text(enabled === options.length ? "disableAll" : "enableAll")}
                      </Button> : <label className="relative inline-flex h-8 cursor-pointer items-center gap-2 rounded-md border px-3 text-sm focus-within:ring-2">
                        {uploadingAvatar ? <Spinner /> : <Upload className="size-4" />}{text("uploadAvatars")}
                        <input className="sr-only" type="file" multiple accept="image/*" disabled={uploadingAvatar} onChange={e => { const files = Array.from(e.target.files || []); e.target.value = ""; void uploadAvatars(files); }} />
                      </label>}
                    </div>
                    {options.length > 0 && <div className="flex flex-wrap gap-3">
                      {options.map(({ file, enabled: checked }, index) => <div key={file} className="grid w-12 shrink-0 justify-items-center gap-1">
                        <button type="button" aria-pressed={checked} aria-label={text(builtin ? "builtinAvatars" : "customAvatars") + " " + (index + 1)} onClick={() => updatePool(items => items.map(item => item.file === file ? { ...item, enabled: !item.enabled } : item))} className="relative size-12 rounded-full outline-offset-2 focus-visible:outline-2 focus-visible:outline-ring">
                          <img loading="lazy" decoding="async" src={builtin ? file : `/file/img/${file}`} alt="" className={checked ? avatarImageClass : avatarImageClass + " grayscale opacity-60"} />
                          {checked && <span aria-hidden="true" className="absolute bottom-0 right-0 grid size-4 place-items-center rounded-full bg-primary text-xs text-primary-foreground">✓</span>}
                        </button>
                        {!builtin && <Button type="button" size="icon-sm" variant="ghost" title={text("removeAvatar")} aria-label={text("removeAvatar") + " " + (index + 1)} onClick={() => updatePool(items => items.filter(item => item.file !== file))}><CircleMinus aria-hidden="true" /></Button>}
                      </div>)}
                    </div>}
                  </fieldset>;
                })}
              </div>
              {missingSelection && <p role="alert" className="text-sm text-destructive">{text("selectAvatar")}</p>}
            </section>
          </div>
        </form>
      )}
    </AdminPage>
  );
}
