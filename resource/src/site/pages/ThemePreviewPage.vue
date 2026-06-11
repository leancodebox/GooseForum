<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  AlertTriangle,
  Check,
  Circle,
  Clipboard,
  Code2,
  Eye,
  FileText,
  MessageSquare,
  Moon,
  PaintBucket,
  RotateCcw,
  Rocket,
  Save,
  Search,
  SlidersHorizontal,
  Sparkles,
  Sun,
} from '@lucide/vue'
import { publishSiteTheme, saveSiteTheme } from '@/runtime/api'
import { useFlashMessages } from '@/runtime/flash-message'
import { applySiteThemeCss, applySiteThemePayload, setTheme, type SiteTheme } from '@/runtime/site-theme'
import {
  cloneSiteThemeTokens,
  createEmptySiteThemeTokens,
  siteThemeTokenKeys,
  type LayoutPayload,
  type SiteThemeConfig,
  type SiteThemeDefinition,
  type SiteThemeTokenKey,
  type ThemePayload,
  type ThemePreviewProps,
} from '@/types/payload'
import ThemeColorPicker from '@/site/components/ThemeColorPicker.vue'
import { themePresets, type SiteThemePreset } from '@/site/theme-presets'

type PreviewMode = 'forum' | 'components' | 'code'

const props = defineProps<{
  props: ThemePreviewProps
  layout: LayoutPayload
}>()

const SITE_MANAGER_PERMISSION = 5
const { push: pushFlash } = useFlashMessages()
const { t } = useI18n()

const colorGroups = [
  {
    key: 'base',
    label: 'Base',
    description: 'page, panels, borders',
    tokens: [
      ['color-base-100', 'Canvas'],
      ['color-base-200', 'Page'],
      ['color-base-300', 'Muted'],
      ['color-base-content', 'Text'],
      ['color-line', 'Line'],
      ['color-icon-muted', 'Icon'],
    ],
  },
  {
    key: 'brand',
    label: 'Brand',
    description: 'primary actions and navigation',
    tokens: [
      ['color-primary', 'Primary'],
      ['color-primary-content', 'On primary'],
      ['color-secondary', 'Secondary'],
      ['color-secondary-content', 'On secondary'],
      ['color-accent', 'Accent'],
      ['color-accent-content', 'On accent'],
      ['color-neutral', 'Neutral'],
      ['color-neutral-content', 'On neutral'],
    ],
  },
  {
    key: 'state',
    label: 'State',
    description: 'feedback and badges',
    tokens: [
      ['color-info', 'Info'],
      ['color-info-content', 'On info'],
      ['color-success', 'Success'],
      ['color-success-content', 'On success'],
      ['color-warning', 'Warning'],
      ['color-warning-content', 'On warning'],
      ['color-error', 'Error'],
      ['color-error-content', 'On error'],
    ],
  },
] as const

const contrastPairs = [
  ['Base text', 'color-base-content', 'color-base-100'],
  ['Primary', 'color-primary-content', 'color-primary'],
  ['Secondary', 'color-secondary-content', 'color-secondary'],
  ['Accent', 'color-accent-content', 'color-accent'],
  ['Info', 'color-info-content', 'color-info'],
  ['Success', 'color-success-content', 'color-success'],
  ['Warning', 'color-warning-content', 'color-warning'],
  ['Error', 'color-error-content', 'color-error'],
] as const

const radiusTokens = [
  ['radius-box', 'Boxes', 'card, modal, alert'],
  ['radius-field', 'Fields', 'button, input, select, tab'],
  ['radius-selector', 'Selectors', 'checkbox, toggle, badge'],
] as const

const radiusOptions = [
  ['0', 0],
  ['sm', 4],
  ['md', 8],
  ['lg', 16],
  ['xl', 24],
] as const

const statusTokens = [
  ['color-info', 'Info'],
  ['color-success', 'Success'],
  ['color-warning', 'Warning'],
  ['color-error', 'Error'],
] as const

const previewModes = [
  ['forum', 'Forum', FileText],
  ['components', 'Components', Circle],
  ['code', 'CSS', Code2],
] as const

const selectedTheme = ref<SiteTheme>('gf-light')
const previewMode = ref<PreviewMode>('forum')
const saving = ref(false)
const publishing = ref(false)
const copying = ref(false)
const message = ref('')
const error = ref('')
const prepublish = reactive<SiteThemeConfig>(configFromPrepublish(props.props.theme))
const savedConfig = ref<SiteThemeConfig>(cloneConfig(props.props.theme))

const activeTheme = computed(() => themeByName(selectedTheme.value))
const previewTitle = computed(() => selectedTheme.value === 'gf-dark' ? 'Dark preview' : 'Light preview')
const canManageSiteTheme = computed(() => props.layout.viewer.adminPermissions.includes(SITE_MANAGER_PERMISSION))
const activeThemeCss = computed(() => buildThemeCss(prepublish))
const dirty = computed(() => themeEditSignature(configFromPrepublish(savedConfig.value)) !== themeEditSignature(prepublish))
const selectedDefaultTheme = computed(() => props.props.defaults.themes.find((theme) => theme.name === selectedTheme.value))
const themeAccentLabel = computed(() => selectedTheme.value === 'gf-dark' ? 'Dark' : 'Light')

const contrastScores = computed(() => contrastPairs.map(([label, foreground, background]) => {
  const ratio = contrastRatio(tokenValue(foreground), tokenValue(background))
  return {
    label,
    ratio,
    grade: ratio >= 7 ? 'AAA' : ratio >= 4.5 ? 'AA' : ratio >= 3 ? 'UI' : 'Low',
  }
}))

watch(
  prepublish,
  () => {
    applySiteThemeCss(activeThemeCss.value)
  },
  { deep: true },
)

onMounted(() => {
  applySiteThemeCss(activeThemeCss.value)
  setTheme(selectedTheme.value)
})

onBeforeUnmount(() => {
  applySiteThemeCss('')
})

function cloneConfig(config: SiteThemeConfig): SiteThemeConfig {
  return {
    version: config.version || 1,
    enabled: Boolean(config.enabled),
    themes: config.themes.map((theme) => ({
      name: theme.name,
      label: theme.label,
      colorScheme: theme.colorScheme,
      tokens: cloneSiteThemeTokens(theme.tokens),
    })),
    prepublish: config.prepublish
      ? {
          enabled: Boolean(config.prepublish.enabled),
          updatedAt: config.prepublish.updatedAt,
          themes: config.prepublish.themes.map((theme) => ({
            name: theme.name,
            label: theme.label,
            colorScheme: theme.colorScheme,
            tokens: cloneSiteThemeTokens(theme.tokens),
          })),
        }
      : undefined,
    publishedAt: config.publishedAt,
  }
}

function configFromPrepublish(config: SiteThemeConfig): SiteThemeConfig {
  const cloned = cloneConfig(config)
  if (cloned.prepublish) {
    cloned.enabled = cloned.prepublish.enabled
    cloned.themes = cloned.prepublish.themes.map((theme) => ({
      name: theme.name,
      label: theme.label,
      colorScheme: theme.colorScheme,
      tokens: cloneSiteThemeTokens(theme.tokens),
    }))
  }
  return cloned
}

function themeByName(name: SiteTheme): SiteThemeDefinition {
  let theme = prepublish.themes.find((item) => item.name === name)
  if (!theme) {
    const fallback = props.props.defaults.themes.find((item) => item.name === name)
    theme = fallback
      ? {
          name: fallback.name,
          label: fallback.label,
          colorScheme: fallback.colorScheme,
          tokens: cloneSiteThemeTokens(fallback.tokens),
        }
      : {
          name,
          label: name === 'gf-dark' ? 'Dark' : 'Light',
          colorScheme: name === 'gf-dark' ? 'dark' : 'light',
          tokens: createEmptySiteThemeTokens(),
        }
    prepublish.themes.push(theme)
  }
  return theme
}

function selectTheme(name: SiteTheme) {
  selectedTheme.value = name
  setTheme(name)
}

function tokenValue(key: SiteThemeTokenKey) {
  return activeTheme.value.tokens[key] || selectedDefaultTheme.value?.tokens[key] || ''
}

function setToken(key: SiteThemeTokenKey, value: string) {
  activeTheme.value.tokens[key] = value
}

function presetLabel(preset: SiteThemePreset) {
  return t(`themePreview.presets.${preset.key}.label`)
}

function presetDescription(preset: SiteThemePreset) {
  return t(`themePreview.presets.${preset.key}.description`)
}

function applyPreset(preset: SiteThemePreset) {
  ;(['gf-light', 'gf-dark'] as const).forEach((name) => {
    const theme = themeByName(name)
    siteThemeTokenKeys.forEach((key) => {
      const value = preset.themes[name][key]
      if (value) theme.tokens[key] = value
    })
  })
  setTheme(selectedTheme.value)
  message.value = t('themePreview.presetApplied', { name: presetLabel(preset) })
  error.value = ''
  pushFlash(message.value, 'success')
}

function fromRange(key: SiteThemeTokenKey, value: number) {
  if (key === 'border') return `${value}px`
  return `${value / 16}rem`
}

function toRange(key: SiteThemeTokenKey, value: string) {
  const numeric = Number.parseFloat(value)
  if (!Number.isFinite(numeric)) return key === 'border' ? 1 : 8
  if (value.endsWith('rem')) return Math.round(numeric * 16)
  return numeric
}

function isRadiusSelected(key: SiteThemeTokenKey, value: number) {
  return toRange(key, tokenValue(key)) === value
}

function radiusPreviewStyle(value: number, selected: boolean) {
  const radius = `${value}px`
  return {
    borderTopLeftRadius: '0',
    borderTopRightRadius: radius,
    borderBottomRightRadius: '0',
    borderBottomLeftRadius: '0',
    borderTopColor: selected ? 'var(--gf-color-primary)' : 'var(--gf-color-base-content)',
    borderRightColor: selected ? 'var(--gf-color-primary)' : 'var(--gf-color-base-content)',
  }
}

function sanitizeThemeName(value: string) {
  return value === 'gf-light' || value === 'gf-dark' ? value : ''
}

function sanitizeThemeValue(value: string) {
  const trimmed = value.trim()
  if (!trimmed || /[{};<>]/.test(trimmed)) return ''
  return trimmed
}

function buildThemeCss(config: SiteThemeConfig) {
  if (!config.enabled) return ''
  return config.themes.map((theme) => {
    const name = sanitizeThemeName(theme.name)
    if (!name) return ''
    const declarations = siteThemeTokenKeys
      .map((key) => {
        const value = sanitizeThemeValue(theme.tokens[key])
        return value ? `--gf-${key}:${value}` : ''
      })
      .filter(Boolean)
      .join(';')
    const colorScheme = theme.colorScheme === 'dark' || theme.colorScheme === 'light' ? `color-scheme:${theme.colorScheme};` : ''
    return `[data-theme="${name}"]{${colorScheme}${declarations}}`
  }).filter(Boolean).join('\n')
}

function buildThemePayload(config: SiteThemeConfig): ThemePayload {
  return {
    enabled: config.enabled,
    href: themeHref(config),
    colors: themeColors(config),
  }
}

function themeHref(config: SiteThemeConfig) {
  if (!config.enabled) return ''
  const version = config.publishedAt || String(config.version || 1)
  return `/site-theme.css?v=${encodeURIComponent(version)}`
}

function themeColors(config: SiteThemeConfig) {
  if (!config.enabled) return {}
  return Object.fromEntries(config.themes.map((theme) => [
    theme.name,
    theme.tokens['color-base-100'] || '',
  ]))
}

function themeEditSignature(config: SiteThemeConfig) {
  return JSON.stringify({
    enabled: Boolean(config.enabled),
    themes: config.themes.map((theme) => ({
      name: theme.name,
      label: theme.label,
      colorScheme: theme.colorScheme,
      tokens: Object.fromEntries(siteThemeTokenKeys.map((key) => [key, theme.tokens[key]])),
    })),
  })
}

async function save() {
  if (saving.value || !canManageSiteTheme.value) return
  saving.value = true
  message.value = ''
  error.value = ''
  try {
    const config = await saveSiteTheme(cloneConfig(prepublish))
    savedConfig.value = cloneConfig(config)
    applyPrepublishConfig(configFromPrepublish(config))
    message.value = '预发布已保存，发布后才会影响全站'
    pushFlash(message.value, 'success')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '主题配置保存失败'
    pushFlash(error.value, 'error')
  } finally {
    saving.value = false
  }
}

async function publish() {
  if (publishing.value || !canManageSiteTheme.value) return
  publishing.value = true
  message.value = ''
  error.value = ''
  try {
    const config = await publishSiteTheme()
    savedConfig.value = cloneConfig(config)
    applyPrepublishConfig(configFromPrepublish(config))
    applySiteThemeCss('')
    applySiteThemePayload(buildThemePayload(config))
    message.value = '预发布已发布'
    pushFlash(message.value, 'success')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '主题配置发布失败'
    pushFlash(error.value, 'error')
  } finally {
    publishing.value = false
  }
}

function resetPrepublish() {
  applyPrepublishConfig(configFromPrepublish(savedConfig.value))
  message.value = '已还原到已保存配置'
  error.value = ''
}

function resetSelectedThemeToDefault() {
  const defaultTheme = selectedDefaultTheme.value
  if (!defaultTheme) return
  const theme = themeByName(selectedTheme.value)
  theme.label = defaultTheme.label
  theme.colorScheme = defaultTheme.colorScheme
  theme.tokens = cloneSiteThemeTokens(defaultTheme.tokens)
  message.value = `${defaultTheme.label} 已恢复内置默认`
  error.value = ''
}

function resetAllThemesToDefault() {
  const defaults = cloneConfig(props.props.defaults)
  prepublish.version = defaults.version
  prepublish.enabled = defaults.enabled
  prepublish.themes.splice(0, prepublish.themes.length, ...defaults.themes)
  prepublish.prepublish = undefined
  prepublish.publishedAt = savedConfig.value.publishedAt
  message.value = '已恢复为内置默认主题，保存预发布后可发布到全站'
  error.value = ''
}

async function copyThemeCss() {
  copying.value = true
  message.value = ''
  error.value = ''
  try {
    await navigator.clipboard.writeText(activeThemeCss.value || '/* custom theme disabled */')
    message.value = 'CSS 已复制'
  } catch {
    error.value = 'CSS 复制失败'
  } finally {
    window.setTimeout(() => {
      copying.value = false
    }, 700)
  }
}

function applyPrepublishConfig(next: SiteThemeConfig) {
  prepublish.version = next.version
  prepublish.enabled = next.enabled
  prepublish.themes.splice(0, prepublish.themes.length, ...next.themes)
  prepublish.prepublish = next.prepublish
  prepublish.publishedAt = next.publishedAt
}

function contrastRatio(foreground: string, background: string) {
  const fg = relativeLuminance(foreground)
  const bg = relativeLuminance(background)
  if (fg === null || bg === null) return 0
  const light = Math.max(fg, bg)
  const dark = Math.min(fg, bg)
  return (light + 0.05) / (dark + 0.05)
}

function relativeLuminance(value: string) {
  const rgb = hexToRgb(value)
  if (!rgb) return null
  const [r, g, b] = [rgb.r, rgb.g, rgb.b].map((channel) => {
    const normalized = channel / 255
    return normalized <= 0.03928 ? normalized / 12.92 : ((normalized + 0.055) / 1.055) ** 2.4
  })
  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

function hexToRgb(value: string) {
  const normalized = value.trim()
  if (!/^#[0-9a-fA-F]{6}$/.test(normalized)) return null
  return {
    r: Number.parseInt(normalized.slice(1, 3), 16),
    g: Number.parseInt(normalized.slice(3, 5), 16),
    b: Number.parseInt(normalized.slice(5, 7), 16),
  }
}
</script>

<template>
  <div class="grid gap-3 lg:h-[calc(100vh-5.5rem)] lg:min-h-0 lg:grid-cols-[minmax(280px,330px)_minmax(0,1fr)] lg:overflow-hidden">
    <section class="overflow-hidden rounded-box border border-line bg-base-100 lg:flex lg:min-h-0 lg:flex-col">
      <div class="sticky top-0 z-10 border-b border-line bg-base-100 p-3">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <PaintBucket class="h-4 w-4 text-primary" />
              <h1 class="truncate text-base font-semibold text-base-content">主题预览设置</h1>
            </div>
            <div class="mt-2 flex flex-wrap items-center gap-2 text-xs font-semibold">
              <span class="rounded-full bg-base-200 px-2 py-1 text-base-content/65">v{{ prepublish.version || 1 }}</span>
              <span class="rounded-full px-2 py-1" :class="prepublish.enabled ? 'bg-success/10 text-success' : 'bg-base-200 text-base-content/55'">{{ prepublish.enabled ? 'Enabled' : 'Disabled' }}</span>
              <span v-if="dirty" class="gf-badge gf-badge-warning py-1">Unsaved</span>
              <span v-if="!canManageSiteTheme" class="gf-badge gf-badge-error py-1">Read only</span>
            </div>
          </div>
          <div class="flex shrink-0 items-center gap-1.5">
            <button type="button" class="gf-button gf-button-sm gf-button-secondary text-xs" @click="resetAllThemesToDefault">
              默认
            </button>
            <button type="button" class="gf-icon-button h-8 w-8 bg-base-200" title="还原到已保存配置" @click="resetPrepublish">
              <RotateCcw class="h-4 w-4" />
            </button>
          </div>
        </div>

        <div class="mt-3 grid grid-cols-[minmax(0,1fr)_auto] gap-2">
          <div class="gf-segmented grid-cols-2">
            <button
              type="button"
              class="gf-segmented-item"
              :class="selectedTheme === 'gf-light' ? 'gf-segmented-item-active' : 'gf-segmented-item-idle'"
              @click="selectTheme('gf-light')"
            >
              <Sun class="h-4 w-4" /> Light
            </button>
            <button
              type="button"
              class="gf-segmented-item"
              :class="selectedTheme === 'gf-dark' ? 'gf-segmented-item-active' : 'gf-segmented-item-idle'"
              @click="selectTheme('gf-dark')"
            >
              <Moon class="h-4 w-4" /> Dark
            </button>
          </div>
          <label class="inline-flex cursor-pointer items-center gap-2 rounded-md bg-base-200 px-2 text-sm font-semibold text-base-content/75">
            <span>启用</span>
            <span class="relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition" :class="prepublish.enabled ? 'bg-primary' : 'bg-base-300'">
              <input v-model="prepublish.enabled" type="checkbox" class="peer sr-only" />
              <span class="absolute left-0.5 h-4 w-4 rounded-full bg-primary-content transition" :class="prepublish.enabled ? 'translate-x-4' : 'translate-x-0 bg-base-100'" />
            </span>
          </label>
        </div>
      </div>

      <div class="min-h-full space-y-4 bg-base-100 p-3 lg:min-h-0 lg:flex-1 lg:overflow-y-auto">
        <section class="space-y-2">
          <div class="flex items-center gap-2">
            <Sparkles class="h-4 w-4 text-icon-muted" />
            <h2 class="text-sm font-semibold text-base-content">{{ t('themePreview.presetsTitle') }}</h2>
            <span class="h-px flex-1 bg-line" />
          </div>
          <div class="grid grid-cols-2 gap-2">
            <button
              v-for="preset in themePresets"
              :key="preset.key"
              type="button"
              class="min-w-0 rounded-md border border-line bg-base-100 p-2 text-left transition hover:border-primary/30 hover:bg-base-200 disabled:cursor-not-allowed disabled:opacity-55"
              :disabled="!canManageSiteTheme"
              @click="applyPreset(preset)"
            >
              <div class="flex items-center justify-between gap-2">
                <span class="truncate text-xs font-bold text-base-content">{{ presetLabel(preset) }}</span>
                <span class="flex shrink-0 -space-x-1">
                  <span
                    v-for="color in preset.swatches"
                    :key="color"
                    class="h-3.5 w-3.5 rounded-full border border-base-100"
                    :style="{ backgroundColor: color }"
                  />
                </span>
              </div>
              <div class="mt-1 truncate text-[11px] text-base-content/45">{{ presetDescription(preset) }}</div>
            </button>
          </div>
        </section>

        <section class="space-y-3">
          <div class="flex items-center gap-2">
            <PaintBucket class="h-4 w-4 text-icon-muted" />
            <h2 class="text-sm font-semibold text-base-content">Change Colors</h2>
            <span class="h-px flex-1 bg-line" />
            <button type="button" class="gf-button gf-button-sm gf-button-secondary h-7 text-xs" @click="resetSelectedThemeToDefault">
              默认
            </button>
          </div>

          <div v-for="group in colorGroups" :key="group.key" class="border-b border-line pb-3 last:border-b-0">
            <div class="flex items-center justify-between gap-3">
              <div>
                <div class="text-xs font-bold uppercase tracking-wide text-base-content/70">{{ group.label }}</div>
                <div class="text-[11px] text-base-content/45">{{ group.description }}</div>
              </div>
              <div class="flex -space-x-1">
                <span
                  v-for="[key] in group.tokens.slice(0, 5)"
                  :key="key"
                  class="h-4 w-4 rounded-full border border-base-100"
                  :style="{ backgroundColor: tokenValue(key) }"
                />
              </div>
            </div>
            <div class="mt-2 grid grid-cols-4 gap-x-2 gap-y-2">
              <ThemeColorPicker
                v-for="[key, label] in group.tokens"
                :key="key"
                :model-value="tokenValue(key)"
                :token-label="label"
                @update:model-value="setToken(key, $event)"
              />
            </div>
          </div>

          <div class="border-t border-line pt-3">
            <div class="mb-2 flex items-center justify-between gap-3">
              <h3 class="text-xs font-bold uppercase tracking-wide text-base-content/70">Contrast</h3>
              <span class="text-[11px] font-semibold text-base-content/45">{{ themeAccentLabel }}</span>
            </div>
            <div class="grid gap-1.5">
              <div v-for="item in contrastScores" :key="item.label" class="grid grid-cols-[minmax(0,1fr)_56px_44px] items-center gap-2 text-xs">
                <span class="truncate text-base-content/65">{{ item.label }}</span>
                <span class="font-mono text-base-content/55">{{ item.ratio ? item.ratio.toFixed(2) : 'n/a' }}</span>
                <span
                  class="rounded px-1.5 py-0.5 text-center font-bold"
                  :class="item.grade === 'Low' ? 'bg-error/10 text-error' : item.grade === 'UI' ? 'bg-warning/10 text-warning' : 'bg-success/10 text-success'"
                >
                  {{ item.grade }}
                </span>
              </div>
            </div>
          </div>
        </section>

        <section class="border-t border-line pt-4">
          <div class="mb-2.5 flex items-center gap-2">
            <SlidersHorizontal class="h-4 w-4 text-icon-muted" />
            <h2 class="text-sm font-semibold text-base-content">Radius</h2>
            <span class="h-px flex-1 bg-line" />
          </div>
          <div class="space-y-3">
            <div v-for="[key, label, hint] in radiusTokens" :key="key">
              <div class="mb-1.5 flex items-end justify-between gap-3">
                <div>
                  <div class="text-xs font-semibold text-base-content/75">{{ label }}</div>
                  <div class="text-[11px] text-base-content/45">{{ hint }}</div>
                </div>
                <span class="font-mono text-xs text-base-content/55">{{ tokenValue(key) }}</span>
              </div>
              <div class="grid grid-cols-5 gap-1.5">
                <button
                  v-for="[, value] in radiusOptions"
                  :key="`${key}-${value}`"
                  type="button"
                  class="grid h-10 place-items-center rounded-md border transition"
                  :class="isRadiusSelected(key, value) ? 'border-primary bg-info/10' : 'border-line bg-base-200 hover:bg-base-300'"
                  :aria-label="`${label} ${value}px`"
                  @click="setToken(key, fromRange(key, value))"
                >
                  <span
                    class="block h-5 w-5 border-r-2 border-t-2"
                    :class="isRadiusSelected(key, value) ? 'opacity-100' : 'opacity-35'"
                    :style="radiusPreviewStyle(value, isRadiusSelected(key, value))"
                  />
                </button>
              </div>
            </div>
          </div>
        </section>

        <section class="border-t border-line pt-4">
          <div class="mb-2.5 flex items-center gap-2">
            <SlidersHorizontal class="h-4 w-4 text-icon-muted" />
            <h2 class="text-sm font-semibold text-base-content">Effects & Options</h2>
            <span class="h-px flex-1 bg-line" />
          </div>
          <div class="grid gap-2">
            <label class="flex items-center justify-between rounded-md bg-base-200 px-2.5 py-1.5 text-sm font-medium text-base-content/75">
              <span>Depth Effect</span>
              <input type="checkbox" class="h-4 w-4 accent-primary" :checked="tokenValue('depth') === '1'" @change="setToken('depth', ($event.target as HTMLInputElement).checked ? '1' : '0')" />
            </label>
            <label class="flex items-center justify-between rounded-md bg-base-200 px-2.5 py-1.5 text-sm font-medium text-base-content/75">
              <span>Noise Effect</span>
              <input type="checkbox" class="h-4 w-4 accent-primary" :checked="tokenValue('noise') === '1'" @change="setToken('noise', ($event.target as HTMLInputElement).checked ? '1' : '0')" />
            </label>
            <label class="flex items-center justify-between rounded-md bg-base-200 px-2.5 py-1.5 text-sm font-medium text-base-content/75">
              <span>Dark color scheme</span>
              <input type="checkbox" class="h-4 w-4 accent-primary" :checked="activeTheme.colorScheme === 'dark'" @change="activeTheme.colorScheme = ($event.target as HTMLInputElement).checked ? 'dark' : 'light'" />
            </label>
          </div>
        </section>
      </div>
    </section>

    <section class="overflow-hidden rounded-box border border-line bg-base-100 lg:flex lg:min-h-0 lg:flex-col">
      <div class="sticky top-0 z-10 flex flex-wrap items-center justify-between gap-2 border-b border-line bg-base-100 p-3">
        <div class="flex items-center gap-2">
          <Eye class="h-4 w-4 text-icon-muted" />
          <div>
            <h2 class="text-sm font-semibold text-base-content">{{ previewTitle }}</h2>
            <p class="mt-0.5 text-xs text-base-content/55">{{ props.layout.site.name || 'GooseForum' }}</p>
          </div>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <div class="gf-segmented grid-cols-3">
            <button
              v-for="[mode, label, Icon] in previewModes"
              :key="mode"
              type="button"
              class="gf-segmented-item text-xs sm:text-sm"
              :class="previewMode === mode ? 'gf-segmented-item-active' : 'gf-segmented-item-idle'"
              @click="previewMode = mode"
            >
              <component :is="Icon" class="h-4 w-4" /> {{ label }}
            </button>
          </div>
          <button
            type="button"
            class="gf-button gf-button-sm gf-button-secondary w-32 whitespace-nowrap disabled:bg-base-100 disabled:text-base-content/75 disabled:opacity-100"
            :class="saving ? 'cursor-wait bg-base-200 ring-2 ring-primary/20' : ''"
            :aria-busy="saving"
            :disabled="!canManageSiteTheme"
            @click="save"
          >
            <Save class="h-4 w-4" :class="saving ? 'animate-pulse text-primary' : ''" /> 保存预发布
          </button>
          <button
            type="button"
            class="gf-button gf-button-sm gf-button-primary w-24 whitespace-nowrap disabled:bg-primary disabled:text-primary-content disabled:opacity-100"
            :class="publishing ? 'cursor-wait brightness-95 ring-2 ring-primary/25' : ''"
            :aria-busy="publishing"
            :disabled="!canManageSiteTheme"
            @click="publish"
          >
            <Rocket class="h-4 w-4" :class="publishing ? 'animate-pulse' : ''" /> 发布
          </button>
        </div>
      </div>

      <div class="min-h-full bg-base-200/60 p-3 lg:min-h-0 lg:flex-1 lg:overflow-y-auto">
        <div v-if="previewMode === 'forum'" class="grid gap-3 2xl:grid-cols-[minmax(0,1fr)_248px]">
          <div class="space-y-3">
            <section class="gf-card overflow-hidden">
              <header class="flex flex-col gap-2 border-b border-line bg-base-100 px-3 py-2.5 sm:flex-row sm:items-center sm:justify-between">
                <div class="flex min-w-0 items-center gap-2 overflow-x-auto">
                  <button class="gf-tab gf-tab-active">最新</button>
                  <button class="gf-tab gf-tab-idle">热门</button>
                  <button class="gf-tab gf-tab-idle">精华</button>
                </div>
                <button class="gf-button gf-button-sm gf-button-primary">
                  <Rocket class="h-4 w-4" /> 发布主题
                </button>
              </header>
              <div class="hidden grid-cols-[minmax(0,1fr)_88px_56px_56px_64px] gap-2 border-b border-line bg-base-200/60 px-3 py-2 text-[11px] font-bold uppercase text-base-content/75 xl:grid">
                <div>Topic</div>
                <div class="text-center">Users</div>
                <div class="text-center">Replies</div>
                <div class="text-center">Views</div>
                <div class="text-right">Activity</div>
              </div>
              <div class="divide-y divide-line bg-base-100">
                <article v-for="index in 3" :key="index" class="grid gap-2 px-3 py-2.5 transition hover:bg-base-200/70 xl:grid-cols-[minmax(0,1fr)_88px_56px_56px_64px] xl:items-center">
                  <div class="min-w-0">
                    <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                      <a href="#" class="min-w-0 truncate text-[15px] font-semibold leading-6 text-base-content hover:text-primary sm:text-base">
                        {{ index === 1 ? '主题系统重构讨论：颜色、圆角和组件状态' : index === 2 ? 'Markdown 正文在深色模式下的可读性' : '新用户引导和消息通知的视觉检查' }}
                      </a>
                      <span class="gf-topic-chip">
                        <span class="h-1.5 w-1.5 rounded-full bg-primary" /> design
                      </span>
                      <span v-if="index === 1" class="inline-flex h-6 items-center gap-1 text-[11px] font-semibold text-warning">
                        <AlertTriangle class="h-3 w-3" /> hot
                      </span>
                    </div>
                    <p class="mt-1 min-h-5 truncate text-[13px] leading-5 text-base-content/55">观察弱文本、分隔线、标签和悬停背景在当前主题下的层级。</p>
                  </div>
                  <div class="hidden justify-center xl:flex">
                    <div class="flex h-7 min-w-7 -space-x-2.5">
                      <span class="h-7 w-7 rounded-full bg-primary ring-2 ring-base-100" />
                      <span class="h-7 w-7 rounded-full bg-accent ring-2 ring-base-100" />
                      <span class="h-7 w-7 rounded-full bg-neutral ring-2 ring-base-100" />
                    </div>
                  </div>
                  <div class="hidden text-center text-sm font-semibold tabular-nums text-base-content/75 xl:block">{{ index * 12 }}</div>
                  <div class="hidden text-center text-sm tabular-nums text-base-content/55 xl:block">{{ index * 480 }}</div>
                  <div class="hidden text-right text-[13px] font-medium tabular-nums text-base-content/55 xl:block">{{ index }}h</div>
                </article>
              </div>
            </section>

            <article class="gf-card p-3">
              <div class="flex flex-wrap items-center gap-2">
                <span class="gf-badge gf-badge-info">Preview</span>
                <span class="gf-badge gf-badge-muted">Theme</span>
              </div>
              <h3 class="mt-3 text-xl font-semibold leading-tight text-base-content">一篇主题帖的标题</h3>
              <div class="gf-prose gf-prose-article mt-2 text-sm">
                <p>这里模拟正文、链接、引用和代码块。主题变量应该让内容在明暗模式下都清晰，尤其是正文、边框和弱文本。</p>
                <blockquote>颜色层级应该安静，但不能含糊。</blockquote>
                <pre><code>const theme = 'gf-{{ selectedTheme === 'gf-dark' ? 'dark' : 'light' }}'</code></pre>
              </div>
            </article>
          </div>

          <aside class="grid gap-3 md:grid-cols-3 2xl:block 2xl:space-y-3">
            <section class="gf-panel p-3">
              <h3 class="text-sm font-semibold text-base-content">Search</h3>
              <label class="mt-3 flex h-9 items-center gap-2 rounded-md border border-line bg-base-200 px-3 text-sm text-base-content/55 transition focus-within:border-primary focus-within:bg-base-100 focus-within:ring-4 focus-within:ring-primary/20">
                <Search class="h-4 w-4" />
                <input class="min-w-0 flex-1 bg-transparent text-base-content outline-none" value="theme preview" />
              </label>
            </section>
            <section class="gf-panel p-3">
              <h3 class="text-sm font-semibold text-base-content">Messages</h3>
              <div class="mt-3 space-y-3">
                <div class="rounded-box bg-base-300 px-3 py-2 text-sm text-base-content">这个背景还舒服吗？</div>
                <div class="rounded-box bg-primary px-3 py-2 text-sm text-primary-content">对比度需要稳。</div>
              </div>
            </section>
            <section class="gf-panel p-3">
              <h3 class="text-sm font-semibold text-base-content">Status</h3>
              <div class="mt-3 grid gap-2">
                <div class="rounded-md bg-info/10 px-3 py-1.5 text-sm font-medium text-primary">Info notification</div>
                <div class="rounded-md bg-success/10 px-3 py-1.5 text-sm font-medium text-success">Success state</div>
                <div class="rounded-md bg-warning/10 px-3 py-1.5 text-sm font-medium text-warning">Warning state</div>
                <div class="rounded-md bg-error/10 px-3 py-1.5 text-sm font-medium text-error">Error state</div>
              </div>
            </section>
          </aside>
        </div>

        <div v-else-if="previewMode === 'components'" class="grid gap-3 xl:grid-cols-[minmax(0,1fr)_320px]">
          <section class="gf-panel p-4">
            <div class="grid gap-3 md:grid-cols-2">
              <div class="rounded-box border border-line bg-base-100 p-4 shadow-[0_10px_24px_-20px_rgba(0,0,0,calc(var(--gf-depth)*0.35))]">
                <div class="h-28 rounded-box bg-base-300" />
                <div class="mt-3 flex items-start justify-between gap-3">
                  <div>
                    <h3 class="font-semibold text-base-content">Card title</h3>
                    <p class="mt-1 text-sm text-base-content/55">Base, line, radius and shadow.</p>
                  </div>
                  <span class="rounded-selector bg-accent px-2 py-1 text-xs font-bold text-accent-content">NEW</span>
                </div>
                <div class="mt-4 flex flex-wrap gap-2">
                  <button class="gf-button gf-button-sm gf-button-primary">Primary</button>
                  <button class="gf-button gf-button-sm gf-button-secondary">Secondary</button>
                </div>
              </div>

              <div class="rounded-box border border-line bg-base-100 p-4">
                <h3 class="font-semibold text-base-content">Form states</h3>
                <label class="mt-3 block text-xs font-medium text-base-content/55">Input</label>
                <input class="gf-input mt-1 bg-base-200" value="GooseForum" />
                <label class="mt-3 block text-xs font-medium text-base-content/55">Textarea</label>
                <textarea class="gf-textarea mt-1 min-h-20 resize-none bg-base-200">主题变量覆盖输入、焦点和正文。</textarea>
                <div class="mt-3 flex items-center justify-between rounded-selector bg-base-200 px-3 py-2">
                  <span class="text-sm font-medium text-base-content/75">Selector</span>
                  <span class="h-5 w-9 rounded-full bg-primary p-0.5"><span class="block h-4 w-4 translate-x-4 rounded-full bg-primary-content" /></span>
                </div>
              </div>
            </div>

            <div class="mt-3 grid gap-3 md:grid-cols-4">
              <div v-for="[key, label] in statusTokens" :key="key" class="rounded-box border border-line bg-base-100 p-3">
                <div class="h-2 rounded-full" :style="{ backgroundColor: tokenValue(key) }" />
                <div class="mt-3 text-sm font-semibold text-base-content">{{ label }}</div>
                <div class="mt-1 truncate font-mono text-xs text-base-content/55">{{ tokenValue(key) }}</div>
              </div>
            </div>
          </section>

          <aside class="gf-panel p-4">
            <h3 class="text-sm font-semibold text-base-content">Admin table</h3>
            <div class="mt-3 overflow-hidden rounded-md border border-line">
              <div class="grid grid-cols-[1fr_72px_80px] border-b border-line bg-base-200 px-3 py-2 text-xs font-bold uppercase text-base-content/65">
                <div>Name</div>
                <div>Status</div>
                <div class="text-right">Action</div>
              </div>
              <div v-for="item in ['Users', 'Posts', 'Badges']" :key="item" class="grid grid-cols-[1fr_72px_80px] items-center border-b border-line px-3 py-2 text-sm last:border-b-0">
                <div class="font-medium text-base-content">{{ item }}</div>
                <div><span class="gf-badge gf-badge-success">OK</span></div>
                <div class="text-right"><button class="rounded-md px-2 py-1 text-xs font-semibold text-primary hover:bg-info/10">Edit</button></div>
              </div>
            </div>
            <div class="mt-4 rounded-md border border-line bg-base-200 p-3">
              <div class="mb-2 flex items-center gap-2 text-xs font-bold uppercase text-base-content/55">
                <MessageSquare class="h-3.5 w-3.5" /> Toast
              </div>
              <p class="text-sm text-base-content/75">There are 9 new messages.</p>
            </div>
          </aside>
        </div>

        <div v-else class="gf-panel">
          <header class="flex items-center justify-between gap-3 border-b border-line px-4 py-3">
            <div class="flex items-center gap-2">
              <Code2 class="h-4 w-4 text-icon-muted" />
              <h3 class="text-sm font-semibold text-base-content">Generated CSS</h3>
            </div>
            <button type="button" class="gf-button gf-button-sm gf-button-secondary" @click="copyThemeCss">
              <Clipboard class="h-4 w-4" /> {{ copying ? 'Copied' : 'Copy' }}
            </button>
          </header>
          <pre class="max-h-[calc(100vh-14rem)] overflow-auto p-4 text-xs leading-6 text-base-content"><code>{{ activeThemeCss || '/* custom theme disabled */' }}</code></pre>
        </div>

        <div class="mt-3 flex flex-wrap items-center gap-3">
          <p v-if="message" class="inline-flex items-center gap-1.5 text-sm font-medium text-success"><Check class="h-4 w-4" />{{ message }}</p>
          <p v-if="error" class="inline-flex items-center gap-1.5 text-sm font-medium text-error"><AlertTriangle class="h-4 w-4" />{{ error }}</p>
        </div>
      </div>
    </section>
  </div>
</template>
