import { useTranslation } from "react-i18next";
import { useGooseRuntime } from "@gooseforum/runtime";
import { ThemeMenuOptions as Options, ThemeMenuTrigger as Trigger } from "@gooseforum/ui/components/theme-menu";

export function ThemeMenuTrigger() {
  const runtime = useGooseRuntime();
  const { t } = useTranslation("shared.shell");
  return <Trigger preference={runtime.themePreference ?? "system"} label={t("chooseTheme")} />;
}

export function ThemeMenuOptions() {
  const runtime = useGooseRuntime();
  const { t } = useTranslation("shared.shell");
  return <Options preference={runtime.themePreference ?? "system"} onChange={value => runtime.setThemePreference?.(value)} t={t} />;
}
