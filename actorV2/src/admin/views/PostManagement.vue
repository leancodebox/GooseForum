<script setup lang="ts">
import {computed, h, onMounted, ref} from 'vue'
import {NButton, NSpace, NTag, useMessage} from 'naive-ui'
import {BanOutline, CheckmarkCircleOutline, EyeOutline, SearchOutline, TrashOutline} from '@vicons/ionicons5'
import {getAdminArticlesList,editArticle} from "@/admin/utils/authService.ts";
import type {Articles} from "@/admin/types/adminInterfaces.ts";

const message = useMessage()
const searchText = ref('')
const categoryFilter = ref(null)
const statusFilter = ref(null)
const loading = ref(false)

// 分页设置
const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  onChange: (page: number) => {
    pagination.value.page = page
    fetchArticles()
  }
})

// 分类选项
const categoryOptions = [
  {label: '技术讨论', value: 1},
  {label: '问答', value: 2},
  {label: '分享', value: 3},
  {label: '公告', value: 4}
]

// 状态选项
const statusOptions = [
  {label: '已发布', value: 1},
  {label: '待审核', value: 0},
  {label: '已删除', value: 2}
]

// 文章数据
const posts = ref<Articles[]>([])

// 获取文章列表
const fetchArticles = async () => {
  loading.value = true
  try {
    const response = await getAdminArticlesList(pagination.value.page, pagination.value.pageSize)
    if (response.code === 0) {
      posts.value = response.result.list
      pagination.value.itemCount = response.result.total
    } else {
      message.error(response.message || '获取文章列表失败')
    }
  } catch (error) {
    message.error('获取文章列表失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

// 过滤后的帖子列表
const filteredPosts = computed(() => {
  let result = posts.value

  // 搜索过滤
  if (searchText.value) {
    result = result.filter(post =>
        post.title.toLowerCase().includes(searchText.value.toLowerCase()) ||
        post.username.toLowerCase().includes(searchText.value.toLowerCase())
    )
  }

  // 分类过滤
  if (categoryFilter.value) {
    result = result.filter(post => post.type === categoryFilter.value)
  }

  // 状态过滤
  if (statusFilter.value) {
    result = result.filter(post => post.processStatus === statusFilter.value)
  }

  return result
})

// 表格列定义
const columns = [
  {
    title: 'ID',
    key: 'id'
  },
  {
    title: '标题',
    key: 'title',
    ellipsis: {
      tooltip: true
    }
  },
  {
    title: '作者',
    key: 'username'
  },
  {
    title: '分类',
    key: 'type',
    render(row: Articles) {
      const categoryMap = {
        1: '技术讨论',
        2: '问答',
        3: '分享',
        4: '公告'
      }
      return categoryMap[row.type] || row.type
    }
  },
  {
    title: '状态',
    key: 'processStatus',
    render(row: Articles) {
      const statusMap = {
        1: {text: '已发布', type: 'success'},
        0: {text: '待审核', type: 'warning'},
      }
      const status = statusMap[row.articleStatus] || {text: row.articleStatus, type: 'default'}
      return h(
          NTag,
          {type: status.type},
          {default: () => status.text}
      )
    }
  },
  {
    title: '锁定',
    key: 'processStatus',
    render(row: Articles) {
      const statusMap = {
        1: {text: '已锁定', type: 'warning'},
        0: {text: '未锁定', type: 'success'},
      }
      const status = statusMap[row.processStatus] || {text: row.processStatus, type: 'default'}
      return h(
          NTag,
          {type: status.type},
          {default: () => status.text}
      )
    }
  },
  {
    title: '创建时间',
    key: 'createdAt'
  },
  {
    title: '更新时间',
    key: 'updatedAt'
  },
  {
    title: '操作',
    key: 'actions',
    render(row: Articles) {
      return h(
          NSpace,
          {},
          {
            default: () => [
              h(
                  NButton,
                  {
                    size: 'small',
                    quaternary: true,
                    onClick: () => handleViewPost(row)
                  },
                  {
                    default: () => '查看',
                    icon: () => h(EyeOutline)
                  }
              ),
              row.processStatus === 0 ? h(
                  NButton,
                  {
                    size: 'small',
                    quaternary: true,
                    type: 'warning',
                    onClick: () => handleBlockPost(row)
                  },
                  {
                    default: () => '冻结',
                    icon: () => h(BanOutline)
                  }
              ) : null,
              row.processStatus === 1 ? h(
                  NButton,
                  {
                    size: 'small',
                    quaternary: true,
                    type: 'success',
                    onClick: () => handleBlockPost(row)
                  },
                  {
                    default: () => '解冻',
                    icon: () => h(CheckmarkCircleOutline)
                  }
              ):null
            ]
          }
      )
    }
  }
]

// 查看帖子
const handleViewPost = (post) => {
  message.info(`查看帖子: ${post.title}`)
  // 实际应用中可以跳转到帖子详情页或打开预览对话框
  window.open(`/post/${post.id}`, '_blank')
}

// 屏蔽和解冻
const handleBlockPost = async(row:Articles) => {
  try {
    // 切换状态：0->1 或 1->0
    const newStatus = row.processStatus === 0 ? 1 : 0
    await editArticle(row.id, newStatus)

    // 更新本地数据状态
    row.processStatus = newStatus
    message.success(`${newStatus === 1 ? '冻结' : '解冻'}成功`)

    // 刷新列表
    await fetchArticles()
  } catch (err) {
    message.error('操作失败')
    console.error(err)
  }
}

// 页面加载时获取文章列表
onMounted(() => {
  fetchArticles()
})
</script>
<template>
  <div>
    <div class="action-bar">
      <n-space>
        <n-input v-model:value="searchText" placeholder="搜索帖子" clearable>
          <template #prefix>
            <n-icon>
              <SearchOutline/>
            </n-icon>
          </template>
        </n-input>
        <n-select
            v-model:value="categoryFilter"
            placeholder="分类筛选"
            clearable
            :options="categoryOptions"
            style="width: 200px"
        />
        <n-select
            v-model:value="statusFilter"
            placeholder="状态筛选"
            clearable
            :options="statusOptions"
            style="width: 200px"
        />
      </n-space>
    </div>

    <n-data-table
        :columns="columns"
        :data="filteredPosts"
        :pagination="pagination"
        :bordered="true"
        :loading="loading"
        striped
    />
  </div>
</template>
<style scoped>
.action-bar {
  margin-bottom: 16px;
}
</style>
