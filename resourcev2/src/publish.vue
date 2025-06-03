<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { getArticleEnum, getArticlesOrigin, submitArticle } from './utils/articleService.js'

// 表单数据 - 匹配后端接口结构
const articleData = ref({
  id: 0,
  articleContent: '',
  articleTitle: '',
  categoryId: [],
  type: 1
})

// 动态选项数据
const categories = ref([
  {label: '分享', value: 1},
  {label: '求助', value: 2},
])

const typeList = ref([
  {label: 'GooseForum', value: 1},
])

// 编辑器配置
const toolbars = [
  'bold', 'underline', 'italic', 'strikeThrough', '-',
  'title', 'sub', 'sup', 'quote', 'unorderedList', 'orderedList', 'task', '-',
  'codeRow', 'code', 'link', 'image', 'table', 'mermaid', 'katex', '-',
  'revoke', 'next', 'save', '=', 'pageFullscreen', 'fullscreen', 'preview', 'htmlPreview', 'catalog'
]

// 编辑器主题
const editorTheme = computed(() => {
  // 这里可以根据系统主题动态切换
  return 'light'
})

// 状态管理
const isSubmitting = ref(false)
const isCategoryDropdownOpen = ref(false)

// 获取URL参数（用于编辑模式）
const getUrlParams = () => {
  const urlParams = new URLSearchParams(window.location.search)
  return urlParams.get('id')
}

// 表单验证
const validateForm = () => {
  if (!articleData.value.articleTitle.trim()) {
    alert('请输入文章标题')
    return false
  }
  if (!articleData.value.categoryId || articleData.value.categoryId.length === 0) {
    alert('请选择文章分类')
    return false
  }
  if (!articleData.value.articleContent.trim()) {
    alert('请输入文章内容')
    return false
  }
  return true
}

// 提交表单
const handleSubmit = async () => {
  if (!validateForm()) return
  if (isSubmitting.value) return

  isSubmitting.value = true
  try {
    const response = await submitArticle(articleData.value)
    if (response.code !== 0) {
      alert(response.message || '发布失败，请重试')
      return
    }

    alert('文章发布成功！')
    // 跳转到新发布的文章地址
    window.location.href = `/post/${response.result}`
  } catch (error) {
    console.error('发布失败:', error)
    alert('发布失败，请重试')
  } finally {
    isSubmitting.value = false
  }
}

// 保存草稿
const saveDraft = async () => {
  try {
    // 这里可以添加保存草稿的逻辑
    console.log('保存草稿:', articleData.value)
    alert('草稿保存成功！')
  } catch (error) {
    console.error('保存草稿失败:', error)
    alert('保存草稿失败，请重试')
  }
}

// 获取原始文章数据（编辑模式）
const getOriginData = async () => {
  const id = getUrlParams()
  if (!id) return

  try {
    const res = await getArticlesOrigin(id)
    if (res.code === 0 && res.result) {
      articleData.value.articleTitle = res.result.articleTitle
      articleData.value.articleContent = res.result.articleContent
      articleData.value.categoryId = res.result.categoryId
      articleData.value.type = res.result.type
      articleData.value.id = parseInt(id)
    }
  } catch (err) {
    console.error('获取文章数据失败:', err)
    alert('获取文章数据失败')
  }
}

// 切换分类选择
const toggleCategory = (categoryId) => {
  const index = articleData.value.categoryId.indexOf(categoryId)
  if (index > -1) {
    articleData.value.categoryId.splice(index, 1)
  } else {
    articleData.value.categoryId.push(categoryId)
  }
}

// 点击外部关闭下拉框
const handleClickOutside = (event) => {
  if (!event.target.closest('.relative')) {
    isCategoryDropdownOpen.value = false
  }
}

// 页面初始化
onMounted(async () => {
  try {
    // 获取分类和类型选项
    const enumInfo = await getArticleEnum()
    if (enumInfo.code === 0) {
      categories.value = enumInfo.result.category.map((item) => ({
        label: item.name,
        value: item.value
      }))
      typeList.value = enumInfo.result.type.map((item) => ({
        label: item.name,
        value: item.value
      }))
    }

    // 如果有 id 参数，说明是编辑模式
    if (getUrlParams()) {
      await getOriginData()
    }
    
    // 添加点击外部事件监听
    document.addEventListener('click', handleClickOutside)
  } catch (error) {
    console.error('初始化失败:', error)
  }
})

// 组件卸载时移除事件监听
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})


</script>
<template>

  <div class="max-w-6xl mx-auto py-8 px-4">
    <!-- 页面标题 -->
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-base-content mb-2">发布文章</h1>
      <p class="text-base-content/70">分享您的技术见解和经验</p>
    </div>

    <!-- 发布表单 -->
    <form @submit.prevent="handleSubmit" class="space-y-6">
      <!-- 基本信息网格布局 -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- 文章标题 -->
        <div class="lg:col-span-2">
          <label class="form-control w-full">
            <div class="label">
              <span class="label-text font-medium">文章标题</span>
              <span class="label-text-alt text-error">*</span>
            </div>
            <input v-model="articleData.articleTitle" type="text" placeholder="请输入文章标题" class="input input-bordered w-full"
                   required />
          </label>
        </div>

        <!-- 文章类型 -->
        <div>
          <label class="form-control w-full">
            <div class="label">
              <span class="label-text font-medium">文章类型</span>
              <span class="label-text-alt text-error">*</span>
            </div>
            <select v-model="articleData.type" class="select select-bordered w-full" required>
              <option v-for="type in typeList" :key="type.value" :value="type.value">
                {{ type.label }}
              </option>
            </select>
          </label>
        </div>

        <!-- 分类选择 -->
        <div>
          <label class="form-control w-full">
            <div class="label">
              <span class="label-text font-medium">文章分类</span>
              <span class="label-text-alt text-error">*</span>
            </div>
            
            <!-- 自定义多选组件 -->
            <div class="relative">
              <!-- 选择框 -->
              <div 
                @click="isCategoryDropdownOpen = !isCategoryDropdownOpen"
                class="min-h-12 px-3 py-2 border border-base-300 rounded-lg bg-base-100 cursor-pointer hover:border-base-400 focus-within:border-primary transition-colors"
              >
                <!-- 已选择的标签 -->
                <div v-if="articleData.categoryId.length > 0" class="flex flex-wrap gap-1">
                  <span 
                    v-for="categoryId in articleData.categoryId" 
                    :key="categoryId"
                    class="inline-flex items-center gap-1 px-2 py-1 bg-primary/10 text-primary text-sm rounded-md"
                  >
                    {{ categories.find(c => c.value === categoryId)?.label }}
                    <button 
                      @click.stop="articleData.categoryId = articleData.categoryId.filter(id => id !== categoryId)"
                      class="hover:bg-primary/20 rounded-full p-0.5 transition-colors"
                    >
                      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                      </svg>
                    </button>
                  </span>
                </div>
                
                <!-- 占位符 -->
                <div v-else class="text-base-content/50">
                  请选择文章分类
                </div>
                
                <!-- 下拉箭头 -->
                <div class="absolute right-3 top-1/2 transform -translate-y-1/2 pointer-events-none">
                  <svg 
                    class="w-4 h-4 transition-transform duration-200" 
                    :class="{ 'rotate-180': isCategoryDropdownOpen }"
                    fill="none" 
                    stroke="currentColor" 
                    viewBox="0 0 24 24"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
                  </svg>
                </div>
              </div>
              
              <!-- 下拉选项 -->
              <div 
                v-show="isCategoryDropdownOpen"
                class="absolute z-10 w-full mt-1 bg-base-100 border border-base-300 rounded-lg shadow-lg max-h-60 overflow-y-auto"
              >
                <div 
                  v-for="category in categories" 
                  :key="category.value"
                  @click="toggleCategory(category.value)"
                  class="flex items-center gap-3 px-3 py-2 hover:bg-base-200 cursor-pointer transition-colors"
                >
                  <!-- 复选框 -->
                  <div class="flex items-center">
                    <input 
                      type="checkbox" 
                      :checked="articleData.categoryId.includes(category.value)"
                      class="checkbox checkbox-primary checkbox-sm"
                      readonly
                    >
                  </div>
                  
                  <!-- 分类名称 -->
                  <span class="flex-1">{{ category.label }}</span>
                </div>
              </div>
            </div>
            
            <div class="label">
              <span class="label-text-alt">可以选择多个分类</span>
            </div>
          </label>
        </div>
      </div>

      <!-- Markdown 编辑器 -->


      <div class="border border-base-300 bg-base-100 rounded-lg overflow-hidden">
        <MdEditor
            v-model="articleData.articleContent"
            :height="500"
            :preview="true"
            :toolbars="toolbars"
            :theme="editorTheme"
            :previewTheme="'github'"
            :uploadImg="false"
            placeholder="请输入文章内容，支持 Markdown 语法"
        />
      </div>




      <!-- 操作按钮 -->
      <div class="flex flex-col sm:flex-row gap-4 pt-6">
        <button type="submit" class="btn btn-primary flex-1 sm:flex-none sm:min-w-32" :disabled="isSubmitting">
          <span v-if="isSubmitting" class="loading loading-spinner loading-sm"></span>
          {{ isSubmitting ? '发布中...' : '发布文章' }}
        </button>

        <button type="button" @click="saveDraft" class="btn btn-outline flex-1 sm:flex-none sm:min-w-32"
                :disabled="isSubmitting">
          保存草稿
        </button>

      </div>
    </form>


  </div>
</template>


<style >

/* 修复 md-editor-v3 分割线不可见的问题 */
.md-editor-resize-operate {
  background-color: #e5e7eb !important;
  border-left: 1px solid #d1d5db !important;
  border-right: 1px solid #d1d5db !important;
  width: 6px !important;
}

.md-editor-resize-operate:hover {
  background-color: #3b82f6 !important;
}

/* 确保在暗色主题下也能看到分割线 */
[data-theme="dark"] .md-editor-resize-operate {
  background-color: #4b5563 !important;
  border-left: 1px solid #6b7280 !important;
  border-right: 1px solid #6b7280 !important;
}

[data-theme="dark"] .md-editor-resize-operate:hover {
  background-color: #60a5fa !important;
}

/* 列表样式 */
.md-editor-preview ul,
.md-editor-preview ol {
  padding-left: 1.5em;
  margin: 0.5em 0;
}

.md-editor-preview ul {
  list-style-type: disc;
}

.md-editor-preview ol {
  list-style-type: decimal;
}

.md-editor-preview li {
  margin: 0.25em 0;
}

.md-editor-preview ul ul { list-style-type: circle; }
.md-editor-preview ul ul ul { list-style-type: square; }
.md-editor-preview ol ol { list-style-type: lower-alpha; }
.md-editor-preview ol ol ol { list-style-type: lower-roman; }

</style>