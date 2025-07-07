<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'

// 类型定义
interface Category {
  id: number
  name: string
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
  placeholder: '',
  label: '文章分类',
  labelAlt: '',
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
const isOpen = ref(false)
const isFocused = ref(false)
const searchTerm = ref('')
const selectorRef = ref<HTMLElement>()

// 计算属性
const selectedCategories = computed(() => {
  return props.modelValue
    .map(id => props.categories.find(c => c.id === id))
    .filter(Boolean) as Category[]
})

const filteredCategories = computed(() => {
  if (!searchTerm.value.trim()) return props.categories
  return props.categories.filter(category => 
    category.name.toLowerCase().includes(searchTerm.value.toLowerCase())
  )
})

const hasSelection = computed(() => selectedCategories.value.length > 0)
const isLabelFloating = computed(() => hasSelection.value || isFocused.value || isOpen.value)

// 方法
const toggle = () => {
  if (props.disabled) return
  isOpen.value = !isOpen.value
  isFocused.value = isOpen.value
  if (isOpen.value) searchTerm.value = ''
}

const selectCategory = (categoryId: number) => {
  if (props.disabled) return
  
  const newSelection = props.modelValue.includes(categoryId)
    ? props.modelValue.filter(id => id !== categoryId)
    : props.modelValue.length < props.maxSelection
      ? [...props.modelValue, categoryId]
      : (emit('error', `最多只能选择${props.maxSelection}个分类`), props.modelValue)
  
  if (newSelection !== props.modelValue) {
    emit('update:modelValue', newSelection)
    emit('change', newSelection)
  }
}

const removeCategory = (categoryId: number, event: Event) => {
  event.stopPropagation()
  if (props.disabled) return
  
  const newSelection = props.modelValue.filter(id => id !== categoryId)
  emit('update:modelValue', newSelection)
  emit('change', newSelection)
}

// 点击外部关闭
const handleClickOutside = (event: Event) => {
  if (!selectorRef.value?.contains(event.target as Node)) {
    isOpen.value = false
    isFocused.value = false
  }
}

// ESC键关闭
const handleKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    isOpen.value = false
    isFocused.value = false
  }
}

// 生命周期
onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleKeyDown)
})
</script>

<template>
  <div ref="selectorRef" class="relative">
    <!-- Floating Label Container -->
    <div 
      @click="toggle"
      :class="[
        'relative h-[2.5rem] border rounded-lg bg-base-100 cursor-pointer transition-all duration-200 flex items-center',
        {
          'border-primary': isOpen || isFocused,
          'border-base-content/20': !isOpen && !isFocused && !disabled,
          'border-base-content/10 cursor-not-allowed opacity-60': disabled
        }
      ]"
    >
      <!-- Floating Label -->
      <label 
        :class="[
          'absolute left-3 transition-all duration-200 pointer-events-none select-none',
          {
            'text-xs -top-2 bg-base-100 px-1': isLabelFloating,
            'text-sm top-1/2 -translate-y-1/2 text-base-content/60': !isLabelFloating,
            'text-primary': isLabelFloating && (isOpen || isFocused),
            'text-base-content/80': isLabelFloating && !isOpen && !isFocused
          }
        ]"
      >
        {{ label }}
      </label>
      
      <!-- Selected Categories Display -->
      <div class="flex flex-wrap gap-1.5 px-3 py-1 flex-1 items-center overflow-hidden">
        <span 
          v-if="!hasSelection" 
          class="text-base-content/40 text-sm"
        >
          {{ placeholder }}
        </span>
        
        <span 
          v-for="category in selectedCategories" 
          :key="category.id"
          class="inline-flex items-center gap-1 px-2 py-0.5 bg-primary/10 text-primary text-xs rounded-full border border-primary/20 flex-shrink-0"
        >
          {{ category.name }}
          <button 
            v-if="!disabled"
            @click="removeCategory(category.id, $event)"
            class="w-3 h-3 rounded-full hover:bg-primary/20 flex items-center justify-center transition-colors"
          >
            ×
          </button>
        </span>
      </div>
      
      <!-- Dropdown Arrow -->
      <div class="absolute right-3 top-1/2 -translate-y-1/2">
        <svg 
          :class="['w-4 h-4 transition-transform duration-200', { 'rotate-180': isOpen }]"
          fill="none" 
          stroke="currentColor" 
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </div>
    </div>
    
    <!-- Label Alt -->
    <div v-if="labelAlt" class="text-xs text-base-content/60 mt-1 ml-1">
      {{ labelAlt }}
    </div>
    
    <!-- Dropdown Panel -->
    <div 
      v-show="isOpen && !disabled"
      class="absolute top-full left-0 right-0 mt-1 bg-base-100 border border-base-300 rounded-lg shadow-xl z-50 overflow-hidden"
    >
      <!-- Search Input -->
      <div class="p-3 border-b border-base-300">
        <input 
          v-model="searchTerm"
          type="text" 
          placeholder="搜索分类..." 
          class="w-full px-3 py-2 text-sm border border-base-300 rounded focus:border-primary focus:outline-none"
          autocomplete="off"
        >
      </div>
      
      <!-- Category Options -->
      <div class="max-h-48 overflow-y-auto">
        <div 
          v-for="category in filteredCategories" 
          :key="category.id"
          @click="selectCategory(category.id)"
          :class="[
            'flex items-center justify-between px-3 py-2 cursor-pointer transition-colors',
            {
              'bg-primary text-primary-content': modelValue.includes(category.id),
              'hover:bg-base-200': !modelValue.includes(category.id)
            }
          ]"
        >
          <span class="text-sm">{{ category.name }}</span>
          <svg 
            v-if="modelValue.includes(category.id)"
            class="w-4 h-4" 
            fill="currentColor" 
            viewBox="0 0 20 20"
          >
            <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
          </svg>
        </div>
        
        <div 
          v-if="filteredCategories.length === 0 && searchTerm.trim()"
          class="text-center text-base-content/60 py-6 text-sm"
        >
          未找到匹配的分类
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>