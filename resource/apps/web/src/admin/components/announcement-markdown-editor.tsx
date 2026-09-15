import {
  useRef,
  useState,
  type ChangeEvent,
  type ClipboardEvent,
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
import {
  insertBlock,
  insertInline,
  markdownFromClipboard,
  renderMarkdown,
} from "@gooseforum/markdown";
import { Button } from "@gooseforum/ui/components/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@gooseforum/ui/components/popover";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Textarea } from "@gooseforum/ui/components/textarea";
import {
  ToggleGroup,
  ToggleGroupItem,
} from "@gooseforum/ui/components/toggle-group";
import type { ContentSettingsTextKey } from "../content-settings-i18n";

type Text = (key: ContentSettingsTextKey) => string;
type Selection = { start: number; end: number };
type InlineAction = "bold" | "italic" | "strike" | "inlineCode";

const inlineActions: Array<{
  action: InlineAction;
  icon: typeof Bold;
  label: ContentSettingsTextKey;
}> = [
  { action: "bold", icon: Bold, label: "announcementEditorBold" },
  { action: "italic", icon: Italic, label: "announcementEditorItalic" },
  { action: "strike", icon: Strikethrough, label: "announcementEditorStrike" },
  { action: "inlineCode", icon: Code, label: "announcementEditorInlineCode" },
];

export function AnnouncementMarkdownEditor({
  value,
  disabled,
  text,
  onChange,
  onUploadImage,
}: {
  value: string;
  disabled: boolean;
  text: Text;
  onChange(value: string): void;
  onUploadImage(file: File): Promise<string>;
}) {
  const [mode, setMode] = useState<"markdown" | "preview">("markdown");
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState("");
  const textarea = useRef<HTMLTextAreaElement>(null);
  const fileInput = useRef<HTMLInputElement>(null);
  const pendingSelection = useRef<Selection | undefined>(undefined);

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
    return {
      start: textarea.current?.selectionStart || 0,
      end: textarea.current?.selectionEnd || 0,
    };
  }

  function applyInline(action: InlineAction) {
    const { start, end } = selected();
    const syntax = {
      bold: ["**", "**"],
      italic: ["*", "*"],
      strike: ["~~", "~~"],
      inlineCode: ["`", "`"],
    }[action];
    replace(
      insertInline(
        value,
        start,
        end,
        syntax[0],
        syntax[1],
        text("announcementEditorPlaceholder"),
      ),
    );
  }

  function insertMarkdown(markdown: string) {
    const { start, end } = selected();
    replace(insertBlock(value, start, end, markdown));
  }

  function wrapBlock(prefix: string, fallback: string) {
    const { start, end } = selected();
    const selectedValue = value.slice(start, end) || fallback;
    const content = selectedValue
      .split("\n")
      .map((line) => `${prefix}${line}`)
      .join("\n");
    replace(insertBlock(value, start, end, content));
  }

  function insertLink() {
    const { start, end } = selected();
    replace(
      insertInline(
        value,
        start,
        end,
        "[",
        "](https://)",
        text("announcementEditorLinkText"),
      ),
    );
  }

  function setHeading(level: number) {
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

  async function uploadImage(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file || uploading) return;
    setUploading(true);
    setUploadError("");
    try {
      const url = await onUploadImage(file);
      const alt = file.name.replace(/\.[^.]+$/, "").replace(/[[\]\n\r]/g, " ");
      insertMarkdown(`![${alt || "image"}](${url})`);
    } catch (reason) {
      setUploadError(
        reason instanceof Error
          ? reason.message
          : text("announcementEditorUploadFailed"),
      );
    } finally {
      setUploading(false);
    }
  }

  function paste(event: ClipboardEvent<HTMLTextAreaElement>) {
    const markdown = markdownFromClipboard(event.clipboardData);
    if (!markdown) return;
    event.preventDefault();
    insertMarkdown(markdown);
  }

  const toolbarDisabled = disabled || mode === "preview";
  return (
    <div className="overflow-hidden rounded-lg border bg-background">
      <div
        data-slot="announcement-markdown-toolbar"
        className="flex flex-wrap items-center gap-1 border-b bg-muted/30 p-2"
      >
        <Popover>
          <PopoverTrigger asChild>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              disabled={toolbarDisabled}
              aria-label={text("announcementEditorHeading")}
            >
              <Heading />
            </Button>
          </PopoverTrigger>
          <PopoverContent align="start" className="w-44">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="w-full justify-start"
              onClick={() => setHeading(0)}
            >
              {text("announcementEditorParagraph")}
            </Button>
            {[1, 2, 3, 4].map((level) => (
              <Button
                key={level}
                type="button"
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
        {inlineActions.map(({ action, icon: Icon, label }) => (
          <Button
            key={action}
            type="button"
            variant="ghost"
            size="icon-sm"
            disabled={toolbarDisabled}
            aria-label={text(label)}
            onMouseDown={(event) => event.preventDefault()}
            onClick={() => applyInline(action)}
          >
            <Icon />
          </Button>
        ))}
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          disabled={toolbarDisabled}
          aria-label={text("announcementEditorQuote")}
          onClick={() =>
            wrapBlock("> ", text("announcementEditorPlaceholder"))
          }
        >
          <MessageSquareQuote />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          disabled={toolbarDisabled}
          aria-label={text("announcementEditorCode")}
          onClick={() =>
            insertMarkdown(
              `\`\`\`\n${text("announcementEditorPlaceholder")}\n\`\`\``,
            )
          }
        >
          <Code2 />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          disabled={toolbarDisabled}
          aria-label={text("announcementEditorBulletList")}
          onClick={() =>
            wrapBlock("- ", text("announcementEditorListItem"))
          }
        >
          <List />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          disabled={toolbarDisabled}
          aria-label={text("announcementEditorOrderedList")}
          onClick={() =>
            wrapBlock("1. ", text("announcementEditorListItem"))
          }
        >
          <ListOrdered />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          disabled={toolbarDisabled}
          aria-label={text("announcementEditorRule")}
          onClick={() => insertMarkdown("---")}
        >
          <Minus />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          disabled={toolbarDisabled}
          aria-label={text("announcementEditorLink")}
          onClick={insertLink}
        >
          <Link />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          disabled={toolbarDisabled}
          aria-label={text("announcementEditorTable")}
          onClick={() =>
            insertMarkdown("|   |   |\n| --- | --- |\n|   |   |")
          }
        >
          <Table2 />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          disabled={disabled || uploading || mode === "preview"}
          aria-label={text("announcementEditorUploadImage")}
          onClick={() => fileInput.current?.click()}
        >
          {uploading ? <Spinner /> : <Image />}
        </Button>
        <input
          ref={fileInput}
          type="file"
          accept="image/*"
          className="sr-only"
          disabled={disabled || uploading || mode === "preview"}
          onChange={(event) => void uploadImage(event)}
        />
        <ToggleGroup
          type="single"
          value={mode}
          variant="outline"
          size="sm"
          spacing={0}
          className="ml-auto"
          onValueChange={(next) => {
            if (next) setMode(next as "markdown" | "preview");
          }}
        >
          <ToggleGroupItem value="markdown" disabled={disabled}>
            {text("announcementEditorMarkdown")}
          </ToggleGroupItem>
          <ToggleGroupItem value="preview" disabled={disabled}>
            <Eye data-icon="inline-start" />
            {text("announcementEditorPreview")}
          </ToggleGroupItem>
        </ToggleGroup>
      </div>
      {mode === "preview" ? (
        value.trim() ? (
          <div
            data-slot="announcement-markdown-preview"
            className="typeset typeset-announcement gf-prose min-h-80 px-4 py-3"
            dangerouslySetInnerHTML={{ __html: renderMarkdown(value) }}
          />
        ) : (
          <p
            data-slot="announcement-markdown-preview"
            className="min-h-80 px-4 py-3 text-sm text-muted-foreground"
          >
            {text("announcementEditorEmptyPreview")}
          </p>
        )
      ) : (
        <Textarea
          ref={textarea}
          value={value}
          disabled={disabled}
          aria-label={text("announcementContent")}
          className="min-h-80 resize-y rounded-none border-0 px-4 py-3 font-mono text-sm shadow-none focus-visible:ring-0"
          placeholder={text("announcementEditorPlaceholder")}
          onChange={(event) => onChange(event.target.value)}
          onPaste={paste}
        />
      )}
      {uploadError ? (
        <p role="alert" className="border-t px-4 py-2 text-sm text-destructive">
          {uploadError}
        </p>
      ) : null}
    </div>
  );
}
