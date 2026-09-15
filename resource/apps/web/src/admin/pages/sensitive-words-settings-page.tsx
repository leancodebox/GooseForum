import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type {
  GooseAdminApi,
  SensitiveWord,
  SensitiveWordSettings,
} from "@gooseforum/client";
import { Button } from "@gooseforum/ui/components/button";
import {
  Field,
  FieldDescription,
  FieldLabel,
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
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Switch } from "@gooseforum/ui/components/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@gooseforum/ui/components/table";
import { Pencil, Plus, RefreshCw, Save, Search, Trash2, X } from "lucide-react";
import { toast } from "sonner";
import type { ModerationSettingsTextKey } from "../moderation-settings-i18n";
type Text = (k: ModerationSettingsTextKey) => string;
const blank = (): SensitiveWord => ({
  id: 0,
  word: "",
  action: "reject",
  replacement: "",
  enabled: true,
});
export function SensitiveWordsSettingsPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text: Text;
}) {
  const [settings, setSettings] = useState<SensitiveWordSettings>({
    enabled: false,
    mode: "after_review",
  });
  const [words, setWords] = useState<SensitiveWord[]>([]);
  const [editor, setEditor] = useState(blank);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const tr = useRef(text);
  useEffect(() => {
    tr.current = text;
  }, [text]);
  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [a, b] = await Promise.all([
        api.settings.sensitiveWordSettings(),
        api.settings.sensitiveWords(),
      ]);
      setSettings(a);
      setWords(b);
    } catch (r) {
      toast.error(r instanceof Error ? r.message : tr.current("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [api]);
  useEffect(() => {
    void load();
  }, [load]);
  const filtered = useMemo(
    () =>
      words.filter((x) => `${x.word} ${x.replacement}`.includes(search.trim())),
    [search, words],
  );
  const visible = filtered.slice((page - 1) * 50, page * 50);
  async function saveSettings() {
    setSaving(true);
    try {
      await api.settings.saveSensitiveWordSettings(settings);
      toast.success(text("saved"));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    } finally {
      setSaving(false);
    }
  }
  async function saveWord() {
    if (!editor.word.trim()) return;
    setSaving(true);
    try {
      const saved = await api.settings.saveSensitiveWord({
        ...editor,
        word: editor.word.trim(),
      });
      setWords((current) => {
        const i = current.findIndex((x) => x.id === saved.id);
        return i < 0
          ? [...current, saved]
          : current.map((x) => (x.id === saved.id ? saved : x));
      });
      setEditor(blank());
      toast.success(text("saved"));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    } finally {
      setSaving(false);
    }
  }
  async function toggle(word: SensitiveWord, enabled: boolean) {
    setSaving(true);
    try {
      const saved = await api.settings.saveSensitiveWord({ ...word, enabled });
      setWords(words.map((x) => (x.id === saved.id ? saved : x)));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    } finally {
      setSaving(false);
    }
  }
  async function remove(id: number) {
    setSaving(true);
    try {
      await api.settings.deleteSensitiveWord(id);
      setWords(words.filter((x) => x.id !== id));
      if (editor.id === id) setEditor(blank());
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    } finally {
      setSaving(false);
    }
  }
  return (
    <main className="flex flex-1 flex-col gap-4 px-3 py-3 lg:px-4">
      <header className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">{text("sensitive")}</h2>
          <p className="text-xs text-muted-foreground">
            {text("sensitiveHint")}
          </p>
        </div>
        <Button
          size="sm"
          disabled={loading || saving}
          onClick={() => void saveSettings()}
        >
          <Save data-icon="inline-start" />
          {text("save")}
        </Button>
      </header>
      <section className="flex flex-wrap items-center gap-6 rounded-lg border p-3">
        <Field orientation="horizontal" className="w-auto">
          <Switch
            checked={settings.enabled}
            disabled={loading}
            onCheckedChange={(enabled) => setSettings({ ...settings, enabled })}
          />
          <FieldLabel>{text("enabled")}</FieldLabel>
        </Field>
        <Field orientation="horizontal" className="w-auto">
          <FieldLabel>{text("mode")}</FieldLabel>
          <Select
            value={settings.mode}
            onValueChange={(mode) =>
              setSettings({
                ...settings,
                mode: mode as SensitiveWordSettings["mode"],
              })
            }
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="after_review">
                  {text("afterReview")}
                </SelectItem>
                <SelectItem value="visible_then_review">
                  {text("visibleThenReview")}
                </SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </Field>
        <FieldDescription>{text("asyncHint")}</FieldDescription>
      </section>
      <section className="overflow-hidden rounded-lg border">
        <div className="flex flex-col gap-2 border-b bg-muted/20 p-2 md:flex-row md:items-center md:justify-between">
          <div className="flex items-center gap-3">
            <span className="font-medium">
              {text("dictionary")} {filtered.length}
            </span>
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  setPage(1);
                }}
                className="h-8 pl-8"
                placeholder={text("search")}
              />
            </div>
          </div>
          <div className="flex items-center gap-1">
            <Button
              variant="outline"
              size="sm"
              disabled={loading || saving}
              onClick={() => void load()}
            >
              <RefreshCw data-icon="inline-start" />
              {text("refresh")}
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={() => setPage(page - 1)}
            >
              {text("previous")}
            </Button>
            <span>{page}</span>
            <Button
              variant="outline"
              size="sm"
              disabled={page * 50 >= filtered.length}
              onClick={() => setPage(page + 1)}
            >
              {text("next")}
            </Button>
          </div>
        </div>
        <div className="grid items-end gap-2 border-b p-3 md:grid-cols-[minmax(0,2fr)_140px_minmax(0,2fr)_auto]">
          <F
            label={text("word")}
            value={editor.word}
            onChange={(word) => setEditor({ ...editor, word })}
          />
          <Field>
            <FieldLabel>{text("action")}</FieldLabel>
            <Select
              value={editor.action}
              onValueChange={(action) =>
                setEditor({
                  ...editor,
                  action: action as SensitiveWord["action"],
                })
              }
            >
              <SelectTrigger className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="reject">{text("reject")}</SelectItem>
                  <SelectItem value="replace">{text("replace")}</SelectItem>
                  <SelectItem value="record">{text("record")}</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
          <F
            label={text("replacement")}
            value={editor.replacement}
            disabled={editor.action !== "replace"}
            onChange={(replacement) => setEditor({ ...editor, replacement })}
          />
          <div className="flex gap-1">
            <Button
              disabled={saving || !editor.word.trim()}
              onClick={() => void saveWord()}
            >
              {editor.id ? (
                <Save data-icon="inline-start" />
              ) : (
                <Plus data-icon="inline-start" />
              )}
              {editor.id ? text("edit") : text("add")}
            </Button>
            {editor.id ? (
              <Button variant="ghost" onClick={() => setEditor(blank())}>
                <X />
              </Button>
            ) : null}
          </div>
        </div>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{text("word")}</TableHead>
              <TableHead className="w-28">{text("action")}</TableHead>
              <TableHead>{text("replacement")}</TableHead>
              <TableHead className="w-20">{text("enabled")}</TableHead>
              <TableHead className="w-20" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={5} className="h-24 text-center">
                  <Spinner />
                </TableCell>
              </TableRow>
            ) : visible.length ? (
              visible.map((word) => (
                <TableRow key={word.id}>
                  <TableCell className="break-all font-medium">
                    {word.word}
                  </TableCell>
                  <TableCell>{text(word.action)}</TableCell>
                  <TableCell className="break-all font-mono text-sm">
                    {word.action === "replace" ? word.replacement || "—" : "—"}
                  </TableCell>
                  <TableCell>
                    <Switch
                      checked={word.enabled}
                      disabled={saving}
                      onCheckedChange={(v) => void toggle(word, v)}
                    />
                  </TableCell>
                  <TableCell>
                    <div className="flex">
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => setEditor({ ...word })}
                      >
                        <Pencil />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => void remove(word.id)}
                      >
                        <Trash2 />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="h-24 text-center text-muted-foreground"
                >
                  {text("empty")}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </section>
    </main>
  );
}
function F({
  label,
  value,
  onChange,
  disabled = false,
}: {
  label: string;
  value: string;
  onChange(v: string): void;
  disabled?: boolean;
}) {
  return (
    <Field>
      <FieldLabel>{label}</FieldLabel>
      <Input
        value={value}
        disabled={disabled}
        onChange={(e) => onChange(e.target.value)}
      />
    </Field>
  );
}
