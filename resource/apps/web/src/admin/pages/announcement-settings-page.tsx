import { useCallback, useEffect, useRef, useState, type ReactNode } from "react";
import { closestCenter, DndContext, KeyboardSensor, PointerSensor, useSensor, useSensors } from "@dnd-kit/core";
import { arrayMove, SortableContext, sortableKeyboardCoordinates, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import type { AnnouncementConfig, AnnouncementItemConfig, GooseAdminApi } from "@gooseforum/client";
import { Button } from "@gooseforum/ui/components/button";
import { Field, FieldLabel } from "@gooseforum/ui/components/field";
import { Input } from "@gooseforum/ui/components/input";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Switch } from "@gooseforum/ui/components/switch";
import { ArrowDown, ArrowUp, Code, GripVertical, Plus, Save, Trash2 } from "lucide-react";
import { toast } from "sonner";
import type { ContentSettingsTextKey } from "../content-settings-i18n";
import { AnnouncementMarkdownEditor } from "../components/announcement-markdown-editor";

type Text = (k: ContentSettingsTextKey) => string;
type Form = AnnouncementConfig & { items: AnnouncementItemConfig[] };

function normalize(config: AnnouncementConfig): Form {
  return { ...config, content: "", items: config.items ?? (config.content.trim()
    ? [{ id: "legacy", title: "", content: config.content, enabled: true }] : []) };
}

function SortableAnnouncement({ id, disabled, children, text }: {
  id: string; disabled: boolean; children: ReactNode; text: Text;
}) {
  const { setNodeRef, transform, transition, isDragging, attributes, listeners } = useSortable({ id, disabled });
  return <div ref={setNodeRef}
    style={{ transform: CSS.Transform.toString(transform), transition, opacity: isDragging ? 0.5 : undefined }}
    className="flex items-center gap-0.5 rounded-lg border bg-background p-1">
    <Button type="button" variant="ghost" size="icon-xs" className="touch-none cursor-grab shrink-0" disabled={disabled}
      aria-label={text("announcementReorder")} {...attributes} {...listeners}><GripVertical /></Button>
    {children}
  </div>;
}

export function AnnouncementSettingsPage({ api, text }: { api: GooseAdminApi; text: Text }) {
  const [form, setForm] = useState<Form>({ enabled: false, content: "", items: [] });
  const [selectedId, setSelectedId] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const tr = useRef(text);
  const disabled = loading || saving || uploading;
  const selected = form.items.find(item => item.id === selectedId);
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }));
  useEffect(() => { tr.current = text; }, [text]);
  const load = useCallback(async () => {
    setLoading(true);
    try {
      const next = normalize(await api.settings.announcement());
      setForm(next);
      setSelectedId(next.items[0]?.id ?? "");
    } catch (reason) {
      toast.error(reason instanceof Error ? reason.message : tr.current("loadFailed"));
    } finally { setLoading(false); }
  }, [api]);
  useEffect(() => { void load(); }, [load]);

  function update(id: string, patch: Partial<AnnouncementItemConfig>) {
    setForm(current => ({ ...current, items: current.items.map(item => item.id === id ? { ...item, ...patch } : item) }));
  }
  function add() {
    let timestamp = Date.now();
    while (form.items.some(item => item.id === String(timestamp))) timestamp += 1;
    const item = { id: String(timestamp), title: "", content: "", enabled: true };
    setForm(current => ({ ...current, items: [...current.items, item] }));
    setSelectedId(item.id);
  }
  function remove(index: number) {
    const items = form.items.filter((_, i) => i !== index);
    if (form.items[index].id === selectedId) setSelectedId(items[Math.min(index, items.length - 1)]?.id ?? "");
    setForm(current => ({ ...current, items }));
  }
  function move(from: number, to: number) {
    if (from < 0 || to < 0 || to >= form.items.length || disabled) return;
    setForm(current => ({ ...current, items: arrayMove(current.items, from, to) }));
  }
  async function save() {
    if (disabled) return;
    const empty = form.items.find(item => item.enabled && !item.content.trim());
    if (empty) { setSelectedId(empty.id); toast.error(text("announcementContentRequired")); return; }
    setSaving(true);
    try {
      await api.settings.saveAnnouncement({ enabled: form.enabled, content: "", items: form.items });
      toast.success(text("saved"));
    } catch (reason) { toast.error(reason instanceof Error ? reason.message : text("saveFailed")); }
    finally { setSaving(false); }
  }

  return <main className="flex flex-1 flex-col gap-4 px-3 py-3 lg:px-4">
    <header className="flex items-center justify-between gap-3">
      <div><h2 className="text-lg font-semibold">{text("announcement")}</h2>
        <p className="text-xs text-muted-foreground">{text("announcementOrderHint")}</p></div>
      <div className="flex shrink-0 items-center gap-3">
        <label className="flex items-center gap-2 text-xs text-muted-foreground">
          <span className="hidden sm:inline">{text("announcementEnabled")}</span>
          <Switch aria-label={text("announcementEnabled")} checked={form.enabled} disabled={disabled}
            onCheckedChange={enabled => setForm(current => ({ ...current, enabled }))} />
        </label>
        <Button size="sm" disabled={disabled} onClick={() => void save()}>
          {saving ? <Spinner data-icon="inline-start" /> : <Save data-icon="inline-start" />}{text("save")}
        </Button>
      </div>
    </header>
    <div className="grid max-w-6xl items-start gap-3 lg:grid-cols-[18rem_minmax(0,1fr)]">
      <section className="min-w-0 rounded-xl border bg-muted/20 p-3" aria-label={text("announcementList")}>
        <div className="mb-2 flex items-center justify-between gap-2"><h3 className="text-sm font-semibold">{text("announcementList")}</h3>
          <Button size="sm" variant="outline" disabled={disabled} onClick={add}><Plus data-icon="inline-start" />{text("add")}</Button></div>
        <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={({ active, over }) => {
          if (!over || active.id === over.id) return;
          move(form.items.findIndex(item => item.id === active.id), form.items.findIndex(item => item.id === over.id));
        }}>
          <SortableContext items={form.items.map(item => item.id)} strategy={verticalListSortingStrategy}>
            <div className="space-y-1">
              {form.items.map((item, index) => <SortableAnnouncement key={item.id} id={item.id} disabled={disabled} text={text}>
                  <button type="button" disabled={disabled} aria-pressed={item.id === selectedId}
                    onClick={() => setSelectedId(item.id)}
                    className={`flex min-w-0 flex-1 items-center gap-1.5 rounded-md px-1 py-1.5 text-left text-sm focus-visible:outline-2 focus-visible:outline-ring ${item.id === selectedId ? "bg-primary/10 text-primary" : "hover:bg-muted"}`}>
                    <span className="shrink-0 text-xs tabular-nums">{index + 1}.</span>
                    <span className="truncate font-medium">{item.title || text("announcementUntitled")}</span>
                  </button>
                    <Switch aria-label={`${text("announcementEnabled")} ${index + 1}`} checked={item.enabled} disabled={disabled} onCheckedChange={enabled => update(item.id, { enabled })} />
                    <div className="flex shrink-0 gap-0.5">
                      <Button variant="ghost" size="icon-xs" disabled={disabled || index === 0} aria-label={text("announcementMoveUp")} onClick={() => move(index, index - 1)}><ArrowUp /></Button>
                      <Button variant="ghost" size="icon-xs" disabled={disabled || index === form.items.length - 1} aria-label={text("announcementMoveDown")} onClick={() => move(index, index + 1)}><ArrowDown /></Button>
                      <Button variant="ghost" size="icon-xs" disabled={disabled} aria-label={text("remove")} onClick={() => remove(index)}><Trash2 /></Button>
                    </div>
              </SortableAnnouncement>)}
            </div>
          </SortableContext>
        </DndContext>
        {!form.items.length && !loading ? <p className="py-4 text-center text-xs text-muted-foreground">{text("announcementEmpty")}</p> : null}
      </section>
      <section className="min-w-0 rounded-xl border bg-background p-3">
        {selected ? <div className="space-y-3">
          <Field orientation="horizontal"><FieldLabel className="shrink-0 whitespace-nowrap" htmlFor="announcement-title">{text("announcementTitle")}</FieldLabel>
            <Input id="announcement-title" value={selected.title} disabled={disabled} onChange={event => update(selected.id, { title: event.target.value })} /></Field>
          <Field>
            <div className="flex items-center justify-between gap-2"><FieldLabel>{text("announcementContent")}</FieldLabel>
              <Button variant="ghost" size="sm" disabled={disabled || !!selected.content} onClick={() => update(selected.id, { content: text("exampleContent") })}>
                <Code data-icon="inline-start" />{text("example")}</Button>
            </div>
            <AnnouncementMarkdownEditor key={selected.id} value={selected.content} disabled={loading || saving} text={text}
              onBusyChange={setUploading} onChange={content => update(selected.id, { content })}
              onUploadImage={async file => { const result = await api.pages.uploadImage(file); if (!result.url) throw new Error(text("announcementEditorUploadFailed")); return result.url; }} />
          </Field>
        </div> : <div className="flex min-h-40 flex-col items-center justify-center gap-2 text-center text-sm text-muted-foreground">
          {loading ? <Spinner /> : <><p>{text("announcementEmptyHint")}</p><Button variant="outline" size="sm" onClick={add}><Plus />{text("announcementAdd")}</Button></>}
        </div>}
      </section>
    </div>
  </main>;
}
