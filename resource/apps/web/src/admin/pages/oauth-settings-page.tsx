import { useCallback, useEffect, useRef, useState } from "react";
import type { GooseAdminApi, OAuthProviderSettings } from "@gooseforum/client";
import { Badge } from "@gooseforum/ui/components/badge";
import { Button } from "@gooseforum/ui/components/button";
import {
  Field,
  FieldDescription,
  FieldLabel,
} from "@gooseforum/ui/components/field";
import { Input } from "@gooseforum/ui/components/input";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Switch } from "@gooseforum/ui/components/switch";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@gooseforum/ui/components/tabs";
import { KeyRound, Plus, Save, Trash2, X } from "lucide-react";
import { toast } from "sonner";
import type { IdentityTextKey } from "../identity-settings-i18n";
type Text = (k: IdentityTextKey) => string;
type Form = OAuthProviderSettings & { scopeDraft: string };
function callbackUrl(siteUrl: string, key: string) {
  const provider = key.trim().toLowerCase();
  return siteUrl.trim() && provider
    ? `${siteUrl.trim().replace(/\/+$/, "")}/api/auth/${encodeURIComponent(provider)}/callback`
    : "";
}
function toForm(p: OAuthProviderSettings): Form {
  return {
    ...p,
    clientSecret: "",
    clearClientSecret: false,
    scopes: p.scopes?.length ? [...p.scopes] : ["openid", "profile", "email"],
    scopeDraft: "",
  };
}
export function OAuthSettingsPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text: Text;
}) {
  const [items, setItems] = useState<Form[]>([]);
  const [siteUrl, setSiteUrl] = useState("");
  const [active, setActive] = useState("0");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const tr = useRef(text);
  useEffect(() => {
    tr.current = text;
  }, [text]);
  const load = useCallback(async () => {
    setLoading(true);
    try {
      const settings = await api.settings.oauth();
      setSiteUrl(settings.siteUrl);
      setItems(settings.providers.map(toForm));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : tr.current("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [api]);
  useEffect(() => {
    void load();
  }, [load]);
  const index = Math.min(Number(active) || 0, Math.max(items.length - 1, 0));
  const current = items[index];
  function update(fn: (p: Form) => void) {
    setItems((list) =>
      list.map((p, i) => {
        if (i !== index) return p;
        const next = { ...p, scopes: [...(p.scopes || [])] };
        fn(next);
        return next;
      }),
    );
  }
  function add() {
    setItems((list) => [
      ...list,
      {
        key: "",
        displayName: "",
        kind: "oidc",
        enabled: false,
        clientId: "",
        clientSecret: "",
        clientSecretConfigured: false,
        clearClientSecret: false,
        callbackUrl: "",
        discoveryUrl: "",
        scopes: ["openid", "profile", "email"],
        scopeDraft: "",
      },
    ]);
    setActive(String(items.length));
  }
  async function save() {
    setSaving(true);
    try {
      const result = await api.settings.saveOAuth({
        providers: items.map((p) => ({
          key: p.key.trim().toLowerCase(),
          displayName: p.displayName.trim(),
          kind: p.kind,
          enabled: p.enabled,
          clientId: p.clientId.trim(),
          clientSecret: p.clientSecret?.trim() || "",
          clientSecretConfigured: p.clientSecretConfigured,
          clearClientSecret: Boolean(p.clearClientSecret),
          callbackUrl: callbackUrl(siteUrl, p.key),
          discoveryUrl: p.discoveryUrl?.trim() || "",
          scopes: p.scopes || [],
        })),
      });
      setSiteUrl(result.siteUrl);
      setItems(result.providers.map(toForm));
      setActive(String(Math.min(index, result.providers.length - 1)));
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
          <h2 className="text-lg font-semibold">{text("oauth")}</h2>
          <p className="text-xs text-muted-foreground">{text("oauthHint")}</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={add}>
            <Plus data-icon="inline-start" />
            {text("addProvider")}
          </Button>
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
        </div>
      </header>
      {loading ? (
        <div className="grid min-h-64 place-items-center">
          <Spinner />
        </div>
      ) : current ? (
        <Tabs value={String(index)} onValueChange={setActive}>
          <TabsList>
            {items.map((p, i) => (
              <TabsTrigger key={`${p.kind}-${p.key}-${i}`} value={String(i)}>
                <KeyRound />
                {p.displayName || p.key || text("provider")}
                <span
                  className={`size-1.5 rounded-full ${p.enabled ? "bg-success" : "bg-muted-foreground/35"}`}
                />
              </TabsTrigger>
            ))}
          </TabsList>
          <TabsContent value={String(index)} className="flex flex-col gap-6">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="font-semibold">
                  {current.displayName || current.key || text("provider")}
                </h3>
                <p className="text-xs text-muted-foreground">
                  {current.kind === "oidc" ? text("custom") : text("builtin")} ·{" "}
                  {current.clientSecretConfigured
                    ? text("secretConfigured")
                    : text("secretMissing")}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <Field orientation="horizontal" className="w-auto">
                  <FieldLabel>{text("enabled")}</FieldLabel>
                  <Switch
                    checked={current.enabled}
                    onCheckedChange={(enabled) =>
                      update((p) => {
                        p.enabled = enabled;
                      })
                    }
                  />
                </Field>
                {current.kind === "oidc" ? (
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => {
                      setItems(items.filter((_, i) => i !== index));
                      setActive(String(Math.max(0, index - 1)));
                    }}
                  >
                    <Trash2 />
                  </Button>
                ) : null}
              </div>
            </div>
            {current.kind === "oidc" ? (
              <div className="grid gap-4 md:grid-cols-2">
                <F
                  label={text("displayName")}
                  value={current.displayName}
                  onChange={(displayName) =>
                    update((p) => {
                      p.displayName = displayName;
                    })
                  }
                />
                <F
                  label={text("key")}
                  value={current.key}
                  onChange={(key) =>
                    update((p) => {
                      p.key = key;
                    })
                  }
                />
              </div>
            ) : null}
            <div className="grid gap-4 md:grid-cols-2">
              <F
                label={text("clientId")}
                value={current.clientId}
                onChange={(clientId) =>
                  update((p) => {
                    p.clientId = clientId;
                  })
                }
              />
              <Field>
                <FieldLabel>{text("clientSecret")}</FieldLabel>
                <div className="flex gap-2">
                  <Input
                    type="password"
                    autoComplete="new-password"
                    value={current.clientSecret || ""}
                    disabled={current.clearClientSecret}
                    placeholder={
                      current.clientSecretConfigured ? text("keepSecret") : ""
                    }
                    onChange={(e) =>
                      update((p) => {
                        p.clientSecret = e.target.value;
                      })
                    }
                  />
                  {current.clientSecretConfigured ? (
                    <Button
                      variant="outline"
                      onClick={() =>
                        update((p) => {
                          p.clearClientSecret = !p.clearClientSecret;
                        })
                      }
                    >
                      {current.clearClientSecret
                        ? text("undoClear")
                        : text("clearSecret")}
                    </Button>
                  ) : null}
                </div>
                {current.clearClientSecret ? (
                  <FieldDescription className="text-destructive">
                    {text("clearWarning")}
                  </FieldDescription>
                ) : null}
              </Field>
            </div>
            {current.kind === "oidc" ? (
              <>
                <F
                  label={text("discovery")}
                  value={current.discoveryUrl || ""}
                  onChange={(discoveryUrl) =>
                    update((p) => {
                      p.discoveryUrl = discoveryUrl;
                    })
                  }
                />
                <Field>
                  <FieldLabel>{text("scopes")}</FieldLabel>
                  <div className="flex gap-2">
                    <Input
                      value={current.scopeDraft}
                      onChange={(e) =>
                        update((p) => {
                          p.scopeDraft = e.target.value;
                        })
                      }
                      onKeyDown={(e) => {
                        if (e.key === "Enter") {
                          e.preventDefault();
                          const v = current.scopeDraft.trim();
                          if (v && !current.scopes?.includes(v))
                            update((p) => {
                              p.scopes = [...(p.scopes || []), v];
                              p.scopeDraft = "";
                            });
                        }
                      }}
                    />
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() => {
                        const v = current.scopeDraft.trim();
                        if (v)
                          update((p) => {
                            p.scopes = [...new Set([...(p.scopes || []), v])];
                            p.scopeDraft = "";
                          });
                      }}
                    >
                      <Plus />
                    </Button>
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {current.scopes?.map((s) => (
                      <Badge
                        key={s}
                        variant="secondary"
                        className="gap-1 font-mono"
                      >
                        {s}
                        {s !== "openid" ? (
                          <Button
                            variant="ghost"
                            size="icon-xs"
                            onClick={() =>
                              update((p) => {
                                p.scopes = p.scopes?.filter((x) => x !== s);
                              })
                            }
                          >
                            <X />
                          </Button>
                        ) : null}
                      </Badge>
                    ))}
                  </div>
                </Field>
              </>
            ) : null}
            <Field>
              <FieldLabel htmlFor="oauth-callback">{text("callback")}</FieldLabel>
              <Input
                id="oauth-callback"
                readOnly
                value={callbackUrl(siteUrl, current.key)}
                className="font-mono text-xs"
              />
            </Field>
          </TabsContent>
        </Tabs>
      ) : null}
    </main>
  );
}
function F({
  label,
  value,
  onChange,
}: {
  label: string;
  value: string;
  onChange(v: string): void;
}) {
  return (
    <Field>
      <FieldLabel>{label}</FieldLabel>
      <Input aria-label={label} value={value} onChange={(e) => onChange(e.target.value)} />
    </Field>
  );
}
