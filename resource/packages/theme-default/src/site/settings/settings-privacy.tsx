import type { PrivacySettingsPayload } from "@gooseforum/client";
import { Shield } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Checkbox } from "@gooseforum/ui/components/checkbox";
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@gooseforum/ui/components/field";
import { SettingsSectionHeader } from "./settings-section-header";

export function PrivacySettings({
  settings,
  saving,
  onChange,
}: {
  settings: PrivacySettingsPayload;
  saving: boolean;
  onChange(key: keyof PrivacySettingsPayload, checked: boolean): void;
}) {
  const { t } = useTranslation("settings");
  return (
    <section>
      <SettingsSectionHeader icon={Shield} title={t("privacy.title")} />
      <FieldGroup className="max-w-2xl gap-0 p-4">
        {(["showTopics", "showActivity", "showFollowing"] as const).map(
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
                disabled={saving}
                checked={settings[key]}
                onCheckedChange={(checked) => onChange(key, checked === true)}
              />
            </Field>
          ),
        )}
      </FieldGroup>
    </section>
  );
}
