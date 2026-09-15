import { useCallback, useEffect, useRef, useState } from "react";
import type { GooseAdminApi, SecuritySettings } from "@gooseforum/client";
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
import { MailCheck, Plus, Save, Trash2 } from "lucide-react";
import { toast } from "sonner";
import type { SystemSettingsTextKey } from "../system-settings-i18n";
type Text = (k: SystemSettingsTextKey) => string;
export function SecuritySettingsPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text: Text;
}) {
  const [form, setForm] = useState<SecuritySettings>({
    enableSignup: true,
    enableEmailVerification: false,
    allowedDomains: [],
  });
  const [domain, setDomain] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const tr = useRef(text);
  useEffect(() => {
    tr.current = text;
  }, [text]);
  const load = useCallback(async () => {
    setLoading(true);
    try {
      const v = await api.settings.security();
      setForm({
        ...v,
        allowedDomains: Array.isArray(v.allowedDomains) ? v.allowedDomains : [],
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
  function add() {
    const v = domain.trim().toLowerCase();
    if (v && !form.allowedDomains.includes(v)) {
      setForm({ ...form, allowedDomains: [...form.allowedDomains, v] });
      setDomain("");
    }
  }
  async function save() {
    setSaving(true);
    try {
      await api.settings.saveSecurity(form);
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
          <h2 className="text-lg font-semibold">{text("security")}</h2>
          <p className="text-xs text-muted-foreground">
            {text("securityHint")}
          </p>
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
      <div className="flex max-w-2xl flex-col gap-7">
        <Field orientation="horizontal">
          <div className="flex-1">
            <FieldLabel>{text("signup")}</FieldLabel>
            <FieldDescription>{text("signupHint")}</FieldDescription>
          </div>
          <Switch
            checked={form.enableSignup}
            disabled={loading}
            onCheckedChange={(enableSignup) =>
              setForm({ ...form, enableSignup })
            }
          />
        </Field>
        <Field orientation="horizontal">
          <div className="flex-1">
            <FieldLabel className="flex items-center gap-2">
              <MailCheck className="size-4" />
              {text("verification")}
            </FieldLabel>
            <FieldDescription>{text("verificationHint")}</FieldDescription>
          </div>
          <Switch
            checked={form.enableEmailVerification}
            disabled={loading}
            onCheckedChange={(enableEmailVerification) =>
              setForm({ ...form, enableEmailVerification })
            }
          />
        </Field>
        <section className="flex flex-col gap-4">
          <div>
            <h3 className="font-medium">{text("domains")}</h3>
            <p className="text-sm text-muted-foreground">
              {text("domainsHint")}
            </p>
          </div>
          <div className="flex gap-2">
            <Input
              className="max-w-sm"
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  add();
                }
              }}
              placeholder={text("domainPlaceholder")}
            />
            <Button variant="secondary" onClick={add}>
              <Plus data-icon="inline-start" />
              {text("add")}
            </Button>
          </div>
          <div className="flex flex-wrap gap-2">
            {form.allowedDomains.length ? (
              form.allowedDomains.map((v) => (
                <Badge
                  key={v}
                  variant="secondary"
                  className="gap-2 px-3 py-1.5 text-sm font-normal"
                >
                  {v}
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    title={text("remove")}
                    onClick={() =>
                      setForm({
                        ...form,
                        allowedDomains: form.allowedDomains.filter(
                          (x) => x !== v,
                        ),
                      })
                    }
                  >
                    <Trash2 />
                  </Button>
                </Badge>
              ))
            ) : (
              <span className="text-sm italic text-muted-foreground">
                {text("noDomains")}
              </span>
            )}
          </div>
        </section>
      </div>
    </main>
  );
}
