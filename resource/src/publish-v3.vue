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
const isUploading = ref(false)
const uploadProgress = ref(0)
const isDragOver = ref(false)


// 计算属性
const previewTitle = computed(() => {
  return articleData.title.trim() || '文章标题预览'
})

// 配置 markdown-it 以保持与服务端 Goldmark 的高度一致性
const md = new MarkdownIt({
  html: true,        // 启用 HTML 标签
  linkify: true,     // 自动转换 URL 为链接
  typographer: true, // 启用排版替换
  breaks: true,     // 
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
    
    return `<div id="${id}" class="mermaid">${code}</div>`
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

// 图片上传相关方法
const uploadImage = async (file: File): Promise<string> => {
  if (!file.type.startsWith('image/')) {
    throw new Error('请选择图片文件')
  }

  if (file.size > 10 * 1024 * 1024) {
    throw new Error('图片大小不能超过10MB')
  }

  const formData = new FormData()
  formData.append('file', file)

  isUploading.value = true
  uploadProgress.value = 0

  try {
    const response = await fetch('/file/img-upload', {
      method: 'POST',
      body: formData
    })

    if (!response.ok) {
      throw new Error(`上传失败: ${response.status}`)
    }

    const result = await response.json()
    console.log(result)
    if (result.code === 0 && result.result) {
      return result.result
    } else {
      throw new Error(result.msg || '上传失败')
    }
  } finally {
    isUploading.value = false
    uploadProgress.value = 0
  }
}

const insertImageToContent = (imageUrl: string, altText: string = '') => {
  const textarea = document.querySelector('textarea') as HTMLTextAreaElement
  if (!textarea) return

  const cursorPos = textarea.selectionStart
  const textBefore = articleData.content.substring(0, cursorPos)
  const textAfter = articleData.content.substring(textarea.selectionEnd)
  
  const imageMarkdown = `![${altText}](${imageUrl})`
  articleData.content = textBefore + imageMarkdown + textAfter
  
  // 设置光标位置到插入的图片后面
  setTimeout(() => {
    const newPos = cursorPos + imageMarkdown.length
    textarea.setSelectionRange(newPos, newPos)
    textarea.focus()
  }, 0)
}

const handleFileSelect = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const files = input.files
  if (!files || files.length === 0) return

  try {
    for (const file of Array.from(files)) {
      showMessage('正在上传图片...', 'info')
      const imageUrl = await uploadImage(file)
      insertImageToContent(imageUrl, file.name.split('.')[0])
      showMessage('图片上传成功！', 'success')
    }
  } catch (error) {
    console.error('图片上传失败:', error)
    showMessage((error as Error).message || '图片上传失败', 'error')
  } finally {
    // 清空input值，允许重复选择同一文件
    input.value = ''
  }
}

const handlePaste = async (event: ClipboardEvent) => {
  const items = event.clipboardData?.items
  if (!items) return

  for (const item of Array.from(items)) {
    if (item.type.startsWith('image/')) {
      event.preventDefault()
      const file = item.getAsFile()
      if (!file) continue

      try {
        showMessage('正在上传剪贴板图片...', 'info')
        const imageUrl = await uploadImage(file)
        insertImageToContent(imageUrl, '剪贴板图片')
        showMessage('图片上传成功！', 'success')
      } catch (error) {
        console.error('图片上传失败:', error)
        showMessage((error as Error).message || '图片上传失败', 'error')
      }
      break
    }
  }
}

const handleDragOver = (event: DragEvent) => {
  event.preventDefault()
  event.stopPropagation()
  event.dataTransfer!.dropEffect = 'copy'
  isDragOver.value = true
}

const handleDragLeave = (event: DragEvent) => {
  event.preventDefault()
  event.stopPropagation()
  // 只有当离开整个拖拽区域时才隐藏提示
  const currentTarget = event.currentTarget as HTMLElement
  if (!currentTarget?.contains(event.relatedTarget as Node)) {
    isDragOver.value = false
  }
}

const handleDrop = async (event: DragEvent) => {
  event.preventDefault()
  event.stopPropagation()
  isDragOver.value = false
  const files = event.dataTransfer?.files
  if (!files || files.length === 0) return

  try {
    for (const file of Array.from(files)) {
      if (file.type.startsWith('image/')) {
        showMessage('正在上传拖拽图片...', 'info')
        const imageUrl = await uploadImage(file)
        insertImageToContent(imageUrl, file.name.split('.')[0])
        showMessage('图片上传成功！', 'success')
      }
    }
  } catch (error) {
    console.error('图片上传失败:', error)
    showMessage((error as Error).message || '图片上传失败', 'error')
  }
}

const triggerFileInput = () => {
  const fileInput = document.getElementById('imageUpload') as HTMLInputElement
  fileInput?.click()
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
        <input type="radio" name="my_tabs_3" class="tab" aria-label="文章编写" :checked="true"/>
        <div class="tab-content bg-base-100 border-base-300 p-0">
          <div class="flex flex-col h-full">
            <!-- 编辑区域 -->
            <div class="flex-1 p-6 space-y-2">
              <!-- 文章标题区域 -->
              <div class="form-control">
                <label class="floating-label pb-1">
                  <span class="label-text font-normal text-base-content">📝 文章标题</span>
                  <input type="text" v-model="articleData.title" placeholder="请输入一个吸引人的标题..."
                         class="input input-bordered input-md w-full focus:input-primary"/>
                </label>
              </div>

              <!-- 分类和类型选择区域 -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                <!-- 文章类型 -->
                <div class="form-control">
                  <label class="floating-label pb-1">
                    <span class="label-text font-normal text-base-content">🏷️ 文章类型(必选)</span>
                    <select v-model="articleData.type"
                            class="select select-bordered w-full focus:select-primary">
                      <option value="">请选择类型</option>
                      <option v-for="type in typeList" :key="type.value" :value="type.value">
                        {{ type.name }}
                      </option>
                    </select>
                  </label>
                </div>

                <!-- 文章分类 -->
                <CategorySelector v-model="articleData.categoryId" :categories="categories"
                                  :max-selection="3" @change="handleCategoryChange" @error="handleCategoryError"/>
              </div>

              <!-- 图片上传工具栏 -->
              <div class="flex items-center gap-2 mb-2 p-2 bg-base-200 rounded-lg">
                <button @click="triggerFileInput" type="button" class="btn btn-sm btn-ghost gap-1">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
                  </svg>
                  上传图片
                </button>
                <div class="text-xs text-base-content/60">
                  支持拖拽、粘贴上传 | 最大10MB
                </div>
                <div v-if="isUploading" class="flex items-center gap-2 ml-auto">
                  <span class="loading loading-spinner loading-sm"></span>
                  <span class="text-xs">上传中...</span>
                </div>
              </div>

              <!-- 隐藏的文件输入框 -->
              <input 
                id="imageUpload" 
                type="file" 
                accept="image/*" 
                multiple 
                @change="handleFileSelect" 
                class="hidden"
              />

              <!-- 文章内容区域 -->
              <div class="form-control flex-1">
                <label class="floating-label pb-2">
                  <span class="font-normal text-base-content">✍️ 文章内容-支持 Markdown 语法</span>
                  <div class="relative flex-1">
                    

                    
                    <!-- 文本编辑区域 -->
                    <div 
                      class="relative"
                      @dragover="handleDragOver"
                      @dragleave="handleDragLeave"
                      @drop="handleDrop"
                    >
                      <textarea 
                        v-model="articleData.content"
                        @paste="handlePaste"
                        class="textarea textarea-bordered w-full h-full min-h-96 resize-none focus:textarea-primary font-mono text-sm leading-relaxed"
                        placeholder="开启你的创作...\n\n💡 提示：\n• 直接粘贴图片即可上传\n• 拖拽图片到此区域上传\n• 点击上方按钮选择图片"
                      ></textarea>
                      
                      <!-- 拖拽提示层 -->
                       <div 
                         class="absolute inset-0 bg-primary/10 border-2 border-dashed border-primary rounded-lg flex items-center justify-center opacity-0 pointer-events-none transition-opacity duration-200"
                         :class="{ 'opacity-100': isDragOver || isUploading }"
                       >
                        <div class="text-center">
                           <svg class="w-12 h-12 mx-auto mb-2 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                             <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"></path>
                           </svg>
                           <p class="text-primary font-medium">
                             {{ isUploading ? '正在上传图片...' : '释放以上传图片' }}
                           </p>
                         </div>
                      </div>
                      
                      <!-- 字数统计 -->
                      <div class="absolute bottom-2 right-4 text-xs text-base-content/50 bg-base-100 px-2 py-1 rounded">
                        <span>{{ charCount }}</span> 字符
                      </div>
                    </div>
                  </div>
                </label>
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