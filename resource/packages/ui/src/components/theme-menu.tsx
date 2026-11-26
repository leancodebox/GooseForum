import { MoonIcon, SunIcon } from "lucide-react";
import { Button } from "@gooseforum/ui/components/button";
import { RoutineIcon } from "@gooseforum/ui/components/icons/material-symbols/routine-icon";
import { DropdownMenuRadioGroup, DropdownMenuRadioItem, DropdownMenuTrigger } from "@gooseforum/ui/components/dropdown-menu";

export type ThemePreference = "system" | "gf-light" | "gf-dark";

export function ThemeMenuTrigger({ size = "icon", preference, label }: { size?: "icon" | "icon-lg"; preference: ThemePreference; label: string }) {
  return (
    <DropdownMenuTrigger asChild>
      <Button
        variant="ghost"
        size={size}
        className={size === "icon-lg" ? "[&_svg]:size-5" : undefined}
        aria-label={label}
      >
        {preference === "system" ? (
          <RoutineIcon />
        ) : preference === "gf-light" ? (
          <SunIcon />
        ) : (
          <MoonIcon />
        )}
      </Button>
    </DropdownMenuTrigger>
  );
}

export function ThemeMenuOptions({ preference, onChange, t }: { preference: ThemePreference; onChange(value: ThemePreference): void; t(key: string): string }) {
  return (
    <DropdownMenuRadioGroup
      value={preference}
      onValueChange={(value) => {
        if (value === "system" || value === "gf-light" || value === "gf-dark")
          onChange(value);
      }}
    >
      <DropdownMenuRadioItem value="system">{t("followSystem")}</DropdownMenuRadioItem>
      <DropdownMenuRadioItem value="gf-light">{t("lightTheme")}</DropdownMenuRadioItem>
      <DropdownMenuRadioItem value="gf-dark">{t("darkTheme")}</DropdownMenuRadioItem>
    </DropdownMenuRadioGroup>
  );
}
