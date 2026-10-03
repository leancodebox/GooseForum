import { useEffect, useState } from "react";
import type { AuthSessionPayload } from "@gooseforum/client";
import { Monitor, Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "@gooseforum/ui/components/button";
import { useGooseRuntime } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";
import { SettingsSectionHeader } from "./settings-section-header";

export function SessionSettings({
  showStatus,
  showError,
}: {
  showStatus(message: string): void;
  showError(message: string): void;
}) {
  const { t } = useTranslation("settings");
  const runtime = useGooseRuntime();
  const serverError = useServerErrorMessage();
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

  return (
    <section>
      <SettingsSectionHeader icon={Monitor} title={t("account.sessionsTitle")} />
      <div className="p-4">
        {sessions.some((item) => !item.current) && (
          <div className="mb-3 flex justify-end">
            <Button type="button" variant="outline" size="sm" disabled={revoking !== null} onClick={() => void revoke("others")}>
              {t("account.revokeOthers")}
            </Button>
          </div>
        )}
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
