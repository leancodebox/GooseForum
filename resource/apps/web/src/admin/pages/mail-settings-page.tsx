import { useCallback, useEffect, useRef, useState } from "react";
import type { GooseAdminApi, MailSettings } from "@gooseforum/client";
import { Button } from "@gooseforum/ui/components/button";
import {
  Empty,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
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
import { Mail, Save, Send, Shield } from "lucide-react";
import { toast } from "sonner";
import type { SystemSettingsTextKey } from "../system-settings-i18n";
type Text = (k: SystemSettingsTextKey) => string;
const empty: MailSettings = {
  enableMail: true,
  smtpHost: "",
  smtpPort: 587,
  useSSL: false,
  smtpUsername: "",
  smtpPassword: "",
  fromName: "",
  fromEmail: "",
};
export function MailSettingsPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text: Text;
}) {
  const [form, setForm] = useState(empty);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testEmail, setTestEmail] = useState("");
  const textRef = useRef(text);
  useEffect(() => {
    textRef.current = text;
  }, [text]);
  const load = useCallback(async () => {
    setLoading(true);
    try {
      setForm({ ...empty, ...(await api.settings.mail()) });
    } catch (r) {
      toast.error(
        r instanceof Error ? r.message : textRef.current("loadFailed"),
      );
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
      await api.settings.saveMail(form);
      toast.success(text("saved"));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    } finally {
      setSaving(false);
    }
  }
  async function test() {
    if (!testEmail.trim()) return;
    setTesting(true);
    try {
      await api.settings.testMail(form, testEmail.trim());
      toast.success(text("testSent"));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    } finally {
      setTesting(false);
    }
  }
  return (
    <main className="flex flex-1 flex-col gap-3 px-3 py-3 lg:px-4">
      <header className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">{text("mail")}</h2>
          <p className="text-xs text-muted-foreground">{text("mailHint")}</p>
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
        <SettingEmpty text={text("loading")} />
      ) : (
        <div className="grid gap-8 lg:grid-cols-[minmax(0,2fr)_minmax(260px,1fr)]">
          <div className="flex flex-col gap-6">
            <Field
              orientation="horizontal"
              className="rounded-lg border bg-muted/20 p-4"
            >
              <div className="flex-1">
                <FieldLabel>{text("enabled")}</FieldLabel>
                <FieldDescription>{text("enabledHint")}</FieldDescription>
              </div>
              <Switch
                checked={form.enableMail}
                onCheckedChange={(enableMail) =>
                  setForm({ ...form, enableMail })
                }
              />
            </Field>
            <FieldSet>
              <FieldLegend className="border-b pb-2">
                {text("smtp")}
              </FieldLegend>
              <div className="grid gap-4 md:grid-cols-2">
                <F
                  label={text("host")}
                  value={form.smtpHost}
                  disabled={!form.enableMail}
                  onChange={(smtpHost) => setForm({ ...form, smtpHost })}
                />
                <F
                  label={text("port")}
                  value={String(form.smtpPort)}
                  type="number"
                  disabled={!form.enableMail}
                  onChange={(v) => setForm({ ...form, smtpPort: Number(v) })}
                />
                <F
                  label={text("username")}
                  value={form.smtpUsername}
                  disabled={!form.enableMail}
                  onChange={(smtpUsername) =>
                    setForm({ ...form, smtpUsername })
                  }
                />
                <F
                  label={text("password")}
                  value={form.smtpPassword}
                  type="password"
                  disabled={!form.enableMail}
                  onChange={(smtpPassword) =>
                    setForm({ ...form, smtpPassword })
                  }
                />
              </div>
            </FieldSet>
            <Field
              orientation="horizontal"
              className="rounded-lg border bg-muted/20 p-4"
            >
              <div className="flex-1">
                <FieldLabel className="flex items-center gap-2">
                  <Shield className="size-4" />
                  {text("ssl")}
                </FieldLabel>
                <FieldDescription>{text("sslHint")}</FieldDescription>
              </div>
              <Switch
                checked={form.useSSL}
                disabled={!form.enableMail}
                onCheckedChange={(useSSL) => setForm({ ...form, useSSL })}
              />
            </Field>
          </div>
          <aside className="flex flex-col gap-6">
            <FieldSet>
              <FieldLegend className="border-b pb-2">
                {text("sender")}
              </FieldLegend>
              <FieldGroup>
                <F
                  label={text("fromName")}
                  value={form.fromName}
                  disabled={!form.enableMail}
                  onChange={(fromName) => setForm({ ...form, fromName })}
                />
                <F
                  label={text("fromEmail")}
                  value={form.fromEmail}
                  disabled={!form.enableMail}
                  onChange={(fromEmail) => setForm({ ...form, fromEmail })}
                />
              </FieldGroup>
            </FieldSet>
            <section className="flex flex-col gap-3 rounded-lg border p-4">
              <h3 className="font-semibold">{text("test")}</h3>
              <p className="text-sm text-muted-foreground">
                {text("testHint")}
              </p>
              <Input
                value={testEmail}
                disabled={!form.enableMail}
                onChange={(e) => setTestEmail(e.target.value)}
                placeholder={text("testEmail")}
              />
              <Button
                variant="secondary"
                disabled={testing || !form.enableMail || !testEmail.trim()}
                onClick={() => void test()}
              >
                {testing ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <Send data-icon="inline-start" />
                )}
                {text("sendTest")}
              </Button>
            </section>
          </aside>
        </div>
      )}
    </main>
  );
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
        type={type}
        value={value}
        disabled={disabled}
        onChange={(e) => onChange(e.target.value)}
      />
    </Field>
  );
}
function SettingEmpty({ text }: { text: string }) {
  return (
    <Empty className="min-h-64 border">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Mail />
        </EmptyMedia>
        <EmptyTitle>{text}</EmptyTitle>
      </EmptyHeader>
    </Empty>
  );
}
