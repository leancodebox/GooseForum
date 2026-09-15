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
import { SettingsSectionHeader } from "./settings-section-header";

export function AccountSettings({
  showStatus,
  showError,
}: {
  showStatus(message: string): void;
  showError(message: string): void;
}) {
  const { t } = useTranslation("settings");
  const runtime = useGooseRuntime();
  const [current, setCurrent] = useState("");
  const [password, setPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [saving, setSaving] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (password !== confirmation)
      return showError(t("validation.passwordMismatch"));
    setSaving(true);
    try {
      await runtime.api.users.changePassword(current, password);
      setCurrent("");
      setPassword("");
      setConfirmation("");
      showStatus(t("status.passwordChanged"));
    } catch (reason) {
      showError(
        reason instanceof Error ? reason.message : t("errors.password"),
      );
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
            <Button type="submit" disabled={saving}>
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
