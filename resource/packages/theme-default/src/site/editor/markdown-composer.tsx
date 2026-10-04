import {
  useRef,
  useState,
  useEffect,
  type ChangeEvent,
  type ClipboardEvent as ReactClipboardEvent,
  type DragEvent as ReactDragEvent,
  type ReactNode,
} from "react";
import {
  Bold,
  Code,
  Code2,
  Eye,
  Heading,
  Image,
  Italic,
  Link,
  List,
  ListOrdered,
  MessageSquareQuote,
  Minus,
  Strikethrough,
  Table2,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { Badge } from "@gooseforum/ui/components/badge";
import { Button } from "@gooseforum/ui/components/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@gooseforum/ui/components/popover";
import { Textarea } from "@gooseforum/ui/components/textarea";
import { ToggleGroup, ToggleGroupItem } from "@gooseforum/ui/components/toggle-group";
import { cn } from "@gooseforum/ui/lib/utils";
import { useGooseRuntime } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";
import {
  hasUnsupportedVisualMarkdown,
  insertBlock,
  insertInline,
  markdownFromClipboard,
  renderMarkdown,
  mentionMarkdown,
  mentionQuery,
} from "@gooseforum/markdown";
import { optimizeImage, validateImage } from "./image";
import {
  VisualMarkdownEditor,
  type EditorAction,
  type VisualMarkdownEditorHandle,
} from "./visual-markdown-editor";

type EditorMode = "visual" | "markdown";
type Selection = { start: number; end: number };
const toolbar: Array<{ action: EditorAction; icon: typeof Bold }> = [
  { action: "bold", icon: Bold },
  { action: "italic", icon: Italic },
  { action: "strike", icon: Strikethrough },
  { action: "inlineCode", icon: Code },
  { action: "quote", icon: MessageSquareQuote },
  { action: "code", icon: Code2 },
  { action: "bulletList", icon: List },
  { action: "orderedList", icon: ListOrdered },
  { action: "horizontalRule", icon: Minus },
];

export function MarkdownComposer({
  value,
  onChange,
  minHeight = "min-h-80",
  toolbarPlacement = "top",
  actions,
  status,
  sourceVersion = 1,
}: {
  value: string;
  onChange(value: string): void;
  minHeight?: string;
  toolbarPlacement?: "top" | "bottom";
  actions?: ReactNode;
  status?: ReactNode;
  sourceVersion?: 0 | 1;
}) {
  const { t } = useTranslation("publish");
  const runtime = useGooseRuntime();
  const serverError = useServerErrorMessage();
  const [selectedMode, setMode] = useState<EditorMode>(() =>
    hasUnsupportedVisualMarkdown(value,sourceVersion) ? "markdown" : "visual",
  );
  const mode = hasUnsupportedVisualMarkdown(value, sourceVersion) ? "markdown" : selectedMode;
  const [preview, setPreview] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState<[number, number]>([
    0, 0,
  ]);
  const [dragging, setDragging] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const textarea = useRef<HTMLTextAreaElement>(null);
  const visual = useRef<VisualMarkdownEditorHandle>(null);
  const pendingSelection = useRef<Selection | undefined>(undefined);
  const [query, setQuery] = useState<string | null>(null);
  const [candidateResult, setCandidateResult] = useState<{
    query: string | null;
    items: Array<{ id: string; username: string }>;
  }>({ query: null, items: [] });
  const candidates = candidateResult.query === query ? candidateResult.items : [];
  const candidateLoading = query !== null && candidateResult.query !== query;
  const [candidateIndex, setCandidateIndex] = useState(0);
  const mentionSelection = useRef<Selection | null>(null);
  useEffect(() => {
    if (query === null || sourceVersion === 0) return;
    let active = true;
    const timer = setTimeout(() => {
      void runtime.api.users.mentions(query)
        .then((items) => { if (active) setCandidateResult({ query, items }); })
        .catch(() => { if (active) setCandidateResult({ query, items: [] }); });
    }, 200);
    return () => { active = false; clearTimeout(timer); };
  }, [query, sourceVersion, runtime]);
  function updateQuery(next: string | null) {
    setQuery(next);
    if (next !== query) setCandidateIndex(0);
  }
  function updateMention(target: HTMLTextAreaElement) {
    if (sourceVersion === 0 || target.selectionStart !== target.selectionEnd) { updateQuery(null); return; }
    const end = target.selectionStart;
    const match = mentionQuery(target.value.slice(0, end));
    mentionSelection.current = match ? { start: end - match.length, end } : null;
    updateQuery(match?.query ?? null);
  }
  function selectMention(user: { id: string; username: string }) {
    if (mode === "visual") visual.current?.insertMention(user.id, user.username);
    else {
      const selection = mentionSelection.current;
      if (!selection) return;
      const text = mentionMarkdown(user.id, user.username);
      replace({ value: value.slice(0, selection.start) + text + value.slice(selection.end), start: selection.start + text.length, end: selection.start + text.length });
    }
    updateQuery(null);
  }
  function mentionKey(event: {
    key: string;
    isComposing?: boolean;
    keyCode?: number;
    nativeEvent?: { isComposing?: boolean; keyCode?: number };
    preventDefault(): void;
  }) {
    if (sourceVersion === 0 || query === null || event.isComposing || event.nativeEvent?.isComposing || event.keyCode === 229 || event.nativeEvent?.keyCode === 229) return false;
    if (event.key === "Escape") { event.preventDefault(); updateQuery(null); return true; }
    if (!candidates.length) return false;
    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault();
      setCandidateIndex((index) => (index + (event.key === "ArrowDown" ? 1 : -1) + candidates.length) % candidates.length);
      return true;
    }
    if (event.key === "Enter") {
      event.preventDefault();
      selectMention(candidates[candidateIndex] ?? candidates[0]);
      return true;
    }
    return false;
  }

  function updateMode(next: string) {
    if (!next) return;
    if (next === "visual" && hasUnsupportedVisualMarkdown(value,sourceVersion)) {
      setError(t("visualUnsupported"));
      return;
    }
    setError("");
    setPreview(false);
    setMode(next as EditorMode);
    updateQuery(null);
    requestAnimationFrame(() =>
      next === "visual" ? visual.current?.focus() : textarea.current?.focus(),
    );
  }
  function replace(next: { value: string; start: number; end: number }) {
    pendingSelection.current = next;
    onChange(next.value);
    requestAnimationFrame(() => {
      const target = textarea.current;
      const selection = pendingSelection.current;
      if (!target || !selection) return;
      target.focus();
      target.setSelectionRange(selection.start, selection.end);
      pendingSelection.current = undefined;
    });
  }
  function selected() {
    const target = textarea.current;
    return {
      start: target?.selectionStart || 0,
      end: target?.selectionEnd || 0,
    };
  }
  function apply(action: EditorAction) {
    if (mode === "visual") {
      visual.current?.applyAction(action);
      return;
    }
    const { start, end } = selected();
    const placeholder = t(
      `placeholder.${action === "bulletList" || action === "orderedList" ? "listItem" : action === "quote" ? "quote" : action === "bold" || action === "italic" || action === "strike" ? action : "link"}`,
    );
    if (action === "bold")
      replace(insertInline(value, start, end, "**", "**", placeholder));
    else if (action === "italic")
      replace(insertInline(value, start, end, "*", "*", placeholder));
    else if (action === "strike")
      replace(insertInline(value, start, end, "~~", "~~", placeholder));
    else if (action === "inlineCode")
      replace(insertInline(value, start, end, "`", "`", "code"));
    else if (action === "quote")
      replace(
        insertBlock(
          value,
          start,
          end,
          `> ${value.slice(start, end) || placeholder}`,
        ),
      );
    else if (action === "code")
      replace(
        insertBlock(
          value,
          start,
          end,
          `\`\`\`\n${value.slice(start, end) || "code"}\n\`\`\``,
        ),
      );
    else if (action === "bulletList" || action === "orderedList")
      replace(
        insertBlock(
          value,
          start,
          end,
          `${action === "bulletList" ? "-" : "1."} ${value.slice(start, end) || placeholder}`,
        ),
      );
    else if (action === "horizontalRule")
      replace(insertBlock(value, start, end, "---"));
    else replace(insertInline(value, start, end, "  \n", "", ""));
  }
  function insertMarkdown(markdown: string) {
    if (mode === "visual") visual.current?.insertMarkdown(markdown);
    else {
      const { start, end } = selected();
      replace(insertBlock(value, start, end, markdown));
    }
  }
  function insertLink() {
    if (mode === "visual")
      visual.current?.insertMarkdown(`[${t("placeholder.link")}](https://)`);
    else {
      const { start, end } = selected();
      replace(
        insertInline(
          value,
          start,
          end,
          "[",
          "](https://)",
          t("placeholder.link"),
        ),
      );
    }
  }
  function setHeading(level: number) {
    if (mode === "visual")
      visual.current?.setBlock(
        level === 0
          ? "paragraph"
          : (`heading_${level}` as `heading_${1 | 2 | 3 | 4 | 5 | 6}`),
      );
    else {
      const { start, end } = selected();
      const lineStart = value.lastIndexOf("\n", start - 1) + 1;
      const nextBreak = value.indexOf("\n", end);
      const lineEnd = nextBreak < 0 ? value.length : nextBreak;
      const prefix = level ? `${"#".repeat(level)} ` : "";
      const replacement = value
        .slice(lineStart, lineEnd)
        .split("\n")
        .map((line) =>
          line.trim() ? `${prefix}${line.replace(/^#{1,6}\s+/, "")}` : line,
        )
        .join("\n");
      replace({
        value: `${value.slice(0, lineStart)}${replacement}${value.slice(lineEnd)}`,
        start: lineStart,
        end: lineStart + replacement.length,
      });
    }
  }
  function insertTable() {
    const markdown = "|   |   |\n| --- | --- |\n|   |   |";
    if (mode === "visual") visual.current?.insertTable(2, 2);
    else insertMarkdown(markdown);
  }
  async function upload(files: File[]) {
    const images = files.filter((file) => file.type.startsWith("image/"));
    if (!images.length || uploading) return;
    setUploading(true);
    setUploadProgress([0, images.length]);
    setMessage("");
    setError("");
    const inserted: string[] = [];
    const failures: string[] = [];
    for (const [index, file] of images.entries()) {
      if (!validateImage(file))
        failures.push(`${file.name}: ${t("imageTooLarge")}`);
      else {
        try {
          const optimized = await optimizeImage(file);
          const url = await runtime.api.uploads.image(optimized);
          const alt =
            file.name
              .replace(/\.[^.]+$/, "")
              .replace(/[[\]\n\r]/g, " ")
              .trim() || "image";
          inserted.push(`![${alt}](${url})`);
        } catch (reason) {
          failures.push(
            `${file.name}: ${serverError(reason, t("noUploadableImages"))}`,
          );
        }
      }
      setUploadProgress([index + 1, images.length]);
    }
    if (inserted.length) {
      insertMarkdown(inserted.join("\n"));
      setMessage(
        t(inserted.length > 1 ? "imagesInserted" : "imageInserted", {
          count: inserted.length,
        }),
      );
    }
    if (failures.length) setError(failures.slice(0, 3).join("；"));
    setUploading(false);
  }
  async function handleInput(event: ChangeEvent<HTMLInputElement>) {
    const files = Array.from(event.target.files || []);
    event.target.value = "";
    await upload(files);
  }
  function handlePasteData(data: DataTransfer, preventDefault: () => void) {
    const files = Array.from(data.items)
      .filter((item) => item.kind === "file" && item.type.startsWith("image/"))
      .map((item) => item.getAsFile())
      .filter((file): file is File => Boolean(file));
    if (files.length) {
      preventDefault();
      void upload(files);
      return;
    }
    const markdown = markdownFromClipboard(data);
    if (markdown) {
      preventDefault();
      insertMarkdown(markdown);
    }
  }
  function handlePaste(event: ReactClipboardEvent) {
    handlePasteData(event.clipboardData, () => event.preventDefault());
  }
  function handleDrop(event: ReactDragEvent) {
    setDragging(false);
    const files = Array.from(event.dataTransfer.files).filter((file) =>
      file.type.startsWith("image/"),
    );
    if (!files.length) return;
    event.preventDefault();
    void upload(files);
  }

  const toolbarRow = (
    <div
      data-slot="markdown-composer-toolbar"
      className={cn(
        "flex flex-wrap items-center gap-1.5",
        toolbarPlacement === "top" ? "mb-2" : "shrink-0 border-t bg-muted/30 px-3 py-2 sm:px-4",
      )}
    >
        <div className="flex flex-wrap items-center gap-1">
          <Popover>
            <PopoverTrigger asChild>
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                aria-label={t("toolbar.heading")}
              >
                <Heading />
              </Button>
            </PopoverTrigger>
            <PopoverContent align="start" className="w-44">
              <Button
                variant="ghost"
                size="sm"
                className="w-full justify-start"
                onClick={() => setHeading(0)}
              >
                {t("toolbar.paragraph")}
              </Button>
              {[1, 2, 3, 4, 5, 6].map((level) => (
                <Button
                  key={level}
                  variant="ghost"
                  size="sm"
                  className="w-full justify-start"
                  onClick={() => setHeading(level)}
                >
                  H{level}
                </Button>
              ))}
            </PopoverContent>
          </Popover>
          {toolbar.map(({ action, icon: Icon }) => (
            <Button
              key={action}
              type="button"
              variant="ghost"
              size="icon-sm"
              title={t(`toolbar.${action}`)}
              aria-label={t(`toolbar.${action}`)}
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => apply(action)}
            >
              <Icon />
            </Button>
          ))}
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            title={t("toolbar.link")}
            aria-label={t("toolbar.link")}
            onClick={insertLink}
          >
            <Link />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            title={t("toolbar.table")}
            aria-label={t("toolbar.table")}
            onClick={insertTable}
          >
            <Table2 />
          </Button>
          <label
            className="inline-flex size-7 cursor-pointer items-center justify-center rounded-lg hover:bg-muted"
            title={t("uploadImageTitle")}
          >
            <Image className="size-4" />
            <span className="sr-only">{t("uploadImageTitle")}</span>
            <input
              type="file"
              accept="image/*"
              multiple
              className="sr-only"
              disabled={uploading}
              onChange={(event) => void handleInput(event)}
            />
          </label>
          {uploading ? (
            <Badge variant="secondary">
              {t(
                uploadProgress[1] > 1 ? "processingImages" : "processingImage",
                { done: uploadProgress[0], total: uploadProgress[1] },
              )}
            </Badge>
          ) : null}
        </div>
        <div className="ml-auto flex items-center gap-1">
          <ToggleGroup
            type="single"
            value={mode}
            variant="outline"
            size="sm"
            spacing={0}
            onValueChange={updateMode}
          >
            <ToggleGroupItem value="visual">{t("visualMode")}</ToggleGroupItem>
            <ToggleGroupItem value="markdown">
              {t("markdownMode")}
            </ToggleGroupItem>
          </ToggleGroup>
          <Button
            type="button"
            size="sm"
            variant={preview ? "secondary" : "outline"}
            onClick={() => setPreview((current) => !current)}
          >
            <Eye data-icon="inline-start" />
            {t("preview")}
          </Button>
          {actions}
        </div>
    </div>
  );

  return (
    <div
      className={cn(
        toolbarPlacement === "bottom" && "flex h-full min-h-0 flex-col",
      )}
    >
      {toolbarPlacement === "top" ? toolbarRow : null}
      <div
        className={cn(
          toolbarPlacement === "bottom" && "min-h-0 flex-1 overflow-auto p-3 sm:p-4",
        )}
      >
        <div
          className={cn(
            "relative rounded-lg border bg-background px-3",
            dragging && "ring-2 ring-primary",
          )}
          onDragOver={(event) => {
            if (
              Array.from(event.dataTransfer.items).some((item) =>
                item.type.startsWith("image/"),
              )
            ) {
              event.preventDefault();
              setDragging(true);
            }
          }}
          onDragLeave={() => setDragging(false)}
          onDrop={handleDrop}
        >
          {dragging ? (
            <div className="pointer-events-none absolute inset-0 grid place-items-center bg-primary/10 font-semibold text-primary">
              {t("dropToUpload")}
            </div>
          ) : null}
          {preview ? (
            value.trim() ? (
              <div
                data-slot="markdown-composer-preview"
                className={cn(
                  "markdown-composer-content markdown-composer-preview typeset typeset-forum gf-prose gf-prose-post px-1 py-4",
                  minHeight,
                )}
                dangerouslySetInnerHTML={{ __html: renderMarkdown(value,sourceVersion) }}
              />
            ) : (
              <p
                data-slot="markdown-composer-preview"
                className={cn(
                  "markdown-composer-content typeset typeset-forum px-1 py-4 text-muted-foreground",
                  minHeight,
                )}
              >
                {t("emptyPreview")}
              </p>
            )
          ) : mode === "visual" ? (
            <VisualMarkdownEditor
              ref={visual}
              value={value}
              onChange={onChange}
              sourceVersion={sourceVersion}
              onMentionQuery={updateQuery}
              onMentionKeyDown={mentionKey}
              placeholder={t("visualPlaceholder")}
              onPaste={(event) => {
                if (event.clipboardData)
                  handlePasteData(event.clipboardData, () =>
                    event.preventDefault(),
                  );
              }}
              onDrop={(event) => {
                setDragging(false);
                const files = Array.from(event.dataTransfer?.files || []).filter(
                  (file) => file.type.startsWith("image/"),
                );
                if (!files.length) return;
                event.preventDefault();
                void upload(files);
              }}
            />
          ) : (
            <Textarea
              ref={textarea}
              value={value}
              onChange={(event) => {onChange(event.target.value);updateMention(event.target);}}
              onSelect={(event)=>updateMention(event.currentTarget)}
              onPaste={handlePaste}
              onKeyDown={(event) => {
                if(mentionKey(event))return;
                if (!(event.ctrlKey || event.metaKey)) return;
                const key = event.key.toLowerCase();
                if (key === "b" || key === "i" || key === "k") {
                  event.preventDefault();
                  if (key === "k") insertLink();
                  else apply(key === "b" ? "bold" : "italic");
                }
              }}
              className={cn(
                "markdown-composer-content typeset typeset-forum resize-none rounded-none border-0 px-1 py-4 shadow-none focus-visible:ring-0",
                minHeight,
              )}
              placeholder={t("bodyPlaceholder")}
            />
          )}
          {query!==null && !preview && sourceVersion===1 ? (
            <div className="border-t py-1" role="listbox" aria-label={t("mentionUsers")}>
              {candidateLoading ? <p className="px-2 py-1 text-sm text-muted-foreground">{t("mentionLoading")}</p> : candidates.length ? candidates.map((user,index)=>(
                <button key={user.id} type="button" role="option" aria-selected={index===candidateIndex} className={cn("block w-full px-2 py-1.5 text-left text-sm hover:bg-muted focus-visible:bg-muted",index===candidateIndex&&"bg-muted")} onMouseDown={(event)=>event.preventDefault()} onClick={()=>selectMention(user)}>@{user.username}</button>
              )) : <p className="px-2 py-1 text-sm text-muted-foreground">{t("mentionEmpty")}</p>}
            </div>
          ) : null}
        </div>
        {error ? <p className="mt-2 text-sm text-destructive">{error}</p> : null}
        {message ? <p className="mt-2 text-sm text-success">{message}</p> : null}
        {status}
      </div>
      {toolbarPlacement === "bottom" ? toolbarRow : null}
    </div>
  );
}
