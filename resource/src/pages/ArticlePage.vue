<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { Bookmark, Clock, CornerDownLeft, Eye, Heart, MessageSquare, PencilLine, Send, X } from '@lucide/vue'
import AppShell from '../components/AppShell.vue'
import { bookmarkArticle, likeArticle, postReply } from '../runtime/api'
import { formatDateTime, formatNumber } from '../runtime/format'
import { fetchPage } from '../runtime/router'
import { scheduleHideUserCard, showUserCard } from '../runtime/user-card-events'
import type { ArticleDetailProps, LayoutPayload } from '../types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: ArticleDetailProps
}>()

const replyContent = ref('')
const replyContents = reactive<Record<number, string>>({})
const openReplyId = ref<number | null>(null)
const currentReplyId = ref(0)
const likeCount = ref(page.props.article.likeCount)
const isLiked = ref(page.props.article.isLiked)
const isBookmarked = ref(page.props.article.isBookmarked)
const actionMessage = ref('')
const actingLike = ref(false)
const actingBookmark = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const titleEl = ref<HTMLElement | null>(null)
const showHeaderTitle = ref(false)
let titleObserver: IntersectionObserver | undefined

function observeTitle() {
  titleObserver?.disconnect()
  showHeaderTitle.value = false

  if (!titleEl.value || !('IntersectionObserver' in window)) return

  titleObserver = new IntersectionObserver(
    (entries) => {
      showHeaderTitle.value = !entries[0]?.isIntersecting
    },
    { threshold: 0, rootMargin: '-80px 0px 0px 0px' },
  )
  titleObserver.observe(titleEl.value)
}

onMounted(() => {
  void nextTick(observeTitle)
})

watch(
  () => page.props.article.id,
  () => {
    likeCount.value = page.props.article.likeCount
    isLiked.value = page.props.article.isLiked
    isBookmarked.value = page.props.article.isBookmarked
    void nextTick(observeTitle)
  },
)

onBeforeUnmount(() => {
  titleObserver?.disconnect()
})

async function toggleLike() {
  if (actingLike.value) return

  const nextLiked = !isLiked.value
  const previousLiked = isLiked.value
  const previousCount = likeCount.value
  actingLike.value = true
  actionMessage.value = ''
  isLiked.value = nextLiked
  likeCount.value = Math.max(0, likeCount.value + (nextLiked ? 1 : -1))
  try {
    await likeArticle(page.props.article.id, nextLiked ? 1 : 2)
  } catch (error) {
    isLiked.value = previousLiked
    likeCount.value = previousCount
    actionMessage.value = error instanceof Error ? error.message : '点赞失败'
  } finally {
    actingLike.value = false
  }
}

async function toggleBookmark() {
  if (actingBookmark.value) return

  const nextBookmarked = !isBookmarked.value
  const previousBookmarked = isBookmarked.value
  actingBookmark.value = true
  actionMessage.value = ''
  isBookmarked.value = nextBookmarked
  try {
    await bookmarkArticle(page.props.article.id, nextBookmarked ? 1 : 2)
    actionMessage.value = nextBookmarked ? '已收藏。' : '已取消收藏。'
  } catch (error) {
    isBookmarked.value = previousBookmarked
    actionMessage.value = error instanceof Error ? error.message : '收藏失败'
  } finally {
    actingBookmark.value = false
  }
}

function toggleReplyForm(replyId: number) {
  openReplyId.value = openReplyId.value === replyId ? null : replyId
  if (replyContents[replyId] === undefined) {
    replyContents[replyId] = ''
  }
}

async function submitReply(replyId = 0) {
  const content = replyId > 0 ? (replyContents[replyId] || '').trim() : replyContent.value.trim()
  if (!content || submitting.value) return

  submitting.value = true
  currentReplyId.value = replyId
  errorMessage.value = ''
  successMessage.value = ''
  try {
    await postReply(page.props.article.id, content, replyId)
    if (replyId > 0) {
      replyContents[replyId] = ''
      openReplyId.value = null
    } else {
      replyContent.value = ''
    }
    successMessage.value = replyId > 0 ? '回复已发布。' : '回复已发布。'
    const payload = await fetchPage(new URL(window.location.href))
    history.replaceState({ goose: true, payload }, '', window.location.href)
    window.dispatchEvent(new CustomEvent('goose:page', { detail: payload }))
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '回复失败'
  } finally {
    submitting.value = false
    currentReplyId.value = 0
  }
}
</script>

<template>
  <AppShell
    :layout="layout"
    :header-title="page.props.article.title"
    :show-header-title="showHeaderTitle"
    rail
  >
    <article class="min-w-0 pb-12">
      <header class="mb-4 border-b border-gray-200/70 pb-4">
        <h1 ref="titleEl" class="text-2xl font-bold leading-tight text-gray-950 sm:text-3xl">{{ page.props.article.title }}</h1>
        <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-[13px] text-gray-500">
          <a
            :href="`/u/${page.props.article.author.id}`"
            class="inline-flex items-center gap-2 font-medium text-gray-700 hover:text-blue-600"
            @mouseenter="showUserCard(page.props.article.author, $event)"
            @mouseleave="scheduleHideUserCard"
            @focus="showUserCard(page.props.article.author, $event)"
            @blur="scheduleHideUserCard"
          >
            <img :src="page.props.article.author.avatarUrl" :alt="page.props.article.author.username" class="h-5 w-5 rounded-full object-cover" />
            {{ page.props.article.author.username }}
          </a>
          <span class="inline-flex items-center gap-1.5">
            <Clock class="h-3.5 w-3.5" />
            {{ formatDateTime(page.props.article.createdAt) }}
          </span>
          <a
            v-for="category in page.props.article.categories"
            :key="category.id"
            :href="category.url"
            class="inline-flex items-center gap-1.5 rounded-sm text-gray-600 hover:text-blue-600"
          >
            <span class="h-2 w-2 rounded-[3px]" :style="{ backgroundColor: category.color }" />
            {{ category.name }}
          </a>
        </div>
      </header>

      <section class="rounded-lg border border-gray-200/70 bg-white shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
        <div class="grid grid-cols-[44px_minmax(0,1fr)] gap-3 p-4 sm:grid-cols-[52px_minmax(0,1fr)] sm:gap-4 sm:p-5">
          <a
            :href="`/u/${page.props.article.author.id}`"
            class="sticky top-19 self-start pt-1"
            @mouseenter="showUserCard(page.props.article.author, $event)"
            @mouseleave="scheduleHideUserCard"
            @focus="showUserCard(page.props.article.author, $event)"
            @blur="scheduleHideUserCard"
          >
            <img :src="page.props.article.author.avatarUrl" :alt="page.props.article.author.username" class="h-11 w-11 rounded-full object-cover ring-1 ring-gray-100" />
          </a>
          <div class="min-w-0">
            <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
              <div>
                <a :href="`/u/${page.props.article.author.id}`" class="font-semibold text-gray-950 hover:text-blue-600">{{ page.props.article.author.username }}</a>
                <div class="text-xs font-medium text-gray-600">原帖</div>
              </div>
              <div class="flex flex-wrap items-center justify-end gap-3 text-xs font-medium text-gray-600">
                <div class="flex items-center gap-3">
                  <span class="inline-flex items-center gap-1"><MessageSquare class="h-3.5 w-3.5" />{{ formatNumber(page.props.article.replyCount) }}</span>
                  <span class="inline-flex items-center gap-1"><Eye class="h-3.5 w-3.5" />{{ formatNumber(page.props.article.viewCount) }}</span>
                  <span class="inline-flex items-center gap-1"><Heart class="h-3.5 w-3.5" />{{ formatNumber(likeCount) }}</span>
                </div>
                <a
                  v-if="page.props.permissions.isOwnArticle"
                  :href="`/publish?id=${page.props.article.id}`"
                  class="inline-flex h-7 items-center gap-1.5 rounded-md border border-gray-200 bg-white px-2 text-xs font-semibold text-gray-600 transition hover:border-blue-200 hover:bg-blue-50 hover:text-blue-700"
                >
                  <PencilLine class="h-3.5 w-3.5" />
                  编辑
                </a>
              </div>
            </div>
            <div class="gf-prose gf-prose-article" v-html="page.props.article.html" />
            <div class="mt-6 flex flex-wrap items-center gap-3 border-t border-gray-100 pt-4">
              <button
                type="button"
                class="inline-flex h-8 items-center gap-1.5 rounded-md px-2.5 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-60"
                :class="isLiked ? 'bg-red-50 text-red-600 hover:bg-red-100' : 'text-gray-500 hover:bg-gray-50 hover:text-gray-800'"
                :disabled="actingLike"
                @click="toggleLike"
              >
                <Heart class="h-4 w-4" :fill="isLiked ? 'currentColor' : 'none'" />
                {{ likeCount ? formatNumber(likeCount) : '点赞' }}
              </button>
              <button
                type="button"
                class="inline-flex h-8 items-center gap-1.5 rounded-md px-2.5 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-60"
                :class="isBookmarked ? 'bg-blue-50 text-blue-700 hover:bg-blue-100' : 'text-gray-500 hover:bg-gray-50 hover:text-gray-800'"
                :disabled="actingBookmark"
                @click="toggleBookmark"
              >
                <Bookmark class="h-4 w-4" :fill="isBookmarked ? 'currentColor' : 'none'" />
                {{ isBookmarked ? '已收藏' : '收藏' }}
              </button>
              <span v-if="actionMessage" class="text-xs" :class="actionMessage.includes('失败') ? 'text-red-600' : 'text-gray-600'">{{ actionMessage }}</span>
            </div>
          </div>
        </div>

        <span v-if="page.props.replies.length" id="replies" class="block scroll-mt-20" aria-hidden="true" />

        <div
          v-for="(reply, index) in page.props.replies"
          :id="`reply-${reply.id}`"
          :key="reply.id"
          class="group grid grid-cols-[44px_minmax(0,1fr)] gap-3 border-t border-gray-100 p-4 transition hover:bg-gray-50/70 sm:grid-cols-[52px_minmax(0,1fr)] sm:gap-4 sm:p-5"
        >
          <a
            :href="`/u/${reply.author.id}`"
            class="sticky top-19 self-start pt-1"
            @mouseenter="showUserCard(reply.author, $event)"
            @mouseleave="scheduleHideUserCard"
            @focus="showUserCard(reply.author, $event)"
            @blur="scheduleHideUserCard"
          >
            <img :src="reply.author.avatarUrl" :alt="reply.author.username" class="h-10 w-10 rounded-full object-cover ring-1 ring-gray-100" />
          </a>
          <div class="min-w-0">
            <div class="mb-2 flex flex-wrap items-center justify-between gap-3">
              <div class="flex items-center gap-2">
                <a :href="`/u/${reply.author.id}`" class="font-semibold text-gray-950 hover:text-blue-600">{{ reply.author.username }}</a>
                <span class="rounded bg-gray-100 px-1.5 py-0.5 text-[11px] font-medium text-gray-500">#{{ index + 1 }}</span>
              </div>
              <div class="flex items-center gap-3">
                <button
                  v-if="page.props.permissions.canReply"
                  type="button"
                  class="inline-flex min-h-9 items-center gap-1.5 rounded-md px-3 text-xs font-semibold text-gray-600 opacity-100 transition hover:bg-blue-50 hover:text-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 sm:opacity-0 sm:group-hover:opacity-100 sm:focus-visible:opacity-100"
                  @click="toggleReplyForm(reply.id)"
                >
                  <CornerDownLeft class="h-3.5 w-3.5" />
                  回复
                </button>
                <time class="text-xs text-gray-400">{{ formatDateTime(reply.createdAt) }}</time>
              </div>
            </div>
            <p v-if="reply.replyToUsername" class="mb-2 w-fit rounded bg-gray-50 px-2 py-1 text-sm text-gray-500">
              回复 <a :href="`/u/${reply.replyToUserId}`" class="font-medium text-gray-700 hover:text-blue-600">@{{ reply.replyToUsername }}</a>
            </p>
            <p class="whitespace-pre-wrap leading-relaxed text-gray-800">{{ reply.content }}</p>

            <div v-if="openReplyId === reply.id" class="mt-4 border-l-2 border-blue-100 pl-3">
              <div class="mb-2 flex items-center justify-between">
                <div class="text-xs font-medium text-gray-500">回复 @{{ reply.author.username }}</div>
                <button type="button" class="rounded-md p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-700" @click="openReplyId = null">
                  <X class="h-4 w-4" />
                </button>
              </div>
              <textarea
                v-model="replyContents[reply.id]"
                class="min-h-20 w-full resize-y rounded-md border border-gray-200 bg-white p-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
                placeholder="写下你的回复..."
              />
              <div class="mt-2 flex justify-end gap-2">
                <button type="button" class="h-8 rounded-md px-3 text-xs font-semibold text-gray-500 hover:bg-gray-100" @click="openReplyId = null">取消</button>
                <button
                  type="button"
                  class="inline-flex h-8 items-center gap-1.5 rounded-md bg-blue-600 px-3 text-xs font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
                  :disabled="submitting || !replyContents[reply.id]?.trim()"
                  @click="submitReply(reply.id)"
                >
                  <Send class="h-3.5 w-3.5" />
                  {{ submitting && currentReplyId === reply.id ? '发布中...' : '回复' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="mt-4 rounded-lg border border-gray-200/70 bg-white p-4 shadow-[0_2px_8px_rgba(0,0,0,0.02)] sm:p-5">
        <template v-if="page.props.permissions.canReply">
          <div class="mb-3 flex items-center justify-between">
            <label class="text-sm font-semibold text-gray-950" for="reply-content">参与讨论</label>
            <span class="text-xs text-gray-400">Markdown 支持稍后补齐</span>
          </div>
          <textarea
            id="reply-content"
            v-model="replyContent"
            class="min-h-28 w-full resize-y rounded-md border border-gray-200 p-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
            placeholder="写下你的回复..."
          />
          <p v-if="errorMessage" class="mt-2 text-sm text-red-600">{{ errorMessage }}</p>
          <p v-if="successMessage" class="mt-2 text-sm text-green-600">{{ successMessage }}</p>
          <div class="mt-3 flex justify-end">
            <button
              class="inline-flex h-9 items-center gap-1.5 rounded-md bg-blue-600 px-3 text-sm font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="submitting || !replyContent.trim()"
              @click="submitReply()"
            >
              <Send class="h-4 w-4" />
              {{ submitting ? '发布中...' : '发布回复' }}
            </button>
          </div>
        </template>
        <template v-else>
          <div class="text-center">
            <h2 class="text-base font-semibold text-gray-950">加入讨论</h2>
            <p class="mt-1 text-sm text-gray-500">登录后可以回复这个主题。</p>
            <a href="/login" class="mt-4 inline-flex h-9 items-center rounded-md bg-blue-600 px-3 text-sm font-semibold text-white hover:bg-blue-700">登录</a>
          </div>
        </template>
      </section>
    </article>

    <template #rail>
      <div class="sticky top-19 space-y-3">
        <div class="overflow-hidden rounded-lg border border-gray-200/70 bg-white shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
          <div class="border-b border-gray-100 px-4 py-4">
            <h2 class="text-sm font-semibold text-gray-500">主题概览</h2>
          </div>

          <dl class="space-y-4 px-4 py-5 text-sm">
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-gray-500">创建于</dt>
              <dd class="text-right font-semibold tabular-nums text-gray-950">{{ formatDateTime(page.props.article.createdAt) }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-gray-500">最后回复</dt>
              <dd class="text-right font-semibold tabular-nums text-gray-950">{{ formatDateTime(page.props.article.updatedAt) }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-gray-500">回复数</dt>
              <dd class="text-right font-semibold tabular-nums text-gray-950">{{ formatNumber(page.props.article.replyCount) }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-gray-500">阅读数</dt>
              <dd class="text-right font-semibold tabular-nums text-gray-950">{{ formatNumber(page.props.article.viewCount) }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-gray-500">参与者</dt>
              <dd class="text-right font-semibold tabular-nums text-gray-950">{{ page.props.article.participants.length }}</dd>
            </div>
          </dl>

          <div v-if="page.props.article.participants.length" class="border-t border-gray-100 px-4 py-4">
            <h3 class="mb-3 text-sm font-semibold text-gray-500">活跃参与者</h3>
            <div class="flex flex-wrap gap-1.5">
              <a
                v-for="participant in page.props.article.participants"
                :key="participant.id"
                :href="`/u/${participant.id}`"
                class="rounded-full"
                @mouseenter="showUserCard(participant, $event)"
                @mouseleave="scheduleHideUserCard"
                @focus="showUserCard(participant, $event)"
                @blur="scheduleHideUserCard"
              >
                <img :src="participant.avatarUrl" :alt="participant.username" class="h-8 w-8 rounded-full object-cover ring-1 ring-gray-100 transition hover:ring-blue-300" />
              </a>
            </div>
          </div>
        </div>

        <div v-if="page.props.hotTopics.length" class="rounded-lg border border-gray-200/70 bg-white p-3 shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
          <div class="mb-2 flex items-center justify-between">
            <h2 class="text-sm font-semibold text-gray-950">热门内容</h2>
            <a href="/?sort=hot" class="text-xs font-medium text-blue-600 hover:text-blue-700">更多</a>
          </div>
          <div class="space-y-0.5">
            <a
              v-for="topic in page.props.hotTopics"
              :key="topic.id"
              :href="topic.url"
              class="block rounded-md px-2 py-2 hover:bg-gray-50"
            >
              <div class="line-clamp-2 text-sm font-semibold leading-snug text-gray-900">{{ topic.title }}</div>
              <div class="mt-1 flex items-center gap-2 text-xs font-medium text-gray-600">
                <span class="truncate">{{ topic.author.username }}</span>
                <span class="shrink-0 tabular-nums">{{ formatNumber(topic.replyCount) }} 回复</span>
              </div>
            </a>
          </div>
        </div>

      </div>
    </template>
  </AppShell>
</template>
