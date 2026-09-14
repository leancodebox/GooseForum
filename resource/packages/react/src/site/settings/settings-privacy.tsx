import { useState } from "react";
import { Shield } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Checkbox } from "../../components/ui/checkbox";
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "../../components/ui/field";
import { SettingsSectionHeader } from "./settings-section-header";

type PrivacyState = {
  showTopics: boolean;
  showFollowing: boolean;
  emailNotifications: boolean;
};
const defaults: PrivacyState = {
  showTopics: true,
  showFollowing: true,
  emailNotifications: true,
};

export function PrivacySettings({
  showStatus,
  showError,
}: {
  showStatus(message: string): void;
  showError(message: string): void;
}) {
  const { t } = useTranslation("settings");
  const [settings, setSettings] = useState<PrivacyState>(readPrivacy);
  function update(key: keyof PrivacyState, checked: boolean) {
    const next = { ...settings, [key]: checked };
    setSettings(next);
    try {
      localStorage.setItem("goose-privacy-settings", JSON.stringify(next));
      showStatus(t("status.privacySaved"));
    } catch {
      showError(t("errors.privacy"));
    }
  }
  return (
    <section>
      <SettingsSectionHeader icon={Shield} title={t("privacy.title")} />
      <FieldGroup className="max-w-2xl gap-0 p-4">
        {(["showTopics", "showFollowing", "emailNotifications"] as const).map(
          (key) => (
            <Field
              key={key}
              orientation="horizontal"
              className="border-b py-4 last:border-b-0"
            >
              <FieldContent>
                <FieldLabel htmlFor={`privacy-${key}`}>
                  {t(`privacy.${key}`)}
                </FieldLabel>
                <FieldDescription>
                  {t(`privacy.${key}Description`)}
                </FieldDescription>
              </FieldContent>
              <Checkbox
                id={`privacy-${key}`}
                checked={settings[key]}
                onCheckedChange={(checked) => update(key, checked === true)}
              />
            </Field>
          ),
        )}
      </FieldGroup>
    </section>
  );
}

function readPrivacy(): PrivacyState {
  try {
    const value = JSON.parse(
      localStorage.getItem("goose-privacy-settings") || "null",
    );
    return value && typeof value === "object"
      ? { ...defaults, ...value }
      : defaults;
  } catch {
    return defaults;
  }
}
