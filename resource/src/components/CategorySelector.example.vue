<script setup lang="ts">
import { ref } from 'vue'
import CategorySelector from './CategorySelector.vue'

// 示例数据
const categories = ref([
  { id: 1, name: '技术分享' },
  { id: 2, name: '生活随笔' },
  { id: 3, name: '学习笔记' },
  { id: 4, name: '项目经验' },
  { id: 5, name: '工具推荐' },
  { id: 6, name: '前端开发' },
  { id: 7, name: '后端开发' },
  { id: 8, name: '移动开发' },
  { id: 9, name: '数据库' },
  { id: 10, name: '运维部署' }
])

const selectedCategories = ref<number[]>([1, 3])

// 事件处理
const handleCategoryChange = (newCategories: number[]) => {
  console.log('分类已更改:', newCategories)
}

const handleCategoryError = (message: string) => {
  alert(message)
}
</script>

<template>
  <div class="p-8 max-w-2xl mx-auto">
    <h1 class="text-2xl font-bold mb-6">CategorySelector 组件示例</h1>
    
    <!-- 基础用法 -->
    <div class="mb-8">
      <h2 class="text-lg font-semibold mb-4">基础用法</h2>
      <CategorySelector
        v-model="selectedCategories"
        :categories="categories"
        :max-selection="3"
        @change="handleCategoryChange"
        @error="handleCategoryError"
      />
    </div>
    
    <!-- 自定义配置 -->
    <div class="mb-8">
      <h2 class="text-lg font-semibold mb-4">自定义配置</h2>
      <CategorySelector
        v-model="selectedCategories"
        :categories="categories"
        :max-selection="5"
        placeholder="请选择您感兴趣的分类..."
        label="🏷️ 兴趣分类"
        label-alt="最多选择5个"
        @change="handleCategoryChange"
        @error="handleCategoryError"
      />
    </div>
    
    <!-- 禁用状态 -->
    <div class="mb-8">
      <h2 class="text-lg font-semibold mb-4">禁用状态</h2>
      <CategorySelector
        v-model="selectedCategories"
        :categories="categories"
        :max-selection="3"
        :disabled="true"
        @change="handleCategoryChange"
        @error="handleCategoryError"
      />
    </div>
    
    <!-- 当前选择的分类 -->
    <div class="mt-8 p-4 bg-base-200 rounded-lg">
      <h3 class="font-semibold mb-2">当前选择的分类 ID:</h3>
      <code class="text-sm">{{ selectedCategories }}</code>
    </div>
  </div>
</template>

<style scoped>
/* 示例页面样式 */
</style>