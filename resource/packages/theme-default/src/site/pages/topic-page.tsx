import {
  lazy,
  Suspense,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type {
  LayoutPayload,
  PostPayload,
  PostWindowPayload,
  ReplyTargetPayload,
  TopicDetailProps,
} from "@gooseforum/client";
import {
  Ban,
  Bell,
  Bookmark,
  ChevronLeft,
  ChevronRight,
  Clock,
  Eye,
  Flag,
  Heart,
  Maximize2,
  Minimize2,
  MessageSquare,
  PencilLine,
  Reply,
  RotateCcw,
  Trash2,
  X,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { Alert, AlertDescription } from "@gooseforum/ui/components/alert";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@gooseforum/ui/components/avatar";
import { Badge } from "@gooseforum/ui/components/badge";
import { Button } from "@gooseforum/ui/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@gooseforum/ui/components/dialog";
import { Slider } from "@gooseforum/ui/components/slider";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Textarea } from "@gooseforum/ui/components/textarea";
import { ToggleGroup, ToggleGroupItem } from "@gooseforum/ui/components/toggle-group";
import { cn } from "@gooseforum/ui/lib/utils";
import { GooseLink, useGooseRuntime } from "@gooseforum/runtime";
import { RenderedContent } from "../content/rendered-content";
import { emptyShellHeader, useShellHeader } from "../layout/shell-header";
import { TopicTable } from "../topics/topic-list";
import { ReplyReference } from "../topics/reply-reference";
import { UserCardPopover } from "../users/user-card-popover";
import { ProfileAvatar } from "../users/profile-avatar";

type PendingReport = {
  targetType: "topic" | "post";
  targetId: number;
  title: string;
  excerpt: string;
};
type PendingModeration = {
  target: "topic" | "post";
  id: number;
  action: "ban" | "unban";
};

const loadTopicComposer = () => import("../topics/topic-composer");
const TopicComposer = lazy(() =>
  loadTopicComposer().then((module) => ({ default: module.TopicComposer })),
);

export function TopicPageView({
  layout,
  page,
}: {
  layout: LayoutPayload;
  page: TopicDetailProps;
}) {
  const { t } = useTranslation("topic");
  const { t: homeT } = useTranslation("home");
  const runtime = useGooseRuntime();
  const setShellHeader = useShellHeader();
  const titleRef = useRef<HTMLHeadingElement>(null);
  const [posts, setPosts] = useState(page.postStream.posts);
  const [replyTargets, setReplyTargets] = useState(
    page.postStream.replyTargets || [],
  );
  const [hasBefore, setHasBefore] = useState(page.postStream.hasBefore);
  const [hasAfter, setHasAfter] = useState(page.postStream.hasAfter);
  const [beforeNo, setBeforeNo] = useState(
    page.postStream.beforePostNo || firstNo(page.postStream.posts),
  );
  const [afterNo, setAfterNo] = useState(
    page.postStream.afterPostNo || lastNo(page.postStream.posts),
  );
  const [maxNo, setMaxNo] = useState(
    page.postStream.maxPostNo || page.topic.maxPostNo || 1,
  );
  const [loadingDirection, setLoadingDirection] = useState<
    "before" | "after" | "anchor" | ""
  >("");
  const [windowError, setWindowError] = useState("");
  const [likeCount, setLikeCount] = useState(page.topic.likeCount);
  const [replyCount, setReplyCount] = useState(page.topic.replyCount);
  const [liked, setLiked] = useState(page.topic.isLiked);
  const [bookmarked, setBookmarked] = useState(page.topic.isBookmarked);
  const [watched, setWatched] = useState(page.topic.isWatched);
  const [topicStatus, setTopicStatus] = useState(page.topic.processStatus);
  const [actionBusy, setActionBusy] = useState("");
  const [actionError, setActionError] = useState("");
  const [composerOpen, setComposerOpen] = useState(false);
  const [composerMinimized, setComposerMinimized] = useState(false);
  const [composerExpanded, setComposerExpanded] = useState(false);
  const [composerContent, setComposerContent] = useState("");
  const [replyTargetId, setReplyTargetId] = useState(0);
  const [editingId, setEditingId] = useState(0);
  const [composerBusy, setComposerBusy] = useState(false);
  const [composerError, setComposerError] = useState("");
  const [pendingDelete, setPendingDelete] = useState<PostPayload>();
  const [deleteBusy, setDeleteBusy] = useState(false);
  const [pendingReport, setPendingReport] = useState<PendingReport>();
  const [reportReason, setReportReason] = useState("spam");
  const [reportNote, setReportNote] = useState("");
  const [reportBusy, setReportBusy] = useState(false);
  const [pendingModeration, setPendingModeration] =
    useState<PendingModeration>();
  const [moderationBusy, setModerationBusy] = useState(false);
  const [imageIndex, setImageIndex] = useState(-1);
  const [imageActualSize, setImageActualSize] = useState(false);
  const [images, setImages] = useState<Array<{ src: string; alt: string }>>([]);
  const [activePostNo, setActivePostNo] = useState(
    firstNo(page.postStream.posts) || 1,
  );
  const sentinel = useRef<HTMLDivElement>(null);
  const postsRef = useRef<HTMLDivElement>(null);
  const windowRevision = useRef(0);
  const [revealPostId, setRevealPostId] = useState(0);
  useEffect(() => () => { windowRevision.current++; }, []);
  const targetMap = useMemo(
    () => new Map(replyTargets.map((target) => [target.id, target])),
    [replyTargets],
  );
  const replyTarget = posts.find((post) => post.id === replyTargetId);
  const maxPostNo = Math.max(
    maxNo,
    ...posts.map((post) => post.postNo || 0),
    1,
  );

  useEffect(() => {
    const title = page.topic.title;
    const tags = page.topic.categories.map((category) => ({
      id: category.id,
      name: category.name,
      color: category.color,
    }));
    setShellHeader({ title, tags, visible: false });
    const titleNode = titleRef.current;
    if (!titleNode || !("IntersectionObserver" in window)) {
      return () =>
        setShellHeader((current) =>
          current.title === title ? emptyShellHeader : current,
        );
    }
    const observer = new IntersectionObserver(
      ([entry]) => {
        setShellHeader((current) =>
          current.title === title
            ? { ...current, visible: !entry?.isIntersecting }
            : current,
        );
      },
      { threshold: 0, rootMargin: "-80px 0px 0px 0px" },
    );
    observer.observe(titleNode);
    return () => {
      observer.disconnect();
      setShellHeader((current) =>
        current.title === title ? emptyShellHeader : current,
      );
    };
  }, [page.topic.categories, page.topic.id, page.topic.title, setShellHeader]);

  const loadWindow = useCallback(
    async (direction: "before" | "after" | "anchor", anchor?: number) => {
      if (loadingDirection) return;
      const revision = ++windowRevision.current;
      setLoadingDirection(direction);
      setWindowError("");
      const previousHeight = document.documentElement.scrollHeight;
      try {
        const input =
          direction === "before"
            ? { topicId: page.topic.id, beforePostNo: beforeNo, limit: 20 }
            : direction === "after"
              ? { topicId: page.topic.id, afterPostNo: afterNo, limit: 20 }
              : { topicId: page.topic.id, anchorPostNo: anchor, limit: 20 };
        const result = await runtime.api.posts.window(input);
        if (revision !== windowRevision.current) return;
        applyWindow(result, direction);
        if (direction === "before")
          requestAnimationFrame(() =>
            window.scrollBy({
              top: document.documentElement.scrollHeight - previousHeight,
              behavior: "auto",
            }),
          );
        if (direction === "anchor")
          requestAnimationFrame(() =>
            document
              .querySelector<HTMLElement>(`[data-post-no="${anchor}"]`)
              ?.scrollIntoView({ block: "center" }),
          );
      } catch (reason) {
        if (revision !== windowRevision.current) return;
        setWindowError(
          reason instanceof Error ? reason.message : t("loadFailed"),
        );
      } finally {
        if (revision === windowRevision.current) setLoadingDirection("");
      }
    },
    [afterNo, beforeNo, loadingDirection, page.topic.id, runtime.api.posts, t],
  );
  function applyWindow(
    result: PostWindowPayload,
    direction: "before" | "after" | "anchor",
  ) {
    setPosts((current) =>
      direction === "anchor" ? result.posts : mergePosts(current, result.posts),
    );
    setReplyTargets((current) => mergeById(current, result.replyTargets || []));
    if (direction !== "after") {
      setHasBefore(result.hasBefore);
      setBeforeNo(result.beforePostNo || firstNo(result.posts));
    }
    if (direction !== "before") {
      setHasAfter(result.hasAfter);
      setAfterNo(result.afterPostNo || lastNo(result.posts));
    }
    setMaxNo((current) => Math.max(result.maxPostNo || 0, current));
  }
  useEffect(() => {
    if (!revealPostId || composerOpen) return;
    const node = postsRef.current?.querySelector<HTMLElement>(`#post-${revealPostId}`);
    if (!node) return;
    const frame = requestAnimationFrame(() => {
      node.scrollIntoView({ block: "start", behavior: "auto" });
      setActivePostNo(Number(node.dataset.postNo) || 1);
      setRevealPostId(0);
    });
    return () => cancelAnimationFrame(frame);
  }, [posts, composerOpen, revealPostId]);
  useEffect(() => {
    const node = sentinel.current;
    if (!node || !hasAfter || !("IntersectionObserver" in window)) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting))
          void loadWindow("after");
      },
      { rootMargin: "400px 0px" },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [hasAfter, loadWindow]);
  useEffect(() => {
    let frame = 0;
    const nodes = Array.from(
      postsRef.current?.querySelectorAll<HTMLElement>("[data-post-no]") || [],
    );
    function measure() {
      frame = 0;
      const marker = Math.min(innerHeight * 0.38, 340);
      let closest = nodes[0];
      let distance = Number.POSITIVE_INFINITY;
      for (const node of nodes) {
        const rect = node.getBoundingClientRect();
        const next =
          rect.top <= marker && rect.bottom >= marker
            ? 0
            : Math.min(
                Math.abs(rect.top - marker),
                Math.abs(rect.bottom - marker),
              );
        if (next < distance) {
          closest = node;
          distance = next;
        }
      }
      if (closest) setActivePostNo(Number(closest.dataset.postNo) || 1);
    }
    function onScroll() {
      if (!frame) frame = requestAnimationFrame(measure);
    }
    window.addEventListener("scroll", onScroll, { passive: true });
    measure();
    return () => {
      window.removeEventListener("scroll", onScroll);
      cancelAnimationFrame(frame);
    };
  }, [posts]);
  useEffect(() => {
    const preload = () => void loadTopicComposer();
    if ("requestIdleCallback" in window) {
      const handle = window.requestIdleCallback(preload, { timeout: 1500 });
      return () => window.cancelIdleCallback(handle);
    }
    const handle = setTimeout(preload, 600);
    return () => clearTimeout(handle);
  }, []);

  async function toggleAction(type: "like" | "bookmark" | "watch") {
    if (actionBusy) return;
    setActionBusy(type);
    setActionError("");
    const previous = { liked, bookmarked, watched, likeCount };
    const next =
      type === "like" ? !liked : type === "bookmark" ? !bookmarked : !watched;
    if (type === "like") {
      setLiked(next);
      setLikeCount((count) => Math.max(0, count + (next ? 1 : -1)));
    } else if (type === "bookmark") setBookmarked(next);
    else setWatched(next);
    try {
      if (type === "like")
        await runtime.api.topics.like(page.topic.id, next ? 1 : 2);
      else if (type === "bookmark")
        await runtime.api.topics.bookmark(page.topic.id, next ? 1 : 2);
      else await runtime.api.topics.watch(page.topic.id, next ? 1 : 2);
    } catch (reason) {
      setLiked(previous.liked);
      setBookmarked(previous.bookmarked);
      setWatched(previous.watched);
      setLikeCount(previous.likeCount);
      setActionError(
        reason instanceof Error ? reason.message : t("actionFailed"),
      );
    } finally {
      setActionBusy("");
    }
  }
  function openComposer(post?: PostPayload) {
    if (!layout.viewer.isAuthenticated || !page.permissions.canPost) {
      void runtime.navigate(loginUrl(runtime.currentUrl));
      return;
    }
    setEditingId(0);
    setReplyTargetId(post?.id || 0);
    setComposerContent("");
    setComposerError("");
    setComposerMinimized(false);
    setComposerOpen(true);
  }
  function editPost(post: PostPayload) {
    if (post.postNo === 1) {
      void runtime.navigate(`/publish?id=${page.topic.id}`);
      return;
    }
    setEditingId(post.id);
    setReplyTargetId(0);
    setComposerContent(post.content);
    setComposerError("");
    setComposerMinimized(false);
    setComposerOpen(true);
  }
  async function submitComposer() {
    const content = composerContent.trim();
    if (!content) {
      setComposerError(t("replyRequired"));
      return;
    }
    setComposerBusy(true);
    setComposerError("");
    try {
      if (editingId) {
        const updated = await runtime.api.posts.update({
          postId: editingId,
          content,
        });
        setPosts((items) =>
          items.map((post) =>
            post.id === editingId
              ? {
                  ...post,
                  content: updated.content,
                  renderedContent: updated.renderedContent,
                  updatedAt: updated.updatedAt,
                  processStatus: updated.processStatus ?? post.processStatus,
                }
              : post,
          ),
        );
        runtime.queueFlash(t("replyUpdated"), "success");
      } else {
        const created = await runtime.api.posts.create({
          topicId: page.topic.id,
          content,
          replyToPostId: replyTargetId,
        });
        if (
          typeof created === "object" &&
          created &&
          "moderationStatus" in created &&
          created.moderationStatus === "rejected"
        ) {
          setComposerError(t("actionFailed"));
          return;
        }
        const result = typeof created === "object" && created ? created : null;
        if (result) {
          const newPost: PostPayload = {
            id: result.id,
            topicId: page.topic.id,
            postNo: result.postNo || maxPostNo + 1,
            content,
            renderedContent: result.renderedContent,
            processStatus: result.processStatus || 0,
            isHidden: false,
            canModerate: false,
            author: {
              id: layout.viewer.id,
              username: layout.viewer.username,
              avatarUrl: layout.viewer.avatarUrl,
            },
            createdAt: new Date().toISOString(),
            replyToPostId: replyTargetId || undefined,
            replyToUserId: replyTarget?.author.id,
            replyToUsername: replyTarget?.author.username,
            isOwnPost: true,
          };
          // A new tail reply must not create a hole in the current window.
          // Start with a navigable local window, then load its server context.
          const revision = ++windowRevision.current;
          setLoadingDirection("");
          setWindowError("");
          setPosts([newPost]);
          setBeforeNo(newPost.postNo);
          setAfterNo(newPost.postNo);
          setHasBefore(newPost.postNo > 1);
          setHasAfter(false);
          if (replyTarget) setReplyTargets((items) => mergeById(items, [{
            id: replyTarget.id,
            postNo: replyTarget.postNo,
            author: replyTarget.author,
            renderedContent: replyTarget.renderedContent,
            unavailable: replyTarget.isHidden,
          }]));
          setMaxNo(current => Math.max(current, newPost.postNo));
          if (newPost.processStatus === 0) setReplyCount((count) => count + 1);
          setComposerOpen(false);
          if (newPost.processStatus === 0) {
            setLoadingDirection("anchor");
            try {
              const result = await runtime.api.posts.window({ topicId: page.topic.id, anchorPostNo: newPost.postNo, limit: 20 });
              if (revision === windowRevision.current) {
                // Retain the submitted reply if it is not yet in the read window.
                if (result.posts.some(post => post.id === newPost.id)) applyWindow(result, "anchor");
              }
            } catch (reason) {
              if (revision === windowRevision.current) setWindowError(reason instanceof Error ? reason.message : t("loadFailed"));
            } finally {
              if (revision === windowRevision.current) setLoadingDirection("");
            }
          }
          if (revision === windowRevision.current) setRevealPostId(newPost.id);
        } else await runtime.refresh();
        runtime.queueFlash(t("replyPosted"), "success");
      }
      setComposerOpen(false);
      setComposerContent("");
      setEditingId(0);
      setReplyTargetId(0);
    } catch (reason) {
      setComposerError(
        reason instanceof Error ? reason.message : t("actionFailed"),
      );
    } finally {
      setComposerBusy(false);
    }
  }
  async function deletePost() {
    if (!pendingDelete || deleteBusy) return;
    setDeleteBusy(true);
    try {
      await runtime.api.posts.delete(pendingDelete.id);
      if (pendingDelete.postNo === 1) {
        await runtime.navigate("/");
        return;
      }
      setPosts((items) => items.filter((post) => post.id !== pendingDelete.id));
      if (pendingDelete.processStatus === 0)
        setReplyCount((count) => Math.max(0, count - 1));
      setPendingDelete(undefined);
      runtime.queueFlash(t("replyDeleted"), "success");
    } catch (reason) {
      setActionError(
        reason instanceof Error ? reason.message : t("actionFailed"),
      );
    } finally {
      setDeleteBusy(false);
    }
  }
  function requestReport(target: PendingReport) {
    if (!layout.viewer.isAuthenticated) {
      void runtime.navigate(
        `/login?next=${encodeURIComponent(runtime.currentUrl)}`,
      );
      return;
    }
    setPendingReport(target);
    setReportReason("spam");
    setReportNote("");
  }
  async function submitReport() {
    if (!pendingReport || reportBusy) return;
    setReportBusy(true);
    try {
      await runtime.api.moderation.report(
        pendingReport.targetType,
        pendingReport.targetId,
        reportReason,
        reportNote,
      );
      setPendingReport(undefined);
      runtime.queueFlash(t("reportSubmitted"), "success");
    } catch (reason) {
      setActionError(
        reason instanceof Error ? reason.message : t("actionFailed"),
      );
    } finally {
      setReportBusy(false);
    }
  }
  async function applyModeration() {
    if (!pendingModeration || moderationBusy) return;
    setModerationBusy(true);
    try {
      if (pendingModeration.target === "topic") {
        await runtime.api.moderation.setTopicStatus(
          pendingModeration.id,
          pendingModeration.action,
        );
        setTopicStatus(pendingModeration.action === "ban" ? 1 : 0);
      } else {
        await runtime.api.moderation.setPostStatus(
          pendingModeration.id,
          pendingModeration.action,
        );
        setPosts((items) =>
          items.map((post) =>
            post.id === pendingModeration.id
              ? {
                  ...post,
                  processStatus: pendingModeration.action === "ban" ? 1 : 0,
                  isHidden: pendingModeration.action === "ban",
                }
              : post,
          ),
        );
      }
      runtime.queueFlash(
        t(
          pendingModeration.action === "ban"
            ? "moderationBanSuccess"
            : "moderationUnbanSuccess",
        ),
        "success",
      );
      setPendingModeration(undefined);
    } catch (reason) {
      setActionError(
        reason instanceof Error ? reason.message : t("actionFailed"),
      );
    } finally {
      setModerationBusy(false);
    }
  }
  function contentClick(event: React.MouseEvent) {
    const image = (event.target as HTMLElement).closest(".gf-prose-post img");
    if (!(image instanceof HTMLImageElement)) return;
    event.preventDefault();
    const all = Array.from(
      postsRef.current?.querySelectorAll<HTMLImageElement>(
        ".gf-prose-post img",
      ) || [],
    ).map((item) => ({
      src: item.currentSrc || item.src,
      alt: item.alt || "",
    }));
    setImages(all);
    setImageActualSize(false);
    setImageIndex(
      Math.max(
        0,
        all.findIndex((item) => item.src === (image.currentSrc || image.src)),
      ),
    );
  }

  return (
    <main className="min-w-0 pb-10">
      <header className="border-b px-4 py-4 lg:mb-4 lg:px-0 lg:pt-0">
        <h1
          ref={titleRef}
          className="break-words text-2xl font-bold leading-tight lg:text-3xl"
        >
          {page.topic.title}
        </h1>
        <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-[13px] text-muted-foreground">
          <Person user={page.topic.author} compact />
          <time
            dateTime={page.topic.createdAt}
            className="inline-flex items-center gap-1.5"
          >
            <Clock className="size-3.5" />
            {formatDate(page.topic.createdAt, runtime.locale)}
          </time>
          {page.topic.categories.map((category) => (
            <GooseLink
              key={category.id}
              href={category.url}
              className="inline-flex items-center gap-1.5 hover:text-primary"
            >
              <span
                className="size-2 rounded-sm"
                style={{ backgroundColor: category.color }}
              />
              {category.name}
            </GooseLink>
          ))}
          <span className="inline-flex items-center gap-1">
            <MessageSquare className="size-3.5" />
            {replyCount}
          </span>
          <span className="inline-flex items-center gap-1">
            <Eye className="size-3.5" />
            {page.topic.viewCount}
          </span>
          <span className="inline-flex items-center gap-1">
            <Heart className="size-3.5" />
            {likeCount}
          </span>
        </div>
      </header>
      {actionError ? (
        <Alert variant="destructive" className="mb-3">
          <AlertDescription>{actionError}</AlertDescription>
        </Alert>
      ) : null}
      <section
        data-event-delegation="image-viewer"
        className="min-w-0 border-b bg-background lg:rounded-xl lg:border xl:grid xl:grid-cols-[minmax(0,1fr)_256px]"
        onClick={contentClick}
      >
        <div ref={postsRef} className="min-w-0">
          {hasBefore ? (
            <div className="px-4 py-3 text-center">
              <Button
                variant="ghost"
                size="sm"
                disabled={Boolean(loadingDirection)}
                onClick={() => void loadWindow("before")}
              >
                {loadingDirection === "before" ? (
                  <Spinner data-icon="inline-start" />
                ) : null}
                {t("loadEarlierReplies")}
              </Button>
            </div>
          ) : null}
          {posts.map((post) => (
            <PostRow
              key={post.id}
              post={post}
              target={
                post.replyToPostId
                  ? targetMap.get(post.replyToPostId)
                  : undefined
              }
              first={post.postNo === 1}
              locale={runtime.locale}
              liked={liked}
              bookmarked={bookmarked}
              watched={watched}
              likeCount={likeCount}
              actionBusy={actionBusy}
              canPost={page.permissions.canPost}
              isOwnTopic={page.permissions.isOwnTopic}
              topicCanModerate={page.permissions.canModerateTopic}
              topicStatus={topicStatus}
              t={t}
              onToggle={toggleAction}
              onReply={() => openComposer(post)}
              onEdit={() => editPost(post)}
              onDelete={() => setPendingDelete(post)}
              onReport={() =>
                requestReport({
                  targetType: post.postNo === 1 ? "topic" : "post",
                  targetId: post.postNo === 1 ? page.topic.id : post.id,
                  title:
                    post.postNo === 1
                      ? page.topic.title
                      : t("replyReportTitle", { no: post.postNo }),
                  excerpt: post.content,
                })
              }
              onModerate={(action) =>
                setPendingModeration({
                  target: post.postNo === 1 ? "topic" : "post",
                  id: post.postNo === 1 ? page.topic.id : post.id,
                  action,
                })
              }
            />
          ))}
          <div
            ref={sentinel}
            className={cn("px-4 py-3 text-center", hasAfter && "border-t")}
          >
            {hasAfter ? (
              <Button
                variant="ghost"
                size="sm"
                disabled={Boolean(loadingDirection)}
                onClick={() => void loadWindow("after")}
              >
                {loadingDirection === "after" ? (
                  <Spinner data-icon="inline-start" />
                ) : null}
                {windowError ? t("retryLoadReplies") : t("loadMoreReplies")}
              </Button>
            ) : posts.length ? (
              <span className="text-xs text-muted-foreground">
                {t("allRepliesShown")}
              </span>
            ) : null}
            {windowError ? (
              <p className="mt-1 text-xs text-destructive">{windowError}</p>
            ) : null}
          </div>
        </div>
        <TopicAside
          page={page}
          replyCount={replyCount}
          active={activePostNo}
          max={maxPostNo}
          busy={Boolean(loadingDirection)}
          posts={posts}
          hasAfter={hasAfter}
          locale={runtime.locale}
          t={t}
          onSelect={(value) => {
            setActivePostNo(value);
            const loaded = posts.find((post) => post.postNo === value);
            if (loaded) {
              document
                .querySelector<HTMLElement>(`[data-post-no="${value}"]`)
                ?.scrollIntoView({ behavior: "smooth", block: "center" });
              return;
            }
            void loadWindow("anchor", value);
          }}
        />
      </section>
      {page.hotTopics.length ? (
        <section className="mt-0 overflow-hidden border-b bg-background lg:mt-4 lg:rounded-xl lg:border">
          <h2 className="border-b px-4 py-3 text-sm font-semibold">
            {t("hotContent")}
          </h2>
          <TopicTable topics={page.hotTopics} showCategories t={homeT} />
        </section>
      ) : null}
      {page.permissions.canPost || !layout.viewer.isAuthenticated ? (
        <div className="pointer-events-none fixed inset-x-0 bottom-4 z-30">
          <div
            data-slot="topic-reply-float-boundary"
            className="mx-auto flex w-full max-w-[1600px] justify-end pl-4 pr-6 lg:pl-8 lg:pr-10"
          >
            <Button
              size="icon-lg"
              className="pointer-events-auto rounded-full shadow-lg"
              aria-label={t("joinDiscussion")}
              onClick={() =>
                page.permissions.canPost
                  ? openComposer()
                  : void runtime.navigate(loginUrl(runtime.currentUrl))
              }
            >
              <Reply />
            </Button>
          </div>
        </div>
      ) : null}
      <Suspense fallback={null}>
        <TopicComposer
          open={composerOpen}
          minimized={composerMinimized}
          expanded={composerExpanded}
          content={composerContent}
          editing={Boolean(editingId)}
          target={replyTarget}
          viewer={layout.viewer}
          topicTitle={page.topic.title}
          busy={composerBusy}
          error={composerError}
          onContent={setComposerContent}
          onSubmit={() => void submitComposer()}
          onClose={() => setComposerOpen(false)}
          onMinimize={() => setComposerMinimized((value) => !value)}
          onExpand={() => setComposerExpanded((value) => !value)}
          onClearTarget={() => setReplyTargetId(0)}
        />
      </Suspense>
      <ConfirmDialog
        open={Boolean(pendingDelete)}
        title={
          pendingDelete?.postNo === 1
            ? t("deleteTopicTitle")
            : t("deleteReplyTitle")
        }
        description={
          pendingDelete?.postNo === 1
            ? t("deleteTopicDescription")
            : t("deleteReplyDescription")
        }
        confirm={t("confirmDelete")}
        busy={deleteBusy}
        destructive
        onClose={() => setPendingDelete(undefined)}
        onConfirm={() => void deletePost()}
      />
      <ConfirmDialog
        open={Boolean(pendingModeration)}
        title={t(
          pendingModeration?.action === "unban"
            ? "moderationUnbanTitle"
            : "moderationBanTitle",
        )}
        description={t(
          pendingModeration?.action === "unban"
            ? "moderationUnbanDescription"
            : "moderationBanDescription",
        )}
        confirm={t(
          pendingModeration?.action === "unban"
            ? "confirmModerationUnban"
            : "confirmModerationBan",
        )}
        busy={moderationBusy}
        destructive
        onClose={() => setPendingModeration(undefined)}
        onConfirm={() => void applyModeration()}
      />
      <Dialog
        open={Boolean(pendingReport)}
        onOpenChange={(open) => !open && setPendingReport(undefined)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("reportTitle")}</DialogTitle>
            <DialogDescription>{pendingReport?.title}</DialogDescription>
          </DialogHeader>
          <ToggleGroup
            type="single"
            value={reportReason}
            onValueChange={(value) => value && setReportReason(value)}
            variant="outline"
            className="flex-wrap"
          >
            {["spam", "abuse", "illegal", "irrelevant", "other"].map(
              (reason) => (
                <ToggleGroupItem key={reason} value={reason}>
                  {t(`reportReasons.${reason}`)}
                </ToggleGroupItem>
              ),
            )}
          </ToggleGroup>
          <Textarea
            value={reportNote}
            onChange={(event) => setReportNote(event.target.value)}
            maxLength={300}
            placeholder={t("reportNotePlaceholder")}
          />
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setPendingReport(undefined)}
            >
              {t("cancel")}
            </Button>
            <Button disabled={reportBusy} onClick={() => void submitReport()}>
              {reportBusy ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <Flag data-icon="inline-start" />
              )}
              {t("submitReport")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog
        open={imageIndex >= 0}
        onOpenChange={(open) => !open && setImageIndex(-1)}
      >
        <DialogContent
          className="flex h-dvh w-full max-w-none flex-col gap-0 rounded-none bg-transparent p-0 ring-0 sm:max-w-none"
          showCloseButton={false}
          onKeyDown={(event) => {
            if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
              event.preventDefault();
              setImageActualSize(false);
              setImageIndex((index) => (index + (event.key === 'ArrowRight' ? 1 : -1) + images.length) % images.length);
            } else if (event.key === '0') {
              event.preventDefault();
              setImageActualSize((value) => !value);
            }
          }}
        >
          <DialogHeader className="sr-only">
            <DialogTitle>{t("imageViewer")}</DialogTitle>
            <DialogDescription>{images[imageIndex]?.alt}</DialogDescription>
          </DialogHeader>
          <div className="flex shrink-0 items-center justify-between gap-2 px-3 py-2">
            <span className="text-sm tabular-nums" aria-live="polite">{imageIndex + 1} / {images.length}</span>
            <div className="flex items-center gap-2">
              <Button variant="secondary" size="sm" aria-pressed={imageActualSize} onClick={() => setImageActualSize((value) => !value)}>
                {imageActualSize ? <Minimize2 data-icon="inline-start" /> : <Maximize2 data-icon="inline-start" />}
                {t(imageActualSize ? 'fitImage' : 'originalImageSize')}
              </Button>
              <Button variant="secondary" size="icon" aria-label={t('close')} onClick={() => setImageIndex(-1)}><X /></Button>
            </div>
          </div>
          <div key={`${imageIndex}-${imageActualSize}`} className="min-h-0 flex-1 overflow-auto p-3 sm:px-16">
            <div className="grid min-h-full min-w-full place-items-center">
              <img
                src={images[imageIndex]?.src}
                alt={images[imageIndex]?.alt || ""}
                role="button"
                tabIndex={0}
                aria-label={`${images[imageIndex]?.alt || t("imageViewer")} · ${t(imageActualSize ? "fitImage" : "originalImageSize")}`}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    setImageActualSize((value) => !value);
                  }
                }}
                style={imageActualSize ? { maxWidth: 'none' } : undefined}
                className={imageActualSize ? 'cursor-zoom-out' : 'max-h-[calc(100dvh-5rem)] max-w-full cursor-zoom-in object-contain'}
                onClick={() => setImageActualSize((value) => !value)}
              />
            </div>
          </div>
          <Button
            variant="secondary"
            size="icon"
            className="absolute left-3 top-1/2 -translate-y-1/2"
            disabled={images.length < 2}
            aria-label={t("previousImage")}
            onClick={() => {
              setImageActualSize(false);
              setImageIndex((index) => (index - 1 + images.length) % images.length);
            }}
          >
            <ChevronLeft />
          </Button>
          <Button
            variant="secondary"
            size="icon"
            className="absolute right-3 top-1/2 -translate-y-1/2"
            disabled={images.length < 2}
            aria-label={t("nextImage")}
            onClick={() => {
              setImageActualSize(false);
              setImageIndex((index) => (index + 1) % images.length);
            }}
          >
            <ChevronRight />
          </Button>
        </DialogContent>
      </Dialog>
    </main>
  );
}

type Translate = ReturnType<typeof useTranslation>["t"];
function PostRow({
  post,
  target,
  first,
  locale,
  liked,
  bookmarked,
  watched,
  likeCount,
  actionBusy,
  canPost,
  isOwnTopic,
  topicCanModerate,
  topicStatus,
  t,
  onToggle,
  onReply,
  onEdit,
  onDelete,
  onReport,
  onModerate,
}: {
  post: PostPayload;
  target?: ReplyTargetPayload;
  first: boolean;
  locale: string;
  liked: boolean;
  bookmarked: boolean;
  watched: boolean;
  likeCount: number;
  actionBusy: string;
  canPost: boolean;
  isOwnTopic: boolean;
  topicCanModerate: boolean;
  topicStatus: number;
  t: Translate;
  onToggle(type: "like" | "bookmark" | "watch"): void;
  onReply(): void;
  onEdit(): void;
  onDelete(): void;
  onReport(): void;
  onModerate(action: "ban" | "unban"): void;
}) {
  const hidden = post.isHidden && !post.canModerate;
  return (
    <article
      id={`post-${post.id}`}
      data-post-no={post.postNo}
      className="group relative grid scroll-mt-24 grid-cols-[42px_minmax(0,1fr)] gap-3 px-4 py-4 after:absolute after:inset-x-4 after:bottom-0 after:h-px after:bg-border last:after:hidden"
    >
      <Person user={post.author} avatarOnly showBadge sticky />
      <div className="min-w-0">
        <header className="mb-1.5 flex min-w-0 items-start justify-between gap-2 text-xs text-muted-foreground">
          <div className="min-w-0">
            <div className="flex min-w-0 items-center gap-2">
              <Person user={post.author} textOnly />
              {first ? (
                <Badge variant="secondary" className="px-1.5 py-0.5 text-xs">
                  {t("originalPost")}
                </Badge>
              ) : null}
              <PostPermalink post={post} className="hidden lg:inline" />
            </div>
            <div className="mt-0.5 flex items-center gap-2 lg:hidden">
              <PostPermalink post={post} />
              <time dateTime={post.createdAt} className="truncate">
                {formatDate(post.createdAt, locale)}
              </time>
            </div>
          </div>
          <div className="flex shrink-0 items-center gap-0.5 lg:gap-1.5">
            {post.isOwnPost ? (
              <>
                <Button
                  variant="ghost"
                  size="icon-xs"
                  className="text-muted-foreground hover:bg-primary/10 hover:text-primary"
                  aria-label={t("edit")}
                  onClick={onEdit}
                >
                  <PencilLine />
                </Button>
                <Button
                  variant="ghost"
                  size="icon-xs"
                  className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                  aria-label={t("delete")}
                  onClick={onDelete}
                >
                  <Trash2 />
                </Button>
              </>
            ) : null}
            {canPost && !post.isHidden ? (
              <Button
                variant="ghost"
                size="icon-xs"
                className="text-muted-foreground hover:bg-primary/10 hover:text-primary"
                aria-label={t("reply")}
                onClick={onReply}
              >
                <Reply />
              </Button>
            ) : null}
            {!first && !post.isOwnPost && !post.isHidden ? (
              <Button
                variant="ghost"
                size="icon-xs"
                className="text-muted-foreground hover:bg-warning/10 hover:text-warning"
                aria-label={t("report")}
                onClick={onReport}
              >
                <Flag />
              </Button>
            ) : null}
            {!first && post.canModerate ? (
              <Button
                variant="ghost"
                size="icon-xs"
                className={cn(
                  "text-muted-foreground",
                  post.processStatus === 1
                    ? "hover:bg-primary/10 hover:text-primary"
                    : "hover:bg-destructive/10 hover:text-destructive",
                )}
                aria-label={
                  post.processStatus === 1
                    ? t("moderationUnban")
                    : t("moderationBan")
                }
                onClick={() =>
                  onModerate(post.processStatus === 1 ? "unban" : "ban")
                }
              >
                {post.processStatus === 1 ? <RotateCcw /> : <Ban />}
              </Button>
            ) : null}
            <time
              dateTime={post.createdAt}
              className="hidden w-36 shrink-0 text-right lg:block"
            >
              {formatDate(post.createdAt, locale)}
            </time>
          </div>
        </header>
        {post.replyToPostId ? <ReplyReference topicId={post.topicId} target={target} t={t} /> : null}
        {hidden ? (
          <div className="rounded-lg bg-muted px-3 py-2 text-sm text-muted-foreground">
            {t("hiddenReplyPlaceholder")}
          </div>
        ) : (
          <RenderedContent html={post.renderedContent} />
        )}
        {post.isHidden && post.canModerate ? (
          <Badge variant="secondary" className="mt-2">
            {t("hiddenReplyBadge")}
          </Badge>
        ) : null}
        {post.processStatus === 1 && post.isOwnPost && !post.isHidden ? (
          <Badge variant="secondary" className="mt-2">
            {t("pendingReviewBadge")}
          </Badge>
        ) : null}
        {post.updatedAt && post.updatedAt !== post.createdAt ? (
          <p className="mt-2 text-xs text-muted-foreground">
            {t("editedAt", { time: formatDate(post.updatedAt, locale) })}
          </p>
        ) : null}
        {first ? (
          <div className="mt-4 flex flex-wrap gap-2 border-t pt-3">
            <ActionButton
              tone="like"
              active={liked}
              busy={actionBusy === "like"}
              icon={Heart}
              label={likeCount ? String(likeCount) : t("like")}
              onClick={() => onToggle("like")}
            />
            <ActionButton
              tone="bookmark"
              active={bookmarked}
              busy={actionBusy === "bookmark"}
              icon={Bookmark}
              label={t(bookmarked ? "bookmarked" : "bookmark")}
              onClick={() => onToggle("bookmark")}
            />
            <ActionButton
              tone="watch"
              active={watched}
              busy={actionBusy === "watch"}
              icon={Bell}
              label={t(watched ? "watched" : "watch")}
              onClick={() => onToggle("watch")}
            />
            {!isOwnTopic ? (
              <Button variant="ghost" size="sm" onClick={onReport}>
                <Flag data-icon="inline-start" />
                {t("report")}
              </Button>
            ) : null}
            {topicCanModerate ? (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onModerate(topicStatus === 1 ? "unban" : "ban")}
              >
                {topicStatus === 1 ? (
                  <RotateCcw data-icon="inline-start" />
                ) : (
                  <Ban data-icon="inline-start" />
                )}
                {t(topicStatus === 1 ? "moderationUnban" : "moderationBan")}
              </Button>
            ) : null}
          </div>
        ) : null}
      </div>
    </article>
  );
}
function PostPermalink({
  post,
  className,
}: {
  post: PostPayload;
  className?: string;
}) {
  return (
    <GooseLink
      href={`/p/post/${post.topicId}${post.postNo > 1 ? `/${post.postNo}` : ""}#post-${post.id}`}
      className={cn(
        "shrink-0 font-semibold tabular-nums text-muted-foreground hover:text-primary",
        className,
      )}
    >
      #{post.postNo}
    </GooseLink>
  );
}
function ActionButton({
  tone,
  active,
  busy,
  icon: Icon,
  label,
  onClick,
}: {
  tone: "like" | "bookmark" | "watch";
  active: boolean;
  busy: boolean;
  icon: typeof Heart;
  label: string;
  onClick(): void;
}) {
  return (
    <Button
      data-tone={tone}
      variant="ghost"
      size="sm"
      className={cn(
        "px-2.5 text-muted-foreground hover:bg-muted hover:text-foreground",
        active &&
          (tone === "like"
            ? "bg-destructive/10 text-destructive hover:bg-destructive/15 hover:text-destructive"
            : tone === "bookmark"
              ? "bg-primary/10 text-primary hover:bg-primary/15 hover:text-primary"
              : "bg-success/10 text-success hover:bg-success/15 hover:text-success"),
      )}
      disabled={busy}
      onClick={onClick}
    >
      <Icon data-icon="inline-start" fill={active ? "currentColor" : "none"} />
      {label}
    </Button>
  );
}
function Person({
  user,
  avatarOnly = false,
  textOnly = false,
  compact = false,
  small = false,
  showBadge = false,
  sticky = false,
}: {
  user: PostPayload["author"];
  avatarOnly?: boolean;
  textOnly?: boolean;
  compact?: boolean;
  small?: boolean;
  showBadge?: boolean;
  sticky?: boolean;
}) {
  const avatarClassName = compact
    ? "size-5"
    : small
      ? "size-8"
      : "size-9 lg:size-10";
  return (
    <UserCardPopover user={user}>
      <GooseLink
        href={`/u/${user.id}`}
        className={cn(
          "inline-flex min-w-0 items-center gap-2 font-medium hover:text-primary",
          avatarOnly && avatarClassName,
          sticky && "sticky top-19 self-start mt-1",
          textOnly && "truncate",
        )}
        title={user.username}
      >
        {textOnly ? (
          user.username
        ) : (
          showBadge ? (
            <ProfileAvatar
              src={user.avatarUrl}
              name={user.username}
              badge={user.wornBadge}
              className={cn(
                avatarClassName,
                "rounded-full",
              )}
              framed={false}
              compactBadge
            />
          ) : (
            <Avatar
              className={cn(
                avatarClassName,
                small && "transition hover:ring-1 hover:ring-primary/40",
              )}
            >
              <AvatarImage src={user.avatarUrl} alt="" />
              <AvatarFallback>
                {user.username.slice(0, 1).toUpperCase()}
              </AvatarFallback>
            </Avatar>
          )
        )}
        {!avatarOnly && !textOnly ? user.username : null}
      </GooseLink>
    </UserCardPopover>
  );
}
function TopicAside({
  page,
  replyCount,
  active,
  max,
  busy,
  posts,
  hasAfter,
  locale,
  t,
  onSelect,
}: {
  page: TopicDetailProps;
  replyCount: number;
  active: number;
  max: number;
  busy: boolean;
  posts: PostPayload[];
  hasAfter: boolean;
  locale: string;
  t: Translate;
  onSelect(value: number): void;
}) {
  const currentPost = posts.find((post) => post.postNo === active);
  const endDate = hasAfter
    ? page.topic.updatedAt
    : posts.at(-1)?.createdAt || page.topic.updatedAt;
  return (
    <aside className="hidden min-w-0 xl:block">
      <div className="sticky top-19">
        <h2 className="px-4 py-4 text-sm font-semibold text-muted-foreground">
          {t("overview")}
        </h2>
        <dl className="flex flex-col gap-4 border-t px-4 py-5 text-sm">
          <Stat label={t("replyCount")} value={replyCount} />
          <Stat label={t("viewCount")} value={page.topic.viewCount} />
          <Stat
            label={t("participants")}
            value={page.topic.participants.length}
          />
        </dl>
        {page.topic.participants.length ? (
          <div className="border-t px-4 py-4">
            <h3 className="mb-3 text-sm font-semibold text-muted-foreground">
              {t("activeParticipants")}
            </h3>
            <div className="flex flex-wrap gap-1.5">
              {page.topic.participants.map((user) => (
                <Person key={user.id} user={user} avatarOnly small />
              ))}
            </div>
          </div>
        ) : null}
        {max > 1 ? (
          <PostPositionRail
            current={active}
            max={max}
            startLabel={formatRailDate(page.topic.createdAt, locale)}
            endLabel={formatRailDate(endDate, locale)}
            currentLabel={
              currentPost
                ? formatRailDate(currentPost.createdAt, locale)
                : undefined
            }
            busy={busy}
            locale={locale}
            t={t}
            onSelect={onSelect}
          />
        ) : null}
      </div>
    </aside>
  );
}

function PostPositionRail({
  current,
  max,
  startLabel,
  endLabel,
  currentLabel,
  busy,
  locale,
  t,
  onSelect,
}: {
  current: number;
  max: number;
  startLabel: string;
  endLabel: string;
  currentLabel?: string;
  busy: boolean;
  locale: string;
  t: Translate;
  onSelect(value: number): void;
}) {
  const [preview, setPreview] = useState<number>();
  const displayed = Math.min(max, Math.max(1, preview ?? current));
  const progress = max <= 1 ? 0 : (displayed - 1) / (max - 1);
  const thumbTop = Math.min(164, Math.max(0, progress * 192 - 14));
  const indicatorTop = Math.min(156, Math.max(0, thumbTop - 4));

  return (
    <div className="border-t px-3 py-2">
      <Button
        type="button"
        variant="ghost"
        className="mb-3 block max-w-full truncate px-0 text-left text-base font-semibold leading-tight text-muted-foreground hover:bg-transparent hover:text-foreground"
        title={t("earliestContent")}
        aria-label={`${t("earliestContent")}: ${startLabel}`}
        disabled={busy}
        onClick={() => onSelect(1)}
      >
        {startLabel}
      </Button>
      <div className="grid h-48 grid-cols-[24px_minmax(0,1fr)] gap-3">
        <Slider
          orientation="vertical"
          inverted
          aria-label={t("replyPosition")}
          value={[displayed]}
          min={1}
          max={max}
          step={1}
          disabled={busy}
          className="h-full [&_[data-slot=slider-range]]:bg-transparent [&_[data-slot=slider-thumb]]:h-7 [&_[data-slot=slider-thumb]]:w-1.5 [&_[data-slot=slider-thumb]]:border-0 [&_[data-slot=slider-thumb]]:bg-success [&_[data-slot=slider-track]]:w-px [&_[data-slot=slider-track]]:bg-border"
          onValueChange={(values) => setPreview(values[0])}
          onValueCommit={(values) => {
            setPreview(undefined);
            if (values[0]) onSelect(values[0]);
          }}
        />
        <div className="relative min-w-0">
          <div
            className="absolute left-0 min-w-0 transition-[top] duration-200 ease-out"
            style={{ top: indicatorTop }}
          >
            <div className="whitespace-nowrap text-base font-black leading-none tabular-nums">
              {displayed} / {max.toLocaleString(locale)}
            </div>
            {currentLabel && preview === undefined && !busy ? (
              <div className="mt-2 truncate text-sm font-semibold leading-tight text-muted-foreground">
                {currentLabel}
              </div>
            ) : null}
          </div>
        </div>
      </div>
      <Button
        type="button"
        variant="ghost"
        className="mt-3 block max-w-full truncate px-0 text-left text-base font-semibold leading-tight text-muted-foreground hover:bg-transparent hover:text-foreground"
        title={t("latestReply")}
        aria-label={`${t("latestReply")}: ${endLabel}`}
        disabled={busy}
        onClick={() => onSelect(max)}
      >
        {endLabel}
      </Button>
    </div>
  );
}
function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <dt className="font-semibold text-muted-foreground">{label}</dt>
      <dd className="font-semibold tabular-nums">{value.toLocaleString()}</dd>
    </div>
  );
}
function ConfirmDialog({
  open,
  title,
  description,
  confirm,
  busy,
  destructive = false,
  onClose,
  onConfirm,
}: {
  open: boolean;
  title: string;
  description: string;
  confirm: string;
  busy: boolean;
  destructive?: boolean;
  onClose(): void;
  onConfirm(): void;
}) {
  const { t } = useTranslation("topic");
  return (
    <Dialog open={open} onOpenChange={(value) => !value && onClose()}>
      <DialogContent showCloseButton={false}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" disabled={busy} onClick={onClose}>
            {t("cancel")}
          </Button>
          <Button
            variant={destructive ? "destructive" : "default"}
            disabled={busy}
            onClick={onConfirm}
          >
            {busy ? <Spinner data-icon="inline-start" /> : null}
            {confirm}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
function firstNo(posts: PostPayload[]) {
  return posts[0]?.postNo || 0;
}
function lastNo(posts: PostPayload[]) {
  return posts.at(-1)?.postNo || 0;
}
function mergeById<T extends { id: number }>(current: T[], incoming: T[]) {
  const map = new Map(current.map((item) => [item.id, item]));
  for (const item of incoming) map.set(item.id, item);
  return [...map.values()];
}
function mergePosts(current: PostPayload[], incoming: PostPayload[]) {
  return mergeById(current, incoming).sort(
    (left, right) => left.postNo - right.postNo,
  );
}
function formatDate(value: string, locale: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : new Intl.DateTimeFormat(locale, {
        dateStyle: "medium",
        timeStyle: "short",
      }).format(date);
}

function formatRailDate(value: string, locale: string) {
  const date = new Date(value.replace(" ", "T"));
  if (Number.isNaN(date.getTime())) return value.slice(0, 10);
  const now = new Date();
  return new Intl.DateTimeFormat(locale, {
    year: date.getFullYear() === now.getFullYear() ? undefined : "numeric",
    month: "short",
    day: "numeric",
  }).format(date);
}

function loginUrl(currentUrl: string) {
  const current = new URL(currentUrl, window.location.href);
  return `/login?next=${encodeURIComponent(`${current.pathname}${current.search}${current.hash}`)}`;
}
