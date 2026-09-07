<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { KeyRound, Loader2, Plus, Save, Trash2, X } from '@lucide/vue'
import { siDiscord, siGithub, siGoogle, type SimpleIcon } from 'simple-icons'
import { BasicPage } from '@/admin/components/global-layout'
import { Badge } from '@/admin/components/ui/badge'
import { Button } from '@/admin/components/ui/button'
import { Input } from '@/admin/components/ui/input'
import { Switch } from '@/admin/components/ui/switch'
import { adminText } from '@/admin/runtime/i18n-text'
import { getOAuthSettings, saveOAuthSettings } from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import type { OAuthProviderSettings, OAuthSettings } from '@/admin/types'

interface OAuthProviderForm extends OAuthProviderSettings {
  scopeDraft: string
}

const providers = ref<OAuthProviderForm[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const activeProviderIndex = ref(0)
const activeProvider = computed(() => providers.value[activeProviderIndex.value])

const providerIcons: Record<string, SimpleIcon> = {
  github: siGithub,
  google: siGoogle,
  discord: siDiscord,
}

function toForm(provider: OAuthProviderSettings): OAuthProviderForm {
  return {
    ...provider,
    clientSecret: '',
    clearClientSecret: false,
    scopes: provider.scopes?.length ? [...provider.scopes] : ['openid', 'profile', 'email'],
    scopeDraft: '',
  }
}

function applySettings(settings: OAuthSettings) {
  providers.value = settings.providers.map(toForm)
  activeProviderIndex.value = Math.min(activeProviderIndex.value, Math.max(providers.value.length - 1, 0))
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    applySettings(await getOAuthSettings())
  } catch (err) {
    error.value = err instanceof Error ? err.message : adminText('k00g4')
  } finally {
    loading.value = false
  }
}

function addProvider() {
  providers.value.push({
    key: '',
    displayName: '',
    kind: 'oidc',
    enabled: false,
    clientId: '',
    clientSecret: '',
    clientSecretConfigured: false,
    clearClientSecret: false,
    callbackUrl: '',
    discoveryUrl: '',
    scopes: ['openid', 'profile', 'email'],
    scopeDraft: '',
  })
  activeProviderIndex.value = providers.value.length - 1
}

function removeProvider(index: number) {
  if (providers.value[index]?.kind !== 'oidc') return
  providers.value.splice(index, 1)
  activeProviderIndex.value = Math.min(index, Math.max(providers.value.length - 1, 0))
}

function addScope(provider: OAuthProviderForm) {
  const scope = provider.scopeDraft.trim()
  if (!scope || provider.scopes?.includes(scope)) return
  provider.scopes = [...(provider.scopes || []), scope]
  provider.scopeDraft = ''
}

function removeScope(provider: OAuthProviderForm, scope: string) {
  if (scope === 'openid') return
  provider.scopes = (provider.scopes || []).filter(item => item !== scope)
}

async function save() {
  saving.value = true
  try {
    const settings: OAuthSettings = {
      providers: providers.value.map(provider => ({
        key: provider.key.trim().toLowerCase(),
        displayName: provider.displayName.trim(),
        kind: provider.kind,
        enabled: provider.enabled,
        clientId: provider.clientId.trim(),
        clientSecret: provider.clientSecret?.trim() || '',
        clientSecretConfigured: provider.clientSecretConfigured,
        clearClientSecret: Boolean(provider.clearClientSecret),
        callbackUrl: provider.callbackUrl,
        discoveryUrl: provider.discoveryUrl?.trim() || '',
        scopes: provider.scopes || [],
      })),
    }
    applySettings(await saveOAuthSettings(settings))
    adminToast.success(adminText('k00g0'))
  } catch (err) {
    adminToast.error(err, adminText('k000f'))
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <BasicPage :title="adminText('k00fl')" :description="adminText('k00fm')">
    <template #primary-action>
      <div class="flex items-center gap-2">
        <Button type="button" variant="outline" @click="addProvider">
          <Plus class="size-4" />
          {{ adminText('k00fn') }}
        </Button>
        <Button type="button" :disabled="saving" @click="save">
          <Loader2 v-if="saving" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          {{ adminText('k004f') }}
        </Button>
      </div>
    </template>

    <div v-if="loading" class="flex h-[360px] items-center justify-center">
      <Loader2 class="size-8 animate-spin text-primary" />
    </div>
    <div v-else-if="error" class="rounded-md border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">
      {{ error }}
    </div>
    <form v-else-if="activeProvider" class="overflow-hidden rounded-md border bg-background" @submit.prevent="save">
      <nav class="flex gap-1 overflow-x-auto border-b bg-muted/15 p-2" :aria-label="adminText('k00fl')">
        <button
          v-for="(provider, index) in providers"
          :key="`${provider.kind}-${provider.key || index}`"
          type="button"
          class="flex h-10 shrink-0 items-center gap-2 rounded-md px-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-background/70 hover:text-foreground"
          :class="activeProviderIndex === index ? 'bg-background text-foreground shadow-xs ring-1 ring-border' : ''"
          @click="activeProviderIndex = index"
        >
          <svg
            v-if="providerIcons[provider.key]"
            class="size-4"
            viewBox="0 0 24 24"
            fill="currentColor"
            aria-hidden="true"
            :style="{ color: `#${providerIcons[provider.key].hex}` }"
          >
            <path :d="providerIcons[provider.key].path" />
          </svg>
          <KeyRound v-else class="size-4" />
          <span>{{ provider.displayName || provider.key || adminText('k00fp') }}</span>
          <span class="size-1.5 rounded-full" :class="provider.enabled ? 'bg-emerald-500' : 'bg-muted-foreground/35'" />
        </button>
      </nav>

      <section class="space-y-5 p-4 sm:p-5">
        <header class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-base font-semibold">{{ activeProvider.displayName || activeProvider.key || adminText('k00fp') }}</h2>
            <p class="mt-0.5 text-xs text-muted-foreground">
              {{ activeProvider.kind === 'oidc' ? adminText('k00fp') : adminText('k00fo') }} ·
              {{ activeProvider.clientSecretConfigured ? adminText('k00ft') : adminText('k00fu') }}
            </p>
          </div>
          <div class="flex items-center gap-3">
            <label class="flex items-center gap-2 text-sm font-medium">
              {{ adminText('k00fq') }}
              <Switch v-model="activeProvider.enabled" />
            </label>
            <Button v-if="activeProvider.kind === 'oidc'" type="button" size="icon" variant="ghost" :title="adminText('k00fz')" @click="removeProvider(activeProviderIndex)">
              <Trash2 class="size-4 text-destructive" />
            </Button>
          </div>
        </header>

        <div v-if="activeProvider.kind === 'oidc'" class="grid gap-4 md:grid-cols-2">
          <label class="grid gap-2 text-sm font-medium">
            {{ adminText('k00g1') }}
            <Input v-model="activeProvider.displayName" placeholder="Company SSO" />
          </label>
          <label class="grid gap-2 text-sm font-medium">
            {{ adminText('k00g2') }}
            <Input v-model="activeProvider.key" placeholder="company-sso" :title="adminText('k00g3')" />
          </label>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="grid gap-2 text-sm font-medium">
            {{ adminText('k00fr') }}
            <Input v-model="activeProvider.clientId" autocomplete="off" />
          </label>
          <label class="grid gap-2 text-sm font-medium">
            {{ adminText('k00fs') }}
            <div class="flex gap-2">
              <Input
                v-model="activeProvider.clientSecret"
                class="min-w-0"
                type="password"
                autocomplete="new-password"
                :disabled="activeProvider.clearClientSecret"
                :placeholder="activeProvider.clientSecretConfigured ? adminText('k00fv') : ''"
              />
              <Button
                v-if="activeProvider.clientSecretConfigured"
                type="button"
                variant="outline"
                class="shrink-0"
                @click="activeProvider.clearClientSecret = !activeProvider.clearClientSecret"
              >
                {{ activeProvider.clearClientSecret ? adminText('k00g7') : adminText('k00g6') }}
              </Button>
            </div>
            <span v-if="activeProvider.clearClientSecret" class="text-xs font-normal text-destructive">{{ adminText('k00g5') }}</span>
          </label>
        </div>

        <div v-if="activeProvider.kind === 'oidc'" class="grid gap-4 md:grid-cols-2">
          <label class="grid gap-2 text-sm font-medium">
            {{ adminText('k00fx') }}
            <Input v-model="activeProvider.discoveryUrl" placeholder="https://id.example.com/.well-known/openid-configuration" />
          </label>
          <label class="grid gap-2 text-sm font-medium">
            {{ adminText('k00fy') }}
            <div class="flex min-w-0 gap-2">
              <Input
                v-model="activeProvider.scopeDraft"
                class="min-w-0"
                placeholder="email"
                @keydown.enter.prevent="addScope(activeProvider)"
              />
              <Button type="button" size="icon" variant="outline" :title="adminText('k0094')" @click="addScope(activeProvider)">
                <Plus class="size-4" />
              </Button>
            </div>
          </label>
          <div class="flex min-h-7 flex-wrap gap-1.5 md:col-start-2">
            <Badge v-for="scope in activeProvider.scopes" :key="scope" variant="secondary" class="gap-1 px-2 py-1 font-mono text-xs font-normal">
              {{ scope }}
              <button
                v-if="scope !== 'openid'"
                type="button"
                class="rounded-sm text-muted-foreground hover:text-foreground"
                :title="adminText('k005i')"
                @click="removeScope(activeProvider, scope)"
              >
                <X class="size-3" />
              </button>
            </Badge>
          </div>
        </div>

        <label class="grid gap-2 text-sm font-medium">
          {{ adminText('k00fw') }}
          <Input :model-value="activeProvider.callbackUrl" readonly class="font-mono text-xs" />
        </label>
      </section>
    </form>
  </BasicPage>
</template>
