import { MoonIcon, SunIcon } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useGooseRuntime } from "@gooseforum/runtime";
import { Button } from "@gooseforum/ui/components/button";
import { RoutineIcon } from "@gooseforum/ui/components/icons/material-symbols/routine-icon";
import { DropdownMenuRadioGroup, DropdownMenuRadioItem, DropdownMenuTrigger } from "@gooseforum/ui/components/dropdown-menu";

export function ThemeMenuTrigger({ size = "icon" }: { size?: "icon" | "icon-lg" }) {
  const runtime = useGooseRuntime();
  const { t } = useTranslation("shell");
  return (
    <DropdownMenuTrigger asChild>
      <Button
        variant="ghost"
        size={size}
        className={size === "icon-lg" ? "[&_svg]:size-5" : undefined}
        aria-label={t("chooseTheme")}
      >
        {(!runtime.themePreference || runtime.themePreference === "system") ? (
          <RoutineIcon />
        ) : runtime.themePreference === "gf-light" ? (
          <SunIcon />
        ) : (
          <MoonIcon />
        )}
      </Button>
    </DropdownMenuTrigger>
  );
}

export function ThemeMenuOptions() {
  const runtime = useGooseRuntime();
  const { t } = useTranslation("shell");
  return (
    <DropdownMenuRadioGroup
      value={runtime.themePreference ?? "system"}
      onValueChange={(value) => {
        if (value === "system" || value === "gf-light" || value === "gf-dark")
          runtime.setThemePreference?.(value);
      }}
    >
      <DropdownMenuRadioItem value="system">{t("followSystem")}</DropdownMenuRadioItem>
      <DropdownMenuRadioItem value="gf-light">{t("lightTheme")}</DropdownMenuRadioItem>
      <DropdownMenuRadioItem value="gf-dark">{t("darkTheme")}</DropdownMenuRadioItem>
    </DropdownMenuRadioGroup>
  );
}
