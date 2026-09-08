<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { Check, CircleAlert, Copy, KeyRound, Loader2, Pencil, Plus, RefreshCw, RotateCw } from '@lucide/vue'
import { BasicPage } from '@/admin/components/global-layout'
import AdminConfirmDialog from '@/admin/components/AdminConfirmDialog.vue'
import { Badge } from '@/admin/components/ui/badge'
import { Button } from '@/admin/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/admin/components/ui/dialog'
import { Input } from '@/admin/components/ui/input'
import { Label } from '@/admin/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/admin/components/ui/select'
import { Switch } from '@/admin/components/ui/switch'
import { Textarea } from '@/admin/components/ui/textarea'
import { createOIDCClient, getOIDCClients, getOIDCProviderStatus, rotateOIDCClientSecret, rotateOIDCSigningKey, saveOIDCProviderSettings, updateOIDCClient } from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import type { OIDCClient, OIDCClientAuthMethod, OIDCClientInput, OIDCProviderStatus } from '@/admin/types'

const supportedScopes = ['openid', 'profile', 'email', 'offline_access'] as const
const clients = ref<OIDCClient[]>([])
const loading = ref(false)
const saving = ref(false)
const togglingId = ref('')
const error = ref('')
const editorOpen = ref(false)
const rotateTarget = ref<OIDCClient | null>(null)
const rotating = ref(false)
const secretOpen = ref(false)
const revealedSecret = ref('')
const secretClientId = ref('')
const copied = ref(false)
const providerStatus = ref<OIDCProviderStatus | null>(null)
const providerSaving = ref(false)
const signingKeyDialogOpen = ref(false)
const signingKeyRotating = ref(false)

interface ClientForm {
  clientId: string
  name: string
  redirectUris: string
  scopes: string[]
  tokenEndpointAuthMethod: OIDCClientAuthMethod
  requirePkce: boolean
  public: boolean
  enabled: boolean
}

const form = reactive<ClientForm>(emptyForm())
const editing = computed(() => Boolean(form.clientId))

function emptyForm(): ClientForm {
  return {
    clientId: '',
    name: '',
    redirectUris: '',
    scopes: ['openid', 'profile', 'email'],
    tokenEndpointAuthMethod: 'client_secret_basic',
    requirePkce: true,
    public: false,
    enabled: true,
  }
}

function resetForm(client?: OIDCClient) {
  Object.assign(form, client ? {
    clientId: client.clientId,
    name: client.name,
    redirectUris: client.redirectUris.join('\n'),
    scopes: [...client.scopes],
    tokenEndpointAuthMethod: client.tokenEndpointAuthMethod,
    requirePkce: client.requirePkce,
    public: client.public,
    enabled: client.enabled,
  } : emptyForm())
}

async function loadClients() {
  loading.value = true
  error.value = ''
  try {
    clients.value = await getOIDCClients()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载 OIDC 客户端失败'
  } finally {
    loading.value = false
  }
}

async function loadProviderStatus() {
  try {
    providerStatus.value = await getOIDCProviderStatus()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载 OIDC Provider 状态失败'
  }
}

async function toggleProvider(enabled: boolean) {
  if (providerSaving.value) return
  providerSaving.value = true
  try {
    providerStatus.value = await saveOIDCProviderSettings(enabled)
    if (providerStatus.value.available) adminToast.success('OIDC Provider 已启用')
    else if (!providerStatus.value.enabled) adminToast.success('OIDC Provider 已停用')
    else adminToast.warning(providerStatus.value.error || 'OIDC Provider 配置尚未就绪')
  } catch (err) {
    adminToast.error(err, '保存 OIDC Provider 设置失败')
  } finally {
    providerSaving.value = false
  }
}

async function confirmSigningKeyRotation() {
  if (signingKeyRotating.value) return
  signingKeyRotating.value = true
  try {
    providerStatus.value = await rotateOIDCSigningKey()
    signingKeyDialogOpen.value = false
    adminToast.success('OIDC 签名密钥已轮换')
  } catch (err) {
    adminToast.error(err, '轮换 OIDC 签名密钥失败')
  } finally {
    signingKeyRotating.value = false
  }
}

function openCreate() {
  resetForm()
  editorOpen.value = true
}

function openEdit(client: OIDCClient) {
  resetForm(client)
  editorOpen.value = true
}

function setPublic(value: boolean) {
  form.public = value
  if (value) {
    form.tokenEndpointAuthMethod = 'none'
    form.requirePkce = true
  } else if (form.tokenEndpointAuthMethod === 'none') {
    form.tokenEndpointAuthMethod = 'client_secret_basic'
  }
}

function toggleScope(scope: string, enabled: boolean) {
  if (scope === 'openid') return
  form.scopes = enabled
    ? Array.from(new Set([...form.scopes, scope]))
    : form.scopes.filter(item => item !== scope)
}

function toInput(): OIDCClientInput {
  const scopes = Array.from(new Set(['openid', ...form.scopes]))
  return {
    name: form.name.trim(),
    redirectUris: Array.from(new Set(form.redirectUris.split(/\r?\n/).map(value => value.trim()).filter(Boolean))),
    scopes,
    grantTypes: scopes.includes('offline_access') ? ['authorization_code', 'refresh_token'] : ['authorization_code'],
    tokenEndpointAuthMethod: form.public ? 'none' : form.tokenEndpointAuthMethod,
    requirePkce: form.public || form.requirePkce,
    public: form.public,
    enabled: form.enabled,
  }
}

function validateInput(input: OIDCClientInput) {
  if (!input.name) return '请输入应用名称'
  if (!input.redirectUris.length) return '请至少填写一个回调地址'
  if (input.redirectUris.some(uri => !/^https?:\/\//i.test(uri))) return '回调地址必须是完整的 HTTP(S) URL'
  return ''
}

async function submit() {
  if (saving.value) return
  const input = toInput()
  const message = validateInput(input)
  if (message) {
    adminToast.warning(message)
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      const updated = await updateOIDCClient({ ...input, clientId: form.clientId })
      clients.value = clients.value.map(client => client.clientId === updated.clientId ? updated : client)
      editorOpen.value = false
      adminToast.success('OIDC 客户端已保存')
    } else {
      const credentials = await createOIDCClient(input)
      clients.value = [...clients.value, credentials.client]
      editorOpen.value = false
      showSecret(credentials.client.clientId, credentials.clientSecret || '')
    }
  } catch (err) {
    adminToast.error(err, editing.value ? '保存 OIDC 客户端失败' : '创建 OIDC 客户端失败')
  } finally {
    saving.value = false
  }
}

async function toggleEnabled(client: OIDCClient, enabled: boolean) {
  if (togglingId.value) return
  togglingId.value = client.clientId
  try {
    const updated = await updateOIDCClient({ ...client, enabled })
    clients.value = clients.value.map(item => item.clientId === updated.clientId ? updated : item)
  } catch (err) {
    adminToast.error(err, '更新客户端状态失败')
  } finally {
    togglingId.value = ''
  }
}

async function confirmRotate() {
  if (!rotateTarget.value || rotating.value) return
  rotating.value = true
  const clientId = rotateTarget.value.clientId
  try {
    const credentials = await rotateOIDCClientSecret(clientId)
    clients.value = clients.value.map(client => client.clientId === credentials.client.clientId ? credentials.client : client)
    rotateTarget.value = null
    showSecret(credentials.client.clientId, credentials.clientSecret || '')
  } catch (err) {
    adminToast.error(err, '重置客户端密钥失败')
  } finally {
    rotating.value = false
  }
}

function showSecret(clientId: string, secret: string) {
  secretClientId.value = clientId
  revealedSecret.value = secret
  copied.value = false
  secretOpen.value = true
}

function clearSecret() {
  revealedSecret.value = ''
  secretClientId.value = ''
  copied.value = false
}

function setSecretOpen(open: boolean) {
  secretOpen.value = open
  if (!open) clearSecret()
}

async function copySecret() {
  if (!revealedSecret.value) return
  await navigator.clipboard.writeText(revealedSecret.value)
  copied.value = true
}

onMounted(() => void Promise.all([loadProviderStatus(), loadClients()]))
onBeforeUnmount(clearSecret)
</script>

<template>
  <BasicPage title="OIDC Provider" description="管理使用 GooseForum 统一登录的应用。">
    <template #primary-action>
      <Button type="button" @click="openCreate">
        <Plus class="size-4" />
        新建客户端
      </Button>
    </template>

    <section class="mb-4 flex flex-col gap-3 border-y bg-muted/20 px-4 py-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="min-w-0">
        <div class="flex items-center gap-2">
          <span class="text-sm font-semibold">服务状态</span>
          <Badge v-if="providerStatus" :variant="providerStatus.available ? 'default' : 'outline'">
            {{ providerStatus.available ? '运行中' : providerStatus.enabled ? '配置错误' : '未启用' }}
          </Badge>
        </div>
        <p v-if="providerStatus?.issuer" class="mt-1 truncate font-mono text-xs text-muted-foreground">{{ providerStatus.issuer }}</p>
        <p v-else-if="providerStatus?.error" class="mt-1 flex items-start gap-1.5 text-xs text-destructive">
          <CircleAlert class="mt-0.5 size-3.5 shrink-0" />
          <span>{{ providerStatus.error }}</span>
        </p>
        <p v-else class="mt-1 text-xs text-muted-foreground">启用前请先在站点设置中配置规范站点地址。</p>
      </div>
      <div class="flex shrink-0 flex-wrap items-center gap-3">
        <Button type="button" size="sm" variant="outline" :disabled="!providerStatus?.available || signingKeyRotating" @click="signingKeyDialogOpen = true">
          <RotateCw class="size-4" />
          轮换签名密钥
        </Button>
        <label class="flex items-center gap-2 text-sm">
          启用 Provider
          <Switch :model-value="providerStatus?.enabled ?? false" :disabled="!providerStatus || providerSaving" @update:model-value="toggleProvider(Boolean($event))" />
        </label>
      </div>
    </section>

    <div v-if="error" class="mb-4 flex items-center justify-between gap-4 rounded-md border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">
      <span>{{ error }}</span>
      <Button type="button" size="sm" variant="outline" @click="loadClients">重试</Button>
    </div>

    <section class="overflow-hidden rounded-lg border bg-background">
      <header class="flex items-center justify-between gap-3 border-b bg-muted/20 px-4 py-3">
        <div>
          <h2 class="text-sm font-semibold">已注册应用</h2>
          <p class="mt-0.5 text-xs text-muted-foreground">{{ clients.length }} 个客户端</p>
        </div>
        <Button type="button" size="icon-sm" variant="ghost" :disabled="loading" title="刷新" @click="loadClients">
          <RefreshCw class="size-4" :class="loading ? 'animate-spin' : ''" />
        </Button>
      </header>

      <div v-if="loading && !clients.length" class="grid min-h-64 place-items-center">
        <Loader2 class="size-7 animate-spin text-muted-foreground" />
      </div>
      <div v-else-if="!clients.length" class="grid min-h-64 place-items-center px-6 text-center">
        <div>
          <KeyRound class="mx-auto size-8 text-muted-foreground" />
          <p class="mt-3 text-sm font-medium">还没有 OIDC 客户端</p>
          <p class="mt-1 text-xs text-muted-foreground">创建客户端后，外部应用即可使用 GooseForum 登录。</p>
        </div>
      </div>
      <div v-else class="divide-y">
        <article v-for="client in clients" :key="client.clientId" class="grid gap-4 px-4 py-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h3 class="truncate text-sm font-semibold">{{ client.name }}</h3>
              <Badge variant="secondary">{{ client.public ? 'Public' : 'Confidential' }}</Badge>
              <Badge :variant="client.enabled ? 'default' : 'outline'">{{ client.enabled ? '已启用' : '已停用' }}</Badge>
            </div>
            <p class="mt-1 truncate font-mono text-xs text-muted-foreground">{{ client.clientId }}</p>
            <div class="mt-2 flex flex-wrap gap-1.5">
              <Badge v-for="scope in client.scopes" :key="scope" variant="outline" class="font-mono text-[11px] font-normal">{{ scope }}</Badge>
              <span class="text-xs text-muted-foreground">{{ client.redirectUris.length }} 个回调地址</span>
            </div>
          </div>
          <div class="flex items-center gap-1 md:justify-end">
            <label class="mr-2 flex items-center gap-2 text-xs text-muted-foreground">
              启用
              <Switch :model-value="client.enabled" :disabled="Boolean(togglingId)" @update:model-value="toggleEnabled(client, Boolean($event))" />
            </label>
            <Button v-if="!client.public" type="button" size="icon-sm" variant="ghost" title="重置客户端密钥" @click="rotateTarget = client">
              <RotateCw class="size-4" />
            </Button>
            <Button type="button" size="icon-sm" variant="ghost" title="编辑" @click="openEdit(client)">
              <Pencil class="size-4" />
            </Button>
          </div>
        </article>
      </div>
    </section>

    <Dialog v-model:open="editorOpen">
      <DialogContent class="max-h-[calc(100vh-2rem)] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{{ editing ? '编辑 OIDC 客户端' : '新建 OIDC 客户端' }}</DialogTitle>
          <DialogDescription>客户端类型创建后不可更改；每行填写一个回调地址。</DialogDescription>
        </DialogHeader>

        <form id="oidc-client-form" class="grid gap-5" @submit.prevent="submit">
          <div class="grid gap-2">
            <Label for="oidc-client-name">应用名称</Label>
            <Input id="oidc-client-name" v-model="form.name" maxlength="255" placeholder="Internal Wiki" autofocus />
          </div>

          <div class="grid gap-2">
            <Label for="oidc-client-redirects">回调地址</Label>
            <Textarea id="oidc-client-redirects" v-model="form.redirectUris" class="min-h-24 font-mono text-xs" placeholder="https://wiki.example.com/oauth/callback" />
          </div>

          <fieldset class="grid gap-3">
            <legend class="text-sm font-medium">客户端类型</legend>
            <div class="grid gap-2 sm:grid-cols-2">
              <button type="button" class="rounded-md border p-3 text-left text-sm transition-colors" :class="!form.public ? 'border-primary bg-primary/5' : 'hover:bg-muted/50'" :disabled="editing" @click="setPublic(false)">
                <span class="font-medium">Confidential</span>
                <span class="mt-1 block text-xs text-muted-foreground">服务端应用，使用客户端密钥认证。</span>
              </button>
              <button type="button" class="rounded-md border p-3 text-left text-sm transition-colors" :class="form.public ? 'border-primary bg-primary/5' : 'hover:bg-muted/50'" :disabled="editing" @click="setPublic(true)">
                <span class="font-medium">Public</span>
                <span class="mt-1 block text-xs text-muted-foreground">浏览器或原生应用，强制 PKCE。</span>
              </button>
            </div>
          </fieldset>

          <div class="grid gap-4 sm:grid-cols-2">
            <div class="grid gap-2">
              <Label>Token 端点认证</Label>
              <Select v-model="form.tokenEndpointAuthMethod" :disabled="form.public">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="client_secret_basic">client_secret_basic</SelectItem>
                  <SelectItem value="client_secret_post">client_secret_post</SelectItem>
                  <SelectItem v-if="form.public" value="none">none</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="flex items-center justify-between gap-4 rounded-md border px-3 py-2.5">
              <div>
                <p class="text-sm font-medium">要求 PKCE</p>
                <p class="text-xs text-muted-foreground">Public 客户端始终开启。</p>
              </div>
              <Switch :model-value="form.public || form.requirePkce" :disabled="form.public" @update:model-value="form.requirePkce = Boolean($event)" />
            </div>
          </div>

          <fieldset class="grid gap-2">
            <legend class="text-sm font-medium">允许的 Scope</legend>
            <div class="grid gap-2 sm:grid-cols-2">
              <label v-for="scope in supportedScopes" :key="scope" class="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
                <span class="font-mono text-xs">{{ scope }}</span>
                <Switch :model-value="form.scopes.includes(scope)" :disabled="scope === 'openid'" @update:model-value="toggleScope(scope, Boolean($event))" />
              </label>
            </div>
          </fieldset>

          <label class="flex items-center justify-between gap-4 rounded-md border px-3 py-2.5">
            <span>
              <span class="block text-sm font-medium">启用客户端</span>
              <span class="block text-xs text-muted-foreground">停用后将拒绝新的授权和 Token 请求。</span>
            </span>
            <Switch v-model="form.enabled" />
          </label>
        </form>

        <DialogFooter>
          <Button type="button" variant="outline" @click="editorOpen = false">取消</Button>
          <Button type="submit" form="oidc-client-form" :disabled="saving">
            <Loader2 v-if="saving" class="size-4 animate-spin" />
            {{ saving ? '保存中' : '保存' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <AdminConfirmDialog
      :open="signingKeyDialogOpen"
      :loading="signingKeyRotating"
      title="轮换 OIDC 签名密钥"
      description="新 ID Token 将立即使用新密钥签名；旧公钥会保留到所有已签发 ID Token 过期。"
      confirm-text="轮换密钥"
      @update:open="signingKeyDialogOpen = $event"
      @confirm="confirmSigningKeyRotation"
    />

    <AdminConfirmDialog
      :open="Boolean(rotateTarget)"
      :loading="rotating"
      title="重置客户端密钥"
      description="现有密钥会立即失效，依赖此客户端的应用需要同步更新。"
      confirm-text="重置密钥"
      @update:open="!$event && (rotateTarget = null)"
      @confirm="confirmRotate"
    />

    <Dialog :open="secretOpen" @update:open="setSecretOpen">
      <DialogContent class="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>客户端密钥</DialogTitle>
          <DialogDescription>该密钥只显示这一次。关闭后无法再次查看，只能重新生成。</DialogDescription>
        </DialogHeader>
        <div class="grid gap-3">
          <div>
            <p class="mb-1 text-xs font-medium text-muted-foreground">Client ID</p>
            <code class="block overflow-x-auto rounded-md border bg-muted/40 p-3 text-xs">{{ secretClientId }}</code>
          </div>
          <div>
            <p class="mb-1 text-xs font-medium text-muted-foreground">Client Secret</p>
            <div class="flex min-w-0 gap-2">
              <Input :model-value="revealedSecret" readonly class="min-w-0 font-mono text-xs" />
              <Button type="button" size="icon" variant="outline" :title="copied ? '已复制' : '复制密钥'" @click="copySecret">
                <Check v-if="copied" class="size-4" />
                <Copy v-else class="size-4" />
              </Button>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button type="button" @click="setSecretOpen(false)">我已妥善保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </BasicPage>
</template>
