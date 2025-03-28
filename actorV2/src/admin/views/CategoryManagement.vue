<template>
  <div>
    <div class="action-bar">
      <n-space>
        <n-button type="primary" @click="handleAddCategory">
          <template #icon>
            <n-icon>
              <AddOutline />
            </n-icon>
          </template>
          添加分类
        </n-button>
      </n-space>
    </div>

    <n-data-table
      :columns="columns"
      :data="categories"
      :pagination="pagination"
      :bordered="false"
      striped
    />

    <!-- 添加/编辑分类对话框 -->
    <n-modal v-model:show="showCategoryModal" :title="isEditing ? '编辑分类' : '添加分类'">
      <n-card>
        <n-form
          ref="categoryFormRef"
          :model="categoryForm"
          :rules="categoryRules"
          label-placement="left"
          label-width="auto"
          require-mark-placement="right-hanging"
        >
          <n-form-item path="name" label="分类名称">
            <n-input v-model:value="categoryForm.name" placeholder="请输入分类名称" />
          </n-form-item>
          <n-form-item path="slug" label="分类标识">
            <n-input v-model:value="categoryForm.slug" placeholder="请输入分类标识" />
          </n-form-item>
          <n-form-item path="description" label="分类描述">
            <n-input
              v-model:value="categoryForm.description"
              type="textarea"
              placeholder="请输入分类描述"
            />
          </n-form-item>
          <n-form-item path="order" label="排序">
            <n-input-number v-model:value="categoryForm.order" :min="0" />
          </n-form-item>
          <n-form-item path="status" label="状态">
            <n-switch v-model:value="categoryForm.status" />
          </n-form-item>
        </n-form>
        <div class="action-buttons">
          <n-space justify="end">
            <n-button @click="showCategoryModal = false">取消</n-button>
            <n-button type="primary" @click="handleSaveCategory">保存</n-button>
          </n-space>
        </div>
      </n-card>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { h, ref, reactive } from 'vue'
import { useMessage } from 'naive-ui'
import { AddOutline, CreateOutline, TrashOutline } from '@vicons/ionicons5'
import type { FormInst, FormRules } from 'naive-ui'

const message = useMessage()
const showCategoryModal = ref(false)
const isEditing = ref(false)
const categoryFormRef = ref<FormInst | null>(null)

// 分类表单
const categoryForm = reactive({
  id: '',
  name: '',
  slug: '',
  description: '',
  order: 0,
  status: true
})

// 表单验证规则
const categoryRules: FormRules = {
  name: [
    { required: true, message: '请输入分类名称', trigger: 'blur' }
  ],
  slug: [
    { required: true, message: '请输入分类标识', trigger: 'blur' },
    { pattern: /^[a-z0-9-]+$/, message: '分类标识只能包含小写字母、数字和连字符', trigger: 'blur' }
  ]
}

// 分页设置
const pagination = {
  pageSize: 10
}

// 模拟分类数据
const categories = ref([
  { id: '1', name: '技术讨论', slug: 'tech', description: '讨论各种技术问题', order: 0, status: true, postCount: 42, createdAt: '2023-01-01' },
  { id: '2', name: '问答', slug: 'qa', description: '提问和回答', order: 1, status: true, postCount: 128, createdAt: '2023-01-02' },
  { id: '3', name: '分享', slug: 'share', description: '分享各种资源和经验', order: 2, status: true, postCount: 64, createdAt: '2023-01-03' },
  { id: '4', name: '公告', slug: 'announcement', description: '系统公告', order: 3, status: true, postCount: 10, createdAt: '2023-01-04' },
  { id: '5', name: '闲聊', slug: 'chat', description: '随意聊天', order: 4, status: false, postCount: 0, createdAt: '2023-01-05' }
])

// 表格列定义
const columns = [
  {
    title: 'ID',
    key: 'id'
  },
  {
    title: '分类名称',
    key: 'name'
  },
  {
    title: '分类标识',
    key: 'slug'
  },
  {
    title: '描述',
    key: 'description',
    ellipsis: {
      tooltip: true
    }
  },
  {
    title: '排序',
    key: 'order'
  },
  {
    title: '状态',
    key: 'status',
    render(row) {
      return h(
        'n-tag',
        { type: row.status ? 'success' : 'warning' },
        { default: () => row.status ? '启用' : '禁用' }
      )
    }
  },
  {
    title: '帖子数',
    key: 'postCount'
  },
  {
    title: '创建时间',
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
                onClick: () => handleEditCategory(row)
              },
              {
                default: () => '编辑',
                icon: () => h(CreateOutline)
              }
            ),
            h(
              'n-button',
              {
                size: 'small',
                quaternary: true,
                type: 'error',
                onClick: () => handleDeleteCategory(row)
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

// 添加分类
const handleAddCategory = () => {
  isEditing.value = false
  categoryForm.id = ''
  categoryForm.name = ''
  categoryForm.slug = ''
  categoryForm.description = ''
  categoryForm.order = categories.value.length
  categoryForm.status = true
  showCategoryModal.value = true
}

// 编辑分类
const handleEditCategory = (category) => {
  isEditing.value = true
  categoryForm.id = category.id
  categoryForm.name = category.name
  categoryForm.slug = category.slug
  categoryForm.description = category.description
  categoryForm.order = category.order
  categoryForm.status = category.status
  showCategoryModal.value = true
}

// 删除分类
const handleDeleteCategory = (category) => {
  // 实际应用中应该弹出确认对话框
  if (category.postCount > 0) {
    message.error(`无法删除分类 "${category.name}"，该分类下还有 ${category.postCount} 篇帖子`)
    return
  }
  
  message.success(`分类 "${category.name}" 已删除`)
  categories.value = categories.value.filter(c => c.id !== category.id)
}

// 保存分类
const handleSaveCategory = () => {
  categoryFormRef.value?.validate((errors) => {
    if (!errors) {
      if (isEditing.value) {
        // 更新分类
        const index = categories.value.findIndex(c => c.id === categoryForm.id)
        if (index !== -1) {
          categories.value[index] = { 
            ...categories.value[index], 
            name: categoryForm.name,
            slug: categoryForm.slug,
            description: categoryForm.description,
            order: categoryForm.order,
            status: categoryForm.status
          }
          message.success('分类更新成功')
        }
      } else {
        // 添加分类
        const newId = (parseInt(categories.value[categories.value.length - 1].id) + 1).toString()
        categories.value.push({
          id: newId,
          name: categoryForm.name,
          slug: categoryForm.slug,
          description: categoryForm.description,
          order: categoryForm.order,
          status: categoryForm.status,
          postCount: 0,
          createdAt: new Date().toISOString().split('T')[0]
        })
        message.success('分类添加成功')
      }
      showCategoryModal.value = false
    }
  })
}
</script>

<style scoped>
.action-bar {
  margin-bottom: 16px;
}

.action-buttons {
  margin-top: 24px;
}
</style>