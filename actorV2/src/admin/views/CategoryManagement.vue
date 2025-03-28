<script setup lang="ts">
import { h, ref, reactive, onMounted } from 'vue'
import { useMessage, NTag, NSpace, NButton, NPopconfirm } from 'naive-ui'
import { AddOutline, CreateOutline, TrashOutline } from '@vicons/ionicons5'
import type { FormInst, FormRules } from 'naive-ui'
import { getCategoryList, saveCategory, deleteCategory } from "@/admin/utils/authService.ts";
import type { Category} from "../types/adminInterfaces.ts";

const message = useMessage()
const showCategoryModal = ref(false)
const isEditing = ref(false)
const categoryFormRef = ref<FormInst | null>(null)
const loading = ref(false)

// 分类表单
const categoryForm = reactive({
  id: 0,
  category: '',
  sort: 0,
  status: 1
})

// 表单验证规则
const categoryRules: FormRules = {
  category: [
    { required: true, message: '请输入分类名称', trigger: 'blur' }
  ],
  sort: [
    { type: 'number', required: true, message: '请输入排序值', trigger: ['blur', 'change'] }
  ]
}

// 分页设置
const pagination = {
  pageSize: 10
}

// 分类数据
const categories = ref([])

// 获取分类列表
const fetchCategories = async () => {
  loading.value = true
  try {
    const res = await getCategoryList()
    if (res.code === 0) {
      categories.value = res.result
    } else {
      message.error(res.message || '获取分类列表失败')
    }
  } catch (error) {
    message.error('获取分类列表失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

// 初始化时获取分类列表
onMounted(() => {
  fetchCategories()
})

// 表格列定义
const columns = [
  {
    title: 'ID',
    key: 'id'
  },
  {
    title: '分类名称',
    key: 'category'
  },
  {
    title: '排序',
    key: 'sort'
  },
  {
    title: '状态',
    key: 'status',
    render(row:Category) {
      return h(
        NTag,
        { type: row.status === 1 ?  'warning':'success'  },
        { default: () => row.status === 1 ?  '禁用':'启用' }
      )
    }
  },
  {
    title: '操作',
    key: 'actions',
    render(row:Category) {
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
                onClick: () => handleEditCategory(row)
              },
              {
                default: () => '编辑',
                icon: () => h(CreateOutline)
              }
            ),
            h(
              NPopconfirm,
              {
                onPositiveClick: () => handleDeleteCategory(row)
              },
              {
                default: () => '确定要删除该分类吗？',
                trigger: () => h(
                  NButton,
                  {
                    size: 'small',
                    quaternary: true,
                    type: 'error'
                  },
                  {
                    default: () => '删除',
                    icon: () => h(TrashOutline)
                  }
                )
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
  categoryForm.id = 0
  categoryForm.category = ''
  categoryForm.sort = 0
  categoryForm.status = 1
  showCategoryModal.value = true
}

// 编辑分类
const handleEditCategory = (category) => {
  isEditing.value = true
  categoryForm.id = category.id
  categoryForm.category = category.category
  categoryForm.sort = category.sort
  categoryForm.status = category.status
  showCategoryModal.value = true
}

// 删除分类
const handleDeleteCategory = async (category) => {
  try {
    const res = await deleteCategory(category.id)
    if (res.code === 0) {
      message.success('分类删除成功')
      fetchCategories() // 重新获取分类列表
    } else {
      message.error(res.message || '删除分类失败')
    }
  } catch (error) {
    message.error('删除分类失败')
    console.error(error)
  }
}

// 保存分类
const handleSaveCategory = () => {
  categoryFormRef.value?.validate(async (errors) => {
    if (!errors) {
      try {
        const res = await saveCategory({
          id: isEditing.value ? categoryForm.id : undefined,
          category: categoryForm.category,
          sort: categoryForm.sort,
          status: categoryForm.status
        })

        if (res.code === 0) {
          message.success(isEditing.value ? '分类更新成功' : '分类添加成功')
          showCategoryModal.value = false
          fetchCategories() // 重新获取分类列表
        } else {
          message.error(res.message || '保存分类失败')
        }
      } catch (error) {
        message.error('保存分类失败')
        console.error(error)
      }
    }
  })
}
</script>
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
      :loading="loading"
      striped
    />

    <!-- 添加/编辑分类对话框 -->
    <n-modal 
      v-model:show="showCategoryModal" 
      :title="isEditing ? '编辑分类' : '添加分类'"
      style="width: 500px;"
      preset="card"
    >
      <n-card>
        <n-form
          ref="categoryFormRef"
          :model="categoryForm"
          :rules="categoryRules"
          label-placement="left"
          label-width="auto"
          require-mark-placement="right-hanging"
        >
          <n-form-item path="category" label="分类名称">
            <n-input v-model:value="categoryForm.category" placeholder="请输入分类名称" />
          </n-form-item>
          <n-form-item path="sort" label="排序">
            <n-input-number v-model:value="categoryForm.sort" :min="0" />
          </n-form-item>
          <n-form-item path="status" label="状态">
            <n-radio-group v-model:value="categoryForm.status">
              <n-radio :value="1">启用</n-radio>
              <n-radio :value="0">禁用</n-radio>
            </n-radio-group>
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

<style scoped>
.action-bar {
  margin-bottom: 16px;
}

.action-buttons {
  margin-top: 24px;
}
</style>
