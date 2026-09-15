import { useEffect, useId, useState } from "react";
import { Check, Palette, RotateCcw } from "lucide-react";
import { Button } from "./button";
import { Input } from "./input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./tabs";
import {
  Dialog,
  DialogContent,
  DialogTrigger,
  DialogTitle,
  DialogDescription,
} from "./dialog";
import {
  colorPalette,
  cssColorHex,
  hexChannels,
  rgbHex,
} from "../lib/color";
import { cn } from "../lib/utils";

const copy = {
  zh: {
    palette: "调色板",
    recent: "最近使用",
    channels: "RGB 调节",
    value: "颜色值",
    invalid: "请输入有效的不透明颜色，如 HEX、RGB、HSL 或 OKLCH",
    reset: "还原打开前颜色",
    fine: "精细取色",
    done: "完成",
    previous: "原颜色",
    current: "当前颜色",
  },
  en: {
    palette: "Palette",
    recent: "Recent colors",
    channels: "RGB channels",
    value: "Color value",
    invalid: "Enter an opaque HEX, RGB, HSL or OKLCH color",
    reset: "Restore original color",
    fine: "Fine color selection",
    done: "Done",
    previous: "Original",
    current: "Current",
  },
  ja: {
    palette: "パレット",
    recent: "最近の色",
    channels: "RGB 調整",
    value: "色の値",
    invalid: "有効な不透明色を入力してください",
    reset: "元の色に戻す",
    fine: "色を選択",
    done: "完了",
    previous: "元の色",
    current: "現在の色",
  },
  it: {
    palette: "Tavolozza",
    recent: "Colori recenti",
    channels: "Canali RGB",
    value: "Valore colore",
    invalid: "Inserisci un colore opaco HEX, RGB, HSL o OKLCH",
    reset: "Ripristina colore",
    fine: "Seleziona colore",
    done: "Fine",
    previous: "Originale",
    current: "Attuale",
  },
};
const storageKey = "goose:recent-colors";
function readRecent(): string[] {
  try {
    const value: unknown = JSON.parse(localStorage.getItem(storageKey) || "[]");
    return Array.isArray(value)
      ? value
          .filter(
            (item): item is string =>
              typeof item === "string" && !!hexChannels(item),
          )
          .slice(0, 10)
      : [];
  } catch {
    return [];
  }
}

export function ColorPicker({
  value,
  onChange,
  label,
  id,
  disabled = false,
  locale,
  className,
}: {
  value: string;
  onChange(value: string): void;
  label: string;
  id?: string;
  disabled?: boolean;
  locale?: string;
  className?: string;
}) {
  const language =
    locale ||
    (typeof document !== "undefined" ? document.documentElement.lang : "zh");
  const t = copy[language.split("-")[0] as keyof typeof copy] || copy.en;
  const fieldId = useId();
  const [open, setOpen] = useState(false);
  const [original, setOriginal] = useState(value);
  const [draft, setDraft] = useState(value);
  const [hex, setHex] = useState("#ffffff");
  const [invalid, setInvalid] = useState(false);
  const [recent, setRecent] = useState<string[]>([]);
  useEffect(() => {
    setDraft(value);
    setHex(cssColorHex(value) || "#ffffff");
    setInvalid(false);
  }, [value]);
  const channels = hexChannels(hex)!;
  function choose(next: string) {
    setHex(next);
    setDraft(next);
    setInvalid(false);
    onChange(next);
  }
  function remember(next: string) {
    const colors = [
      next,
      ...readRecent().filter((color) => color !== next),
    ].slice(0, 10);
    setRecent(colors);
    try {
      localStorage.setItem(storageKey, JSON.stringify(colors));
    } catch {}
  }
  function changeOpen(next: boolean) {
    if (next) {
      setOriginal(value);
      setDraft(value);
      setHex(cssColorHex(value) || "#ffffff");
      setRecent(readRecent());
      setInvalid(false);
    } else {
      const selected = cssColorHex(value);
      if (selected) remember(selected);
    }
    setOpen(next);
  }
  function applyDraft() {
    const next = cssColorHex(draft.trim());
    if (!next) {
      setInvalid(true);
      return false;
    }
    choose(next);
    return true;
  }
  return (
    <Dialog open={open} onOpenChange={changeOpen}>
      <DialogTrigger asChild>
        <Button
          id={id}
          type="button"
          variant="outline"
          disabled={disabled}
          aria-label={label}
          className={cn("justify-start", className)}
        >
          <span
            className="size-5 shrink-0 rounded border border-border/80"
            style={{ backgroundColor: value }}
          />
          <span className="min-w-0 truncate font-mono text-xs">{value}</span>
          <Palette data-icon="inline-end" className="ml-auto" />
        </Button>
      </DialogTrigger>
      <DialogContent
        className="max-h-[85dvh] overflow-y-auto p-4 sm:max-w-lg sm:p-5"
        showCloseButton={false}
      >
        <DialogDescription className="sr-only">
          {t.palette} · {t.channels} · {t.value}
        </DialogDescription>
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between gap-2">
            <DialogTitle>{label}</DialogTitle>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label={t.reset}
              onClick={() => {
                onChange(original);
                setDraft(original);
                setHex(cssColorHex(original) || "#ffffff");
                setInvalid(false);
              }}
            >
              <RotateCcw />
            </Button>
          </div>
          <div className="grid grid-cols-2 overflow-hidden rounded-md border">
            {[
              ["previous", original],
              ["current", value],
            ].map(([key, color]) => (
              <div key={key} className="flex flex-col">
                <span className="h-8" style={{ backgroundColor: color }} />
                <span className="bg-muted px-2 py-1 text-center text-xs text-muted-foreground">
                  {t[key as "previous" | "current"]}
                </span>
              </div>
            ))}
          </div>
          <Tabs defaultValue="palette" className="gap-3">
            <TabsList className="w-full">
              <TabsTrigger value="palette">{t.palette}</TabsTrigger>
              <TabsTrigger value="channels">{t.channels}</TabsTrigger>
            </TabsList>
            <TabsContent value="palette" className="flex flex-col gap-3">
              <div className="flex flex-col gap-1.5">
                <div className="grid grid-cols-10 gap-1">
                  {colorPalette.flat().map((color) => (
                    <button
                      type="button"
                      key={color}
                      aria-label={color}
                      aria-pressed={hex === color}
                      title={color}
                      className="relative h-5 w-full rounded-sm border border-foreground/10 outline-none hover:scale-110 focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1"
                      style={{ backgroundColor: color }}
                      onClick={() => choose(color)}
                    >
                      {hex === color ? (
                        <Check className="absolute inset-0 m-auto size-3.5 text-white drop-shadow-[0_1px_1px_black]" />
                      ) : null}
                    </button>
                  ))}
                </div>
              </div>
              {recent.length ? (
                <div className="flex flex-col gap-1.5">
                  <span className="text-xs font-medium text-muted-foreground">
                    {t.recent}
                  </span>
                  <div className="grid grid-cols-10 gap-1">
                    {recent.map((color) => (
                      <button
                        type="button"
                        key={color}
                        title={color}
                        aria-label={`${t.recent} ${color}`}
                        className="aspect-square rounded-sm border border-foreground/10 focus-visible:ring-2 focus-visible:ring-ring"
                        style={{ backgroundColor: color }}
                        onClick={() => choose(color)}
                      />
                    ))}
                  </div>
                </div>
              ) : null}
            </TabsContent>
            <TabsContent value="channels">
              <fieldset className="flex min-h-36 flex-col justify-center gap-4">
                <legend className="mb-2 text-xs font-medium text-muted-foreground">
                  {t.channels}
                </legend>
                {["R", "G", "B"].map((channel, index) => (
                  <label
                    key={channel}
                    className="flex items-center gap-2 text-xs"
                  >
                    <span className="w-3">{channel}</span>
                    <input
                      aria-label={channel}
                      type="range"
                      min={0}
                      max={255}
                      value={channels[index]}
                      className="h-4 min-w-0 flex-1 cursor-pointer accent-primary"
                      onChange={(event) =>
                        choose(
                          rgbHex(
                            channels.map((value, i) =>
                              i === index ? Number(event.target.value) : value,
                            ),
                          ),
                        )
                      }
                    />
                    <span className="w-7 text-right tabular-nums">
                      {channels[index]}
                    </span>
                  </label>
                ))}
              </fieldset>
            </TabsContent>
          </Tabs>
          <div className="flex flex-col gap-1.5">
            <label htmlFor={fieldId} className="text-xs font-medium">
              {t.value}
            </label>
            <Input
              id={fieldId}
              value={draft}
              aria-invalid={invalid}
              aria-describedby={invalid ? `${fieldId}-error` : undefined}
              className="font-mono text-xs"
              onChange={(event) => {
                setDraft(event.target.value);
                setInvalid(false);
              }}
              onBlur={() => {
                if (draft !== value) applyDraft();
              }}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  event.preventDefault();
                  event.stopPropagation();
                  applyDraft();
                }
              }}
            />
            {invalid ? (
              <p
                id={`${fieldId}-error`}
                role="alert"
                className="text-xs text-destructive"
              >
                {t.invalid}
              </p>
            ) : null}
          </div>
          <div className="flex items-center justify-between">
            <label className="flex cursor-pointer items-center gap-2 text-xs text-muted-foreground">
              <input
                type="color"
                value={hex}
                aria-label={t.fine}
                className="size-7 cursor-pointer"
                onChange={(event) => choose(event.target.value)}
              />
              {t.fine}
            </label>
            <Button
              type="button"
              size="sm"
              onClick={() => {
                if (draft === value || applyDraft()) {
                  remember(cssColorHex(draft) || hex);
                  setOpen(false);
                }
              }}
            >
              {t.done}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
