<script setup lang="ts">
import type { DateValue } from '@internationalized/date'
import type { DateRange } from 'reka-ui'
import { CalendarDate, DateFormatter, getLocalTimeZone } from '@internationalized/date'
import { useMediaQuery } from '@vueuse/core'
import { Calendar } from '@lucide/vue'
import { computed, ref, shallowRef, watch } from 'vue'
import { Button } from '@/admin/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/admin/components/ui/popover'
import { RangeCalendar } from '@/admin/components/ui/range-calendar'
import { cn } from '@/admin/utils/cn'

const props = defineProps<{
  startDate: string
  endDate: string
}>()

const emit = defineEmits<{
  'update:startDate': [value: string]
  'update:endDate': [value: string]
  change: []
}>()

const formatter = new DateFormatter('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
})

const open = ref(false)
const isDesktop = useMediaQuery('(min-width: 640px)')
const value = shallowRef<DateRange>({
  start: parseDate(props.startDate),
  end: parseDate(props.endDate),
})

const label = computed(() => {
  if (!value.value.start) return '选择日期范围'
  const start = formatDateValue(value.value.start)
  if (!value.value.end) return start
  return `${start} - ${formatDateValue(value.value.end)}`
})

watch(
  () => [props.startDate, props.endDate],
  ([startDate, endDate]) => {
    const nextStart = parseDate(startDate)
    const nextEnd = parseDate(endDate)
    if (toISO(value.value.start) === startDate && toISO(value.value.end) === endDate) return
    value.value = { start: nextStart, end: nextEnd }
  },
)

function handleValueChange(next: DateRange) {
  value.value = next
  const start = toISO(next.start)
  const end = toISO(next.end)
  if (!start || !end) return
  emit('update:startDate', start)
  emit('update:endDate', end)
  emit('change')
  open.value = false
}

function parseDate(value?: string): CalendarDate | undefined {
  if (!value) return undefined
  const [year, month, day] = value.split('-').map(Number)
  if (!year || !month || !day) return undefined
  return new CalendarDate(year, month, day)
}

function toISO(value?: DateValue): string {
  if (!value) return ''
  return value.toString()
}

function formatDateValue(value: DateValue): string {
  return formatter.format(value.toDate(getLocalTimeZone()))
}
</script>

<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <Button
        variant="outline"
        :class="cn(
          'h-10 w-full min-w-0 justify-start rounded-lg bg-background px-3 text-left font-normal shadow-sm sm:w-[320px]',
          !value.start && 'text-muted-foreground',
        )"
      >
        <Calendar class="mr-2 h-4 w-4 text-muted-foreground" />
        <span class="truncate">{{ label }}</span>
      </Button>
    </PopoverTrigger>
    <PopoverContent class="max-h-[min(80vh,640px)] w-[calc(100vw-2rem)] overflow-auto p-0 sm:w-auto" align="end">
      <RangeCalendar
        :model-value="value"
        :number-of-months="isDesktop ? 2 : 1"
        initial-focus
        @update:model-value="handleValueChange"
      />
    </PopoverContent>
  </Popover>
</template>
