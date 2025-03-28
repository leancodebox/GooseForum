<script setup lang="ts">
import { h, ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { SearchOutline, EyeOutline, TrashOutline, CheckmarkCircleOutline, BanOutline } from '@vicons/ionicons5'

const message = useMessage()
const searchText = ref('')
const categoryFilter = ref(null)
const statusFilter = ref(null)

// 分页设置
const pagination = {
  pageSize: 10
}

// 分类选项
const categoryOptions = [
  { label: '技术讨论', value: 'tech' },
  { label: '问答', value: 'qa' },
  { label: '分享', value: 'share' },
  { label: '公告', value: 'announcement' }
]

// 状态选项
const statusOptions = [
  { label: '已发布', value: 'published' },
  { label: '待审核', value: 'pending' },
  { label: '已删除', value: 'deleted' }
]

// 模拟帖子数据
const posts = ref([
  { id: '1', title: 'Vue3 最佳实践', author: 'admin', category: 'tech', status: 'published', views: 1024, comments: 42, createdAt: '2023-01-01' },
  { id: '2', title: '如何使用 Naive UI', author: 'user1', category: 'qa', status: 'published', views: 768, comments: 23, createdAt: '2023-01-02' },
  { id: '3', title: '分享一个好用的工具', author: 'user2', category: 'share', status: 'pending', views: 0, comments: 0, createdAt: '2023-01-03' },
  { id: '4', title: '系统维护公告', author: 'admin', category: 'announcement', status: 'published', views: 2048, comments: 15, createdAt: '2023-01-04' },
  { id: '5', title: '被举报的帖子', author: 'user3', category: 'share', status: 'deleted', views: 128, comments: 5, createdAt: '2023-01-05' }
])

// 过滤后的帖子列表
const filteredPosts = computed(() => {
  let result = posts.value

  // 搜索过滤
  if (searchText.value) {
    result = result.filter(post =>
        post.title.toLowerCase().includes(searchText.value.toLowerCase()) ||
        post.author.toLowerCase().includes(searchText.value.toLowerCase())
    )
  }

  // 分类过滤
  if (categoryFilter.value) {
    result = result.filter(post => post.category === categoryFilter.value)
  }

  // 状态过滤
  if (statusFilter.value) {
    result = result.filter(post => post.status === statusFilter.value)
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
    key: 'author'
  },
  {
    title: '分类',
    key: 'category',
    render(row) {
      const categoryMap = {
        tech: '技术讨论',
        qa: '问答',
        share: '分享',
        announcement: '公告'
      }
      return categoryMap[row.category] || row.category
    }
  },
  {
    title: '状态',
    key: 'status',
    render(row) {
      const statusMap = {
        published: { text: '已发布', type: 'success' },
        pending: { text: '待审核', type: 'warning' },
        deleted: { text: '已删除', type: 'error' }
      }
      const status = statusMap[row.status] || { text: row.status, type: 'default' }
      return h(
          'n-tag',
          { type: status.type },
          { default: () => status.text }
      )
    }
  },
  {
    title: '浏览量',
    key: 'views'
  },
  {
    title: '评论数',
    key: 'comments'
  },
  {
    title: '发布时间',
    key: 'createdAt'
  },
  {
    title: '操作',
    key: 'actions',
    render(row) {
      return h(
          'n-space',
          {},
          {
            default: () => [
              h(
                  'n-button',
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
              row.status === 'pending' ? h(
                  'n-button',
                  {
                    size: 'small',
                    quaternary: true,
                    type: 'success',
                    onClick: () => handleApprovePost(row)
                  },
                  {
                    default: () => '通过',
                    icon: () => h(CheckmarkCircleOutline)
                  }
              ) : null,
              row.status !== 'deleted' ? h(
                  'n-button',
                  {
                    size: 'small',
                    quaternary: true,
                    type: 'warning',
                    onClick: () => handleBlockPost(row)
                  },
                  {
                    default: () => '屏蔽',
                    icon: () => h(BanOutline)
                  }
              ) : null,
              h(
                  'n-button',
                  {
                    size: 'small',
                    quaternary: true,
                    type: 'error',
                    onClick: () => handleDeletePost(row)
                  },
                  {
                    default: () => '删除',
                    icon: () => h(TrashOutline)
                  }
              )
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
}

// 通过帖子
const handleApprovePost = (post) => {
  post.status = 'published'
  message.success(`帖子已通过审核: ${post.title}`)
}

// 屏蔽帖子
const handleBlockPost = (post) => {
  post.status = 'deleted'
  message.warning(`帖子已屏蔽: ${post.title}`)
}

// 删除帖子
const handleDeletePost = (post) => {
  // 实际应用中应该弹出确认对话框
  message.error(`帖子已删除: ${post.title}`)
  posts.value = posts.value.filter(p => p.id !== post.id)
}
</script>
<template>
  <div>
    <div class="action-bar">
      <n-space>
        <n-input v-model:value="searchText" placeholder="搜索帖子" clearable>
          <template #prefix>
            <n-icon>
              <SearchOutline />
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
      :bordered="false"
      striped
    />
  </div>
</template>
<style scoped>
.action-bar {
  margin-bottom: 16px;
}
</style>
