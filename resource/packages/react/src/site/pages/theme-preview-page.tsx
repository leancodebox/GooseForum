import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import {
  cloneSiteThemeTokens,
  siteThemeTokenKeys,
  type LayoutPayload,
  type SiteThemeConfig,
  type SiteThemeDefinition,
  type SiteThemeTokenKey,
  type ThemePreviewProps,
} from "@gooseforum/client";
import {
  Clipboard,
  Eye,
  PaintBucket,
  Rocket,
  RotateCcw,
  Save,
  SlidersHorizontal,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { Alert, AlertDescription } from "../../components/ui/alert";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Checkbox } from "../../components/ui/checkbox";
import { Input } from "../../components/ui/input";
import { Slider } from "../../components/ui/slider";
import { Spinner } from "../../components/ui/spinner";
import { Switch } from "../../components/ui/switch";
import { Tabs, TabsList, TabsTrigger } from "../../components/ui/tabs";
import { Textarea } from "../../components/ui/textarea";
import { useGooseRuntime } from "../../runtime";
import { themePresets } from "../theme/theme-presets";

type ThemeName = "gf-light" | "gf-dark";
type PreviewMode = "forum" | "components" | "code";
const managerPermission = 5;
const colorGroups: Array<{
  label: string;
  description: string;
  tokens: Array<[SiteThemeTokenKey, string]>;
}> = [
  {
    label: "Base",
    description: "Canvas, panels and borders",
    tokens: [
      ["color-base-100", "Canvas"],
      ["color-base-200", "Page"],
      ["color-base-300", "Muted"],
      ["color-base-content", "Text"],
      ["color-line", "Line"],
      ["color-icon-muted", "Icon"],
    ],
  },
  {
    label: "Brand",
    description: "Actions and navigation",
    tokens: [
      ["color-primary", "Primary"],
      ["color-primary-content", "On primary"],
      ["color-secondary", "Secondary"],
      ["color-secondary-content", "On secondary"],
      ["color-accent", "Accent"],
      ["color-accent-content", "On accent"],
      ["color-neutral", "Neutral"],
      ["color-neutral-content", "On neutral"],
    ],
  },
  {
    label: "State",
    description: "Feedback and badges",
    tokens: [
      ["color-info", "Info"],
      ["color-info-content", "On info"],
      ["color-success", "Success"],
      ["color-success-content", "On success"],
      ["color-warning", "Warning"],
      ["color-warning-content", "On warning"],
      ["color-error", "Error"],
      ["color-error-content", "On error"],
    ],
  },
];
const radiusTokens: Array<[SiteThemeTokenKey, string, number]> = [
  ["radius-box", "Boxes", 32],
  ["radius-field", "Fields", 24],
  ["radius-selector", "Selectors", 24],
];

export function ThemePreviewPageView({
  layout,
  page,
}: {
  layout: LayoutPayload;
  page: ThemePreviewProps;
}) {
  const { t } = useTranslation("themePreview");
  const runtime = useGooseRuntime();
  const [draft, setDraft] = useState(() => fromPrepublish(page.theme));
  const [saved, setSaved] = useState(() => cloneConfig(page.theme));
  const [themeName, setThemeName] = useState<ThemeName>(runtime.theme);
  const [mode, setMode] = useState<PreviewMode>("forum");
  const [saving, setSaving] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const originalTheme = useRef(runtime.theme);
  const canManage = layout.viewer.adminPermissions.includes(managerPermission);
  const active = theme(draft, themeName, page.defaults);
  const defaults = theme(page.defaults, themeName, page.defaults);
  const css = useMemo(() => buildThemeCss(draft), [draft]);
  const dirty = editSignature(draft) !== editSignature(fromPrepublish(saved));
  useLayoutEffect(() => {
    applyPreview(active, themeName);
    return () => clearPreview(originalTheme.current);
  }, [active, themeName]);
  useEffect(() => {
    function beforeUnload(event: BeforeUnloadEvent) {
      if (dirty) {
        event.preventDefault();
        event.returnValue = "";
      }
    }
    window.addEventListener("beforeunload", beforeUnload);
    return () => window.removeEventListener("beforeunload", beforeUnload);
  }, [dirty]);
  function setToken(key: SiteThemeTokenKey, value: string) {
    setDraft((current) =>
      updateTheme(current, themeName, (item) => ({
        ...item,
        tokens: { ...item.tokens, [key]: value },
      })),
    );
  }
  function applyPreset(key: string) {
    const preset = themePresets.find((item) => item.key === key);
    if (!preset) return;
    setDraft((current) => ({
      ...current,
      themes: current.themes.map((item) => ({
        ...item,
        tokens: { ...item.tokens, ...preset.themes[item.name] },
      })),
    }));
  }
  function resetSelected() {
    setDraft((current) =>
      updateTheme(current, themeName, () => cloneTheme(defaults)),
    );
    setMessage(t("themeRestoredDefault", { name: defaults.label }));
  }
  function resetAll() {
    const next = cloneConfig(page.defaults);
    next.publishedAt = saved.publishedAt;
    setDraft(next);
    setMessage(t("allRestoredDefault"));
  }
  function restoreSaved() {
    setDraft(fromPrepublish(saved));
    setMessage(t("restoredToSaved"));
    setError("");
  }
  async function saveDraft() {
    if (!canManage || saving || publishing) return;
    setSaving(true);
    setError("");
    try {
      const result = await runtime.api.themes.save(cloneConfig(draft));
      setSaved(cloneConfig(result));
      setDraft(fromPrepublish(result));
      setMessage(t("draftSaved"));
      runtime.queueFlash(t("draftSaved"), "success");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }
  async function publish() {
    if (!canManage || saving || publishing) return;
    setPublishing(true);
    setError("");
    try {
      await runtime.api.themes.save(cloneConfig(draft));
      const result = await runtime.api.themes.publish();
      setSaved(cloneConfig(result));
      setDraft(fromPrepublish(result));
      setMessage(t("published"));
      runtime.queueFlash(t("published"), "success");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("publishFailed"));
    } finally {
      setPublishing(false);
    }
  }
  async function copyCss() {
    try {
      await navigator.clipboard.writeText(css || "/* custom theme disabled */");
      setMessage(t("cssCopied"));
    } catch {
      setError(t("cssCopyFailed"));
    }
  }
  return (
    <main className="grid gap-3 lg:h-[calc(100vh-5.5rem)] lg:min-h-0 lg:grid-cols-[minmax(280px,340px)_minmax(0,1fr)] lg:overflow-hidden">
      <section className="overflow-hidden rounded-xl border bg-background lg:flex lg:min-h-0 lg:flex-col">
        <header className="border-b p-3">
          <div className="flex items-start justify-between gap-3">
            <div>
              <div className="flex items-center gap-2">
                <PaintBucket className="text-primary" />
                <h1 className="font-semibold">{t("pageTitle")}</h1>
              </div>
              <div className="mt-2 flex flex-wrap gap-2 text-xs">
                <Badge variant="secondary">v{draft.version || 1}</Badge>
                <Badge variant={draft.enabled ? "default" : "outline"}>
                  {t(draft.enabled ? "enabled" : "disabled")}
                </Badge>
                {dirty ? (
                  <Badge variant="secondary">{t("unsaved")}</Badge>
                ) : null}
                {!canManage ? (
                  <Badge variant="destructive">{t("readOnly")}</Badge>
                ) : null}
              </div>
            </div>
            <div className="flex gap-1">
              <Button variant="outline" size="sm" onClick={resetAll}>
                {t("resetDefault")}
              </Button>
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={t("restoreToSavedTitle")}
                onClick={restoreSaved}
              >
                <RotateCcw />
              </Button>
            </div>
          </div>
          <Tabs
            value={themeName}
            onValueChange={(value) => setThemeName(value as ThemeName)}
            className="mt-3 gap-0"
          >
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="gf-light">{t("light")}</TabsTrigger>
              <TabsTrigger value="gf-dark">{t("dark")}</TabsTrigger>
            </TabsList>
          </Tabs>
          <div className="mt-3 flex items-center justify-between rounded-lg bg-muted px-3 py-2">
            <span className="text-sm font-medium">{t("enableLabel")}</span>
            <Switch
              checked={draft.enabled}
              disabled={!canManage}
              onCheckedChange={(enabled) =>
                setDraft((current) => ({ ...current, enabled }))
              }
            />
          </div>
        </header>
        <div className="min-h-0 flex-1 overflow-y-auto p-3">
          <section>
            <h2 className="mb-2 text-sm font-semibold">{t("presetsTitle")}</h2>
            <div className="grid grid-cols-2 gap-2">
              {themePresets.map((preset) => (
                <Button
                  key={preset.key}
                  variant="outline"
                  className="h-auto justify-between px-2 py-2"
                  onClick={() => applyPreset(preset.key)}
                >
                  <span className="truncate">{t(`presets.${preset.key}`)}</span>
                  <span className="flex -space-x-1">
                    {preset.swatches.map((color) => (
                      <span
                        key={color}
                        className="size-3 rounded-full border border-background"
                        style={{ backgroundColor: color }}
                      />
                    ))}
                  </span>
                </Button>
              ))}
            </div>
          </section>
          <section className="mt-5">
            <div className="mb-3 flex items-center gap-2">
              <PaintBucket className="text-muted-foreground" />
              <h2 className="text-sm font-semibold">{t("colors")}</h2>
              <Button
                className="ml-auto"
                variant="ghost"
                size="sm"
                onClick={resetSelected}
              >
                {t("resetDefault")}
              </Button>
            </div>
            <div className="flex flex-col gap-4">
              {colorGroups.map((group) => (
                <div key={group.label} className="border-b pb-4 last:border-0">
                  <h3 className="text-xs font-bold uppercase">{group.label}</h3>
                  <p className="text-[11px] text-muted-foreground">
                    {group.description}
                  </p>
                  <div className="mt-2 grid grid-cols-2 gap-2">
                    {group.tokens.map(([key, label]) => (
                      <ColorToken
                        key={key}
                        label={label}
                        value={active.tokens[key] || "#ffffff"}
                        disabled={!canManage}
                        onChange={(value) => setToken(key, value)}
                      />
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </section>
          <section className="mt-5">
            <div className="mb-3 flex items-center gap-2">
              <SlidersHorizontal className="text-muted-foreground" />
              <h2 className="text-sm font-semibold">{t("radius")}</h2>
            </div>
            <div className="flex flex-col gap-4">
              {radiusTokens.map(([key, label, max]) => (
                <div key={key}>
                  <div className="mb-2 flex justify-between text-xs">
                    <span>{label}</span>
                    <code>{active.tokens[key]}</code>
                  </div>
                  <Slider
                    aria-label={`${label} radius`}
                    value={[tokenPixels(active.tokens[key])]}
                    max={max}
                    step={1}
                    disabled={!canManage}
                    onValueChange={(value) =>
                      setToken(key, `${(value[0] || 0) / 16}rem`)
                    }
                  />
                </div>
              ))}
            </div>
          </section>
          <section className="mt-5">
            <h2 className="mb-2 text-sm font-semibold">{t("effects")}</h2>
            <label className="flex items-center justify-between rounded-lg bg-muted px-3 py-2 text-sm">
              <span>{t("depth")}</span>
              <Checkbox
                checked={active.tokens.depth === "1"}
                disabled={!canManage}
                onCheckedChange={(checked) =>
                  setToken("depth", checked ? "1" : "0")
                }
              />
            </label>
            <Contrast theme={active} label={t("contrast")} />
          </section>
        </div>
      </section>
      <section className="overflow-hidden rounded-xl border bg-background lg:flex lg:min-h-0 lg:flex-col">
        <header className="flex flex-wrap items-center justify-between gap-2 border-b p-3">
          <div className="flex items-center gap-2">
            <Eye className="text-muted-foreground" />
            <div>
              <h2 className="text-sm font-semibold">{t("preview")}</h2>
              <p className="text-xs text-muted-foreground">
                {layout.site.name || "GooseForum"}
              </p>
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            <Tabs
              value={mode}
              onValueChange={(value) => setMode(value as PreviewMode)}
              className="gap-0"
            >
              <TabsList>
                <TabsTrigger value="forum">{t("forum")}</TabsTrigger>
                <TabsTrigger value="components">{t("components")}</TabsTrigger>
                <TabsTrigger value="code">{t("code")}</TabsTrigger>
              </TabsList>
            </Tabs>
            <Button
              variant="outline"
              size="sm"
              disabled={!canManage || saving || publishing}
              onClick={() => void saveDraft()}
            >
              {saving ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <Save data-icon="inline-start" />
              )}
              {t("saveDraft")}
            </Button>
            <Button
              size="sm"
              disabled={!canManage || saving || publishing}
              onClick={() => void publish()}
            >
              {publishing ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <Rocket data-icon="inline-start" />
              )}
              {t("publishSite")}
            </Button>
          </div>
        </header>
        <div className="min-h-0 flex-1 overflow-y-auto bg-muted/40 p-3">
          <Alert className="mb-3">
            <AlertDescription>
              <strong>{t("workflowTitle")}</strong> {t("workflowDescription")}
            </AlertDescription>
          </Alert>
          {error ? (
            <Alert variant="destructive" className="mb-3">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : null}
          {message ? (
            <Alert className="mb-3">
              <AlertDescription>{message}</AlertDescription>
            </Alert>
          ) : null}
          {mode === "forum" ? (
            <ForumPreview t={t} />
          ) : mode === "components" ? (
            <ComponentPreview />
          ) : (
            <div className="rounded-xl border bg-background p-3">
              <div className="mb-2 flex justify-between">
                <h3 className="font-semibold">CSS</h3>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => void copyCss()}
                >
                  <Clipboard data-icon="inline-start" />
                  {t("copyCss")}
                </Button>
              </div>
              <Textarea
                readOnly
                value={css || "/* custom theme disabled */"}
                className="min-h-[60vh] font-mono text-xs"
              />
            </div>
          )}
        </div>
      </section>
    </main>
  );
}

type Translate = ReturnType<typeof useTranslation>["t"];
function ColorToken({
  label,
  value,
  disabled,
  onChange,
}: {
  label: string;
  value: string;
  disabled: boolean;
  onChange(value: string): void;
}) {
  const color = /^#[0-9a-f]{6}$/i.test(value) ? value : "#ffffff";
  return (
    <label className="flex items-center gap-2 rounded-lg border p-2">
      <input
        type="color"
        aria-label={`${label} color`}
        value={color}
        disabled={disabled}
        onChange={(event) => onChange(event.target.value)}
        className="size-7 cursor-pointer rounded border-0 bg-transparent p-0"
      />
      <span className="min-w-0 flex-1">
        <span className="block truncate text-xs font-medium">{label}</span>
        <Input
          aria-label={`${label} value`}
          value={value}
          disabled={disabled}
          onChange={(event) => onChange(event.target.value)}
          className="mt-1 h-6 px-1.5 font-mono text-[10px]"
        />
      </span>
    </label>
  );
}
function ForumPreview({ t }: { t: Translate }) {
  return (
    <div className="grid gap-3 2xl:grid-cols-[minmax(0,1fr)_240px]">
      <div className="flex flex-col gap-3">
        <section className="overflow-hidden rounded-xl border bg-background">
          <header className="flex items-center justify-between border-b p-3">
            <div className="flex gap-2">
              <Badge>{t("sampleTopic1")}</Badge>
              <Badge variant="secondary">{t("sampleTopic2")}</Badge>
            </div>
            <Button size="sm">
              <Rocket data-icon="inline-start" />
              {t("samplePublish")}
            </Button>
          </header>
          {[1, 2, 3].map((index) => (
            <article
              key={index}
              className="border-b px-3 py-3 last:border-0 hover:bg-muted/50"
            >
              <h3 className="font-semibold">{t(`sampleTopic${index}`)}</h3>
              <p className="mt-1 text-sm text-muted-foreground">
                {t("sampleTopicExcerpt")}
              </p>
            </article>
          ))}
        </section>
        <article className="rounded-xl border bg-background p-4">
          <h2 className="text-xl font-semibold">{t("sampleTopicTitle")}</h2>
          <div className="gf-prose mt-3">
            <p>{t("sampleTopicBody")}</p>
            <blockquote>{t("sampleTopicQuote")}</blockquote>
            <pre>
              <code>const theme = 'goose'</code>
            </pre>
          </div>
        </article>
      </div>
      <aside className="flex flex-col gap-3">
        <div className="rounded-xl border bg-background p-3">
          <h3 className="font-semibold">Messages</h3>
          <div className="mt-3 rounded-lg bg-muted p-2 text-sm">
            {t("sampleMessageIncoming")}
          </div>
          <div className="ml-6 mt-2 rounded-lg bg-primary p-2 text-sm text-primary-foreground">
            {t("sampleMessageOutgoing")}
          </div>
        </div>
        <StatusPreview />
      </aside>
    </div>
  );
}
function ComponentPreview() {
  return (
    <div className="grid gap-3 md:grid-cols-2">
      <div className="rounded-xl border bg-background p-4">
        <div className="h-28 rounded-xl bg-muted" />
        <h3 className="mt-3 font-semibold">Card title</h3>
        <p className="text-sm text-muted-foreground">
          Base, line, radius and shadow.
        </p>
        <div className="mt-4 flex gap-2">
          <Button size="sm">Primary</Button>
          <Button variant="outline" size="sm">
            Secondary
          </Button>
        </div>
      </div>
      <div className="rounded-xl border bg-background p-4">
        <h3 className="font-semibold">Form states</h3>
        <Input className="mt-3" defaultValue="GooseForum" />
        <Textarea
          className="mt-3"
          defaultValue="Theme variables cover form controls."
        />
        <div className="mt-3 flex items-center justify-between rounded-lg bg-muted px-3 py-2">
          <span>Selector</span>
          <Switch defaultChecked />
        </div>
      </div>
      <StatusPreview />
    </div>
  );
}
function StatusPreview() {
  return (
    <div className="rounded-xl border bg-background p-3">
      <h3 className="font-semibold">Status</h3>
      <div className="mt-3 grid gap-2 text-sm">
        <div className="rounded-lg bg-primary/10 px-3 py-2 text-primary">
          Info notification
        </div>
        <div className="rounded-lg bg-success/10 px-3 py-2 text-success">
          Success state
        </div>
        <div className="rounded-lg bg-warning/10 px-3 py-2 text-warning">
          Warning state
        </div>
        <div className="rounded-lg bg-destructive/10 px-3 py-2 text-destructive">
          Error state
        </div>
      </div>
    </div>
  );
}
function Contrast({
  theme: value,
  label,
}: {
  theme: SiteThemeDefinition;
  label: string;
}) {
  const pairs: Array<[string, SiteThemeTokenKey, SiteThemeTokenKey]> = [
    ["Base", "color-base-content", "color-base-100"],
    ["Primary", "color-primary-content", "color-primary"],
    ["Success", "color-success-content", "color-success"],
    ["Error", "color-error-content", "color-error"],
  ];
  return (
    <div className="mt-4">
      <h3 className="mb-2 text-xs font-bold uppercase">{label}</h3>
      {pairs.map(([name, foreground, background]) => {
        const ratio = contrastRatio(
          value.tokens[foreground],
          value.tokens[background],
        );
        return (
          <div
            key={name}
            className="grid grid-cols-[1fr_48px_38px] gap-2 py-1 text-xs"
          >
            <span>{name}</span>
            <span className="font-mono">
              {ratio ? ratio.toFixed(2) : "n/a"}
            </span>
            <Badge
              variant={ratio >= 4.5 ? "secondary" : "destructive"}
              className="justify-center px-1"
            >
              {ratio >= 7 ? "AAA" : ratio >= 4.5 ? "AA" : "Low"}
            </Badge>
          </div>
        );
      })}
    </div>
  );
}
function cloneTheme(value: SiteThemeDefinition): SiteThemeDefinition {
  return { ...value, tokens: cloneSiteThemeTokens(value.tokens) };
}
function cloneConfig(value: SiteThemeConfig): SiteThemeConfig {
  return {
    ...value,
    themes: value.themes.map(cloneTheme),
    prepublish: value.prepublish
      ? { ...value.prepublish, themes: value.prepublish.themes.map(cloneTheme) }
      : undefined,
  };
}
function fromPrepublish(value: SiteThemeConfig) {
  const next = cloneConfig(value);
  if (next.prepublish) {
    next.enabled = next.prepublish.enabled;
    next.themes = next.prepublish.themes.map(cloneTheme);
  }
  return next;
}
function theme(
  config: SiteThemeConfig,
  name: ThemeName,
  defaults: SiteThemeConfig,
) {
  return (
    config.themes.find((item) => item.name === name) ||
    defaults.themes.find((item) => item.name === name)!
  );
}
function updateTheme(
  config: SiteThemeConfig,
  name: ThemeName,
  update: (theme: SiteThemeDefinition) => SiteThemeDefinition,
): SiteThemeConfig {
  return {
    ...config,
    themes: config.themes.map((item) =>
      item.name === name ? update(cloneTheme(item)) : item,
    ),
  };
}
function editSignature(config: SiteThemeConfig) {
  return JSON.stringify({ enabled: config.enabled, themes: config.themes });
}
function tokenPixels(value: string) {
  const number = Number.parseFloat(value);
  if (!Number.isFinite(number)) return 0;
  return value.endsWith("rem") ? Math.round(number * 16) : Math.round(number);
}
function applyPreview(value: SiteThemeDefinition, name: ThemeName) {
  document.documentElement.dataset.theme = name;
  for (const key of siteThemeTokenKeys)
    document.documentElement.style.setProperty(
      `--gf-${key}`,
      value.tokens[key],
    );
}
function clearPreview(name: ThemeName) {
  for (const key of siteThemeTokenKeys)
    document.documentElement.style.removeProperty(`--gf-${key}`);
  document.documentElement.dataset.theme = name;
}
function buildThemeCss(config: SiteThemeConfig) {
  if (!config.enabled) return "";
  return config.themes
    .map(
      (item) =>
        `[data-theme="${item.name}"]{color-scheme:${item.colorScheme};${siteThemeTokenKeys
          .map((key) => [key, sanitize(item.tokens[key])] as const)
          .filter((entry) => entry[1])
          .map(([key, value]) => `--gf-${key}:${value}`)
          .join(";")}}`,
    )
    .join("\n");
}
function sanitize(value: string) {
  const next = value.trim();
  return /[{};<>]/.test(next) ? "" : next;
}
function contrastRatio(foreground: string, background: string) {
  const fg = luminance(foreground);
  const bg = luminance(background);
  if (fg === null || bg === null) return 0;
  return (Math.max(fg, bg) + 0.05) / (Math.min(fg, bg) + 0.05);
}
function luminance(value: string) {
  if (!/^#[0-9a-f]{6}$/i.test(value)) return null;
  const channels = [1, 3, 5]
    .map((index) => Number.parseInt(value.slice(index, index + 2), 16) / 255)
    .map((channel) =>
      channel <= 0.03928 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4,
    );
  return 0.2126 * channels[0]! + 0.7152 * channels[1]! + 0.0722 * channels[2]!;
}
