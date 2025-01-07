<script setup lang="ts">
import {DataTableColumns, NButton, NDataTable, NTag, useMessage, NCard, NSpace, NGrid, NGridItem} from 'naive-ui'
import {h, onMounted, reactive, ref} from 'vue'
import {getAdminArticlesList, editArticle} from '@/service/request'
import { useRouter } from 'vue-router'

const message = useMessage()
const router = useRouter()
type ArticleItem = {
  id: string
  title: string
  type: string
  userId: string
  username: string
  articleStatus: number
  processStatus: number
  createdAt: string
  updatedAt: string
}

const data = ref<ArticleItem[]>([])
let isSmallScreen = ref(window.innerWidth < 800)

// 检查屏幕尺寸
function checkScreenSize() {
  isSmallScreen.value = window.innerWidth < 800
}

onMounted(() => {
  checkScreenSize()
  window.addEventListener('resize', checkScreenSize)
})

// 状态标签渲染函数
const renderArticleStatus = (status: number) => {
  if (status === 1) {
    return h(NTag, {type: 'success', style: 'margin: 4px 0'}, () => "已发布")
  }
  return h(NTag, {type: 'warning', style: 'margin: 4px 0'}, () => "草稿")
}

const renderProcessStatus = (status: number) => {
  switch (status) {
    case 0:
      return h(NTag, {type: 'success', style: 'margin: 4px 0'}, () => "未锁定")
    case 1:
      return h(NTag, {type: 'warning', style: 'margin: 4px 0'}, () => "锁定")
    case 2:
      return h(NTag, {type: 'warning', style: 'margin: 4px 0'}, () => "机器锁定")
  }
}

// PC端表格列配置
const createColumns = (): DataTableColumns<ArticleItem> => {
  return [
    {
      title: 'Title',
      key: 'title',
      width: "160px",
      ellipsis: true
    },
    {
      title: "作者",
      key: "username"
    },
    {
      title: '状态',
      key: 'articleStatus',
      render: (row) => renderArticleStatus(row.articleStatus)
    },
    {
      title: "锁定状态",
      key: "processStatus",
      render: (row) => renderProcessStatus(row.processStatus)
    },
    {
      title: '创建时间',
      key: 'createdAt'
    },
    {
      title: 'Action',
      key: 'actions',
      fixed: 'right',
      render(row: ArticleItem) {
        return h(NSpace, {}, {
          default: () => [
            h(
                NButton,
                {
                  type: row.processStatus === 0 ? 'warning' : 'success',
                  size: 'small',
                  onClick: () => handleFreeze(row)
                },
                { default: () => row.processStatus === 0 ? '冻结' : '解冻' }
            ),
            h(
                NButton,
                {
                  type: 'primary',
                  size: 'small',
                  onClick: () => handleView(row)
                },
                {default: () => '查看'}
            )
          ]
        })
      }
    }
  ]
}

const columns = createColumns()

const paginationReactive = reactive({
  page: 1,
  pageCount: 1,
  pageSize: 10,
  itemCount: 0,
  search: "",
  prefix({itemCount}) {
    return `共 ${itemCount} 条`
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

// 操作按钮
const handleFreeze = async (row: ArticleItem) => {
  try {
    // 切换状态：0->1 或 1->0
    const newStatus = row.processStatus === 0 ? 1 : 0
    await editArticle(row.id, newStatus)

    // 更新本地数据状态
    row.processStatus = newStatus
    message.success(`${newStatus === 1 ? '冻结' : '解冻'}成功`)

    // 刷新列表
    await searchPage(paginationReactive.page)
  } catch (err) {
    message.error('操作失败')
    console.error(err)
  }
}

const handleView = (row: ArticleItem) => {
  router.push({
    path: '/home/bbs/articlesPage',
    query: {
      id: row.id,
      title: row.title
    }
  })
}
</script>

<template>
  <!-- PC端显示表格 -->
  <div v-if="!isSmallScreen" class="pc-view">
    <n-data-table
        :size="'small'"
        remote
        :columns="columns"
        :data="data"
        :pagination="paginationReactive"
        :bordered="false"
        @update:page="searchPage"
        striped
        flex-height
        style="height: calc(100vh - var(--header-height) - 28px);"
    />
  </div>

  <!-- 移动端显示卡片 -->
  <div v-else class="mobile-view">
    <n-space vertical>
      <n-card
          v-for="item in data"
          :key="item.id"
          class="article-card"
          :bordered="false"
          size="small"
      >
        <n-space vertical>
          <div class="article-title">{{ item.title }}</div>
          <n-space justify="space-between" align="center">
            <span class="article-author">{{ item.username }}</span>
            <div>
              <n-tag :type="item.articleStatus === 1 ? 'success' : 'warning'" style="margin: 4px 0">
                {{ item.articleStatus === 1 ? '已发布' : '草稿' }}
              </n-tag>
              <n-tag :type="item.processStatus === 0 ? 'success' : 'warning'" style="margin: 4px 0">
                {{ item.processStatus === 0 ? '未锁定' : item.processStatus === 1 ? '锁定' : '机器锁定' }}
              </n-tag>
            </div>
          </n-space>
          <div class="article-time">创建于: {{ item.createdAt }}</div>
          <n-space justify="end">
            <n-button
                size="small"
                :type="item.processStatus === 0 ? 'warning' : 'success'"
                @click="handleFreeze(item)"
            >
              {{ item.processStatus === 0 ? '冻结' : '解冻' }}
            </n-button>
            <n-button size="small" type="primary" @click="handleView(item)">
              查看
            </n-button>
          </n-space>
        </n-space>
      </n-card>
    </n-space>

    <!-- 移动端分页 -->
    <div class="mobile-pagination">
      <n-space justify="center" align="center">
        <n-button
          size="small"
          :disabled="paginationReactive.page === 1"
          @click="searchPage(paginationReactive.page - 1)"
        >
          上一页
        </n-button>
        <span>{{ paginationReactive.page }} / {{ paginationReactive.pageCount }}</span>
        <n-button
          size="small"
          :disabled="paginationReactive.page === paginationReactive.pageCount"
          @click="searchPage(paginationReactive.page + 1)"
        >
          下一页
        </n-button>
      </n-space>
    </div>
  </div>
</template>

<style scoped>
.mobile-view {
  padding: 8px;
}

.article-card {
  margin-bottom: 12px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.article-title {
  font-size: 16px;
  font-weight: bold;
  color: #333;
}

.article-author {
  font-size: 14px;
  color: #666;
}

.article-time {
  font-size: 12px;
  color: #999;
}

.mobile-pagination {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px;
  background: white;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.08);
}

/* 为底部分页器留出空间 */
.mobile-view {
  padding-bottom: 60px;
}
</style>
