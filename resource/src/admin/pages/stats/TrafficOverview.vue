<script setup lang="ts">import { adminText } from '@/admin/runtime/i18n-text'

import { VisArea, VisAxis, VisLine, VisXYContainer } from '@unovis/vue'
import { computed } from 'vue'
import type { ChartConfig } from '@/admin/components/ui/chart'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/admin/components/ui/card'
import {
  ChartContainer,
  ChartCrosshair,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
  componentToString,
} from '@/admin/components/ui/chart'
import type { DailyTraffic } from '@/admin/types'

const props = defineProps<{
  data: DailyTraffic[]
  loading?: boolean
}>()

type Data = {
  date: Date
  regCount: number
  articleCount: number
  replyCount: number
}

const chartData = computed<Data[]>(() =>
  props.data.map(item => ({
    date: new Date(item.date),
    regCount: item.regCount,
    articleCount: item.articleCount,
    replyCount: item.replyCount,
  })),
)

const chartConfig = {
  regCount: {
    label: adminText('k001p'),
    color: 'var(--chart-1)',
  },
  articleCount: {
    label: adminText('k001q'),
    color: 'var(--chart-2)',
  },
  replyCount: {
    label: adminText('k001r'),
    color: 'var(--chart-4)',
  },
} satisfies ChartConfig

const svgDefs = `
  <linearGradient id="fillRegCount" x1="0" y1="0" x2="0" y2="1">
    <stop offset="5%" stop-color="var(--color-regCount)" stop-opacity="0.8" />
    <stop offset="95%" stop-color="var(--color-regCount)" stop-opacity="0.1" />
  </linearGradient>
  <linearGradient id="fillArticleCount" x1="0" y1="0" x2="0" y2="1">
    <stop offset="5%" stop-color="var(--color-articleCount)" stop-opacity="0.8" />
    <stop offset="95%" stop-color="var(--color-articleCount)" stop-opacity="0.1" />
  </linearGradient>
  <linearGradient id="fillReplyCount" x1="0" y1="0" x2="0" y2="1">
    <stop offset="5%" stop-color="var(--color-replyCount)" stop-opacity="0.8" />
    <stop offset="95%" stop-color="var(--color-replyCount)" stop-opacity="0.1" />
  </linearGradient>
`
</script>

<template>
  <Card class="h-full overflow-hidden pt-0">
    <CardHeader class="grid gap-3 space-y-0 border-b px-4 py-4 sm:flex sm:flex-row sm:items-center sm:px-5">
      <div class="grid min-w-0 flex-1 gap-1">
        <CardTitle>{{ adminText('k001s') }}</CardTitle>
        <CardDescription>{{ adminText('k001t') }}</CardDescription>
      </div>
      <div class="min-w-0 sm:shrink-0">
        <slot name="headerAction" />
      </div>
    </CardHeader>
    <CardContent class="px-3 pb-3 pt-4 sm:px-5">
      <div v-if="loading" class="flex h-[260px] items-center justify-center text-muted-foreground sm:h-[320px]">
        {{ adminText('k0046') }}
      </div>
      <ChartContainer v-else :config="chartConfig" class="aspect-auto h-[260px] w-full sm:h-[320px]" :cursor="false">
        <VisXYContainer
          :data="chartData"
          :svg-defs="svgDefs"
          :margin="{ top: 18, right: 18, left: -24, bottom: 28 }"
        >
          <VisArea
            :x="(d: Data) => d.date"
            :y="[(d: Data) => d.regCount, (d: Data) => d.articleCount, (d: Data) => d.replyCount]"
            :color="(_d: Data, i: number) => ['url(#fillRegCount)', 'url(#fillArticleCount)', 'url(#fillReplyCount)'][i]"
            :opacity="1"
          />
          <VisLine
            :x="(d: Data) => d.date"
            :y="[
              (d: Data) => d.regCount,
              (d: Data) => d.regCount + d.articleCount,
              (d: Data) => d.regCount + d.articleCount + d.replyCount,
            ]"
            :color="(_d: Data, i: number) => [chartConfig.regCount.color, chartConfig.articleCount.color, chartConfig.replyCount.color][i]"
            :line-width="1"
          />
          <VisAxis
            type="x"
            :x="(d: Data) => d.date"
            :tick-line="false"
            :domain-line="false"
            :grid-line="false"
            :num-ticks="7"
            :tick-format="(d: number) => {
              const date = new Date(d)
              return date.toLocaleDateString('en-US', {
                month: 'short',
                day: 'numeric',
              })
            }"
          />
          <VisAxis
            type="y"
            :num-ticks="5"
            :tick-line="false"
            :domain-line="false"
          />
          <ChartTooltip />
          <ChartCrosshair
            :template="componentToString(chartConfig, ChartTooltipContent, {
              labelFormatter: (d) => {
                const date = new Date(d)
                return date.toLocaleDateString('en-US', {
                  month: 'short',
                  day: 'numeric',
                  year: 'numeric',
                })
              },
            })"
            :color="(_d: Data, i: number) => [chartConfig.regCount.color, chartConfig.articleCount.color, chartConfig.replyCount.color][i % 3]"
          />
        </VisXYContainer>
        <ChartLegendContent />
      </ChartContainer>
    </CardContent>
  </Card>
</template>
