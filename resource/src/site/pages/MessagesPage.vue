<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { ArrowLeft, MessageSquare, MoreVertical, PenLine, Search, Send, Smile, X } from '@lucide/vue'
import { getChatMessages, markChatRead, sendChatMessage, type ChatMessagePayload } from '@/runtime/api'
import { formatChatTime } from '@/runtime/format'
import { useUnreadStatus } from '@/runtime/unread-status'
import UserAvatar from '@/site/components/UserAvatar.vue'
import type { LayoutPayload, MessageConversationPayload, MessagesPageProps, UserConnectionPayload } from '@/types/payload'
import { useI18n } from 'vue-i18n'

type ChatConversation = MessageConversationPayload & {
  messages: ChatMessagePayload[]
  loading?: boolean
  loadingOlder?: boolean
  messagesLoaded?: boolean
  hasMoreBefore?: boolean
  nextBeforeId?: number
  latestId?: number
}

const page = defineProps<{
  layout: LayoutPayload
  props: MessagesPageProps
}>()

const { t } = useI18n()
const conversations = ref<ChatConversation[]>(page.props.conversations.map((item) => ({ ...item, messages: [] })))
const active = ref<ChatConversation | null>(null)
const newMessage = ref('')
const search = ref('')
const userSearch = ref('')
const showNewChat = ref(false)
const showEmoji = ref(false)
const sending = ref(false)
const error = ref('')
const messagesEl = ref<HTMLElement | null>(null)
const messageInput = ref<HTMLTextAreaElement | null>(null)
const unreadStatus = useUnreadStatus()
const messagePageLimit = 30
const emojis = ['😀', '😂', '😍', '😊', '😭', '👍', '🙏', '🔥', '✨', '🎉', '🤔', '👀', '❤️', '🙌', '👏', '✅']

const filteredConversations = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  if (!keyword) return conversations.value
  return conversations.value.filter((item) =>
    item.peerUsername.toLowerCase().includes(keyword) || item.lastMsg.toLowerCase().includes(keyword),
  )
})

const filteredUsers = computed(() => {
  const keyword = userSearch.value.trim().toLowerCase()
  if (!keyword) return page.props.suggestedUsers
  return page.props.suggestedUsers.filter((user) =>
    user.username.toLowerCase().includes(keyword) || user.nickname.toLowerCase().includes(keyword),
  )
})

onMounted(() => {
  const params = new URLSearchParams(window.location.search)
  const targetUserId = Number(params.get('userId') || 0)
  if (targetUserId) {
    const existing = conversations.value.find((item) => item.peerId === targetUserId)
    if (existing) {
      void selectConversation(existing)
      return
    }
    const username = params.get('username') || t('notifications.actorFallback')
    const avatar = params.get('avatar') || '/static/pic/default-avatar.webp'
    void startChat({ id: targetUserId, username, nickname: username, avatarUrl: avatar, bio: '', url: `/u/${targetUserId}` })
    return
  }

  if (window.matchMedia('(min-width: 768px)').matches && conversations.value.length > 0) {
    void selectConversation(conversations.value[0])
  }
})

async function selectConversation(conversation: ChatConversation) {
  active.value = conversation
  error.value = ''
  showNewChat.value = false
  if (conversation.convId && !conversation.messagesLoaded) {
    await loadInitialMessages(conversation)
  }
  if (conversation.unreadCount > 0 && conversation.convId) {
    conversation.unreadCount = 0
    if (!conversations.value.some((item) => item.peerId !== conversation.peerId && item.unreadCount > 0)) {
      unreadStatus.clearMessages()
    }
    void markChatRead(conversation.convId).catch(() => undefined)
  }
  await scrollToBottom()
}

async function loadInitialMessages(conversation: ChatConversation) {
  if (!conversation.convId || conversation.loading) return
  conversation.loading = true
  try {
    const result = await getChatMessages({ convId: conversation.convId, limit: messagePageLimit })
    conversation.messages = result.list
    conversation.hasMoreBefore = result.hasMoreBefore
    conversation.nextBeforeId = result.nextBeforeId
    conversation.latestId = result.latestId
    conversation.messagesLoaded = true
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('api.messagesLoadFailed')
  } finally {
    conversation.loading = false
  }
}

async function loadOlderMessages() {
  const conversation = active.value
  const el = messagesEl.value
  if (!conversation?.convId || !el || conversation.loading || conversation.loadingOlder || !conversation.hasMoreBefore) return
  const beforeId = conversation.nextBeforeId || conversation.messages[0]?.id || 0
  if (!beforeId) return

  conversation.loadingOlder = true
  const oldScrollHeight = el.scrollHeight
  try {
    const result = await getChatMessages({ convId: conversation.convId, beforeId, limit: messagePageLimit })
    const existingIds = new Set(conversation.messages.map((message) => message.id))
    const olderMessages = result.list.filter((message) => !existingIds.has(message.id))
    conversation.messages = [...olderMessages, ...conversation.messages]
    conversation.hasMoreBefore = result.hasMoreBefore
    conversation.nextBeforeId = result.nextBeforeId || conversation.nextBeforeId
    if (result.latestId) conversation.latestId = Math.max(conversation.latestId || 0, result.latestId)
    await nextTick()
    el.scrollTop = el.scrollHeight - oldScrollHeight
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('api.messagesLoadFailed')
  } finally {
    conversation.loadingOlder = false
  }
}

function handleMessagesScroll() {
  if (!messagesEl.value || messagesEl.value.scrollTop > 48) return
  void loadOlderMessages()
}

async function scrollToBottom() {
  await nextTick()
  if (messagesEl.value) {
    messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  }
}

async function submitMessage() {
  const content = newMessage.value.trim()
  if (!content || !active.value || sending.value) return

  sending.value = true
  error.value = ''
  try {
    const convId = await sendChatMessage(active.value.peerId, content)
    if (!active.value.convId && convId) {
      active.value.convId = convId
      active.value.id = convId
    }
    const message: ChatMessagePayload = {
      id: Date.now(),
      senderId: page.layout.viewer.id,
      content,
      msgType: 1,
      isRead: 0,
      createdAt: new Date().toISOString(),
      isSelf: true,
    }
    active.value.messages.push(message)
    active.value.lastMsg = content
    active.value.lastMsgTime = message.createdAt
    active.value.messagesLoaded = true
    conversations.value = [active.value, ...conversations.value.filter((item) => item.peerId !== active.value?.peerId)]
    newMessage.value = ''
    resizeMessageInput()
    showEmoji.value = false
    await scrollToBottom()
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('api.sendFailed')
  } finally {
    sending.value = false
  }
}

function handleEnter(event: KeyboardEvent) {
  if (event.shiftKey) return
  event.preventDefault()
  void submitMessage()
}

function resizeMessageInput() {
  void nextTick(() => {
    const input = messageInput.value
    if (!input) return
    input.style.height = 'auto'
    input.style.height = `${Math.min(input.scrollHeight, 132)}px`
  })
}

function appendEmoji(emoji: string) {
  newMessage.value += emoji
  showEmoji.value = false
  resizeMessageInput()
  void nextTick(() => messageInput.value?.focus())
}

async function startChat(user: UserConnectionPayload) {
  const existing = conversations.value.find((item) => item.peerId === user.id)
  if (existing) {
    await selectConversation(existing)
    return
  }
  const conversation: ChatConversation = {
    id: 0,
    peerId: user.id,
    peerUsername: user.nickname || user.username,
    peerAvatar: user.avatarUrl,
    lastMsg: '',
    lastMsgTime: '',
    unreadCount: 0,
    convId: 0,
    peerUrl: user.url,
    messages: [],
    messagesLoaded: true,
    hasMoreBefore: false,
    nextBeforeId: 0,
    latestId: 0,
  }
  conversations.value = [conversation, ...conversations.value]
  await selectConversation(conversation)
}
</script>

<template>
    <main class="h-[calc(100dvh-4rem)] min-h-0 min-w-0 overflow-hidden sm:-mx-5 sm:-my-3 md:mx-0 md:my-0 md:h-[calc(100dvh-5.5rem)] md:min-h-[620px] md:pb-3">
      <section class="grid h-full overflow-hidden bg-base-100 md:grid-cols-[300px_minmax(0,1fr)] md:[border-radius:var(--gf-radius-box)] md:border md:border-line/70 md:shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
        <aside
          class="flex min-h-0 flex-col border-line bg-base-100 md:border-r md:bg-base-200/50"
          :class="active ? 'hidden md:flex' : 'flex'"
        >
          <div class="flex h-14 shrink-0 items-center justify-between border-b border-line bg-base-100 px-4 md:h-15">
            <h1 class="text-base font-bold text-base-content md:text-lg">{{ t('messages.title') }}</h1>
            <button
              type="button"
              class="gf-icon-button h-8 w-8 hover:bg-base-300 hover:text-base-content"
              :title="t('messages.newMessage')"
              @click="showNewChat = true"
            >
              <PenLine class="h-4 w-4" />
            </button>
          </div>

          <div class="border-b border-line p-3 md:bg-transparent">
            <label class="flex h-9 items-center gap-2 border border-line bg-base-200 px-3 text-sm text-base-content/55 [border-radius:var(--gf-radius-field)] md:bg-base-100">
              <Search class="h-4 w-4" />
              <input v-model="search" class="min-w-0 flex-1 bg-transparent outline-none" :placeholder="t('messages.searchConversations')" />
            </label>
          </div>

          <div v-if="filteredConversations.length" class="min-h-0 flex-1 overflow-y-auto">
            <button
              v-for="conversation in filteredConversations"
              :key="conversation.peerId"
              type="button"
              class="flex w-full gap-3 border-b border-line px-4 py-3 text-left transition hover:bg-base-200 md:hover:bg-base-100"
              :class="active?.peerId === conversation.peerId ? 'bg-info/10 shadow-[inset_3px_0_0_var(--gf-color-primary)] md:bg-base-100' : ''"
              @click="selectConversation(conversation)"
            >
              <span class="relative h-10 w-10 shrink-0">
                <UserAvatar :src="conversation.peerAvatar" :alt="conversation.peerUsername" class="h-10 w-10 rounded-full object-cover ring-1 ring-line" />
                <span
                  v-if="conversation.unreadCount"
                  class="absolute -right-0.5 -top-0.5 h-2.5 w-2.5 rounded-full bg-error/100 ring-2 ring-base-100"
                  aria-hidden="true"
                />
              </span>
              <div class="min-w-0 flex-1">
                <div class="flex items-baseline justify-between gap-2">
                  <span class="truncate text-sm font-semibold text-base-content">{{ conversation.peerUsername }}</span>
                  <time class="shrink-0 text-[11px] text-base-content/55">{{ conversation.lastMsgTime ? formatChatTime(conversation.lastMsgTime) : '' }}</time>
                </div>
                <div class="mt-1 flex items-center gap-2">
                  <p class="min-w-0 flex-1 truncate text-sm" :class="conversation.unreadCount ? 'font-semibold text-base-content' : 'text-base-content/55'">
                    {{ conversation.lastMsg || t('messages.noMessagesYet') }}
                  </p>
                </div>
              </div>
            </button>
          </div>

          <div v-else class="flex min-h-0 flex-1 flex-col items-center justify-center px-6 text-center">
            <MessageSquare class="h-10 w-10 text-base-content/35" />
            <h2 class="mt-3 text-base font-semibold text-base-content">{{ t('messages.emptyConversationsTitle') }}</h2>
            <p class="mt-1 text-sm text-base-content/55">{{ t('messages.emptyConversationsDescription') }}</p>
            <button type="button" class="gf-button gf-button-md gf-button-primary mt-4" @click="showNewChat = true">
              {{ t('messages.newMessage') }}
            </button>
          </div>
        </aside>

        <section class="relative min-h-0 min-w-0 bg-base-100" :class="active ? 'flex' : 'hidden md:flex'">
          <div v-if="active" class="flex min-h-0 w-full flex-col">
            <header class="flex h-14 shrink-0 items-center justify-between border-b border-line px-3 md:h-15 md:px-4">
              <div class="flex min-w-0 items-center gap-3">
                <button type="button" class="gf-icon-button -ml-2 h-8 w-8 hover:bg-base-300 hover:text-base-content md:hidden" @click="active = null">
                  <ArrowLeft class="h-5 w-5" />
                </button>
                <UserAvatar :src="active.peerAvatar" :alt="active.peerUsername" class="h-9 w-9 rounded-full object-cover ring-1 ring-line" />
                <div class="min-w-0">
                  <a :href="active.peerUrl" class="truncate text-sm font-bold text-base-content hover:text-primary">{{ active.peerUsername }}</a>
                  <p class="text-xs text-base-content/55">{{ t('messages.conversation') }}</p>
                </div>
              </div>
              <button type="button" class="gf-icon-button h-8 w-8 hover:bg-base-300 hover:text-base-content">
                <MoreVertical class="h-4 w-4" />
              </button>
            </header>

            <div ref="messagesEl" class="min-h-0 flex-1 space-y-3 overflow-y-auto px-3 py-3 md:space-y-4 md:px-4 md:py-4" @scroll.passive="handleMessagesScroll">
              <div class="flex justify-center">
                <span class="bg-base-200 px-2 py-1 text-xs font-medium text-base-content/55 [border-radius:var(--gf-radius-selector)]">{{ t('messages.today') }}</span>
              </div>

              <div v-if="active.loading" class="py-12 text-center text-sm text-base-content/55">{{ t('messages.loading') }}</div>
              <template v-else-if="active.messages.length">
                <div v-if="active.loadingOlder" class="py-1 text-center text-xs text-base-content/45">{{ t('messages.loading') }}</div>
                <div
                  v-for="message in active.messages"
                  :key="message.id"
                  class="flex max-w-[88%] gap-2 md:max-w-[82%]"
                  :class="message.isSelf ? 'ml-auto flex-row-reverse' : ''"
                >
                  <UserAvatar v-if="!message.isSelf" :src="active.peerAvatar" :alt="active.peerUsername" class="mt-auto h-8 w-8 rounded-full object-cover ring-1 ring-line" />
                  <div class="group relative min-w-0">
                    <div
                      class="whitespace-pre-wrap break-words px-3 py-2 text-sm leading-relaxed shadow-sm [border-radius:var(--gf-radius-box)] md:px-4"
                      :class="message.isSelf ? 'bg-primary text-primary-content' : 'bg-base-300 text-base-content'"
                    >
                      {{ message.content }}
                    </div>
                    <time class="mt-1 block text-[11px] text-base-content/55" :class="message.isSelf ? 'text-right' : ''">{{ formatChatTime(message.createdAt) }}</time>
                  </div>
                </div>
              </template>
              <div v-else class="flex h-full flex-col items-center justify-center text-center">
                <MessageSquare class="h-10 w-10 text-base-content/35" />
                <h2 class="mt-3 text-base font-semibold text-base-content">{{ t('messages.startChat') }}</h2>
                <p class="mt-1 text-sm text-base-content/55">{{ t('messages.firstMessageTo', { user: active.peerUsername }) }}</p>
              </div>
            </div>

            <footer class="shrink-0 border-t border-line bg-base-100/95 px-3 py-2 md:px-4 md:py-3">
              <div class="mx-auto max-w-4xl">
                <p v-if="error" class="gf-status-message gf-status-message-error mb-2">{{ error }}</p>
                <div class="gf-panel bg-base-200/80 p-2 transition focus-within:border-primary/40 focus-within:bg-base-100 focus-within:shadow-[0_10px_28px_rgba(37,99,235,0.08)]">
                  <textarea
                    ref="messageInput"
                    v-model="newMessage"
                    rows="1"
                    class="block max-h-36 min-h-11 w-full resize-none bg-transparent px-2 py-2 text-[15px] leading-relaxed text-base-content outline-none placeholder:text-base-content/55"
                    :placeholder="t('messages.inputPlaceholder')"
                    @input="resizeMessageInput"
                    @keydown.enter="handleEnter"
                  />
                  <div class="mt-1 flex items-center justify-between gap-3 border-t border-line/70 px-1 pt-2">
                    <div class="flex min-w-0 items-center gap-2">
                      <div class="relative">
                        <button
                          type="button"
                          class="gf-icon-button h-8 w-8 hover:bg-base-100 hover:text-primary"
                          :title="t('messages.emoji')"
                          @click="showEmoji = !showEmoji"
                        >
                          <Smile class="h-5 w-5" />
                        </button>
                        <div v-if="showEmoji" class="gf-menu-surface absolute bottom-full left-0 z-20 mb-3 grid w-48 grid-cols-4 gap-1 p-2">
                          <button
                            v-for="emoji in emojis"
                            :key="emoji"
                            type="button"
                            class="p-1.5 text-xl hover:bg-base-200 [border-radius:var(--gf-radius-field)]"
                            @click="appendEmoji(emoji)"
                          >
                            {{ emoji }}
                          </button>
                        </div>
                      </div>
                      <span class="hidden truncate text-[11px] font-medium text-base-content/55 sm:inline">{{ t('messages.enterHint') }}</span>
                    </div>
                    <button
                      type="button"
                      class="gf-button gf-button-sm gf-button-primary shrink-0 disabled:bg-base-300 disabled:text-base-content/55"
                      :disabled="!newMessage.trim() || sending"
                      @click="submitMessage"
                    >
                      <Send class="h-4 w-4" />
                      <span>{{ t('messages.send') }}</span>
                    </button>
                  </div>
                </div>
              </div>
            </footer>
          </div>

          <div v-else class="hidden flex-1 flex-col items-center justify-center p-8 text-center md:flex">
            <MessageSquare class="h-12 w-12 text-base-content/35" />
            <h2 class="mt-3 text-lg font-semibold text-base-content">{{ t('messages.selectConversation') }}</h2>
            <p class="mt-1 text-sm text-base-content/55">{{ t('messages.selectConversationDescription') }}</p>
            <button type="button" class="gf-button gf-button-md gf-button-primary mt-4" @click="showNewChat = true">
              {{ t('messages.startChat') }}
            </button>
          </div>
        </section>
      </section>

      <div v-if="showNewChat" class="fixed inset-0 z-[80] flex items-center justify-center bg-neutral/20 px-4 backdrop-blur-sm" @click.self="showNewChat = false">
        <div class="gf-menu-surface flex max-h-[80vh] w-full max-w-md flex-col overflow-hidden">
          <div class="flex h-13 items-center justify-between border-b border-line px-4">
            <h2 class="text-sm font-semibold text-base-content">{{ t('messages.newMessage') }}</h2>
            <button type="button" class="gf-icon-button p-1.5 hover:bg-base-300 hover:text-base-content" @click="showNewChat = false">
              <X class="h-4 w-4" />
            </button>
          </div>
          <div class="border-b border-line p-3">
            <label class="flex h-9 items-center gap-2 border border-line bg-base-200 px-3 text-sm text-base-content/55 [border-radius:var(--gf-radius-field)]">
              <Search class="h-4 w-4" />
              <input v-model="userSearch" class="min-w-0 flex-1 bg-transparent outline-none" :placeholder="t('messages.searchUsers')" />
            </label>
          </div>
          <div class="min-h-0 overflow-y-auto p-2">
            <button
              v-for="user in filteredUsers"
              :key="user.id"
              type="button"
              class="flex w-full items-center gap-3 p-3 text-left hover:bg-base-200 [border-radius:var(--gf-radius-field)]"
              @click="startChat(user)"
            >
              <UserAvatar :src="user.avatarUrl" :alt="user.username" class="h-10 w-10 rounded-full object-cover ring-1 ring-line" />
              <div class="min-w-0">
                <div class="truncate text-sm font-semibold text-base-content">{{ user.nickname || user.username }}</div>
                <div class="truncate text-xs text-base-content/55">@{{ user.username }}</div>
              </div>
            </button>
            <p v-if="!filteredUsers.length" class="px-4 py-8 text-center text-sm text-base-content/55">{{ t('messages.noContactableUsers') }}</p>
          </div>
        </div>
      </div>
    </main>
</template>
