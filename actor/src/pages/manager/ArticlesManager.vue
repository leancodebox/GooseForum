<script setup lang="ts">
import type {DataTableColumns} from 'naive-ui'
import {NButton, NDataTable, NSpace, useMessage} from 'naive-ui'
import {h} from 'vue'

const message = useMessage()
type Song = {
  no: number
  title: string
  length: string
}
const data: Song[] = [
  {no: 3, title: 'Wonderwall', length: '4:18'},
  {no: 4, title: "Don't Look Back in Anger", length: '4:48'},
  {no: 12, title: 'Champagne Supernova', length: '7:27'}
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
      title: '文章状态',
      key: 'status'
    },
    {
      title: '创建时间',
      key: 'length'
    },
    {
      title: '修改时间',
      key: 'length'
    },
    {
      title: 'Action',
      key: 'actions',
      render(row) {
        return [h(
            NButton,
            {
              strong: true,
              tertiary: true,
              size: 'small',
              onClick: () => play(row)
            },
            {default: () => '冻结'}
        ),
          h(
              NButton,
              {
                strong: true,
                tertiary: true,
                size: 'small',
                onClick: () => play(row)
              },
              {default: () => '查看'}
          )]
      }
    }
  ]
}
let columns = createColumns({
  play(row: Song) {
    message.info(`Play ${row.title}`)
  }
})
let pagination = true
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
