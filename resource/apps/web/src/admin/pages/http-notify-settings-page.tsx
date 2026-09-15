import { useCallback, useEffect, useRef, useState } from "react";
import type {
  GooseAdminApi,
  HttpNotifyEndpoint,
  HttpNotifySettings,
} from "@gooseforum/client";
import { Badge } from "@gooseforum/ui/components/badge";
import { Button } from "@gooseforum/ui/components/button";
import { Checkbox } from "@gooseforum/ui/components/checkbox";
import { Field, FieldLabel } from "@gooseforum/ui/components/field";
import { Input } from "@gooseforum/ui/components/input";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Switch } from "@gooseforum/ui/components/switch";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@gooseforum/ui/components/tabs";
import { Plus, Save, Trash2, Webhook } from "lucide-react";
import { toast } from "sonner";
import type { ModerationSettingsTextKey } from "../moderation-settings-i18n";
type Text = (k: ModerationSettingsTextKey) => string;
const events = [
  ["topic.published", "topicPublished"],
  ["topic.updated", "topicUpdated"],
  ["comment.created", "commentCreated"],
  ["user.signup", "userSignup"],
  ["moderation.report.created", "reportCreated"],
] as const;
const empty = (): HttpNotifyEndpoint => ({
  id: crypto.randomUUID(),
  name: "",
  enabled: true,
  url: "",
  secret: "",
  events: ["topic.published"],
  timeoutSeconds: 2,
  failureCount: 0,
  lastError: "",
  abnormalTerminated: false,
});
export function HttpNotifySettingsPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text: Text;
}) {
  const [form, setForm] = useState<HttpNotifySettings>({
    enabled: false,
    endpoints: [],
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
      const v = await api.settings.httpNotify();
      setForm(normalizeSettings(v));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : tr.current("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [api]);
  useEffect(() => {
    void load();
  }, [load]);
  function update(i: number, fn: (e: HttpNotifyEndpoint) => void) {
    setForm((current) => {
      const endpoints = current.endpoints.map((x) => ({
        ...x,
        events: [...x.events],
      }));
      fn(endpoints[i]);
      return { ...current, endpoints };
    });
  }
  function valid(settings: HttpNotifySettings) {
    if (!settings.enabled) return true;
    const active = settings.endpoints.filter((x) => x.enabled);
    if (!active.length) {
      toast.warning(text("needEndpoint"));
      return false;
    }
    for (const e of active) {
      try {
        const u = new URL(e.url);
        if (!["http:", "https:"].includes(u.protocol)) {
          throw new Error("Unsupported notification URL protocol");
        }
      } catch {
        toast.warning(text("invalidUrl"));
        return false;
      }
      if (!e.events.length) {
        toast.warning(text("needEvent"));
        return false;
      }
      if (e.timeoutSeconds < 1 || e.timeoutSeconds > 15) {
        toast.warning(text("invalidTimeout"));
        return false;
      }
    }
    return true;
  }
  async function save() {
    const normalized = normalizeSettings(form);
    if (!valid(normalized)) return;
    setSaving(true);
    try {
      await api.settings.saveHttpNotify(normalized);
      setForm(normalized);
      toast.success(text("saved"));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    } finally {
      setSaving(false);
    }
  }
  return (
    <main className="flex flex-1 flex-col gap-4 px-3 py-3 lg:px-4">
      <header className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">{text("http")}</h2>
          <p className="text-xs text-muted-foreground">{text("httpHint")}</p>
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
      <Tabs defaultValue="config">
        <TabsList>
          <TabsTrigger value="config">{text("config")}</TabsTrigger>
          <TabsTrigger value="guide">{text("guide")}</TabsTrigger>
        </TabsList>
        <TabsContent value="config">
          <div className="flex max-w-5xl flex-col gap-4">
            <div className="flex items-center justify-between border-b pb-4">
              <div>
                <h3 className="flex items-center gap-2 font-medium">
                  <Webhook className="size-4" />
                  {text("enabled")}
                </h3>
                <p className="text-sm text-muted-foreground">
                  {text("httpHint")}
                </p>
              </div>
              <Switch
                checked={form.enabled}
                disabled={loading}
                onCheckedChange={(enabled) => setForm({ ...form, enabled })}
              />
            </div>
            <section className="flex flex-col gap-3">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium">{text("endpoints")}</h3>
                  <p className="text-sm text-muted-foreground">
                    {text("endpointHint")}
                  </p>
                </div>
                <Button
                  variant="secondary"
                  onClick={() =>
                    setForm({
                      ...form,
                      endpoints: [...form.endpoints, empty()],
                    })
                  }
                >
                  <Plus data-icon="inline-start" />
                  {text("addEndpoint")}
                </Button>
              </div>
              {form.endpoints.length ? (
                form.endpoints.map((endpoint, i) => (
                  <article
                    key={endpoint.id}
                    className="flex flex-col gap-3 rounded-lg border p-3"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <Switch
                          checked={endpoint.enabled}
                          disabled={!form.enabled}
                          onCheckedChange={(enabled) =>
                            update(i, (e) => {
                              e.enabled = enabled;
                            })
                          }
                        />
                        <div>
                          <div className="flex items-center gap-2 font-medium">
                            {endpoint.name || endpoint.url || text("endpoints")}
                            {endpoint.abnormalTerminated ? (
                              <Badge variant="destructive">
                                {text("abnormal")}
                              </Badge>
                            ) : null}
                          </div>
                          {endpoint.lastError ? (
                            <p className="text-xs text-muted-foreground">
                              {endpoint.lastError}
                            </p>
                          ) : null}
                        </div>
                      </div>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() =>
                          setForm({
                            ...form,
                            endpoints: form.endpoints.filter((_, n) => n !== i),
                          })
                        }
                      >
                        <Trash2 />
                      </Button>
                    </div>
                    <div className="grid gap-3 md:grid-cols-[minmax(140px,220px)_minmax(0,1fr)_120px]">
                      <F
                        label={text("name")}
                        value={endpoint.name}
                        disabled={!form.enabled}
                        onChange={(v) =>
                          update(i, (e) => {
                            e.name = v;
                          })
                        }
                      />
                      <F
                        label={text("url")}
                        value={endpoint.url}
                        disabled={!form.enabled}
                        onChange={(v) =>
                          update(i, (e) => {
                            e.url = v;
                          })
                        }
                      />
                      <F
                        label={text("timeout")}
                        value={String(endpoint.timeoutSeconds)}
                        type="number"
                        disabled={!form.enabled}
                        onChange={(v) =>
                          update(i, (e) => {
                            e.timeoutSeconds = Number(v);
                          })
                        }
                      />
                    </div>
                    <F
                      label={text("secret")}
                      value={endpoint.secret}
                      type="password"
                      disabled={!form.enabled}
                      onChange={(v) =>
                        update(i, (e) => {
                          e.secret = v;
                        })
                      }
                    />
                    <Field>
                      <FieldLabel>{text("events")}</FieldLabel>
                      <div className="flex flex-wrap gap-2">
                        {events.map(([value, key]) => (
                          <label
                            key={value}
                            className="flex items-center gap-2 rounded-md border px-2.5 py-1.5 text-sm"
                          >
                            <Checkbox
                              checked={endpoint.events.includes(value)}
                              disabled={!form.enabled}
                              onCheckedChange={(checked) =>
                                update(i, (e) => {
                                  e.events = checked
                                    ? [...new Set([...e.events, value])]
                                    : e.events.filter((x) => x !== value);
                                })
                              }
                            />
                            {text(key)}
                          </label>
                        ))}
                      </div>
                    </Field>
                  </article>
                ))
              ) : (
                <div className="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
                  {text("noEndpoints")}
                </div>
              )}
            </section>
          </div>
        </TabsContent>
        <TabsContent value="guide">
          <div className="max-w-3xl rounded-lg border bg-background p-5 text-sm leading-7 text-muted-foreground">
            <h3 className="mb-2 font-semibold text-foreground">
              {text("guide")}
            </h3>
            <p>{text("guideText")}</p>
            <pre className="mt-3 overflow-auto rounded-lg bg-muted p-3 font-mono text-xs">
              POST application/json{`\n`}X-Goose-Signature: sha256=…
            </pre>
          </div>
        </TabsContent>
      </Tabs>
    </main>
  );
}
function normalize(e: HttpNotifyEndpoint): HttpNotifyEndpoint {
  const enabled = e.enabled !== false;
  return {
    ...empty(),
    ...e,
    id: e.id || crypto.randomUUID(),
    events: Array.isArray(e.events) ? e.events : [],
    timeoutSeconds: Math.min(Math.max(Number(e.timeoutSeconds || 2), 1), 15),
    enabled,
    failureCount: enabled ? 0 : Number(e.failureCount || 0),
    lastError: enabled ? "" : e.lastError || "",
    abnormalTerminated: enabled ? false : Boolean(e.abnormalTerminated),
  };
}
function normalizeSettings(value: HttpNotifySettings): HttpNotifySettings {
  return {
    enabled: Boolean(value.enabled),
    endpoints: (value.endpoints || [])
      .map(normalize)
      .filter((endpoint) => endpoint.url.trim()),
  };
}
function F({
  label,
  value,
  onChange,
  type,
  disabled,
}: {
  label: string;
  value: string;
  onChange(v: string): void;
  type?: string;
  disabled: boolean;
}) {
  return (
    <Field>
      <FieldLabel>{label}</FieldLabel>
      <Input
        value={value}
        type={type}
        disabled={disabled}
        onChange={(e) => onChange(e.target.value)}
      />
    </Field>
  );
}
