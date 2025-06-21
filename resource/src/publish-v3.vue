<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { marked } from 'marked'

// 类型定义
interface ArticleData {
  id: number
  content: string
  title: string
  categoryId: number[]
  type: number
}

interface Category {
  id: number
  name: string
}

interface TypeOption {
  value: number
  name: string
}

interface CategoryConfig {
  maxSelection: number
  selectedCategories: Set<number>
}

// 响应式数据
const articleData = reactive<ArticleData>({
  id: 0,
  content: '',
  title: '',
  categoryId: [],
  type: 1
})

const categories = ref<Category[]>([])
const typeList = ref<TypeOption[]>([])
const isSubmitting = ref(false)
const showCategoryPopup = ref(false)
const categorySearchTerm = ref('')

const categoryConfig = reactive<CategoryConfig>({
  maxSelection: 3,
  selectedCategories: new Set()
})

// 计算属性
const previewTitle = computed(() => {
  return articleData.title.trim() || '文章标题预览'
})

// 优化 marked 配置 - 使用 marked.use() 替代已废弃的 setOptions()
marked.use({
  breaks: true,  // 支持换行符转换
  gfm: true,     // 启用 GitHub Flavored Markdown
  pedantic: false, // 不严格遵循原始 markdown.pl
  silent: false    // 不静默错误
})

const previewContent = computed(() => {
  if (!articleData.content.trim()) {
    return '<p class="text-base-content/60">在左侧编辑区域输入内容，预览将在这里显示...</p>'
  }
  
  try {
    // 直接返回 marked 解析的结果，不使用 DOMPurify
    return marked.parse(articleData.content)
  } catch (error) {
    console.error('Markdown解析错误:', error)
    return '<p class="text-error">Markdown解析出错，请检查语法</p>'
  }
})

const charCount = computed(() => {
  return articleData.content.length.toLocaleString()
})

const filteredCategories = computed(() => {
  if (!categorySearchTerm.value.trim()) {
    return categories.value
  }
  return categories.value.filter(category => 
    category.name.toLowerCase().includes(categorySearchTerm.value.toLowerCase())
  )
})

const selectedCategoriesDisplay = computed(() => {
  return Array.from(categoryConfig.selectedCategories)
    .map(id => categories.value.find(c => c.id === id))
    .filter(Boolean) as Category[]
})



// 消息提示相关
interface Message {
  id: number
  text: string
  type: 'info' | 'success' | 'error'
}

const messages = ref<Message[]>([])
let messageIdCounter = 0

// 方法
const showMessage = (message: string, type: 'info' | 'success' | 'error' = 'info') => {
  const messageId = ++messageIdCounter
  const newMessage: Message = {
    id: messageId,
    text: message,
    type
  }
  
  messages.value.push(newMessage)
  
  // 自动移除
  setTimeout(() => {
    removeMessage(messageId)
  }, 5000)
}

const removeMessage = (messageId: number) => {
  const index = messages.value.findIndex(msg => msg.id === messageId)
  if (index > -1) {
    messages.value.splice(index, 1)
  }
}

const getArticleEnum = async () => {
  try {
    const response = await fetch('/api/forum/get-articles-enum', {
      method: 'GET'
    })
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    
    const result = await response.json()
    
    if (result.code === 0) {
      // 填充类型选项
      if (result.result.type) {
        typeList.value = result.result.type
      }
      
      // 填充分类选项
      if (result.result.category) {
        categories.value = result.result.category.map((category: any) => ({
          id: category.value,
          name: category.name
        }))
      }
    } else {
      throw new Error(result.msg || '获取枚举数据失败')
    }
  } catch (error) {
    console.error('获取枚举数据失败:', error)
    throw error
  }
}

const getOriginData = async (articleId: string) => {
  try {
    const response = await fetch('/api/forum/get-articles-origin', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        id: parseInt(articleId)
      })
    })
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    
    const result = await response.json()
    
    if (result.code === 0 && result.result) {
      const data = result.result
      
      // 更新文章数据
      articleData.title = data.articleTitle || ''
      articleData.content = data.articleContent || ''
      articleData.type = data.type || 1
      articleData.categoryId = data.categoryId || []
      
      // 设置分类选择
      if (articleData.categoryId && articleData.categoryId.length > 0) {
        categoryConfig.selectedCategories.clear()
        articleData.categoryId.forEach(id => {
          categoryConfig.selectedCategories.add(id)
        })
      }
    } else {
      throw new Error(result.msg || '获取文章数据失败')
    }
  } catch (error) {
    console.error('获取文章数据失败:', error)
    throw error
  }
}

const validateForm = (): boolean => {
  if (!articleData.title.trim()) {
    showMessage('请输入文章标题', 'error')
    return false
  }
  
  if (!articleData.content.trim()) {
    showMessage('请输入文章内容', 'error')
    return false
  }
  
  if (!articleData.type) {
    showMessage('请选择文章类型', 'error')
    return false
  }
  
  if (!articleData.categoryId.length) {
    showMessage('请选择文章分类', 'error')
    return false
  }
  
  return true
}

const submitArticle = async () => {
  if (isSubmitting.value) return
  
  if (!validateForm()) return
  
  isSubmitting.value = true
  
  try {
    const response = await fetch('/api/forum/write-articles', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(articleData)
    })
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    
    const result = await response.json()
    
    if (result.code === 0) {
      showMessage(result.result ? '文章更新成功！' : '文章发布成功！', 'success')
      
      // 延迟跳转到文章列表或详情页
      setTimeout(() => {
        window.location.href = '/post/' + result.result
      }, 300)
    } else {
      throw new Error(result.msg || '提交失败')
    }
  } catch (error) {
    console.error('提交文章失败:', error)
    showMessage((error as Error).message || '提交失败，请重试', 'error')
  } finally {
    isSubmitting.value = false
  }
}

const toggleCategoryPopup = () => {
  showCategoryPopup.value = !showCategoryPopup.value
  if (showCategoryPopup.value) {
    nextTick(() => {
      categorySearchTerm.value = ''
    })
  }
}

const selectCategory = (categoryId: number) => {
  const category = categories.value.find(c => c.id === categoryId)
  if (!category) return
  
  if (categoryConfig.selectedCategories.has(categoryId)) {
    categoryConfig.selectedCategories.delete(categoryId)
  } else {
    if (categoryConfig.selectedCategories.size >= categoryConfig.maxSelection) {
      showMessage(`最多只能选择${categoryConfig.maxSelection}个分类`, 'error')
      return
    }
    categoryConfig.selectedCategories.add(categoryId)
  }
  
  articleData.categoryId = Array.from(categoryConfig.selectedCategories)
}

const removeCategory = (categoryId: number) => {
  categoryConfig.selectedCategories.delete(categoryId)
  articleData.categoryId = Array.from(categoryConfig.selectedCategories)
}

const clearContent = () => {
  if (confirm('确定要清空所有内容吗？')) {
    articleData.title = ''
    articleData.content = ''
    articleData.type = 1
    articleData.categoryId = []
    categoryConfig.selectedCategories.clear()
  }
}

const initData = async () => {
  try {
    // 获取分类和类型选项
    await getArticleEnum()
    
    // 检查是否为编辑模式
    const urlParams = new URLSearchParams(window.location.search)
    const articleId = urlParams.get('id')
    
    if (articleId) {
      articleData.id = parseInt(articleId)
      await getOriginData(articleId)
    }
  } catch (error) {
    console.error('初始化数据失败:', error)
    showMessage('初始化失败，请刷新页面重试', 'error')
  }
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

// 生命周期
onMounted(() => {
  initData()
  
  // 添加全局事件监听器
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleKeyDown)
})

// 组件卸载时清理事件监听器
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleKeyDown)
})
</script>

<template>
  <div class="min-h-screen flex flex-col bg-base-200">
    <!-- 消息提示组件 -->
    <div class="fixed top-4 right-4 z-50 space-y-2">
      <div 
        v-for="message in messages" 
        :key="message.id"
        :class="[
          'alert w-auto max-w-sm transition-all duration-300',
          {
            'alert-info': message.type === 'info',
            'alert-success': message.type === 'success',
            'alert-error': message.type === 'error'
          }
        ]"
      >
        <span>{{ message.text }}</span>
        <button 
          @click="removeMessage(message.id)"
          class="btn btn-sm btn-ghost"
        >
          ×
        </button>
      </div>
    </div>
    <main class="flex-1 container mx-auto px-4 py-4">
      <div class="tabs tabs-lift">
        <input type="radio" name="my_tabs_3" class="tab" aria-label="文章编写" checked="checked"/>
        <div class="tab-content bg-base-100 border-base-300 p-0">
          <div class="flex flex-col h-full">
            <!-- 编辑区域 -->
            <div class="flex-1 p-6 space-y-6">
              <!-- 文章标题区域 -->
              <div class="form-control">
                <label class="label pb-1">
                  <span class="label-text font-medium text-base-content">📝 文章标题</span>
                  <span class="label-text-alt text-base-content/60">必填</span>
                </label>
                <input 
                  type="text" 
                  v-model="articleData.title"
                  placeholder="请输入一个吸引人的标题..."
                  class="input input-bordered input-md w-full focus:input-primary"
                />
              </div>
              
              <!-- 分类和类型选择区域 -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- 文章类型 -->
                <div class="form-control">
                  <label class="label pb-1">
                    <span class="label-text font-medium text-base-content">🏷️ 文章类型</span>
                    <span class="label-text-alt text-base-content/60">必选</span>
                  </label>
                  <select 
                    v-model="articleData.type" 
                    class="select select-bordered w-full focus:select-primary"
                  >
                    <option value="">请选择类型</option>
                    <option 
                      v-for="type in typeList" 
                      :key="type.value" 
                      :value="type.value"
                    >
                      {{ type.name }}
                    </option>
                  </select>
                </div>
                
                <!-- 文章分类 -->
                <div class="form-control">
                  <label class="label pb-1">
                    <span class="label-text font-medium text-base-content">📂 文章分类</span>
                    <span class="label-text-alt text-base-content/60">最多选择3个</span>
                  </label>
                  
                  <!-- 分类选择器容器 -->
                  <div class="category-selector relative">
                    <!-- 已选分类标签展示区 -->
                    <div 
                      @click="toggleCategoryPopup"
                      class="selected-tags mb-2 min-h-8 flex flex-wrap gap-2 p-2 border border-base-300 rounded-lg bg-base-100 cursor-pointer hover:border-primary transition-colors"
                    >
                      <span 
                        v-if="selectedCategoriesDisplay.length === 0" 
                        class="text-base-content/60 text-sm"
                      >
                        点击此处选择分类...
                      </span>
                      <span 
                        v-for="category in selectedCategoriesDisplay" 
                        :key="category.id"
                        class="category-tag inline-flex items-center gap-1 px-2 py-1 bg-primary text-primary-content text-sm rounded-full"
                      >
                        <span>{{ category.name }}</span>
                        <button 
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
                      v-show="showCategoryPopup"
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
              </div>
              
              <!-- 文章内容区域 -->
              <div class="form-control flex-1">
                <label class="label pb-2">
                  <span class="label-text font-medium text-base-content">✍️ 文章内容</span>
                  <span class="label-text-alt text-base-content/60">支持 Markdown 语法</span>
                </label>
                <div class="relative flex-1">
                  <textarea 
                    v-model="articleData.content"
                    class="textarea textarea-bordered w-full h-full min-h-96 resize-none focus:textarea-primary font-mono text-sm leading-relaxed"
                    placeholder="开启你的创作..."
                  ></textarea>
                  <!-- 字数统计 -->
                  <div class="absolute bottom-2 right-4 text-xs text-base-content/50 bg-base-100 px-2 py-1 rounded">
                    <span>{{ charCount }}</span> 字符
                  </div>
                </div>
              </div>

              <!-- 底部操作区域 -->
              <div class="card bg-base-50 border border-base-300">
                <div class="card-body p-4">
                  <div class="flex items-center justify-between">
                    <div class="flex items-center gap-4">
                      <div class="form-control">
                        <label class="label cursor-pointer gap-2">
                          <input type="checkbox" class="checkbox checkbox-sm" disabled/>
                          <span class="label-text text-sm">保存为草稿</span>
                        </label>
                      </div>
                      <div class="form-control">
                        <label class="label cursor-pointer gap-2">
                          <input type="checkbox" class="checkbox checkbox-sm" checked disabled/>
                          <span class="label-text text-sm">允许评论</span>
                        </label>
                      </div>
                    </div>
                    <div class="flex items-center gap-2">
                      <button 
                        @click="clearContent"
                        class="btn btn-ghost btn-sm"
                      >
                        清空内容
                      </button>
                      <button 
                        @click="submitArticle"
                        :disabled="isSubmitting"
                        class="btn btn-primary btn-sm"
                      >
                        {{ isSubmitting ? '发布中...' : '发布文章' }}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <input type="radio" name="my_tabs_3" class="tab" aria-label="预览"/>
        <div class="tab-content bg-base-100 border-base-300 p-6">
          <div class="mb-4">
            <h1 class="text-2xl font-normal text-base-content mb-4">标题：{{ previewTitle }}</h1>
          </div>
          <div 
            class="prose lg:prose-base md:prose-lg prose-h1:font-normal prose-h2:font-normal prose-h3:font-normal prose-pre:bg-base-200 prose-code:text-base-content max-w-none text-base-content overflow-hidden min-w-0"
            v-html="previewContent"
          ></div>
        </div>
      </div>
    </main>
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