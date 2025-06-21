<script setup lang="ts">
import { ref, reactive, computed, nextTick, onMounted, onUnmounted, watch } from 'vue'

// 类型定义
interface Category {
  id: number
  name: string
}

interface CategoryConfig {
  maxSelection: number
  selectedCategories: Set<number>
}

// Props 定义
interface Props {
  categories: Category[]
  modelValue: number[]
  maxSelection?: number
  placeholder?: string
  label?: string
  labelAlt?: string
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  maxSelection: 3,
  placeholder: '点击此处选择分类...',
  label: '📂 文章分类',
  labelAlt: '最多选择3个',
  disabled: false
})

// Emits 定义
interface Emits {
  'update:modelValue': [value: number[]]
  'change': [value: number[]]
  'error': [message: string]
}

const emit = defineEmits<Emits>()

// 响应式数据
const showCategoryPopup = ref(false)
const categorySearchTerm = ref('')

const categoryConfig = reactive<CategoryConfig>({
  maxSelection: props.maxSelection,
  selectedCategories: new Set(props.modelValue)
})

// 计算属性
const filteredCategories = computed(() => {
  if (!categorySearchTerm.value.trim()) {
    return props.categories
  }
  return props.categories.filter(category => 
    category.name.toLowerCase().includes(categorySearchTerm.value.toLowerCase())
  )
})

const selectedCategoriesDisplay = computed(() => {
  return Array.from(categoryConfig.selectedCategories)
    .map(id => props.categories.find(c => c.id === id))
    .filter(Boolean) as Category[]
})

// 方法
const toggleCategoryPopup = () => {
  if (props.disabled) return
  
  showCategoryPopup.value = !showCategoryPopup.value
  if (showCategoryPopup.value) {
    nextTick(() => {
      categorySearchTerm.value = ''
    })
  }
}

const selectCategory = (categoryId: number) => {
  if (props.disabled) return
  
  const category = props.categories.find(c => c.id === categoryId)
  if (!category) return
  
  if (categoryConfig.selectedCategories.has(categoryId)) {
    categoryConfig.selectedCategories.delete(categoryId)
  } else {
    if (categoryConfig.selectedCategories.size >= categoryConfig.maxSelection) {
      emit('error', `最多只能选择${categoryConfig.maxSelection}个分类`)
      return
    }
    categoryConfig.selectedCategories.add(categoryId)
  }
  
  const newValue = Array.from(categoryConfig.selectedCategories)
  emit('update:modelValue', newValue)
  emit('change', newValue)
}

const removeCategory = (categoryId: number) => {
  if (props.disabled) return
  
  categoryConfig.selectedCategories.delete(categoryId)
  const newValue = Array.from(categoryConfig.selectedCategories)
  emit('update:modelValue', newValue)
  emit('change', newValue)
}

// 点击外部关闭分类弹窗
const handleClickOutside = (event: Event) => {
  const target = event.target as Element
  const categorySelector = target.closest('.category-selector')
  const categoryOption = target.closest('.category-option')
  
  // 如果点击的是分类选项，不关闭弹窗
  if (categoryOption) {
    return
  }
  
  // 如果点击在选择器外部，关闭弹窗
  if (!categorySelector && showCategoryPopup.value) {
    showCategoryPopup.value = false
  }
}

// ESC键关闭分类弹窗
const handleKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && showCategoryPopup.value) {
    showCategoryPopup.value = false
  }
}

// 监听 props 变化
const updateSelectedCategories = () => {
  categoryConfig.selectedCategories.clear()
  props.modelValue.forEach(id => {
    categoryConfig.selectedCategories.add(id)
  })
  categoryConfig.maxSelection = props.maxSelection
}

// 生命周期
onMounted(() => {
  updateSelectedCategories()
  
  // 添加全局事件监听器
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleKeyDown)
})

// 组件卸载时清理事件监听器
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleKeyDown)
})

// 监听 props 变化
watch(() => props.modelValue, updateSelectedCategories, { deep: true })
watch(() => props.maxSelection, (newVal) => {
  categoryConfig.maxSelection = newVal
})
</script>

<template>
  <div class="form-control">
    <label class="label pb-1" v-if="label">
      <span class="label-text font-medium text-base-content">{{ label }}</span>
      <span class="label-text-alt text-base-content/60" v-if="labelAlt">{{ labelAlt }}</span>
    </label>
    
    <!-- 分类选择器容器 -->
    <div class="category-selector relative">
      <!-- 已选分类标签展示区 -->
      <div 
        @click="toggleCategoryPopup"
        :class="[
          'selected-tags mb-2 min-h-8 max-h-20 flex flex-wrap gap-2 p-2 border border-base-300 rounded-lg bg-base-100 transition-colors overflow-y-auto',
          {
            'cursor-pointer hover:border-primary': !disabled,
            'cursor-not-allowed opacity-60': disabled
          }
        ]"
      >
        <span 
          v-if="selectedCategoriesDisplay.length === 0" 
          class="text-base-content/60 text-sm"
        >
          {{ placeholder }}
        </span>
        <span 
          v-for="category in selectedCategoriesDisplay" 
          :key="category.id"
          class="category-tag inline-flex items-center gap-1 px-2 py-0.5 bg-primary text-primary-content text-sm rounded-full"
        >
          <span>{{ category.name }}</span>
          <button 
            v-if="!disabled"
            type="button" 
            @click.stop="removeCategory(category.id)"
            class="remove-tag w-4 h-4 rounded-full hover:bg-white/20 flex items-center justify-center transition-colors"
          >
            ×
          </button>
        </span>
      </div>
      
      <!-- 分类选择浮层 -->
      <div 
        v-show="showCategoryPopup && !disabled"
        class="absolute top-full left-0 right-0 mt-1 bg-base-100 border border-base-300 rounded-lg shadow-xl z-50"
      >
        <!-- 搜索框 -->
        <div class="p-3 border-b border-base-300">
          <div class="relative">
            <input 
              type="text" 
              v-model="categorySearchTerm"
              placeholder="搜索分类..." 
              class="input input-bordered w-full focus:input-primary"
              autocomplete="off"
            >
            <div class="absolute inset-y-0 right-0 flex items-center pr-3">
              <svg class="w-4 h-4 text-base-content/60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z"></path>
              </svg>
            </div>
          </div>
        </div>
        
        <!-- 分类选项 -->
        <div class="max-h-60 overflow-y-auto">
          <div class="p-2">
            <div class="space-y-1">
              <div 
                v-for="category in filteredCategories" 
                :key="category.id"
                @click="selectCategory(category.id)"
                class="category-option p-2 cursor-pointer rounded transition-colors"
                :class="{
                  'bg-primary text-primary-content': categoryConfig.selectedCategories.has(category.id),
                  'text-base-content hover:bg-base-200 hover:text-base-content': !categoryConfig.selectedCategories.has(category.id)
                }"
              >
                <div class="flex items-center justify-between">
                  <span>{{ category.name }}</span>
                  <svg 
                    v-if="categoryConfig.selectedCategories.has(category.id)"
                    class="w-4 h-4" 
                    fill="currentColor" 
                    viewBox="0 0 20 20"
                  >
                    <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"></path>
                  </svg>
                </div>
              </div>
            </div>
            <div 
              v-if="filteredCategories.length === 0 && categorySearchTerm.trim()"
              class="text-center text-base-content/60 py-4"
            >
              未找到匹配的分类
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 分类标签动画 */
.category-tag {
  animation: fadeIn 0.2s ease-in;
}

@keyframes fadeIn {
  from { opacity: 0; transform: scale(0.9); }
  to { opacity: 1; transform: scale(1); }
}
</style>