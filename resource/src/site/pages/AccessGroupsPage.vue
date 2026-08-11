<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { CheckCircle2, Clock3, Lock, UsersRound } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '@/site/components/PageHeader.vue'
import {
  applyToAccessGroup,
  getJoinableAccessGroups,
  getManagedAccessGroups,
  reviewManagedAccessGroupApplication,
  type JoinableAccessGroup,
  type ManagedAccessGroup,
} from '@/runtime/api'
import type { AccessGroupsPageProps, LayoutPayload } from '@gooseforum/client'

defineProps<{ layout: LayoutPayload; props: AccessGroupsPageProps }>()

const { t } = useI18n()
const groups = ref<JoinableAccessGroup[]>([])
const managedGroups = ref<ManagedAccessGroup[]>([])
const loading = ref(true)
const applyingId = ref(0)
const reviewingMemberId = ref(0)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    ;[groups.value, managedGroups.value] = await Promise.all([getJoinableAccessGroups(), getManagedAccessGroups()])
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('accessGroups.loadFailed')
  } finally {
    loading.value = false
  }
}

async function review(groupId: number, memberId: number, approve: boolean) {
  if (reviewingMemberId.value) return
  reviewingMemberId.value = memberId
  error.value = ''
  try {
    await reviewManagedAccessGroupApplication(groupId, memberId, approve)
    managedGroups.value = await getManagedAccessGroups()
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('accessGroups.memberSaveFailed')
  } finally {
    reviewingMemberId.value = 0
  }
}

async function apply(group: JoinableAccessGroup) {
  if (applyingId.value) return
  applyingId.value = group.id
  error.value = ''
  try {
    await applyToAccessGroup(group.id)
    group.status = 2
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('accessGroups.applicationFailed')
  } finally {
    applyingId.value = 0
  }
}

onMounted(() => void load())
</script>

<template>
  <main class="min-w-0 pb-8">
    <PageHeader :title="t('accessGroups.joinTitle')" :description="t('accessGroups.joinDescription')" />
    <div v-if="error" class="gf-status-message gf-status-message-error mb-3">
      {{ error }}
    </div>
    <section class="gf-card overflow-hidden">
      <div v-if="loading" class="px-5 py-12 text-center text-sm text-base-content/55">
        {{ t('common.loadingShort') }}
      </div>
      <div v-else-if="!groups.length" class="px-5 py-12 text-center">
        <UsersRound class="mx-auto h-8 w-8 text-base-content/35" />
        <h2 class="mt-3 font-semibold">
          {{ t('accessGroups.noJoinableGroups') }}
        </h2>
        <p class="mt-1 text-sm text-base-content/55">
          {{ t('accessGroups.noJoinableGroupsHint') }}
        </p>
      </div>
      <div v-else class="divide-y divide-line">
        <article
          v-for="group in groups"
          :key="group.id"
          class="flex flex-col gap-4 px-5 py-4 sm:flex-row sm:items-center sm:justify-between"
        >
          <div class="min-w-0">
            <h2 class="flex items-center gap-2 font-semibold">
              <Lock class="h-4 w-4 text-base-content/45" />{{ group.name }}
            </h2>
            <p class="mt-1 text-sm text-base-content/55">
              {{
                group.categories.length
                  ? t('accessGroups.unlocksCategories', {
                      categories: group.categories.join('、'),
                    })
                  : t('accessGroups.noCategoryGrants')
              }}
            </p>
          </div>
          <span v-if="group.status === 1" class="inline-flex items-center gap-1.5 text-sm font-medium text-success"
            ><CheckCircle2 class="h-4 w-4" />{{ t('accessGroups.joined') }}</span
          >
          <span v-else-if="group.status === 2" class="inline-flex items-center gap-1.5 text-sm font-medium text-warning"
            ><Clock3 class="h-4 w-4" />{{ t('accessGroups.pending') }}</span
          >
          <button v-else class="gf-button gf-button-primary" :disabled="applyingId !== 0" @click="apply(group)">
            {{ applyingId === group.id ? t('common.saving') : t('accessGroups.apply') }}
          </button>
        </article>
      </div>
    </section>

    <section v-if="managedGroups.length" class="gf-card mt-4 overflow-hidden">
      <div class="border-b border-line px-5 py-4">
        <h2 class="font-semibold">
          {{ t('accessGroups.applicationsToReview') }}
        </h2>
        <p class="mt-1 text-sm text-base-content/55">
          {{ t('accessGroups.managerReviewHint') }}
        </p>
      </div>
      <div class="divide-y divide-line">
        <article v-for="group in managedGroups" :key="group.id" class="px-5 py-4">
          <h3 class="font-medium">{{ group.name }}</h3>
          <p v-if="!group.applications.length" class="mt-2 text-sm text-base-content/55">
            {{ t('accessGroups.noPendingApplications') }}
          </p>
          <div
            v-for="application in group.applications"
            :key="application.id"
            class="mt-3 flex items-center justify-between gap-3 rounded-md border border-line p-3"
          >
            <span class="font-medium">{{ application.username || `#${application.userId}` }}</span>
            <div class="flex gap-2">
              <button
                class="gf-button gf-button-muted"
                :disabled="reviewingMemberId !== 0"
                @click="review(group.id, application.id, false)"
              >
                {{ t('accessGroups.reject') }}</button
              ><button
                class="gf-button gf-button-primary"
                :disabled="reviewingMemberId !== 0"
                @click="review(group.id, application.id, true)"
              >
                {{ t('accessGroups.approve') }}
              </button>
            </div>
          </div>
        </article>
      </div>
    </section>
  </main>
</template>
