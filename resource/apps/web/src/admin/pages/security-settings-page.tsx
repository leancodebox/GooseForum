import { errorMessage } from "../error-message";
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
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from "@gooseforum/ui/components/select";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Switch } from "@gooseforum/ui/components/switch";
import { MailCheck, Plus, Save, Trash2 } from "lucide-react";
import { toast } from "@gooseforum/ui/components/sonner";
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
    inviteOnly: false,
    enableEmailVerification: false,
    allowedDomains: [],
    registrationIPLimit: 5,
    registrationEmailLimit: 3,
    registrationGlobalLimit: 100,
  });
  const [savedCaptcha, setSavedCaptcha] = useState<SecuritySettings["registrationCaptcha"]>();
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
      setSavedCaptcha(v.registrationCaptcha);
      setForm({
        ...v,
        inviteOnly: v.inviteOnly ?? false,
        registrationIPLimit: v.registrationIPLimit ?? 5,
        registrationEmailLimit: v.registrationEmailLimit ?? 3,
        registrationGlobalLimit: v.registrationGlobalLimit ?? 100,
        allowedDomains: Array.isArray(v.allowedDomains) ? v.allowedDomains : [],
      });
    } catch (r) {
      toast.error(errorMessage(r, tr.current("loadFailed")));
    } finally {
      setLoading(false);
    }
  }, [api]);
  useEffect(() => {
    void load();
  }, [load]);
  const captcha = form.registrationCaptcha ?? { provider: "image" };
  const hosted = captcha.provider !== "image" && captcha.provider !== "none" && captcha.provider !== "altcha";
  const secretConfigured = hosted && savedCaptcha?.provider === captcha.provider && savedCaptcha.siteKey === captcha.siteKey?.trim();
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
      const settings: SecuritySettings = captcha.provider === "altcha"
        ? { ...form, registrationCaptcha: { provider: captcha.provider, difficulty: captcha.difficulty ?? "medium" } }
        : form;
      await api.settings.saveSecurity(settings);
      setSavedCaptcha({ ...captcha, siteKey: captcha.siteKey?.trim(), secret: undefined });
      if (captcha.secret) {
        setForm((current) => current.registrationCaptcha === captcha
          ? { ...current, registrationCaptcha: { ...captcha, secret: "" } }
          : current);
      }
      toast.success(text("saved"));
    } catch (r) {
      toast.error(errorMessage(r, text("saveFailed")));
    } finally {
      setSaving(false);
    }
  }
  return (
    <main className="flex flex-1 flex-col gap-6 px-3 py-3 lg:px-4">

      <div className="flex max-w-2xl flex-col gap-7">
        <Field orientation="horizontal">
          <div className="flex-1">
            <FieldLabel htmlFor="enable-signup">{text("signup")}</FieldLabel>
            <FieldDescription>{text("signupHint")}</FieldDescription>
          </div>
          <Switch
            id="enable-signup"
            checked={form.enableSignup}
            disabled={loading}
            onCheckedChange={(enableSignup) =>
              setForm({ ...form, enableSignup })
            }
          />
        </Field>
        <Field orientation="horizontal">
          <div className="flex-1">
            <FieldLabel htmlFor="invite-only">{text("inviteOnly")}</FieldLabel>
            <FieldDescription>{text("inviteOnlyHint")}</FieldDescription>
          </div>
          <Switch id="invite-only" checked={Boolean(form.inviteOnly)} disabled={loading} onCheckedChange={(inviteOnly) => setForm({ ...form, inviteOnly })} />
        </Field>
        <Field orientation="horizontal">
          <div className="flex-1">
            <FieldLabel htmlFor="enable-verification" className="flex items-center gap-2">
              <MailCheck className="size-4" />
              {text("verification")}
            </FieldLabel>
            <FieldDescription>{text("verificationHint")}</FieldDescription>
          </div>
          <Switch
            id="enable-verification"
            checked={form.enableEmailVerification}
            disabled={loading}
            onCheckedChange={(enableEmailVerification) =>
              setForm({ ...form, enableEmailVerification })
            }
          />
        </Field>
        <section aria-labelledby="captcha-heading" className="flex flex-col gap-4 border-y py-5">
          <h3 id="captcha-heading" className="font-medium">
            <label htmlFor="registration-captcha">{text("registrationCaptcha")}</label>
          </h3>
          <Select value={captcha.provider} disabled={loading || saving} onValueChange={(value) => {
            const provider = value as typeof captcha.provider;
            setForm({ ...form, registrationCaptcha: savedCaptcha?.provider === provider ? savedCaptcha : { provider } });
          }}>
            <SelectTrigger id="registration-captcha" className="w-full sm:max-w-sm"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="none">{text("captchaNone")}</SelectItem>
              <SelectGroup>
                <SelectLabel>{text("captchaBuiltIn")}</SelectLabel>
                <SelectItem value="image">{text("captchaImage")}</SelectItem>
                <SelectItem value="altcha">{text("captchaAltcha")}</SelectItem>
              </SelectGroup>
              <SelectGroup>
                <SelectLabel>{text("captchaThirdParty")}</SelectLabel>
                <SelectItem value="hcaptcha">hCaptcha</SelectItem>
                <SelectItem value="recaptcha_v2">{text("captchaRecaptchaV2")}</SelectItem>
                <SelectItem value="recaptcha_v3">{text("captchaRecaptchaV3")}</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
          {captcha.provider === "altcha" && (
            <Field className="border-s-2 ps-4">
              <FieldLabel htmlFor="altcha-difficulty">{text("captchaDifficulty")}</FieldLabel>
              <Select value={captcha.difficulty ?? "medium"} disabled={loading || saving} onValueChange={(difficulty) => {
                if (difficulty === "low" || difficulty === "medium" || difficulty === "high") {
                  setForm({ ...form, registrationCaptcha: { ...captcha, difficulty } });
                }
              }}>
                <SelectTrigger id="altcha-difficulty" aria-describedby="altcha-difficulty-hint" className="w-full sm:max-w-sm"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">{text("captchaDifficultyLow")}</SelectItem>
                  <SelectItem value="medium">{text("captchaDifficultyMedium")}</SelectItem>
                  <SelectItem value="high">{text("captchaDifficultyHigh")}</SelectItem>
                </SelectContent>
              </Select>
              <FieldDescription id="altcha-difficulty-hint">{text("captchaDifficultyHint")}</FieldDescription>
            </Field>
          )}
          {hosted && (
            <fieldset disabled={loading || saving} aria-labelledby="registration-captcha" className="grid min-w-0 gap-4 border-s-2 ps-4">
              <Field>
                <FieldLabel htmlFor="captcha-site-key">{text("captchaSiteKey")}</FieldLabel>
                <Input id="captcha-site-key" value={captcha.siteKey ?? ""} onChange={(e) =>
                  setForm({ ...form, registrationCaptcha: { ...captcha, siteKey: e.target.value } })} />
              </Field>
              <Field>
                <FieldLabel htmlFor="captcha-secret">{text("captchaSecretKey")}</FieldLabel>
                <Input type="password" autoComplete="new-password" id="captcha-secret" aria-describedby="captcha-secret-hint"
                  value={captcha.secret ?? ""} onChange={(e) =>
                    setForm({ ...form, registrationCaptcha: { ...captcha, secret: e.target.value } })} />
                <FieldDescription id="captcha-secret-hint">{text(secretConfigured ? "captchaSecretHint" : "captchaSecretNew")}</FieldDescription>
              </Field>
              {captcha.provider === "recaptcha_v3" && (
                <Field>
                  <FieldLabel htmlFor="captcha-score">{text("captchaScore")}</FieldLabel>
                  <Input type="number" id="captcha-score" className="w-28" min={0} max={1} step={0.1} value={captcha.scoreThreshold ?? 0.5}
                    onChange={(e) => setForm({ ...form, registrationCaptcha: { ...captcha, scoreThreshold: Number(e.target.value) } })} />
                </Field>
              )}
            </fieldset>
          )}
        </section>
        <section className="flex flex-col gap-4">
          <div>
            <h3 className="font-medium">{text("domains")}</h3>
            <p className="text-sm text-muted-foreground">
              {text("domainsHint")}
            </p>
          </div>
          <div className="flex gap-2">
            <Input
              aria-label={text("domains")}
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
        <section className="flex flex-col gap-4 border-t pt-5">
          <div>
            <h3 className="font-medium">{text("registrationLimits")}</h3>
            <p className="text-sm text-muted-foreground">{text("registrationLimitsHint")}</p>
          </div>
          {([
            ["registrationIPLimit", "registrationIPLimit"],
            ["registrationEmailLimit", "registrationEmailLimit"],
            ["registrationGlobalLimit", "registrationGlobalLimit"],
          ] as const).map(([key, label]) => (
            <Field key={key} orientation="horizontal">
              <FieldLabel htmlFor={key} className="flex-1">{text(label)}</FieldLabel>
              <Input id={key} type="number" min={0} max={100000} step={1} className="w-28 shrink-0" disabled={loading}
                value={form[key] ?? 0} onChange={(event) => setForm({ ...form, [key]: Math.max(0, Math.trunc(Number(event.target.value))) })} />
            </Field>
          ))}
        </section>
      </div>
      <footer className="flex justify-end border-t pt-4">
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
      </footer>
    </main>
  );
}
