import { AdminPage } from '../components/admin-page'
import { errorMessage } from '../error-message'
import { useLatestRequest } from '@gooseforum/runtime'
import { useCallback, useEffect, useRef, useState } from "react";
import type { AdminFileResource, GooseAdminApi } from "@gooseforum/client";
import { Badge } from "@gooseforum/ui/components/badge";
import { Button } from "@gooseforum/ui/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@gooseforum/ui/components/dialog";
import {
  Empty,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@gooseforum/ui/components/select";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { ChevronLeft, ChevronRight, Copy, File, LayoutGrid, List, RefreshCw } from "lucide-react";
import { toast } from "@gooseforum/ui/components/sonner";
import type { AssetTextKey } from "../assets-i18n";
type Text = (key: AssetTextKey) => string;
export function FileResourcesManagementPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text: Text;
}) {
  const [view, setView] = useState<"grid" | "list">("grid");
  const sentinel = useRef<HTMLDivElement>(null);
  const [items, setItems] = useState<AdminFileResource[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [preview, setPreview] = useState<AdminFileResource | null>(null);
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const beginRequest = useLatestRequest();
  const load = useCallback(async () => {
    const isCurrent = beginRequest();
    setLoading(true);
    setError("");
    try {
      const r = await api.assets.files({ page, pageSize });
      if (!isCurrent()) return;
      setItems(previous => view === "grid" && page > 1
        ? [...previous, ...(r.list || []).filter(item => !previous.some(existing => existing.id === item.id))]
        : r.list || []);
      setTotal(r.total || 0);
      setPage(r.page || page);
      setPageSize(r.pageSize || r.size || pageSize);
    } catch (reason) {
      if (!isCurrent()) return;
      setError(errorMessage(reason, text("loadFailed")));
    } finally {
      if (isCurrent()) setLoading(false);
    }
  }, [beginRequest, api, page, pageSize, text, view]);
  useEffect(() => {
    void load();
  }, [load]);
  useEffect(() => {
    if (view !== "grid" || loading || error || page >= totalPages || !sentinel.current) return;
    const observer = new IntersectionObserver(([entry]) => {
      if (entry.isIntersecting) {
        observer.disconnect();
        setPage(current => current + 1);
      }
    }, { rootMargin: "400px" });
    observer.observe(sentinel.current);
    return () => observer.disconnect();
  }, [view, loading, error, page, totalPages]);
  function changeView(next: "grid" | "list") {
    if (next === view) return;
    beginRequest();
    setItems([]);
    setLoading(true);
    setPage(1);
    setView(next);
  }
  return (
    <AdminPage>
      <header className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">{text("files")}</h2>
          <p className="text-xs text-muted-foreground">{text("filesHint")}</p>
        </div>
        <div className="flex items-center gap-2">
        <div className="flex gap-1 rounded-lg border p-1" role="group" aria-label={text("viewMode")}>
          <Button variant={view === "grid" ? "secondary" : "ghost"} size="icon-sm" aria-label={text("gridView")} aria-pressed={view === "grid"} onClick={() => changeView("grid")}><LayoutGrid /></Button>
          <Button variant={view === "list" ? "secondary" : "ghost"} size="icon-sm" aria-label={text("listView")} aria-pressed={view === "list"} onClick={() => changeView("list")}><List /></Button>
        </div>
        <Button
          variant="outline"
          size="sm"
          disabled={loading}
          onClick={() => { if (view === "grid" && page !== 1) { setItems([]); setPage(1); } else void load(); }}
        >
          <RefreshCw
            data-icon="inline-start"
            className={loading ? "animate-spin" : undefined}
          />
          {text("refresh")}
        </Button>
        </div>
      </header>
      <section className="overflow-hidden rounded-lg border bg-background">
        {loading && !items.length ? (
          <AssetEmpty title={text("loading")} icon={<Spinner />} />
        ) : error && !items.length ? (
          <AssetEmpty title={error} icon={<File />} />
        ) : !items.length ? (
          <AssetEmpty title={text("empty")} icon={<File />} />
        ) : (
          view === "grid" ? (
            <div className="grid grid-flow-row-dense auto-rows-[1px] grid-cols-2 items-start gap-x-4 p-4 md:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
              {items.map(item => (
                <MasonryCard key={item.id}>
                  <button type="button" className="relative block w-full cursor-pointer text-left focus-visible:outline-2 focus-visible:outline-primary" onClick={() => setPreview(item)} aria-label={item.name}>
                    <ResourceThumbnail item={item} />
                  </button>
                </MasonryCard>
              ))}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="border-b bg-muted/30 text-muted-foreground"><tr>
                  {[text("name"), text("type"), text("size"), text("uploader"), text("created")].map(label => <th key={label} className="whitespace-nowrap p-3 font-medium">{label}</th>)}
                </tr></thead>
                <tbody>{items.map(item => <tr key={item.id} className="border-b last:border-0 hover:bg-muted/30">
                  <td className="p-3"><button type="button" className="flex max-w-96 items-center gap-3 text-left hover:text-primary" onClick={() => setPreview(item)}>
                    <span className="flex size-12 shrink-0 items-center justify-center overflow-hidden rounded-lg bg-muted">{isImage(item) ? <ResourceListImage item={item} /> : <File />}</span>
                    <span className="truncate">{item.name}</span>
                  </button></td>
                  <td className="p-3"><Badge variant="secondary">{item.type || "file"}</Badge></td>
                  <td className="whitespace-nowrap p-3">{formatBytes(item.size)}</td>
                  <td className="p-3"><a href={`/u/${item.userId}`} target="_blank" rel="noreferrer" className="hover:text-primary">{item.uploaderUsername || `#${item.userId}`}</a></td>
                  <td className="whitespace-nowrap p-3 text-muted-foreground">{formatTime(item.createdAt)}</td>
                </tr>)}</tbody>
              </table>
            </div>
          )
        )}
        {view === "grid" ? (
          <div ref={sentinel} className="flex min-h-16 items-center justify-center gap-2 p-4 text-sm text-muted-foreground" aria-live="polite">
            {loading ? <><Spinner />{text("loading")}</> : error ? <><span>{error}</span><Button variant="outline" onClick={() => void load()}>{text("retry")}</Button></> : page < totalPages ? <Button variant="ghost" onClick={() => setPage(current => current + 1)}>{text("loadMore")}</Button> : <span>{text("allLoaded")} · {total}</span>}
          </div>
        ) : <footer className="flex items-center justify-between gap-3 border-t bg-muted/10 px-3 py-2 text-sm text-muted-foreground">
          <span>{total}</span>
          <div className="flex items-center gap-1.5">
            <Select
              value={String(pageSize)}
              onValueChange={(v) => {
                setPageSize(Number(v));
                setPage(1);
              }}
            >
              <SelectTrigger className="w-20">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {[10, 20, 30, 50].map((n) => (
                    <SelectItem key={n} value={String(n)}>
                      {n}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              size="icon-sm"
              disabled={loading || page <= 1}
              onClick={() => setPage(page - 1)}
            >
              <ChevronLeft />
            </Button>
            <span>
              {text("page")} {page}/{totalPages}
            </span>
            <Button
              variant="outline"
              size="icon-sm"
              disabled={loading || page >= totalPages}
              onClick={() => setPage(page + 1)}
            >
              <ChevronRight />
            </Button>
          </div>
        </footer>}
      </section>
      <Preview item={preview} text={text} onClose={() => setPreview(null)} />
    </AdminPage>
  );
}
function MasonryCard({ children }: { children: React.ReactNode }) {
  const content = useRef<HTMLDivElement>(null);
  const [span, setSpan] = useState(12);
  useEffect(() => {
    const element = content.current;
    if (!element) return;
    const measure = () => setSpan(Math.ceil(element.getBoundingClientRect().height) + 16);
    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(element);
    return () => observer.disconnect();
  }, []);
  return <article style={{ gridRowEnd: `span ${span}` }} className="group min-w-0 pb-4">
    <div ref={content} className="overflow-hidden rounded-2xl border bg-muted/20">{children}</div>
  </article>;
}
function ResourceThumbnail({ item }: { item: AdminFileResource }) {
  const [failed, setFailed] = useState(false);
  if (isImage(item) && !failed) return <img src={item.url} alt="" loading="lazy" onError={() => setFailed(true)} className="block h-auto max-h-[32rem] min-h-24 w-full object-contain" />;
  return <div className="flex min-h-56 flex-col items-center justify-center gap-4 p-6 text-muted-foreground">
    <File className="size-12" />
    <span className="line-clamp-3 break-all text-center text-sm font-medium text-foreground">{item.name}</span>
    <Badge variant="secondary">{item.type || "file"}</Badge>
  </div>;
}
function ResourceListImage({ item }: { item: AdminFileResource }) {
  const [failed, setFailed] = useState(false);
  if (failed) return <File className="size-5 text-muted-foreground" />;
  return <img src={item.url} alt="" loading="lazy" onError={() => setFailed(true)} className="size-full object-cover" />;
}
function Preview({
  item,
  text,
  onClose,
}: {
  item: AdminFileResource | null;
  text: Text;
  onClose(): void;
}) {
  async function copy() {
    if (!item) return;
    try {
      await navigator.clipboard.writeText(
        new URL(item.url, window.location.origin).toString(),
      );
      toast.success(text("copied"));
    } catch {
      toast.error(text("loadFailed"));
    }
  }
  return (
    <Dialog open={Boolean(item)} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-h-[calc(100vh-2rem)] sm:max-w-5xl">
        <DialogHeader>
          <DialogTitle className="truncate">{item?.name}</DialogTitle>
          <DialogDescription className="truncate font-mono text-xs">
            {item?.url}
          </DialogDescription>
        </DialogHeader>
        {item ? (
          <div className="grid min-h-0 gap-3 lg:grid-cols-[minmax(0,1fr)_18rem]">
            <div className="flex max-h-[70vh] min-h-64 items-center justify-center overflow-hidden rounded-lg border bg-muted/50">
              {isImage(item) ? (
                <PreviewImage item={item} />
              ) : (
                <File className="size-12 text-muted-foreground" />
              )}
            </div>
            <aside className="flex flex-col gap-3 rounded-lg border bg-muted/20 p-3 text-sm">
              <Detail
                label={text("uploader")}
                value={item.uploaderUsername || `#${item.userId}`}
              />
              <Detail label={text("type")} value={item.type} />
              <Detail label={text("size")} value={formatBytes(item.size)} />
              <Detail
                label={text("created")}
                value={formatTime(item.createdAt)}
              />
              <p className="break-all font-mono text-xs text-muted-foreground">
                {new URL(item.url, window.location.origin).toString()}
              </p>
              <Button variant="outline" size="sm" onClick={() => void copy()}>
                <Copy data-icon="inline-start" />
                {text("copy")}
              </Button>
            </aside>
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}
function PreviewImage({ item }: { item: AdminFileResource }) {
  const [failed, setFailed] = useState(false);
  if (failed) {
    return (
      <div className="flex flex-col items-center gap-3 p-6 text-muted-foreground">
        <File className="size-12" />
        <span className="break-all text-center text-sm">{item.name}</span>
      </div>
    );
  }
  return (
    <img
      src={item.url}
      alt={item.name}
      onError={() => setFailed(true)}
      className="max-h-full max-w-full object-contain"
    />
  );
}
function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-xs text-muted-foreground">{label}</div>
      <div>{value}</div>
    </div>
  );
}
function AssetEmpty({ title, icon }: { title: string; icon: React.ReactNode }) {
  return (
    <Empty className="min-h-48 border-0">
      <EmptyHeader>
        <EmptyMedia variant="icon">{icon}</EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
      </EmptyHeader>
    </Empty>
  );
}
function isImage(x: AdminFileResource) {
  return x.type.startsWith("image/");
}
function formatBytes(v: number) {
  if (!Number.isFinite(v) || v <= 0) return "0 B";
  if (v < 1024) return `${v} B`;
  if (v < 1048576) return `${(v / 1024).toFixed(1)} KB`;
  return `${(v / 1048576).toFixed(2)} MB`;
}
function formatTime(v: string) {
  if (!v) return "—";
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? v : d.toLocaleString();
}
