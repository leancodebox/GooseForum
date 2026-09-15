import {
  Fragment,
  useEffect,
  useMemo,
  useRef,
  useState,
  type FormEvent,
  type KeyboardEvent,
} from "react";
import type {
  ChatItemPayload,
  ChatMessagePayload,
  LayoutPayload,
  MessagesPageProps,
  UserConnectionPayload,
} from "@gooseforum/client";
import {
  ArrowLeft,
  MessageSquare,
  MessageSquarePlus,
  MoreVertical,
  Search,
  Send,
  Smile,
  X,
} from "lucide-react";
import { useTranslation } from "react-i18next";
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
  DialogHeader,
  DialogTitle,
} from "@gooseforum/ui/components/dialog";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
  InputGroupTextarea,
} from "@gooseforum/ui/components/input-group";
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemMedia,
  ItemSeparator,
  ItemTitle,
} from "@gooseforum/ui/components/item";
import {
  Message,
  MessageAvatar,
  MessageContent,
  MessageFooter,
  MessageGroup,
} from "@gooseforum/ui/components/message";
import {
  MessageScroller,
  MessageScrollerButton,
  MessageScrollerContent,
  MessageScrollerItem,
  MessageScrollerProvider,
  MessageScrollerViewport,
} from "@gooseforum/ui/components/message-scroller";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@gooseforum/ui/components/popover";
import { ScrollArea } from "@gooseforum/ui/components/scroll-area";
import { cn } from "@gooseforum/ui/lib/utils";
import { announceUnreadStatus } from "@gooseforum/runtime/unread-status";
import { GooseLink, useGooseRuntime } from "@gooseforum/runtime";

type Conversation = ChatItemPayload & {
  messages: ChatMessagePayload[];
  loading: boolean;
  loadingOlder: boolean;
  messagesLoaded: boolean;
  hasMoreBefore: boolean;
  nextBeforeId: number;
  latestId: number;
};

const messagePageLimit = 30;
const emojis = [
  "😀",
  "😂",
  "😍",
  "😊",
  "😭",
  "👍",
  "🙏",
  "🔥",
  "✨",
  "🎉",
  "🤔",
  "👀",
  "❤️",
  "🙌",
  "👏",
  "✅",
];

export function MessagesPageView({
  layout,
  page,
}: {
  layout: LayoutPayload;
  page: MessagesPageProps;
}) {
  const { t } = useTranslation("messages");
  const runtime = useGooseRuntime();
  const [conversations, setConversations] = useState<Conversation[]>(() =>
    page.conversations.map(toConversation),
  );
  const [activePeerId, setActivePeerId] = useState(0);
  const [conversationSearch, setConversationSearch] = useState("");
  const [userSearch, setUserSearch] = useState("");
  const [showNewChat, setShowNewChat] = useState(false);
  const [showEmoji, setShowEmoji] = useState(false);
  const [newMessage, setNewMessage] = useState("");
  const [sending, setSending] = useState(false);
  const [error, setError] = useState("");
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const loadingConversations = useRef(new Set<number>());
  const loadingOlderConversations = useRef(new Set<number>());
  const active = conversations.find((item) => item.peerId === activePeerId);

  useEffect(() => {
    const url = new URL(runtime.currentUrl, "http://gooseforum.local");
    const targetUserId = Number(url.searchParams.get("userId") || 0);
    if (targetUserId) {
      const existing = page.conversations.find(
        (item) => item.peerId === targetUserId,
      );
      if (existing) setActivePeerId(existing.peerId);
      else {
        const username = url.searchParams.get("username") || t("userFallback");
        const avatarUrl =
          url.searchParams.get("avatar") || "/static/pic/default-avatar.webp";
        setConversations((items) => [
          newConversation({
            id: targetUserId,
            username,
            nickname: username,
            avatarUrl,
            bio: "",
            url: `/u/${targetUserId}`,
          }),
          ...items,
        ]);
        setActivePeerId(targetUserId);
      }
      return;
    }
    if (
      window.matchMedia?.("(min-width: 768px)").matches &&
      page.conversations[0]
    )
      setActivePeerId(page.conversations[0].peerId);
  }, [page.conversations, runtime.currentUrl]);

  useEffect(() => {
    if (!activePeerId) return;
    const conversation = conversations.find(
      (item) => item.peerId === activePeerId,
    );
    if (!conversation) return;
    if (
      conversation.convId &&
      !conversation.messagesLoaded &&
      !conversation.loading &&
      !loadingConversations.current.has(conversation.convId)
    ) {
      loadingConversations.current.add(conversation.convId);
      updateConversation(conversation.peerId, (item) => ({
        ...item,
        loading: true,
      }));
      void runtime.api.chat
        .messages({ convId: conversation.convId, limit: messagePageLimit })
        .then((result) =>
          updateConversation(conversation.peerId, (item) => ({
            ...item,
            messages: result.list,
            hasMoreBefore: result.hasMoreBefore,
            nextBeforeId: result.nextBeforeId,
            latestId: result.latestId,
            messagesLoaded: true,
            loading: false,
          })),
        )
        .catch((reason) => {
          setError(reason instanceof Error ? reason.message : t("loadFailed"));
          updateConversation(conversation.peerId, (item) => ({
            ...item,
            loading: false,
          }));
        })
        .finally(() =>
          loadingConversations.current.delete(conversation.convId),
        );
    }
    if (conversation.unreadCount && conversation.convId) {
      updateConversation(conversation.peerId, (item) => ({
        ...item,
        unreadCount: 0,
      }));
      const hasOtherUnread = conversations.some(
        (item) => item.peerId !== conversation.peerId && item.unreadCount > 0,
      );
      announceUnreadStatus({ messages: hasOtherUnread });
      void runtime.api.chat
        .markRead(conversation.convId)
        .catch(() => undefined);
    }
    // Only a new conversation or unread change should trigger this load;
    // loading/error updates must not automatically retry a failed request.
  }, [activePeerId, active?.convId, active?.unreadCount, runtime.api.chat, t]);

  const filteredConversations = useMemo(() => {
    const keyword = conversationSearch.trim().toLowerCase();
    return keyword
      ? conversations.filter(
          (item) =>
            item.peerUsername.toLowerCase().includes(keyword) ||
            item.lastMsg.toLowerCase().includes(keyword),
        )
      : conversations;
  }, [conversationSearch, conversations]);
  const filteredUsers = useMemo(() => {
    const keyword = userSearch.trim().toLowerCase();
    return keyword
      ? page.suggestedUsers.filter(
          (user) =>
            user.username.toLowerCase().includes(keyword) ||
            user.nickname.toLowerCase().includes(keyword),
        )
      : page.suggestedUsers;
  }, [page.suggestedUsers, userSearch]);

  function updateConversation(
    peerId: number,
    update: (item: Conversation) => Conversation,
  ) {
    setConversations((items) =>
      items.map((item) => (item.peerId === peerId ? update(item) : item)),
    );
  }

  async function loadOlder() {
    if (
      !active?.convId ||
      active.loadingOlder ||
      !active.hasMoreBefore ||
      loadingOlderConversations.current.has(active.convId)
    )
      return;
    const beforeId = active.nextBeforeId || active.messages[0]?.id || 0;
    if (!beforeId) return;
    loadingOlderConversations.current.add(active.convId);
    updateConversation(active.peerId, (item) => ({
      ...item,
      loadingOlder: true,
    }));
    try {
      const result = await runtime.api.chat.messages({
        convId: active.convId,
        beforeId,
        limit: messagePageLimit,
      });
      updateConversation(active.peerId, (item) => ({
        ...item,
        messages: prependUnique(item.messages, result.list),
        hasMoreBefore: result.hasMoreBefore,
        nextBeforeId: result.nextBeforeId || item.nextBeforeId,
        latestId: Math.max(item.latestId, result.latestId || 0),
        loadingOlder: false,
      }));
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("loadFailed"));
      updateConversation(active.peerId, (item) => ({
        ...item,
        loadingOlder: false,
      }));
    } finally {
      loadingOlderConversations.current.delete(active.convId);
    }
  }

  async function sendMessage() {
    const content = newMessage.trim();
    if (!content || !active || sending) return;
    setSending(true);
    setError("");
    try {
      const result = await runtime.api.chat.send(active.peerId, content);
      const message: ChatMessagePayload = {
        id: Date.now(),
        senderId: layout.viewer.id,
        content,
        msgType: 1,
        isRead: 0,
        createdAt: new Date().toISOString(),
        isSelf: true,
      };
      setConversations((items) => {
        const updated = items.find((item) => item.peerId === active.peerId);
        if (!updated) return items;
        const next = {
          ...updated,
          id: updated.id || result.convId,
          convId: updated.convId || result.convId,
          messages: [...updated.messages, message],
          lastMsg: content,
          lastMsgTime: message.createdAt,
          messagesLoaded: true,
        };
        return [next, ...items.filter((item) => item.peerId !== active.peerId)];
      });
      setNewMessage("");
      setShowEmoji(false);
      if (inputRef.current) inputRef.current.style.height = "auto";
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("sendFailed"));
    } finally {
      setSending(false);
    }
  }

  function startChat(user: UserConnectionPayload) {
    const existing = conversations.find((item) => item.peerId === user.id);
    if (!existing)
      setConversations((items) => [newConversation(user), ...items]);
    setActivePeerId(user.id);
    setShowNewChat(false);
  }

  return (
    <main className="h-[calc(100dvh-4rem)] min-h-0 min-w-0 overflow-hidden lg:h-[calc(100dvh-5.5rem)] lg:min-h-[620px] lg:pb-3">
      <section className="grid h-full overflow-hidden bg-background lg:grid-cols-[300px_minmax(0,1fr)] lg:rounded-xl lg:border">
        <aside
          className={cn(
            "min-h-0 flex-col border-r bg-muted/30",
            active ? "hidden lg:flex" : "flex",
          )}
        >
          <header className="flex h-14 shrink-0 items-center justify-between border-b bg-background px-4 lg:h-15">
            <h1 className="text-base font-bold lg:text-lg">{t("title")}</h1>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              title={t("newMessage")}
              aria-label={t("newMessage")}
              onClick={() => setShowNewChat(true)}
            >
              <MessageSquarePlus />
            </Button>
          </header>
          <div className="border-b p-3">
            <InputGroup className="bg-background">
              <InputGroupAddon>
                <Search />
              </InputGroupAddon>
              <InputGroupInput
                value={conversationSearch}
                onChange={(event) => setConversationSearch(event.target.value)}
                placeholder={t("searchConversations")}
              />
            </InputGroup>
          </div>
          {filteredConversations.length ? (
            <ScrollArea className="min-h-0 flex-1">
              <ItemGroup className="gap-0 has-data-[size=sm]:gap-0">
                {filteredConversations.map((conversation, index) => (
                  <Fragment key={conversation.peerId}>
                    <Item
                      asChild
                      size="sm"
                      className={cn(
                        "rounded-none border-0 px-4 py-3",
                        active?.peerId === conversation.peerId &&
                          "bg-primary/5 shadow-[inset_3px_0_0_var(--primary)]",
                      )}
                    >
                      <button
                        type="button"
                        onClick={() => {
                          setError("");
                          setActivePeerId(conversation.peerId);
                        }}
                      >
                        <ItemMedia className="relative">
                          <Avatar className="size-10">
                            <AvatarImage
                              src={conversation.peerAvatar}
                              alt={conversation.peerUsername}
                            />
                            <AvatarFallback>
                              {conversation.peerUsername
                                .slice(0, 1)
                                .toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          {conversation.unreadCount ? (
                            <span className="absolute -right-0.5 -top-0.5 size-2.5 rounded-full bg-destructive ring-2 ring-background" />
                          ) : null}
                        </ItemMedia>
                        <ItemContent className="min-w-0">
                          <ItemTitle className="w-full justify-between">
                            <span className="truncate font-semibold">
                              {conversation.peerUsername}
                            </span>
                            <time className="shrink-0 text-[11px] font-normal text-muted-foreground">
                              {formatChatTime(
                                conversation.lastMsgTime,
                                runtime.locale,
                              )}
                            </time>
                          </ItemTitle>
                          <ItemDescription
                            className={cn(
                              "line-clamp-1",
                              conversation.unreadCount &&
                                "font-semibold text-foreground",
                            )}
                          >
                            {conversation.lastMsg || t("noMessagesYet")}
                          </ItemDescription>
                        </ItemContent>
                      </button>
                    </Item>
                    {index < filteredConversations.length - 1 ? (
                      <ItemSeparator className="my-0" />
                    ) : null}
                  </Fragment>
                ))}
              </ItemGroup>
            </ScrollArea>
          ) : (
            <Empty className="min-h-0 flex-1 border-0">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <MessageSquare />
                </EmptyMedia>
                <EmptyTitle>{t("emptyConversationsTitle")}</EmptyTitle>
                <EmptyDescription>
                  {t("emptyConversationsDescription")}
                </EmptyDescription>
              </EmptyHeader>
              <Button onClick={() => setShowNewChat(true)}>
                {t("newMessage")}
              </Button>
            </Empty>
          )}
        </aside>

        <section
          className={cn(
            "relative min-h-0 min-w-0 bg-background",
            active ? "flex" : "hidden lg:flex",
          )}
        >
          {active ? (
            <div className="flex min-h-0 w-full flex-col">
              <header className="flex h-14 shrink-0 items-center justify-between border-b px-3 lg:h-15 lg:px-4">
                <div className="flex min-w-0 items-center gap-3">
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-sm"
                    className="-ml-2 lg:hidden"
                    aria-label={t("back")}
                    onClick={() => setActivePeerId(0)}
                  >
                    <ArrowLeft />
                  </Button>
                  <Avatar className="size-9">
                    <AvatarImage
                      src={active.peerAvatar}
                      alt={active.peerUsername}
                    />
                    <AvatarFallback>
                      {active.peerUsername.slice(0, 1).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div className="min-w-0">
                    <GooseLink
                      href={active.peerUrl}
                      className="block truncate text-sm font-bold hover:text-primary"
                    >
                      {active.peerUsername}
                    </GooseLink>
                    <p className="text-xs text-muted-foreground">
                      {t("conversation")}
                    </p>
                  </div>
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  aria-label={t("more")}
                >
                  <MoreVertical />
                </Button>
              </header>

              <MessageScrollerProvider autoScroll defaultScrollPosition="end">
                <MessageScroller className="min-h-0 flex-1">
                  <MessageScrollerViewport
                    preserveScrollOnPrepend
                    aria-label={t("conversation")}
                    onScroll={(event) => {
                      if (event.currentTarget.scrollTop <= 48) void loadOlder();
                    }}
                  >
                    <MessageScrollerContent className="gap-3 px-3 py-4 lg:gap-4 lg:px-4">
                      <div className="flex justify-center">
                        <Badge variant="secondary">{t("today")}</Badge>
                      </div>
                      {active.loading ? (
                        <p className="py-12 text-center text-sm text-muted-foreground">
                          {t("loading")}
                        </p>
                      ) : active.messages.length ? (
                        <MessageGroup className="gap-3">
                          {active.loadingOlder ? (
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              disabled
                            >
                              {t("loading")}
                            </Button>
                          ) : active.hasMoreBefore ? (
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              onClick={() => void loadOlder()}
                            >
                              {t("loadOlder")}
                            </Button>
                          ) : null}
                          {active.messages.map((message) => (
                            <MessageScrollerItem
                              key={message.id}
                              messageId={String(message.id)}
                            >
                              <Message
                                align={message.isSelf ? "end" : "start"}
                                className={cn(
                                  "max-w-[88%] lg:max-w-[82%]",
                                  message.isSelf ? "ml-auto" : "mr-auto",
                                )}
                              >
                                {!message.isSelf ? (
                                  <MessageAvatar>
                                    <Avatar className="size-8">
                                      <AvatarImage
                                        src={active.peerAvatar}
                                        alt={active.peerUsername}
                                      />
                                      <AvatarFallback>
                                        {active.peerUsername.slice(0, 1)}
                                      </AvatarFallback>
                                    </Avatar>
                                  </MessageAvatar>
                                ) : null}
                                <MessageContent className="w-auto max-w-full gap-1">
                                  <div
                                    className={cn(
                                      "whitespace-pre-wrap rounded-xl px-3 py-2 text-sm leading-relaxed shadow-sm lg:px-4",
                                      message.isSelf
                                        ? "bg-primary text-primary-foreground"
                                        : "bg-muted text-foreground",
                                    )}
                                  >
                                    {message.content}
                                  </div>
                                  <MessageFooter>
                                    {formatChatTime(
                                      message.createdAt,
                                      runtime.locale,
                                    )}
                                  </MessageFooter>
                                </MessageContent>
                              </Message>
                            </MessageScrollerItem>
                          ))}
                        </MessageGroup>
                      ) : (
                        <Empty className="min-h-64 flex-1 border-0">
                          <EmptyHeader>
                            <EmptyMedia variant="icon">
                              <MessageSquare />
                            </EmptyMedia>
                            <EmptyTitle>{t("startChat")}</EmptyTitle>
                            <EmptyDescription>
                              {t("firstMessageTo", {
                                user: active.peerUsername,
                              })}
                            </EmptyDescription>
                          </EmptyHeader>
                        </Empty>
                      )}
                    </MessageScrollerContent>
                  </MessageScrollerViewport>
                  <MessageScrollerButton
                    direction="end"
                    aria-label={t("scrollToLatest")}
                  />
                </MessageScroller>
              </MessageScrollerProvider>

              <footer className="shrink-0 border-t bg-background/95 px-3 py-2 lg:px-4 lg:py-3">
                <form
                  className="mx-auto max-w-4xl"
                  onSubmit={(event: FormEvent) => {
                    event.preventDefault();
                    void sendMessage();
                  }}
                >
                  {error ? (
                    <p className="mb-2 text-sm text-destructive">{error}</p>
                  ) : null}
                  <InputGroup className="h-auto flex-col items-stretch bg-muted/50 p-2 focus-within:bg-background">
                    <InputGroupTextarea
                      ref={inputRef}
                      rows={1}
                      value={newMessage}
                      placeholder={t("inputPlaceholder")}
                      className="max-h-36 min-h-11 text-[15px]"
                      onChange={(event) => setNewMessage(event.target.value)}
                      onInput={(event) => {
                        event.currentTarget.style.height = "auto";
                        event.currentTarget.style.height = `${Math.min(event.currentTarget.scrollHeight, 132)}px`;
                      }}
                      onKeyDown={(
                        event: KeyboardEvent<HTMLTextAreaElement>,
                      ) => {
                        if (event.key === "Enter" && !event.shiftKey) {
                          event.preventDefault();
                          void sendMessage();
                        }
                      }}
                    />
                    <InputGroupAddon
                      align="block-end"
                      className="justify-between border-t"
                    >
                      <div className="flex min-w-0 items-center gap-2">
                        <Popover open={showEmoji} onOpenChange={setShowEmoji}>
                          <PopoverTrigger asChild>
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon-sm"
                              aria-label={t("emoji")}
                            >
                              <Smile />
                            </Button>
                          </PopoverTrigger>
                          <PopoverContent
                            side="top"
                            align="start"
                            className="grid w-48 grid-cols-4 gap-1"
                          >
                            {emojis.map((emoji) => (
                              <Button
                                key={emoji}
                                type="button"
                                variant="ghost"
                                className="text-xl"
                                onClick={() => {
                                  setNewMessage((value) => value + emoji);
                                  setShowEmoji(false);
                                  inputRef.current?.focus();
                                }}
                              >
                                {emoji}
                              </Button>
                            ))}
                          </PopoverContent>
                        </Popover>
                        <span className="hidden truncate text-[11px] text-muted-foreground lg:inline">
                          {t("enterHint")}
                        </span>
                      </div>
                      <Button
                        type="submit"
                        size="sm"
                        disabled={!newMessage.trim() || sending}
                      >
                        <Send data-icon="inline-start" />
                        {sending ? t("sending") : t("send")}
                      </Button>
                    </InputGroupAddon>
                  </InputGroup>
                </form>
              </footer>
            </div>
          ) : (
            <Empty className="hidden min-h-0 flex-1 border-0 lg:flex">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <MessageSquare />
                </EmptyMedia>
                <EmptyTitle>{t("selectConversation")}</EmptyTitle>
                <EmptyDescription>
                  {t("selectConversationDescription")}
                </EmptyDescription>
              </EmptyHeader>
              <Button onClick={() => setShowNewChat(true)}>
                {t("newMessage")}
              </Button>
            </Empty>
          )}
        </section>
      </section>

      <Dialog open={showNewChat} onOpenChange={setShowNewChat}>
        <DialogContent
          showCloseButton={false}
          className="flex max-h-[80vh] flex-col gap-0 overflow-hidden p-0 sm:max-w-md"
        >
          <DialogHeader className="flex h-13 flex-row items-center justify-between border-b px-4 text-left">
            <DialogTitle className="text-sm">{t("newMessage")}</DialogTitle>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label={t("close")}
              onClick={() => setShowNewChat(false)}
            >
              <X />
            </Button>
          </DialogHeader>
          <DialogDescription className="sr-only">
            {t("searchUsers")}
          </DialogDescription>
          <div className="border-b p-3">
            <InputGroup className="bg-muted">
              <InputGroupAddon>
                <Search />
              </InputGroupAddon>
              <InputGroupInput
                value={userSearch}
                onChange={(event) => setUserSearch(event.target.value)}
                placeholder={t("searchUsers")}
              />
            </InputGroup>
          </div>
          <ScrollArea className="min-h-0 flex-1">
            <ItemGroup className="gap-0 p-2">
              {filteredUsers.map((user, index) => (
                <Fragment key={user.id}>
                  <Item asChild className="border-0">
                    <button type="button" onClick={() => startChat(user)}>
                      <ItemMedia>
                        <Avatar className="size-10">
                          <AvatarImage src={user.avatarUrl} alt={user.username} />
                          <AvatarFallback>
                            {user.username.slice(0, 1).toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                      </ItemMedia>
                      <ItemContent>
                        <ItemTitle>{user.nickname || user.username}</ItemTitle>
                        <ItemDescription>@{user.username}</ItemDescription>
                      </ItemContent>
                    </button>
                  </Item>
                  {index < filteredUsers.length - 1 ? (
                    <ItemSeparator className="my-0" />
                  ) : null}
                </Fragment>
              ))}
            </ItemGroup>
            {!filteredUsers.length ? (
              <p className="px-4 py-8 text-center text-sm text-muted-foreground">
                {t("noContactableUsers")}
              </p>
            ) : null}
          </ScrollArea>
        </DialogContent>
      </Dialog>
    </main>
  );
}

function toConversation(item: ChatItemPayload): Conversation {
  return {
    ...item,
    messages: [],
    loading: false,
    loadingOlder: false,
    messagesLoaded: false,
    hasMoreBefore: false,
    nextBeforeId: 0,
    latestId: 0,
  };
}
function newConversation(user: UserConnectionPayload): Conversation {
  return {
    id: 0,
    peerId: user.id,
    peerUsername: user.nickname || user.username,
    peerAvatar: user.avatarUrl,
    lastMsg: "",
    lastMsgTime: "",
    unreadCount: 0,
    convId: 0,
    peerUrl: user.url,
    messages: [],
    loading: false,
    loadingOlder: false,
    messagesLoaded: true,
    hasMoreBefore: false,
    nextBeforeId: 0,
    latestId: 0,
  };
}
function prependUnique(
  current: ChatMessagePayload[],
  incoming: ChatMessagePayload[],
) {
  const ids = new Set(current.map((message) => message.id));
  return [...incoming.filter((message) => !ids.has(message.id)), ...current];
}
function formatChatTime(value: string, locale: string) {
  if (!value) return "";
  const date = new Date(value.includes("T") ? value : value.replace(" ", "T"));
  if (Number.isNaN(date.getTime())) return value;
  const today = new Date();
  return date.toDateString() === today.toDateString()
    ? date.toLocaleTimeString(locale, { hour: "2-digit", minute: "2-digit" })
    : date.toLocaleDateString(locale, { month: "short", day: "numeric" });
}
