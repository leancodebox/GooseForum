import { useEffect, useState, type FormEvent } from "react";
import type { AuthSessionPayload } from "@gooseforum/client";
import { KeyRound, Mail, Monitor, Trash2 } from "lucide-react";
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

export function AccountSettings({
  showStatus,
  showError,
}: {
  showStatus(message: string): void;
  showError(message: string): void;
}) {
  const { t } = useTranslation("settings");
  const runtime = useGooseRuntime();
  const serverError = useServerErrorMessage();
  const [current, setCurrent] = useState("");
  const [password, setPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [saving, setSaving] = useState(false);
  const [sessions, setSessions] = useState<AuthSessionPayload[]>([]);
  const [loadingSessions, setLoadingSessions] = useState(true);
  const [revoking, setRevoking] = useState<number | "others" | null>(null);

  useEffect(() => {
    let active = true;
    void runtime.api.users.authSessions().then((items) => {
      if (active) setSessions(items);
    }).catch((reason) => {
      if (active) showError(serverError(reason, t("account.sessionsLoadFailed")));
    }).finally(() => { if (active) setLoadingSessions(false); });
    return () => { active = false; };
  }, [runtime.api.users, serverError, showError, t]);

  async function revoke(id: number | "others") {
    setRevoking(id);
    try {
      if (id === "others") await runtime.api.users.revokeOtherAuthSessions();
      else await runtime.api.users.revokeAuthSession(id);
      setSessions((items) => items.filter((item) => item.current || (id !== "others" && item.id !== id)));
      showStatus(t("account.sessionRevoked"));
    } catch (reason) {
      showError(serverError(reason, t("account.sessionRevokeFailed")));
    } finally { setRevoking(null); }
  }

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
      <div className="border-t p-4">
        <div className="mb-3 flex items-center justify-between gap-3">
          <h3 className="flex items-center gap-2 text-sm font-semibold"><Monitor className="size-4" />{t("account.sessionsTitle")}</h3>
          {sessions.some((item) => !item.current) && <Button type="button" variant="outline" size="sm" disabled={revoking !== null} onClick={() => void revoke("others")}>{t("account.revokeOthers")}</Button>}
        </div>
        {loadingSessions ? <p className="text-sm text-muted-foreground">{t("account.sessionsLoading")}</p> : sessions.length === 0 ? <p className="text-sm text-muted-foreground">{t("account.sessionsEmpty")}</p> : (
          <div className="divide-y border-y">
            {sessions.map((item) => <div key={item.id} className="flex min-w-0 items-center justify-between gap-3 py-3">
              <div className="min-w-0 text-sm">
                <p className="truncate font-medium">{item.userAgent || t("account.unknownDevice")}{item.current && <span className="ml-2 text-primary">{t("account.currentSession")}</span>}</p>
                <p className="text-xs text-muted-foreground">{item.clientIp || t("account.unknownIP")} · {item.oauthProvider || item.authMethod} · {new Date(item.lastSeenAt).toLocaleString(runtime.locale)}</p>
              </div>
              {!item.current && <Button type="button" variant="ghost" size="icon" title={t("account.revokeSession")} aria-label={t("account.revokeSession")} disabled={revoking !== null} onClick={() => void revoke(item.id)}><Trash2 className="size-4" /></Button>}
            </div>)}
          </div>
        )}
      </div>
    </section>
  );
}
