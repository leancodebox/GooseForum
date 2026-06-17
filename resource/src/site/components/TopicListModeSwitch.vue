<script setup lang="ts">
import { computed } from 'vue'
import { Grid3X3, List } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { useFlashMessages } from '@/runtime/flash-message'
import type { TopicListMode } from '@/site/composables/useTopicListMode'

const props = defineProps<{
  modelValue: TopicListMode
}>()

const emit = defineEmits<{
  'update:modelValue': [value: TopicListMode]
}>()

const { t } = useI18n()
const flash = useFlashMessages()

const nextMode = computed<TopicListMode>(() => props.modelValue === 'waterfall' ? 'pagination' : 'waterfall')
const label = computed(() => nextMode.value === 'pagination' ? t('topicList.mode.pagination') : t('topicList.mode.waterfall'))
const ModeIcon = computed(() => nextMode.value === 'pagination' ? List : Grid3X3)

function switchMode() {
  emit('update:modelValue', nextMode.value)
  flash.push(t('topicList.mode.changed', { mode: label.value }), 'info')
}
</script>

<template>
  <div class="gf-list-mode-switch">
    <button
      type="button"
      class="gf-list-mode-button"
      :aria-label="t('topicList.mode.switchTo', { mode: label })"
      :title="t('topicList.mode.switchTo', { mode: label })"
      @click="switchMode"
    >
      <component :is="ModeIcon" class="h-4 w-4" aria-hidden="true" />
      <span class="sr-only">{{ label }}</span>
    </button>
  </div>
</template>
