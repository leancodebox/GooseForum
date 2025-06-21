# CategorySelector 组件

一个可复用的Vue 3分类多选组件，支持搜索、标签展示、点击外部关闭等功能。

## 功能特性

- ✅ 多选分类支持
- ✅ 实时搜索过滤
- ✅ 标签式展示已选分类
- ✅ 点击外部自动关闭
- ✅ ESC键关闭
- ✅ 最大选择数量限制
- ✅ 自定义样式和文本
- ✅ 禁用状态支持
- ✅ TypeScript支持

## 基础用法

```vue
<template>
  <CategorySelector
    v-model="selectedCategories"
    :categories="categories"
    :max-selection="3"
    @change="handleCategoryChange"
    @error="handleCategoryError"
  />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import CategorySelector from './components/CategorySelector.vue'

const categories = ref([
  { id: 1, name: '技术分享' },
  { id: 2, name: '生活随笔' },
  { id: 3, name: '学习笔记' }
])

const selectedCategories = ref<number[]>([])

const handleCategoryChange = (newCategories: number[]) => {
  console.log('分类已更改:', newCategories)
}

const handleCategoryError = (message: string) => {
  alert(message)
}
</script>
```

## Props

| 属性名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `categories` | `Category[]` | `[]` | 分类列表数据 |
| `modelValue` | `number[]` | `[]` | 已选分类ID数组（v-model） |
| `maxSelection` | `number` | `3` | 最大选择数量 |
| `placeholder` | `string` | `'点击此处选择分类...'` | 占位符文本 |
| `label` | `string` | `'📂 文章分类'` | 标签文本 |
| `labelAlt` | `string` | `'最多选择3个'` | 副标签文本 |
| `disabled` | `boolean` | `false` | 是否禁用 |

## Events

| 事件名 | 参数 | 说明 |
|--------|------|------|
| `update:modelValue` | `number[]` | 选择的分类ID数组更新 |
| `change` | `number[]` | 分类选择变化 |
| `error` | `string` | 错误信息（如超出最大选择数量） |

## 类型定义

```typescript
interface Category {
  id: number
  name: string
}
```

## 自定义样式

组件使用了DaisyUI的样式类，你可以通过以下CSS变量来自定义样式：

```css
/* 自定义主题色 */
:root {
  --primary: your-primary-color;
  --primary-content: your-primary-content-color;
}
```

## 高级用法

### 自定义配置

```vue
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
```

### 禁用状态

```vue
<CategorySelector
  v-model="selectedCategories"
  :categories="categories"
  :disabled="true"
/>
```

## 注意事项

1. 确保传入的`categories`数组中每个分类都有唯一的`id`
2. `modelValue`应该是分类ID的数组，而不是分类对象
3. 组件会自动处理点击外部关闭和ESC键关闭功能
4. 当选择数量达到最大限制时，会触发`error`事件

## 依赖

- Vue 3.x
- DaisyUI（用于样式）
- TypeScript（可选，但推荐）