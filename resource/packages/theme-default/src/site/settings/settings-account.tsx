import { useState, type FormEvent } from "react";
import { KeyRound, Mail } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "@gooseforum/ui/components/button";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldSeparator,
} from "@gooseforum/ui/components/field";
import { Input } from "@gooseforum/ui/components/input";
import { GooseLink, useGooseRuntime } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";
import { SettingsSectionHeader } from "./settings-section-header";
import { useMFAEnabled } from "./use-mfa-enabled";

export function AccountSettings({
  showError,
}: {
  showError(message: string): void;
}) {
  const { t } = useTranslation("settings");
  const runtime = useGooseRuntime();
  const serverError = useServerErrorMessage();
  const [current, setCurrent] = useState("");
  const [password, setPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [saving, setSaving] = useState(false);
  const mfa = useMFAEnabled();
  const [mfaCode, setMfaCode] = useState("");
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (saving || !mfa.ready) return;
    if (password !== confirmation)
      return showError(t("validation.passwordMismatch"));
    setSaving(true);
    try {
      if (mfa.enabled) await runtime.api.users.changePassword(current, password, mfaCode);
      else await runtime.api.users.changePassword(current, password);
      setCurrent("");
      setPassword("");
      setConfirmation("");
      runtime.redirect("/login");
    } catch (reason) {
      showError(serverError(reason, t("errors.password")));
    } finally {
      setSaving(false);
    }
  }

  return (
    <section>
      <SettingsSectionHeader icon={KeyRound} title={t("account.title")} />
      <form className="max-w-xl p-4" onSubmit={(event) => void submit(event)}>
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="current-password">
              {t("account.currentPassword")}
            </FieldLabel>
            <Input
              id="current-password"
              required
              type="password"
              autoComplete="current-password"
              value={current}
              onChange={(event) => setCurrent(event.target.value)}
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="new-password">
              {t("account.newPassword")}
            </FieldLabel>
            <Input
              id="new-password"
              required
              minLength={6}
              type="password"
              autoComplete="new-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
            <FieldDescription>{t("account.passwordHint")}</FieldDescription>
          </Field>
          <Field>
            <FieldLabel htmlFor="confirm-password">
              {t("account.confirmPassword")}
            </FieldLabel>
            <Input
              id="confirm-password"
              required
              type="password"
              autoComplete="new-password"
              value={confirmation}
              onChange={(event) => setConfirmation(event.target.value)}
            />
          </Field>
          <div>
            {mfa.enabled && <Field className="mb-4">
              <FieldLabel htmlFor="password-mfa-code">{t("mfa.code")}</FieldLabel>
              <Input id="password-mfa-code" autoComplete="one-time-code" required maxLength={64} value={mfaCode} onChange={(event) => setMfaCode(event.target.value)} />
            </Field>}
            {mfa.error && <Button type="button" variant="outline" onClick={mfa.retry}>{t('mfa.retry')}</Button>}
            <Button type="submit" variant="outline" disabled={saving || !mfa.ready}>
              {saving ? t("savingShort") : t("account.changePassword")}
            </Button>
          </div>
          <FieldSeparator />
          <Field>
            <FieldLabel>{t("account.forgotPasswordTitle")}</FieldLabel>
            <FieldDescription>
              {t("account.forgotPasswordDescription")}
            </FieldDescription>
            <div>
              <Button asChild variant="outline">
                <GooseLink href="/login?mode=forgot">
                  <Mail data-icon="inline-start" />
                  {t("account.resetByEmail")}
                </GooseLink>
              </Button>
            </div>
          </Field>
        </FieldGroup>
      </form>
    </section>
  );
}
