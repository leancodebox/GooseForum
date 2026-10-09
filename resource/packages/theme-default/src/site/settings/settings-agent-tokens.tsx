import { useEffect, useRef, useState } from "react";
import type { AgentTokenPayload } from "@gooseforum/client";
import { Ban, Check, Copy, KeyRound, Plus, RefreshCw, Trash2, X } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useGooseRuntime } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";
import { Alert, AlertDescription } from "@gooseforum/ui/components/alert";
import { Badge } from "@gooseforum/ui/components/badge";
import { Button } from "@gooseforum/ui/components/button";
import { Checkbox } from "@gooseforum/ui/components/checkbox";
import {
  Field,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@gooseforum/ui/components/field";
import { Input } from "@gooseforum/ui/components/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@gooseforum/ui/components/select";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { SettingsSectionHeader } from "./settings-section-header";
import { useMFAEnabled } from "./use-mfa-enabled";

export function AgentTokenSettings() {
  const { t } = useTranslation("settings");
  const runtime = useGooseRuntime();
  const serverError = useServerErrorMessage();
  const [rows, setRows] = useState<AgentTokenPayload[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const inFlight = useRef(false);
  const [error, setError] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [days, setDays] = useState("30");
  const [scopes, setScopes] = useState<string[]>(["forum:read"]);
  const [secret, setSecret] = useState("");
  const [copied, setCopied] = useState(false);
  const [confirm, setConfirm] = useState("");
  const [deleteID, setDeleteID] = useState("");
  const mfa = useMFAEnabled();

  useEffect(() => {
    let active = true;
    void runtime.api.users
      .agentTokens()
      .then((items) => {
        if (active) setRows(items);
      })
      .catch((reason) => {
        if (active) setError(serverError(reason, t("agentTokens.failed")));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [runtime.api.users, serverError, t]);

  async function run(action: () => Promise<void>) {
    if (inFlight.current) return;
    inFlight.current = true;
    setBusy(true);
    setError("");
    try {
      await action();
    } catch (reason) {
      setError(serverError(reason, t("agentTokens.failed")));
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  async function create(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (
      !mfa.ready ||
      mfa.error ||
      !name.trim() ||
      !password ||
      (mfa.enabled && !code.trim()) ||
      secret
    )
      return;
    await run(async () => {
      try {
        const result = await runtime.api.users.createAgentToken({
          name: name.trim(),
          scopes,
          days: Number(days),
          password,
          mfaCode: code,
        });
        setRows((items) => [result.entry, ...items]);
        setSecret(result.token);
        setCopied(false);
        setName("");
      } finally {
        setPassword("");
        setCode("");
      }
    });
  }

  return (
    <section>
      <SettingsSectionHeader
        icon={KeyRound}
        title={t("agentTokens.title")}
        actions={
          <div className="flex gap-2">
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              title={t("binding.refresh")}
              aria-label={t("binding.refresh")}
              disabled={busy}
              onClick={() =>
                void run(async () => {
                  setRows(await runtime.api.users.agentTokens());
                  setLoading(false);
                })
              }
            >
              <RefreshCw />
            </Button>
            <Button
              type="button"
              variant="outline"
              disabled={busy || !rows.some((row) => !row.revokedAt)}
              onClick={() => { setDeleteID(""); setConfirm("all"); }}
            >
              <Ban data-icon="inline-start" />
              {t("agentTokens.revokeAll")}
            </Button>
          </div>
        }
      />
      <div className="flex flex-col gap-4 p-4">
        {error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}
        {mfa.error ? (
          <Alert>
            <AlertDescription>
              <Button type="button" variant="outline" onClick={mfa.retry}>
                <RefreshCw data-icon="inline-start" />
                {t("mfa.retry")}
              </Button>
            </AlertDescription>
          </Alert>
        ) : null}
        {secret ? (
          <Alert>
            <AlertDescription className="flex flex-col gap-3">
              <span>{t("agentTokens.secretOnce")}</span>
              <Input
                readOnly
                value={secret}
                aria-label={t("agentTokens.secret")}
                className="font-mono"
              />
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() =>
                    void run(async () => {
                      await navigator.clipboard.writeText(secret);
                      setCopied(true);
                    })
                  }
                >
                  {copied ? (
                    <Check data-icon="inline-start" />
                  ) : (
                    <Copy data-icon="inline-start" />
                  )}
                  {t(copied ? "agentTokens.copied" : "agentTokens.copy")}
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  onClick={() => {
                    setSecret("");
                    setCopied(false);
                  }}
                >
                  <X data-icon="inline-start" />
                  {t("agentTokens.dismiss")}
                </Button>
              </div>
            </AlertDescription>
          </Alert>
        ) : null}
        <form onSubmit={(event) => void create(event)}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="agent-token-name">
                {t("agentTokens.name")}
              </FieldLabel>
              <Input
                id="agent-token-name"
                value={name}
                maxLength={100}
                required
                disabled={busy || Boolean(secret)}
                onChange={(event) => setName(event.target.value)}
              />
            </Field>
            <FieldSet>
              <FieldLegend>{t("agentTokens.permissions")}</FieldLegend>
              <FieldGroup>
                {["forum:read", "topics:create", "posts:create", "images:upload"].map(
                  (scope) => (
                    <Field key={scope} orientation="horizontal">
                      <Checkbox
                        id={`agent-${scope}`}
                        checked={scopes.includes(scope)}
                        disabled={scope === "forum:read" || busy}
                        onCheckedChange={(checked) =>
                          setScopes((items) =>
                            checked
                              ? [...new Set([...items, scope])]
                              : items.filter((item) => item !== scope),
                          )
                        }
                      />
                      <FieldLabel htmlFor={`agent-${scope}`}>
                        {t(`agentTokens.scopes.${scope}`, {
                          nsSeparator: false,
                        })}
                      </FieldLabel>
                    </Field>
                  ),
                )}
              </FieldGroup>
            </FieldSet>
            <Field>
              <FieldLabel>{t("agentTokens.expiry")}</FieldLabel>
              <Select
                value={days}
                onValueChange={(value) => {
                  if (value) setDays(value);
                }}
                disabled={busy}
              >
                <SelectTrigger aria-label={t("agentTokens.expiry")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {[7, 30, 90].map((value) => (
                      <SelectItem key={value} value={String(value)}>
                        {t("agentTokens.days", { count: value })}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel htmlFor="agent-token-password">
                {t("agentTokens.password")}
              </FieldLabel>
              <Input
                id="agent-token-password"
                type="password"
                autoComplete="current-password"
                required
                value={password}
                disabled={busy || Boolean(secret)}
                onChange={(event) => setPassword(event.target.value)}
              />
            </Field>
            {mfa.enabled ? (
              <Field>
                <FieldLabel htmlFor="agent-token-code">
                  {t("mfa.code")}
                </FieldLabel>
                <Input
                  id="agent-token-code"
                  autoComplete="one-time-code"
                  required
                  maxLength={64}
                  value={code}
                  disabled={busy}
                  onChange={(event) => setCode(event.target.value)}
                />
              </Field>
            ) : null}
            <Field>
              <Button
                type="submit"
                disabled={
                  loading ||
                  busy ||
                  !mfa.ready ||
                  Boolean(mfa.error) ||
                  Boolean(secret) ||
                  !name.trim() ||
                  !password ||
                  (mfa.enabled && !code.trim())
                }
              >
                {busy ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <Plus data-icon="inline-start" />
                )}
                {t("agentTokens.create")}
              </Button>
            </Field>
          </FieldGroup>
        </form>
        {confirm || deleteID ? (
          <Alert>
            <AlertDescription className="flex flex-wrap items-center gap-2">
              <span>{t(deleteID ? "agentTokens.confirmDelete" : "agentTokens.confirmRevoke")}</span>
              <Button
                type="button"
                variant="destructive"
                disabled={busy}
                onClick={() =>
                  void run(async () => {
                    if (deleteID) await runtime.api.users.deleteAgentToken(deleteID);
                    else if (confirm === "all")
                      await runtime.api.users.revokeAllAgentTokens();
                    else await runtime.api.users.revokeAgentToken(confirm);
                    setSecret("");
                    setConfirm("");
                    setDeleteID("");
                    setRows(await runtime.api.users.agentTokens());
                  })
                }
              >
                {deleteID ? <Trash2 data-icon="inline-start" /> : <Ban data-icon="inline-start" />}
                {t(deleteID ? "agentTokens.delete" : "applications.revoke")}
              </Button>
              <Button
                type="button"
                variant="ghost"
                disabled={busy}
                onClick={() => { setConfirm(""); setDeleteID(""); }}
              >
                {t("cancel")}
              </Button>
            </AlertDescription>
          </Alert>
        ) : null}
        {loading ? (
          <Spinner />
        ) : !rows.length ? (
          <p className="text-sm text-muted-foreground">
            {t("agentTokens.empty")}
          </p>
        ) : (
          <div className="divide-y">
            {rows.map((row) => (
              <article
                key={row.id}
                className="flex flex-wrap items-center justify-between gap-3 py-3"
              >
                <div className="min-w-0">
                  <h3 className="break-words font-medium">{row.name}</h3>
                  <p className="font-mono text-xs text-muted-foreground">
                    {row.prefix}
                  </p>
                  <div className="my-2 flex flex-wrap gap-1">
                    {row.scopes.map((scope) => (
                      <Badge key={scope} variant="outline">
                        {t(`agentTokens.scopes.${scope}`, {
                          nsSeparator: false,
                        })}
                      </Badge>
                    ))}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {row.revokedAt
                      ? t("agentTokens.revoked")
                      : t("agentTokens.expires", {
                          date: new Date(row.expiresAt).toLocaleString(),
                        })}
                  </p>
                </div>
                <div className="flex shrink-0 gap-1"><Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  title={t("applications.revoke")}
                  aria-label={`${t("applications.revoke")} ${row.name}`}
                  disabled={busy || Boolean(row.revokedAt)}
                  onClick={() => { setDeleteID(""); setConfirm(row.id); }}
                >
                  <Ban />
                </Button>
                <Button type="button" variant="ghost" size="icon-sm" title={t("agentTokens.delete")} aria-label={`${t("agentTokens.delete")} ${row.name}`} disabled={busy} onClick={() => { setConfirm(""); setDeleteID(row.id); }}><Trash2 /></Button></div>
              </article>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
