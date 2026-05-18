<script setup lang="ts">
import { ExternalLink, Link, Send, ShieldCheck } from '@lucide/vue'
import AppShell from '../components/AppShell.vue'
import type { LayoutPayload, LinksPageProps } from '../types/payload'

defineProps<{
  layout: LayoutPayload
  props: LinksPageProps
}>()
</script>

<template>
  <AppShell :layout="layout">
    <div class="pb-12">
      <header class="mb-3 border-b border-gray-200/70 pb-4">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="text-2xl font-bold text-gray-950">友情链接</h1>
          <span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-semibold text-gray-500">{{ props.totalCount }}</span>
        </div>
        <p class="mt-2 max-w-2xl text-sm leading-6 text-gray-500">社区伙伴、开源项目和值得访问的友好站点。</p>
      </header>

      <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_260px]">
        <div class="space-y-5">
          <section v-for="group in props.groups" :key="group.name" class="space-y-3">
            <div class="flex items-center justify-between gap-3 border-b border-gray-100 pb-2">
              <h2 class="flex min-w-0 items-center gap-2 text-base font-bold text-gray-950">
                <span
                  class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-gray-50 text-sm"
                  :style="{ color: group.color || '#64748b' }"
                >
                  {{ group.emoji || '↗' }}
                </span>
                <span class="truncate">{{ group.name }}</span>
              </h2>
              <span class="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] font-semibold text-gray-500">{{ group.links.length }}</span>
            </div>

            <div class="grid grid-cols-2 gap-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5">
              <a
                v-for="link in group.links"
                :key="`${group.name}-${link.url}`"
                :href="link.url"
                target="_blank"
                rel="noopener noreferrer"
                class="group rounded-md border border-gray-200 bg-white px-2.5 py-2 transition hover:border-blue-200 hover:bg-blue-50/20"
              >
                <div class="flex items-center gap-2">
                  <div class="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-md border border-gray-100 bg-gray-50">
                    <img
                      v-if="link.logoUrl"
                      :src="link.logoUrl"
                      :alt="link.name"
                      class="h-full w-full object-cover"
                      loading="lazy"
                    />
                    <Link v-else class="h-4 w-4 text-gray-400" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex min-w-0 items-center gap-2">
                      <h3 class="truncate text-[13px] font-semibold text-gray-950 group-hover:text-blue-600">{{ link.name }}</h3>
                      <ExternalLink class="h-3 w-3 shrink-0 text-gray-300 group-hover:text-blue-500" />
                    </div>
                    <p class="mt-0.5 truncate text-[11px] leading-4 text-gray-500">{{ link.desc || link.url }}</p>
                  </div>
                </div>
              </a>
            </div>
          </section>

          <div v-if="!props.groups.length" class="rounded-lg border border-gray-200 bg-white px-5 py-16 text-center">
            <Link class="mx-auto h-8 w-8 text-gray-300" />
            <h2 class="mt-3 text-base font-semibold text-gray-900">暂无链接</h2>
            <p class="mt-1 text-sm text-gray-500">站点还没有配置友情链接。</p>
          </div>
        </div>

        <aside class="space-y-3">
          <div class="rounded-lg border border-gray-200/70 bg-white p-4">
            <h2 class="text-sm font-semibold text-gray-950">申请友链</h2>
            <p class="mt-2 text-sm leading-6 text-gray-500">发帖提交网站名称、描述、地址和 Logo，管理员审核后会出现在这里。</p>
            <a href="/publish" class="mt-4 inline-flex h-9 items-center gap-1.5 rounded-md bg-blue-600 px-3 text-sm font-semibold text-white hover:bg-blue-700">
              <Send class="h-4 w-4" />
              去发帖申请
            </a>
          </div>

          <div class="rounded-lg border border-gray-200/70 bg-white p-4">
            <h2 class="text-sm font-semibold text-gray-950">收录原则</h2>
            <div class="mt-3 space-y-2 text-sm text-gray-600">
              <div class="flex gap-2">
                <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" />
                <span>内容健康、长期可访问。</span>
              </div>
              <div class="flex gap-2">
                <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" />
                <span>优先技术、开源、社区相关站点。</span>
              </div>
              <div class="flex gap-2">
                <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" />
                <span>站点信息清晰，Logo 可稳定加载。</span>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>
  </AppShell>
</template>
