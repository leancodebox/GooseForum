<template>
  <div class="space-y-6">
    <!-- 页面标题和操作按钮 -->
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900">分类管理</h1>
      <button
          @click="openCreateModal"
          class="btn btn-primary"
      >
        <PlusIcon class="w-5 h-5 mr-2"/>
        新增分类
      </button>
    </div>

    <!-- 分类列表 -->
    <div class="bg-white rounded-lg shadow">
      <div class="overflow-x-auto">
        <!-- 加载状态 -->
        <div v-if="loading" class="flex justify-center items-center py-12">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
        
        <!-- 空状态 -->
        <div v-else-if="categories.length === 0" class="text-center py-12">
          <div class="text-gray-500 mb-4">暂无分类数据</div>
          <button @click="openCreateModal" class="btn btn-primary btn-sm">
            <PlusIcon class="w-4 h-4 mr-1"/>
            创建第一个分类
          </button>
        </div>
        
        <!-- 分类表格 -->
        <table v-else class="table table-zebra w-full">
          <thead>
          <tr>
            <th class="text-left">ID</th>
            <th class="text-left">分类名称</th>
            <th class="text-left">排序</th>
            <th class="text-left">状态</th>
            <th class="text-left">操作</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="category in categories" :key="category.id" class="hover">
            <td class="font-mono text-sm">{{ category.id }}</td>
            <td class="font-medium">{{ category.category }}</td>
            <td>
              <span class="badge badge-outline">{{ category.sort }}</span>
            </td>
            <td>
              <span class="badge whitespace-nowrap" :class="getStatusClass(category.status)">
                {{ getStatusText(category.status) }}
              </span>
            </td>
            <td>
              <div class="flex gap-2">
                <button 
                  @click="openEditModal(category)"
                  class="btn btn-sm btn-outline btn-primary"
                  title="编辑分类"
                >
                  <PencilIcon class="w-4 h-4"/>
                </button>
                <button 
                  @click="openDeleteModal(category)"
                  class="btn btn-sm btn-outline btn-error"
                  title="删除分类"
                >
                  <TrashIcon class="w-4 h-4"/>
                </button>
              </div>
            </td>
          </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 分类编辑模态框 -->
    <dialog ref="categoryModal" class="modal">
      <div class="modal-box w-11/12 max-w-md">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2" @click="closeModal">✕</button>
        </form>
        
        <h3 class="font-bold text-lg mb-4">
          {{ modalMode === 'create' ? '新增分类' : '编辑分类' }}
        </h3>
        
        <div class="space-y-4">
          <!-- 分类名称 -->
          <div class="form-control">
            <label class="label">
              <span class="label-text font-medium">分类名称 <span class="text-error">*</span></span>
            </label>
            <input 
              v-model="formData.category"
              type="text" 
              placeholder="请输入分类名称"
              class="input input-bordered w-full"
              :disabled="saving"
            />
          </div>
          
          <!-- 排序 -->
          <div class="form-control">
            <label class="label">
              <span class="label-text font-medium">排序</span>
            </label>
            <input 
              v-model.number="formData.sort"
              type="number" 
              placeholder="请输入排序值"
              class="input input-bordered w-full"
              :disabled="saving"
              min="0"
            />
          </div>
          
          <!-- 状态 -->
          <div class="form-control">
            <label class="label">
              <span class="label-text font-medium">状态</span>
            </label>
            <select v-model.number="formData.status" class="select select-bordered w-full" :disabled="saving">
              <option :value="1">启用</option>
              <option :value="0">禁用</option>
            </select>
          </div>
        </div>
        
        <div class="modal-action">
          <button 
            @click="closeModal"
            class="btn btn-outline"
            :disabled="saving"
          >
            取消
          </button>
          <button 
            @click="saveCategory_"
            class="btn btn-primary"
            :disabled="saving"
          >
            <span v-if="saving" class="loading loading-spinner loading-sm mr-2"></span>
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="closeModal">close</button>
      </form>
    </dialog>

    <!-- 删除确认模态框 -->
    <dialog ref="deleteModal" class="modal">
      <div class="modal-box w-11/12 max-w-sm">
        <h3 class="font-bold text-lg mb-4 text-error">确认删除</h3>
        
        <div class="mb-6">
          <p class="text-gray-600 mb-2">您确定要删除以下分类吗？</p>
          <div v-if="categoryToDelete" class="bg-gray-50 p-3 rounded-lg">
            <div class="font-medium">{{ categoryToDelete.category }}</div>
            <div class="text-sm text-gray-500">ID: {{ categoryToDelete.id }}</div>
          </div>
          <p class="text-sm text-error mt-2">此操作不可撤销！</p>
        </div>
        
        <div class="modal-action">
          <button 
            @click="closeDeleteModal"
            class="btn btn-outline"
            :disabled="deleting"
          >
            取消
          </button>
          <button 
            @click="confirmDelete"
            class="btn btn-error"
            :disabled="deleting"
          >
            <span v-if="deleting" class="loading loading-spinner loading-sm mr-2"></span>
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="closeDeleteModal">close</button>
      </form>
    </dialog>
  </div>
</template>

<script setup lang="ts">
import {onMounted, ref, reactive} from 'vue'
import {PlusIcon, PencilIcon, TrashIcon} from '@heroicons/vue/24/outline'
import {getCategoryList,saveCategory,deleteCategory} from "@/admin/utils/adminService.ts";
import type {Category} from "@/admin/utils/adminInterfaces.ts";

// 响应式数据
const categories = ref<Category[]>([])
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const editingCategory = ref<Category | null>(null)
const categoryModal = ref<HTMLDialogElement>()
const deleteModal = ref<HTMLDialogElement>()
const modalMode = ref<'create' | 'edit'>('create')
const categoryToDelete = ref<Category | null>(null)

// 表单数据
const formData = reactive({
  id: 0,
  category: '',
  sort: 0,
  status: 1
})

// 方法
const fetchCategories = async () => {
  loading.value = true
  try {
    let resp = await getCategoryList()
    categories.value = resp.result
  } catch (error) {
    console.error('获取分类列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 组件挂载时获取数据
onMounted(() => {
  fetchCategories()
})

const resetForm = () => {
  formData.id = 0
  formData.category = ''
  formData.sort = 0
  formData.status = 1
}

const openCreateModal = () => {
  modalMode.value = 'create'
  resetForm()
  editingCategory.value = null
  categoryModal.value?.showModal()
}

const openEditModal = (category: Category) => {
  modalMode.value = 'edit'
  editingCategory.value = category
  formData.id = category.id
  formData.category = category.category
  formData.sort = category.sort
  formData.status = category.status
  categoryModal.value?.showModal()
}

const closeModal = () => {
  categoryModal.value?.close()
  resetForm()
}

const saveCategory_ = async () => {
  if (!formData.category.trim()) {
    alert('请输入分类名称')
    return
  }
  
  saving.value = true
  try {
    const id = modalMode.value === 'edit' ? formData.id : 0
    await saveCategory(id, formData.category, formData.sort, formData.status)
    closeModal()
    await fetchCategories()
  } catch (error) {
    console.error('保存分类失败:', error)
    alert('保存失败，请重试')
  } finally {
    saving.value = false
  }
}

const openDeleteModal = (category: Category) => {
  categoryToDelete.value = category
  deleteModal.value?.showModal()
}

const closeDeleteModal = () => {
  deleteModal.value?.close()
  categoryToDelete.value = null
}

const confirmDelete = async () => {
  if (!categoryToDelete.value) return
  
  deleting.value = true
  try {
    await deleteCategory(categoryToDelete.value.id)
    closeDeleteModal()
    await fetchCategories()
  } catch (error) {
    console.error('删除分类失败:', error)
    alert(error.message)
  } finally {
    deleting.value = false
  }
}

const getStatusText = (status: number) => {
  return status === 0 ? '启用' : '禁用'
}

const getStatusClass = (status: number) => {
  return status === 0 ? 'badge-success' : 'badge-error'
}
</script>

<style scoped>
</style>