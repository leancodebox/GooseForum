<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  CalendarDays,
  Camera,
  Check,
  Image,
  KeyRound,
  Link as LinkIcon,
  Loader2,
  Mail,
  Pencil,
  Shield,
  Sparkles,
  UserRound,
} from '@lucide/vue'
import {
  changePassword,
  getOAuthBindings,
  resendActivationEmail,
  saveUserEmail,
  saveUserInfo,
  saveUserName,
  saveUserProfileCover,
  unbindOAuth,
  type OAuthBindingsPayload,
} from '@/runtime/api'
import { formatDate, formatNumber } from '@/runtime/format'
import { useFlashMessages, type FlashMessageType } from '@/runtime/flash-message'
import { useAvatarCropUpload } from '@/site/composables/useAvatarCropUpload'
import PageHeader from '@/site/components/PageHeader.vue'
import SectionHeader from '@/site/components/SectionHeader.vue'
import UserAvatar from '@/site/components/UserAvatar.vue'
import { socialIcons, socialLabels } from '@/site/utils/social-icons'
import type { LayoutPayload, SettingsPageProps } from '@/types/payload'
import { useI18n } from 'vue-i18n'

const page = defineProps<{
  layout: LayoutPayload
  props: SettingsPageProps
}>()

const { t } = useI18n()
const tabKeys = ['profile', 'account', 'privacy', 'binding'] as const
type TabKey = (typeof tabKeys)[number]

const activeTab = ref<TabKey>('profile')
const status = ref('')
const error = ref('')
const savingProfile = ref(false)
const savingUsername = ref(false)
const savingEmail = ref(false)
const sendingActivationEmail = ref(false)
const savingPassword = ref(false)
const loadingBindings = ref(false)
const bindingAction = ref('')
const editingUsername = ref(false)
const editingEmail = ref(false)
const editingCover = ref(false)
const savingCover = ref(false)
const coverUrl = ref(page.props.user.profileCoverUrl || '')
const coverDraft = ref(page.props.user.profileCoverUrl || '')
const bindings = ref<OAuthBindingsPayload>({})
const { push: pushFlash } = useFlashMessages()
const {
  uploadingAvatar,
  avatarInput,
  cropperImage,
  avatarUrl,
  cropModalOpen,
  cropImageUrl,
  cropPreviewUrl,
  cropError,
  chooseAvatar,
  handleAvatarChange,
  closeCropModal,
  uploadCroppedAvatar,
} = useAvatarCropUpload({
  initialAvatarUrl: page.props.user.avatarUrl,
  onStatus: showStatus,
  onError: showError,
})

const socialKeys = ['github', 'twitter', 'linkedIn', 'weibo', 'bilibili', 'zhihu'] as const

const profileForm = reactive({
  nickname: page.props.user.nickname || '',
  bio: page.props.user.bio || '',
  signature: page.props.user.signature || '',
  websiteName: page.props.user.websiteName || '',
  website: page.props.user.website || '',
  externalInformation: buildExternalInfo(),
})

const usernameForm = reactive({
  username: page.props.user.username,
})

const emailForm = reactive({
  email: page.props.user.email,
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const privacy = reactive({
  showArticles: true,
  showFollowing: true,
  emailNotifications: true,
})

const displayName = computed(() => profileForm.nickname || usernameForm.username)
const profileBioText = computed(() => profileForm.bio || profileForm.signature || t('user.emptyBio'))
const profileCoverStyle = computed(() => {
  const activeCoverUrl = coverUrl.value.trim()
  const defaultCover = 'linear-gradient(135deg, var(--gf-color-base-200) 0%, var(--gf-color-info-content) 52%, var(--gf-color-base-200) 100%)'
  if (!activeCoverUrl) {
    return {
      backgroundImage: defaultCover,
    }
  }
  return {
    backgroundImage: `url(${JSON.stringify(activeCoverUrl)}), ${defaultCover}`,
  }
})
const hasStatus = computed(() => Boolean(status.value || error.value))
const statsItems = computed(() => [
  { label: t('user.stats.topics'), value: page.props.stats.articleCount },
  { label: t('user.stats.replies'), value: page.props.stats.replyCount },
  { label: t('user.stats.likesReceived'), value: page.props.stats.likeReceivedCount },
  { label: t('user.stats.likesGiven'), value: page.props.stats.likeGivenCount },
  { label: t('user.stats.followers'), value: page.props.stats.followerCount },
  { label: t('user.stats.following'), value: page.props.stats.followingCount },
  { label: t('user.stats.bookmarks'), value: page.props.stats.collectionCount },
])
const socialItems = computed(() => socialKeys.map((key) => ({
  key,
  label: socialLabels[key],
  icon: socialIcons[key],
})))
const providers = computed(() => [
  { key: 'github', label: 'GitHub', supported: true },
  { key: 'google', label: 'Google', supported: false },
])

const easterEggMessages: Array<{ type: FlashMessageType; message: string }> = [
  { type: 'success', message: t('settings.easterEgg.success') },
  { type: 'info', message: t('settings.easterEgg.info') },
  { type: 'warning', message: t('settings.easterEgg.warning') },
  { type: 'error', message: t('settings.easterEgg.error') },
]

watch(
  () => page.props.user.id,
  () => {
    avatarUrl.value = page.props.user.avatarUrl
    usernameForm.username = page.props.user.username
    emailForm.email = page.props.user.email
    profileForm.nickname = page.props.user.nickname || ''
    profileForm.bio = page.props.user.bio || ''
    profileForm.signature = page.props.user.signature || ''
    profileForm.websiteName = page.props.user.websiteName || ''
    profileForm.website = page.props.user.website || ''
    profileForm.externalInformation = buildExternalInfo()
    coverUrl.value = page.props.user.profileCoverUrl || ''
    coverDraft.value = page.props.user.profileCoverUrl || ''
    editingCover.value = false
  },
)

onMounted(() => {
  const urlTab = new URL(window.location.href).searchParams.get('tab')
  if (tabKeys.includes(urlTab as TabKey)) activeTab.value = urlTab as TabKey

  const savedPrivacy = localStorage.getItem('goose-privacy-settings')
  if (savedPrivacy) {
    Object.assign(privacy, JSON.parse(savedPrivacy))
  }
  void loadBindings()
})

function buildExternalInfo() {
  const info: Record<string, { link?: string }> = {}
  for (const key of socialKeys) {
    info[key] = { link: page.props.user.externalInformation?.[key]?.link || '' }
  }
  return info
}

function setActiveTab(key: TabKey) {
  activeTab.value = key
  const url = new URL(window.location.href)
  if (key === 'profile') url.searchParams.delete('tab')
  else url.searchParams.set('tab', key)
  history.replaceState(history.state, '', url)
}

function settingsTabLabel(key: string, fallback?: string) {
  if (key === 'profile') return t('settings.tabs.profile')
  if (key === 'account') return t('settings.tabs.account')
  if (key === 'privacy') return t('settings.tabs.privacy')
  if (key === 'binding') return t('settings.tabs.binding')
  return fallback || key
}

function triggerAvatarFlash() {
  const item = easterEggMessages[Math.floor(Math.random() * easterEggMessages.length)]
  pushFlash(item.message, item.type)
}

function showStatus(message: string) {
  error.value = ''
  status.value = message
  window.setTimeout(() => {
    if (status.value === message) status.value = ''
  }, 3000)
}

function showError(message: string) {
  status.value = ''
  error.value = message
}

async function saveProfile() {
  savingProfile.value = true
  try {
    await saveUserInfo({ ...profileForm })
    showStatus(t('settings.status.profileSaved'))
  } catch (err) {
    showError(err instanceof Error ? err.message : t('api.profileSaveFailed'))
  } finally {
    savingProfile.value = false
  }
}

function toggleCoverEditor() {
  coverDraft.value = coverUrl.value
  editingCover.value = !editingCover.value
}

function cancelCoverEditor() {
  coverDraft.value = coverUrl.value
  editingCover.value = false
}

async function saveCover() {
  savingCover.value = true
  try {
    await saveUserProfileCover(coverDraft.value)
    coverUrl.value = coverDraft.value.trim()
    editingCover.value = false
    showStatus(t('user.coverSaved'))
  } catch (err) {
    showError(err instanceof Error ? err.message : t('api.coverSaveFailed'))
  } finally {
    savingCover.value = false
  }
}

async function saveUsername() {
  const username = usernameForm.username.trim()
  if (!username) return showError(t('settings.validation.usernameRequired'))

  savingUsername.value = true
  try {
    await saveUserName(username)
    editingUsername.value = false
    showStatus(t('settings.status.usernameSaved'))
  } catch (err) {
    showError(err instanceof Error ? err.message : t('api.usernameSaveFailed'))
  } finally {
    savingUsername.value = false
  }
}

function cancelUsernameEdit() {
  usernameForm.username = page.props.user.username
  editingUsername.value = false
}

async function saveEmail() {
  const email = emailForm.email.trim()
  if (!email) return showError(t('settings.validation.emailRequired'))

  savingEmail.value = true
  try {
    await saveUserEmail(email)
    editingEmail.value = false
    showStatus(t('settings.status.emailSaved'))
  } catch (err) {
    showError(err instanceof Error ? err.message : t('api.emailSaveFailed'))
  } finally {
    savingEmail.value = false
  }
}

function cancelEmailEdit() {
  emailForm.email = page.props.user.email
  editingEmail.value = false
}

async function sendActivationEmail() {
  sendingActivationEmail.value = true
  try {
    showStatus(await resendActivationEmail())
  } catch (err) {
    showError(err instanceof Error ? err.message : t('api.activationEmailSendFailed'))
  } finally {
    sendingActivationEmail.value = false
  }
}

async function submitPassword() {
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    return showError(t('auth.validation.passwordMismatch'))
  }

  savingPassword.value = true
  try {
    await changePassword(passwordForm.oldPassword, passwordForm.newPassword)
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
    showStatus(t('settings.status.passwordChanged'))
  } catch (err) {
    showError(err instanceof Error ? err.message : t('api.passwordChangeFailed'))
  } finally {
    savingPassword.value = false
  }
}

function savePrivacy() {
  localStorage.setItem('goose-privacy-settings', JSON.stringify(privacy))
  showStatus(t('settings.status.privacySaved'))
}

async function loadBindings() {
  loadingBindings.value = true
  try {
    bindings.value = await getOAuthBindings()
  } catch (err) {
    showError(err instanceof Error ? err.message : t('api.bindingsLoadFailed'))
  } finally {
    loadingBindings.value = false
  }
}

function isBound(provider: string) {
  return Boolean(bindings.value[provider]?.bound)
}

function providerActionLabel(provider: { key: string; supported: boolean }) {
  if (!provider.supported) return t('settings.binding.unsupported')
  return isBound(provider.key) ? t('settings.binding.disconnect') : t('settings.binding.connect')
}

async function toggleBinding(provider: string) {
  const item = providers.value.find((entry) => entry.key === provider)
  if (!item?.supported) return

  if (!isBound(provider)) {
    window.location.href = `/api/auth/${provider}`
    return
  }

  bindingAction.value = provider
  try {
    await unbindOAuth(provider)
    await loadBindings()
    showStatus(t('settings.status.bindingDisconnected'))
  } catch (err) {
    showError(err instanceof Error ? err.message : t('api.unbindFailed'))
  } finally {
    bindingAction.value = ''
  }
}
</script>

<template>
    <main class="min-w-0 pb-8">
      <PageHeader :title="t('shell.settings')" :description="t('settings.profile.description')" compact />

      <section class="gf-card overflow-hidden">
        <div class="h-20 border-b border-line bg-base-300 bg-cover bg-center sm:h-24" :style="profileCoverStyle" />
        <div class="px-4 pb-4 sm:px-5">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
            <div class="flex min-w-0 gap-4">
              <button
                type="button"
                class="group relative -mt-9 h-24 w-24 shrink-0 rounded-lg border-4 border-base-100 bg-base-100 shadow-sm outline-none focus-visible:ring-4 focus-visible:ring-primary/20 sm:-mt-10 sm:h-28 sm:w-28"
                :disabled="uploadingAvatar"
                :aria-label="t('settings.avatar.upload')"
                @click="chooseAvatar"
              >
                <UserAvatar :src="avatarUrl" :alt="usernameForm.username" size="large" class="h-full w-full rounded object-cover transition group-hover:brightness-90" />
                <span class="absolute inset-0 flex items-center justify-center rounded bg-neutral/0 text-neutral-content transition group-hover:bg-neutral/20">
                  <Loader2 v-if="uploadingAvatar" class="h-6 w-6 animate-spin opacity-100" />
                  <Camera v-else class="h-6 w-6 opacity-0 transition group-hover:opacity-100" />
                </span>
                <span class="absolute -bottom-2 -right-2 flex h-8 w-8 items-center justify-center rounded-full bg-neutral text-neutral-content shadow-sm ring-2 ring-base-100">
                  <Loader2 v-if="uploadingAvatar" class="h-4 w-4 animate-spin" />
                  <Camera v-else class="h-4 w-4" />
                </span>
                <input ref="avatarInput" type="file" class="hidden" accept="image/*" @change="handleAvatarChange" />
              </button>

              <div class="min-w-0 pt-3">
                <div class="flex min-w-0 flex-wrap items-center gap-2">
                  <h2 class="truncate text-2xl font-bold leading-tight text-base-content">{{ displayName }}</h2>
                  <span class="gf-badge gf-badge-info rounded text-[11px]">{{ t('settings.editing') }}</span>
                  <button
                    type="button"
                    class="inline-flex h-7 w-7 items-center justify-center rounded outline-none transition hover:bg-base-300 focus-visible:ring-4 focus-visible:ring-primary/20"
                    :aria-label="t('settings.easterEgg.aria')"
                    :title="t('settings.easterEgg.title')"
                    @click="triggerAvatarFlash"
                  >
                    <Sparkles class="h-4 w-4 text-primary" />
                  </button>
                </div>
                <p class="mt-1 text-sm font-medium text-base-content/55">@{{ usernameForm.username }}</p>
                <p class="mt-2 max-w-3xl text-sm leading-relaxed text-base-content/75">{{ profileBioText }}</p>
              </div>
            </div>

            <div class="flex shrink-0 flex-col items-start gap-2 sm:items-end">
              <div class="flex flex-wrap items-center gap-2">
                <div v-if="layout.viewer.canAccessAdmin" class="relative">
                  <button
                    type="button"
                    class="gf-button gf-button-md gf-button-secondary"
                    :aria-expanded="editingCover"
                    @click="toggleCoverEditor"
                  >
                    <Image class="h-4 w-4" />
                    {{ t('user.editCover') }}
                  </button>
                  <form
                    v-if="editingCover"
                    class="gf-menu-surface absolute left-0 top-11 z-20 w-80 max-w-[calc(100vw-2rem)] p-3 sm:left-auto sm:right-0"
                    @submit.prevent="saveCover"
                  >
                    <label class="block">
                      <span class="text-xs font-semibold text-base-content/55">{{ t('user.coverUrl') }}</span>
                      <input
                        v-model="coverDraft"
                        type="url"
                        class="gf-input mt-1 h-9"
                        :placeholder="t('user.coverUrl')"
                      />
                    </label>
                    <div class="mt-3 flex justify-end gap-2">
                      <button
                        type="button"
                        class="gf-button gf-button-sm gf-button-secondary"
                        :disabled="savingCover"
                        @click="cancelCoverEditor"
                      >
                        {{ t('common.cancel') }}
                      </button>
                      <button
                        type="submit"
                        class="gf-button gf-button-sm gf-button-primary min-w-16 disabled:cursor-wait"
                        :disabled="savingCover"
                      >
                        <Loader2 v-if="savingCover" class="h-4 w-4 animate-spin" />
                        <span v-else>{{ t('common.save') }}</span>
                      </button>
                    </div>
                  </form>
                </div>
                <button
                  type="button"
                  class="gf-button gf-button-md gf-button-secondary"
                  @click="chooseAvatar"
                >
                  <Camera class="h-4 w-4" />
                  {{ t('settings.avatar.change') }}
                </button>
              </div>
            </div>
          </div>

          <div class="mt-5 grid grid-cols-4 border-y border-line py-3 lg:grid-cols-7 lg:py-4">
            <div v-for="item in statsItems" :key="item.label" class="px-1 py-1 text-center lg:px-0 lg:text-left">
              <div class="text-lg font-bold tabular-nums text-base-content lg:text-xl">{{ formatNumber(item.value) }}</div>
              <div class="mt-0.5 text-[11px] font-medium text-base-content/55 lg:text-xs">{{ item.label }}</div>
            </div>
          </div>

          <div class="mt-4 flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-base-content/55">
            <span class="inline-flex items-center gap-1.5"><CalendarDays class="h-3.5 w-3.5" /> {{ t('user.joinedAt', { date: formatDate(props.stats.createdAt) }) }}</span>
            <span v-if="profileForm.website" class="inline-flex min-w-0 items-center gap-1.5">
              <LinkIcon class="h-3.5 w-3.5 shrink-0" />
              <span class="truncate">{{ profileForm.websiteName || profileForm.website }}</span>
            </span>
          </div>
        </div>
      </section>

      <p
        v-if="hasStatus"
        class="my-3 rounded-md px-3 py-2 text-sm font-medium"
        :class="error ? 'bg-error/10 text-error' : 'bg-success/10 text-success'"
      >
        {{ error || status }}
      </p>

      <section class="gf-card mt-3 overflow-hidden">
        <nav class="flex overflow-x-auto border-b border-line px-3">
            <button
              v-for="tab in props.tabs"
              :key="tab.key"
              type="button"
              class="inline-flex h-11 shrink-0 items-center border-b-2 px-4 text-sm font-semibold"
              :class="activeTab === tab.key ? 'border-primary text-primary' : 'border-transparent text-base-content/55 hover:text-base-content'"
              @click="setActiveTab(tab.key as TabKey)"
            >
              {{ settingsTabLabel(tab.key, tab.label) }}
            </button>
          </nav>

        <div class="space-y-3">
          <section v-show="activeTab === 'profile'">
            <SectionHeader :icon="UserRound" :title="t('settings.profile.title')" :description="t('settings.profile.description')" />
            <div class="space-y-6 p-4">
              <div class="grid gap-4 sm:grid-cols-2">
                <label class="block min-w-0">
                  <span class="text-sm font-medium text-base-content/75">{{ t('auth.username') }}</span>
                  <div v-if="!editingUsername" class="mt-1 flex min-w-0 items-center gap-2">
                    <div class="flex h-10 min-w-0 flex-1 items-center rounded-md border border-line bg-base-200/70 px-3 text-sm font-medium text-base-content">
                      <span class="truncate">{{ usernameForm.username }}</span>
                    </div>
                    <button
                      type="button"
                      class="inline-flex h-10 shrink-0 items-center gap-1.5 rounded-md border border-primary/20 bg-info/10 px-3 text-sm font-semibold text-primary hover:border-primary/20 hover:bg-info/10"
                      @click="editingUsername = true"
                    >
                      <Pencil class="h-4 w-4" />
                      {{ t('common.edit') }}
                    </button>
                  </div>
                  <div v-else class="mt-1 flex min-w-0 gap-2">
                    <input v-model="usernameForm.username" class="gf-input min-w-0 flex-1 border-primary/40 ring-4 ring-primary/20" />
                    <button type="button" class="gf-button gf-button-lg gf-button-primary shrink-0" :disabled="savingUsername" @click="saveUsername">
                      {{ savingUsername ? t('settings.savingShort') : t('common.save') }}
                    </button>
                    <button type="button" class="gf-button gf-button-lg gf-button-muted shrink-0 px-2.5 font-medium" @click="cancelUsernameEdit">{{ t('common.cancel') }}</button>
                  </div>
                </label>
                <div class="block min-w-0">
                  <span class="text-sm font-medium text-base-content/75">{{ t('auth.email') }}</span>
                  <div v-if="!editingEmail" class="mt-1 flex min-w-0 items-center gap-2">
                    <div class="flex h-10 min-w-0 flex-1 items-center rounded-md border border-line bg-base-200/70 px-3 text-sm font-medium text-base-content">
                      <span class="truncate">{{ emailForm.email }}</span>
                    </div>
                    <button
                      type="button"
                      class="inline-flex h-10 shrink-0 items-center gap-1.5 rounded-md border border-primary/20 bg-info/10 px-3 text-sm font-semibold text-primary hover:border-primary/20 hover:bg-info/10"
                      @click="editingEmail = true"
                    >
                      <Pencil class="h-4 w-4" />
                      {{ t('common.edit') }}
                    </button>
                  </div>
                  <div v-else class="mt-1 flex min-w-0 gap-2">
                    <input v-model="emailForm.email" type="email" class="gf-input min-w-0 flex-1 border-primary/40 ring-4 ring-primary/20" />
                    <button type="button" class="gf-button gf-button-lg gf-button-primary shrink-0" :disabled="savingEmail" @click="saveEmail">
                      {{ savingEmail ? t('settings.savingShort') : t('common.save') }}
                    </button>
                    <button type="button" class="gf-button gf-button-lg gf-button-muted shrink-0 px-2.5 font-medium" @click="cancelEmailEdit">{{ t('common.cancel') }}</button>
                  </div>
                  <div v-if="layout.viewer.requiresEmailVerification" class="mt-2 flex flex-col gap-2 border-l-2 border-warning bg-warning/10 px-3 py-2 sm:flex-row sm:items-center sm:justify-between">
                    <span class="min-w-0 text-sm text-warning">
                      <span class="font-semibold">{{ t('settings.emailVerification.title') }}</span>
                      <span class="ml-1 text-warning">{{ t('settings.emailVerification.description') }}</span>
                    </span>
                    <button
                      type="button"
                      class="inline-flex h-8 shrink-0 items-center justify-center gap-1.5 rounded-md border border-warning/30 bg-base-100 px-3 text-sm font-semibold text-warning hover:bg-warning/15 disabled:cursor-wait disabled:opacity-70"
                      :disabled="sendingActivationEmail"
                      @click="sendActivationEmail"
                    >
                      <Loader2 v-if="sendingActivationEmail" class="h-4 w-4 animate-spin" />
                      <Mail v-else class="h-4 w-4" />
                      {{ sendingActivationEmail ? t('settings.emailVerification.sending') : t('settings.emailVerification.action') }}
                    </button>
                  </div>
                </div>
                <label class="block sm:col-span-2">
                  <span class="text-sm font-medium text-base-content/75">{{ t('settings.profile.displayName') }}</span>
                  <input v-model="profileForm.nickname" class="gf-input mt-1" />
                </label>
                <label class="block">
                  <span class="text-sm font-medium text-base-content/75">{{ t('settings.profile.websiteName') }}</span>
                  <input v-model="profileForm.websiteName" class="gf-input mt-1" />
                </label>
                <label class="block">
                  <span class="text-sm font-medium text-base-content/75">{{ t('settings.profile.website') }}</span>
                  <input v-model="profileForm.website" class="gf-input mt-1" placeholder="https://example.com" />
                </label>
              </div>

              <label class="block">
                <span class="text-sm font-medium text-base-content/75">{{ t('settings.profile.bio') }}</span>
                <textarea v-model="profileForm.bio" class="gf-textarea mt-1 min-h-24 py-2" />
              </label>
              <label class="block">
                <span class="text-sm font-medium text-base-content/75">{{ t('settings.profile.signature') }}</span>
                <textarea v-model="profileForm.signature" class="gf-textarea mt-1 min-h-20 py-2" />
              </label>

              <div class="border-t border-line pt-5">
                <div class="mb-3 flex items-center gap-2">
                  <LinkIcon class="h-4 w-4 text-base-content/55" />
                  <h3 class="text-sm font-semibold text-base-content">{{ t('settings.profile.social') }}</h3>
                </div>
                <div class="grid gap-3 sm:grid-cols-2">
                  <label v-for="item in socialItems" :key="item.key" class="block">
                    <span class="inline-flex items-center gap-2 text-sm font-medium text-base-content/75">
                      <span class="inline-flex h-5 w-5 items-center justify-center text-base-content/55">
                        <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                          <path :d="item.icon.path" />
                        </svg>
                      </span>
                      {{ item.label }}
                    </span>
                    <input v-model="profileForm.externalInformation[item.key].link" class="gf-input mt-1" />
                  </label>
                </div>
              </div>

              <div class="border-t border-line pt-5">
                <button
                  type="button"
                  class="gf-button gf-button-lg gf-button-primary min-w-28 disabled:cursor-wait"
                  :disabled="savingProfile"
                  @click="saveProfile"
                >
                  <Loader2 v-if="savingProfile" class="h-4 w-4 animate-spin" />
                  <span>{{ savingProfile ? t('settings.savingShort') : t('settings.profile.save') }}</span>
                </button>
              </div>
            </div>
          </section>

          <section v-show="activeTab === 'account'">
            <SectionHeader :icon="KeyRound" :title="t('settings.account.title')" />
            <form class="max-w-xl space-y-4 p-4" @submit.prevent="submitPassword">
              <label class="block">
                <span class="text-sm font-medium text-base-content/75">{{ t('settings.account.currentPassword') }}</span>
                <input v-model="passwordForm.oldPassword" required type="password" class="gf-input mt-1" />
              </label>
              <label class="block">
                <span class="text-sm font-medium text-base-content/75">{{ t('auth.newPassword') }}</span>
                <input v-model="passwordForm.newPassword" required type="password" class="gf-input mt-1" />
                <span class="mt-1 block text-xs text-base-content/55">{{ t('settings.account.passwordHint') }}</span>
              </label>
              <label class="block">
                <span class="text-sm font-medium text-base-content/75">{{ t('auth.confirmPassword') }}</span>
                <input v-model="passwordForm.confirmPassword" required type="password" class="gf-input mt-1" />
              </label>
              <button type="submit" class="gf-button gf-button-lg gf-button-primary disabled:cursor-wait" :disabled="savingPassword">
                <Loader2 v-if="savingPassword" class="h-4 w-4 animate-spin" />
                {{ t('settings.account.changePassword') }}
              </button>
            </form>
          </section>

          <section v-show="activeTab === 'privacy'">
            <SectionHeader :icon="Shield" :title="t('settings.privacy.title')" />
            <div class="max-w-2xl divide-y divide-line p-4">
              <label class="flex items-center justify-between gap-4 py-4">
                <span>
                  <span class="block text-sm font-semibold text-base-content">{{ t('settings.privacy.showArticles') }}</span>
                  <span class="text-sm text-base-content/55">{{ t('settings.privacy.showArticlesDescription') }}</span>
                </span>
                <input v-model="privacy.showArticles" type="checkbox" class="h-5 w-5 rounded border-line text-primary" @change="savePrivacy" />
              </label>
              <label class="flex items-center justify-between gap-4 py-4">
                <span>
                  <span class="block text-sm font-semibold text-base-content">{{ t('settings.privacy.showFollowing') }}</span>
                  <span class="text-sm text-base-content/55">{{ t('settings.privacy.showFollowingDescription') }}</span>
                </span>
                <input v-model="privacy.showFollowing" type="checkbox" class="h-5 w-5 rounded border-line text-primary" @change="savePrivacy" />
              </label>
              <label class="flex items-center justify-between gap-4 py-4">
                <span>
                  <span class="block text-sm font-semibold text-base-content">{{ t('settings.privacy.emailNotifications') }}</span>
                  <span class="text-sm text-base-content/55">{{ t('settings.privacy.emailNotificationsDescription') }}</span>
                </span>
                <input v-model="privacy.emailNotifications" type="checkbox" class="h-5 w-5 rounded border-line text-primary" @change="savePrivacy" />
              </label>
            </div>
          </section>

          <section v-show="activeTab === 'binding'">
            <SectionHeader :icon="Mail" :title="t('settings.binding.title')">
              <template #actions>
                <button type="button" class="text-xs font-medium text-primary hover:text-primary" @click="loadBindings">{{ t('settings.binding.refresh') }}</button>
              </template>
            </SectionHeader>
            <div v-if="loadingBindings" class="p-4 py-8 text-center text-sm text-base-content/55">
              <Loader2 class="mx-auto mb-2 h-5 w-5 animate-spin" />
              {{ t('settings.binding.loading') }}
            </div>
            <div v-else class="space-y-3 p-4">
              <div
                v-for="provider in providers"
                :key="provider.key"
                class="flex items-center justify-between gap-4 rounded-lg border p-4"
                :class="provider.supported ? 'border-line bg-base-100' : 'border-line bg-base-200/70'"
              >
                <div class="flex min-w-0 items-center gap-3">
                  <div
                    class="flex h-11 w-11 shrink-0 items-center justify-center rounded-full border"
                    :class="provider.supported ? 'border-line bg-base-100 shadow-sm' : 'border-line bg-base-300 opacity-60'"
                  >
                    <svg v-if="provider.key === 'github'" class="h-6 w-6 text-base-content" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                      <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.44 9.8 8.21 11.39.6.11.82-.26.82-.58v-2.04c-3.34.73-4.04-1.61-4.04-1.61-.55-1.39-1.34-1.76-1.34-1.76-1.09-.75.08-.73.08-.73 1.21.08 1.84 1.24 1.84 1.24 1.07 1.83 2.81 1.3 3.49.99.11-.78.42-1.3.76-1.6-2.67-.3-5.47-1.33-5.47-5.93 0-1.31.47-2.38 1.24-3.22-.12-.3-.54-1.52.12-3.18 0 0 1.01-.32 3.3 1.23A11.5 11.5 0 0 1 12 5.8c1.02.01 2.05.14 3.01.4 2.29-1.55 3.3-1.23 3.3-1.23.65 1.66.24 2.88.12 3.18.77.84 1.23 1.91 1.23 3.22 0 4.61-2.81 5.62-5.48 5.92.43.37.82 1.1.82 2.22v3.29c0 .32.22.7.82.58A12.01 12.01 0 0 0 24 12c0-6.63-5.37-12-12-12Z" />
                    </svg>
                    <svg v-else class="h-6 w-6" viewBox="0 0 24 24" aria-hidden="true">
                      <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.31v2.77h3.56c2.08-1.92 3.28-4.74 3.28-8.09Z" />
                      <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.56-2.77c-.99.66-2.24 1.06-3.72 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84A11 11 0 0 0 12 23Z" />
                      <path fill="#FBBC05" d="M5.84 14.1A6.61 6.61 0 0 1 5.5 12c0-.73.12-1.44.34-2.1V7.07H2.18A11 11 0 0 0 1 12c0 1.78.43 3.45 1.18 4.93l3.66-2.83Z" />
                      <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15A10.6 10.6 0 0 0 12 1 11 11 0 0 0 2.18 7.07L5.84 9.9C6.71 7.31 9.14 5.38 12 5.38Z" />
                    </svg>
                  </div>
                  <div>
                    <h3 class="font-semibold text-base-content">{{ provider.label }}</h3>
                    <p class="text-sm" :class="provider.supported ? 'text-base-content/55' : 'text-base-content/55'">
                      {{ provider.supported ? (isBound(provider.key) ? t('settings.binding.connected') : t('settings.binding.disconnected')) : t('settings.binding.siteUnsupported') }}
                    </p>
                  </div>
                </div>
                <button
                  type="button"
                  class="inline-flex h-9 min-w-24 items-center justify-center gap-2 rounded-md border px-3 text-sm font-semibold disabled:cursor-not-allowed"
                  :class="[
                    !provider.supported
                      ? 'border-line bg-base-300 text-base-content/55'
                      : isBound(provider.key)
                        ? 'border-error/30 bg-error/10 text-error hover:bg-error/10'
                        : 'border-neutral bg-neutral text-neutral-content hover:bg-neutral/90',
                  ]"
                  :disabled="bindingAction === provider.key || !provider.supported"
                  @click="toggleBinding(provider.key)"
                >
                  <Loader2 v-if="bindingAction === provider.key" class="h-4 w-4 animate-spin" />
                  <Check v-else-if="isBound(provider.key)" class="h-4 w-4" />
                  {{ providerActionLabel(provider) }}
                </button>
              </div>
            </div>
          </section>
        </div>
      </section>

      <div v-if="cropModalOpen" class="fixed inset-0 z-[100] overflow-y-auto bg-neutral/50 px-3 py-4 backdrop-blur-sm sm:px-4" role="dialog" aria-modal="true">
        <div class="mx-auto flex min-h-full max-w-[760px] items-center justify-center">
          <div class="gf-menu-surface flex max-h-[calc(100vh-2rem)] w-full flex-col overflow-hidden">
            <div class="flex items-center justify-between border-b border-line px-5 py-3">
              <div>
                <h2 class="text-base font-semibold text-base-content">{{ t('settings.avatar.cropTitle') }}</h2>
                <p class="mt-0.5 text-sm text-base-content/55">{{ t('settings.avatar.cropDescription') }}</p>
              </div>
              <button type="button" class="rounded-md px-2 py-1 text-sm font-medium text-base-content/55 hover:bg-base-300 hover:text-base-content" @click="closeCropModal">
                {{ t('common.close') }}
              </button>
            </div>

            <div class="grid gap-4 overflow-y-auto p-4 md:grid-cols-[minmax(280px,420px)_180px] md:items-start md:justify-center">
              <div class="avatar-crop-workspace aspect-square w-full max-w-[420px] justify-self-center overflow-hidden rounded-lg border border-line bg-base-200">
                <img ref="cropperImage" :src="cropImageUrl" :alt="t('settings.avatar.cropAlt')" class="block" />
              </div>

              <aside class="grid gap-4 sm:grid-cols-[128px_minmax(0,1fr)] md:block md:space-y-4">
                <div class="min-w-0">
                  <div class="mb-2 text-sm font-semibold text-base-content">{{ t('settings.avatar.preview') }}</div>
                  <div class="flex h-32 w-32 items-center justify-center overflow-hidden rounded-full border border-line bg-base-200">
                    <img v-if="cropPreviewUrl" :src="cropPreviewUrl" :alt="t('settings.avatar.previewAlt')" class="h-full w-full object-cover" />
                  </div>
                </div>
                <div class="self-start rounded-lg bg-base-200 p-3 text-sm leading-6 text-base-content/55">
                  {{ t('settings.avatar.cropTip') }}
                </div>
              </aside>
            </div>

            <div v-if="cropError" class="border-t border-error/20 bg-error/10 px-5 py-3 text-sm font-medium text-error">
              {{ cropError }}
            </div>

            <div class="flex items-center justify-end gap-2 border-t border-line bg-base-200 px-5 py-3">
              <button type="button" class="gf-button gf-button-lg gf-button-muted font-medium" @click="closeCropModal">
                {{ t('common.cancel') }}
              </button>
              <button
                type="button"
                class="gf-button gf-button-lg gf-button-primary min-w-28 disabled:cursor-wait"
                :disabled="uploadingAvatar"
                @click="uploadCroppedAvatar"
              >
                <Loader2 v-if="uploadingAvatar" class="h-4 w-4 animate-spin" />
                {{ uploadingAvatar ? t('settings.avatar.uploading') : t('settings.avatar.confirmUpload') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>
</template>
