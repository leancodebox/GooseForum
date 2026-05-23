<script setup lang="ts">
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
    label: '注册用户',
    color: 'var(--chart-1)',
  },
  articleCount: {
    label: '新增文章',
    color: 'var(--chart-2)',
  },
  replyCount: {
    label: '新增回复',
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
  <Card class="h-full pt-0">
    <CardHeader class="flex items-center gap-2 space-y-0 border-b py-5 sm:flex-row">
      <div class="grid flex-1 gap-1">
        <CardTitle>流量概览</CardTitle>
        <CardDescription>展示注册用户、文章和回复的增长趋势</CardDescription>
      </div>
      <slot name="headerAction" />
    </CardHeader>
    <CardContent class="px-2 pt-4 pb-4 sm:px-6 sm:pt-6">
      <div v-if="loading" class="flex h-[350px] items-center justify-center text-muted-foreground">
        加载中...
      </div>
      <ChartContainer v-else :config="chartConfig" class="aspect-auto h-[300px] w-full" :cursor="false">
        <VisXYContainer
          :data="chartData"
          :svg-defs="svgDefs"
          :margin="{ top: 12, right: 12, left: -40, bottom: 24 }"
        >
          <VisArea
            :x="(d: Data) => d.date"
            :y="[(d: Data) => d.regCount, (d: Data) => d.articleCount, (d: Data) => d.replyCount]"
            :color="(_d: Data, i: number) => ['url(#fillRegCount)', 'url(#fillArticleCount)', 'url(#fillReplyCount)'][i]"
            :opacity="0.6"
          />
          <VisLine
            :x="(d: Data) => d.date"
            :y="[(d: Data) => d.regCount, (d: Data) => d.articleCount, (d: Data) => d.replyCount]"
            :color="(_d: Data, i: number) => [chartConfig.regCount.color, chartConfig.articleCount.color, chartConfig.replyCount.color][i]"
            :line-width="1"
          />
          <VisAxis
            type="x"
            :x="(d: Data) => d.date"
            :tick-line="false"
            :domain-line="false"
            :grid-line="false"
            :num-ticks="6"
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
            :num-ticks="4"
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
