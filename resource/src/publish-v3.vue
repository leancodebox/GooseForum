<script setup lang="ts">
import {computed, onMounted, reactive, ref} from 'vue'
import MarkdownIt from 'markdown-it'
import markdownItTaskLists from 'markdown-it-task-lists'
import mermaid from 'mermaid'
import CategorySelector from './components/CategorySelector.vue'

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


// 计算属性
const previewTitle = computed(() => {
  return articleData.title.trim() || '文章标题预览'
})

// 配置 markdown-it 以保持与服务端 Goldmark 的高度一致性
const md = new MarkdownIt({
  html: true,        // 启用 HTML 标签
  linkify: true,     // 自动转换 URL 为链接
  typographer: true, // 启用排版替换
  breaks: false,     // 不自动转换换行符（与 CommonMark 一致）
})
  .use(markdownItTaskLists, { enabled: false }) // 启用任务列表支持

// 自定义 fence 渲染器以支持 Mermaid
const defaultFence = md.renderer.rules.fence
md.renderer.rules.fence = function (tokens, idx, options, env, slf) {
  const token = tokens[idx]
  const info = token.info ? token.info.trim() : ''
  const langName = info ? info.split(/\s+/g)[0] : ''
  
  if (langName === 'mermaid') {
    const id = 'mermaid-' + Math.random().toString(36).substr(2, 9)
    const code = token.content.trim()
    
    // 异步渲染 Mermaid 图表
    setTimeout(() => {
      try {
        mermaid.render(id + '_svg', code).then(({svg}) => {
          const element = document.getElementById(id)
          if (element) {
            element.innerHTML = svg
          }
        }).catch(error => {
          console.error('Mermaid渲染错误:', error)
          const element = document.getElementById(id)
          if (element) {
            element.innerHTML = `<div class="text-error">Mermaid图表渲染失败: ${error.message}</div>`
          }
        })
      } catch (error) {
        console.error('Mermaid渲染错误:', error)
        const element = document.getElementById(id)
        if (element) {
          element.innerHTML = `<div class="text-error">Mermaid图表渲染失败: ${error.message}</div>`
        }
      }
    }, 0)
    
    return `<div id="${id}" class="mermaid-container">${code}</div>`
  }
  
  // 使用默认的 fence 渲染器处理其他代码块
  return defaultFence(tokens, idx, options, env, slf)
}

const previewContent = computed(() => {
  if (!articleData.content.trim()) {
    return '<p class="text-base-content/60">在左侧编辑区域输入内容，预览将在这里显示...</p>'
  }

  try {
    // 使用 markdown-it 解析 Markdown 内容
    const htmlContent = md.render(articleData.content)
    return htmlContent
  } catch (error) {
    console.error('Markdown解析错误:', error)
    return '<p class="text-error">Markdown解析出错，请检查语法</p>'
  }
})


const charCount = computed(() => {
  return articleData.content.length.toLocaleString()
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
  } else {
    throw new Error(result.msg || '获取文章数据失败')
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

// 分类选择处理
const handleCategoryChange = (selectedCategories: number[]) => {
  articleData.categoryId = selectedCategories
}

const handleCategoryError = (message: string) => {
  showMessage(message, 'error')
}

const clearContent = () => {
  if (confirm('确定要清空所有内容吗？')) {
    articleData.title = ''
    articleData.content = ''
    articleData.type = 1
    articleData.categoryId = []
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


// 生命周期
onMounted(async () => {
  // 初始化 mermaid
  mermaid.initialize({
    startOnLoad: false,
    theme: 'default',
    securityLevel: 'loose'
  })

  await initData()
})
</script>

<template>
  <div class="min-h-screen flex flex-col bg-base-200">
    <!-- 消息提示组件 -->
    <div class="fixed top-16 right-4 z-150 space-y-2">
      <div v-for="message in messages" :key="message.id" :class="[
                'alert w-auto max-w-sm transition-all duration-300',
                {
                    'alert-info': message.type === 'info',
                    'alert-success': message.type === 'success',
                    'alert-error': message.type === 'error'
                }
            ]">
        <span>{{ message.text }}</span>
        <button @click="removeMessage(message.id)" class="btn btn-sm btn-ghost">
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
            <div class="flex-1 p-6 space-y-2">
              <!-- 文章标题区域 -->
              <div class="form-control">
                <label class="label pb-1">
                  <span class="label-text font-medium text-base-content">📝 文章标题</span>
                  <span class="label-text-alt text-base-content/60">必填</span>
                </label>
                <input type="text" v-model="articleData.title" placeholder="请输入一个吸引人的标题..."
                       class="input input-bordered input-md w-full focus:input-primary"/>
              </div>

              <!-- 分类和类型选择区域 -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                <!-- 文章类型 -->
                <div class="form-control">
                  <label class="label pb-1">
                    <span class="label-text font-medium text-base-content">🏷️ 文章类型</span>
                    <span class="label-text-alt text-base-content/60">必选</span>
                  </label>
                  <select v-model="articleData.type"
                          class="select select-bordered w-full focus:select-primary">
                    <option value="">请选择类型</option>
                    <option v-for="type in typeList" :key="type.value" :value="type.value">
                      {{ type.name }}
                    </option>
                  </select>
                </div>

                <!-- 文章分类 -->
                <CategorySelector v-model="articleData.categoryId" :categories="categories"
                                  :max-selection="3" @change="handleCategoryChange" @error="handleCategoryError"/>
              </div>

              <!-- 文章内容区域 -->
              <div class="form-control flex-1">
                <label class="label pb-2">
                  <span class="label-text font-medium text-base-content">✍️ 文章内容</span>
                  <span class="label-text-alt text-base-content/60">支持 Markdown 语法</span>
                </label>
                <div class="relative flex-1">
                                    <textarea v-model="articleData.content"
                                              class="textarea textarea-bordered w-full h-full min-h-96 resize-none focus:textarea-primary font-mono text-sm leading-relaxed"
                                              placeholder="开启你的创作..."></textarea>
                  <!-- 字数统计 -->
                  <div
                      class="absolute bottom-2 right-4 text-xs text-base-content/50 bg-base-100 px-2 py-1 rounded">
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
                          <input type="checkbox" class="checkbox checkbox-sm" checked
                                 disabled/>
                          <span class="label-text text-sm">允许评论</span>
                        </label>
                      </div>
                    </div>
                    <div class="flex items-center gap-2">
                      <button @click="clearContent" class="btn btn-ghost btn-sm">
                        清空内容
                      </button>
                      <button @click="submitArticle" :disabled="isSubmitting"
                              class="btn btn-primary btn-sm">
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
            <h1 class="text-3xl font-normal text-base-content mb-4 border-b border-b-gray-200 pb-2">{{ previewTitle }}</h1>
          </div>
          <div
              class="prose lg:prose-base md:prose-lg prose-h1:font-normal prose-h2:font-normal prose-h3:font-normal prose-pre:bg-base-200 prose-code:text-base-content max-w-none text-base-content overflow-hidden min-w-0"
              v-html="previewContent"></div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
</style>