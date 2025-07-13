<template>
  <div class="directory-tree-editor">
    <!-- 工具栏 -->
    <div class="flex justify-between items-center mb-4">
      <h4 class="font-semibold">目录结构编辑器</h4>
      <div class="flex gap-2">
        <button class="btn btn-sm btn-outline" @click="addRootItem">
          <PlusIcon class="w-4 h-4"/>
          添加根目录
        </button>
        <button class="btn btn-sm btn-ghost" @click="expandAll">
          <ChevronDownIcon class="w-4 h-4"/>
          展开全部
        </button>
        <button class="btn btn-sm btn-ghost" @click="collapseAll">
          <ChevronRightIcon class="w-4 h-4"/>
          收起全部
        </button>
      </div>
    </div>

    <!-- 树形结构 -->
    <div class="border border-base-300 rounded-lg p-4 bg-base-50 min-h-[300px]">
      <div v-if="treeData.length === 0" class="text-center text-base-content/50 py-8">
        <FolderIcon class="w-12 h-12 mx-auto mb-2 opacity-50"/>
        <p>暂无目录结构，点击"添加根目录"开始创建</p>
      </div>
      
      <draggable
        v-model="treeData"
        group="directory"
        item-key="id"
        class="space-y-1"
        @change="onTreeChange"
      >
        <template #item="{ element, index }">
          <DirectoryTreeItem
            :item="element"
            :level="0"
            @update="updateItem"
            @delete="deleteItem"
            @add-child="addChild"
          />
        </template>
      </draggable>
    </div>

    <!-- JSON预览 -->
    <div class="mt-4">
      <div class="flex justify-between items-center mb-2">
        <label class="font-semibold">JSON预览</label>
        <button class="btn btn-xs btn-ghost" @click="copyToClipboard">
          <DocumentDuplicateIcon class="w-3 h-3"/>
          复制
        </button>
      </div>
      <textarea
        :value="jsonPreview"
        readonly
        class="textarea textarea-bordered w-full h-32 font-mono text-xs"
        placeholder="JSON结构将在这里显示"
      ></textarea>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import draggable from 'vuedraggable'
import DirectoryTreeItem from './DirectoryTreeItem.vue'
import {
  PlusIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  FolderIcon,
  DocumentDuplicateIcon
} from '@heroicons/vue/24/outline'

const props = defineProps({
  modelValue: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue'])

// 树形数据
const treeData = ref([])
let nextId = 1

// JSON预览
const jsonPreview = computed(() => {
  if (treeData.value.length === 0) return ''
  return JSON.stringify(convertToApiFormat(treeData.value), null, 2)
})

// 防止循环更新的标志
let isUpdatingFromProps = false

// 监听树形数据变化，更新父组件
watch(treeData, (newValue) => {
  if (!isUpdatingFromProps) {
    emit('update:modelValue', convertToApiFormat(newValue))
  }
}, { deep: true })

// 初始化数据
watch(() => props.modelValue, (newValue) => {
  if (newValue && Array.isArray(newValue)) {
    isUpdatingFromProps = true
    treeData.value = convertFromApiFormat(newValue)
    nextTick(() => {
      isUpdatingFromProps = false
    })
  }
}, { immediate: true })

// 转换为API格式
function convertToApiFormat(items) {
  return items.map(item => {
    const result = {
      title: item.title,
      slug: item.slug,
      description: item.description || ''
    }
    if (item.children && item.children.length > 0) {
      result.children = convertToApiFormat(item.children)
    }
    return result
  })
}

// 从API格式转换（添加内部属性）
function convertFromApiFormat(items) {
  return items.map(item => {
    const result = {
      id: nextId++,
      title: item.title || '',
      slug: item.slug || '',
      description: item.description || '',
      expanded: true,
      children: []
    }
    if (item.children && item.children.length > 0) {
      result.children = convertFromApiFormat(item.children)
    }
    return result
  })
}

// 添加根目录项
function addRootItem() {
  const newItem = {
    id: nextId++,
    title: '新目录',
    slug: 'new-directory',
    description: '',
    expanded: true,
    children: []
  }
  treeData.value.push(newItem)
}

// 更新项目
function updateItem(id, updates) {
  updateItemRecursive(treeData.value, id, updates)
}

function updateItemRecursive(items, id, updates) {
  for (const item of items) {
    if (item.id === id) {
      Object.assign(item, updates)
      return true
    }
    if (item.children && updateItemRecursive(item.children, id, updates)) {
      return true
    }
  }
  return false
}

// 删除项目
function deleteItem(id) {
  deleteItemRecursive(treeData.value, id)
}

function deleteItemRecursive(items, id) {
  for (let i = 0; i < items.length; i++) {
    if (items[i].id === id) {
      items.splice(i, 1)
      return true
    }
    if (items[i].children && deleteItemRecursive(items[i].children, id)) {
      return true
    }
  }
  return false
}

// 添加子项目
function addChild(parentId) {
  const newItem = {
    id: nextId++,
    title: '新子目录',
    slug: 'new-subdirectory',
    description: '',
    expanded: true,
    children: []
  }
  addChildRecursive(treeData.value, parentId, newItem)
}

function addChildRecursive(items, parentId, newItem) {
  for (const item of items) {
    if (item.id === parentId) {
      item.children.push(newItem)
      item.expanded = true
      return true
    }
    if (item.children && addChildRecursive(item.children, parentId, newItem)) {
      return true
    }
  }
  return false
}

// 展开/收起全部
function expandAll() {
  setExpandedRecursive(treeData.value, true)
}

function collapseAll() {
  setExpandedRecursive(treeData.value, false)
}

function setExpandedRecursive(items, expanded) {
  items.forEach(item => {
    item.expanded = expanded
    if (item.children) {
      setExpandedRecursive(item.children, expanded)
    }
  })
}

// 树形结构变化
function onTreeChange() {
  // 拖拽排序后的处理
}

// 复制到剪贴板
function copyToClipboard() {
  navigator.clipboard.writeText(jsonPreview.value).then(() => {
    // 可以添加提示
  })
}
</script>

<style scoped>
.directory-tree-editor {
  @apply w-full;
}

.bg-base-50 {
  background-color: rgb(248 250 252);
}
</style>