import { useState, type Dispatch, type SetStateAction } from "react";
import type { SaveUserInfoInput, SettingsPageProps } from "@gooseforum/client";
import { Link as LinkIcon, Mail, Pencil, UserRound } from "lucide-react";
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
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@gooseforum/ui/components/select";
import { Textarea } from "@gooseforum/ui/components/textarea";
import { useGooseRuntime } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";
import { SettingsSectionHeader } from "./settings-section-header";

const socialKeys = [
  "github",
  "twitter",
  "linkedIn",
  "weibo",
  "bilibili",
  "zhihu",
] as const;

export function ProfileSettings({
  page,
  requiresEmailVerification,
  profile,
  username,
  email,
  setProfile,
  setUsername,
  setEmail,
  showStatus,
  showError,
}: {
  page: SettingsPageProps;
  requiresEmailVerification: boolean;
  profile: SaveUserInfoInput;
  username: string;
  email: string;
  setProfile: Dispatch<SetStateAction<SaveUserInfoInput>>;
  setUsername(value: string): void;
  setEmail(value: string): void;
  showStatus(message: string): void;
  showError(message: string): void;
}) {
  const { t } = useTranslation("settings");
  const runtime = useGooseRuntime();
  const serverError = useServerErrorMessage();
  const [editingUsername, setEditingUsername] = useState(false);
  const [editingEmail, setEditingEmail] = useState(false);
  const [saving, setSaving] = useState("");

  async function run(
    key: string,
    action: () => Promise<unknown>,
    success?: string,
  ) {
    setSaving(key);
    try {
      await action();
      if (success) showStatus(success);
      return true;
    } catch (reason) {
      showError(serverError(reason, t("errors.save")));
    } finally {
      setSaving("");
    }
    return false;
  }

  async function saveUsername() {
    if (!username.trim()) return showError(t("validation.usernameRequired"));
    if (
      await run(
        "username",
        () => runtime.api.users.saveUsername(username.trim()),
        t("status.usernameSaved"),
      )
    )
      setEditingUsername(false);
  }

  async function saveEmail() {
    if (!email.trim()) return showError(t("validation.emailRequired"));
    if (
      await run(
        "email",
        () => runtime.api.users.saveEmail(email.trim()),
        t("status.emailSaved"),
      )
    )
      setEditingEmail(false);
  }

  async function sendActivation() {
    setSaving("activation");
    try {
      const result = await runtime.api.users.resendActivationEmail();
      showStatus(result.message || t("status.activationEmailSent"));
    } catch (reason) {
      showError(serverError(reason, t("errors.save")));
    } finally {
      setSaving("");
    }
  }

  return (
    <section>
      <SettingsSectionHeader
        icon={UserRound}
        title={t("profile.title")}
        description={t("profile.description")}
      />
      <div className="p-4">
        <FieldGroup>
          <div className="grid gap-4 lg:grid-cols-2">
            <Field>
              <FieldLabel htmlFor="settings-username">
                {t("username")}
              </FieldLabel>
              <div className="flex gap-2">
                <Input
                  id="settings-username"
                  value={username}
                  disabled={!editingUsername}
                  onChange={(event) => setUsername(event.target.value)}
                />
                {editingUsername ? (
                  <>
                    <Button
                      type="button"
                      disabled={saving === "username"}
                      onClick={() => void saveUsername()}
                    >
                      {t("save")}
                    </Button>
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => {
                        setUsername(page.user.username);
                        setEditingUsername(false);
                      }}
                    >
                      {t("cancel")}
                    </Button>
                  </>
                ) : (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setEditingUsername(true)}
                  >
                    <Pencil data-icon="inline-start" />
                    {t("edit")}
                  </Button>
                )}
              </div>
            </Field>
            <Field>
              <FieldLabel htmlFor="settings-email">{t("email")}</FieldLabel>
              <div className="flex gap-2">
                <Input
                  id="settings-email"
                  type="email"
                  value={email}
                  disabled={!editingEmail}
                  onChange={(event) => setEmail(event.target.value)}
                />
                {editingEmail ? (
                  <>
                    <Button
                      type="button"
                      disabled={saving === "email"}
                      onClick={() => void saveEmail()}
                    >
                      {t("save")}
                    </Button>
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => {
                        setEmail(page.user.email);
                        setEditingEmail(false);
                      }}
                    >
                      {t("cancel")}
                    </Button>
                  </>
                ) : (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setEditingEmail(true)}
                  >
                    <Pencil data-icon="inline-start" />
                    {t("edit")}
                  </Button>
                )}
              </div>
              {requiresEmailVerification ? (
                <div className="flex flex-col gap-2 border-l-2 border-warning bg-warning/10 px-3 py-2 lg:flex-row lg:items-center lg:justify-between">
                  <FieldDescription className="text-warning">
                    {t("emailVerification.description")}
                  </FieldDescription>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={saving === "activation"}
                    onClick={() => void sendActivation()}
                  >
                    <Mail data-icon="inline-start" />
                    {t("emailVerification.action")}
                  </Button>
                </div>
              ) : null}
            </Field>
            <Field>
              <FieldLabel htmlFor="settings-nickname">
                {t("profile.displayName")}
              </FieldLabel>
              <Input
                id="settings-nickname"
                value={profile.nickname}
                onChange={(event) => update("nickname", event.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel>{t("profile.language")}</FieldLabel>
              <Select
                value={profile.locale}
                onValueChange={(value) => update("locale", value)}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {["zh", "en", "ja", "it"].map((locale) => (
                      <SelectItem key={locale} value={locale}>
                        {t(`locales.${locale}`)}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel htmlFor="settings-website-name">
                {t("profile.websiteName")}
              </FieldLabel>
              <Input
                id="settings-website-name"
                value={profile.websiteName}
                onChange={(event) => update("websiteName", event.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="settings-website">
                {t("profile.website")}
              </FieldLabel>
              <Input
                id="settings-website"
                type="url"
                placeholder="https://example.com"
                value={profile.website}
                onChange={(event) => update("website", event.target.value)}
              />
            </Field>
          </div>
          <Field>
            <FieldLabel htmlFor="settings-bio">{t("profile.bio")}</FieldLabel>
            <Textarea
              id="settings-bio"
              className="min-h-24"
              value={profile.bio}
              onChange={(event) => update("bio", event.target.value)}
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="settings-signature">
              {t("profile.signature")}
            </FieldLabel>
            <Textarea
              id="settings-signature"
              className="min-h-20"
              value={profile.signature}
              onChange={(event) => update("signature", event.target.value)}
            />
          </Field>
          <FieldSeparator />
          <Field>
            <div className="flex items-center gap-2">
              <LinkIcon className="size-4 text-muted-foreground" />
              <FieldLabel>{t("profile.social")}</FieldLabel>
            </div>
            <div className="grid gap-3 lg:grid-cols-2">
              {socialKeys.map((key) => (
                <Field key={key}>
                  <FieldLabel htmlFor={`settings-social-${key}`}>
                    {key === "linkedIn"
                      ? "LinkedIn"
                      : key[0].toUpperCase() + key.slice(1)}
                  </FieldLabel>
                  <Input
                    id={`settings-social-${key}`}
                    type="url"
                    value={profile.externalInformation[key]?.link || ""}
                    onChange={(event) =>
                      setProfile((current) => ({
                        ...current,
                        externalInformation: {
                          ...current.externalInformation,
                          [key]: { link: event.target.value },
                        },
                      }))
                    }
                  />
                </Field>
              ))}
            </div>
          </Field>
          <div>
            <Button
              type="button"
              disabled={saving === "profile"}
              onClick={() =>
                void run(
                  "profile",
                  () => runtime.api.users.saveInfo(profile),
                  t("status.profileSaved"),
                )
              }
            >
              {saving === "profile" ? t("savingShort") : t("profile.save")}
            </Button>
          </div>
        </FieldGroup>
      </div>
    </section>
  );

  function update<K extends keyof SaveUserInfoInput>(
    key: K,
    value: SaveUserInfoInput[K],
  ) {
    setProfile((current) => ({ ...current, [key]: value }));
  }
}
