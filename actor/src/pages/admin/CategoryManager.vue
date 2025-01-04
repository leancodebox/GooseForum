<template>
  <n-card :bordered="false">
    <template #header>
      <n-flex justify="space-between">
        <h3 style="margin: 0">分类管理</h3>
        <n-button type="primary" @click="handleAdd">
          添加分类
        </n-button>
      </n-flex>
    </template>

    <n-data-table
        :columns="columns"
        :data="tableData"
        :loading="loading"
    />

    <n-modal v-model:show="showModal" :mask-closable="false">
      <n-card
          style="width: 600px"
          :title="formData.id ? '编辑分类' : '添加分类'"
          :bordered="false"
          size="huge"
          role="dialog"
          aria-modal="true"
      >
        <n-form
            ref="formRef"
            :model="formData"
            :rules="rules"
            label-placement="left"
            label-width="100"
            require-mark-placement="right-hanging"
        >
          <n-form-item label="分类名称" path="category">
            <n-input v-model:value="formData.category" placeholder="请输入分类名称"/>
          </n-form-item>
          <n-form-item label="排序" path="sort">
            <n-input-number v-model:value="formData.sort" placeholder="请输入排序值"/>
          </n-form-item>
          <n-form-item label="状态" path="status">
            <n-switch v-model:value="formData.status" :checked-value="1" :unchecked-value="0">
              <template #checked>
                启用
              </template>
              <template #unchecked>
                禁用
              </template>
            </n-switch>
          </n-form-item>
        </n-form>

        <template #footer>
          <n-flex justify="end" :x-gap="12">
            <n-button @click="closeModal">取消</n-button>
            <n-button type="primary" :loading="submitting" @click="handleSubmit">
              确定
            </n-button>
          </n-flex>
        </template>
      </n-card>
    </n-modal>
  </n-card>
</template>

<script setup>
import {
  NButton,
  NCard,
  NDataTable,
  NFlex,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NSwitch,
  useMessage,
} from 'naive-ui'
import {h, onMounted, ref} from 'vue'
import {getCategoryList, saveCategory, deleteCategory} from '@/service/request'

const message = useMessage()
const loading = ref(false)
const tableData = ref([])
const showModal = ref(false)
const submitting = ref(false)

const formRef = ref(null)
const formData = ref({
  id: 0,
  category: '',
  sort: 0,
  status: 1,
})

const rules = {
  category: {
    required: true,
    message: '请输入分类名称',
    trigger: 'blur'
  }
}

const columns = [
  {
    title: 'ID',
    key: 'id',
    width: 80
  },
  {
    title: '分类名称',
    key: 'category'
  },
  {
    title: '排序',
    key: 'sort',
    width: 100
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render(row) {
      return row.status === 1 ? '启用' : '禁用'
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render(row) {
      return h(NFlex, {justify: 'center', xGap: 12}, {
        default: () => [
          h(
              NButton,
              {
                size: 'small',
                onClick: () => handleEdit(row)
              },
              {default: () => '编辑'}
          ),
          h(
              NButton,
              {
                size: 'small',
                type: 'error',
                onClick: () => handleDelete(row)
              },
              {default: () => '删除'}
          )
        ]
      })
    }
  }
]

async function loadData() {
  loading.value = true
  try {
    const res = await getCategoryList()
    if (res.code === 0) {
      tableData.value = res.result
    }
  } catch (err) {
    console.error('获取分类列表失败:', err)
    message.error('获取分类列表失败')
  } finally {
    loading.value = false
  }
}

function handleAdd() {
  formData.value = {
    id: 0,
    category: '',
    sort: 0,
    status: 1
  }
  showModal.value = true
}

function handleEdit(row) {
  formData.value = {...row}
  showModal.value = true
}

async function handleDelete(row) {
  if (await window.$dialog.warning({
    title: '确认删除',
    content: '确定要删除这个分类吗？',
    positiveText: '确定',
    negativeText: '取消'
  })) {
    try {
      const res = await deleteCategory(row.id)
      if (res.code === 0) {
        message.success('删除成功')
        loadData()
      }
    } catch (err) {
      console.error('删除失败:', err)
      message.error('删除失败')
    }
  }
}

async function handleSubmit() {
  await formRef.value?.validate()
  
  submitting.value = true
  try {
    const res = await saveCategory(formData.value)
    if (res.code === 0) {
      message.success('保存成功')
      closeModal()
      loadData()
    }
  } catch (err) {
    console.error('保存失败:', err)
    message.error('保存失败')
  } finally {
    submitting.value = false
  }
}

function closeModal() {
  showModal.value = false
  formData.value = {
    id: 0,
    category: '',
    sort: 0,
    status: 1
  }
}

onMounted(() => {
  loadData()
})
</script> 