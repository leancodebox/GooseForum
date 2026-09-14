import { useCallback, useEffect, useRef, useState } from "react";
import type {
  GooseAdminApi,
  OIDCClient,
  OIDCClientAuthMethod,
  OIDCClientInput,
  OIDCProviderStatus,
} from "@gooseforum/client";
import { Badge } from "@gooseforum/react/components/ui/badge";
import { Button } from "@gooseforum/react/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@gooseforum/react/components/ui/dialog";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@gooseforum/react/components/ui/field";
import { Input } from "@gooseforum/react/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@gooseforum/react/components/ui/select";
import { Spinner } from "@gooseforum/react/components/ui/spinner";
import { Switch } from "@gooseforum/react/components/ui/switch";
import { Textarea } from "@gooseforum/react/components/ui/textarea";
import { Check, Copy, Pencil, Plus, RefreshCw, RotateCw } from "lucide-react";
import { toast } from "sonner";
import type { IdentityTextKey } from "../identity-settings-i18n";
type Text = (k: IdentityTextKey) => string;
const scopes = ["openid", "profile", "email", "offline_access"];
type Form = {
  clientId: string;
  name: string;
  redirectUris: string;
  scopes: string[];
  tokenEndpointAuthMethod: OIDCClientAuthMethod;
  requirePkce: boolean;
  public: boolean;
  enabled: boolean;
};
const blank = (): Form => ({
  clientId: "",
  name: "",
  redirectUris: "",
  scopes: ["openid", "profile", "email"],
  tokenEndpointAuthMethod: "client_secret_basic",
  requirePkce: true,
  public: false,
  enabled: true,
});
export function OIDCProviderSettingsPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text: Text;
}) {
  const [status, setStatus] = useState<OIDCProviderStatus | null>(null);
  const [clients, setClients] = useState<OIDCClient[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [editor, setEditor] = useState<Form | null>(null);
  const [rotate, setRotate] = useState<OIDCClient | null>(null);
  const [rotateSigning, setRotateSigning] = useState(false);
  const [secret, setSecret] = useState<{
    clientId: string;
    value: string;
  } | null>(null);
  const [copied, setCopied] = useState(false);
  const tr = useRef(text);
  useEffect(() => {
    tr.current = text;
  }, [text]);
  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [a, b] = await Promise.allSettled([
        api.settings.oidcStatus(),
        api.settings.oidcClients(),
      ]);
      if (a.status === "fulfilled") setStatus(a.value);
      if (b.status === "fulfilled") setClients(b.value);
      const failure =
        a.status === "rejected"
          ? a.reason
          : b.status === "rejected"
            ? b.reason
            : null;
      if (failure)
        toast.error(
          failure instanceof Error ? failure.message : tr.current("loadFailed"),
        );
    } finally {
      setLoading(false);
    }
  }, [api]);
  useEffect(() => {
    void load();
  }, [load]);
  async function toggleProvider(enabled: boolean) {
    setSaving(true);
    try {
      setStatus(await api.settings.saveOIDCStatus(enabled));
      toast.success(text("saved"));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    } finally {
      setSaving(false);
    }
  }
  async function submit() {
    if (!editor) return;
    const input = toInput(editor);
    if (!input.name) {
      toast.warning(text("nameRequired"));
      return;
    }
    if (!input.redirectUris.length) {
      toast.warning(text("redirectRequired"));
      return;
    }
    if (input.redirectUris.some((uri) => !/^https?:\/\//i.test(uri))) {
      toast.warning(text("redirectInvalid"));
      return;
    }
    setSaving(true);
    try {
      if (editor.clientId) {
        const result = await api.settings.updateOIDCClient({
          ...input,
          clientId: editor.clientId,
        });
        setClients(
          clients.map((c) => (c.clientId === result.clientId ? result : c)),
        );
      } else {
        const result = await api.settings.createOIDCClient(input);
        setClients([...clients, result.client]);
        setSecret({
          clientId: result.client.clientId,
          value: result.clientSecret || "",
        });
      }
      setEditor(null);
      toast.success(text("saved"));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    } finally {
      setSaving(false);
    }
  }
  async function rotateClient() {
    if (!rotate) return;
    setSaving(true);
    try {
      const r = await api.settings.rotateOIDCClientSecret(rotate.clientId);
      setClients(
        clients.map((c) => (c.clientId === r.client.clientId ? r.client : c)),
      );
      setSecret({ clientId: r.client.clientId, value: r.clientSecret || "" });
      setRotate(null);
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
          <h2 className="text-lg font-semibold">{text("oidc")}</h2>
          <p className="text-xs text-muted-foreground">{text("oidcHint")}</p>
        </div>
        <Button size="sm" onClick={() => setEditor(blank())}>
          <Plus data-icon="inline-start" />
          {text("create")}
        </Button>
      </header>
      <section className="flex flex-col gap-3 rounded-lg border p-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2 font-semibold">
            {text("status")}
            <Badge variant={status?.available ? "default" : "outline"}>
              {status?.available
                ? text("running")
                : status?.enabled
                  ? text("configError")
                  : text("notEnabled")}
            </Badge>
          </div>
          <p className="mt-1 font-mono text-xs text-muted-foreground">
            {status?.issuer || status?.error || text("issuerHint")}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            disabled={!status?.available || saving}
            onClick={() => setRotateSigning(true)}
          >
            <RotateCw data-icon="inline-start" />
            {text("rotateSigning")}
          </Button>
          <Switch
            checked={status?.enabled || false}
            disabled={!status || saving}
            onCheckedChange={(v) => void toggleProvider(v)}
          />
        </div>
      </section>
      <section className="overflow-hidden rounded-lg border">
        <div className="flex items-center justify-between border-b p-3">
          <div>
            <h3 className="font-semibold">{text("apps")}</h3>
            <p className="text-xs text-muted-foreground">{clients.length}</p>
          </div>
          <Button
            variant="ghost"
            size="icon-sm"
            disabled={loading}
            onClick={() => void load()}
          >
            <RefreshCw className={loading ? "animate-spin" : undefined} />
          </Button>
        </div>
        {loading && !clients.length ? (
          <div className="grid h-28 place-items-center">
            <Spinner />
          </div>
        ) : clients.length ? (
          <div className="divide-y">
            {clients.map((client) => (
              <article
                key={client.clientId}
                className="grid gap-3 p-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-center"
              >
                <div>
                  <div className="flex items-center gap-2">
                    <h4 className="font-semibold">{client.name}</h4>
                    <Badge variant="secondary">
                      {client.public ? text("public") : text("confidential")}
                    </Badge>
                    <Badge variant={client.enabled ? "default" : "outline"}>
                      {client.enabled ? text("enabled") : text("notEnabled")}
                    </Badge>
                  </div>
                  <p className="mt-1 font-mono text-xs text-muted-foreground">
                    {client.clientId}
                  </p>
                  <div className="mt-2 flex gap-1">
                    {client.scopes.map((s) => (
                      <Badge key={s} variant="outline" className="font-mono">
                        {s}
                      </Badge>
                    ))}
                  </div>
                </div>
                <div className="flex items-center gap-1">
                  <Switch
                    checked={client.enabled}
                    disabled={saving}
                    onCheckedChange={async (enabled) => {
                      setSaving(true);
                      try {
                        const r = await api.settings.updateOIDCClient({
                          ...client,
                          enabled,
                        });
                        setClients(
                          clients.map((c) =>
                            c.clientId === r.clientId ? r : c,
                          ),
                        );
                      } catch (r) {
                        toast.error(
                          r instanceof Error ? r.message : text("saveFailed"),
                        );
                      } finally {
                        setSaving(false);
                      }
                    }}
                  />
                  {!client.public ? (
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      onClick={() => setRotate(client)}
                    >
                      <RotateCw />
                    </Button>
                  ) : null}
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => setEditor(fromClient(client))}
                  >
                    <Pencil />
                  </Button>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <div className="p-8 text-center text-muted-foreground">
            {text("empty")}
          </div>
        )}
      </section>
      <ClientEditor
        form={editor}
        saving={saving}
        text={text}
        onChange={setEditor}
        onClose={() => setEditor(null)}
        onSave={() => void submit()}
      />
      <Confirm
        open={rotateSigning}
        title={text("rotateSigningTitle")}
        description={text("rotateSigningHint")}
        text={text}
        saving={saving}
        onClose={() => setRotateSigning(false)}
        onConfirm={async () => {
          setSaving(true);
          try {
            setStatus(await api.settings.rotateOIDCSigningKey());
            setRotateSigning(false);
            toast.success(text("saved"));
          } catch (r) {
            toast.error(r instanceof Error ? r.message : text("saveFailed"));
          } finally {
            setSaving(false);
          }
        }}
      />
      <Confirm
        open={Boolean(rotate)}
        title={text("rotateSecret")}
        description={text("rotateSecretHint")}
        text={text}
        saving={saving}
        onClose={() => setRotate(null)}
        onConfirm={rotateClient}
      />
      <Secret
        value={secret}
        copied={copied}
        text={text}
        onClose={() => {
          setSecret(null);
          setCopied(false);
        }}
        onCopy={async () => {
          if (secret?.value) {
            await navigator.clipboard.writeText(secret.value);
            setCopied(true);
          }
        }}
      />
    </main>
  );
}
function ClientEditor({
  form,
  saving,
  text,
  onChange,
  onClose,
  onSave,
}: {
  form: Form | null;
  saving: boolean;
  text: Text;
  onChange(v: Form): void;
  onClose(): void;
  onSave(): void;
}) {
  if (!form) return <Dialog open={false} />;
  function set<K extends keyof Form>(k: K, v: Form[K]) {
    onChange({ ...form!, [k]: v });
  }
  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-h-[calc(100vh-2rem)] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {form.clientId ? text("edit") : text("create")}
          </DialogTitle>
          <DialogDescription>{text("oidcHint")}</DialogDescription>
        </DialogHeader>
        <FieldGroup>
          <Field>
            <FieldLabel>{text("appName")}</FieldLabel>
            <Input
              autoFocus
              value={form.name}
              onChange={(e) => set("name", e.target.value)}
            />
          </Field>
          <Field>
            <FieldLabel>{text("redirects")}</FieldLabel>
            <Textarea
              className="min-h-24 font-mono text-xs"
              value={form.redirectUris}
              onChange={(e) => set("redirectUris", e.target.value)}
            />
          </Field>
          <FieldSet>
            <FieldLegend>{text("clientType")}</FieldLegend>
            <div className="grid gap-2 sm:grid-cols-2">
              <Button
                type="button"
                variant={form.public ? "outline" : "default"}
                disabled={Boolean(form.clientId)}
                onClick={() =>
                  onChange({
                    ...form,
                    public: false,
                    tokenEndpointAuthMethod:
                      form.tokenEndpointAuthMethod === "none"
                        ? "client_secret_basic"
                        : form.tokenEndpointAuthMethod,
                  })
                }
              >
                {text("confidential")}
              </Button>
              <Button
                type="button"
                variant={form.public ? "default" : "outline"}
                disabled={Boolean(form.clientId)}
                onClick={() =>
                  onChange({
                    ...form,
                    public: true,
                    tokenEndpointAuthMethod: "none",
                    requirePkce: true,
                  })
                }
              >
                {text("public")}
              </Button>
            </div>
          </FieldSet>
          <div className="grid gap-3 sm:grid-cols-2">
            <Field>
              <FieldLabel>{text("tokenAuth")}</FieldLabel>
              <Select
                value={form.tokenEndpointAuthMethod}
                disabled={form.public}
                onValueChange={(v) =>
                  set("tokenEndpointAuthMethod", v as OIDCClientAuthMethod)
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value="client_secret_basic">
                      client_secret_basic
                    </SelectItem>
                    <SelectItem value="client_secret_post">
                      client_secret_post
                    </SelectItem>
                    {form.public ? (
                      <SelectItem value="none">none</SelectItem>
                    ) : null}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
            <Field orientation="horizontal">
              <div className="flex-1">
                <FieldLabel>{text("pkce")}</FieldLabel>
                <FieldDescription>{text("pkceHint")}</FieldDescription>
              </div>
              <Switch
                checked={form.public || form.requirePkce}
                disabled={form.public}
                onCheckedChange={(v) => set("requirePkce", v)}
              />
            </Field>
          </div>
          <FieldSet>
            <FieldLegend>{text("scopes")}</FieldLegend>
            <div className="grid gap-2 sm:grid-cols-2">
              {scopes.map((scope) => (
                <Field key={scope} orientation="horizontal">
                  <FieldLabel className="font-mono">{scope}</FieldLabel>
                  <Switch
                    checked={form.scopes.includes(scope)}
                    disabled={scope === "openid"}
                    onCheckedChange={(v) =>
                      set(
                        "scopes",
                        v
                          ? [...new Set([...form.scopes, scope])]
                          : form.scopes.filter((x) => x !== scope),
                      )
                    }
                  />
                </Field>
              ))}
            </div>
          </FieldSet>
          <Field orientation="horizontal">
            <FieldLabel>{text("clientEnabled")}</FieldLabel>
            <Switch
              checked={form.enabled}
              onCheckedChange={(v) => set("enabled", v)}
            />
          </Field>
        </FieldGroup>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {text("cancel")}
          </Button>
          <Button disabled={saving} onClick={onSave}>
            {saving ? <Spinner data-icon="inline-start" /> : null}
            {text("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
function Confirm({
  open,
  title,
  description,
  text,
  saving,
  onClose,
  onConfirm,
}: {
  open: boolean;
  title: string;
  description: string;
  text: Text;
  saving: boolean;
  onClose(): void;
  onConfirm(): void | Promise<void>;
}) {
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {text("cancel")}
          </Button>
          <Button disabled={saving} onClick={() => void onConfirm()}>
            {text("confirm")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
function Secret({
  value,
  copied,
  text,
  onClose,
  onCopy,
}: {
  value: { clientId: string; value: string } | null;
  copied: boolean;
  text: Text;
  onClose(): void;
  onCopy(): void;
}) {
  return (
    <Dialog open={Boolean(value)} onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{text("secretTitle")}</DialogTitle>
          <DialogDescription>{text("secretHint")}</DialogDescription>
        </DialogHeader>
        <code className="rounded-md border bg-muted p-3 text-xs">
          {value?.clientId}
        </code>
        <div className="flex gap-2">
          <Input readOnly value={value?.value || ""} className="font-mono" />
          <Button variant="outline" size="icon" onClick={onCopy}>
            {copied ? <Check /> : <Copy />}
          </Button>
        </div>
        <DialogFooter>
          <Button onClick={onClose}>{text("stored")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
function fromClient(c: OIDCClient): Form {
  return { ...c, redirectUris: c.redirectUris.join("\n") };
}
function toInput(f: Form): OIDCClientInput {
  const scopes = [...new Set(["openid", ...f.scopes])];
  return {
    name: f.name.trim(),
    redirectUris: [
      ...new Set(
        f.redirectUris
          .split(/\r?\n/)
          .map((x) => x.trim())
          .filter(Boolean),
      ),
    ],
    scopes,
    grantTypes: scopes.includes("offline_access")
      ? ["authorization_code", "refresh_token"]
      : ["authorization_code"],
    tokenEndpointAuthMethod: f.public ? "none" : f.tokenEndpointAuthMethod,
    requirePkce: f.public || f.requirePkce,
    public: f.public,
    enabled: f.enabled,
  };
}
