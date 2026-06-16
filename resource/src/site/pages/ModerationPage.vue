<script setup lang="ts">
import { ref, watch } from 'vue'
import { Ban, CircleAlert, RotateCcw, Scale } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { updateModerationArticleStatus } from '@/runtime/api'
import { fetchPage } from '@/runtime/router'
import EmptyState from '@/site/components/EmptyState.vue'
import PageHeader from '@/site/components/PageHeader.vue'
import TopicList from '@/site/components/TopicList.vue'
import type { LayoutPayload, ModerationPageProps, PagePayload, TopicPayload } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: ModerationPageProps
}>()

const { t } = useI18n()
const currentProps = ref<ModerationPageProps>(page.props)
const topics = ref<TopicPayload[]>([...page.props.topics])
const busyIds = ref<number[]>([])
const actionError = ref('')
const loadingList = ref(false)
const activeConsoleTab = ref<'ban' | 'guidance'>('ban')

const managementTabs = [
  { key: 'ban', icon: Ban },
  { key: 'guidance', icon: Scale },
]

watch(
  () => page.props,
  (next) => {
    currentProps.value = next
    topics.value = [...next.topics]
    actionError.value = ''
    busyIds.value = []
    loadingList.value = false
  },
  { immediate: true },
)

function isBusy(id: number) {
  return busyIds.value.includes(id)
}

async function loadModerationURL(url: string, options: { push?: boolean } = {}) {
  if (loadingList.value) return
  loadingList.value = true
  actionError.value = ''
  busyIds.value = []
  try {
    const nextURL = new URL(url, window.location.origin)
    const payload = (await fetchPage(nextURL)) as PagePayload<ModerationPageProps>
    currentProps.value = payload.props
    topics.value = [...payload.props.topics]
    if (options.push !== false) {
      window.history.pushState(window.history.state, '', `${nextURL.pathname}${nextURL.search}${nextURL.hash}`)
    }
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : t('common.loadFailed')
  } finally {
    loadingList.value = false
  }
}

async function moderateTopic(topic: TopicPayload) {
  if (isBusy(topic.id)) return
  busyIds.value = [...busyIds.value, topic.id]
  actionError.value = ''
  try {
    await updateModerationArticleStatus(topic.id, 'unban')
    topics.value = topics.value.filter(item => item.id !== topic.id)
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : t('api.moderationActionFailed')
  } finally {
    busyIds.value = busyIds.value.filter(id => id !== topic.id)
  }
}
</script>

<template>
  <main class="min-w-0 pb-8">
    <PageHeader :title="t('moderation.title')" :description="t('moderation.description')" compact />

    <div class="mb-4 flex flex-wrap gap-2 border-b border-line">
      <button
        v-for="tab in managementTabs"
        :key="tab.key"
        type="button"
        class="-mb-px inline-flex h-10 items-center gap-2 border-b-2 px-1 text-sm font-semibold transition"
        :class="activeConsoleTab === tab.key ? 'border-primary text-primary' : 'border-transparent text-base-content/55 hover:text-base-content'"
        @click="activeConsoleTab = tab.key as 'ban' | 'guidance'"
      >
        <component :is="tab.icon" class="h-4 w-4" />
        {{ t(`moderation.managementTabs.${tab.key}`) }}
      </button>
    </div>

    <section v-if="activeConsoleTab === 'ban'" class="space-y-3">
      <div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <div>
          <h2 class="text-base font-semibold text-base-content">{{ t('moderation.blockedTitle') }}</h2>
          <p class="mt-1 text-sm leading-6 text-base-content/60">
            {{ t('moderation.blockedDescription') }}
          </p>
        </div>
      </div>

      <div class="flex flex-wrap gap-2">
        <a
          v-for="tab in currentProps.categoryTabs"
          :key="tab.key"
          :href="tab.url"
          class="gf-button gf-button-sm text-xs"
          :class="tab.active ? 'gf-button-secondary' : 'gf-button-ghost'"
          @click.prevent="loadModerationURL(tab.url)"
        >
          {{ tab.label }}
        </a>
      </div>

      <p v-if="actionError" class="rounded border border-error/25 bg-error/10 px-3 py-2 text-sm text-error">
        {{ actionError }}
      </p>

      <div class="overflow-hidden rounded border border-line bg-base-100">
        <TopicList :topics="topics" :show-hot="false">
          <template #activity-header>
            {{ t('moderation.table.action') }}
          </template>
          <template #activity="{ topic }">
            <button
              type="button"
              class="gf-button gf-button-sm gf-button-primary shrink-0 text-xs"
              :disabled="isBusy(topic.id)"
              @click="moderateTopic(topic)"
            >
              <RotateCcw class="h-4 w-4" />
              {{ isBusy(topic.id) ? t('common.loadingShort') : t('moderation.unbanAction') }}
            </button>
          </template>
          <template #mobile-action="{ topic }">
            <span class="ml-auto">
              <button
                type="button"
                class="gf-button gf-button-sm gf-button-primary shrink-0 text-xs"
                :disabled="isBusy(topic.id)"
                @click="moderateTopic(topic)"
              >
                <RotateCcw class="h-4 w-4" />
                {{ isBusy(topic.id) ? t('common.loadingShort') : t('moderation.unbanAction') }}
              </button>
            </span>
          </template>
          <template #empty>
            <EmptyState v-if="!topics.length" :icon="Ban" :title="t('moderation.blockedEmptyTitle')" :description="t('moderation.emptyDescription')" />
          </template>
        </TopicList>

        <footer v-if="currentProps.pagination.hasNext" class="border-t border-line bg-base-200/50 px-4 py-3 text-center">
          <a
            :href="currentProps.pagination.nextUrl"
            class="gf-button gf-button-sm gf-button-secondary"
            rel="next"
            @click.prevent="loadModerationURL(currentProps.pagination.nextUrl)"
          >
            {{ t('common.nextPage') }}
          </a>
        </footer>
      </div>
    </section>

    <section v-else class="space-y-3 px-4 sm:px-0">
      <div>
        <h2 class="text-base font-semibold text-base-content">{{ t('moderation.guidanceTitle') }}</h2>
        <p class="mt-1 text-sm leading-6 text-base-content/60">{{ t('moderation.guidanceDescription') }}</p>
      </div>

      <div class="border-y border-line">
        <div class="flex items-start gap-3 border-b border-line py-3">
          <CircleAlert class="mt-0.5 h-4 w-4 shrink-0 text-warning" />
          <p class="text-sm leading-6 text-base-content/70">{{ t('moderation.notice') }}</p>
        </div>
        <div class="divide-y divide-line">
          <div class="py-3">
            <h3 class="text-sm font-semibold text-base-content">{{ t('moderation.guidanceItems.rule.title') }}</h3>
            <p class="mt-1 text-sm leading-6 text-base-content/60">{{ t('moderation.guidanceItems.rule.description') }}</p>
          </div>
          <div class="py-3">
            <h3 class="text-sm font-semibold text-base-content">{{ t('moderation.guidanceItems.context.title') }}</h3>
            <p class="mt-1 text-sm leading-6 text-base-content/60">{{ t('moderation.guidanceItems.context.description') }}</p>
          </div>
          <div class="py-3">
            <h3 class="text-sm font-semibold text-base-content">{{ t('moderation.guidanceItems.restraint.title') }}</h3>
            <p class="mt-1 text-sm leading-6 text-base-content/60">{{ t('moderation.guidanceItems.restraint.description') }}</p>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>
