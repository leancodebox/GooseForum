<script setup>
import {computed, onMounted, onUnmounted, ref} from 'vue'
import {MdEditor} from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import {getArticleEnum, getArticlesOrigin, submitArticle} from './utils/articleService.js'
import {NConfigProvider, NSelect} from 'naive-ui'
import {createSelectThemeOverrides, getBaseTheme} from './components/nu-theme.ts'

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
  {label: 'GooseForum', value: 1},
])

const typeList = ref([
  {label: '分享', value: 1},
  {label: '求助', value: 2},
])

// 编辑器配置
const toolbars = [
  'bold', 'underline', 'italic', 'strikeThrough', '-',
  'title', 'sub', 'sup', 'quote', 'unorderedList', 'orderedList', 'task', '-',
  'codeRow', 'code', 'link', 'image', 'table', 'mermaid', 'katex', '-',
  'revoke', 'next', 'save', '=', 'pageFullscreen', 'fullscreen', 'preview', 'htmlPreview', 'catalog'
]

// 状态管理
const isSubmitting = ref(false)

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
  console.log(articleData.value)
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

// 主题相关
const currentTheme = ref('light')

// Naive UI 主题配置
const naiveTheme = computed(() => {
  return getBaseTheme(currentTheme.value)
})

const naiveThemeOverrides = computed(() => {
  console.log(currentTheme.value)
  return createSelectThemeOverrides(currentTheme.value)
})


// 监听主题变化
const observeThemeChange = () => {
  // 初始化当前主题
  const initialTheme = document.documentElement.getAttribute('data-theme') || 'light'
  currentTheme.value = initialTheme

  // 创建 MutationObserver 监听主题变化
  const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (mutation.type === 'attributes' && mutation.attributeName === 'data-theme') {
        const newTheme = document.documentElement.getAttribute('data-theme') || 'light'
        currentTheme.value = newTheme
      }
    })
  })

  // 开始监听
  observer.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['data-theme']
  })

  return observer
}


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

    // 启动主题监听器
    const themeObserver = observeThemeChange()
  } catch (error) {
    console.error('初始化失败:', error)
  }
})

onUnmounted(() => {
})


</script>
<template>
  <n-config-provider :theme="naiveTheme" :theme-overrides="naiveThemeOverrides">
    <div class="max-w-6xl mx-auto py-2 px-4">
      <!-- 发布表单 -->
      <form @submit.prevent="handleSubmit" class="space-y-6">
        <!-- 基本信息响应式布局 -->
        <div class="space-y-4">
          <!-- 文章标题 - 始终占满宽度 -->
          <div class="w-full">
            <label class="form-control w-full">
              <div class="label">
                <span class="label-text font-medium">文章标题</span>
                <span class="label-text-alt text-error">*</span>
              </div>
              <input v-model="articleData.articleTitle" type="text" placeholder="请输入文章标题"
                     class="input input-bordered w-full"
                     required/>
            </label>
          </div>

          <!-- 类型和分类 - 响应式布局 -->
          <div class="flex flex-col sm:flex-row gap-4">
            <!-- 文章类型 -->
            <div class="w-full sm:w-1/3">
              <label class="form-control w-full">
                <div class="label">
                  <span class="label-text font-medium">类型</span>
                  <span class="label-text-alt text-error">*</span>
                </div>
                <n-select
                    v-model:value="articleData.type"
                    :options="typeList"
                    placeholder="选择类型"
                />
              </label>
            </div>

            <!-- 分类选择 -->
            <div class="w-full sm:w-2/3">
              <label class="form-control w-full">
                <div class="label">
                  <span class="label-text font-medium">文章分类</span>
                  <span class="label-text-alt text-error">*</span>
                </div>
                <n-select
                    v-model:value="articleData.categoryId"
                    :options="categories"
                    placeholder="请选择分类"
                    :max-tag-count="3"
                    multiple
                />
              </label>
            </div>
          </div>
        </div>

        <!-- Markdown 编辑器 -->


        <div class="border border-base-300 bg-base-100 rounded-lg overflow-hidden">
          <MdEditor
              v-model="articleData.articleContent"
              :height="500"
              :preview="true"
              :toolbars="toolbars"
              :theme="['dark','garden','luxury','dracula'].indexOf(currentTheme)=== -1?'light':'dark'"
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
  </n-config-provider>
</template>


<style>

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

.md-editor-preview ul ul {
  list-style-type: circle;
}

.md-editor-preview ul ul ul {
  list-style-type: square;
}

.md-editor-preview ol ol {
  list-style-type: lower-alpha;
}

.md-editor-preview ol ol ol {
  list-style-type: lower-roman;
}

</style>