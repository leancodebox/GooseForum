import { useLatestRequest } from '../use-latest-request'
import { useCallback, useEffect, useState } from "react";
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
import { ChevronLeft, ChevronRight, Copy, File, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import type { AssetTextKey } from "../assets-i18n";
type Text = (key: AssetTextKey) => string;
export function FileResourcesManagementPage({
  api,
  text,
}: {
  api: GooseAdminApi;
  text: Text;
}) {
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
      setItems(r.list || []);
      setTotal(r.total || 0);
      setPage(r.page || page);
      setPageSize(r.pageSize || r.size || pageSize);
    } catch (reason) {
      if (!isCurrent()) return;
      setError(message(reason, text("loadFailed")));
    } finally {
      if (isCurrent()) setLoading(false);
    }
  }, [beginRequest, api, page, pageSize, text]);
  useEffect(() => {
    void load();
  }, [load]);
  return (
    <main className="flex flex-1 flex-col gap-3 px-3 py-3 lg:px-4">
      <header className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">{text("files")}</h2>
          <p className="text-xs text-muted-foreground">{text("filesHint")}</p>
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
      </header>
      <section className="overflow-hidden rounded-lg border bg-background">
        {loading && !items.length ? (
          <AssetEmpty title={text("loading")} icon={<Spinner />} />
        ) : error ? (
          <AssetEmpty title={error} icon={<File />} />
        ) : !items.length ? (
          <AssetEmpty title={text("empty")} icon={<File />} />
        ) : (
          <div className="grid gap-3 p-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
            {items.map((item) => (
              <article
                key={item.id}
                className="group overflow-hidden rounded-lg border bg-background transition hover:border-primary/35 hover:shadow-sm"
              >
                <Button
                  variant="ghost"
                  className="aspect-[4/3] h-auto w-full overflow-hidden rounded-none bg-muted p-0 hover:bg-muted"
                  onClick={() => setPreview(item)}
                >
                  {isImage(item) ? (
                    <img
                      src={item.url}
                      alt={item.name}
                      className="size-full object-cover transition group-hover:scale-[1.02]"
                    />
                  ) : (
                    <span className="flex size-full flex-col items-center justify-center gap-2 text-muted-foreground">
                      <File className="size-8" />
                      <span className="max-w-full truncate px-4 text-xs">
                        {item.type || "file"}
                      </span>
                    </span>
                  )}
                </Button>
                <div className="flex flex-col gap-2 p-3">
                  <div>
                    <h3 className="truncate text-sm font-medium">
                      {item.name}
                    </h3>
                    <p className="mt-1 truncate font-mono text-xs text-muted-foreground">
                      {item.url}
                    </p>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <Badge variant="secondary" className="font-mono">
                      {item.type}
                    </Badge>
                    <span>{formatBytes(item.size)}</span>
                  </div>
                  <div className="flex justify-between gap-2 text-xs text-muted-foreground">
                    <a
                      href={`/u/${item.userId}`}
                      target="_blank"
                      rel="noreferrer"
                      className="truncate hover:text-primary"
                    >
                      {text("uploader")}{" "}
                      {item.uploaderUsername || `#${item.userId}`}
                    </a>
                    <span className="whitespace-nowrap">
                      {formatTime(item.createdAt)}
                    </span>
                  </div>
                </div>
              </article>
            ))}
          </div>
        )}
        <footer className="flex items-center justify-between gap-3 border-t bg-muted/10 px-3 py-2 text-sm text-muted-foreground">
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
              disabled={page <= 1}
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
              disabled={page >= totalPages}
              onClick={() => setPage(page + 1)}
            >
              <ChevronRight />
            </Button>
          </div>
        </footer>
      </section>
      <Preview item={preview} text={text} onClose={() => setPreview(null)} />
    </main>
  );
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
                <img
                  src={item.url}
                  alt={item.name}
                  className="max-h-full max-w-full object-contain"
                />
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
function message(r: unknown, f: string) {
  return r instanceof Error && r.message ? r.message : f;
}
