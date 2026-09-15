import { useCallback, useEffect, useState } from "react";
import type {
  LayoutPayload,
  SaveUserInfoInput,
  SettingsPageProps,
} from "@gooseforum/client";
import {
  CalendarDays,
  Camera,
  Check,
  Image,
  Link as LinkIcon,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { Alert, AlertDescription } from "@gooseforum/ui/components/alert";
import { Badge } from "@gooseforum/ui/components/badge";
import { Button } from "@gooseforum/ui/components/button";
import { Field, FieldLabel } from "@gooseforum/ui/components/field";
import { Input } from "@gooseforum/ui/components/input";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@gooseforum/ui/components/popover";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@gooseforum/ui/components/tabs";
import { ProfileAvatar } from "../users/profile-avatar";
import { ProfileBadge } from "../users/profile-badge";
import { AccountSettings } from "./settings-account";
import { ConnectionsSettings } from "./settings-connections";
import { PrivacySettings } from "./settings-privacy";
import { ProfileSettings } from "./settings-profile";
import { AvatarCropDialog } from "./avatar-crop-dialog";
import { useAvatarCrop } from "./use-avatar-crop";
import { useGooseRuntime } from "@gooseforum/runtime";
import { cn } from "@gooseforum/ui/lib/utils";

const tabKeys = [
  "profile",
  "account",
  "privacy",
  "binding",
  "applications",
] as const;
type TabKey = (typeof tabKeys)[number];
const presetAvatars = Array.from(
  { length: 12 },
  (_, index) => `/static/pic/${index + 1}.webp?t=1788958424`,
);

export function SettingsPageView({
  page,
  layout,
}: {
  page: SettingsPageProps;
  layout: LayoutPayload;
}) {
  const { t } = useTranslation("settings");
  const runtime = useGooseRuntime();
  const [activeTab, setActiveTab] = useState<TabKey>(readTab);
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");
  const [cover, setCover] = useState(page.user.profileCoverUrl || "");
  const [profile, setProfile] = useState<SaveUserInfoInput>(() => ({
    nickname: page.user.nickname || "",
    locale: page.user.locale || runtime.locale,
    bio: page.user.bio || "",
    signature: page.user.signature || "",
    websiteName: page.user.websiteName || "",
    website: page.user.website || "",
    externalInformation: buildExternalInformation(page),
  }));
  const [username, setUsername] = useState(page.user.username);
  const [email, setEmail] = useState(page.user.email);
  const [presetDraft, setPresetDraft] = useState(page.user.avatarUrl);
  const [savingPreset, setSavingPreset] = useState(false);
  const [wornCode, setWornCode] = useState(page.user.wornBadgeCode || "");
  const [savingBadge, setSavingBadge] = useState(false);
  useEffect(() => {
    if (!status) return;
    const timer = window.setTimeout(() => setStatus(""), 3000);
    return () => window.clearTimeout(timer);
  }, [status]);

  const showStatus = useCallback((message: string) => {
    setError("");
    setStatus(message);
  }, []);

  const showError = useCallback((message: string) => {
    setStatus("");
    setError(message);
  }, []);

  const crop = useAvatarCrop({
    initialUrl: page.user.avatarUrl,
    onSuccess(url) {
      setPresetDraft(url);
      showStatus(t("status.avatarSaved"));
    },
    onError: showError,
  });

  function changeTab(tab: TabKey) {
    setActiveTab(tab);
    const url = new URL(window.location.href);
    if (tab === "profile") url.searchParams.delete("tab");
    else url.searchParams.set("tab", tab);
    window.history.replaceState(window.history.state, "", url);
  }

  async function applyPreset() {
    if (savingPreset || presetDraft === crop.avatarUrl) return;
    setSavingPreset(true);
    try {
      const result = await runtime.api.users.savePresetAvatar(presetDraft);
      const url = result.avatarUrl || presetDraft;
      crop.setAvatarUrl(url);
      setPresetDraft(url);
      showStatus(t("status.avatarSaved"));
    } catch (reason) {
      showError(errorMessage(reason, t("errors.avatar")));
    } finally {
      setSavingPreset(false);
    }
  }

  async function applyBadge() {
    setSavingBadge(true);
    try {
      await runtime.api.users.wearBadge(wornCode);
      showStatus(t("status.badgeSaved"));
    } catch (reason) {
      showError(errorMessage(reason, t("errors.badge")));
    } finally {
      setSavingBadge(false);
    }
  }

  const wornBadge =
    page.user.wearableBadges.find((badge) => badge.code === wornCode) || null;
  const displayName = profile.nickname || username;
  return (
    <main className="min-w-0 pb-8">
      <section className="site-panel overflow-hidden rounded-xl border bg-background">
        <div
          data-slot="profile-cover"
          className="relative h-20 border-b bg-muted bg-cover bg-center lg:h-24"
          style={{
            backgroundImage: cover
              ? `url(${JSON.stringify(cover)}), ${defaultCover}`
              : defaultCover,
          }}
        >
          {runtime.api.users.saveCover ? (
            <div className="absolute right-2 top-2">
              <CoverEditor
                value={cover}
                text={t}
                onSave={async (value) => {
                  try {
                    await runtime.api.users.saveCover(value);
                    setCover(value.trim());
                    showStatus(t("status.coverSaved"));
                    return true;
                  } catch (reason) {
                    showError(errorMessage(reason, t("errors.cover")));
                    return false;
                  }
                }}
              />
            </div>
          ) : null}
        </div>
        <div className="px-4 pb-4 lg:px-5">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-end">
            <div className="flex min-w-0 flex-1 gap-4">
              <button
                type="button"
                className="group relative -mt-9 size-24 shrink-0 rounded-full outline-none focus-visible:ring-4 focus-visible:ring-primary/20 lg:-mt-10 lg:size-28"
                aria-label={t("avatar.change")}
                onClick={crop.choose}
              >
                <ProfileAvatar
                  src={presetDraft}
                  name={username}
                  badge={wornBadge}
                  className="size-full"
                />
                <span className="pointer-events-none absolute inset-0 flex items-center justify-center rounded-full bg-foreground/0 text-background transition group-hover:bg-foreground/25">
                  <Camera className="size-8 opacity-0 drop-shadow-sm transition group-hover:opacity-100" />
                </span>
              </button>
              <input
                ref={crop.inputRef}
                type="file"
                accept="image/*"
                className="hidden"
                onChange={(event) => crop.selectFile(event.target.files?.[0])}
              />
              <div className="min-w-0 flex-1 pt-3">
                <div className="flex min-w-0 flex-wrap items-center gap-2">
                  <h1 className="truncate text-2xl font-bold leading-tight">
                    {displayName}
                  </h1>
                  <Badge variant="secondary">{t("editing")}</Badge>
                </div>
                <p className="mt-1 text-sm font-medium text-muted-foreground">
                  @{username}
                </p>
                <p className="mt-2 text-sm leading-relaxed text-foreground/75">
                  {profile.bio || profile.signature || t("emptyBio")}
                </p>
              </div>
            </div>
          </div>
          <Stats page={page} />
          <div className="mt-4 flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-muted-foreground">
            <span className="inline-flex items-center gap-1.5">
              <CalendarDays className="size-3.5" />
              {t("joinedAt", {
                date: new Date(page.stats.createdAt).toLocaleDateString(
                  runtime.locale,
                ),
              })}
            </span>
            {profile.website ? (
              <span className="inline-flex min-w-0 items-center gap-1.5">
                <LinkIcon className="size-3.5" />
                <span className="truncate">
                  {profile.websiteName || profile.website}
                </span>
              </span>
            ) : null}
          </div>
          <section className="mt-4 border-t pt-3">
            <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
              <div>
                <h2 className="text-sm font-semibold">
                  {t("avatar.presetsTitle")}
                </h2>
                <p className="text-xs text-muted-foreground">
                  {t("avatar.presetsDescription")}
                </p>
              </div>
              <div className="flex gap-2">
                {presetDraft !== crop.avatarUrl ? (
                  <Button
                    size="sm"
                    disabled={savingPreset}
                    onClick={() => void applyPreset()}
                  >
                    <Check data-icon="inline-start" />
                    {t("avatar.applyPreset")}
                  </Button>
                ) : null}
                <Button size="sm" variant="outline" onClick={crop.choose}>
                  <Camera data-icon="inline-start" />
                  {t("avatar.uploadCustom")}
                </Button>
              </div>
            </div>
            <div className="flex gap-1.5 overflow-x-auto pb-1">
              {presetAvatars.map((url) => (
                <button
                  key={url}
                  type="button"
                  aria-label={t("avatar.selectPreset")}
                  className={cn(
                    "relative size-11 shrink-0 overflow-hidden rounded-md border p-0.5",
                    presetDraft === url &&
                      "border-primary ring-2 ring-primary/15",
                  )}
                  onClick={() => setPresetDraft(url)}
                >
                  <img
                    src={url}
                    alt=""
                    className="size-full rounded object-cover"
                  />
                  {crop.avatarUrl === url ? (
                    <span className="absolute right-0.5 top-0.5 grid size-4 place-items-center rounded-full bg-primary text-primary-foreground">
                      <Check className="size-3" />
                    </span>
                  ) : null}
                </button>
              ))}
            </div>
          </section>
          {page.user.badges.length ? (
            <section className="mt-4 border-t pt-3">
              <div className="mb-2 flex items-center justify-between gap-2">
                <div>
                  <h2 className="text-sm font-semibold">
                    {t("avatar.wornBadgeTitle")}
                  </h2>
                  <p className="text-xs text-muted-foreground">
                    {t("avatar.wornBadgeDescription")}
                  </p>
                </div>
                <Button
                  size="sm"
                  disabled={savingBadge}
                  onClick={() => void applyBadge()}
                >
                  <Check data-icon="inline-start" />
                  {t("avatar.applyWornBadge")}
                </Button>
              </div>
              <div className="flex gap-2 overflow-x-auto pb-1">
                <button
                  type="button"
                  className={cn(
                    "grid size-16 shrink-0 place-items-center rounded-md border text-xs",
                    wornCode === "" && "border-primary ring-2 ring-primary/15",
                  )}
                  onClick={() => setWornCode("")}
                >
                  {t("avatar.noWornBadge")}
                </button>
                {page.user.badges.map((badge) => (
                  <button
                    key={badge.code}
                    type="button"
                    disabled={!badge.isWearable || savingBadge}
                    title={badge.description}
                    className={cn(
                      "relative flex size-16 shrink-0 flex-col items-center justify-center gap-1 rounded-md border",
                      !badge.isWearable &&
                        "cursor-not-allowed opacity-45 grayscale",
                      wornCode === badge.code &&
                        "border-primary ring-2 ring-primary/15",
                    )}
                    onClick={() => badge.isWearable && setWornCode(badge.code)}
                  >
                    <ProfileBadge badge={badge} compact />
                    <span className="max-w-full truncate px-1 text-[10px]">
                      {badge.name}
                    </span>
                  </button>
                ))}
              </div>
            </section>
          ) : null}
        </div>

        {status || error ? (
          <Alert
            variant={error ? "destructive" : "default"}
            className="mx-4 mb-3 w-auto lg:mx-5"
          >
            <AlertDescription>{error || status}</AlertDescription>
          </Alert>
        ) : null}
        <Tabs
          value={activeTab}
          onValueChange={(value) => changeTab(value as TabKey)}
          className="gap-0"
        >
          <div className="border-t bg-muted/30 px-3">
            <TabsList
              variant="line"
              aria-label={t("tabsLabel")}
              className="-m-1 h-auto w-full justify-start gap-1 overflow-x-auto rounded-none bg-transparent p-1 group-data-horizontal/tabs:h-auto"
            >
              {page.tabs.map((tab) => (
                <TabsTrigger
                  key={tab.key}
                  value={tab.key}
                  className="h-9 flex-none rounded-md border-0 px-3 text-sm font-semibold text-muted-foreground after:bg-primary group-data-horizontal/tabs:after:bottom-0 data-active:text-primary focus-visible:ring-2"
                >
                  {tabLabel(tab.key, tab.label, t)}
                </TabsTrigger>
              ))}
            </TabsList>
          </div>

          <TabsContent value="profile">
            <ProfileSettings
              page={page}
              requiresEmailVerification={layout.viewer.requiresEmailVerification}
              profile={profile}
              username={username}
              email={email}
              setProfile={setProfile}
              setUsername={setUsername}
              setEmail={setEmail}
              showStatus={showStatus}
              showError={showError}
            />
          </TabsContent>
          <TabsContent value="account">
            <AccountSettings showStatus={showStatus} showError={showError} />
          </TabsContent>
          <TabsContent value="privacy">
            <PrivacySettings showStatus={showStatus} showError={showError} />
          </TabsContent>
          <TabsContent value="binding">
            <ConnectionsSettings
              section="binding"
              showStatus={showStatus}
              showError={showError}
            />
          </TabsContent>
          <TabsContent value="applications">
            <ConnectionsSettings
              section="applications"
              showStatus={showStatus}
              showError={showError}
            />
          </TabsContent>
        </Tabs>
      </section>
      <AvatarCropDialog crop={crop} />
    </main>
  );
}

function Stats({ page }: { page: SettingsPageProps }) {
  const { t } = useTranslation("user");
  const items = [
    ["topics", page.stats.topicCount],
    ["replies", page.stats.replyCount],
    ["likesReceived", page.stats.likeReceivedCount],
    ["likesGiven", page.stats.likeGivenCount],
    ["followers", page.stats.followerCount],
    ["following", page.stats.followingCount],
    ["bookmarks", page.stats.collectionCount],
  ] as const;
  return (
    <div className="mt-5 grid grid-cols-4 border-y py-3 lg:grid-cols-7 lg:py-4">
      {items.map(([key, value]) => (
        <div key={key} className="px-1 py-1 text-center lg:text-left">
          <div className="text-lg font-bold tabular-nums lg:text-xl">
            {value.toLocaleString()}
          </div>
          <div className="text-[11px] font-medium text-muted-foreground lg:text-xs">
            {t(`stats.${key}`)}
          </div>
        </div>
      ))}
    </div>
  );
}

function CoverEditor({
  value,
  text,
  onSave,
}: {
  value: string;
  text: ReturnType<typeof useTranslation>["t"];
  onSave(value: string): Promise<boolean>;
}) {
  const [draft, setDraft] = useState(value);
  const [saving, setSaving] = useState(false);
  const [open, setOpen] = useState(false);
  return (
    <Popover
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (next) setDraft(value);
      }}
    >
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="secondary"
          size="sm"
          className="border border-background/40 bg-background/80 shadow-sm backdrop-blur-sm hover:bg-background/95"
        >
          <Image data-icon="inline-start" />
          {text("cover")}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-80 max-w-[calc(100vw-2rem)]">
        <form
          className="flex flex-col gap-3"
          onSubmit={async (event) => {
            event.preventDefault();
            setSaving(true);
            try {
              if (await onSave(draft)) setOpen(false);
            } finally {
              setSaving(false);
            }
          }}
        >
          <Field>
            <FieldLabel htmlFor="settings-cover">{text("coverUrl")}</FieldLabel>
            <Input
              id="settings-cover"
              type="url"
              value={draft}
              placeholder="https://example.com/cover.webp"
              onChange={(event) => setDraft(event.target.value)}
            />
          </Field>
          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="secondary"
              size="sm"
              onClick={() => setOpen(false)}
            >
              {text("cancel")}
            </Button>
            <Button type="submit" size="sm" disabled={saving}>
              {saving ? text("savingShort") : text("save")}
            </Button>
          </div>
        </form>
      </PopoverContent>
    </Popover>
  );
}

function readTab(): TabKey {
  const key = new URL(window.location.href).searchParams.get("tab") as TabKey;
  return tabKeys.includes(key) ? key : "profile";
}

function buildExternalInformation(page: SettingsPageProps) {
  return Object.fromEntries(
    ["github", "twitter", "linkedIn", "weibo", "bilibili", "zhihu"].map(
      (key) => [
        key,
        { link: page.user.externalInformation?.[key]?.link || "" },
      ],
    ),
  );
}

function tabLabel(
  key: string,
  fallback: string | undefined,
  t: ReturnType<typeof useTranslation>["t"],
) {
  return key === "applications"
    ? t("tabs.applications")
    : t(`tabs.${key}`, { defaultValue: fallback || key });
}

function errorMessage(reason: unknown, fallback: string) {
  return reason instanceof Error ? reason.message : fallback;
}

const defaultCover =
  "linear-gradient(135deg, var(--muted) 0%, var(--gf-color-info-content) 52%, var(--muted) 100%)";
