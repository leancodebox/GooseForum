<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import Cropper from 'cropperjs'
import {
  CalendarDays,
  Camera,
  Check,
  KeyRound,
  Link as LinkIcon,
  Loader2,
  Mail,
  Pencil,
  Shield,
  UserRound,
} from '@lucide/vue'
import {
  changePassword,
  getOAuthBindings,
  saveUserEmail,
  saveUserInfo,
  saveUserName,
  unbindOAuth,
  uploadAvatar,
  type OAuthBindingsPayload,
} from '@/runtime/api'
import { formatDate, formatNumber } from '@/runtime/format'
import { useFlashMessages, type FlashMessageType } from '@/runtime/flash-message'
import { canvasToImageFile, validateImageFile } from '@/runtime/image'
import type { LayoutPayload, SettingsPageProps } from '@/types/payload'
import FlashSpriteIcon from '@/site/components/FlashSpriteIcon.vue'
import { socialIcons, socialLabels } from '@/site/utils/social-icons'

const page = defineProps<{
  layout: LayoutPayload
  props: SettingsPageProps
}>()

const tabKeys = ['profile', 'account', 'privacy', 'binding'] as const
type TabKey = (typeof tabKeys)[number]

const activeTab = ref<TabKey>('profile')
const status = ref('')
const error = ref('')
const savingProfile = ref(false)
const savingUsername = ref(false)
const savingEmail = ref(false)
const savingPassword = ref(false)
const uploadingAvatar = ref(false)
const loadingBindings = ref(false)
const bindingAction = ref('')
const avatarInput = ref<HTMLInputElement | null>(null)
const cropperImage = ref<HTMLImageElement | null>(null)
const avatarUrl = ref(page.props.user.avatarUrl)
const cropModalOpen = ref(false)
const cropImageUrl = ref('')
const cropPreviewUrl = ref('')
const cropSourceFile = ref<File | null>(null)
let cropper: Cropper | undefined
let cropPreviewFrame = 0
const editingUsername = ref(false)
const editingEmail = ref(false)
const bindings = ref<OAuthBindingsPayload>({})
const { push: pushFlash } = useFlashMessages()

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
const profileBioText = computed(() => profileForm.bio || profileForm.signature || '这个用户还没有留下简介。')
const hasStatus = computed(() => Boolean(status.value || error.value))
const statsItems = computed(() => [
  { label: '主题', value: page.props.stats.articleCount },
  { label: '回复', value: page.props.stats.replyCount },
  { label: '获赞', value: page.props.stats.likeReceivedCount },
  { label: '点赞', value: page.props.stats.likeGivenCount },
  { label: '粉丝', value: page.props.stats.followerCount },
  { label: '关注', value: page.props.stats.followingCount },
  { label: '收藏', value: page.props.stats.collectionCount },
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
  { type: 'success', message: '彩蛋触发成功：今天的头像也很有精神。' },
  { type: 'info', message: '系统提示：这条消息只是来检查展示效果的。' },
  { type: 'warning', message: '小提醒：你发现了设置页侧栏头像的隐藏开关。' },
  { type: 'error', message: '模拟错误：别紧张，这只是消息框验收专用。' },
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

onBeforeUnmount(() => {
  destroyCropper()
  revokeCropImageUrl()
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
    showStatus('资料已保存')
  } catch (err) {
    showError(err instanceof Error ? err.message : '资料保存失败')
  } finally {
    savingProfile.value = false
  }
}

async function saveUsername() {
  const username = usernameForm.username.trim()
  if (!username) return showError('用户名不能为空')

  savingUsername.value = true
  try {
    await saveUserName(username)
    editingUsername.value = false
    showStatus('用户名已更新')
  } catch (err) {
    showError(err instanceof Error ? err.message : '用户名保存失败')
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
  if (!email) return showError('邮箱不能为空')

  savingEmail.value = true
  try {
    await saveUserEmail(email)
    editingEmail.value = false
    showStatus('邮箱已更新，请留意验证邮件')
  } catch (err) {
    showError(err instanceof Error ? err.message : '邮箱保存失败')
  } finally {
    savingEmail.value = false
  }
}

function cancelEmailEdit() {
  emailForm.email = page.props.user.email
  editingEmail.value = false
}

async function submitPassword() {
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    return showError('两次输入的新密码不一致')
  }

  savingPassword.value = true
  try {
    await changePassword(passwordForm.oldPassword, passwordForm.newPassword)
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
    showStatus('密码已修改')
  } catch (err) {
    showError(err instanceof Error ? err.message : '密码修改失败')
  } finally {
    savingPassword.value = false
  }
}

function savePrivacy() {
  localStorage.setItem('goose-privacy-settings', JSON.stringify(privacy))
  showStatus('隐私偏好已保存')
}

function chooseAvatar() {
  avatarInput.value?.click()
}

async function handleAvatarChange(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  const validationError = validateImageFile(file, 5 * 1024 * 1024)
  if (validationError) return showError(validationError)

  openCropModal(file)
  if (avatarInput.value) avatarInput.value.value = ''
}

function openCropModal(file: File) {
  destroyCropper()
  revokeCropImageUrl()
  cropSourceFile.value = file
  cropImageUrl.value = URL.createObjectURL(file)
  cropModalOpen.value = true
  void nextTick(initCropper)
}

function initCropper() {
  if (!cropperImage.value) return
  cropper = new Cropper(cropperImage.value, {
    template: `
      <cropper-canvas background>
        <cropper-image translatable scalable rotatable></cropper-image>
        <cropper-shade hidden></cropper-shade>
        <cropper-handle action="select" plain></cropper-handle>
        <cropper-selection initial-coverage="0.92" aspect-ratio="1" movable resizable zoomable outlined>
          <cropper-grid role="grid" bordered covered></cropper-grid>
          <cropper-crosshair centered></cropper-crosshair>
          <cropper-handle action="move" theme-color="rgba(37, 99, 235, 0.35)"></cropper-handle>
          <cropper-handle action="n-resize"></cropper-handle>
          <cropper-handle action="e-resize"></cropper-handle>
          <cropper-handle action="s-resize"></cropper-handle>
          <cropper-handle action="w-resize"></cropper-handle>
          <cropper-handle action="ne-resize"></cropper-handle>
          <cropper-handle action="nw-resize"></cropper-handle>
          <cropper-handle action="se-resize"></cropper-handle>
          <cropper-handle action="sw-resize"></cropper-handle>
        </cropper-selection>
      </cropper-canvas>
    `,
  })
  window.setTimeout(updateCropPreview, 120)
  cropper.container.addEventListener('pointerup', scheduleCropPreview)
  cropper.container.addEventListener('wheel', scheduleCropPreview, { passive: true })
  cropper.container.addEventListener('keyup', scheduleCropPreview)
}

function closeCropModal() {
  cropModalOpen.value = false
  cropSourceFile.value = null
  destroyCropper()
  revokeCropImageUrl()
}

function destroyCropper() {
  window.cancelAnimationFrame(cropPreviewFrame)
  cropPreviewFrame = 0
  cropper?.destroy()
  cropper = undefined
  cropPreviewUrl.value = ''
}

function revokeCropImageUrl() {
  if (cropImageUrl.value) URL.revokeObjectURL(cropImageUrl.value)
  cropImageUrl.value = ''
}

async function uploadCroppedAvatar() {
  if (!cropper || !cropSourceFile.value) return

  uploadingAvatar.value = true
  try {
    const selection = cropper.getCropperSelection()
    if (!selection) throw new Error('请选择裁切区域')
    const canvas = await selection.$toCanvas({
      width: 400,
      height: 400,
      beforeDraw(context) {
        context.imageSmoothingEnabled = true
        context.imageSmoothingQuality = 'high'
      },
    })
    const avatarFile = await canvasToImageFile(canvas, cropSourceFile.value.name, undefined, 0.86)
    avatarUrl.value = await uploadAvatar(avatarFile)
    closeCropModal()
    showStatus('头像已更新')
  } catch (err) {
    showError(err instanceof Error ? err.message : '头像上传失败')
  } finally {
    uploadingAvatar.value = false
  }
}

function scheduleCropPreview() {
  window.cancelAnimationFrame(cropPreviewFrame)
  cropPreviewFrame = window.requestAnimationFrame(() => {
    void updateCropPreview()
  })
}

async function updateCropPreview() {
  const selection = cropper?.getCropperSelection()
  if (!selection) return
  try {
    const canvas = await selection.$toCanvas({
      width: 160,
      height: 160,
      beforeDraw(context) {
        context.imageSmoothingEnabled = true
        context.imageSmoothingQuality = 'high'
      },
    })
    cropPreviewUrl.value = canvas.toDataURL('image/webp', 0.82)
  } catch {
    cropPreviewUrl.value = ''
  }
}

async function loadBindings() {
  loadingBindings.value = true
  try {
    bindings.value = await getOAuthBindings()
  } catch (err) {
    showError(err instanceof Error ? err.message : '绑定状态加载失败')
  } finally {
    loadingBindings.value = false
  }
}

function isBound(provider: string) {
  return Boolean(bindings.value[provider]?.bound)
}

function providerActionLabel(provider: { key: string; supported: boolean }) {
  if (!provider.supported) return '暂不支持'
  return isBound(provider.key) ? '解除绑定' : '连接'
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
    showStatus('账号绑定已解除')
  } catch (err) {
    showError(err instanceof Error ? err.message : '解绑失败')
  } finally {
    bindingAction.value = ''
  }
}
</script>

<template>
    <main class="min-w-0 pb-12">
      <section class="mb-4 overflow-hidden rounded-lg border border-gray-200/70 bg-white shadow-[0_2px_12px_rgba(0,0,0,0.04)]">
        <div class="h-24 border-b border-gray-100 bg-[linear-gradient(135deg,#f8fafc_0%,#eff6ff_48%,#f8fafc_100%)]" />
        <div class="px-4 pb-4 sm:px-5">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
            <div class="flex min-w-0 gap-4">
              <button
                type="button"
                class="group relative -mt-10 h-20 w-20 shrink-0 rounded-lg border-4 border-white bg-white shadow-sm outline-none focus-visible:ring-4 focus-visible:ring-blue-100 sm:h-24 sm:w-24"
                :disabled="uploadingAvatar"
                aria-label="上传头像"
                @click="chooseAvatar"
              >
                <img :src="avatarUrl" :alt="usernameForm.username" class="h-full w-full rounded object-cover transition group-hover:brightness-90" />
                <span class="absolute inset-0 flex items-center justify-center rounded bg-gray-950/0 text-white transition group-hover:bg-gray-950/20">
                  <Loader2 v-if="uploadingAvatar" class="h-6 w-6 animate-spin opacity-100" />
                  <Camera v-else class="h-6 w-6 opacity-0 transition group-hover:opacity-100" />
                </span>
                <span class="absolute -bottom-2 -right-2 flex h-8 w-8 items-center justify-center rounded-full bg-gray-900 text-white shadow-sm ring-2 ring-white">
                  <Loader2 v-if="uploadingAvatar" class="h-4 w-4 animate-spin" />
                  <Camera v-else class="h-4 w-4" />
                </span>
                <input ref="avatarInput" type="file" class="hidden" accept="image/*" @change="handleAvatarChange" />
              </button>

              <div class="min-w-0 pt-3">
                <div class="flex min-w-0 flex-wrap items-center gap-2">
                  <h2 class="truncate text-2xl font-bold leading-tight text-gray-950">{{ displayName }}</h2>
                  <span class="rounded bg-blue-50 px-1.5 py-0.5 text-[11px] font-semibold text-blue-700">编辑中</span>
                  <button
                    type="button"
                    class="inline-flex h-7 w-7 items-center justify-center rounded-full outline-none transition hover:-translate-y-0.5 focus-visible:ring-4 focus-visible:ring-blue-100"
                    aria-label="触发随机系统提示"
                    title="触发彩蛋"
                    @click="triggerAvatarFlash"
                  >
                    <FlashSpriteIcon type="info" class="h-6 w-6" />
                  </button>
                </div>
                <p class="mt-1 text-sm font-medium text-gray-400">@{{ usernameForm.username }}</p>
                <p class="mt-2 max-w-3xl text-sm leading-relaxed text-gray-600">{{ profileBioText }}</p>
              </div>
            </div>

            <div class="flex shrink-0 flex-wrap items-center gap-2">
              <button
                type="button"
                class="inline-flex h-9 items-center gap-1.5 rounded-md border border-gray-200 bg-white px-3 text-sm font-semibold text-gray-700 hover:bg-gray-50"
                @click="chooseAvatar"
              >
                <Camera class="h-4 w-4" />
                更换头像
              </button>
            </div>
          </div>

          <div class="mt-5 grid grid-cols-4 border-y border-gray-100 py-3 lg:grid-cols-7 lg:py-4">
            <div v-for="item in statsItems" :key="item.label" class="px-1 py-1 text-center lg:px-0 lg:text-left">
              <div class="text-lg font-bold tabular-nums text-gray-950 lg:text-xl">{{ formatNumber(item.value) }}</div>
              <div class="mt-0.5 text-[11px] font-medium text-gray-400 lg:text-xs">{{ item.label }}</div>
            </div>
          </div>

          <div class="mt-4 flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-gray-400">
            <span class="inline-flex items-center gap-1.5"><CalendarDays class="h-3.5 w-3.5" /> 加入于 {{ formatDate(props.stats.createdAt) }}</span>
            <span v-if="profileForm.website" class="inline-flex min-w-0 items-center gap-1.5">
              <LinkIcon class="h-3.5 w-3.5 shrink-0" />
              <span class="truncate">{{ profileForm.websiteName || profileForm.website }}</span>
            </span>
          </div>
        </div>
      </section>

      <p
        v-if="hasStatus"
        class="mb-3 rounded-md px-3 py-2 text-sm font-medium"
        :class="error ? 'bg-red-50 text-red-600' : 'bg-green-50 text-green-700'"
      >
        {{ error || status }}
      </p>

      <section class="overflow-hidden rounded-lg border border-gray-200/70 bg-white shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
        <nav class="flex overflow-x-auto border-b border-gray-100 px-3">
            <button
              v-for="tab in props.tabs"
              :key="tab.key"
              type="button"
              class="inline-flex h-11 shrink-0 items-center border-b-2 px-4 text-sm font-semibold"
              :class="activeTab === tab.key ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-900'"
              @click="setActiveTab(tab.key as TabKey)"
            >
              {{ tab.label }}
            </button>
          </nav>

        <div class="space-y-3">
          <section v-show="activeTab === 'profile'">
            <div class="flex items-center gap-2 border-b border-gray-100 px-4 py-3">
              <UserRound class="h-4 w-4 text-gray-400" />
              <div>
                <h2 class="text-sm font-semibold text-gray-950">公开资料</h2>
                <p class="mt-0.5 text-xs text-gray-400">这里的内容会展示在你的个人主页和悬停卡片中。</p>
              </div>
            </div>
            <div class="space-y-6 p-4">
              <div class="grid gap-4 sm:grid-cols-2">
                <label class="block min-w-0">
                  <span class="text-sm font-medium text-gray-700">用户名</span>
                  <div v-if="!editingUsername" class="mt-1 flex min-w-0 items-center gap-2">
                    <div class="flex h-10 min-w-0 flex-1 items-center rounded-md border border-gray-100 bg-gray-50/70 px-3 text-sm font-medium text-gray-900">
                      <span class="truncate">{{ usernameForm.username }}</span>
                    </div>
                    <button
                      type="button"
                      class="inline-flex h-10 shrink-0 items-center gap-1.5 rounded-md border border-blue-100 bg-blue-50 px-3 text-sm font-semibold text-blue-700 hover:border-blue-200 hover:bg-blue-100"
                      @click="editingUsername = true"
                    >
                      <Pencil class="h-4 w-4" />
                      编辑
                    </button>
                  </div>
                  <div v-else class="mt-1 flex min-w-0 gap-2">
                    <input v-model="usernameForm.username" class="h-10 min-w-0 flex-1 rounded-md border border-blue-300 px-3 text-sm outline-none ring-4 ring-blue-50 focus:border-blue-500" />
                    <button type="button" class="h-10 shrink-0 rounded-md bg-blue-600 px-3 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-60" :disabled="savingUsername" @click="saveUsername">
                      {{ savingUsername ? '保存中' : '保存' }}
                    </button>
                    <button type="button" class="h-10 shrink-0 rounded-md px-2.5 text-sm font-medium text-gray-500 hover:bg-gray-100" @click="cancelUsernameEdit">取消</button>
                  </div>
                </label>
                <label class="block min-w-0">
                  <span class="text-sm font-medium text-gray-700">邮箱</span>
                  <div v-if="!editingEmail" class="mt-1 flex min-w-0 items-center gap-2">
                    <div class="flex h-10 min-w-0 flex-1 items-center rounded-md border border-gray-100 bg-gray-50/70 px-3 text-sm font-medium text-gray-900">
                      <span class="truncate">{{ emailForm.email }}</span>
                    </div>
                    <button
                      type="button"
                      class="inline-flex h-10 shrink-0 items-center gap-1.5 rounded-md border border-blue-100 bg-blue-50 px-3 text-sm font-semibold text-blue-700 hover:border-blue-200 hover:bg-blue-100"
                      @click="editingEmail = true"
                    >
                      <Pencil class="h-4 w-4" />
                      编辑
                    </button>
                  </div>
                  <div v-else class="mt-1 flex min-w-0 gap-2">
                    <input v-model="emailForm.email" type="email" class="h-10 min-w-0 flex-1 rounded-md border border-blue-300 px-3 text-sm outline-none ring-4 ring-blue-50 focus:border-blue-500" />
                    <button type="button" class="h-10 shrink-0 rounded-md bg-blue-600 px-3 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-60" :disabled="savingEmail" @click="saveEmail">
                      {{ savingEmail ? '保存中' : '保存' }}
                    </button>
                    <button type="button" class="h-10 shrink-0 rounded-md px-2.5 text-sm font-medium text-gray-500 hover:bg-gray-100" @click="cancelEmailEdit">取消</button>
                  </div>
                </label>
                <label class="block">
                  <span class="text-sm font-medium text-gray-700">显示名称</span>
                  <input v-model="profileForm.nickname" class="mt-1 h-10 w-full rounded-md border border-gray-200 px-3 text-sm outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-100" />
                </label>
                <label class="block">
                  <span class="text-sm font-medium text-gray-700">网站名称</span>
                  <input v-model="profileForm.websiteName" class="mt-1 h-10 w-full rounded-md border border-gray-200 px-3 text-sm outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-100" />
                </label>
                <label class="block sm:col-span-2">
                  <span class="text-sm font-medium text-gray-700">个人网站</span>
                  <input v-model="profileForm.website" class="mt-1 h-10 w-full rounded-md border border-gray-200 px-3 text-sm outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="https://example.com" />
                </label>
              </div>

              <label class="block">
                <span class="text-sm font-medium text-gray-700">简介</span>
                <textarea v-model="profileForm.bio" class="mt-1 min-h-24 w-full rounded-md border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-100" />
              </label>
              <label class="block">
                <span class="text-sm font-medium text-gray-700">签名</span>
                <textarea v-model="profileForm.signature" class="mt-1 min-h-20 w-full rounded-md border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-100" />
              </label>

              <div class="border-t border-gray-100 pt-5">
                <div class="mb-3 flex items-center gap-2">
                  <LinkIcon class="h-4 w-4 text-gray-400" />
                  <h3 class="text-sm font-semibold text-gray-950">社交资料</h3>
                </div>
                <div class="grid gap-3 sm:grid-cols-2">
                  <label v-for="item in socialItems" :key="item.key" class="block">
                    <span class="inline-flex items-center gap-2 text-sm font-medium text-gray-700">
                      <span class="inline-flex h-5 w-5 items-center justify-center text-gray-500">
                        <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                          <path :d="item.icon.path" />
                        </svg>
                      </span>
                      {{ item.label }}
                    </span>
                    <input v-model="profileForm.externalInformation[item.key].link" class="mt-1 h-10 w-full rounded-md border border-gray-200 px-3 text-sm outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-100" />
                  </label>
                </div>
              </div>

              <div class="border-t border-gray-100 pt-5">
                <button
                  type="button"
                  class="inline-flex h-10 min-w-28 items-center justify-center gap-2 rounded-md bg-blue-600 px-4 text-sm font-semibold text-white hover:bg-blue-700 disabled:cursor-wait disabled:bg-blue-600"
                  :disabled="savingProfile"
                  @click="saveProfile"
                >
                  <Loader2 v-if="savingProfile" class="h-4 w-4 animate-spin" />
                  <span>{{ savingProfile ? '保存中' : '保存资料' }}</span>
                </button>
              </div>
            </div>
          </section>

          <section v-show="activeTab === 'account'" class="p-4">
            <div class="mb-4 flex items-center gap-2">
              <KeyRound class="h-4 w-4 text-gray-400" />
              <h2 class="text-sm font-semibold text-gray-950">账号安全</h2>
            </div>
            <form class="max-w-xl space-y-4" @submit.prevent="submitPassword">
              <label class="block">
                <span class="text-sm font-medium text-gray-700">当前密码</span>
                <input v-model="passwordForm.oldPassword" required type="password" class="mt-1 h-10 w-full rounded-md border border-gray-200 px-3 text-sm outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-100" />
              </label>
              <label class="block">
                <span class="text-sm font-medium text-gray-700">新密码</span>
                <input v-model="passwordForm.newPassword" required type="password" class="mt-1 h-10 w-full rounded-md border border-gray-200 px-3 text-sm outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-100" />
                <span class="mt-1 block text-xs text-gray-400">至少 6 个字符。</span>
              </label>
              <label class="block">
                <span class="text-sm font-medium text-gray-700">确认新密码</span>
                <input v-model="passwordForm.confirmPassword" required type="password" class="mt-1 h-10 w-full rounded-md border border-gray-200 px-3 text-sm outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-100" />
              </label>
              <button type="submit" class="inline-flex h-10 items-center gap-2 rounded-md bg-blue-600 px-4 text-sm font-semibold text-white hover:bg-blue-700 disabled:cursor-wait disabled:opacity-70" :disabled="savingPassword">
                <Loader2 v-if="savingPassword" class="h-4 w-4 animate-spin" />
                修改密码
              </button>
            </form>
          </section>

          <section v-show="activeTab === 'privacy'" class="p-4">
            <div class="mb-2 flex items-center gap-2">
              <Shield class="h-4 w-4 text-gray-400" />
              <h2 class="text-sm font-semibold text-gray-950">隐私偏好</h2>
            </div>
            <div class="max-w-2xl divide-y divide-gray-100">
              <label class="flex items-center justify-between gap-4 py-4">
                <span>
                  <span class="block text-sm font-semibold text-gray-900">展示我的主题</span>
                  <span class="text-sm text-gray-500">允许其他人从个人页查看公开主题。</span>
                </span>
                <input v-model="privacy.showArticles" type="checkbox" class="h-5 w-5 rounded border-gray-300 text-blue-600" @change="savePrivacy" />
              </label>
              <label class="flex items-center justify-between gap-4 py-4">
                <span>
                  <span class="block text-sm font-semibold text-gray-900">展示关注关系</span>
                  <span class="text-sm text-gray-500">允许其他人查看关注和粉丝信息。</span>
                </span>
                <input v-model="privacy.showFollowing" type="checkbox" class="h-5 w-5 rounded border-gray-300 text-blue-600" @change="savePrivacy" />
              </label>
              <label class="flex items-center justify-between gap-4 py-4">
                <span>
                  <span class="block text-sm font-semibold text-gray-900">邮件通知</span>
                  <span class="text-sm text-gray-500">保留旧版本地偏好设置，后续可接入服务端策略。</span>
                </span>
                <input v-model="privacy.emailNotifications" type="checkbox" class="h-5 w-5 rounded border-gray-300 text-blue-600" @change="savePrivacy" />
              </label>
            </div>
          </section>

          <section v-show="activeTab === 'binding'" class="p-4">
            <div class="mb-4 flex items-center justify-between gap-3">
              <div class="flex items-center gap-2">
                <Mail class="h-4 w-4 text-gray-400" />
                <h2 class="text-sm font-semibold text-gray-950">账号绑定</h2>
              </div>
              <button type="button" class="text-xs font-medium text-blue-600 hover:text-blue-700" @click="loadBindings">刷新</button>
            </div>
            <div v-if="loadingBindings" class="py-8 text-center text-sm text-gray-500">
              <Loader2 class="mx-auto mb-2 h-5 w-5 animate-spin" />
              正在加载绑定状态
            </div>
            <div v-else class="space-y-3">
              <div
                v-for="provider in providers"
                :key="provider.key"
                class="flex items-center justify-between gap-4 rounded-lg border p-4"
                :class="provider.supported ? 'border-gray-200 bg-white' : 'border-gray-100 bg-gray-50/70'"
              >
                <div class="flex min-w-0 items-center gap-3">
                  <div
                    class="flex h-11 w-11 shrink-0 items-center justify-center rounded-full border"
                    :class="provider.supported ? 'border-gray-100 bg-white shadow-sm' : 'border-gray-100 bg-gray-100 opacity-60'"
                  >
                    <svg v-if="provider.key === 'github'" class="h-6 w-6 text-gray-950" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
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
                    <h3 class="font-semibold text-gray-950">{{ provider.label }}</h3>
                    <p class="text-sm" :class="provider.supported ? 'text-gray-500' : 'text-gray-400'">
                      {{ provider.supported ? (isBound(provider.key) ? '已连接' : '未连接') : '当前站点暂未开放' }}
                    </p>
                  </div>
                </div>
                <button
                  type="button"
                  class="inline-flex h-9 min-w-24 items-center justify-center gap-2 rounded-md border px-3 text-sm font-semibold disabled:cursor-not-allowed"
                  :class="[
                    !provider.supported
                      ? 'border-gray-200 bg-gray-100 text-gray-400'
                      : isBound(provider.key)
                        ? 'border-red-200 bg-red-50 text-red-600 hover:bg-red-100'
                        : 'border-gray-900 bg-gray-900 text-white hover:bg-gray-800',
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

      <div v-if="cropModalOpen" class="fixed inset-0 z-[100] overflow-y-auto bg-gray-950/50 px-4 py-6 backdrop-blur-sm" role="dialog" aria-modal="true">
        <div class="mx-auto flex min-h-full max-w-[980px] items-center justify-center">
          <div class="w-full overflow-hidden rounded-xl bg-white shadow-2xl">
            <div class="flex items-center justify-between border-b border-gray-100 px-5 py-4">
              <div>
                <h2 class="text-base font-semibold text-gray-950">裁切头像</h2>
                <p class="mt-0.5 text-sm text-gray-500">拖动或缩放裁切框，确认后会自动压缩并优化格式。</p>
              </div>
              <button type="button" class="rounded-md px-2 py-1 text-sm font-medium text-gray-500 hover:bg-gray-100 hover:text-gray-900" @click="closeCropModal">
                关闭
              </button>
            </div>

            <div class="grid gap-4 p-5 lg:grid-cols-[minmax(0,1fr)_170px]">
              <div class="avatar-crop-workspace h-[360px] overflow-hidden rounded-lg border border-gray-100 bg-gray-50">
                <img ref="cropperImage" :src="cropImageUrl" alt="待裁切头像" class="block" />
              </div>

              <aside class="space-y-4">
                <div>
                  <div class="mb-2 text-sm font-semibold text-gray-950">预览</div>
                  <div class="flex h-32 w-32 items-center justify-center overflow-hidden rounded-full border border-gray-100 bg-gray-50">
                    <img v-if="cropPreviewUrl" :src="cropPreviewUrl" alt="头像预览" class="h-full w-full object-cover" />
                  </div>
                </div>
                <div class="rounded-lg bg-gray-50 p-3 text-sm leading-6 text-gray-500">
                  建议让脸部或标识位于圆形中心。最终头像会导出为 400 x 400，并按浏览器支持优先转为 WebP。
                </div>
              </aside>
            </div>

            <div class="flex items-center justify-end gap-2 border-t border-gray-100 bg-gray-50 px-5 py-4">
              <button type="button" class="h-10 rounded-md px-4 text-sm font-medium text-gray-600 hover:bg-gray-100" @click="closeCropModal">
                取消
              </button>
              <button
                type="button"
                class="inline-flex h-10 min-w-28 items-center justify-center gap-2 rounded-md bg-blue-600 px-4 text-sm font-semibold text-white hover:bg-blue-700 disabled:cursor-wait disabled:bg-blue-600"
                :disabled="uploadingAvatar"
                @click="uploadCroppedAvatar"
              >
                <Loader2 v-if="uploadingAvatar" class="h-4 w-4 animate-spin" />
                {{ uploadingAvatar ? '上传中' : '确认上传' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>
</template>
