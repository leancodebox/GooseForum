<template>
  <div class="directory-tree-item">
    <!-- 当前项目 -->
    <div class="flex items-center gap-2 p-2 rounded hover:bg-base-100 group" :style="{ paddingLeft: `${level * 20 + 8}px` }">
      <!-- 展开/收起按钮 -->
      <button
        v-if="item.children && item.children.length > 0"
        class="btn btn-xs btn-ghost p-0 w-5 h-5"
        @click="toggleExpanded"
      >
        <ChevronRightIcon v-if="!item.expanded" class="w-3 h-3"/>
        <ChevronDownIcon v-else class="w-3 h-3"/>
      </button>
      <div v-else class="w-5"></div>

      <!-- 图标 -->
      <FolderIcon v-if="item.children && item.children.length > 0" class="w-4 h-4 text-amber-500"/>
      <DocumentIcon v-else class="w-4 h-4 text-blue-500"/>

      <!-- 编辑模式 -->
      <div v-if="isEditing" class="flex-1 flex gap-2">
        <input
          ref="titleInput"
          v-model="editData.title"
          class="input input-xs flex-1"
          placeholder="标题"
          @keyup.enter="saveEdit"
          @keyup.escape="cancelEdit"
        />
        <input
          v-model="editData.slug"
          class="input input-xs w-32"
          placeholder="slug"
          @keyup.enter="saveEdit"
          @keyup.escape="cancelEdit"
        />
        <input
          v-model="editData.description"
          class="input input-xs flex-1"
          placeholder="描述（可选）"
          @keyup.enter="saveEdit"
          @keyup.escape="cancelEdit"
        />
        <div class="flex gap-1">
          <button class="btn btn-xs btn-success" @click="saveEdit">
            <CheckIcon class="w-3 h-3"/>
          </button>
          <button class="btn btn-xs btn-ghost" @click="cancelEdit">
            <XMarkIcon class="w-3 h-3"/>
          </button>
        </div>
      </div>

      <!-- 显示模式 -->
      <div v-else class="flex-1 flex items-center gap-2">
        <span class="font-medium">{{ item.title }}</span>
        <code class="text-xs bg-base-200 px-1 rounded">{{ item.slug }}</code>
        <span v-if="item.description" class="text-sm text-base-content/60">{{ item.description }}</span>
        
        <!-- 操作按钮 -->
        <div class="ml-auto flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <button class="btn btn-xs btn-ghost" @click="startEdit" title="编辑">
            <PencilIcon class="w-3 h-3"/>
          </button>
          <button class="btn btn-xs btn-ghost" @click="addChild" title="添加子项">
            <PlusIcon class="w-3 h-3"/>
          </button>
          <button class="btn btn-xs btn-ghost text-error" @click="deleteItem" title="删除">
            <TrashIcon class="w-3 h-3"/>
          </button>
        </div>
      </div>
    </div>

    <!-- 子项目 -->
    <div v-if="item.expanded && item.children && item.children.length > 0" class="ml-2">
      <draggable
        v-model="item.children"
        group="directory"
        item-key="id"
        class="space-y-1"
      >
        <template #item="{ element }">
          <DirectoryTreeItem
            :item="element"
            :level="level + 1"
            @update="(id, updates) => $emit('update', id, updates)"
            @delete="(id) => $emit('delete', id)"
            @add-child="(id) => $emit('add-child', id)"
          />
        </template>
      </draggable>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, nextTick } from 'vue'
import draggable from 'vuedraggable'
import {
  ChevronRightIcon,
  ChevronDownIcon,
  FolderIcon,
  DocumentIcon,
  PencilIcon,
  PlusIcon,
  TrashIcon,
  CheckIcon,
  XMarkIcon
} from '@heroicons/vue/24/outline'

const props = defineProps({
  item: {
    type: Object,
    required: true
  },
  level: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['update', 'delete', 'add-child'])

// 编辑状态
const isEditing = ref(false)
const titleInput = ref(null)
const editData = reactive({
  title: '',
  slug: '',
  description: ''
})

// 展开/收起
function toggleExpanded() {
  emit('update', props.item.id, { expanded: !props.item.expanded })
}

// 开始编辑
function startEdit() {
  editData.title = props.item.title
  editData.slug = props.item.slug
  editData.description = props.item.description
  isEditing.value = true
  
  nextTick(() => {
    titleInput.value?.focus()
  })
}

// 保存编辑
function saveEdit() {
  if (!editData.title.trim()) {
    return
  }
  
  // 如果slug为空，自动生成
  if (!editData.slug.trim()) {
    editData.slug = editData.title.toLowerCase()
      .replace(/[^a-z0-9\u4e00-\u9fa5]/g, '-')
      .replace(/-+/g, '-')
      .replace(/^-|-$/g, '')
  }
  
  emit('update', props.item.id, {
    title: editData.title.trim(),
    slug: editData.slug.trim(),
    description: editData.description.trim()
  })
  
  isEditing.value = false
}

// 取消编辑
function cancelEdit() {
  isEditing.value = false
}

// 添加子项
function addChild() {
  emit('add-child', props.item.id)
}

// 删除项目
function deleteItem() {
  if (confirm(`确定要删除"${props.item.title}"吗？`)) {
    emit('delete', props.item.id)
  }
}
</script>

<style scoped>
.directory-tree-item {
  @apply select-none;
}
</style>