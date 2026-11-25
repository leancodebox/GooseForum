import { lazy, Suspense, useId, useState, type ReactNode } from "react";
import { ChevronRight } from "lucide-react";
import { Button } from "@gooseforum/ui/components/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@gooseforum/ui/components/dialog";
import { Field, FieldLabel } from "@gooseforum/ui/components/field";
import type { SettingsTextKey } from "../settings-i18n";

const NavigationIconPicker = lazy(() => import("./navigation-icon-picker"));

export function NavigationIconField({
  label, value, displayValue = value, preview, onSelect, text,
}: {
  label: string;
  value?: string;
  displayValue?: string;
  preview?: ReactNode;
  onSelect(name: string): void;
  text(key: SettingsTextKey): string;
}) {
  const id = useId();
  const [open, setOpen] = useState(false);

  return (
    <Field className="min-w-0">
      <FieldLabel htmlFor={id}>{label}</FieldLabel>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogTrigger asChild>
          <Button id={id} type="button" variant="outline" className="w-full min-w-0 justify-start">
            {preview}
            <span className="min-w-0 truncate">{displayValue || text("defaultMenuIcon")}</span>
            <ChevronRight aria-hidden="true" className="ms-auto" />
          </Button>
        </DialogTrigger>
        <DialogContent aria-describedby={undefined} className="flex max-h-[calc(100dvh-2rem)] flex-col overflow-hidden sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{label}</DialogTitle>
          </DialogHeader>
          <Suspense fallback={<p role="status">{text("loading")}</p>}>
            <NavigationIconPicker value={value} text={text} onSelect={name => {
              onSelect(name);
              setOpen(false);
            }} />
          </Suspense>
        </DialogContent>
      </Dialog>
    </Field>
  );
}
