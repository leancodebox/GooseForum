import { useEffect, useId, useRef, useState } from "react";
import type { ReplyTargetPayload } from "@gooseforum/client";
import { ChevronDown, ChevronUp } from "lucide-react";
import { Button } from "@gooseforum/ui/components/button";
import { cn } from "@gooseforum/ui/lib/utils";
import { GooseLink } from "@gooseforum/runtime";
import { RenderedContent } from "../content/rendered-content";
import { ProfileAvatar } from "../users/profile-avatar";

export function ReplyReference({
  topicId,
  target,
  t,
}: {
  topicId: number;
  target?: ReplyTargetPayload;
  t: (key: string) => string;
}) {
  const contentId = useId();
  const content = useRef<HTMLDivElement>(null);
  const [expanded, setExpanded] = useState(false);
  const [overflowing, setOverflowing] = useState(false);
  const available = target && !target.unavailable;
  const sourceURL =
    available && target.postNo
      ? `/p/post/${topicId}${target.postNo > 1 ? `/${target.postNo}` : ""}#post-${target.id}`
      : undefined;

  useEffect(() => {
    setExpanded(false);
    const root = content.current;
    const body = root?.firstElementChild as HTMLElement | null;
    if (!root || !body) {
      setOverflowing(false);
      return;
    }
    const measure = () => {
      const lineHeight =
        Number.parseFloat(getComputedStyle(body).lineHeight) || 24;
      setOverflowing(body.scrollHeight > lineHeight * 4 + 1);
    };
    measure();
    const observer =
      typeof ResizeObserver !== "undefined"
        ? new ResizeObserver(measure)
        : undefined;
    observer?.observe(body);
    root.addEventListener("load", measure, true);
    return () => {
      observer?.disconnect();
      root.removeEventListener("load", measure, true);
    };
  }, [target?.id, target?.renderedContent, target?.unavailable]);

  return (
    <aside
      data-slot="reply-reference"
      className="mb-2 min-w-0 border-l-2 border-primary/45 bg-muted/40 py-2"
    >
      <div className="flex min-h-7 items-center gap-2 px-3 text-sm text-muted-foreground">
        {available ? (
          <ProfileAvatar
            src={target.author.avatarUrl}
            name={target.author.username}
            className="size-6"
            framed={false}
          />
        ) : null}
        {target?.author.username ? (
          <span className="min-w-0 truncate font-medium">
            @{target.author.username}
          </span>
        ) : null}
        {sourceURL ? (
          <GooseLink
            href={sourceURL}
            className="shrink-0 text-xs hover:text-primary"
          >
            #{target?.postNo}
          </GooseLink>
        ) : null}
      </div>
      {available ? (
        <>
          <blockquote cite={sourceURL} className="m-0 px-3 pt-2">
            <div id={contentId} ref={content}>
              <RenderedContent
                html={target.renderedContent || ""}
                className={cn(
                  !expanded && "max-h-[4lh] overflow-hidden",
                  !expanded &&
                    overflowing &&
                    "[mask-image:linear-gradient(to_bottom,black_calc(100%_-_1.5lh),transparent)]",
                )}
              />
            </div>
          </blockquote>
          {overflowing ? (
            <Button
              variant="ghost"
              size="xs"
              className="mx-2.5 mt-1"
              aria-expanded={expanded}
              aria-controls={contentId}
              onClick={() => setExpanded((value) => !value)}
            >
              {expanded ? (
                <ChevronUp data-icon="inline-start" />
              ) : (
                <ChevronDown data-icon="inline-start" />
              )}
              {t(expanded ? "collapseReply" : "expandReply")}
            </Button>
          ) : null}
        </>
      ) : (
        <div className="px-3 pt-2 text-sm text-muted-foreground">
          {t("replyTargetUnavailable")}
        </div>
      )}
    </aside>
  );
}
