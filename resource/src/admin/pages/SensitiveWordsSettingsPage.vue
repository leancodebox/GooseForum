<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Pencil, Plus, RefreshCw, Save, Search, Trash2, X } from '@lucide/vue'
import { adminText } from '@/admin/runtime/i18n-text'
import { adminToast } from '@/admin/runtime/toast'
import { getSensitiveWordSettings, getSensitiveWords, saveSensitiveWordSettings, saveSensitiveWord, deleteSensitiveWord } from '@/admin/runtime/api'
import type { AdminPayload, ManageHomeProps, SensitiveWord, SensitiveWordSettings } from '@/admin/types'
import { BasicPage } from '@/admin/components/global-layout'
import AdminSection from '@/admin/components/AdminSection.vue'
import AdminToolbar from '@/admin/components/AdminToolbar.vue'
import AdminActionButton from '@/admin/components/AdminActionButton.vue'
import { Button } from '@/admin/components/ui/button'
import { Input } from '@/admin/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/admin/components/ui/select'
import { Switch } from '@/admin/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/admin/components/ui/table'

defineProps<{ payload: AdminPayload<ManageHomeProps> }>()

const settings = reactive<SensitiveWordSettings>({ enabled: false, mode: 'after_review' })
const words = ref<SensitiveWord[]>([])
const emptyWord = (): SensitiveWord => ({ id: 0, word: '', action: 'reject', replacement: '', enabled: true })
const editor = reactive(emptyWord())
const loading = ref(true)
const savingSettings = ref(false)
const savingWord = ref(false)
const error = ref('')
const search = ref('')
const page = ref(1)
const filtered = computed(() => words.value.filter(item => `${item.word} ${item.replacement}`.includes(search.value.trim())))
const visible = computed(() => filtered.value.slice((page.value - 1) * 50, page.value * 50))
watch(search, () => { page.value = 1 })
const actions = ['reject', 'replace', 'record'] as const
const actionLabel = (action: string) => adminText({ reject: 'sensitiveReject', replace: 'sensitiveReplace', record: 'sensitiveRecord' }[action] || 'sensitiveReject')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [config, list] = await Promise.all([getSensitiveWordSettings(), getSensitiveWords()])
    Object.assign(settings, config)
    words.value = list
  } catch (err) { error.value = err instanceof Error ? err.message : adminText('reviewFailed') }
  finally { loading.value = false }
}
async function saveSettings() {
  if (savingSettings.value) return
  savingSettings.value = true
  try { await saveSensitiveWordSettings({ ...settings }); adminToast.success(adminText('reviewSaved')) }
  catch (err) { adminToast.error(err, adminText('reviewFailed')) }
  finally { savingSettings.value = false }
}
function resetEditor() { Object.assign(editor, emptyWord()) }
async function persist(word: SensitiveWord) {
  const { word: saved } = await saveSensitiveWord({ ...word })
  const index = words.value.findIndex(item => item.id === saved.id)
  if (index < 0) words.value.push(saved)
  else words.value[index] = saved
}
async function saveWord() {
  if (savingWord.value || !editor.word.trim()) return
  savingWord.value = true
  try {
    await persist(editor)
    resetEditor()
    adminToast.success(adminText('reviewSaved'))
  } catch (err) { adminToast.error(err, adminText('reviewFailed')) }
  finally { savingWord.value = false }
}
async function toggleWord(word: SensitiveWord, enabled: boolean) {
  if (savingWord.value) return
  savingWord.value = true
  try { await persist({ ...word, enabled }); if (editor.id === word.id) editor.enabled = enabled }
  catch (err) { adminToast.error(err, adminText('reviewFailed')) }
  finally { savingWord.value = false }
}
async function removeWord(id: number) {
  if (savingWord.value) return
  savingWord.value = true
  try {
    await deleteSensitiveWord(id)
    words.value = words.value.filter(item => item.id !== id)
    if (editor.id === id) resetEditor()
    page.value = Math.min(page.value, Math.max(1, Math.ceil(filtered.value.length / 50)))
  } catch (err) { adminToast.error(err, adminText('reviewFailed')) }
  finally { savingWord.value = false }
}
onMounted(load)
</script>

<template>
  <BasicPage :title="adminText('sensitiveTitle')" :description="adminText('sensitiveDescription')">
    <template #primary-action>
      <Button size="sm" :disabled="loading || savingSettings || Boolean(error)" @click="saveSettings"><Save class="size-4" />{{ adminText('k004f') }}</Button>
    </template>
    <div v-if="error" class="mb-4 flex items-center justify-between gap-3 rounded-md border border-destructive/30 p-3 text-sm text-destructive">
      {{ error }}<Button variant="outline" size="sm" @click="load">{{ adminText('k004q') }}</Button>
    </div>
    <AdminSection class="mb-4">
      <div class="flex flex-wrap items-center gap-x-8 gap-y-3 px-3 py-3">
        <label class="flex items-center gap-3 text-sm font-medium"><Switch v-model="settings.enabled" :disabled="loading || savingSettings" />{{ adminText('sensitiveEnable') }}</label>
        <label class="flex flex-wrap items-center gap-2 text-sm">
          <span class="text-muted-foreground">{{ adminText('sensitiveMode') }}</span>
          <Select v-model="settings.mode" :disabled="loading || savingSettings">
            <SelectTrigger class="h-9 w-auto" :aria-label="adminText('sensitiveMode')"><SelectValue /></SelectTrigger>
            <SelectContent>
            <SelectItem value="after_review">{{ adminText('sensitiveAfterReview') }}</SelectItem>
            <SelectItem value="visible_then_review">{{ adminText('sensitiveVisibleThenReview') }}</SelectItem>
          </SelectContent></Select>
        </label>
        <span class="text-xs text-muted-foreground">{{ adminText('sensitiveAsyncHint') }}</span>
      </div>
    </AdminSection>
    <AdminSection>
      <template #header>
        <AdminToolbar class="-mx-3 -my-2 border-b-0">
          <div class="flex flex-1 items-center gap-3">
            <span class="shrink-0 text-sm font-medium">{{ adminText('sensitiveDictionary') }} <span class="text-muted-foreground">{{ filtered.length }}</span></span>
            <div class="relative w-full max-w-sm"><Search class="pointer-events-none absolute left-2.5 top-2.5 size-4 text-muted-foreground" /><Input v-model="search" class="h-9 pl-8" :placeholder="adminText('sensitiveSearch')" /></div>
          </div>
          <div class="flex items-center gap-2">
            <Button size="sm" variant="outline" :disabled="loading || savingWord || savingSettings" @click="load"><RefreshCw class="size-4" :class="{ 'animate-spin': loading }" />{{ adminText('k004q') }}</Button>
            <Button size="sm" variant="outline" :disabled="page <= 1" @click="page--">{{ adminText('k00au') }}</Button>
            <span class="text-sm text-muted-foreground">{{ page }}</span>
            <Button size="sm" variant="outline" :disabled="page * 50 >= filtered.length" @click="page++">{{ adminText('k00av') }}</Button>
          </div>
        </AdminToolbar>
      </template>
      <form class="grid items-end gap-3 border-b bg-muted/10 p-3 md:grid-cols-[minmax(0,2fr)_140px_minmax(0,2fr)_auto]" @submit.prevent="saveWord">
        <label class="grid gap-1 text-xs text-muted-foreground">{{ adminText('sensitiveWordLabel') }}<Input v-model="editor.word" :maxlength="128" :disabled="loading || savingWord" :placeholder="adminText('sensitiveWordPlaceholder')" /></label>
        <label class="grid gap-1 text-xs text-muted-foreground">{{ adminText('sensitiveActionLabel') }}<Select v-model="editor.action" :disabled="loading || savingWord"><SelectTrigger class="h-9 w-full text-foreground" :aria-label="adminText('sensitiveActionLabel')"><SelectValue /></SelectTrigger><SelectContent><SelectItem v-for="action in actions" :key="action" :value="action">{{ actionLabel(action) }}</SelectItem></SelectContent></Select></label>
        <label class="grid gap-1 text-xs text-muted-foreground">{{ adminText('sensitiveReplacementLabel') }}<Input v-model="editor.replacement" :maxlength="128" :disabled="loading || savingWord || editor.action !== 'replace'" :placeholder="adminText('sensitiveReplacementPlaceholder')" /></label>
        <div class="flex h-9 gap-2">
          <Button type="submit" :disabled="loading || savingWord || !editor.word.trim()"><Save v-if="editor.id" class="size-4" /><Plus v-else class="size-4" />{{ editor.id ? adminText('sensitiveSave') : adminText('sensitiveAdd') }}</Button>
          <Button v-if="editor.id" type="button" variant="ghost" :disabled="savingWord" @click="resetEditor"><X class="size-4" />{{ adminText('k009q') }}</Button>
        </div>
      </form>
      <Table>
        <TableHeader><TableRow class="bg-muted/20">
          <TableHead class="w-[28%] pl-3">{{ adminText('sensitiveWordLabel') }}</TableHead>
          <TableHead class="w-28">{{ adminText('sensitiveActionLabel') }}</TableHead>
          <TableHead>{{ adminText('sensitiveReplacementLabel') }}</TableHead>
          <TableHead class="w-20 text-center">{{ adminText('sensitiveEnabledLabel') }}</TableHead>
          <TableHead class="w-24 pr-3 text-right">{{ adminText('k007m') }}</TableHead>
        </TableRow></TableHeader>
        <TableBody>
          <TableRow v-if="loading"><TableCell colspan="5" class="h-24 text-center text-muted-foreground">{{ adminText('k0046') }}</TableCell></TableRow>
          <TableRow v-else-if="!visible.length"><TableCell colspan="5" class="h-24 text-center text-muted-foreground">{{ adminText('reviewEmptyWords') }}</TableCell></TableRow>
          <TableRow v-for="word in visible" v-else :key="word.id" :class="{ 'bg-primary/5': editor.id === word.id }">
            <TableCell class="whitespace-normal break-all pl-3 font-medium">{{ word.word }}</TableCell>
            <TableCell>{{ actionLabel(word.action) }}</TableCell>
            <TableCell class="whitespace-normal break-all"><span v-if="word.action === 'replace'" class="font-mono text-sm">{{ word.replacement || adminText('sensitiveEmptyReplacement') }}</span><span v-else class="text-muted-foreground">—</span></TableCell>
            <TableCell class="text-center"><Switch :model-value="word.enabled" :disabled="savingWord" :aria-label="`${adminText('sensitiveEnabledLabel')} ${word.word}`" @update:model-value="value => toggleWord(word, value)" /></TableCell>
            <TableCell class="pr-3 text-right"><div class="flex justify-end gap-1"><AdminActionButton :title="adminText('sensitiveEdit')" :disabled="savingWord" @click="Object.assign(editor, word)"><Pencil class="size-4" /></AdminActionButton><AdminActionButton :title="adminText('sensitiveDelete')" tone="danger" :disabled="savingWord" @click="removeWord(word.id)"><Trash2 class="size-4" /></AdminActionButton></div></TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminSection>
  </BasicPage>
</template>
