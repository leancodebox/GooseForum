import { avatarInitial } from "@gooseforum/ui/lib/avatar-initial";
import { formatDate } from "../date-format";
import { AttachmentList } from "../editor/attachment-editor";
import {
  lazy,
  Suspense,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useId,
  useRef,
  useState,
} from "react";
import type { ReactNode, RefObject } from "react";
import { createPortal } from "react-dom";
import type {
  LayoutPayload,
  PostPayload,
  PostRevision,
  PostWindowPayload,
  ReplyTargetPayload,
  TopicDetailProps,
} from "@gooseforum/client";
import {
  Ban,
  Bell,
  BellRing,
  Bookmark,
  ChevronLeft,
  ChevronRight,
  Clock,
  Eye,
  Heart,
  Maximize2,
  MessageCircle,
  MessageCircleExclamationPoint,
  MessageSquare,
  Minimize2,
  PenLine,
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
import { Input } from "@gooseforum/ui/components/input";
import { Label } from "@gooseforum/ui/components/label";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription, SheetTrigger } from "@gooseforum/ui/components/sheet";
import { SliderPrimitive } from "@gooseforum/ui/components/slider";
import { usePostPosition } from "../topics/use-post-position";
import { Spinner } from "@gooseforum/ui/components/spinner";
import { Textarea } from "@gooseforum/ui/components/textarea";
import { ToggleGroup, ToggleGroupItem } from "@gooseforum/ui/components/toggle-group";
import { cn } from "@gooseforum/ui/lib/utils";
import { GooseLink, useGooseRuntime, useLatestRequest } from "@gooseforum/runtime";
import { useServerErrorMessage } from "@gooseforum/runtime/i18n/server-error";
import { RenderedContent } from "../content/rendered-content";
import { readImageMetadata } from "../content/image-metadata";
import { emptyShellHeader, useShellHeader } from "../layout/shell-header";
import { TopicTable } from "../topics/topic-list";
import { ReplyReference } from "../topics/reply-reference";
import { LocalDraftPanel } from "../drafts/local-draft-panel";
import { clearAccountDrafts, draftOwner, draftFingerprint, type DraftValues, type LocalDraft } from "../drafts/local-draft-store";
import { GooseClientError } from "@gooseforum/client";
import { useLocalDraft } from "../drafts/use-local-draft";
import { useDraftReplyTargets } from "../drafts/use-draft-reply-targets";
import { UserCardPopover } from "../users/user-card-popover";
import { userDisplayName } from "../users/display-name";
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
const topicReadingTop = 80;

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
  const serverError = useServerErrorMessage();
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
  const windowLoading = useRef<"before" | "after" | "anchor" | "reveal" | "">("");
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
  const [editVersion, setEditVersion] = useState<number>();
  const [composerContent, setComposerContent] = useState("");
  const [replyTargetId, setReplyTargetId] = useState(0);
  const [editingId, setEditingId] = useState(0);
  const sourceVersion = 1 as const;
  const [draftBranch, setDraftBranch] = useState(0);
  const [composerBase, setComposerBase] = useState<DraftValues>({ version: undefined, title: "", content: "", categoryIds: [], replyTargetId: 0, sourceVersion: 1 });
  const composerValues = useMemo(() => ({ version: editVersion, title: "", content: composerContent, categoryIds: [], replyTargetId, sourceVersion }), [composerContent, replyTargetId, sourceVersion, editVersion]);
  const localDraft = useLocalDraft({ userId: layout.viewer.id, kind: editingId ? "edit-reply" : "new-reply", objectId: editingId || page.topic.id, topicId: page.topic.id, branch: draftBranch, base: composerBase, values: composerValues, enabled: composerOpen });
  const [composerBusy, setComposerBusy] = useState(false);
  const composerSubmitting = useRef(false);
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
  const [historyPost, setHistoryPost] = useState<PostPayload>();
  const [historyItems, setHistoryItems] = useState<PostRevision[]>([]);
  const [historyCursor, setHistoryCursor] = useState(0);
  const [historyHasMore, setHistoryHasMore] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [historyError, setHistoryError] = useState(false);
  const beginHistoryRequest = useLatestRequest();
  const [imageIndex, setImageIndex] = useState(-1);
  const [imageActualSize, setImageActualSize] = useState(false);
  const imageViewport = useRef<HTMLDivElement>(null);
  const suppressImageClick = useRef(false);
  const imageBackdropPress = useRef<{ x: number; y: number; pointerId: number } | null>(null);
  const dismissImageOnClick = useRef(false);
  const imageDrag = useRef<{ x: number; y: number; left: number; top: number; moved: boolean } | null>(null);
  useEffect(() => {
    imageDrag.current = null;
    suppressImageClick.current = false;
    imageBackdropPress.current = null;
    dismissImageOnClick.current = false;
    const viewport = imageViewport.current;
    if (!viewport) return;
    viewport.scrollLeft = imageActualSize ? (viewport.scrollWidth - viewport.clientWidth) / 2 : 0;
    viewport.scrollTop = imageActualSize ? (viewport.scrollHeight - viewport.clientHeight) / 2 : 0;
  }, [imageActualSize, imageIndex]);
  const [images, setImages] = useState<Array<{ src: string; alt: string } & ReturnType<typeof readImageMetadata>>>([]);
  const [activePostNo, setActivePostNo] = useState(
    firstNo(page.postStream.posts) || 1,
  );
  const sentinel = useRef<HTMLDivElement>(null);
  const postsRef = useRef<HTMLDivElement>(null);
  const firstTopicReply = useRef<HTMLButtonElement>(null);
  const [showFloatingReply, setShowFloatingReply] = useState(false);
  const [topicControls, setTopicControls] = useState<HTMLDivElement | null>(null);
  const prependAnchor = useRef<{ node: HTMLElement; top: number; isCurrent(): boolean } | null>(null);
  const beginWindowRequest = useLatestRequest();
  const [revealPostId, setRevealPostId] = useState(0);
  useEffect(() => {
    windowLoading.current = "";
    // eslint-disable-next-line react-hooks/set-state-in-effect -- Activity resumes with preserved state but invalidated window requests.
    setLoadingDirection("");
    return () => { windowLoading.current = ""; prependAnchor.current = null; };
  }, []);
  const targetMap = useMemo(
    () => new Map(replyTargets.map((target) => [target.id, target])),
    [replyTargets],
  );
  const draftTargets = useDraftReplyTargets({ candidates: localDraft.candidates, posts, references: replyTargets, api: runtime.api.posts, topicId: page.topic.id, context: `${layout.viewer.id}:${page.topic.id}:${draftBranch}`, enabled: composerOpen && page.permissions.canPost });
  const loadHistory = useCallback(async (postId: number, cursor = 0, append = false) => {
    const isCurrent = beginHistoryRequest();
    setHistoryLoading(true);
    setHistoryError(false);
    try {
      const result = await runtime.api.posts.revisions(postId, cursor, 20);
      if (!isCurrent()) return;
      setHistoryItems((current) => append ? [...current, ...result.items] : result.items);
      setHistoryCursor(result.nextCursor);
      setHistoryHasMore(result.hasMore);
    } catch {
      if (isCurrent()) setHistoryError(true);
    } finally {
      if (isCurrent()) setHistoryLoading(false);
    }
  }, [runtime.api.posts, beginHistoryRequest]);
  const openHistory = useCallback((post: PostPayload) => {
    setHistoryPost(post);
    setHistoryItems([]);
    setHistoryCursor(0);
    setHistoryHasMore(false);
    setHistoryError(false);
    void loadHistory(post.id);
  }, [loadHistory]);
  const replyTarget = posts.find((post) => post.id === replyTargetId) || draftTargets.target(replyTargetId).post;
  const composerDirty = JSON.stringify(composerValues) !== JSON.stringify(composerBase);
  const composerDirtyRef = useRef(false);
  const allowComposerLeave = useRef(false);
  composerDirtyRef.current = composerDirty;
  const registerNavigationBlocker = runtime.registerNavigationBlocker;
  useEffect(() => { allowComposerLeave.current = false; }, [composerValues]);
  useEffect(() => {
    const beforeUnload = (event: BeforeUnloadEvent) => {
      if (composerDirtyRef.current && !allowComposerLeave.current) { event.preventDefault(); event.returnValue = ""; }
    };
    window.addEventListener("beforeunload", beforeUnload);
    const beforeLogout = (event: Event) => {
      if (composerDirtyRef.current && !window.confirm(t("localDraft.leave", { ns: "publish" }))) event.preventDefault();
      else allowComposerLeave.current = true;
    };
    window.addEventListener("goose:before-logout", beforeLogout);
    const remove = registerNavigationBlocker?.(() => !composerDirtyRef.current || window.confirm(t("localDraft.leave", { ns: "publish" })));
    return () => { window.removeEventListener("beforeunload", beforeUnload); window.removeEventListener("goose:before-logout", beforeLogout); remove?.(); };
  }, [registerNavigationBlocker, t]);
  const openedLocalCopy = useRef(false);
  useEffect(() => {
    if (openedLocalCopy.current || !layout.viewer.isAuthenticated || !page.permissions.canPost) return;
    const query = new URL(runtime.currentUrl, window.location.origin).searchParams;
    const postId = Number(query.get("localEditPost"));
    if (!query.has("localReply") && !(postId > 0)) return;
    openedLocalCopy.current = true;
    if (!postId) { openComposer(); return; }
    let disposed = false;
    void runtime.api.posts.window({ topicId: page.topic.id, anchorPostId: postId, limit: 20 }).then((result) => {
      if (disposed) return;
      const post = result.posts.find((post) => post.id === postId && post.isOwnPost && !post.isHidden);
      if (post) { applyWindow(result, "anchor"); editPost(post); }
    }).catch((reason) => { if (!disposed) setActionError(serverError(reason, t("loadFailed"))); });
    return () => { disposed = true; openedLocalCopy.current = false; };
    // This is an explicit entry from the local draft list, only once per page.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [layout.viewer.isAuthenticated, page.permissions.canPost, page.topic.id, runtime.currentUrl]);
  const maxPostNo = Math.max(
    maxNo,
    ...posts.map((post) => post.postNo || 0),
    1,
  );
  const canReplyToTopic = page.permissions.canPost || !layout.viewer.isAuthenticated;
  const showFirstTopicReply = canReplyToTopic && posts.some((post) => post.postNo === 1);
  const showDockedReply = canReplyToTopic && (!showFirstTopicReply || showFloatingReply) && !composerOpen;
  const showMobilePosition = maxPostNo > 1 && !composerOpen;
  const shareTopicEndRow = !hasAfter && !windowError && posts.length > 0;

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
      { threshold: 0, rootMargin: `-${topicReadingTop}px 0px 0px 0px` },
    );
    observer.observe(titleNode);
    return () => {
      observer.disconnect();
      setShellHeader((current) =>
        current.title === title ? emptyShellHeader : current,
      );
    };
  }, [page.topic.categories, page.topic.id, page.topic.title, setShellHeader]);

  useEffect(() => {
    const title = titleRef.current;
    const inlineReply = firstTopicReply.current;
    if (!showFirstTopicReply || !title || !inlineReply || !("IntersectionObserver" in window)) return;

    const observer = new IntersectionObserver(([entry]) => {
      if (!entry?.rootBounds) return;
      const { top, bottom } = entry.rootBounds;
      const bounds = inlineReply.getBoundingClientRect();
      setShowFloatingReply(
        title.getBoundingClientRect().bottom < top &&
        (bounds.bottom < top || bounds.top > bottom),
      );
    }, { rootMargin: `-${topicReadingTop}px 0px 0px 0px` });
    observer.observe(title);
    observer.observe(inlineReply);
    return () => observer.disconnect();
  }, [showFirstTopicReply]);

  const loadWindow = useCallback(
    async (direction: "before" | "after" | "anchor", anchor?: number) => {
      if (windowLoading.current && direction !== "anchor") return;
      windowLoading.current = direction;
      const isCurrent = beginWindowRequest();
      setLoadingDirection(direction);
      setWindowError("");
      try {
        const input =
          direction === "before"
            ? { topicId: page.topic.id, beforePostNo: beforeNo, limit: 20 }
            : direction === "after"
              ? { topicId: page.topic.id, afterPostNo: afterNo, limit: 20 }
              : { topicId: page.topic.id, anchorPostNo: anchor, limit: 20 };
        const result = await runtime.api.posts.window(input);
        if (!isCurrent()) return;
        if (direction === "before") {
          // Capture at response time so reading during the request is preserved.
          const line = Math.min(window.innerHeight * 0.38, 340);
          const visible = Array.from(postsRef.current?.querySelectorAll<HTMLElement>("[data-post-no]") || [])
            .map(node => ({ node, rect: node.getBoundingClientRect() }))
            .filter(({ rect }) => rect.bottom > 0 && rect.top < window.innerHeight);
          const anchor = visible.find(({ rect }) => rect.bottom > line) || visible[0];
          prependAnchor.current = anchor ? { node: anchor.node, top: anchor.rect.top, isCurrent } : null;
        }
        applyWindow(result, direction);
        if (direction === "anchor") {
          return result.posts.find(post => (post.postNo || 0) >= (anchor || 1))?.postNo
            ?? result.posts.at(-1)?.postNo;
        }
      } catch (reason) {
        if (!isCurrent()) return;
        setWindowError(
          serverError(reason, t("loadFailed")),
        );
      } finally {
        if (isCurrent()) { windowLoading.current = ""; setLoadingDirection(""); }
      }
    },
    [afterNo, beforeNo, beginWindowRequest, page.topic.id, runtime.api.posts, serverError, t],
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
  useLayoutEffect(() => {
    const anchor = prependAnchor.current;
    prependAnchor.current = null;
    if (!anchor?.isCurrent() || !anchor.node.isConnected || !postsRef.current?.contains(anchor.node)) return;
    // Correct only the remaining displacement. Native scroll anchoring may have
    // already kept this post in place, in which case no second compensation is needed.
    const offset = anchor.node.getBoundingClientRect().top - anchor.top;
    if (Math.abs(offset) > 0.5) window.scrollBy({ top: offset, behavior: "instant" });
  }, [posts]);
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
  const cancelAnchorLoad = useCallback(() => {
    if (windowLoading.current !== "anchor") return;
    beginWindowRequest();
    windowLoading.current = "";
    setLoadingDirection("");
  }, [beginWindowRequest]);
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
    if (type === "watch" && !layout.viewer.isAuthenticated) {
      void runtime.navigate(loginUrl(runtime.currentUrl));
      return;
    }
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
        serverError(reason, t("actionFailed")),
      );
    } finally {
      setActionBusy("");
    }
  }
  function openComposer(post?: PostPayload) {
    if (composerSubmitting.current) return;
    if (!layout.viewer.isAuthenticated || !page.permissions.canPost) {
      void runtime.navigate(loginUrl(runtime.currentUrl));
      return;
    }
    if (composerDirty && !window.confirm(t("localDraft.leave", { ns: "publish" }))) return;
    setEditingId(0);
    setReplyTargetId(post?.id || 0);
    setComposerContent(""); setEditVersion(undefined);
    setComposerBase({ version: undefined, title: "", content: "", categoryIds: [], replyTargetId: post?.id || 0, sourceVersion: 1 });
    setDraftBranch((value) => value + 1);
    setComposerError("");
    setComposerMinimized(false);
    setComposerOpen(true);
  }
  function editPost(post: PostPayload) {
    if (composerSubmitting.current) return;
    if (post.postNo === 1) {
      void runtime.navigate(`/publish?id=${page.topic.id}`);
      return;
    }
    if (composerDirty && !window.confirm(t("localDraft.leave", { ns: "publish" }))) return;
    setEditingId(post.id);
    setReplyTargetId(0);
    setComposerContent(post.content); setEditVersion(post.version);
    setComposerBase({ version: post.version, title: "", content: post.content, categoryIds: [], replyTargetId: 0, sourceVersion: 1 });
    setDraftBranch((value) => value + 1);
    setComposerError("");
    setComposerMinimized(false);
    setComposerOpen(true);
  }
  async function submitComposer() {
    if (localDraft.recoveryPending) return;
    if (composerSubmitting.current) return;
    const content = composerContent.trim();
    if (!content) {
      setComposerError(t("replyRequired"));
      return;
    }
    composerSubmitting.current = true;
    setComposerBusy(true);
    setComposerError("");
    const submission = localDraft.writer?.capture(composerValues);
    let canonical: DraftValues = composerValues;
    try {
      if (editingId) {
        const updated = await runtime.api.posts.update({
          postId: editingId, version: editVersion,
          content,
          sourceVersion,
        });
        if (updated.moderationStatus === "rejected") { setComposerError(t("actionFailed")); return; }
        canonical = { ...composerValues, version: updated.version, content: updated.content, sourceVersion: updated.sourceVersion ?? sourceVersion };
        setPosts((items) =>
          items.map((post) =>
            post.id === editingId
              ? {
                  ...post,
                  content: updated.content, attachments: updated.attachments, version: updated.version,
                  sourceVersion: updated.sourceVersion ?? sourceVersion,
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
          sourceVersion,
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
          canonical = { ...composerValues, content: result.content ?? content, sourceVersion: result.sourceVersion ?? sourceVersion };
          const newPost: PostPayload = {
            id: result.id,
            topicId: page.topic.id,
            postNo: result.postNo || maxPostNo + 1,
            content: canonical.content, attachments: result.attachments, version: result.version,
            sourceVersion: canonical.sourceVersion,
            renderedContent: result.renderedContent,
            processStatus: result.processStatus || 0,
            isHidden: false,
            canModerate: false,
            author: {
              id: layout.viewer.id,
              username: layout.viewer.username,
              nickname: layout.viewer.nickname,
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
          const isCurrent = beginWindowRequest();
          windowLoading.current = "";
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
          if (newPost.processStatus === 0) {
            windowLoading.current = "reveal";
            setLoadingDirection("anchor");
            try {
              const result = await runtime.api.posts.window({ topicId: page.topic.id, anchorPostNo: newPost.postNo, limit: 20 });
              if (isCurrent()) {
                // Retain the submitted reply if it is not yet in the read window.
                if (result.posts.some(post => post.id === newPost.id)) applyWindow(result, "anchor");
              }
            } catch (reason) {
              if (isCurrent())
                setWindowError(serverError(reason, t("loadFailed")));
            } finally {
              if (isCurrent()) { windowLoading.current = ""; setLoadingDirection(""); }
            }
          }
          if (isCurrent()) setRevealPostId(newPost.id);
        } else await runtime.refresh();
        runtime.queueFlash(t("replyPosted"), "success");
      }
      const unchanged = !submission || await localDraft.writer!.submitted(submission, canonical);
      setComposerBase(canonical);
      if (!unchanged) {
        if (editingId) setEditVersion(canonical.version);
        return;
      }
      composerDirtyRef.current = false;
      setComposerOpen(false);
      setComposerContent(""); setEditVersion(undefined);
      setEditingId(0);
      setReplyTargetId(0);
      setComposerBase({ version: undefined, title: "", content: "", categoryIds: [], replyTargetId: 0, sourceVersion: 1 });
    } catch (reason) {
      setComposerError(
        serverError(reason, t("actionFailed")),
      );
      if (reason instanceof GooseClientError && (reason.status === 401 || reason.messageCode === "auth.required")) await clearAccountDrafts(draftOwner(layout.viewer.id));
    } finally {
      composerSubmitting.current = false;
      setComposerBusy(false);
    }
  }
  function availableDraft(draft: LocalDraft) {
    if (composerBusy || !page.permissions.canPost) return false;
    if (editingId && !posts.some((post) => post.id === editingId && post.isOwnPost && !post.isHidden)) return false;
    return draftTargets.target(draft.replyTargetId).status === "available";
  }
  function restoreDraft(draft: LocalDraft) {
    if (composerSubmitting.current) return false;
    if (!availableDraft(draft)) return false;
    if (composerDirty && !window.confirm(t("localDraft.confirmRestore", { ns: "publish" }))) return false;
    localDraft.restore(draft);
    setEditVersion(draft.version ?? composerBase.version); setComposerContent(draft.content); setReplyTargetId(draft.replyTargetId);
    return true;
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
        serverError(reason, t("actionFailed")),
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
        serverError(reason, t("actionFailed")),
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
        serverError(reason, t("actionFailed")),
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
      ...readImageMetadata(item),
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
  const topicReplyButton = (floating = false) => (
    <Button
      ref={floating ? undefined : firstTopicReply}
      data-slot={floating ? "topic-reply-floating" : "topic-reply-first"}
      size={floating ? "lg" : "sm"}
      variant={floating ? "outline" : "default"}
      className={cn(
        "rounded-full",
        floating
          ? "pointer-events-auto ml-auto min-w-0 max-w-[34%] shrink border-primary/25 px-2.5 text-primary shadow-lg hover:text-primary sm:px-4 dark:border-primary/25 dark:bg-background dark:hover:bg-muted"
          : "ml-auto h-9 px-3 lg:h-7",
      )}
      disabled={composerBusy}
      onClick={() => openComposer()}
    >
      <MessageCircle />
      <span className={floating ? "truncate" : undefined}>{t("replyToTopic")}</span>
    </Button>
  );

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
              topicReplyAction={showFirstTopicReply && post.postNo === 1 ? topicReplyButton() : undefined}
              likeCount={likeCount}
              actionBusy={actionBusy}
              composerBusy={composerBusy}
              canPost={page.permissions.canPost}
              isOwnTopic={page.permissions.isOwnTopic}
              topicCanModerate={page.permissions.canModerateTopic}
              topicStatus={topicStatus}
              t={t}
              onToggle={toggleAction}
              onReply={() => openComposer(post)}
              onEdit={() => editPost(post)}
              onHistory={() => openHistory(post)}
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
          <div ref={sentinel} className={cn(
            "px-4 py-3 text-center",
            hasAfter && "border-t",
            shareTopicEndRow && "flex h-13 items-center justify-center",
          )}>
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
              <span className={cn("text-xs text-muted-foreground", shareTopicEndRow && "max-w-[32%] truncate")}>
                {t("allRepliesShown")}
              </span>
            ) : null}
            {windowError ? (
              <p className="mt-1 text-xs text-destructive">{windowError}</p>
            ) : null}
          </div>
          <div
            className={cn(
              "pointer-events-none sticky bottom-[calc(0.5rem+env(safe-area-inset-bottom))] z-30 flex h-13 items-center gap-2 px-4",
              shareTopicEndRow && "-mt-13",
              !showDockedReply && (showMobilePosition ? "xl:mt-0 xl:h-0" : "mt-0 h-0"),
            )}
          >
            <div ref={setTopicControls} className="contents" />
            {showDockedReply ? topicReplyButton(true) : null}
          </div>
        </div>
        <TopicAside
          page={page}
          replyCount={replyCount}
          active={activePostNo}
          mobileHidden={composerOpen}
          max={maxPostNo}
          topicControls={topicControls}
          postsRoot={postsRef}
          cancelLoad={cancelAnchorLoad}
          posts={posts}
          hasAfter={hasAfter}
          locale={runtime.locale}
          t={t}
          onSelect={(value) => loadWindow("anchor", value)}
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
      <Suspense fallback={null}>
        <TopicComposer
          open={composerOpen}
          minimized={composerMinimized}
          expanded={composerExpanded}
          content={composerContent}
          sourceVersion={sourceVersion}
          editing={Boolean(editingId)}
          target={replyTarget}
          viewer={layout.viewer}
          topicTitle={page.topic.title}
          busy={composerBusy}
          submitDisabled={localDraft.recoveryPending}
          error={composerError}
          onContent={setComposerContent}
          onSubmit={() => void submitComposer()}
          onClose={() => { if (!composerSubmitting.current) setComposerOpen(false); }}
          onMinimize={() => setComposerMinimized((value) => !value)}
          onExpand={() => setComposerExpanded((value) => !value)}
          onClearTarget={() => { if (!composerSubmitting.current) setReplyTargetId(0); }}
          draftPanel={<LocalDraftPanel compact status={localDraft.status} candidates={localDraft.candidates} onRestore={restoreDraft} onDiscard={localDraft.discard} onRetry={localDraft.retry} available={availableDraft} checking={(draft) => page.permissions.canPost && draftTargets.target(draft.replyTargetId).status === "checking"} failed={(draft) => draftTargets.target(draft.replyTargetId).status === "error"} onRetryTarget={draftTargets.retry} conflicts={(draft) => draft.sourceFingerprint !== draftFingerprint(composerBase)} />}
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
        open={Boolean(historyPost)}
        onOpenChange={(open) => !open && setHistoryPost(undefined)}
      >
        <DialogContent className="flex max-h-[min(85dvh,48rem)] min-w-0 flex-col gap-0 overflow-hidden p-0 sm:max-w-2xl">
          <DialogHeader className="min-w-0 shrink-0 border-b p-4 pr-12">
            <DialogTitle>{t("historyTitle")}</DialogTitle>
            <DialogDescription className="whitespace-normal leading-relaxed [overflow-wrap:anywhere]">{historyPost ? t("historyDescription", { no: historyPost.postNo }) : null}</DialogDescription>
          </DialogHeader>
          <div className="min-h-0 min-w-0 overflow-y-auto overscroll-contain p-4">
            {historyError ? (
              <Alert variant="destructive"><AlertDescription>{t("historyLoadFailed")}</AlertDescription></Alert>
            ) : historyLoading && historyItems.length === 0 ? (
              <div className="flex justify-center py-8"><Spinner /></div>
            ) : historyItems.length === 0 ? (
              <p className="py-8 text-center text-sm text-muted-foreground">{t("historyEmpty")}</p>
            ) : (
              <div className="min-w-0 space-y-3">
                {historyItems.map((revision) => (
                  <article key={`${revision.postId}-${revision.revisionNo}`} className="min-w-0 rounded-lg border p-3">
                    <div className="mb-2 flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground">
                      <span className="shrink-0">{t("historyRevision", { no: revision.revisionNo })}</span>
                      <time className="min-w-0 [overflow-wrap:anywhere]" dateTime={revision.createdAt}>{formatDate(revision.createdAt, runtime.locale)}</time>
                    </div>
                    <pre className="m-0 max-w-full whitespace-pre-wrap font-sans text-sm leading-relaxed [overflow-wrap:anywhere]">{revision.content}</pre>
                  </article>
                ))}
                {historyHasMore ? (
                  <Button variant="outline" className="w-full" disabled={historyLoading} onClick={() => historyPost && void loadHistory(historyPost.id, historyCursor, true)}>
                    {historyLoading ? <Spinner data-icon="inline-start" /> : null}{t("historyLoadMore")}
                  </Button>
                ) : null}
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
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
                <MessageCircleExclamationPoint data-icon="inline-start" />
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
          className="flex h-dvh w-full max-w-none flex-col gap-0 rounded-none border-0 bg-transparent p-0 text-white ring-0 sm:max-w-none"
          overlayClassName="bg-black/65 supports-backdrop-filter:backdrop-blur-none"
          data-lightbox-backdrop=""
          showCloseButton={false}
          onPointerDownCapture={(event) => {
            suppressImageClick.current = false;
            dismissImageOnClick.current = false;
            imageBackdropPress.current = event.button === 0 && event.isPrimary && (event.target as HTMLElement).hasAttribute("data-lightbox-backdrop")
              ? { x: event.clientX, y: event.clientY, pointerId: event.pointerId }
              : null;
          }}
          onPointerUpCapture={(event) => {
            const press = imageBackdropPress.current;
            dismissImageOnClick.current = Boolean(press && press.pointerId === event.pointerId
              && Math.hypot(event.clientX - press.x, event.clientY - press.y) <= 6
              && (event.target as HTMLElement).hasAttribute("data-lightbox-backdrop"));
            imageBackdropPress.current = null;
          }}
          onPointerCancelCapture={() => { imageBackdropPress.current = null; dismissImageOnClick.current = false; }}
          onClickCapture={(event) => {
            if (dismissImageOnClick.current && !suppressImageClick.current) {
              event.preventDefault();
              event.stopPropagation();
              setImageIndex(-1);
            }
            dismissImageOnClick.current = false;
          }}
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
          <div data-lightbox-backdrop="" className="flex shrink-0 items-center justify-between gap-2 px-3 py-2">
            <span className="text-sm tabular-nums" aria-live="polite">{imageIndex + 1} / {images.length}</span>
            <div className="flex items-center gap-2">
              <Button variant="secondary" size="sm" className="bg-black/50 text-white hover:bg-black/65 focus-visible:ring-white/70" aria-pressed={imageActualSize} onClick={() => setImageActualSize((value) => !value)}>
                {imageActualSize ? <Minimize2 data-icon="inline-start" /> : <Maximize2 data-icon="inline-start" />}
                {t(imageActualSize ? 'fitImage' : 'originalImageSize')}
              </Button>
              <Button variant="secondary" size="icon" className="bg-black/50 text-white hover:bg-black/65 focus-visible:ring-white/70" aria-label={t('close')} onClick={() => setImageIndex(-1)}><X /></Button>
            </div>
          </div>
          <div
            ref={imageViewport}
            data-lightbox-backdrop=""
            className="min-h-0 flex-1 overflow-auto overscroll-contain p-3 sm:px-16"
            style={{ touchAction: imageActualSize ? "none" : "auto" }}
            onPointerDown={(event) => {
              if (!imageActualSize || event.button !== 0) return;
              suppressImageClick.current = false;
              const viewport = event.currentTarget;
              imageDrag.current = { x: event.clientX, y: event.clientY, left: viewport.scrollLeft, top: viewport.scrollTop, moved: false };
            }}
            onPointerMove={(event) => {
              const drag = imageDrag.current;
              if (!drag) return;
              const dx = event.clientX - drag.x;
              const dy = event.clientY - drag.y;
              if (Math.abs(dx) + Math.abs(dy) > 4) {
                drag.moved = true;
                event.currentTarget.setPointerCapture(event.pointerId);
                event.currentTarget.scrollLeft = drag.left - dx;
                event.currentTarget.scrollTop = drag.top - dy;
              }
            }}
            onPointerUp={(event) => {
              if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId);
              suppressImageClick.current = Boolean(imageDrag.current?.moved);
              imageDrag.current = null;
            }}
            onPointerCancel={() => { imageDrag.current = null; }}
            onClickCapture={(event) => {
              if (suppressImageClick.current) {
                event.preventDefault();
                event.stopPropagation();
              }
              suppressImageClick.current = false;
            }}
          >
            <div data-lightbox-backdrop="" className={imageActualSize ? "flex min-h-full w-max min-w-full" : "flex min-h-full w-full items-center justify-center"}>
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
                draggable={false}
                style={imageActualSize ? { maxWidth: 'none', width: 'auto', height: 'auto', flexShrink: 0 } : undefined}
                className={imageActualSize ? 'm-auto cursor-grab select-none active:cursor-grabbing' : 'max-h-[calc(100dvh-7rem)] max-w-full cursor-zoom-in object-contain'}
                onClick={() => setImageActualSize((value) => !value)}
              />
            </div>
          </div>
          <div data-lightbox-backdrop="" className="flex shrink-0 items-center justify-center gap-3 px-4 pb-3 text-xs text-white/80">
            <span className="min-w-0 truncate" title={images[imageIndex]?.name}>{images[imageIndex]?.name || images[imageIndex]?.alt}</span>
            <span className="shrink-0 tabular-nums">{[images[imageIndex]?.dimensions, images[imageIndex]?.size].filter(Boolean).join(" · ")}</span>
          </div>
          <Button
            variant="secondary"
            size="icon"
            className="absolute left-3 top-1/2 -translate-y-1/2 bg-black/50 text-white hover:bg-black/65 focus-visible:ring-white/70"
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
            className="absolute right-3 top-1/2 -translate-y-1/2 bg-black/50 text-white hover:bg-black/65 focus-visible:ring-white/70"
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

function PostActionGroup({
  placement,
  primaryAction,
  children,
}: {
  placement: "header" | "topic";
  primaryAction?: ReactNode;
  children: ReactNode;
}) {
  if (placement === "topic") {
    return (
      <div
        data-slot="topic-actions"
        className="col-span-full row-start-3 flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 pt-1 lg:mt-3 lg:border-t lg:pt-3"
      >
        <div className="flex min-w-0 flex-wrap items-center lg:gap-2 [&_button]:size-9 [&_button]:gap-0 [&_button]:p-0! [&_svg]:size-[18px]! [&_svg]:stroke-2 lg:[&_button]:h-7 lg:[&_button]:w-auto lg:[&_button]:gap-1 lg:[&_button]:px-2.5! lg:[&_svg]:size-3.5!">
          {children}
        </div>
        {primaryAction}
      </div>
    );
  }
  return (
    <div
      data-slot="post-actions"
      className="contents lg:flex lg:shrink-0 lg:items-center lg:gap-1.5 [&_button]:size-9 [&_button]:gap-0 [&_button]:p-0! [&_svg]:size-[18px]! [&_svg]:stroke-2 lg:[&_button]:size-6 lg:[&_svg]:size-3!"
    >
      {children}
    </div>
  );
}

function PostRow({
  post,
  target,
  first,
  locale,
  liked,
  bookmarked,
  watched,
  topicReplyAction,
  likeCount,
  actionBusy,
  canPost,
  composerBusy,
  isOwnTopic,
  topicCanModerate,
  topicStatus,
  t,
  onToggle,
  onReply,
  onEdit,
  onHistory,
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
  topicReplyAction?: ReactNode;
  likeCount: number;
  actionBusy: string;
  canPost: boolean;
  composerBusy: boolean;
  isOwnTopic: boolean;
  topicCanModerate: boolean;
  topicStatus: number;
  t: Translate;
  onToggle(type: "like" | "bookmark" | "watch"): void;
  onReply(): void;
  onEdit(): void;
  onHistory(): void;
  onDelete(): void;
  onReport(): void;
  onModerate(action: "ban" | "unban"): void;
}) {
  const hidden = post.isHidden && !post.canModerate;
  return (
    <article
      id={`post-${post.id}`}
      data-post-no={post.postNo}
      className="group relative grid scroll-mt-24 grid-cols-[42px_minmax(0,1fr)_auto] gap-x-2 gap-y-1 px-4 pt-3 pb-1 after:absolute after:inset-x-4 after:bottom-0 after:h-px after:bg-border last:after:hidden lg:grid-cols-[42px_minmax(0,1fr)] lg:gap-3 lg:py-4"
    >
      <Person user={post.author} avatarOnly showBadge sticky />
      <div className="contents lg:block lg:min-w-0">
        <header className="contents text-xs text-muted-foreground lg:mb-1.5 lg:flex lg:min-w-0 lg:items-start lg:justify-between lg:gap-2">
          <div className="col-start-2 row-start-1 min-w-0 self-center">
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
          <PostActionGroup placement="header">
            <div className="col-start-3 row-start-1 flex items-start justify-end gap-0 self-start lg:contents">
            {post.isOwnPost ? (
              <>
                <Button
                  variant="ghost"
                  size="icon-xs"
                  className="text-muted-foreground hover:bg-primary/10 hover:text-primary"
                  aria-label={t("edit")}
                  disabled={composerBusy}
                  onClick={onEdit}
                >
                  <PenLine />
                </Button>
                <Button
                  variant="ghost"
                  size="icon-xs"
                  className="text-muted-foreground hover:bg-primary/10 hover:text-primary"
                  aria-label={t("viewHistory")}
                  onClick={onHistory}
                >
                  <Clock />
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

            {!post.isOwnPost && !(first && isOwnTopic) && !post.isHidden ? (
              <Button
                variant="ghost"
                size="icon-xs"
                className={cn(
                  "text-muted-foreground hover:bg-warning/10 hover:text-warning lg:order-4",
                  first && "lg:hidden",
                )}
                aria-label={t("report")}
                onClick={onReport}
              >
                <MessageCircleExclamationPoint />
              </Button>
            ) : null}
            {!first && post.canModerate ? (
              <Button
                variant="ghost"
                size="icon-xs"
                className={cn(
                  "text-muted-foreground lg:order-5",
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
            </div>
            {!first && canPost && !post.isHidden ? (
              <Button
                variant="ghost"
                size="icon-xs"
                className="col-start-3 row-start-3 justify-self-end text-muted-foreground hover:bg-primary/10 hover:text-primary lg:order-3"
                aria-label={t("reply")}
                onClick={onReply}
                disabled={composerBusy}
              >
                <MessageCircle />
              </Button>
            ) : null}
            <time
              dateTime={post.createdAt}
              className="hidden w-36 shrink-0 text-right lg:order-6 lg:block"
            >
              {formatDate(post.createdAt, locale)}
            </time>
          </PostActionGroup>
        </header>
        <div data-slot="post-body" className="col-span-3 row-start-2 min-w-0 pt-1 lg:pt-0">
        {post.replyToPostId ? <ReplyReference topicId={post.topicId} target={target} t={t} /> : null}
        {hidden ? (
          <div className="rounded-lg bg-muted px-3 py-2 text-sm text-muted-foreground">
            {t("hiddenReplyPlaceholder")}
          </div>
        ) : (
          <><RenderedContent html={post.renderedContent} /><AttachmentList files={post.attachments || []} /></>
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
        </div>
        {first ? (
          <PostActionGroup placement="topic" primaryAction={topicReplyAction}>
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
              icon={watched ? BellRing : Bell}
              label={t(watched ? "watched" : "watch")}
              onClick={() => onToggle("watch")}
            />
            {!isOwnTopic && !post.isOwnPost ? (
              <Button
                variant="ghost"
                size="sm"
                className="hidden lg:inline-flex"
                onClick={onReport}
              >
                <MessageCircleExclamationPoint data-icon="inline-start" />
                <span className="sr-only lg:not-sr-only">{t("report")}</span>
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
                <span className="sr-only lg:not-sr-only">{t(topicStatus === 1 ? "moderationUnban" : "moderationBan")}</span>
              </Button>
            ) : null}
          </PostActionGroup>
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
      aria-pressed={tone === "watch" ? active : undefined}
      disabled={busy}
      onClick={onClick}
    >
      <Icon data-icon="inline-start" fill={active ? "currentColor" : "none"} />
      <span className="sr-only lg:not-sr-only">{label}</span>
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
  const displayName = userDisplayName(user);
  return (
    <UserCardPopover user={user}>
      <GooseLink
        href={`/u/${user.id}`}
        className={cn(
          "inline-flex min-w-0 items-center gap-2 font-medium hover:text-primary",
          avatarOnly && avatarClassName,
          sticky && "self-start lg:sticky lg:top-19 lg:mt-1",
          textOnly && "truncate",
        )}
        title={displayName}
      >
        {textOnly ? (
          displayName
        ) : (
          showBadge ? (
            <ProfileAvatar
              userId={user.id}
              src={user.avatarUrl}
              name={displayName}
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
              <AvatarFallback userId={user.id}>
                {avatarInitial(displayName)}
              </AvatarFallback>
            </Avatar>
          )
        )}
        {!avatarOnly && !textOnly ? displayName : null}
      </GooseLink>
    </UserCardPopover>
  );
}
function TopicAside({
  page,
  replyCount,
  active,
  mobileHidden,
  max,
  topicControls,
  postsRoot,
  cancelLoad,
  posts,
  hasAfter,
  locale,
  t,
  onSelect,
}: {
  page: TopicDetailProps;
  replyCount: number;
  active: number;
  mobileHidden: boolean;
  max: number;
  topicControls: HTMLDivElement | null;
  postsRoot: RefObject<HTMLDivElement | null>;
  cancelLoad(): void;
  posts: PostPayload[];
  hasAfter: boolean;
  locale: string;
  t: Translate;
  onSelect(value: number): Promise<number | undefined>;
}) {
  const endDate = hasAfter
    ? page.topic.updatedAt
    : posts.at(-1)?.createdAt || page.topic.updatedAt;
  return (
    <aside className="contents min-w-0 xl:block">
      <div className="contents sticky top-19 xl:block">
        <h2 className="hidden px-4 py-4 xl:block text-sm font-semibold text-muted-foreground">
          {t("overview")}
        </h2>
        <dl className="hidden flex-col gap-4 border-t px-4 py-5 text-sm xl:flex">
          <Stat label={t("replyCount")} value={replyCount} />
          <Stat label={t("viewCount")} value={page.topic.viewCount} />
          <Stat
            label={t("participants")}
            value={page.topic.participants.length}
          />
        </dl>
        {page.topic.participants.length ? (
          <div className="hidden border-t px-4 py-4 xl:block">
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
            mobileHidden={mobileHidden}
            max={max}
            startLabel={formatRailDate(page.topic.createdAt, locale)}
            endLabel={formatRailDate(endDate, locale)}
            topicControls={topicControls}
            posts={posts}
            postsRoot={postsRoot}
            hasAfter={hasAfter}
            cancelLoad={cancelLoad}
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
  current, max, mobileHidden, startLabel, endLabel, topicControls, posts, postsRoot, hasAfter, cancelLoad, locale, t, onSelect,
}: {
  current: number;
  mobileHidden: boolean;
  max: number;
  startLabel: string;
  endLabel: string;
  topicControls: HTMLDivElement | null;
  posts: PostPayload[];
  postsRoot: RefObject<HTMLDivElement | null>;
  hasAfter: boolean;
  cancelLoad(): void;
  locale: string;
  t: Translate;
  onSelect(value: number): Promise<number | undefined>;
}) {
  const position = usePostPosition({ root: postsRoot, posts, max, hasAfter, current, load: onSelect, cancelLoad });
  const [preview, setPreview] = useState<number>();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [jumpInput, setJumpInput] = useState("");
  const [mobileCommit, setMobileCommit] = useState<number>();
  const [desktopSliderVersion, setDesktopSliderVersion] = useState(0);
  const [mobileSliderVersion, setMobileSliderVersion] = useState(0);
  const mobileGesture = useRef<{ id: number; start: number; value: number } | null>(null);
  const jumpInputId = useId();
  const mobileTrigger = useRef<HTMLButtonElement>(null);
  const mobileClose = useRef<HTMLButtonElement>(null);
  const pendingMobileSelection = useRef<number | undefined>(undefined);
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- A cached page must not reopen a preserved mobile sheet on Back.
    setMobileOpen(false);
    setMobileCommit(undefined);
    const wideScreen = window.matchMedia?.("(min-width: 80rem)");
    const closeOnDesktop = () => { if (wideScreen?.matches) setMobileOpen(false); };
    wideScreen?.addEventListener("change", closeOnDesktop);
    return () => { pendingMobileSelection.current = undefined; wideScreen?.removeEventListener("change", closeOnDesktop); };
  }, []);
  const indicatedTarget = position.target ?? mobileCommit;
  const displayed = Math.min(max, Math.max(1, Math.round(preview ?? indicatedTarget ?? position.postNo)));
  const progress = preview !== undefined ? (preview - 1) / Math.max(1, max - 1)
    : indicatedTarget !== undefined ? (indicatedTarget - 1) / Math.max(1, max - 1) : position.progress;
  const currentPost = posts.find(post => post.postNo === displayed);
  const select = (value: number) => {
    setPreview(undefined);
    if (mobileOpen) {
      setMobileCommit(Math.round(value));
      pendingMobileSelection.current = Math.round(value);
      setMobileOpen(false);
    } else void position.select(Math.round(value));
  };

  const mobileValue = Math.min(max, Math.max(1, Math.round(Number(jumpInput) || 1)));
  const mobilePost = posts.find(post => post.postNo === mobileValue);
  const mobileProgress = ((preview ?? mobileValue) - 1) / Math.max(1, max - 1);
  const cancelMobileGesture = () => {
    const gesture = mobileGesture.current;
    mobileGesture.current = null;
    if (gesture) {
      setJumpInput(String(gesture.start));
      // Radix clears its cached drag bounds on pointerup, but not pointercancel.
      setMobileSliderVersion(version => version + 1);
    }
    setPreview(undefined);
  };

  const controls = () => (
    <div className="border-t px-3 py-2" data-testid="post-position-rail" data-progress={progress}
      data-state={preview !== undefined ? "dragging" : position.target !== undefined ? "jumping" : "reading"}>
      <PositionJumpButton placement="start" label={startLabel} title={t("earliestContent")} disabled={false} onClick={() => select(1)} />
      <div className="grid h-48 grid-cols-[24px_minmax(0,1fr)] gap-3">
        <SliderPrimitive.Root key={desktopSliderVersion} orientation="vertical" inverted min={1} max={max} step={0.001}
          value={[1 + progress * (max - 1)]}
          className="relative flex h-full w-6 touch-none select-none flex-col items-center"
          onPointerDownCapture={() => position.cancel()}
          onPointerCancel={() => {
            setPreview(undefined);
            // A cancelled drag must discard Radix's cached bounds before the sticky rail moves.
            setDesktopSliderVersion(version => version + 1);
          }}
          onLostPointerCapture={() => setPreview(undefined)}
          onKeyDown={(event) => {
            const values: Record<string, number> = { ArrowDown: displayed + 1, ArrowUp: displayed - 1, ArrowRight: displayed + 1, ArrowLeft: displayed - 1, PageDown: displayed + 10, PageUp: displayed - 10, Home: 1, End: max };
            if (values[event.key] === undefined) return;
            event.preventDefault();
            select(Math.max(1, Math.min(max, values[event.key])));
          }}
          onValueChange={values => setPreview(values[0])}
          onValueCommit={values => select(values[0])}>
          <SliderPrimitive.Track className="relative h-full w-px grow rounded-full bg-border" />
          <SliderPrimitive.Thumb aria-label={t("replyPosition")} aria-valuenow={displayed}
            aria-valuetext={`${displayed} / ${max.toLocaleString(locale)}`}
            className="relative block h-7 w-1.5 rounded-full bg-primary outline-hidden after:absolute after:-inset-2 focus-visible:ring-3 focus-visible:ring-ring/50" />
        </SliderPrimitive.Root>
        <div className="relative min-w-0">
          <div className="absolute top-0 left-0 min-w-0" style={{ transform: `translateY(calc(${progress} * (12rem - 1.75rem)))` }}>
            <div className="whitespace-nowrap text-base font-black leading-none tabular-nums">{displayed} / {max.toLocaleString(locale)}</div>
            <div className="mt-2 h-5 truncate text-sm font-semibold leading-tight text-muted-foreground" aria-busy={position.target !== undefined}>
              {currentPost ? formatRailDate(currentPost.createdAt, locale) : position.target !== undefined ? <Spinner className="size-3" /> : null}
            </div>
          </div>
        </div>
      </div>
      <PositionJumpButton placement="end" label={endLabel} title={t("latestReply")} disabled={false} onClick={() => select(max)} />
    </div>
  );
  return <>
    <div className="hidden xl:block">{controls()}</div>
    <Sheet open={mobileOpen} onOpenChange={open => {
      if (open) { position.cancel(); setMobileCommit(undefined); setJumpInput(String(position.postNo)); }
      setPreview(undefined);
      setMobileOpen(open);
    }}>
      {topicControls && createPortal(<SheetTrigger asChild>
        <Button ref={mobileTrigger} variant="outline" aria-label={t("replyPosition")} aria-describedby={`${jumpInputId}-current`}
          className={cn("pointer-events-auto relative h-9 min-w-[min(7rem,34%)] max-w-[34%] overflow-hidden rounded-full border-primary/25 bg-background px-2.5 text-primary shadow-lg sm:px-4 xl:hidden", mobileHidden && "hidden")}
          data-slot="mobile-topic-position">
          <span aria-hidden="true" className="pointer-events-none absolute inset-y-0 left-0 bg-primary/12" style={{ width: `${progress * 100}%` }} />
          <span id={`${jumpInputId}-current`} className="relative min-w-0 truncate tabular-nums">{indicatedTarget ?? position.postNo} / {max.toLocaleString(locale)}</span>
        </Button>
      </SheetTrigger>, topicControls)}
      <SheetContent side="bottom" showCloseButton={false} className="max-h-[85dvh] overflow-y-auto rounded-t-2xl pb-[max(1rem,env(safe-area-inset-bottom))]"
        onOpenAutoFocus={event => { event.preventDefault(); mobileClose.current?.focus({ preventScroll: true }); }}
        onCloseAutoFocus={event => {
          event.preventDefault();
          mobileTrigger.current?.focus({ preventScroll: true });
          const value = pendingMobileSelection.current;
          pendingMobileSelection.current = undefined;
          // Start after the modal unmounts and releases its body scroll lock.
          if (value !== undefined) void position.select(value);
          setMobileCommit(undefined);
        }}>
        <SheetHeader className="pr-14">
          <SheetTitle>{t("replyPosition")}</SheetTitle>
          <SheetDescription>{t("positionHelp")}</SheetDescription>
        </SheetHeader>
        <Button ref={mobileClose} variant="ghost" size="icon" aria-label={t("close")} className="absolute top-3 right-3" onClick={() => setMobileOpen(false)}><X /></Button>
        <div className="mx-4 border-y px-3 py-2" data-testid="post-position-mobile">
          <Button type="button" variant="ghost" aria-label={`${t("earliestContent")}: ${startLabel}`} className="mb-2 h-11 w-full justify-start px-0 text-base font-semibold text-muted-foreground hover:bg-transparent hover:text-foreground" onClick={() => select(1)}>
            {startLabel}
          </Button>
          <SliderPrimitive.Root key={mobileSliderVersion} data-slot="mobile-position-slider" orientation="vertical" inverted min={1} max={max} step={0.001}
            value={[preview ?? mobileValue]}
            className="relative flex h-[clamp(8rem,calc(85dvh-22rem),22rem)] w-full touch-none select-none flex-col items-start"
            onPointerDownCapture={event => {
              if (event.button !== 0 || !event.isPrimary) { event.preventDefault(); return; }
              mobileGesture.current = { id: event.pointerId, start: mobileValue, value: preview ?? mobileValue };
            }}
            onValueChange={values => {
              if (mobileGesture.current) mobileGesture.current.value = values[0];
              setPreview(values[0]);
              setJumpInput(String(Math.round(values[0])));
            }}
            onPointerUp={event => {
              const gesture = mobileGesture.current;
              if (!gesture || gesture.id !== event.pointerId) return;
              mobileGesture.current = null;
              select(gesture.value);
            }}
            onPointerCancel={cancelMobileGesture}
            onLostPointerCapture={() => { if (mobileGesture.current) cancelMobileGesture(); }}
            onKeyDown={event => {
              const values: Record<string, number> = { ArrowDown: mobileValue + 1, ArrowUp: mobileValue - 1, ArrowRight: mobileValue + 1, ArrowLeft: mobileValue - 1, PageDown: mobileValue + 10, PageUp: mobileValue - 10, Home: 1, End: max };
              if (values[event.key] === undefined) return;
              event.preventDefault();
              select(Math.max(1, Math.min(max, values[event.key])));
            }}>
            <SliderPrimitive.Track className="relative ml-[22px] h-full w-px grow rounded-full bg-border" />
            <SliderPrimitive.Thumb aria-label={t("replyPosition")} aria-valuenow={mobileValue} aria-valuetext={`${mobileValue} / ${max.toLocaleString(locale)}`}
              className="flex size-11 cursor-grab items-center justify-center rounded-md outline-hidden active:cursor-grabbing focus-visible:ring-3 focus-visible:ring-ring/50">
              <span aria-hidden="true" className="pointer-events-none h-7 w-1.5 rounded-full bg-primary" />
            </SliderPrimitive.Thumb>
            <div aria-hidden="true" className="pointer-events-none absolute inset-y-0 left-12 right-0">
              <div className="absolute left-0 w-full" style={{ top: `calc(0.5rem + ${mobileProgress} * (100% - 2.75rem))` }}>
                <div className="whitespace-nowrap text-base font-black leading-none tabular-nums">{mobileValue.toLocaleString(locale)} / {max.toLocaleString(locale)}</div>
                <div className="mt-2 h-5 truncate text-sm font-semibold leading-tight text-muted-foreground">{mobilePost ? formatRailDate(mobilePost.createdAt, locale) : null}</div>
              </div>
            </div>
          </SliderPrimitive.Root>
          <Button type="button" variant="ghost" aria-label={`${t("latestReply")}: ${endLabel}`} className="mt-6 h-11 w-full justify-start px-0 text-base font-semibold text-muted-foreground hover:bg-transparent hover:text-foreground" onClick={() => select(max)}>
            {endLabel}
          </Button>
        </div>
        <form className="px-4" onSubmit={event => {
          event.preventDefault();
          const value = Number(jumpInput);
          if (Number.isInteger(value) && value >= 1 && value <= max) select(value);
        }}>
          <Label htmlFor={jumpInputId}>{t("jumpToPost")}</Label>
          <div className="mt-2 flex gap-2">
            <Input id={jumpInputId} type="number" inputMode="numeric" min={1} max={max} step={1} required value={jumpInput} onChange={event => { setPreview(undefined); setJumpInput(event.target.value); }} />
            <Button type="submit">{t("jumpGo")}</Button>
          </div>
        </form>
      </SheetContent>
    </Sheet>
  </>;
}

function PositionJumpButton({
  placement,
  label,
  title,
  disabled,
  onClick,
}: {
  placement: "start" | "end";
  label: string;
  title: string;
  disabled: boolean;
  onClick(): void;
}) {
  return (
    <Button
      type="button"
      variant="ghost"
      className={cn(
        "block max-w-full truncate px-0 text-left text-base font-semibold leading-tight text-muted-foreground hover:bg-transparent hover:text-foreground",
        placement === "start" ? "mb-3" : "mt-6",
      )}
      title={title}
      aria-label={`${title}: ${label}`}
      disabled={disabled}
      onClick={onClick}
    >
      {label}
    </Button>
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
