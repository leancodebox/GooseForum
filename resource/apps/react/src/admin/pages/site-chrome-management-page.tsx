import { useCallback, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import {
  closestCenter,
  DndContext,
  DragOverlay,
  PointerSensor,
  useDroppable,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import {
  arrayMove,
  horizontalListSortingStrategy,
  SortableContext,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import type {
  GooseAdminApi,
  LayoutPayload,
  SiteChromeConfig,
  SiteChromeGroup,
  SiteChromeItem,
} from "@gooseforum/client";
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
  FieldGroup,
  FieldLabel,
} from "@gooseforum/react/components/ui/field";
import { Input } from "@gooseforum/react/components/ui/input";
import { Spinner } from "@gooseforum/react/components/ui/spinner";
import { Switch } from "@gooseforum/react/components/ui/switch";
import {
  ToggleGroup,
  ToggleGroupItem,
} from "@gooseforum/react/components/ui/toggle-group";
import {
  Bell,
  FileText,
  Flame,
  Eye,
  EyeOff,
  Heart,
  Inbox,
  Link,
  MessageCircle,
  Pencil,
  Plus,
  Save,
  Scale,
  Trash2,
  TrendingUp,
  Upload,
} from "lucide-react";
import { toast } from "sonner";
import type { SettingsTextKey } from "../settings-i18n";
type Text = (k: SettingsTextKey) => string;
const fixedMain = [
  [MessageCircle, "topics"],
  [Flame, "hot"],
  [TrendingUp, "popular"],
  [Inbox, "messages"],
  [Bell, "notifications"],
  [FileText, "drafts"],
  [Scale, "moderation"],
] as const;
const fixedResources = [
  [Link, "links"],
  [Heart, "sponsors"],
] as const;
type Key = "header" | "mainMenu" | "resources";
type ChromeScope =
  | { key: "mainMenu" | "resources" }
  | { key: "sidebar"; groupIndex: number };
type ChromeDragData =
  | { type: "chrome-item"; scope: ChromeScope; index: number }
  | { type: "chrome-container"; scope: ChromeScope };
type ItemDialog =
  | { key: Key; index: number | null }
  | { key: "sidebar"; groupIndex: number; index: number | null };
const emptyItem = (): SiteChromeItem => ({
  id: crypto.randomUUID(),
  enabled: true,
  type: "link",
  label: "",
  i18nLabel: "",
  url: "",
});
const emptyGroup = (): SiteChromeGroup => ({
  id: crypto.randomUUID(),
  title: "Group",
  i18nLabel: "",
  items: [],
});
export function SiteChromeManagementPage({
  api,
  text,
  layout,
}: {
  api: GooseAdminApi;
  text: Text;
  layout: LayoutPayload;
}) {
  const [config, setConfig] = useState<SiteChromeConfig>({
    header: [],
    mainMenu: [],
    resources: [],
    sidebarGroups: [],
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [dialog, setDialog] = useState<ItemDialog | null>(null);
  const [groupDialog, setGroupDialog] = useState<number | null | undefined>();
  const [brandOpen, setBrandOpen] = useState(false);
  const [footerOpen, setFooterOpen] = useState(false);
  const [activeSidebarItem, setActiveSidebarItem] =
    useState<SiteChromeItem | null>(null);
  const textRef = useRef(text);
  useEffect(() => {
    textRef.current = text;
  }, [text]);
  const sidebarSensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 6 } }),
  );
  const load = useCallback(async () => {
    setLoading(true);
    try {
      setConfig(normalize(await api.settings.chrome()));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : textRef.current("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [api]);
  useEffect(() => {
    void load();
  }, [load]);
  function update(fn: (next: SiteChromeConfig) => void) {
    setConfig((current) => {
      const next = structuredClone(current);
      fn(next);
      return next;
    });
  }
  async function save() {
    setSaving(true);
    try {
      await api.settings.saveChrome(normalize(config));
      toast.success(text("saved"));
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    } finally {
      setSaving(false);
    }
  }
  function list(d: ItemDialog) {
    return d.key === "sidebar"
      ? config.sidebarGroups[d.groupIndex].items
      : config[d.key];
  }
  function sidebarList(target: SiteChromeConfig, scope: ChromeScope) {
    return scope.key === "sidebar"
      ? target.sidebarGroups[scope.groupIndex].items
      : target[scope.key];
  }
  function moveSidebarItem(event: DragEndEvent) {
    setActiveSidebarItem(null);
    const active = event.active.data.current as ChromeDragData | undefined;
    const over = event.over?.data.current as ChromeDragData | undefined;
    if (!active || !over || active.type !== "chrome-item") return;
    const targetScope = over.scope;
    update((next) => {
      const source = sidebarList(next, active.scope);
      const target = sidebarList(next, targetScope);
      if (source === target && over.type === "chrome-item") {
        const reordered = arrayMove(source, active.index, over.index);
        source.splice(0, source.length, ...reordered);
        return;
      }
      const [item] = source.splice(active.index, 1);
      if (!item) return;
      target.splice(
        over.type === "chrome-item" ? over.index : target.length,
        0,
        item,
      );
    });
  }
  function startSidebarDrag(event: DragStartEvent) {
    const data = event.active.data.current as ChromeDragData | undefined;
    if (!data || data.type !== "chrome-item") return;
    setActiveSidebarItem(sidebarList(config, data.scope)[data.index] || null);
  }
  return (
    <main className="flex flex-1 flex-col gap-3 px-3 py-3 lg:px-4">
      <header className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">{text("chrome")}</h2>
          <p className="text-xs text-muted-foreground">{text("chromeHint")}</p>
        </div>
        <Button
          size="sm"
          disabled={saving || loading}
          onClick={() => void save()}
        >
          {saving ? (
            <Spinner data-icon="inline-start" />
          ) : (
            <Save data-icon="inline-start" />
          )}
          {text("save")}
        </Button>
      </header>
      {loading ? (
        <Empty className="min-h-64 border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Spinner />
            </EmptyMedia>
            <EmptyTitle>{text("loading")}</EmptyTitle>
          </EmptyHeader>
        </Empty>
      ) : (
        <section className="min-h-[760px] overflow-hidden rounded-xl border bg-muted text-foreground shadow-sm">
          <header className="border-b border-transparent bg-background/0">
            <div className="mx-auto grid h-16 w-full max-w-[1600px] grid-cols-[auto_minmax(0,1fr)] items-center gap-8 px-5">
              <BrandPreview
                config={config}
                layout={layout}
                onEdit={() => setBrandOpen(true)}
              />
              <nav className="flex min-w-0 items-center gap-1">
                <ChromeList
                  items={config.header}
                  horizontal
                  systemIds={["links", "sponsors"]}
                  onChange={(items) =>
                    update((n) => {
                      n.header = items;
                    })
                  }
                  onEdit={(index) =>
                    ["links", "sponsors"].includes(config.header[index].id)
                      ? update((n) => {
                          n.header[index].enabled = !n.header[index].enabled;
                        })
                      : setDialog({ key: "header", index })
                  }
                  onDelete={(index) =>
                    update((n) => {
                      ["links", "sponsors"].includes(n.header[index].id)
                        ? (n.header[index].enabled = false)
                        : n.header.splice(index, 1);
                    })
                  }
                />
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 rounded-md border border-dashed border-border/70 bg-background/40 px-2 text-sm font-medium text-muted-foreground hover:border-primary/40 hover:bg-primary/10 hover:text-primary"
                  onClick={() => setDialog({ key: "header", index: null })}
                >
                  <Plus data-icon="inline-start" />
                  {text("add")}
                </Button>
              </nav>
            </div>
          </header>
          <div className="mx-auto grid w-full max-w-[1600px] grid-cols-[224px_minmax(0,1fr)] gap-3 px-5 py-3">
            <aside className="-my-3 flex min-w-0 flex-col self-start py-3 pr-3">
              <DndContext
                sensors={sidebarSensors}
                collisionDetection={closestCenter}
                onDragStart={startSidebarDrag}
                onDragCancel={() => setActiveSidebarItem(null)}
                onDragEnd={moveSidebarItem}
              >
                <div className="flex flex-col">
                  <PreviewSection
                    title={text("mainMenu")}
                    addLabel={text("add")}
                    hideTitle
                    scope={{ key: "mainMenu" }}
                    fixed={fixedMain.map(([icon, key]) => ({
                      icon,
                      label: text(key),
                    }))}
                    items={config.mainMenu}
                    onChange={(items) =>
                      update((n) => {
                        n.mainMenu = items;
                      })
                    }
                    onEdit={(index) => setDialog({ key: "mainMenu", index })}
                    onDelete={(index) =>
                      update((n) => {
                        n.mainMenu.splice(index, 1);
                      })
                    }
                    onAdd={() => setDialog({ key: "mainMenu", index: null })}
                  />
                  <PreviewSection
                    title={text("resources")}
                    addLabel={text("add")}
                    fixed={fixedResources.map(([icon, key]) => ({
                      icon,
                      label: text(key),
                    }))}
                    scope={{ key: "resources" }}
                    items={config.resources}
                    onChange={(items) =>
                      update((n) => {
                        n.resources = items;
                      })
                    }
                    onEdit={(index) => setDialog({ key: "resources", index })}
                    onDelete={(index) =>
                      update((n) => {
                        n.resources.splice(index, 1);
                      })
                    }
                    onAdd={() => setDialog({ key: "resources", index: null })}
                  />
                  {config.sidebarGroups.map((group, groupIndex) => (
                    <PreviewSection
                      key={group.id}
                      title={group.title}
                      addLabel={text("add")}
                      items={group.items}
                      scope={{ key: "sidebar", groupIndex }}
                      onChange={(items) =>
                        update((n) => {
                          n.sidebarGroups[groupIndex].items = items;
                        })
                      }
                      onEdit={(index) =>
                        setDialog({ key: "sidebar", groupIndex, index })
                      }
                      onDelete={(index) =>
                        update((n) => {
                          n.sidebarGroups[groupIndex].items.splice(index, 1);
                        })
                      }
                      onAdd={() =>
                        setDialog({ key: "sidebar", groupIndex, index: null })
                      }
                      onEditGroup={() => setGroupDialog(groupIndex)}
                      onDeleteGroup={() =>
                        update((n) => {
                          n.sidebarGroups.splice(groupIndex, 1);
                        })
                      }
                    />
                  ))}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setGroupDialog(null)}
                  >
                    <Plus data-icon="inline-start" />
                    {text("group")}
                  </Button>
                  <div className="px-2 text-[10px] font-bold uppercase text-muted-foreground">
                    {text("categories")}
                  </div>
                  {layout.sidebar.categories.map((c) => (
                    <div
                      key={c.id}
                      className="flex h-7 items-center gap-2 px-2 text-[13px]"
                    >
                      <span
                        className="size-2 rounded"
                        style={{ backgroundColor: c.color }}
                      />
                      {c.label}
                    </div>
                  ))}
                  <button
                    type="button"
                    className="mt-1 rounded-md border border-dashed border-transparent px-2 py-1 text-left text-xs leading-5 text-muted-foreground hover:border-primary/25 hover:bg-primary/5"
                    onClick={() => setFooterOpen(true)}
                  >
                    <div className="flex flex-wrap gap-x-3">
                      {config.footerInfo?.list?.map((item, index) => (
                        <span key={`${index}-${item.name}`}>
                          {item.name || text("link")}
                        </span>
                      ))}
                    </div>
                    <div className="flex flex-wrap gap-x-3">
                      {config.footerInfo?.primary?.map((item, index) => (
                        <span key={`${index}-${item.content}`}>
                          {item.content || text("primary")}
                        </span>
                      ))}
                    </div>
                    {!config.footerInfo?.list?.length &&
                    !config.footerInfo?.primary?.length ? (
                      <span>{text("footer")}</span>
                    ) : null}
                  </button>
                </div>
                {typeof document !== "undefined"
                  ? createPortal(
                      <DragOverlay dropAnimation={null}>
                        {activeSidebarItem ? (
                          <ChromeDragPreview item={activeSidebarItem} />
                        ) : null}
                      </DragOverlay>,
                      document.body,
                    )
                  : null}
              </DndContext>
            </aside>
            <div className="min-w-0 overflow-hidden rounded-lg border bg-background">
              <div className="flex h-14 items-center gap-2 border-b px-4">
                <span className="rounded-lg bg-foreground px-4 py-2 text-sm font-semibold text-background">
                  {text("topics")}
                </span>
                <span className="px-3 py-2 text-sm font-medium text-muted-foreground">
                  {text("hot")}
                </span>
                <span className="px-3 py-2 text-sm font-medium text-muted-foreground">
                  {text("popular")}
                </span>
              </div>
              {[1, 2, 3, 4, 5, 6, 7].map((i) => (
                <div key={i} className="border-b px-4 py-4">
                  <div className="h-4 w-2/3 rounded bg-foreground/90" />
                  <div className="mt-3 h-3.5 w-4/5 rounded bg-foreground/20" />
                </div>
              ))}
            </div>
          </div>
        </section>
      )}
      <ItemEditor
        state={dialog}
        value={dialog ? list(dialog)[dialog.index ?? -1] : undefined}
        text={text}
        onClose={() => setDialog(null)}
        onSave={(item) => {
          if (!dialog) return;
          update((n) => {
            const l =
              dialog.key === "sidebar"
                ? n.sidebarGroups[dialog.groupIndex].items
                : n[dialog.key];
            if (dialog.index == null) l.push(item);
            else l[dialog.index] = item;
          });
          setDialog(null);
        }}
      />
      <GroupEditor
        index={groupDialog}
        value={
          groupDialog !== undefined && groupDialog !== null
            ? config.sidebarGroups[groupDialog]
            : undefined
        }
        text={text}
        onClose={() => setGroupDialog(undefined)}
        onSave={(group) => {
          update((n) => {
            if (groupDialog == null) n.sidebarGroups.push(group);
            else n.sidebarGroups[groupDialog] = group;
          });
          setGroupDialog(undefined);
        }}
      />
      <BrandEditor
        open={brandOpen}
        config={config}
        api={api}
        text={text}
        onClose={() => setBrandOpen(false)}
        onSave={(value) => {
          update((n) => Object.assign(n, value));
          setBrandOpen(false);
        }}
      />
      <FooterEditor
        open={footerOpen}
        value={config.footerInfo}
        text={text}
        onClose={() => setFooterOpen(false)}
        onSave={(footerInfo) => {
          update((n) => {
            n.footerInfo = footerInfo;
          });
          setFooterOpen(false);
        }}
      />
    </main>
  );
}
function PreviewSection({
  title,
  addLabel,
  items,
  onChange,
  onEdit,
  onDelete,
  onAdd,
  fixed = [],
  hideTitle = false,
  scope,
  onEditGroup,
  onDeleteGroup,
}: {
  title: string;
  addLabel: string;
  items: SiteChromeItem[];
  onChange(v: SiteChromeItem[]): void;
  onEdit(i: number): void;
  onDelete(i: number): void;
  onAdd(): void;
  fixed?: Array<{ icon: typeof Link; label: string }>;
  hideTitle?: boolean;
  scope: ChromeScope;
  onEditGroup?: () => void;
  onDeleteGroup?: () => void;
}) {
  const drop = useDroppable({
    id: `chrome-container-${scope.key}-${scope.key === "sidebar" ? scope.groupIndex : 0}`,
    data: { type: "chrome-container", scope } satisfies ChromeDragData,
  });
  return (
    <section ref={drop.setNodeRef} className={hideTitle ? "pb-2" : "mt-2"}>
      <div
        className={
          hideTitle
            ? "hidden"
            : "mb-1 flex h-5 items-center justify-between px-2 text-[10px] font-bold uppercase tracking-wide text-muted-foreground"
        }
      >
        <span>{title}</span>
        <span>
          {onEditGroup ? (
            <>
              <Button variant="ghost" size="icon-xs" onClick={onEditGroup}>
                <Pencil />
              </Button>
              <Button variant="ghost" size="icon-xs" onClick={onDeleteGroup}>
                <Trash2 />
              </Button>
            </>
          ) : null}
        </span>
      </div>
      {fixed.map(({ icon: Icon, label }, index) => (
        <div
          key={label}
          className={`flex h-8 items-center gap-2 rounded-md px-2 text-[13px] font-medium ${hideTitle && index === 0 ? "bg-primary/10 text-primary" : "text-foreground/75"}`}
        >
          <Icon className="size-4" />
          <span className="truncate">{label}</span>
        </div>
      ))}
      <ChromeList
        items={items}
        standalone={false}
        scope={scope}
        onChange={onChange}
        onEdit={onEdit}
        onDelete={onDelete}
      />
      <Button
        variant="ghost"
        size="sm"
        className="mt-1 h-7 w-full justify-start rounded-md border border-dashed border-border/70 bg-background/40 px-2 text-[13px] font-medium text-muted-foreground hover:border-primary/40 hover:bg-primary/10 hover:text-primary"
        onClick={onAdd}
      >
        <Plus data-icon="inline-start" />
        {addLabel}
      </Button>
    </section>
  );
}
function ChromeList({
  items,
  onChange,
  onEdit,
  onDelete,
  horizontal = false,
  systemIds = [],
  standalone = true,
  scope,
}: {
  items: SiteChromeItem[];
  onChange(v: SiteChromeItem[]): void;
  onEdit(i: number): void;
  onDelete(i: number): void;
  horizontal?: boolean;
  systemIds?: string[];
  standalone?: boolean;
  scope?: ChromeScope;
}) {
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 6 } }),
  );
  function end(e: DragEndEvent) {
    if (!e.over || e.active.id === e.over.id) return;
    const a = items.findIndex((x) => x.id === e.active.id),
      b = items.findIndex((x) => x.id === e.over?.id);
    if (a >= 0 && b >= 0) onChange(arrayMove(items, a, b));
  }
  const content = (
    <SortableContext
      items={items.map((x) => x.id)}
      strategy={
        horizontal ? horizontalListSortingStrategy : verticalListSortingStrategy
      }
    >
      <div
        className={horizontal ? "flex min-w-0 gap-1" : "flex flex-col gap-0.5"}
      >
        {items.map((item, index) => (
          <ChromeRow
            key={item.id}
            item={item}
            onEdit={() => onEdit(index)}
            onDelete={() => onDelete(index)}
            horizontal={horizontal}
            system={systemIds.includes(item.id)}
            dragData={scope ? { type: "chrome-item", scope, index } : undefined}
          />
        ))}
      </div>
    </SortableContext>
  );
  return standalone ? (
    <DndContext sensors={sensors} onDragEnd={end}>
      {content}
    </DndContext>
  ) : (
    content
  );
}
function ChromeRow({
  item,
  onEdit,
  onDelete,
  horizontal,
  system = false,
  dragData,
}: {
  item: SiteChromeItem;
  onEdit(): void;
  onDelete(): void;
  horizontal: boolean;
  system?: boolean;
  dragData?: ChromeDragData;
}) {
  const s = useSortable({ id: item.id, data: dragData });
  return (
    <div
      ref={s.setNodeRef}
      {...s.attributes}
      {...s.listeners}
      style={{
        transform: CSS.Transform.toString(s.transform),
        transition: s.transition,
        opacity: s.isDragging ? 0.35 : undefined,
      }}
      className={`group relative flex cursor-grab ${horizontal ? "h-7 max-w-56" : "h-8"} items-center gap-2 rounded-md px-2 text-[13px] font-medium text-foreground/75 transition-colors hover:bg-accent hover:text-foreground active:cursor-grabbing ${item.enabled ? "" : "opacity-40"}`}
    >
      {horizontal ? null : <Link className="size-4" />}
      <span className="min-w-0 flex-1 truncate transition-[padding] group-hover:pr-12 group-focus-within:pr-12">
        {item.label || item.i18nLabel || "Item"}
      </span>
      <span className="absolute right-1 top-1/2 z-10 flex -translate-y-1/2 items-center gap-0.5 rounded-sm bg-background p-0.5 opacity-0 shadow-sm ring-1 ring-border group-hover:opacity-100 group-focus-within:opacity-100">
        {system ? (
          <Button
            variant="ghost"
            size="icon-xs"
            className="size-5 rounded-sm p-0"
            onPointerDown={(e) => e.stopPropagation()}
            onClick={onEdit}
          >
            {item.enabled ? <EyeOff /> : <Eye />}
          </Button>
        ) : (
          <>
            <Button
              variant="ghost"
              size="icon-xs"
              className="size-5 rounded-sm p-0"
              onClick={onEdit}
              onPointerDown={(e) => e.stopPropagation()}
            >
              <Pencil />
            </Button>
            <Button
              variant="ghost"
              size="icon-xs"
              className="size-5 rounded-sm p-0"
              onPointerDown={(e) => e.stopPropagation()}
              onClick={onDelete}
            >
              <Trash2 />
            </Button>
          </>
        )}
      </span>
    </div>
  );
}
function ChromeDragPreview({ item }: { item: SiteChromeItem }) {
  return (
    <div className="flex h-8 w-52 items-center gap-2 rounded-md border bg-background px-2 text-[13px] font-medium text-foreground shadow-lg">
      <Link className="size-4 shrink-0" />
      <span className="truncate">{item.label || item.i18nLabel || "Item"}</span>
    </div>
  );
}
function ItemEditor({
  state,
  value,
  text,
  onClose,
  onSave,
}: {
  state: ItemDialog | null;
  value?: SiteChromeItem;
  text: Text;
  onClose(): void;
  onSave(v: SiteChromeItem): void;
}) {
  const [form, setForm] = useState(emptyItem());
  useEffect(() => {
    if (state) setForm(value ? { ...value } : emptyItem());
  }, [state, value]);
  return (
    <Dialog open={Boolean(state)} onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{value ? text("edit") : text("add")}</DialogTitle>
          <DialogDescription>{text("chromeHint")}</DialogDescription>
        </DialogHeader>
        <FieldGroup>
          <Field>
            <FieldLabel>{text("type")}</FieldLabel>
            <ToggleGroup
              type="single"
              value={form.type}
              onValueChange={(v) => v && setForm({ ...form, type: v })}
            >
              <ToggleGroupItem value="link">{text("link")}</ToggleGroupItem>
              <ToggleGroupItem value="text">{text("text")}</ToggleGroupItem>
            </ToggleGroup>
          </Field>
          <Field>
            <FieldLabel>{text("label")}</FieldLabel>
            <Input
              value={form.label}
              onChange={(e) => setForm({ ...form, label: e.target.value })}
            />
          </Field>
          <Field>
            <FieldLabel>{text("i18nLabel")}</FieldLabel>
            <Input
              value={form.i18nLabel}
              onChange={(e) => setForm({ ...form, i18nLabel: e.target.value })}
            />
          </Field>
          <Field>
            <FieldLabel>URL</FieldLabel>
            <Input
              value={form.url}
              onChange={(e) => setForm({ ...form, url: e.target.value })}
            />
          </Field>
          <Field orientation="horizontal">
            <FieldLabel>{text("enabled")}</FieldLabel>
            <Switch
              checked={form.enabled}
              onCheckedChange={(enabled) => setForm({ ...form, enabled })}
            />
          </Field>
        </FieldGroup>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {text("cancel")}
          </Button>
          <Button onClick={() => onSave(form)}>{text("save")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
function GroupEditor({
  index,
  value,
  text,
  onClose,
  onSave,
}: {
  index: number | null | undefined;
  value?: SiteChromeGroup;
  text: Text;
  onClose(): void;
  onSave(v: SiteChromeGroup): void;
}) {
  const [form, setForm] = useState(emptyGroup());
  useEffect(() => {
    if (index !== undefined)
      setForm(
        value
          ? { ...value, items: value.items.map((x) => ({ ...x })) }
          : emptyGroup(),
      );
  }, [index, value]);
  return (
    <Dialog open={index !== undefined} onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {value ? text("edit") : text("add")} {text("group")}
          </DialogTitle>
          <DialogDescription>{text("chromeHint")}</DialogDescription>
        </DialogHeader>
        <FieldGroup>
          <Field>
            <FieldLabel>{text("label")}</FieldLabel>
            <Input
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
            />
          </Field>
          <Field>
            <FieldLabel>{text("i18nLabel")}</FieldLabel>
            <Input
              value={form.i18nLabel}
              onChange={(e) => setForm({ ...form, i18nLabel: e.target.value })}
            />
          </Field>
        </FieldGroup>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {text("cancel")}
          </Button>
          <Button onClick={() => onSave(form)}>{text("save")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
function BrandEditor({
  open,
  config,
  api,
  text,
  onClose,
  onSave,
}: {
  open: boolean;
  config: SiteChromeConfig;
  api: GooseAdminApi;
  text: Text;
  onClose(): void;
  onSave(v: Partial<SiteChromeConfig>): void;
}) {
  const [type, setType] = useState(config.brandType || "default");
  const [value, setValue] = useState("");
  useEffect(() => {
    if (open) {
      setType(config.brandType || "default");
      setValue(
        config.brandType === "image"
          ? config.brandImage || ""
          : config.brandText || "",
      );
    }
  }, [config, open]);
  async function upload(file?: File) {
    if (!file) return;
    try {
      const r = await api.pages.uploadImage(file);
      setValue(r.url || "");
    } catch (r) {
      toast.error(r instanceof Error ? r.message : text("saveFailed"));
    }
  }
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{text("brand")}</DialogTitle>
          <DialogDescription>{text("chromeHint")}</DialogDescription>
        </DialogHeader>
        <ToggleGroup
          type="single"
          value={type}
          onValueChange={(v) => v && setType(v)}
        >
          <ToggleGroupItem value="default">
            {text("brandDefault")}
          </ToggleGroupItem>
          <ToggleGroupItem value="text">{text("brandText")}</ToggleGroupItem>
          <ToggleGroupItem value="image">{text("brandImage")}</ToggleGroupItem>
        </ToggleGroup>
        {type !== "default" ? (
          <Field>
            <FieldLabel>
              {type === "image" ? text("brandImage") : text("brandText")}
            </FieldLabel>
            <div className="flex gap-2">
              <Input value={value} onChange={(e) => setValue(e.target.value)} />
              {type === "image" ? (
                <Button asChild variant="outline">
                  <label>
                    <input
                      className="sr-only"
                      type="file"
                      accept="image/*"
                      onChange={(e) => void upload(e.target.files?.[0])}
                    />
                    <Upload />
                  </label>
                </Button>
              ) : null}
            </div>
          </Field>
        ) : null}
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {text("cancel")}
          </Button>
          <Button
            onClick={() =>
              onSave({
                brandType: type,
                brandText: type === "text" ? value : "",
                brandImage: type === "image" ? value : "",
              })
            }
          >
            {text("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
function FooterEditor({
  open,
  value,
  text,
  onClose,
  onSave,
}: {
  open: boolean;
  value: SiteChromeConfig["footerInfo"];
  text: Text;
  onClose(): void;
  onSave(v: NonNullable<SiteChromeConfig["footerInfo"]>): void;
}) {
  const [links, setLinks] = useState<{ name: string; url: string }[]>([]);
  const [primary, setPrimary] = useState<{ content: string }[]>([]);
  useEffect(() => {
    if (open) {
      setLinks(value?.list?.map((x) => ({ ...x })) || []);
      setPrimary(value?.primary?.map((x) => ({ ...x })) || []);
    }
  }, [open, value]);
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{text("footer")}</DialogTitle>
          <DialogDescription>{text("chromeHint")}</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 sm:grid-cols-2">
          <FieldGroup>
            <FieldLabel>{text("link")}</FieldLabel>
            {links.map((x, i) => (
              <div key={i} className="flex gap-1">
                <Input
                  value={x.name}
                  onChange={(e) =>
                    setLinks(
                      links.map((v, n) =>
                        n === i ? { ...v, name: e.target.value } : v,
                      ),
                    )
                  }
                />
                <Input
                  value={x.url}
                  onChange={(e) =>
                    setLinks(
                      links.map((v, n) =>
                        n === i ? { ...v, url: e.target.value } : v,
                      ),
                    )
                  }
                />
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => setLinks(links.filter((_, n) => n !== i))}
                >
                  <Trash2 />
                </Button>
              </div>
            ))}
            <Button
              variant="outline"
              size="sm"
              onClick={() => setLinks([...links, { name: "", url: "" }])}
            >
              <Plus /> {text("add")}
            </Button>
          </FieldGroup>
          <FieldGroup>
            <FieldLabel>{text("primary")}</FieldLabel>
            {primary.map((x, i) => (
              <div key={i} className="flex gap-1">
                <Input
                  value={x.content}
                  onChange={(e) =>
                    setPrimary(
                      primary.map((v, n) =>
                        n === i ? { content: e.target.value } : v,
                      ),
                    )
                  }
                />
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => setPrimary(primary.filter((_, n) => n !== i))}
                >
                  <Trash2 />
                </Button>
              </div>
            ))}
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPrimary([...primary, { content: "" }])}
            >
              <Plus /> {text("add")}
            </Button>
          </FieldGroup>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {text("cancel")}
          </Button>
          <Button onClick={() => onSave({ list: links, primary })}>
            {text("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
function BrandPreview({
  config,
  layout,
  onEdit,
}: {
  config: SiteChromeConfig;
  layout: LayoutPayload;
  onEdit(): void;
}) {
  const type = config.brandType || layout.site.brandType;
  const image = config.brandImage || layout.site.brandImage || layout.site.logo;
  const label =
    config.brandText ||
    layout.site.brandText ||
    layout.site.name ||
    "GooseForum";
  return (
    <Button
      type="button"
      variant="ghost"
      className="group relative -ml-1 h-auto min-w-0 shrink-0 justify-start gap-2 rounded-md border border-dashed border-transparent px-2 py-1 text-left font-normal shadow-none hover:border-primary/25 hover:bg-primary/5"
      onClick={onEdit}
    >
      {type === "image" && image ? (
        <img
          src={image}
          alt={label}
          className="h-8 w-auto max-w-40 shrink-0 object-contain sm:h-9"
        />
      ) : type === "text" ? (
        <span className="max-w-44 truncate text-xl font-semibold tracking-tighter text-primary sm:text-2xl md:max-w-none">
          {label}
        </span>
      ) : (
        <span className="max-w-44 truncate text-xl font-semibold tracking-tighter sm:text-2xl md:max-w-none">
          <span className="text-primary">Goose</span>
          <span className="text-foreground">Forum</span>
        </span>
      )}
      <span className="ml-1 inline-flex size-5 items-center justify-center rounded-sm text-muted-foreground opacity-0 transition group-hover:opacity-100">
        <Pencil className="size-3.5" />
      </span>
    </Button>
  );
}
function normalize(v: Partial<SiteChromeConfig>): SiteChromeConfig {
  const item = (x: Partial<SiteChromeItem>): SiteChromeItem => ({
    id: x.id || crypto.randomUUID(),
    enabled: x.enabled !== false,
    type: x.type === "text" ? "text" : "link",
    label: x.label || "",
    i18nLabel: x.i18nLabel || "",
    url: x.url || "",
  });
  const headers = [
    {
      id: "sponsors",
      enabled: true,
      type: "link",
      label: "Sponsors",
      i18nLabel: "shell.nav.sponsors",
      url: "/sponsors",
    },
    {
      id: "links",
      enabled: true,
      type: "link",
      label: "Links",
      i18nLabel: "shell.nav.links",
      url: "/links",
    },
  ] as SiteChromeItem[];
  const current = (v.header || []).map(item);
  const ids = new Set(current.map((x) => x.id));
  return {
    header: [...headers.filter((x) => !ids.has(x.id)), ...current],
    mainMenu: (v.mainMenu || []).map(item),
    resources: (v.resources || []).map(item),
    sidebarGroups: (v.sidebarGroups || []).map((g) => ({
      id: g.id || crypto.randomUUID(),
      title: g.title || "Group",
      i18nLabel: g.i18nLabel || "",
      items: (g.items || []).map(item),
    })),
    footerInfo: {
      primary:
        v.footerInfo?.primary?.map((x) => ({ content: x.content || "" })) || [],
      list:
        v.footerInfo?.list?.map((x) => ({
          name: x.name || "",
          url: x.url || "",
        })) || [],
    },
    brandType: ["default", "text", "image"].includes(v.brandType || "")
      ? v.brandType
      : "default",
    brandText: v.brandText || "",
    brandImage: v.brandImage || "",
  };
}
