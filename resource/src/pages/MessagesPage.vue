<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { ArrowLeft, MessageSquare, MoreVertical, PenLine, Search, Send, Smile, X } from '@lucide/vue'
import { getChatMessages, markChatRead, sendChatMessage, type ChatMessagePayload } from '../runtime/api'
import { formatDateTime } from '../runtime/format'
import type { LayoutPayload, MessageConversationPayload, MessagesPageProps, UserConnectionPayload } from '../types/payload'

type ChatConversation = MessageConversationPayload & {
  messages: ChatMessagePayload[]
  loading?: boolean
}

const page = defineProps<{
  layout: LayoutPayload
  props: MessagesPageProps
}>()

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
    const username = params.get('username') || '用户'
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
  if (conversation.convId && conversation.messages.length === 0) {
    conversation.loading = true
    try {
      const messages = await getChatMessages(conversation.convId)
      conversation.messages = messages.reverse()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '消息加载失败'
    } finally {
      conversation.loading = false
    }
  }
  if (conversation.unreadCount > 0 && conversation.convId) {
    conversation.unreadCount = 0
    void markChatRead(conversation.convId).catch(() => undefined)
  }
  await scrollToBottom()
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
    conversations.value = [active.value, ...conversations.value.filter((item) => item.peerId !== active.value?.peerId)]
    newMessage.value = ''
    resizeMessageInput()
    showEmoji.value = false
    await scrollToBottom()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '发送失败'
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
  }
  conversations.value = [conversation, ...conversations.value]
  await selectConversation(conversation)
}
</script>

<template>
    <main class="h-[calc(100dvh-5.5rem)] min-h-[620px] min-w-0 pb-3">
      <section class="grid h-full overflow-hidden rounded-lg border border-gray-200/70 bg-white shadow-[0_2px_8px_rgba(0,0,0,0.02)] md:grid-cols-[300px_minmax(0,1fr)]">
        <aside
          class="flex min-h-0 flex-col border-gray-100 bg-gray-50/50 md:border-r"
          :class="active ? 'hidden md:flex' : 'flex'"
        >
          <div class="flex h-15 shrink-0 items-center justify-between border-b border-gray-100 bg-white px-4">
            <h1 class="text-lg font-bold text-gray-950">私信</h1>
            <button
              type="button"
              class="inline-flex h-8 w-8 items-center justify-center rounded-md text-gray-500 hover:bg-gray-100 hover:text-gray-900"
              title="新私信"
              @click="showNewChat = true"
            >
              <PenLine class="h-4 w-4" />
            </button>
          </div>

          <div class="border-b border-gray-100 p-3">
            <label class="flex h-9 items-center gap-2 rounded-md border border-gray-200 bg-white px-3 text-sm text-gray-500">
              <Search class="h-4 w-4" />
              <input v-model="search" class="min-w-0 flex-1 bg-transparent outline-none" placeholder="搜索会话" />
            </label>
          </div>

          <div v-if="filteredConversations.length" class="min-h-0 flex-1 overflow-y-auto">
            <button
              v-for="conversation in filteredConversations"
              :key="conversation.peerId"
              type="button"
              class="flex w-full gap-3 border-b border-gray-100 px-4 py-3 text-left transition hover:bg-white"
              :class="active?.peerId === conversation.peerId ? 'bg-white shadow-[inset_3px_0_0_#2563eb]' : ''"
              @click="selectConversation(conversation)"
            >
              <img :src="conversation.peerAvatar" :alt="conversation.peerUsername" class="h-10 w-10 shrink-0 rounded-full object-cover ring-1 ring-gray-100" />
              <div class="min-w-0 flex-1">
                <div class="flex items-baseline justify-between gap-2">
                  <span class="truncate text-sm font-semibold text-gray-950">{{ conversation.peerUsername }}</span>
                  <time class="shrink-0 text-[11px] text-gray-400">{{ conversation.lastMsgTime ? formatDateTime(conversation.lastMsgTime) : '' }}</time>
                </div>
                <div class="mt-1 flex items-center gap-2">
                  <p class="min-w-0 flex-1 truncate text-sm" :class="conversation.unreadCount ? 'font-semibold text-gray-900' : 'text-gray-500'">
                    {{ conversation.lastMsg || '还没有消息' }}
                  </p>
                  <span v-if="conversation.unreadCount" class="h-2 w-2 shrink-0 rounded-full bg-blue-600" />
                </div>
              </div>
            </button>
          </div>

          <div v-else class="flex min-h-0 flex-1 flex-col items-center justify-center px-6 text-center">
            <MessageSquare class="h-10 w-10 text-gray-300" />
            <h2 class="mt-3 text-base font-semibold text-gray-950">暂无会话</h2>
            <p class="mt-1 text-sm text-gray-500">开始一次私信聊天。</p>
            <button type="button" class="mt-4 rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white hover:bg-blue-700" @click="showNewChat = true">
              新私信
            </button>
          </div>
        </aside>

        <section class="relative min-h-0 min-w-0 bg-white" :class="active ? 'flex' : 'hidden md:flex'">
          <div v-if="active" class="flex min-h-0 w-full flex-col">
            <header class="flex h-15 shrink-0 items-center justify-between border-b border-gray-100 px-4">
              <div class="flex min-w-0 items-center gap-3">
                <button type="button" class="-ml-2 inline-flex h-8 w-8 items-center justify-center rounded-md text-gray-500 hover:bg-gray-100 md:hidden" @click="active = null">
                  <ArrowLeft class="h-5 w-5" />
                </button>
                <img :src="active.peerAvatar" :alt="active.peerUsername" class="h-9 w-9 rounded-full object-cover ring-1 ring-gray-100" />
                <div class="min-w-0">
                  <a :href="active.peerUrl" class="truncate text-sm font-bold text-gray-950 hover:text-blue-600">{{ active.peerUsername }}</a>
                  <p class="text-xs text-gray-400">私信对话</p>
                </div>
              </div>
              <button type="button" class="inline-flex h-8 w-8 items-center justify-center rounded-md text-gray-400 hover:bg-gray-100 hover:text-gray-700">
                <MoreVertical class="h-4 w-4" />
              </button>
            </header>

            <div ref="messagesEl" class="min-h-0 flex-1 space-y-4 overflow-y-auto px-4 py-4">
              <div class="flex justify-center">
                <span class="rounded-full bg-gray-50 px-2 py-1 text-xs font-medium text-gray-400">今天</span>
              </div>

              <div v-if="active.loading" class="py-12 text-center text-sm text-gray-400">消息加载中...</div>
              <template v-else-if="active.messages.length">
                <div
                  v-for="message in active.messages"
                  :key="message.id"
                  class="flex max-w-[82%] gap-2"
                  :class="message.isSelf ? 'ml-auto flex-row-reverse' : ''"
                >
                  <img v-if="!message.isSelf" :src="active.peerAvatar" :alt="active.peerUsername" class="mt-auto h-8 w-8 rounded-full object-cover ring-1 ring-gray-100" />
                  <div class="group relative min-w-0">
                    <div
                      class="whitespace-pre-wrap break-words rounded-2xl px-4 py-2 text-sm leading-relaxed shadow-sm"
                      :class="message.isSelf ? 'rounded-br-sm bg-blue-600 text-white' : 'rounded-bl-sm bg-gray-100 text-gray-900'"
                    >
                      {{ message.content }}
                    </div>
                    <time class="mt-1 block text-[11px] text-gray-400" :class="message.isSelf ? 'text-right' : ''">{{ formatDateTime(message.createdAt) }}</time>
                  </div>
                </div>
              </template>
              <div v-else class="flex h-full flex-col items-center justify-center text-center">
                <MessageSquare class="h-10 w-10 text-gray-300" />
                <h2 class="mt-3 text-base font-semibold text-gray-950">开始聊天</h2>
                <p class="mt-1 text-sm text-gray-500">给 {{ active.peerUsername }} 发第一条消息。</p>
              </div>
            </div>

            <footer class="shrink-0 border-t border-gray-100 bg-white/95 px-4 py-3">
              <div class="mx-auto max-w-4xl">
                <p v-if="error" class="mb-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-600">{{ error }}</p>
                <div class="rounded-2xl border border-gray-200 bg-gray-50/80 p-2 transition focus-within:border-blue-300 focus-within:bg-white focus-within:shadow-[0_10px_28px_rgba(37,99,235,0.08)]">
                  <textarea
                    ref="messageInput"
                    v-model="newMessage"
                    rows="1"
                    class="block max-h-36 min-h-11 w-full resize-none bg-transparent px-2 py-2 text-[15px] leading-relaxed text-gray-900 outline-none placeholder:text-gray-400"
                    placeholder="输入消息..."
                    @input="resizeMessageInput"
                    @keydown.enter="handleEnter"
                  />
                  <div class="mt-1 flex items-center justify-between gap-3 border-t border-gray-200/70 px-1 pt-2">
                    <div class="flex min-w-0 items-center gap-2">
                      <div class="relative">
                        <button
                          type="button"
                          class="inline-flex h-8 w-8 items-center justify-center rounded-full text-gray-400 transition hover:bg-white hover:text-blue-600 hover:shadow-sm"
                          title="表情"
                          @click="showEmoji = !showEmoji"
                        >
                          <Smile class="h-5 w-5" />
                        </button>
                        <div v-if="showEmoji" class="absolute bottom-full left-0 z-20 mb-3 grid w-48 grid-cols-4 gap-1 rounded-2xl border border-gray-100 bg-white p-2 shadow-xl">
                          <button
                            v-for="emoji in emojis"
                            :key="emoji"
                            type="button"
                            class="rounded-lg p-1.5 text-xl hover:bg-gray-50"
                            @click="appendEmoji(emoji)"
                          >
                            {{ emoji }}
                          </button>
                        </div>
                      </div>
                      <span class="hidden truncate text-[11px] font-medium text-gray-400 sm:inline">Enter 发送，Shift + Enter 换行</span>
                    </div>
                    <button
                      type="button"
                      class="inline-flex h-8 shrink-0 items-center gap-1.5 rounded-full bg-blue-600 px-3 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-gray-200 disabled:text-gray-400 disabled:shadow-none"
                      :disabled="!newMessage.trim() || sending"
                      @click="submitMessage"
                    >
                      <Send class="h-4 w-4" />
                      <span>发送</span>
                    </button>
                  </div>
                </div>
              </div>
            </footer>
          </div>

          <div v-else class="hidden flex-1 flex-col items-center justify-center p-8 text-center md:flex">
            <MessageSquare class="h-12 w-12 text-gray-300" />
            <h2 class="mt-3 text-lg font-semibold text-gray-950">选择一个会话</h2>
            <p class="mt-1 text-sm text-gray-500">从左侧选择会话，或开始新的私信。</p>
            <button type="button" class="mt-4 rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white hover:bg-blue-700" @click="showNewChat = true">
              开始聊天
            </button>
          </div>
        </section>
      </section>

      <div v-if="showNewChat" class="fixed inset-0 z-[80] flex items-center justify-center bg-black/20 px-4 backdrop-blur-sm" @click.self="showNewChat = false">
        <div class="flex max-h-[80vh] w-full max-w-md flex-col overflow-hidden rounded-lg border border-gray-100 bg-white shadow-xl">
          <div class="flex h-13 items-center justify-between border-b border-gray-100 px-4">
            <h2 class="text-sm font-semibold text-gray-950">新私信</h2>
            <button type="button" class="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700" @click="showNewChat = false">
              <X class="h-4 w-4" />
            </button>
          </div>
          <div class="border-b border-gray-100 p-3">
            <label class="flex h-9 items-center gap-2 rounded-md border border-gray-200 bg-gray-50 px-3 text-sm text-gray-500">
              <Search class="h-4 w-4" />
              <input v-model="userSearch" class="min-w-0 flex-1 bg-transparent outline-none" placeholder="搜索用户" />
            </label>
          </div>
          <div class="min-h-0 overflow-y-auto p-2">
            <button
              v-for="user in filteredUsers"
              :key="user.id"
              type="button"
              class="flex w-full items-center gap-3 rounded-md p-3 text-left hover:bg-gray-50"
              @click="startChat(user)"
            >
              <img :src="user.avatarUrl" :alt="user.username" class="h-10 w-10 rounded-full object-cover ring-1 ring-gray-100" />
              <div class="min-w-0">
                <div class="truncate text-sm font-semibold text-gray-950">{{ user.nickname || user.username }}</div>
                <div class="truncate text-xs text-gray-400">@{{ user.username }}</div>
              </div>
            </button>
            <p v-if="!filteredUsers.length" class="px-4 py-8 text-center text-sm text-gray-500">暂无可联系用户。</p>
          </div>
        </div>
      </div>
    </main>
</template>
