<script setup lang="ts">
import type {DataTableColumns} from 'naive-ui'
import {NButton, NDataTable, NSpace, useMessage} from 'naive-ui'
import {h, onMounted, reactive, ref} from 'vue'
import {getAdminArticlesList} from '@/service/request'

const message = useMessage()
type ArticleItem = {
  id: string
  title: string
  type: string
  userId: string
  username: string
  articleStatus: string
  processStatus: string
  createdAt: string
  updatedAt: string
}
const data = ref<ArticleItem[]>([])


const createColumns = ({
                         play
                       }: {
  play: (row: ArticleItem) => void
}): DataTableColumns<ArticleItem> => {
  return [
    {
      title: 'id',
      key: 'id',
      width: "60px",
    },
    {
      title: 'Title',
      key: 'title',
      width: "160px", ellipsis: true
    }, {
      title: "type",
      key: "type"
    },
    {
      title: "用户名",
      key: "username"
    },
    {
      title: '文章状态',
      key: 'articleStatus'
    }, {
      title: "锁定状态",
      key: "processStatus"
    },
    {
      title: '创建时间',
      key: 'createdAt'
    },
    {
      title: '修改时间',
      key: 'updatedAt'
    },
    {
      title: 'Action',
      key: 'actions',
      render(row: ArticleItem) {
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
  play(row: ArticleItem) {
    message.info(`Play ${row.title}`)
  }
})
let pagination = true
const paginationReactive = reactive({
  page: 1,
  pageCount: 1,
  pageSize: 20,
  itemCount: 0,
  search: "",
  prefix({itemCount}) {
    return `Total is ${itemCount}.`
  }
})
onMounted(async () => {
  await searchPage(paginationReactive.page)
})

async function searchPage(page: number) {
  try {
    let res = await getAdminArticlesList(page)
    data.value = res.result.list
    paginationReactive.page = page
    paginationReactive.pageCount = parseInt(String(res.result.total / res.result.size))
    paginationReactive.itemCount = res.result.total
  } catch (e) {
    console.error(e)
  }
}
</script>
<template>
  <n-space vertical>
    <n-data-table
        remote
        :columns="columns"
        :data="data"
        :pagination="paginationReactive"
        :bordered="false"
        @update:page="searchPage"
        flex-height :style="{ height: `600px` }" striped
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
