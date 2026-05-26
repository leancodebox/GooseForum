<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import Draggable from 'vuedraggable'
import { Code, FileText, Globe, Loader2, MailCheck, Plus, Save, Send, Shield, Trash2, Upload } from '@lucide/vue'
import { BasicPage } from '@/admin/components/global-layout'
import AdminLayout from '@/admin/layouts/AdminLayout.vue'
import { Button } from '@/admin/components/ui/button'
import { Badge } from '@/admin/components/ui/badge'
import { Input } from '@/admin/components/ui/input'
import { Textarea } from '@/admin/components/ui/textarea'
import { Switch } from '@/admin/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/admin/components/ui/dialog'
import {
  getAnnouncement,
  getMailSettings,
  getPostingSettings,
  getSecuritySettings,
  getSiteSettings,
  saveAnnouncement,
  saveMailSettings,
  savePostingSettings,
  saveSecuritySettings,
  saveSiteSettings,
  testMailConnection,
} from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import type {
  AdminPayload,
  AnnouncementConfig,
  MailSettings,
  ManageHomeProps,
  PostingSettings,
  SecuritySettings,
  SiteSettings,
} from '@/admin/types'

type Kind = 'site-info' | 'mail' | 'security' | 'posting' | 'announcement'

const props = defineProps<{
  payload: AdminPayload<ManageHomeProps>
  kind: Kind
}>()

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const error = ref('')
const testEmail = ref('')
const newAllowedDomain = ref('')
const newExtension = ref('')
const footerDialog = ref<null | { type: 'link' | 'primary', index: number }>(null)
const footerLinkForm = reactive({ name: '', url: '' })
const footerPrimaryForm = reactive({ content: '' })

const siteForm = reactive<SiteSettings>({
  siteName: '',
  siteUrl: '',
  siteLogo: '',
  siteEmail: '',
  siteDescription: '',
  siteKeywords: '',
  externalLinks: '',
  footerInfo: { primary: [], list: [] },
  brandType: 'default',
  brandText: '',
  brandImage: '',
})

const mailForm = reactive<MailSettings>({
  enableMail: true,
  smtpHost: '',
  smtpPort: 587,
  useSSL: false,
  smtpUsername: '',
  smtpPassword: '',
  fromName: '',
  fromEmail: '',
})

const securityForm = reactive<SecuritySettings>({
  enableSignup: true,
  enableEmailVerification: false,
  allowedDomains: [],
})

const postingForm = reactive<PostingSettings>({
  textControl: {
    minPostLength: 5,
    maxPostLength: 50000,
    minTitleLength: 5,
    maxTitleLength: 100,
    newUserPostCooldownMinutes: 0,
  },
  uploadControl: {
    allowAttachments: true,
    authorizedExtensions: ['.jpg', '.jpeg', '.png', '.gif', '.webp'],
    maxAttachmentSizeKb: 5120,
    maxDailyUploadsPerUser: 10,
    newUserUploadCooldownMinutes: 0,
  },
})

const announcementForm = reactive<AnnouncementConfig>({
  enabled: false,
  content: '',
})

const meta: Record<Kind, { title: string, description: string }> = {
  'site-info': { title: '站点信息', description: '管理论坛的基础信息、SEO 设置、页脚和外部资源。' },
  mail: { title: '邮件设置', description: '配置 SMTP 服务器以发送验证邮件、通知等。' },
  security: { title: '安全与注册', description: '控制用户注册、邮箱验证以及允许的注册邮箱域名。' },
  posting: { title: '发布内容设置', description: '控制帖子标题、内容长度及附件上传规则。' },
  announcement: { title: '系统公告', description: '配置显示在页面顶部的系统公告，正文支持 Markdown 渲染。' },
}

const pageMeta = computed(() => meta[props.kind])
const allowedDomains = computed(() => {
  return securityForm.allowedDomains
})

function toBool(value: unknown, fallback = false) {
  if (typeof value === 'boolean') return value
  if (typeof value === 'number') return value === 1
  if (typeof value === 'string') {
    const normalized = value.trim().toLowerCase()
    if (['true', '1', 'yes', 'on', 'enabled'].includes(normalized)) return true
    if (['false', '0', 'no', 'off', 'disabled', ''].includes(normalized)) return false
  }
  return fallback
}

function normalizeSite(settings: Partial<SiteSettings> = {}) {
  const footerInfo = settings.footerInfo && typeof settings.footerInfo === 'object'
    ? settings.footerInfo
    : { primary: [], list: [] }
  return {
    siteName: settings.siteName ?? '',
    siteUrl: settings.siteUrl ?? '',
    siteLogo: settings.siteLogo ?? '',
    siteEmail: settings.siteEmail ?? '',
    siteDescription: settings.siteDescription ?? '',
    siteKeywords: settings.siteKeywords ?? '',
    externalLinks: settings.externalLinks ?? '',
    footerInfo: {
      primary: Array.isArray(footerInfo.primary)
        ? footerInfo.primary.map(item => ({ content: item?.content ?? '' }))
        : [],
      list: Array.isArray(footerInfo.list)
        ? footerInfo.list.map(item => ({ name: item?.name ?? '', url: item?.url ?? '' }))
        : [],
    },
    brandType: ['default', 'text', 'image'].includes(settings.brandType || '') ? settings.brandType : 'default',
    brandText: settings.brandText ?? '',
    brandImage: settings.brandImage ?? '',
  } satisfies SiteSettings
}

function normalizeMail(settings: Partial<MailSettings> = {}) {
  return {
    enableMail: toBool(settings.enableMail, false),
    smtpHost: settings.smtpHost ?? '',
    smtpPort: Number(settings.smtpPort ?? 587),
    useSSL: toBool(settings.useSSL, false),
    smtpUsername: settings.smtpUsername ?? '',
    smtpPassword: settings.smtpPassword ?? '',
    fromName: settings.fromName ?? '',
    fromEmail: settings.fromEmail ?? '',
  } satisfies MailSettings
}

function normalizeSecurity(settings: Partial<SecuritySettings> = {}) {
  return {
    enableSignup: toBool(settings.enableSignup, true),
    enableEmailVerification: toBool(settings.enableEmailVerification, false),
    allowedDomains: Array.isArray(settings.allowedDomains)
      ? settings.allowedDomains.map(item => String(item).trim().toLowerCase()).filter(Boolean)
      : [],
  } satisfies SecuritySettings
}

function normalizePosting(settings: Partial<PostingSettings> = {}) {
  return {
    textControl: {
      minPostLength: Number(settings.textControl?.minPostLength ?? 5),
      maxPostLength: Number(settings.textControl?.maxPostLength ?? 50000),
      minTitleLength: Number(settings.textControl?.minTitleLength ?? 5),
      maxTitleLength: Number(settings.textControl?.maxTitleLength ?? 100),
      newUserPostCooldownMinutes: Number(settings.textControl?.newUserPostCooldownMinutes ?? 0),
    },
    uploadControl: {
      allowAttachments: toBool(settings.uploadControl?.allowAttachments, true),
      authorizedExtensions: Array.isArray(settings.uploadControl?.authorizedExtensions)
        ? settings.uploadControl.authorizedExtensions.map(item => String(item).trim().toLowerCase()).filter(Boolean)
        : ['.jpg', '.jpeg', '.png', '.gif', '.webp'],
      maxAttachmentSizeKb: Number(settings.uploadControl?.maxAttachmentSizeKb ?? 5120),
      maxDailyUploadsPerUser: Number(settings.uploadControl?.maxDailyUploadsPerUser ?? 10),
      newUserUploadCooldownMinutes: Number(settings.uploadControl?.newUserUploadCooldownMinutes ?? 1440),
    },
  } satisfies PostingSettings
}

function normalizeAnnouncement(settings: Partial<AnnouncementConfig> = {}) {
  return {
    enabled: toBool(settings.enabled, false),
    content: settings.content ?? '',
  } satisfies AnnouncementConfig
}

async function uploadImage(target: 'siteLogo' | 'brandImage', event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const body = new FormData()
  body.append('file', file)
  try {
    const response = await fetch('/file/img-upload', { method: 'POST', body })
    const data = await response.json()
    if (data.code === 0) {
      siteForm[target] = data.result?.url || ''
      adminToast.success('上传成功')
    } else {
      adminToast.error(new Error(data.msg || data.message || '上传失败'), '上传失败')
    }
  } catch (err) {
    adminToast.error(err, '上传失败')
  } finally {
    input.value = ''
  }
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    if (props.kind === 'site-info') Object.assign(siteForm, normalizeSite(await getSiteSettings()))
    else if (props.kind === 'mail') Object.assign(mailForm, normalizeMail(await getMailSettings()))
    else if (props.kind === 'security') Object.assign(securityForm, normalizeSecurity(await getSecuritySettings()))
    else if (props.kind === 'posting') Object.assign(postingForm, normalizePosting(await getPostingSettings()))
    else Object.assign(announcementForm, normalizeAnnouncement(await getAnnouncement()))
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载设置失败'
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    if (props.kind === 'site-info') await saveSiteSettings(normalizeSite(siteForm))
    else if (props.kind === 'mail') await saveMailSettings(normalizeMail(mailForm))
    else if (props.kind === 'security') await saveSecuritySettings(normalizeSecurity(securityForm))
    else if (props.kind === 'posting') await savePostingSettings(normalizePosting(postingForm))
    else await saveAnnouncement(normalizeAnnouncement(announcementForm))
    adminToast.success('保存成功')
  } catch (err) {
    adminToast.error(err, '保存失败')
  } finally {
    saving.value = false
  }
}

async function sendTestMail() {
  if (!testEmail.value.trim()) {
    adminToast.warning('请输入测试邮箱')
    return
  }
  testing.value = true
  try {
    const response = await testMailConnection(normalizeMail(mailForm), testEmail.value.trim())
    adminToast.success(response.result?.message || response.message || response.msg || '测试邮件已发送')
  } catch (err) {
    adminToast.error(err, '发送测试邮件失败')
  } finally {
    testing.value = false
  }
}

function addAllowedDomain() {
  const domain = newAllowedDomain.value.trim().toLowerCase()
  if (!domain || allowedDomains.value.includes(domain)) return
  const next = [...allowedDomains.value, domain]
  securityForm.allowedDomains = next
  newAllowedDomain.value = ''
}

function removeAllowedDomain(domain: string) {
  const next = allowedDomains.value.filter(item => item !== domain)
  securityForm.allowedDomains = next
}

function addExtension() {
  const ext = newExtension.value.trim().toLowerCase()
  if (!ext) return
  if (!ext.startsWith('.')) {
    adminToast.warning('扩展名必须以 . 开头')
    return
  }
  if (!postingForm.uploadControl.authorizedExtensions.includes(ext)) {
    postingForm.uploadControl.authorizedExtensions.push(ext)
  }
  newExtension.value = ''
}

function removeExtension(ext: string) {
  postingForm.uploadControl.authorizedExtensions = postingForm.uploadControl.authorizedExtensions.filter(item => item !== ext)
}

function addFooterPrimary() {
  siteForm.footerInfo ||= { primary: [], list: [] }
  siteForm.footerInfo.primary.push({ content: '' })
  openFooterPrimary(siteForm.footerInfo.primary.length - 1)
}

function addFooterLink() {
  siteForm.footerInfo ||= { primary: [], list: [] }
  siteForm.footerInfo.list.push({ name: '', url: '' })
  openFooterLink(siteForm.footerInfo.list.length - 1)
}

function openFooterLink(index: number) {
  const item = siteForm.footerInfo?.list[index]
  if (!item) return
  footerLinkForm.name = item.name
  footerLinkForm.url = item.url
  footerDialog.value = { type: 'link', index }
}

function openFooterPrimary(index: number) {
  const item = siteForm.footerInfo?.primary[index]
  if (!item) return
  footerPrimaryForm.content = item.content
  footerDialog.value = { type: 'primary', index }
}

function saveFooterItem() {
  if (!footerDialog.value) return
  const { type, index } = footerDialog.value
  if (type === 'link') {
    const item = siteForm.footerInfo?.list[index]
    if (item) {
      item.name = footerLinkForm.name.trim()
      item.url = footerLinkForm.url.trim()
    }
  } else {
    const item = siteForm.footerInfo?.primary[index]
    if (item) {
      item.content = footerPrimaryForm.content.trim()
    }
  }
  footerDialog.value = null
}

function closeFooterDialog(open: boolean) {
  if (!open) footerDialog.value = null
}

function addAnnouncementExample() {
  if (!announcementForm.content) {
    announcementForm.content = '## 维护通知\n\n网站将于今晚 22:00 - 24:00 进行系统维护，期间可能无法访问。'
  }
}

watch(() => props.kind, () => {
  void load()
})

onMounted(load)
</script>

<template>
  <AdminLayout :layout="payload.layout">
    <BasicPage :title="pageMeta.title" :description="pageMeta.description" sticky>
      <template #actions>
        <Button type="button" :disabled="saving" @click="save">
          <Loader2 v-if="saving" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          保存配置
        </Button>
      </template>

      <div v-if="loading" class="flex h-[400px] items-center justify-center">
        <Loader2 class="size-8 animate-spin text-primary" />
      </div>
      <div v-else-if="error" class="rounded-lg border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">{{ error }}</div>

      <form v-else-if="kind === 'site-info'" class="space-y-10" @submit.prevent="save">
        <div class="grid gap-10 md:grid-cols-2">
          <section class="space-y-6">
            <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium"><Globe class="size-5 text-muted-foreground" />基本信息</div>
            <label class="grid gap-2 text-sm font-medium">站点名称<Input v-model="siteForm.siteName" placeholder="GooseForum" /></label>
            <label class="grid gap-2 text-sm font-medium">站点 URL<Input v-model="siteForm.siteUrl" placeholder="https://example.com" /></label>
            <label class="grid gap-2 text-sm font-medium">联系邮箱<Input v-model="siteForm.siteEmail" placeholder="contact@example.com" /></label>
            <label class="grid gap-2 text-sm font-medium">站点 Logo
              <div class="flex items-start gap-4">
                <div class="grid size-20 shrink-0 place-items-center overflow-hidden rounded-lg border bg-muted">
                  <img v-if="siteForm.siteLogo" :src="siteForm.siteLogo" class="size-full object-cover" alt="Logo" />
                  <Upload v-else class="size-8 text-muted-foreground/50" />
                </div>
                <div class="grid flex-1 gap-2">
                  <Input v-model="siteForm.siteLogo" placeholder="Logo URL" />
                  <label class="inline-flex h-9 w-fit cursor-pointer items-center gap-2 rounded-md border bg-background px-3 text-sm font-medium shadow-xs hover:bg-accent">
                    上传图片
                    <input class="hidden" type="file" accept="image/*" @change="uploadImage('siteLogo', $event)" />
                  </label>
                </div>
              </div>
            </label>
            <div class="space-y-4">
              <div class="flex items-center gap-2 border-b pb-2 text-base font-medium">品牌标识</div>
              <div class="flex flex-wrap gap-4 text-sm">
                <label class="flex items-center gap-2"><input v-model="siteForm.brandType" type="radio" value="default" />默认样式</label>
                <label class="flex items-center gap-2"><input v-model="siteForm.brandType" type="radio" value="text" />自定义文字</label>
                <label class="flex items-center gap-2"><input v-model="siteForm.brandType" type="radio" value="image" />图片</label>
              </div>
              <label v-if="siteForm.brandType === 'text'" class="grid gap-2 text-sm font-medium">品牌文字<Input v-model="siteForm.brandText" placeholder="MyBrand" /></label>
              <label v-if="siteForm.brandType === 'image'" class="grid gap-2 text-sm font-medium">品牌图片
                <div class="flex gap-2">
                  <Input v-model="siteForm.brandImage" placeholder="Brand Image URL" />
                  <label class="inline-flex size-9 shrink-0 cursor-pointer items-center justify-center rounded-md border bg-background shadow-xs hover:bg-accent">
                    <Upload class="size-4" />
                    <input class="hidden" type="file" accept="image/*" @change="uploadImage('brandImage', $event)" />
                  </label>
                </div>
              </label>
            </div>
          </section>
          <section class="space-y-6">
            <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium"><FileText class="size-5 text-muted-foreground" />SEO 与资源</div>
            <label class="grid gap-2 text-sm font-medium">站点描述<Textarea v-model="siteForm.siteDescription" class="min-h-24" /></label>
            <label class="grid gap-2 text-sm font-medium">关键词<Input v-model="siteForm.siteKeywords" placeholder="forum, community" /></label>
            <label class="grid gap-2 text-sm font-medium">外部资源链接 / Meta 标签<Textarea v-model="siteForm.externalLinks" class="min-h-28 font-mono text-xs" /></label>
            <div class="space-y-4 border-y py-4">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <div class="font-semibold">页脚内容 (Footer)</div>
                  <p class="text-xs text-muted-foreground">这里尽量按照前台侧栏页脚的两行展示方式编辑。</p>
                </div>
                <div class="flex gap-2">
                  <Button variant="ghost" size="sm" type="button" @click="addFooterLink"><Plus class="size-4" />添加链接</Button>
                  <Button variant="ghost" size="sm" type="button" @click="addFooterPrimary"><Plus class="size-4" />添加文字</Button>
                </div>
              </div>
              <div class="rounded-lg border bg-background px-3 py-3 text-xs leading-5 text-muted-foreground shadow-xs">
                <div class="space-y-1.5">
                  <Draggable
                    v-model="siteForm.footerInfo!.list"
                    item-key="name"
                    direction="horizontal"
                    handle=".js-footer-handle"
                    class="flex min-h-5 flex-wrap items-center gap-x-3 gap-y-1"
                    ghost-class="opacity-40"
                  >
                    <template #item="{ element: item, index }">
                      <div class="group inline-flex min-h-5 items-center rounded text-muted-foreground transition-colors hover:bg-muted/60 hover:text-primary">
                        <span class="js-footer-handle -ml-1 mr-0.5 cursor-grab text-muted-foreground/45 opacity-0 transition-opacity group-hover:opacity-100 active:cursor-grabbing">⋮⋮</span>
                        <button class="inline-flex min-h-5 items-center rounded px-0.5 text-left text-xs" type="button" @click="openFooterLink(index)">
                          {{ item.name || '未命名链接' }}
                        </button>
                        <Button variant="ghost" size="icon" class="ml-0.5 size-5 rounded-sm opacity-0 hover:bg-destructive/10 hover:text-destructive group-hover:opacity-100" type="button" @click="siteForm.footerInfo?.list.splice(index, 1)">
                          <Trash2 class="size-3.5" />
                        </Button>
                      </div>
                    </template>
                    <template #footer>
                      <div v-if="siteForm.footerInfo?.list.length === 0" class="min-h-5 text-muted-foreground/70">暂无链接</div>
                    </template>
                  </Draggable>

                  <Draggable
                    v-model="siteForm.footerInfo!.primary"
                    item-key="content"
                    handle=".js-footer-handle"
                    class="mt-1 flex min-h-5 flex-wrap items-center gap-x-3 gap-y-1 text-muted-foreground"
                    ghost-class="opacity-40"
                  >
                    <template #item="{ element: item, index }">
                      <div class="group inline-flex min-h-5 max-w-full items-center rounded transition-colors hover:bg-muted/60">
                        <span class="js-footer-handle -ml-1 mr-0.5 cursor-grab text-muted-foreground/45 opacity-0 transition-opacity group-hover:opacity-100 active:cursor-grabbing">⋮⋮</span>
                        <button class="inline-flex min-h-5 items-center rounded px-0.5 text-left text-xs text-muted-foreground" type="button" @click="openFooterPrimary(index)">
                          {{ item.content || '空文字内容' }}
                        </button>
                        <Button variant="ghost" size="icon" class="ml-0.5 size-5 rounded-sm opacity-0 hover:bg-destructive/10 hover:text-destructive group-hover:opacity-100" type="button" @click="siteForm.footerInfo?.primary.splice(index, 1)">
                          <Trash2 class="size-3.5" />
                        </Button>
                      </div>
                    </template>
                    <template #footer>
                      <div v-if="siteForm.footerInfo?.primary.length === 0" class="min-h-5 text-muted-foreground/70">暂无内容</div>
                    </template>
                  </Draggable>
                </div>
              </div>
            </div>
          </section>
        </div>
      </form>

      <form v-else-if="kind === 'mail'" class="grid gap-10 lg:grid-cols-[minmax(0,2fr)_minmax(260px,1fr)]" @submit.prevent="save">
        <section class="space-y-6">
          <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium">SMTP 服务器设置</div>
          <div class="flex items-center justify-between rounded-lg border bg-muted/20 p-4">
            <div><div class="font-medium">启用邮件服务</div><p class="text-sm text-muted-foreground">开启后，系统将能够发送验证码、通知等邮件。</p></div>
            <Switch v-model="mailForm.enableMail" />
          </div>
          <div class="grid gap-6 md:grid-cols-2">
            <label class="grid gap-2 text-sm font-medium">SMTP 主机<Input v-model="mailForm.smtpHost" :disabled="!mailForm.enableMail" placeholder="smtp.example.com" /></label>
            <label class="grid gap-2 text-sm font-medium">SMTP 端口<Input v-model.number="mailForm.smtpPort" :disabled="!mailForm.enableMail" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">用户名 (邮箱)<Input v-model="mailForm.smtpUsername" :disabled="!mailForm.enableMail" /></label>
            <label class="grid gap-2 text-sm font-medium">密码 / 授权码<Input v-model="mailForm.smtpPassword" :disabled="!mailForm.enableMail" type="password" /></label>
          </div>
          <div class="flex items-center justify-between rounded-lg border bg-muted/20 p-4">
            <div><div class="flex items-center gap-2 font-medium"><Shield class="size-4" />使用 SSL 加密</div><p class="text-sm text-muted-foreground">通常端口 465 需要开启 SSL，587 通常使用 STARTTLS。</p></div>
            <Switch v-model="mailForm.useSSL" :disabled="!mailForm.enableMail" />
          </div>
        </section>
        <aside class="space-y-8">
          <section class="space-y-4">
            <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium"><Send class="size-5 text-muted-foreground" />发件人信息</div>
            <label class="grid gap-2 text-sm font-medium">发件人名称<Input v-model="mailForm.fromName" :disabled="!mailForm.enableMail" placeholder="GooseForum" /></label>
            <label class="grid gap-2 text-sm font-medium">发件人邮箱<Input v-model="mailForm.fromEmail" :disabled="!mailForm.enableMail" placeholder="noreply@example.com" /></label>
          </section>
          <section class="space-y-4">
            <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium">发送测试</div>
            <div class="space-y-3 rounded-lg border p-4">
              <p class="text-sm text-muted-foreground">保存配置前，可以发送一封测试邮件验证 SMTP 设置。</p>
              <Input v-model="testEmail" :disabled="!mailForm.enableMail" placeholder="输入接收测试的邮箱" />
              <Button class="w-full" type="button" variant="secondary" :disabled="testing || !mailForm.enableMail" @click="sendTestMail">
                <Loader2 v-if="testing" class="size-4 animate-spin" />
                <Send v-else class="size-4" />
                发送测试邮件
              </Button>
            </div>
          </section>
        </aside>
      </form>

      <form v-else-if="kind === 'security'" class="max-w-2xl space-y-8" @submit.prevent="save">
        <div class="flex items-center justify-between">
          <div><div class="text-base font-medium">允许新用户注册</div><p class="text-sm text-muted-foreground">关闭后，前台将不允许创建新账号。</p></div>
          <Switch v-model="securityForm.enableSignup" />
        </div>
        <div class="flex items-center justify-between">
          <div><div class="flex items-center gap-2 text-base font-medium"><MailCheck class="size-4" />要求验证邮箱</div><p class="text-sm text-muted-foreground">新用户完成邮箱验证后才能正常激活和使用账号。</p></div>
          <Switch v-model="securityForm.enableEmailVerification" />
        </div>
        <div class="space-y-4">
          <div><div class="text-base font-medium">允许注册的邮箱域名</div><p class="text-sm text-muted-foreground">留空表示不限制。配置后，仅允许这些邮箱域名完成注册。</p></div>
          <div class="flex gap-2">
            <Input v-model="newAllowedDomain" class="max-w-sm" placeholder="例如: gmail.com" @keydown.enter.prevent="addAllowedDomain" />
            <Button type="button" variant="secondary" @click="addAllowedDomain"><Plus class="size-4" />添加</Button>
          </div>
          <div class="flex flex-wrap gap-2">
            <span v-if="allowedDomains.length === 0" class="text-sm italic text-muted-foreground">未限制域名</span>
            <Badge v-for="domain in allowedDomains" :key="domain" variant="secondary" class="gap-2 px-3 py-1.5 text-sm font-normal">
              {{ domain }}
              <button type="button" @click="removeAllowedDomain(domain)"><Trash2 class="size-3.5 text-muted-foreground hover:text-destructive" /></button>
            </Badge>
          </div>
        </div>
      </form>

      <form v-else-if="kind === 'posting'" class="grid gap-12 lg:grid-cols-2" @submit.prevent="save">
        <section class="space-y-6">
          <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium"><FileText class="size-5 text-muted-foreground" />文本内容控制</div>
          <div class="grid gap-6 sm:grid-cols-2">
            <label class="grid gap-2 text-sm font-medium">最小标题长度<Input v-model.number="postingForm.textControl.minTitleLength" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">最大标题长度<Input v-model.number="postingForm.textControl.maxTitleLength" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">最小正文长度<Input v-model.number="postingForm.textControl.minPostLength" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">最大正文长度<Input v-model.number="postingForm.textControl.maxPostLength" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">新用户发帖冷却 (分钟)<Input v-model.number="postingForm.textControl.newUserPostCooldownMinutes" type="number" /></label>
          </div>
        </section>
        <section class="space-y-6">
          <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium"><Upload class="size-5 text-muted-foreground" />附件控制</div>
          <div class="flex items-center justify-between rounded-lg border bg-muted/10 p-4">
            <div><div class="text-base font-medium">允许上传附件</div><p class="text-sm text-muted-foreground">开启后，用户可以在帖子中上传图片或文件。</p></div>
            <Switch v-model="postingForm.uploadControl.allowAttachments" />
          </div>
          <div class="grid gap-6 sm:grid-cols-2">
            <label class="grid gap-2 text-sm font-medium">每日最大上传数量<Input v-model.number="postingForm.uploadControl.maxDailyUploadsPerUser" :disabled="!postingForm.uploadControl.allowAttachments" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">单个附件大小限制 (KB)<Input v-model.number="postingForm.uploadControl.maxAttachmentSizeKb" :disabled="!postingForm.uploadControl.allowAttachments" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">新用户上传冷却 (分钟)<Input v-model.number="postingForm.uploadControl.newUserUploadCooldownMinutes" :disabled="!postingForm.uploadControl.allowAttachments" type="number" /></label>
          </div>
          <div class="space-y-3">
            <div class="text-sm font-medium">允许的扩展名</div>
            <div class="flex gap-2">
              <Input v-model="newExtension" placeholder="例如: .jpg" @keydown.enter.prevent="addExtension" />
              <Button type="button" variant="secondary" @click="addExtension"><Plus class="size-4" />添加</Button>
            </div>
            <div class="flex flex-wrap gap-2">
              <Badge v-for="ext in postingForm.uploadControl.authorizedExtensions" :key="ext" variant="secondary" class="gap-2 px-3 py-1.5 text-sm font-normal">
                {{ ext }}
                <button type="button" @click="removeExtension(ext)"><Trash2 class="size-3.5 text-muted-foreground hover:text-destructive" /></button>
              </Badge>
            </div>
          </div>
        </section>
      </form>

      <form v-else class="max-w-3xl space-y-6" @submit.prevent="save">
        <div class="flex items-center justify-between">
          <div><div class="text-base font-medium">启用公告</div><p class="text-sm text-muted-foreground">开启后，系统公告将显示在首页顶部。</p></div>
          <Switch v-model="announcementForm.enabled" />
        </div>
        <label class="grid gap-2 text-sm font-medium">
          公告内容
          <span class="text-sm font-normal text-muted-foreground">支持 Markdown 语法，保存后会渲染为富文本公告内容。</span>
          <Textarea v-model="announcementForm.content" class="min-h-64 resize-y font-mono text-sm" placeholder="例如：## 维护通知" />
        </label>
        <Button variant="outline" type="button" @click="addAnnouncementExample"><Code class="size-4" />填入示例</Button>
      </form>

      <Dialog v-if="kind === 'site-info'" :open="footerDialog !== null" @update:open="closeFooterDialog">
        <DialogContent class="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{{ footerDialog?.type === 'link' ? '编辑页脚链接' : '编辑页脚文字' }}</DialogTitle>
            <DialogDescription>
              页脚会按前台侧栏的两行样式展示，第一行是链接，第二行是文字。
            </DialogDescription>
          </DialogHeader>
          <form v-if="footerDialog?.type === 'link'" class="grid gap-4" @submit.prevent="saveFooterItem">
            <label class="grid gap-2 text-sm font-medium">
              链接名称
              <Input v-model="footerLinkForm.name" placeholder="Github" />
            </label>
            <label class="grid gap-2 text-sm font-medium">
              链接地址
              <Input v-model="footerLinkForm.url" placeholder="https://example.com" />
            </label>
            <DialogFooter>
              <Button variant="outline" type="button" @click="footerDialog = null">取消</Button>
              <Button type="submit">保存</Button>
            </DialogFooter>
          </form>
          <form v-else class="grid gap-4" @submit.prevent="saveFooterItem">
            <label class="grid gap-2 text-sm font-medium">
              文字内容
              <Input v-model="footerPrimaryForm.content" placeholder="Powered by GooseForum" />
            </label>
            <DialogFooter>
              <Button variant="outline" type="button" @click="footerDialog = null">取消</Button>
              <Button type="submit">保存</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </BasicPage>
  </AdminLayout>
</template>
