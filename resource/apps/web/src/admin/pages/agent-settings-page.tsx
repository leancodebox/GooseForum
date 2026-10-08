import { useEffect, useRef, useState } from "react";
import type { AgentSettings, GooseAdminApi } from "@gooseforum/client";
import { Button } from "@gooseforum/ui/components/button";
import {
  Field,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@gooseforum/ui/components/field";
import { Input } from "@gooseforum/ui/components/input";
import { Switch } from "@gooseforum/ui/components/switch";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { RefreshCw, Save } from "lucide-react";
import { toast } from "sonner";
import type { SystemSettingsTextKey } from "../system-settings-i18n";

const limits = [
  ["readPerMinute", "agentRead"],
  ["userReadPerMinute", "agentUserRead"],
  ["writePerMinute", "agentWrite"],
  ["anonymousReadPerMinute", "agentAnonymousRead"],
  ["ipPerMinute", "agentIP"],
] as const;

export function AgentSettingsPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text(key: SystemSettingsTextKey): string;
}) {
  const [form, setForm] = useState<AgentSettings>();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [reload, setReload] = useState(0);
  const inFlight = useRef(false);
  const translate = useRef(text);
  useEffect(() => {
    translate.current = text;
  }, [text]);
  useEffect(() => {
    let active = true;
    void api.settings
      .agent()
      .then((value) => {
        if (active) setForm(value);
      })
      .catch((reason: unknown) => {
        if (active)
          toast.error(
            reason instanceof Error
              ? reason.message
              : translate.current("loadFailed"),
          );
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [api, reload]);

  async function save(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!form || inFlight.current) return;
    inFlight.current = true;
    setSaving(true);
    try {
      await api.settings.saveAgent(form);
      toast.success(text("saved"));
    } catch (reason) {
      toast.error(
        reason instanceof Error ? reason.message : text("saveFailed"),
      );
    } finally {
      inFlight.current = false;
      setSaving(false);
    }
  }

  return (
    <main className="flex flex-1 flex-col gap-5 px-3 py-3 lg:px-4">
      <form
        onSubmit={(event) => void save(event)}
        className="flex flex-col gap-6"
      >
        <header className="flex items-center justify-between gap-4">
          <h2 className="text-lg font-semibold">{text("agent")}</h2>
          <Button type="submit" size="sm" disabled={!form || loading || saving}>
            {saving ? (
              <Spinner data-icon="inline-start" />
            ) : (
              <Save data-icon="inline-start" />
            )}
            {text("save")}
          </Button>
        </header>
        {loading ? (
          <Spinner />
        ) : !form ? (
          <Button
            type="button"
            variant="outline"
            className="w-fit"
            onClick={() => {
              setLoading(true);
              setReload((value) => value + 1);
            }}
          >
            <RefreshCw data-icon="inline-start" />
            {text("agentRetry")}
          </Button>
        ) : (
          <>
            <FieldGroup className="max-w-xl">
              {(
                [
                  ["enabled", "agentEnabled"],
                  ["manualTokens", "agentManualTokens"],
                ] as const
              ).map(([key, label]) => (
                <Field key={key} orientation="horizontal">
                  <FieldLabel htmlFor={`agent-${key}`}>
                    {text(label)}
                  </FieldLabel>
                  <Switch
                    id={`agent-${key}`}
                    disabled={saving}
                    checked={form[key]}
                    onCheckedChange={(value) =>
                      setForm(
                        (current) => current && { ...current, [key]: value },
                      )
                    }
                  />
                </Field>
              ))}
            </FieldGroup>
            <FieldSet className="max-w-3xl">
              <FieldLegend>{text("agentLimits")}</FieldLegend>
              <FieldGroup className="grid gap-5 sm:grid-cols-2">
                {limits.map(([key, label]) => (
                  <Field key={key}>
                    <FieldLabel htmlFor={`agent-${key}`}>
                      {text(label)}
                    </FieldLabel>
                    <Input
                      id={`agent-${key}`}
                      type="number"
                      min={1}
                      max={100000}
                      step={1}
                      required
                      disabled={saving}
                      value={form[key]}
                      onChange={(event) => {
                        const value = Number(event.target.value);
                        setForm(
                          (current) => current && { ...current, [key]: value },
                        );
                      }}
                    />
                  </Field>
                ))}
              </FieldGroup>
            </FieldSet>
          </>
        )}
      </form>
    </main>
  );
}
