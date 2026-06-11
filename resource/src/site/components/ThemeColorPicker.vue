<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Box, Check, ChevronDown, Database, Grid3X3, Mountain, SlidersHorizontal, X } from '@lucide/vue'

type PickerMode = 'palette' | 'picker'
type ColorFormat = 'oklch' | 'hsl' | 'rgb' | 'hex'

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
const colorFormat = ref<ColorFormat>('hex')
const formatOpen = ref(false)

const rgb = computed(() => hexToRgb(localHex.value))
const paletteRows = computed(() => buildPaletteRows())
const activeLabel = computed(() => props.tokenLabel.replace(/-/g, ' '))
const displayLabel = computed(() => props.tokenLabel.replace(/-content$/, ' content'))
const formatOptions = [
  ['oklch', 'OKLCH', Mountain],
  ['hsl', 'HSL', Database],
  ['rgb', 'RGB', Box],
  ['hex', 'Hex', Box],
] as const
const activeFormatLabel = computed(() => formatOptions.find(([value]) => value === colorFormat.value)?.[1] || 'Hex')
const formattedColorValue = computed(() => formatColorValue(localHex.value, colorFormat.value))

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
  formatOpen.value = false
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

function updateColorValue(value: string) {
  if (colorFormat.value === 'hex') {
    updateHex(value)
    return
  }
  const parsed = parseFormattedColor(value, colorFormat.value)
  if (parsed) updateHex(parsed)
}

function selectFormat(format: ColorFormat) {
  colorFormat.value = format
  formatOpen.value = false
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

function rgbToHsl({ r, g, b }: { r: number, g: number, b: number }) {
  const red = r / 255
  const green = g / 255
  const blue = b / 255
  const max = Math.max(red, green, blue)
  const min = Math.min(red, green, blue)
  const delta = max - min
  let hue = 0
  if (delta !== 0) {
    if (max === red) hue = ((green - blue) / delta) % 6
    else if (max === green) hue = (blue - red) / delta + 2
    else hue = (red - green) / delta + 4
    hue *= 60
    if (hue < 0) hue += 360
  }
  const lightness = (max + min) / 2
  const saturation = delta === 0 ? 0 : delta / (1 - Math.abs(2 * lightness - 1))
  return {
    h: Math.round(hue),
    s: Math.round(saturation * 100),
    l: Math.round(lightness * 100),
  }
}

function rgbToOklch({ r, g, b }: { r: number, g: number, b: number }) {
  const toLinear = (channel: number) => {
    const normalized = channel / 255
    return normalized <= 0.04045 ? normalized / 12.92 : ((normalized + 0.055) / 1.055) ** 2.4
  }
  const red = toLinear(r)
  const green = toLinear(g)
  const blue = toLinear(b)
  const l = Math.cbrt(0.4122214708 * red + 0.5363325363 * green + 0.0514459929 * blue)
  const m = Math.cbrt(0.2119034982 * red + 0.6806995451 * green + 0.1073969566 * blue)
  const s = Math.cbrt(0.0883024619 * red + 0.2817188376 * green + 0.6299787005 * blue)
  const okL = 0.2104542553 * l + 0.7936177850 * m - 0.0040720468 * s
  const a = 1.9779984951 * l - 2.4285922050 * m + 0.4505937099 * s
  const b2 = 0.0259040371 * l + 0.7827717662 * m - 0.8086757660 * s
  const chroma = Math.sqrt(a * a + b2 * b2)
  const hue = (Math.atan2(b2, a) * 180 / Math.PI + 360) % 360
  return {
    l: Math.round(okL * 1000) / 10,
    c: Math.round(chroma * 1000) / 1000,
    h: Math.round(hue * 10) / 10,
  }
}

function formatColorValue(value: string, format: ColorFormat) {
  const current = hexToRgb(value)
  if (format === 'hex') return normalizeHex(value)
  if (format === 'rgb') return `rgb(${current.r} ${current.g} ${current.b})`
  if (format === 'hsl') {
    const hsl = rgbToHsl(current)
    return `hsl(${hsl.h} ${hsl.s}% ${hsl.l}%)`
  }
  const oklch = rgbToOklch(current)
  return `oklch(${oklch.l}% ${oklch.c} ${oklch.h})`
}

function parseFormattedColor(value: string, format: ColorFormat) {
  if (format === 'hex') return normalizeHex(value)
  const numbers = value.match(/-?\d*\.?\d+/g)?.map(Number) || []
  if (format === 'rgb' && numbers.length >= 3) {
    return rgbToHex(numbers[0], numbers[1], numbers[2])
  }
  if (format === 'hsl' && numbers.length >= 3) {
    return hslToHex(((numbers[0] % 360) + 360) % 360, clamp(numbers[1], 0, 100), clamp(numbers[2], 0, 100))
  }
  return ''
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
  if (event.key === 'Escape') {
    if (formatOpen.value) {
      formatOpen.value = false
      return
    }
    closePicker()
  }
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
      <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-neutral/45 p-3" @click.self="closePicker">
        <section class="gf-menu-surface flex max-h-[84vh] w-full max-w-[736px] flex-col overflow-hidden" role="dialog" aria-modal="true">
          <header class="flex flex-wrap items-center justify-between gap-2 border-b border-line px-3 py-2">
            <div class="flex min-w-0 items-center gap-2.5">
              <span class="grid h-8 w-12 shrink-0 place-items-center rounded-md border border-line bg-base-200 text-base font-black text-base-content">A</span>
              <span class="h-px w-5 shrink-0 bg-line" />
              <h2 class="min-w-0 truncate text-xs text-base-content/55">
                Pick a color for <span class="font-semibold text-base-content">{{ activeLabel }}</span>
              </h2>
            </div>
            <div class="flex items-center gap-2">
              <div class="gf-segmented grid-cols-2">
                <button
                  type="button"
                  class="gf-segmented-item h-6 min-w-20 text-xs leading-none"
                  :class="mode === 'palette' ? 'gf-segmented-item-active text-base-content' : 'gf-segmented-item-idle'"
                  @click="mode = 'palette'"
                >
                  <Grid3X3 class="h-3.5 w-3.5" /> Palette
                </button>
                <button
                  type="button"
                  class="gf-segmented-item h-6 min-w-20 text-xs leading-none"
                  :class="mode === 'picker' ? 'gf-segmented-item-active text-base-content' : 'gf-segmented-item-idle'"
                  @click="mode = 'picker'"
                >
                  <SlidersHorizontal class="h-3.5 w-3.5" /> Picker
                </button>
              </div>
              <button type="button" class="gf-icon-button h-7 w-7" aria-label="Close" @click="closePicker">
                <X class="h-4 w-4" />
              </button>
            </div>
          </header>

          <div class="min-h-0 flex-1 overflow-y-auto p-3">
            <div v-if="mode === 'palette'" class="overflow-x-auto">
              <div class="grid w-max gap-1" :style="{ gridTemplateRows: `repeat(${paletteRows.length}, minmax(0, 1fr))` }">
                <div v-for="(row, rowIndex) in paletteRows" :key="rowIndex" class="flex gap-1">
                  <button
                    v-for="(color, columnIndex) in row"
                    :key="`${rowIndex}-${columnIndex}`"
                    type="button"
                    class="relative grid h-7 w-7 shrink-0 place-items-center rounded-full border border-line/75 text-[8px] font-black transition hover:scale-105"
                    :class="toHex(modelValue) === color.value ? 'ring-2 ring-primary ring-offset-1 ring-offset-base-100' : ''"
                    :style="{ backgroundColor: color.value, color: readableTextColor(color.value) }"
                    :aria-label="color.value"
                    @click="choose(color.value)"
                  >
                    <span v-if="color.label">{{ color.label }}</span>
                  </button>
                </div>
              </div>
            </div>

            <div v-else class="mx-auto max-w-2xl space-y-4 py-2">
              <label v-for="channel in ['r', 'g', 'b']" :key="channel" class="block">
                <span class="mb-1 block text-xs font-semibold capitalize text-base-content">{{ channel === 'r' ? 'Red' : channel === 'g' ? 'Green' : 'Blue' }}</span>
                <input
                  class="h-6 w-full cursor-pointer appearance-none rounded-full border border-line bg-base-200 accent-primary"
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

          <footer class="grid items-end gap-2 border-t border-line bg-base-200 px-3 py-2 md:grid-cols-[minmax(260px,1fr)_auto]">
            <label class="relative min-w-0">
              <span class="mb-1 block text-[11px] font-semibold text-base-content/55">Color value</span>
              <span class="gf-input flex h-8 w-full overflow-hidden p-0">
                <span class="shrink-0">
                  <button
                    type="button"
                    class="gf-button h-full w-24 rounded-none border-r border-line px-2 text-xs text-base-content/75 hover:bg-base-200 hover:text-base-content"
                    @click="formatOpen = !formatOpen"
                  >
                    {{ activeFormatLabel }}
                    <ChevronDown class="h-3.5 w-3.5" />
                  </button>
                </span>
                <input
                  class="h-full min-w-0 flex-1 bg-base-100 px-2.5 text-xs font-semibold text-base-content outline-none"
                  :value="formattedColorValue"
                  @change="updateColorValue(($event.target as HTMLInputElement).value)"
                />
              </span>
              <div v-if="formatOpen" class="gf-menu-surface absolute bottom-9 left-0 z-20 w-36 overflow-hidden p-1">
                <div class="px-2.5 py-1.5 text-[11px] font-bold text-base-content/45">Convert format</div>
                <button
                  v-for="[format, label, Icon] in formatOptions"
                  :key="format"
                  type="button"
                  class="flex h-8 w-full items-center gap-2 rounded px-2.5 text-xs font-semibold"
                  :class="colorFormat === format ? 'bg-neutral text-neutral-content' : 'text-base-content hover:bg-base-200'"
                  @click="selectFormat(format)"
                >
                  <component :is="Icon" class="h-3.5 w-3.5" /> {{ label }}
                </button>
              </div>
            </label>
            <div class="flex h-8 items-center justify-end gap-2.5">
              <span class="inline-flex h-7 items-center rounded-full border border-dashed border-base-content px-2.5 text-[11px] font-black text-base-content">AAA</span>
              <span class="grid h-7 w-7 place-items-center rounded-full bg-neutral text-neutral-content">
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
