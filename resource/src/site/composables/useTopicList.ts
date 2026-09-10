import { computed, nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchPage } from '@/runtime/router'
import { useTopicListMode } from '@/site/composables/useTopicListMode'
import type { HomeProps, PagePayload, TopicPayload } from '@gooseforum/client'

type TopicListProps = Pick<HomeProps, 'topics' | 'pagination'>

export function useTopicList(page: { readonly props: TopicListProps; readonly pageUrl: string }) {
  const { t } = useI18n()
  const { mode: listMode, setMode: setListMode } = useTopicListMode()
  const topics = ref<TopicPayload[]>([])
  const pagination = ref(page.props.pagination)
  const loadingMore = ref(false)
  const loadError = ref('')
  const loadMoreSentinel = ref<HTMLElement | null>(null)
  const hasTopics = computed(() => topics.value.length > 0)
  let observer: IntersectionObserver | undefined
  let revision = 0
  let active = true

  function reset() {
    // A pending response belongs to the old page/mode, even if its URL is revisited.
    revision++
    topics.value = [...page.props.topics]
    pagination.value = page.props.pagination
    loadingMore.value = false
    loadError.value = ''
  }

  watch(() => page.pageUrl, () => {
    reset()
    void nextTick(observeSentinel)
  }, { immediate: true })

  watch(() => page.props.topics, (incoming) => {
    const unseenByID = new Map(incoming.map((topic) => [topic.id, topic.unseen]))
    topics.value = topics.value.map((topic) =>
      unseenByID.has(topic.id) ? { ...topic, unseen: unseenByID.get(topic.id) } : topic,
    )
  })

  watch(listMode, (mode) => {
    if (mode === 'pagination') reset()
    void nextTick(observeSentinel)
  })

  async function loadMore() {
    if (!active || listMode.value !== 'waterfall' || loadingMore.value || !pagination.value.hasNext || !pagination.value.nextUrl) return
    const requestRevision = revision
    loadingMore.value = true
    loadError.value = ''
    try {
      const payload = await fetchPage(new URL(pagination.value.nextUrl, window.location.origin)) as PagePayload<TopicListProps>
      if (requestRevision !== revision) return
      const seen = new Set(topics.value.map((topic) => topic.id))
      topics.value = [...topics.value, ...payload.props.topics.filter((topic) => {
        if (seen.has(topic.id)) return false
        seen.add(topic.id)
        return true
      })]
      pagination.value = payload.props.pagination
    } catch (error) {
      if (requestRevision === revision) loadError.value = error instanceof Error ? error.message : t('common.loadFailed')
    } finally {
      if (requestRevision === revision) loadingMore.value = false
    }
  }

  function observeSentinel() {
    observer?.disconnect()
    if (!active || listMode.value !== 'waterfall' || !loadMoreSentinel.value || !('IntersectionObserver' in window)) return
    observer = new IntersectionObserver((entries) => {
      if (entries.some((entry) => entry.isIntersecting)) void loadMore()
    }, { rootMargin: '480px 0px' })
    observer.observe(loadMoreSentinel.value)
  }

  function deactivate() {
    active = false
    revision++
    loadingMore.value = false
    observer?.disconnect()
  }

  onMounted(observeSentinel)
  onActivated(() => {
    active = true
    void nextTick(observeSentinel)
  })
  onDeactivated(deactivate)
  onBeforeUnmount(deactivate)

  return { topics, pagination, hasTopics, listMode, setListMode, loadingMore, loadError, loadMoreSentinel, loadMore }
}
