import { Maximize2, Minimize2, Minus, Send, X } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { PostPayload, ViewerPayload } from "@gooseforum/client";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@gooseforum/ui/components/avatar";
import { Button } from "@gooseforum/ui/components/button";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { cn } from "@gooseforum/ui/lib/utils";
import { MarkdownComposer } from "../editor/markdown-composer";

export function TopicComposer({
  open,
  minimized,
  expanded,
  content,
  editing,
  target,
  viewer,
  topicTitle,
  busy,
  error,
  onContent,
  onSubmit,
  onClose,
  onMinimize,
  onExpand,
  onClearTarget,
}: {
  open: boolean;
  minimized: boolean;
  expanded: boolean;
  content: string;
  editing: boolean;
  target?: PostPayload;
  viewer: ViewerPayload;
  topicTitle: string;
  busy: boolean;
  error: string;
  onContent(value: string): void;
  onSubmit(): void;
  onClose(): void;
  onMinimize(): void;
  onExpand(): void;
  onClearTarget(): void;
}) {
  const { t } = useTranslation("topic");
  if (!open) return null;
  if (minimized)
    return (
      <div className="pointer-events-none fixed inset-x-0 bottom-3 z-40 flex justify-center px-3">
        <div className="pointer-events-auto flex w-full max-w-lg items-center gap-3 rounded-xl border bg-background p-2 shadow-xl">
          <Avatar className="size-8">
            <AvatarImage src={viewer.avatarUrl} />
            <AvatarFallback>
              {viewer.username.slice(0, 1).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <button
            type="button"
            className="min-w-0 flex-1 text-left"
            onClick={onMinimize}
          >
            <span className="block truncate text-sm font-semibold">
              {editing ? t("editOwnReply") : t("joinDiscussion")}
            </span>
            <span className="block truncate text-xs text-muted-foreground">
              {topicTitle}
            </span>
          </button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("joinDiscussion")}
            onClick={onMinimize}
          >
            <Maximize2 />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("close")}
            onClick={onClose}
          >
            <X />
          </Button>
        </div>
      </div>
    );
  return (
    <div className="pointer-events-none fixed inset-x-0 bottom-0 z-40 flex justify-center px-0 sm:px-8">
      <section
        className={cn(
          "pointer-events-auto flex w-full max-w-4xl flex-col overflow-hidden rounded-t-xl border bg-background shadow-2xl",
          expanded ? "h-[calc(100dvh-1rem)]" : "h-[min(31rem,70dvh)]",
        )}
        aria-label={editing ? t("editOwnReply") : t("joinDiscussion")}
      >
        <header className="flex items-center gap-3 border-b px-3 py-2 sm:px-4">
          <Avatar className="size-9">
            <AvatarImage src={viewer.avatarUrl} />
            <AvatarFallback>
              {viewer.username.slice(0, 1).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1">
            <h2 className="truncate text-sm font-semibold">
              {editing ? t("editOwnReply") : t("joinDiscussion")}
            </h2>
            <p className="truncate text-xs text-muted-foreground">
              {topicTitle}
            </p>
          </div>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("joinDiscussion")}
            disabled={busy}
            onClick={onMinimize}
          >
            <Minus />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("more")}
            onClick={onExpand}
          >
            {expanded ? <Minimize2 /> : <Maximize2 />}
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("close")}
            disabled={busy}
            onClick={onClose}
          >
            <X />
          </Button>
        </header>
        {target ? (
          <div className="flex items-center gap-2 border-b bg-muted/50 px-4 py-2 text-xs">
            <span className="min-w-0 flex-1 truncate">
              {t("replyToPost", {
                user: target.author.username,
                no: target.postNo,
              })}
            </span>
            <Button
              variant="ghost"
              size="icon-xs"
              aria-label={t("close")}
              onClick={onClearTarget}
            >
              <X />
            </Button>
          </div>
        ) : null}
        <div className="min-h-0 flex-1">
          <MarkdownComposer
            value={content}
            onChange={onContent}
            minHeight={expanded ? "min-h-[45vh]" : "min-h-32"}
            toolbarPlacement="bottom"
            status={
              error ? (
                <p className="mt-2 text-sm text-destructive">{error}</p>
              ) : null
            }
            actions={
              <Button disabled={busy} onClick={onSubmit}>
                {busy ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <Send data-icon="inline-start" />
                )}
                {busy
                  ? t("publishing")
                  : editing
                    ? t("save")
                    : t("publishReply")}
              </Button>
            }
          />
        </div>
      </section>
    </div>
  );
}
