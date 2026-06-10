<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Check, Grid3X3, SlidersHorizontal, X } from '@lucide/vue'

type PickerMode = 'palette' | 'picker'

const props = defineProps<{
  modelValue: string
  tokenLabel: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const isOpen = ref(false)
const mode = ref<PickerMode>('palette')
const localHex = ref(toHex(props.modelValue))

const rgb = computed(() => hexToRgb(localHex.value))
const paletteRows = computed(() => buildPaletteRows())
const activeLabel = computed(() => props.tokenLabel.replace(/-/g, ' '))
const displayLabel = computed(() => props.tokenLabel.replace(/-content$/, ' content'))

watch(
  () => props.modelValue,
  (value) => {
    localHex.value = toHex(value)
  },
)

function openPicker() {
  localHex.value = toHex(props.modelValue)
  isOpen.value = true
}

function closePicker() {
  isOpen.value = false
}

function choose(value: string) {
  localHex.value = value
  emit('update:modelValue', value)
}

function updateHex(value: string) {
  const normalized = normalizeHex(value)
  localHex.value = normalized
  emit('update:modelValue', normalized)
}

function updateChannel(channel: 'r' | 'g' | 'b', value: number) {
  const next = { ...rgb.value, [channel]: clamp(Math.round(value), 0, 255) }
  updateHex(rgbToHex(next.r, next.g, next.b))
}

function toHex(value: string) {
  return normalizeHex(value || rgbToHex(255, 255, 255))
}

function normalizeHex(value: string) {
  const trimmed = value.trim()
  if (/^#[0-9a-fA-F]{6}$/.test(trimmed)) return trimmed.toLowerCase()
  if (/^[0-9a-fA-F]{6}$/.test(trimmed)) return `#${trimmed.toLowerCase()}`
  if (/^#[0-9a-fA-F]{3}$/.test(trimmed)) {
    const [, r, g, b] = trimmed.toLowerCase()
    return `#${r}${r}${g}${g}${b}${b}`
  }
  return rgbToHex(255, 255, 255)
}

function hexToRgb(value: string) {
  const normalized = normalizeHex(value).slice(1)
  return {
    r: Number.parseInt(normalized.slice(0, 2), 16),
    g: Number.parseInt(normalized.slice(2, 4), 16),
    b: Number.parseInt(normalized.slice(4, 6), 16),
  }
}

function rgbToHex(r: number, g: number, b: number) {
  return `#${[r, g, b].map((item) => clamp(item, 0, 255).toString(16).padStart(2, '0')).join('')}`
}

function hslToHex(hue: number, saturation: number, lightness: number) {
  const s = saturation / 100
  const l = lightness / 100
  const c = (1 - Math.abs(2 * l - 1)) * s
  const x = c * (1 - Math.abs(((hue / 60) % 2) - 1))
  const m = l - c / 2
  let r = 0
  let g = 0
  let b = 0

  if (hue < 60) [r, g, b] = [c, x, 0]
  else if (hue < 120) [r, g, b] = [x, c, 0]
  else if (hue < 180) [r, g, b] = [0, c, x]
  else if (hue < 240) [r, g, b] = [0, x, c]
  else if (hue < 300) [r, g, b] = [x, 0, c]
  else [r, g, b] = [c, 0, x]

  return rgbToHex(Math.round((r + m) * 255), Math.round((g + m) * 255), Math.round((b + m) * 255))
}

function buildPaletteRows() {
  const grayLightness = [98, 94, 88, 76, 63, 49, 36, 24, 14, 7]
  const chromaLightness = [97, 90, 80, 68, 57, 47, 37, 28, 20, 12]
  const hues = [0, 24, 42, 52, 82, 145, 160, 176, 188, 202, 218, 238, 255, 274, 295, 326, 348]
  const labels = new Map<string, string>([
    ['0:0', 'NC'],
    ['1:0', 'B2'],
    ['2:0', 'B3'],
    ['6:0', 'N'],
    ['3:5', 'A'],
    ['3:10', 'P'],
    ['3:13', 'S'],
    ['4:7', 'SU'],
    ['4:10', 'IN'],
    ['4:15', 'ER'],
    ['4:4', 'WA'],
    ['9:5', 'AC'],
    ['9:7', 'PC'],
    ['9:9', 'SUC'],
    ['9:11', 'INC'],
    ['9:12', 'SC'],
    ['9:15', 'ERC'],
  ])

  return grayLightness.map((gray, rowIndex) => {
    const row = [
      ...[0, 1, 2, 3, 4].map((columnIndex) => ({
        value: hslToHex(220, columnIndex * 6, gray - columnIndex),
        label: labels.get(`${rowIndex}:${columnIndex}`),
      })),
      ...hues.map((hue, hueIndex) => ({
        value: hslToHex(hue, Math.max(18, 90 - rowIndex * 2), chromaLightness[rowIndex]),
        label: labels.get(`${rowIndex}:${hueIndex + 5}`),
      })),
    ]
    return row
  })
}

function readableTextColor(value: string) {
  const { r, g, b } = hexToRgb(value)
  const luminance = (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255
  return luminance > 0.58 ? 'var(--gf-color-base-content)' : 'var(--gf-color-base-100)'
}

function channelGradient(channel: 'r' | 'g' | 'b') {
  const current = rgb.value
  const min = { ...current, [channel]: 0 }
  const max = { ...current, [channel]: 255 }
  return `linear-gradient(to right, ${rgbToHex(min.r, min.g, min.b)}, ${rgbToHex(max.r, max.g, max.b)})`
}

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value))
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') closePicker()
}

onMounted(() => {
  window.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div class="flex min-w-0 flex-col items-center">
    <button
      type="button"
      class="group block rounded-md outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 focus-visible:ring-offset-base-100"
      :aria-label="`Pick color for ${activeLabel}`"
      @click="openPicker"
    >
      <span class="grid h-10 w-10 place-items-center rounded-md border border-line text-base font-black transition group-hover:border-primary" :style="{ backgroundColor: modelValue, color: readableTextColor(modelValue) }">
        <span v-if="tokenLabel.endsWith('content')">A</span>
      </span>
    </button>
    <span class="mt-1 block w-16 min-w-0 truncate text-center text-[10px] leading-none text-base-content/55">
      {{ displayLabel }}
    </span>

    <Teleport to="body">
      <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-neutral/50 p-3" @click.self="closePicker">
        <section class="flex max-h-[88vh] w-full max-w-4xl flex-col overflow-hidden rounded-lg border border-line bg-base-100" role="dialog" aria-modal="true">
          <header class="flex flex-wrap items-center justify-between gap-3 border-b border-line px-4 py-2.5">
            <div class="flex min-w-0 items-center gap-3">
              <span class="grid h-10 w-14 shrink-0 place-items-center rounded-md border border-line bg-base-200 text-lg font-black text-base-content">A</span>
              <span class="h-px w-6 shrink-0 bg-line" />
              <h2 class="min-w-0 truncate text-sm text-base-content/55">
                Pick a color for <span class="font-semibold text-base-content">{{ activeLabel }}</span>
              </h2>
            </div>
            <div class="flex items-center gap-2">
              <div class="inline-flex rounded-md border border-line bg-base-100 p-0.5">
                <button
                  type="button"
                  class="inline-flex h-8 items-center gap-1.5 rounded px-2.5 text-sm font-semibold"
                  :class="mode === 'palette' ? 'bg-base-200 text-base-content' : 'text-base-content/55 hover:text-base-content'"
                  @click="mode = 'palette'"
                >
                  <Grid3X3 class="h-4 w-4" /> Palette
                </button>
                <button
                  type="button"
                  class="inline-flex h-8 items-center gap-1.5 rounded px-2.5 text-sm font-semibold"
                  :class="mode === 'picker' ? 'bg-base-200 text-base-content' : 'text-base-content/55 hover:text-base-content'"
                  @click="mode = 'picker'"
                >
                  <SlidersHorizontal class="h-4 w-4" /> Picker
                </button>
              </div>
              <button type="button" class="grid h-8 w-8 place-items-center rounded-md text-icon-muted hover:bg-base-200 hover:text-base-content" aria-label="Close" @click="closePicker">
                <X class="h-5 w-5" />
              </button>
            </div>
          </header>

          <div class="min-h-0 flex-1 overflow-y-auto p-4">
            <div v-if="mode === 'palette'" class="overflow-x-auto">
              <div class="grid w-max gap-1.5" :style="{ gridTemplateRows: `repeat(${paletteRows.length}, minmax(0, 1fr))` }">
                <div v-for="(row, rowIndex) in paletteRows" :key="rowIndex" class="flex gap-1.5">
                  <button
                    v-for="(color, columnIndex) in row"
                    :key="`${rowIndex}-${columnIndex}`"
                    type="button"
                    class="relative grid h-8 w-8 shrink-0 place-items-center rounded-full border border-line text-[9px] font-black transition hover:scale-105"
                    :class="toHex(modelValue) === color.value ? 'ring-2 ring-primary ring-offset-2 ring-offset-base-100' : ''"
                    :style="{ backgroundColor: color.value, color: readableTextColor(color.value) }"
                    :aria-label="color.value"
                    @click="choose(color.value)"
                  >
                    <span v-if="color.label">{{ color.label }}</span>
                  </button>
                </div>
              </div>
            </div>

            <div v-else class="mx-auto max-w-3xl space-y-6 py-3">
              <label v-for="channel in ['r', 'g', 'b']" :key="channel" class="block">
                <span class="mb-1.5 block text-sm font-semibold capitalize text-base-content">{{ channel === 'r' ? 'Red' : channel === 'g' ? 'Green' : 'Blue' }}</span>
                <input
                  class="h-8 w-full cursor-pointer appearance-none rounded-full border border-line bg-base-200 accent-primary"
                  type="range"
                  min="0"
                  max="255"
                  :value="rgb[channel as 'r' | 'g' | 'b']"
                  :style="{ background: channelGradient(channel as 'r' | 'g' | 'b') }"
                  @input="updateChannel(channel as 'r' | 'g' | 'b', Number(($event.target as HTMLInputElement).value))"
                />
              </label>
            </div>
          </div>

          <footer class="flex flex-wrap items-center justify-between gap-3 border-t border-line bg-base-200 px-4 py-2.5">
            <label class="min-w-0 flex-1">
              <span class="mb-1 block text-xs font-semibold text-base-content/55">Color value</span>
              <span class="flex max-w-xl overflow-hidden rounded-md border border-line bg-base-100">
                <span class="grid h-9 w-14 place-items-center border-r border-line text-sm font-semibold text-base-content/75">Hex</span>
                <input
                  class="h-9 min-w-0 flex-1 bg-base-100 px-3 text-sm font-semibold text-base-content outline-none"
                  :value="localHex"
                  @input="updateHex(($event.target as HTMLInputElement).value)"
                />
              </span>
            </label>
            <div class="flex items-center gap-3">
              <span class="rounded-full border border-dashed border-base-content px-2.5 py-1 text-xs font-black text-base-content">AAA</span>
              <span class="grid h-6 w-6 place-items-center rounded-full bg-neutral text-neutral-content">
                <Check class="h-3.5 w-3.5" />
              </span>
              <span class="h-7 w-16 rounded-full border border-line" :style="{ backgroundColor: localHex }" />
            </div>
          </footer>
        </section>
      </div>
    </Teleport>
  </div>
</template>
