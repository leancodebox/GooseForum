<script setup lang="ts">
import {
    NButton,
    NCard,
    NGrid,
    NGridItem,
    NIcon,
    NSpace,
    NStatistic,
    useMessage,
    NDataTable
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {h, ref} from 'vue'

const message = useMessage()
type Song = {
  no: number
  title: string
  length: string
}
const data: Song[] = [
  { no: 3, title: 'Wonderwall', length: '4:18' },
  { no: 4, title: "Don't Look Back in Anger", length: '4:48' },
  { no: 12, title: 'Champagne Supernova', length: '7:27' }
]


const createColumns = ({
                         play
                       }: {
  play: (row: Song) => void
}): DataTableColumns<Song> => {
  return [
    {
      title: 'No',
      key: 'no'
    },
    {
      title: 'Title',
      key: 'title'
    },
    {
      title: 'Length',
      key: 'length'
    },
    {
      title: 'Action',
      key: 'actions',
      render (row) {
        return h(
            NButton,
            {
              strong: true,
              tertiary: true,
              size: 'small',
              onClick: () => play(row)
            },
            { default: () => 'Play' }
        )
      }
    }
  ]
}
let columns =  createColumns({
  play (row: Song) {
    message.info(`Play ${row.title}`)
  }
})
let pagination =  true
</script>
<template>
    <n-space vertical>
      <n-data-table
          :columns="columns"
          :data="data"
          :pagination="pagination"
          :bordered="false"
      />
    </n-space>
</template>
<style>
.carousel-img {
    width: 100%;
    height: 240px;
    object-fit: cover;
}
</style>
