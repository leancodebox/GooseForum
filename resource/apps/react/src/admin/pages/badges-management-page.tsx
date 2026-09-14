import { useCallback, useEffect, useState } from "react";
import type { AdminBadge, GooseAdminApi } from "@gooseforum/client";
import { Badge } from "@gooseforum/react/components/ui/badge";
import { Button } from "@gooseforum/react/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@gooseforum/react/components/ui/dialog";
import {
  Empty,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/react/components/ui/empty";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@gooseforum/react/components/ui/field";
import { Input } from "@gooseforum/react/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@gooseforum/react/components/ui/select";
import { Spinner } from "@gooseforum/react/components/ui/spinner";
import { Switch } from "@gooseforum/react/components/ui/switch";
import { Textarea } from "@gooseforum/react/components/ui/textarea";
import { Edit3, Plus, RefreshCw, Trash2 } from "lucide-react";
import { toast } from "sonner";
import type { AssetTextKey } from "../assets-i18n";
type Text = (key: AssetTextKey) => string;
const colors = [
  "blue",
  "emerald",
  "teal",
  "sky",
  "cyan",
  "rose",
  "violet",
  "purple",
  "fuchsia",
  "indigo",
  "amber",
  "orange",
  "yellow",
  "slate",
];
const emptyBadge: AdminBadge = {
  code: "",
  type: "custom",
  grantMode: "manual",
  name: "",
  description: "",
  iconType: "asset",
  iconKey: "",
  iconUrl: "/static/badges/contributor.svg",
  color: "blue",
  level: "bronze",
  isEnabled: true,
  isWearable: true,
  sortOrder: 1000,
};
export function BadgesManagementPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text: Text;
}) {
  const [items, setItems] = useState<AdminBadge[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [editing, setEditing] = useState<AdminBadge | null>(null);
  const [deleting, setDeleting] = useState<AdminBadge | null>(null);
  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setItems(await api.assets.badges());
    } catch (reason) {
      setError(message(reason, text("loadFailed")));
    } finally {
      setLoading(false);
    }
  }, [api, text]);
  useEffect(() => {
    void load();
  }, [load]);
  return (
    <main className="flex flex-1 flex-col gap-3 px-3 py-3 lg:px-4">
      <header className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">{text("badges")}</h2>
          <p className="text-xs text-muted-foreground">{text("badgesHint")}</p>
        </div>
        <Button size="sm" onClick={() => setEditing({ ...emptyBadge })}>
          <Plus data-icon="inline-start" />
          {text("newBadge")}
        </Button>
      </header>
      <div className="flex items-center justify-between">
        <div className="flex gap-2">
          <Badge variant="secondary">
            {text("system")} {items.filter((x) => x.type === "system").length}
          </Badge>
          <Badge variant="outline">
            {text("custom")} {items.filter((x) => x.type === "custom").length}
          </Badge>
        </div>
        <Button
          variant="outline"
          size="sm"
          disabled={loading}
          onClick={() => void load()}
        >
          <RefreshCw
            data-icon="inline-start"
            className={loading ? "animate-spin" : undefined}
          />
          {text("refresh")}
        </Button>
      </div>
      {error ? (
        <div className="rounded-lg border border-destructive/30 p-3 text-destructive">
          {error}
        </div>
      ) : loading && !items.length ? (
        <AssetEmpty title={text("loading")} icon={<Spinner />} />
      ) : !items.length ? (
        <AssetEmpty title={text("empty")} icon={<Edit3 />} />
      ) : (
        <div className="grid grid-cols-[repeat(auto-fill,minmax(5.5rem,1fr))] gap-3">
          {items.map((item) => (
            <div
              key={item.code}
              className="group relative flex min-w-0 flex-col items-center rounded-md px-1 py-1.5 hover:bg-muted/60"
            >
              <Button
                variant="ghost"
                className="h-auto min-w-0 flex-col p-0 font-normal hover:bg-transparent"
                onClick={() => setEditing({ ...item })}
              >
                <span
                  className="flex size-12 items-center justify-center bg-muted"
                  style={{
                    clipPath:
                      "polygon(25% 5%, 75% 5%, 100% 50%, 75% 95%, 25% 95%, 0 50%)",
                  }}
                >
                  <img
                    src={item.iconUrl || emptyBadge.iconUrl}
                    alt={item.name}
                    className="size-6 object-contain"
                  />
                </span>
                <span className="mt-1 max-w-full truncate text-xs font-semibold">
                  {item.name}
                </span>
                <span className="max-w-full truncate text-[10px] text-muted-foreground">
                  {item.grantMode === "auto" ? text("auto") : text("manual")} ·{" "}
                  {item.isEnabled ? item.level : text("disabled")}
                </span>
              </Button>
              <div className="absolute right-0.5 top-0.5 flex gap-0.5 rounded-md bg-background/90 p-0.5 opacity-0 ring-1 ring-border group-focus-within:opacity-100 group-hover:opacity-100">
                <Button
                  variant="ghost"
                  size="icon-xs"
                  onClick={() => setEditing({ ...item })}
                >
                  <Edit3 />
                </Button>
                {item.type !== "system" ? (
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => setDeleting(item)}
                  >
                    <Trash2 />
                  </Button>
                ) : null}
              </div>
            </div>
          ))}
        </div>
      )}
      <BadgeEditor
        value={editing}
        api={api}
        text={text}
        onClose={() => setEditing(null)}
        onSaved={load}
      />
      <DeleteBadge
        value={deleting}
        api={api}
        text={text}
        onClose={() => setDeleting(null)}
        onDeleted={load}
      />
    </main>
  );
}
function BadgeEditor({
  value,
  api,
  text,
  onClose,
  onSaved,
}: {
  value: AdminBadge | null;
  api: GooseAdminApi;
  text: Text;
  onClose(): void;
  onSaved(): Promise<void>;
}) {
  const [form, setForm] = useState(emptyBadge);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => {
    if (value) {
      setForm({ ...emptyBadge, ...value });
      setError("");
    }
  }, [value]);
  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!form.name.trim()) {
      setError(text("required"));
      return;
    }
    setSaving(true);
    try {
      await api.assets.saveBadge({
        ...form,
        name: form.name.trim(),
        code: form.code.trim(),
        color: colors.includes(form.color) ? form.color : "blue",
        sortOrder: Number(form.sortOrder) || 1000,
      });
      toast.success(text("saved"));
      onClose();
      await onSaved();
    } catch (r) {
      setError(message(r, text("saveFailed")));
    } finally {
      setSaving(false);
    }
  }
  return (
    <Dialog
      open={Boolean(value)}
      onOpenChange={(open) => !open && !saving && onClose()}
    >
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {value?.code ? text("edit") : text("newBadge")}
          </DialogTitle>
          <DialogDescription>{text("badgesHint")}</DialogDescription>
        </DialogHeader>
        <form id="badge-form" onSubmit={submit}>
          <FieldGroup>
            <div className="grid gap-3 sm:grid-cols-2">
              <Field>
                <FieldLabel>{text("code")}</FieldLabel>
                <Input
                  value={form.code}
                  disabled={Boolean(value?.code)}
                  onChange={(e) => setForm({ ...form, code: e.target.value })}
                />
              </Field>
              <Field data-invalid={Boolean(error)}>
                <FieldLabel>{text("name")}</FieldLabel>
                <Input
                  aria-invalid={Boolean(error)}
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                />
                <FieldError>{error}</FieldError>
              </Field>
            </div>
            <Field>
              <FieldLabel>{text("description")}</FieldLabel>
              <Textarea
                value={form.description}
                onChange={(e) =>
                  setForm({ ...form, description: e.target.value })
                }
              />
            </Field>
            <div className="grid gap-3 sm:grid-cols-2">
              <Field>
                <FieldLabel>{text("icon")}</FieldLabel>
                <Input
                  value={form.iconUrl}
                  onChange={(e) =>
                    setForm({ ...form, iconUrl: e.target.value })
                  }
                />
              </Field>
              <Field>
                <FieldLabel>{text("color")}</FieldLabel>
                <Select
                  value={form.color}
                  onValueChange={(color) => setForm({ ...form, color })}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      {colors.map((color) => (
                        <SelectItem key={color} value={color}>
                          {color}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel>{text("level")}</FieldLabel>
                <Input
                  value={form.level}
                  onChange={(e) => setForm({ ...form, level: e.target.value })}
                />
              </Field>
              <Field>
                <FieldLabel>{text("sort")}</FieldLabel>
                <Input
                  type="number"
                  value={form.sortOrder}
                  onChange={(e) =>
                    setForm({ ...form, sortOrder: Number(e.target.value) })
                  }
                />
              </Field>
              <Field>
                <FieldLabel>{text("grantMode")}</FieldLabel>
                <Select
                  value={form.grantMode}
                  disabled={form.type === "system"}
                  onValueChange={(grantMode) => setForm({ ...form, grantMode })}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem value="auto">{text("auto")}</SelectItem>
                      <SelectItem value="manual">{text("manual")}</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>
              <Field orientation="horizontal">
                <FieldLabel>{text("enabled")}</FieldLabel>
                <Switch
                  checked={form.isEnabled}
                  onCheckedChange={(isEnabled) =>
                    setForm({ ...form, isEnabled })
                  }
                />
              </Field>
              <Field orientation="horizontal">
                <FieldLabel>{text("wearable")}</FieldLabel>
                <Switch
                  checked={form.isWearable}
                  onCheckedChange={(isWearable) =>
                    setForm({ ...form, isWearable })
                  }
                />
              </Field>
            </div>
          </FieldGroup>
        </form>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {text("cancel")}
          </Button>
          <Button form="badge-form" type="submit" disabled={saving}>
            {saving ? <Spinner data-icon="inline-start" /> : null}
            {text("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
function DeleteBadge({
  value,
  api,
  text,
  onClose,
  onDeleted,
}: {
  value: AdminBadge | null;
  api: GooseAdminApi;
  text: Text;
  onClose(): void;
  onDeleted(): Promise<void>;
}) {
  const [saving, setSaving] = useState(false);
  async function remove() {
    if (!value) return;
    setSaving(true);
    try {
      await api.assets.deleteBadge(value.code);
      toast.success(text("saved"));
      onClose();
      await onDeleted();
    } catch (r) {
      toast.error(message(r, text("saveFailed")));
    } finally {
      setSaving(false);
    }
  }
  return (
    <Dialog
      open={Boolean(value)}
      onOpenChange={(open) => !open && !saving && onClose()}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{text("deleteTitle")}</DialogTitle>
          <DialogDescription>{value?.name}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {text("cancel")}
          </Button>
          <Button
            variant="destructive"
            disabled={saving}
            onClick={() => void remove()}
          >
            {text("delete")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
function AssetEmpty({ title, icon }: { title: string; icon: React.ReactNode }) {
  return (
    <Empty className="min-h-48 border">
      <EmptyHeader>
        <EmptyMedia variant="icon">{icon}</EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
      </EmptyHeader>
    </Empty>
  );
}
function message(r: unknown, f: string) {
  return r instanceof Error && r.message ? r.message : f;
}
