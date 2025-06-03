<script setup>
import { ref, computed } from 'vue'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

// 表单数据
const form = ref({
  title: '',
  category: '',
  type: 'original',
  tags: '',
  content: ''
})

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

// 表单验证
const validateForm = () => {
  if (!form.value.title.trim()) {
    alert('请输入文章标题')
    return false
  }
  if (!form.value.category) {
    alert('请选择文章分类')
    return false
  }
  if (!form.value.content.trim()) {
    alert('请输入文章内容')
    return false
  }
  return true
}

// 提交表单
const handleSubmit = async () => {
  if (!validateForm()) return

  isSubmitting.value = true
  try {
    // 这里添加实际的提交逻辑
    console.log('提交文章:', form.value)

    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 2000))

    alert('文章发布成功！')
    // 可以跳转到文章详情页或列表页
    // await navigateTo('/list')
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
    // 这里添加保存草稿的逻辑
    console.log('保存草稿:', form.value)
    alert('草稿保存成功！')
  } catch (error) {
    console.error('保存草稿失败:', error)
    alert('保存草稿失败，请重试')
  }
}


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
            <input v-model="form.title" type="text" placeholder="请输入文章标题" class="input input-bordered w-full"
                   required />
          </label>
        </div>

        <!-- 分类选择 -->
        <div>
          <label class="form-control w-full">
            <div class="label">
              <span class="label-text font-medium">文章分类</span>
              <span class="label-text-alt text-error">*</span>
            </div>
            <select v-model="form.category" class="select select-bordered w-full" required>
              <option value="">请选择分类</option>
              <option value="frontend">前端开发</option>
              <option value="backend">后端开发</option>
              <option value="mobile">移动开发</option>
              <option value="devops">运维部署</option>
              <option value="database">数据库</option>
              <option value="ai">人工智能</option>
              <option value="other">其他技术</option>
            </select>
          </label>
        </div>

        <!-- 文章类型 -->
        <div>
          <label class="form-control w-full">
            <div class="label">
              <span class="label-text font-medium">文章类型</span>
            </div>
            <select v-model="form.type" class="select select-bordered w-full">
              <option value="original">原创</option>
              <option value="reprint">转载</option>
              <option value="translation">翻译</option>
            </select>
          </label>
        </div>
      </div>

      <!-- 标签输入 -->
      <div>
        <label class="form-control w-full">
          <div class="label">
            <span class="label-text font-medium">文章标签</span>
            <span class="label-text-alt">用逗号分隔多个标签</span>
          </div>
          <input v-model="form.tags" type="text" placeholder="例如: Vue.js, JavaScript, 前端"
                 class="input input-bordered w-full" />
        </label>
      </div>

      <!-- Markdown 编辑器 -->


      <div class="border border-base-300 bg-base-100 rounded-lg overflow-hidden">
        <MdEditor
            v-model="form.content"
            :height="500"
            :preview="true"
            :toolbars="toolbars"
            :theme="editorTheme"
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