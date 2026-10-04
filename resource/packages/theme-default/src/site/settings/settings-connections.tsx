import { useEffect, useState } from "react";
import type {
  OIDCGrantPayload,
  OAuthBindingsPayload,
} from "@gooseforum/client";
import { AppWindow, Check, KeyRound, RefreshCw } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Badge } from "@gooseforum/ui/components/badge";
import { Alert, AlertDescription } from "@gooseforum/ui/components/alert";
import { Button } from "@gooseforum/ui/components/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { useGooseRuntime } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";
import { SettingsSectionHeader } from "./settings-section-header";
import { useMFAEnabled } from "./use-mfa-enabled";
import { Input } from "@gooseforum/ui/components/input";
import { Field, FieldLabel } from "@gooseforum/ui/components/field";

export function ConnectionsSettings({
  section,
  showStatus,
  showError,
}: {
  section: "binding" | "applications";
  showStatus(message: string): void;
  showError(message: string): void;
}) {
  const { t } = useTranslation("settings");
  const runtime = useGooseRuntime();
  const serverError = useServerErrorMessage();
  const [bindings, setBindings] = useState<OAuthBindingsPayload>([]);
  const [grants, setGrants] = useState<OIDCGrantPayload[]>([]);
  const [loadingBindings, setLoadingBindings] = useState(true);
  const [loadingGrants, setLoadingGrants] = useState(true);
  const [action, setAction] = useState("");
  const [confirm, setConfirm] = useState("");
  const { enabled: mfaEnabled, ready: mfaReady, error: mfaError, retry: retryMfa } = useMFAEnabled();
  const [mfaCode, setMfaCode] = useState("");

  useEffect(() => {
    let active = true;
    void runtime.api.users
      .oauthBindings()
      .then((items) => {
        if (active) setBindings(items);
      })
      .catch((reason) => { if (active) showError(serverError(reason, t("errors.bindings"))); })
      .finally(() => {
        if (active) setLoadingBindings(false);
      });
    void runtime.api.users
      .oidcGrants()
      .then((items) => {
        if (active) setGrants(items);
      })
      .catch((reason) => { if (active) showError(serverError(reason, t("errors.applications"))); })
      .finally(() => {
        if (active) setLoadingGrants(false);
      });
    return () => {
      active = false;
    };
  }, [runtime.api.users, serverError, showError, t]);

  async function refreshBindings() {
    setLoadingBindings(true);
    try {
      setBindings(await runtime.api.users.oauthBindings());
    } catch (reason) {
      showError(serverError(reason, t("errors.bindings")));
    } finally {
      setLoadingBindings(false);
    }
  }

  async function toggleBinding(provider: OAuthBindingsPayload[number]) {
    if (action || loadingBindings || !mfaReady || (mfaEnabled && !mfaCode.trim())) return;
    if (!provider.enabled && !provider.bound) return;
    setAction(provider.key);
    try {
      if (!provider.bound) {
        const result = await runtime.api.users.prepareOAuthBind(provider.key, mfaCode.trim());
        setMfaCode("");
        runtime.redirect(result.redirect);
        return;
      }
      if (mfaEnabled) await runtime.api.users.unbindOAuth(provider.key, mfaCode);
      else await runtime.api.users.unbindOAuth(provider.key);
      setMfaCode("");
      setBindings((items) =>
        items.map((item) =>
          item.key === provider.key
            ? { ...item, bound: false, boundAt: undefined }
            : item,
        ),
      );
      showStatus(t("status.bindingDisconnected"));
    } catch (reason) {
      showError(serverError(reason, t("errors.unbind")));
    } finally {
      setAction("");
    }
  }

  async function revoke(grant: OIDCGrantPayload) {
    setAction(grant.clientId);
    try {
      await runtime.api.users.revokeOIDCGrant(grant.clientId);
      setGrants((items) =>
        items.filter((item) => item.clientId !== grant.clientId),
      );
      setConfirm("");
      showStatus(t("status.applicationRevoked"));
    } catch (reason) {
      showError(serverError(reason, t("errors.revoke")));
    } finally {
      setAction("");
    }
  }

  if (section === "binding") {
    return (
      <section>
        <SettingsSectionHeader
          icon={KeyRound}
          title={t("binding.title")}
          actions={
            <Button
              type="button"
              variant="ghost"
              size="sm"
              disabled={loadingBindings || Boolean(action)}
              onClick={() => void refreshBindings()}
            >
              <RefreshCw data-icon="inline-start" />
              {t("binding.refresh")}
            </Button>
          }
        />
        {loadingBindings ? (
          <Loading label={t("binding.loading")} />
        ) : (
          <div className="flex flex-col gap-3 p-4">
            {mfaError && <Alert><AlertDescription className="flex flex-wrap items-center gap-2"><span className="flex-1">{t('mfa.statusFailed')}</span><Button type="button" variant="outline" size="sm" onClick={retryMfa}><RefreshCw />{t('mfa.retry')}</Button></AlertDescription></Alert>}
            {mfaEnabled && <Field className="max-w-sm">
              <FieldLabel htmlFor="unbind-mfa-code">{t("mfa.code")}</FieldLabel>
              <Input id="unbind-mfa-code" autoComplete="one-time-code" maxLength={64} value={mfaCode} onChange={(event) => setMfaCode(event.target.value)} />
            </Field>}
            {bindings.map((provider) => (
              <article
                key={provider.key}
                className="flex items-center justify-between gap-4 rounded-lg border p-4"
              >
                <div className="flex min-w-0 items-center gap-3">
                  <span className="grid size-11 shrink-0 place-items-center rounded-full border bg-background">
                    <KeyRound className="size-5" />
                  </span>
                  <div>
                    <h3 className="font-semibold">{provider.displayName}</h3>
                    <p className="text-sm text-muted-foreground">
                      {provider.enabled
                        ? provider.bound
                          ? t("binding.connected")
                          : t("binding.disconnected")
                        : t("binding.siteUnsupported")}
                    </p>
                  </div>
                </div>
                <Button
                  type="button"
                  variant={provider.bound ? "destructive" : "default"}
                  disabled={
                    Boolean(action) ||
                    !mfaReady ||
                    (mfaEnabled && !mfaCode.trim()) ||
                    (!provider.enabled && !provider.bound)
                  }
                  onClick={() => void toggleBinding(provider)}
                >
                  {action === provider.key ? (
                    <Spinner data-icon="inline-start" />
                  ) : provider.bound ? (
                    <Check data-icon="inline-start" />
                  ) : null}
                  {!provider.enabled && !provider.bound
                    ? t("binding.unsupported")
                    : provider.bound
                      ? t("binding.disconnect")
                      : t("binding.connect")}
                </Button>
              </article>
            ))}
          </div>
        )}
      </section>
    );
  }

  return (
    <section>
      <SettingsSectionHeader
        icon={AppWindow}
        title={t("applications.title")}
        description={t("applications.description")}
        actions={
          <Button
            type="button"
            variant="ghost"
            size="sm"
            disabled={loadingGrants || Boolean(action)}
            onClick={async () => {
              setLoadingGrants(true);
              try {
                setGrants(await runtime.api.users.oidcGrants());
              } catch (reason) {
                showError(serverError(reason, t("errors.applications")));
              } finally {
                setLoadingGrants(false);
              }
            }}
          >
            <RefreshCw data-icon="inline-start" />
            {t("binding.refresh")}
          </Button>
        }
      />
      {loadingGrants ? (
        <Loading label={t("applications.loading")} />
      ) : grants.length ? (
        <div className="divide-y">
          {grants.map((grant) => (
            <article
              key={grant.clientId}
              className="flex flex-col gap-3 p-4 lg:flex-row lg:items-center lg:justify-between"
            >
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <h3 className="font-semibold">{grant.name}</h3>
                  {!grant.enabled ? (
                    <Badge variant="secondary">
                      {t("applications.disabled")}
                    </Badge>
                  ) : null}
                </div>
                <p className="mt-1 truncate font-mono text-xs text-muted-foreground">
                  {grant.clientId}
                </p>
                <div className="mt-2 flex flex-wrap gap-1.5">
                  {grant.scopes.map((scope) => (
                    <Badge
                      key={scope}
                      variant="outline"
                      className="font-mono font-normal"
                    >
                      {scope}
                    </Badge>
                  ))}
                </div>
                <p className="mt-2 text-xs text-muted-foreground">
                  {t("applications.grantedAt", {
                    date: new Date(grant.grantedAt).toLocaleDateString(),
                  })}
                </p>
              </div>
              <div className="flex shrink-0 gap-2">
                {confirm === grant.clientId ? (
                  <>
                    <Button
                      type="button"
                      variant="secondary"
                      size="sm"
                      disabled={Boolean(action)}
                      onClick={() => setConfirm("")}
                    >
                      {t("cancel")}
                    </Button>
                    <Button
                      type="button"
                      variant="destructive"
                      size="sm"
                      disabled={Boolean(action)}
                      onClick={() => void revoke(grant)}
                    >
                      {action === grant.clientId ? (
                        <Spinner data-icon="inline-start" />
                      ) : null}
                      {t("applications.confirmRevoke")}
                    </Button>
                  </>
                ) : (
                  <Button
                    type="button"
                    variant="destructive"
                    size="sm"
                    disabled={Boolean(action)}
                    onClick={() => setConfirm(grant.clientId)}
                  >
                    {t("applications.revoke")}
                  </Button>
                )}
              </div>
            </article>
          ))}
        </div>
      ) : (
        <Empty className="min-h-48 border-0">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <AppWindow />
            </EmptyMedia>
            <EmptyTitle>{t("applications.empty")}</EmptyTitle>
            <EmptyDescription>
              {t("applications.emptyDescription")}
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
    </section>
  );
}

function Loading({ label }: { label: string }) {
  return (
    <div className="flex min-h-40 flex-col items-center justify-center gap-2 text-sm text-muted-foreground">
      <Spinner className="size-5" />
      {label}
    </div>
  );
}
