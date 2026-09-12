<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { Check, CircleAlert, Copy, Loader2, Pencil, Plus, RefreshCw, RotateCw } from '@lucide/vue'
import { BasicPage } from '@/admin/components/global-layout'
import { adminText } from '@/admin/runtime/i18n-text'
import { useI18n } from 'vue-i18n'
import AdminSection from '@/admin/components/AdminSection.vue'
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

const { t } = useI18n()

const supportedScopes = ['openid', 'profile', 'email', 'offline_access'] as const
const clients = ref<OIDCClient[]>([])
const loading = ref(false)
const saving = ref(false)
const togglingId = ref('')
const clientError = ref('')
const providerError = ref('')
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
  clientError.value = ''
  try {
    clients.value = await getOIDCClients()
  } catch (err) {
    clientError.value = err instanceof Error ? err.message : adminText('oidcLoadFailed')
  } finally {
    loading.value = false
  }
}

async function loadProviderStatus() {
  providerError.value = ''
  try {
    providerStatus.value = await getOIDCProviderStatus()
  } catch (err) {
    providerError.value = err instanceof Error ? err.message : adminText('oidcStatusFailed')
  }
}

async function toggleProvider(enabled: boolean) {
  if (providerSaving.value) return
  providerSaving.value = true
  try {
    providerStatus.value = await saveOIDCProviderSettings(enabled)
    if (providerStatus.value.available) adminToast.success(adminText('oidcEnabledToast'))
    else if (!providerStatus.value.enabled) adminToast.success(adminText('oidcDisabledToast'))
    else adminToast.warning(providerStatus.value.error || adminText('oidcNotReady'))
  } catch (err) {
    adminToast.error(err, adminText('oidcSaveProviderFailed'))
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
    adminToast.success(adminText('oidcRotated'))
  } catch (err) {
    adminToast.error(err, adminText('oidcRotateFailed'))
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
  if (!input.name) return adminText('oidcNameRequired')
  if (!input.redirectUris.length) return adminText('oidcRedirectRequired')
  if (input.redirectUris.some(uri => !/^https?:\/\//i.test(uri))) return adminText('oidcRedirectInvalid')
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
      adminToast.success(adminText('oidcSaved'))
    } else {
      const credentials = await createOIDCClient(input)
      clients.value = [...clients.value, credentials.client]
      editorOpen.value = false
      showSecret(credentials.client.clientId, credentials.clientSecret || '')
    }
  } catch (err) {
    adminToast.error(err, editing.value ? adminText('oidcSaveFailed') : adminText('oidcCreateFailed'))
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
    adminToast.error(err, adminText('oidcToggleFailed'))
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
    adminToast.error(err, adminText('oidcResetFailed'))
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

async function reload() {
  await Promise.all([loadProviderStatus(), loadClients()])
}

onMounted(reload)
onBeforeUnmount(clearSecret)
</script>

<template>
  <BasicPage :title="adminText('oidcTitle')" :description="adminText('oidcDescription')">
    <template #primary-action>
      <Button size="sm" type="button" @click="openCreate">
        <Plus class="size-4" />
        {{ adminText('oidcCreate') }}
      </Button>
    </template>

    <AdminSection class="mb-4" body-class="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="min-w-0">
        <div class="flex items-center gap-2">
          <span class="text-sm font-semibold">{{ adminText('oidcStatus') }}</span>
          <Badge v-if="providerStatus" :variant="providerStatus.available ? 'default' : 'outline'">
            {{ providerStatus.available ? adminText('oidcRunning') : providerStatus.enabled ? adminText('oidcConfigError') : adminText('oidcNotEnabled') }}
          </Badge>
        </div>
        <p v-if="providerError" class="mt-1 flex flex-wrap items-center gap-2 text-xs text-destructive">
          <span>{{ providerError }}</span>
          <Button type="button" variant="link" size="sm" class="h-auto px-0 text-destructive" @click="loadProviderStatus">{{ t('common.retry') }}</Button>
        </p>
        <p v-else-if="providerStatus?.issuer" class="mt-1 truncate font-mono text-xs text-muted-foreground">{{ providerStatus.issuer }}</p>
        <p v-else-if="providerStatus?.error" class="mt-1 flex items-start gap-1.5 text-xs text-destructive">
          <CircleAlert class="mt-0.5 size-3.5 shrink-0" />
          <span>{{ providerStatus.error }}</span>
        </p>
        <p v-else class="mt-1 text-xs text-muted-foreground">{{ adminText('oidcIssuerHint') }}</p>
      </div>
      <div class="flex shrink-0 flex-wrap items-center gap-3">
        <Button type="button" size="sm" variant="outline" :disabled="!providerStatus?.available || signingKeyRotating" @click="signingKeyDialogOpen = true">
          <RotateCw class="size-4" />
          {{ adminText('oidcRotateSigning') }}
        </Button>
        <label class="flex items-center gap-2 text-sm">
          {{ adminText('oidcEnableProvider') }}
          <Switch :model-value="providerStatus?.enabled ?? false" :disabled="!providerStatus || providerSaving" @update:model-value="toggleProvider(Boolean($event))" />
        </label>
      </div>
    </AdminSection>

    <AdminSection>
      <template #header>
        <div class="flex items-center justify-between gap-3">
          <div>
            <h2 class="text-sm font-semibold">{{ adminText('oidcApps') }}</h2>
            <p class="mt-0.5 text-xs text-muted-foreground">{{ adminText('oidcCount', { count: clients.length }) }}</p>
          </div>
          <Button type="button" size="icon-sm" variant="ghost" :disabled="loading" :title="t('common.refresh')" @click="loadClients">
            <RefreshCw class="size-4" :class="loading ? 'animate-spin' : ''" />
          </Button>
        </div>
      </template>

      <div v-if="loading && !clients.length" class="grid h-28 place-items-center px-4 text-sm text-muted-foreground">
        {{ adminText('k0046') }}
      </div>
      <div v-else-if="clientError" class="flex min-h-28 items-center justify-center p-4 text-sm">
        <div class="inline-flex items-center gap-3 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-2 text-destructive">
          <span>{{ clientError }}</span>
          <Button type="button" variant="link" size="sm" class="h-auto px-0 text-destructive" @click="loadClients">{{ t('common.retry') }}</Button>
        </div>
      </div>
      <div v-else-if="!clients.length" class="flex min-h-28 flex-col items-center justify-center gap-1 px-4 py-6 text-center text-sm text-muted-foreground">
        <p>{{ adminText('oidcEmpty') }}</p>
        <p class="text-xs">{{ adminText('oidcEmptyHint') }}</p>
      </div>
      <div v-else class="divide-y">
        <article v-for="client in clients" :key="client.clientId" class="grid gap-4 px-4 py-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h3 class="truncate text-sm font-semibold">{{ client.name }}</h3>
              <Badge variant="secondary">{{ client.public ? adminText('oidcPublic') : adminText('oidcConfidential') }}</Badge>
              <Badge :variant="client.enabled ? 'default' : 'outline'">{{ client.enabled ? adminText('oidcEnabled') : adminText('oidcDisabled') }}</Badge>
            </div>
            <p class="mt-1 truncate font-mono text-xs text-muted-foreground">{{ client.clientId }}</p>
            <div class="mt-2 flex flex-wrap gap-1.5">
              <Badge v-for="scope in client.scopes" :key="scope" variant="outline" class="font-mono text-[11px] font-normal">{{ scope }}</Badge>
              <span class="text-xs text-muted-foreground">{{ adminText('oidcRedirectCount', { count: client.redirectUris.length }) }}</span>
            </div>
          </div>
          <div class="flex items-center gap-1 md:justify-end">
            <label class="mr-2 flex items-center gap-2 text-xs text-muted-foreground">
              {{ adminText('k00fq') }}
              <Switch :model-value="client.enabled" :disabled="Boolean(togglingId)" @update:model-value="toggleEnabled(client, Boolean($event))" />
            </label>
            <Button v-if="!client.public" type="button" size="icon-sm" variant="ghost" :title="adminText('oidcResetSecret')" @click="rotateTarget = client">
              <RotateCw class="size-4" />
            </Button>
            <Button type="button" size="icon-sm" variant="ghost" :title="t('common.edit')" @click="openEdit(client)">
              <Pencil class="size-4" />
            </Button>
          </div>
        </article>
      </div>
    </AdminSection>

    <Dialog v-model:open="editorOpen">
      <DialogContent class="max-h-[calc(100vh-2rem)] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{{ editing ? adminText('oidcEdit') : adminText('oidcCreateTitle') }}</DialogTitle>
          <DialogDescription>{{ adminText('oidcEditorHint') }}</DialogDescription>
        </DialogHeader>

        <form id="oidc-client-form" class="grid gap-6" @submit.prevent="submit">
          <div class="grid gap-2">
            <Label for="oidc-client-name">{{ adminText('oidcAppName') }}</Label>
            <Input id="oidc-client-name" v-model="form.name" maxlength="255" placeholder="Internal Wiki" autofocus />
          </div>

          <div class="grid gap-2">
            <Label for="oidc-client-redirects">{{ adminText('oidcRedirects') }}</Label>
            <Textarea id="oidc-client-redirects" v-model="form.redirectUris" class="min-h-24 font-mono text-xs" placeholder="https://wiki.example.com/oauth/callback" />
          </div>

          <fieldset class="grid gap-3">
            <legend class="text-sm font-medium">{{ adminText('oidcClientType') }}</legend>
            <div class="grid gap-2 sm:grid-cols-2">
              <button type="button" class="rounded-md border p-3 text-left text-sm transition-colors outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50" :class="!form.public ? 'border-primary bg-primary/5' : 'hover:bg-muted/50'" :disabled="editing" :aria-pressed="!form.public" @click="setPublic(false)">
                <span class="font-medium">{{ adminText('oidcConfidential') }}</span>
                <span class="mt-1 block text-xs text-muted-foreground">{{ adminText('oidcConfidentialHint') }}</span>
              </button>
              <button type="button" class="rounded-md border p-3 text-left text-sm transition-colors outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50" :class="form.public ? 'border-primary bg-primary/5' : 'hover:bg-muted/50'" :disabled="editing" :aria-pressed="form.public" @click="setPublic(true)">
                <span class="font-medium">{{ adminText('oidcPublic') }}</span>
                <span class="mt-1 block text-xs text-muted-foreground">{{ adminText('oidcPublicHint') }}</span>
              </button>
            </div>
          </fieldset>

          <div class="grid gap-4 sm:grid-cols-2">
            <div class="grid gap-2">
              <Label for="oidc-token-auth">{{ adminText('oidcTokenAuth') }}</Label>
              <Select v-model="form.tokenEndpointAuthMethod" :disabled="form.public">
                <SelectTrigger id="oidc-token-auth"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="client_secret_basic">client_secret_basic</SelectItem>
                  <SelectItem value="client_secret_post">client_secret_post</SelectItem>
                  <SelectItem v-if="form.public" value="none">none</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="flex items-center justify-between gap-4 rounded-md border px-3 py-2.5">
              <div>
                <p class="text-sm font-medium">{{ adminText('oidcRequirePkce') }}</p>
                <p class="text-xs text-muted-foreground">{{ adminText('oidcPkceHint') }}</p>
              </div>
              <Switch :model-value="form.public || form.requirePkce" :disabled="form.public" @update:model-value="form.requirePkce = Boolean($event)" />
            </div>
          </div>

          <fieldset class="grid gap-2">
            <legend class="text-sm font-medium">{{ adminText('oidcScopes') }}</legend>
            <div class="grid gap-2 sm:grid-cols-2">
              <label v-for="scope in supportedScopes" :key="scope" class="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
                <span class="font-mono text-xs">{{ scope }}</span>
                <Switch :model-value="form.scopes.includes(scope)" :disabled="scope === 'openid'" @update:model-value="toggleScope(scope, Boolean($event))" />
              </label>
            </div>
          </fieldset>

          <label class="flex items-center justify-between gap-4 rounded-md border px-3 py-2.5">
            <span>
              <span class="block text-sm font-medium">{{ adminText('oidcEnableClient') }}</span>
              <span class="block text-xs text-muted-foreground">{{ adminText('oidcDisableHint') }}</span>
            </span>
            <Switch v-model="form.enabled" />
          </label>
        </form>

        <DialogFooter>
          <Button type="button" variant="outline" @click="editorOpen = false">{{ t('common.cancel') }}</Button>
          <Button type="submit" form="oidc-client-form" :disabled="saving">
            <Loader2 v-if="saving" class="size-4 animate-spin" />
            {{ saving ? t('common.saving') : t('common.save') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <AdminConfirmDialog
      :open="signingKeyDialogOpen"
      :loading="signingKeyRotating"
      :title="adminText('oidcRotateTitle')"
      :description="adminText('oidcRotateHint')"
      :confirm-text="adminText('oidcRotateConfirm')"
      @update:open="signingKeyDialogOpen = $event"
      @confirm="confirmSigningKeyRotation"
    />

    <AdminConfirmDialog
      :open="Boolean(rotateTarget)"
      :loading="rotating"
      :title="adminText('oidcResetSecret')"
      :description="adminText('oidcResetHint')"
      :confirm-text="adminText('oidcResetConfirm')"
      @update:open="!$event && (rotateTarget = null)"
      @confirm="confirmRotate"
    />

    <Dialog :open="secretOpen" @update:open="setSecretOpen">
      <DialogContent class="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>{{ adminText('oidcSecret') }}</DialogTitle>
          <DialogDescription>{{ adminText('oidcSecretHint') }}</DialogDescription>
        </DialogHeader>
        <div class="grid gap-3">
          <div>
            <p class="mb-1 text-xs font-medium text-muted-foreground">{{ adminText('oidcClientId') }}</p>
            <code class="block overflow-x-auto rounded-md border bg-muted/40 p-3 text-xs">{{ secretClientId }}</code>
          </div>
          <div>
            <p class="mb-1 text-xs font-medium text-muted-foreground">{{ adminText('oidcClientSecret') }}</p>
            <div class="flex min-w-0 gap-2">
              <Input :model-value="revealedSecret" readonly class="min-w-0 font-mono text-xs" />
              <Button type="button" size="icon" variant="outline" :title="copied ? adminText('oidcCopied') : adminText('oidcCopySecret')" @click="copySecret">
                <Check v-if="copied" class="size-4" />
                <Copy v-else class="size-4" />
              </Button>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button type="button" @click="setSecretOpen(false)">{{ adminText('oidcStored') }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </BasicPage>
</template>
