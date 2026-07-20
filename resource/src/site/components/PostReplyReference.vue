<script setup lang="ts">
import { ChevronDown, ChevronUp } from '@lucide/vue'
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import UserAvatar from '@/site/components/UserAvatar.vue'
import type { ReplyTargetPayload } from '@/types/payload'

const props = defineProps<{
  target?: ReplyTargetPayload
}>()

const { t } = useI18n()
const expanded = ref(false)
const overflowing = ref(false)
const contentEl = ref<HTMLElement | null>(null)
const collapsedLines = 4
let resizeObserver: ResizeObserver | undefined

function measureOverflow() {
  const content = contentEl.value
  if (!content) {
    overflowing.value = false
    return
  }

  const lineHeight = Number.parseFloat(window.getComputedStyle(content).lineHeight)
  overflowing.value = content.scrollHeight > lineHeight * collapsedLines + 1
}

function toggleExpanded() {
  expanded.value = !expanded.value
}

watch(
  () => props.target?.renderedContent,
  async () => {
    expanded.value = false
    await nextTick()
    measureOverflow()
  },
)

onMounted(async () => {
  await nextTick()
  measureOverflow()
  if ('ResizeObserver' in window && contentEl.value) {
    resizeObserver = new ResizeObserver(measureOverflow)
    resizeObserver.observe(contentEl.value)
  }
})

onBeforeUnmount(() => resizeObserver?.disconnect())
</script>

<template>
  <aside class="mb-2 border-l-2 border-primary/45 bg-base-200/40 py-2">
    <div class="flex min-h-7 items-center gap-2 px-3 text-sm text-base-content/55">
      <UserAvatar
        v-if="target && !target.unavailable"
        :src="target.author.avatarUrl"
        :alt="target.author.username"
        class="h-6 w-6 rounded-full object-cover ring-1 ring-line"
      />
      <span v-if="target?.author.username" class="min-w-0 truncate font-medium text-base-content/75">@{{ target.author.username }}</span>
      <span v-if="target?.postNo" class="shrink-0 text-xs text-base-content/45">#{{ target.postNo }}</span>
    </div>

    <div v-if="!target || target.unavailable" class="px-3 pt-2 text-sm text-base-content/45">
      {{ t('topic.replyTargetUnavailable') }}
    </div>
    <template v-else>
      <div class="px-3 pt-2">
        <div
          ref="contentEl"
          class="gf-prose gf-prose-post"
          :class="{
            'reply-reference-content--collapsed': !expanded,
            'reply-reference-content--faded': !expanded && overflowing,
          }"
          v-html="target.renderedContent"
        />
      </div>
      <button
        v-if="overflowing"
        type="button"
        class="mx-2.5 mt-1 inline-flex h-7 items-center gap-1 rounded px-1.5 text-xs font-medium text-primary transition hover:bg-info/10"
        :aria-expanded="expanded"
        @click="toggleExpanded"
      >
        <ChevronUp v-if="expanded" class="h-3.5 w-3.5" />
        <ChevronDown v-else class="h-3.5 w-3.5" />
        {{ expanded ? t('topic.collapseReply') : t('topic.expandReply') }}
      </button>
    </template>
  </aside>
</template>

<style scoped>
.reply-reference-content--collapsed {
  max-height: 4lh;
  overflow: hidden;
}

.reply-reference-content--faded {
  mask-image: linear-gradient(to bottom, black calc(100% - 1.5lh), transparent 100%);
}
</style>
