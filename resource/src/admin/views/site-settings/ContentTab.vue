<template>
  <div class="space-y-6">
    <!-- 成功提示 -->
    <div v-if="showSuccessAlert" class="alert alert-success">
      <CheckCircleIcon class="w-6 h-6" />
      <span>设置保存成功！</span>
    </div>

    <!-- 错误提示 -->
    <div v-if="showErrorAlert" class="alert alert-error">
      <XCircleIcon class="w-6 h-6" />
      <span>{{ errorMessage }}</span>
    </div>

    <!-- 内容设置表单 -->
    <div class="space-y-8">
      <!-- 文章设置 -->
      <div>
        <div class="flex justify-between items-center">
          <h2 class="flex items-center gap-2 mb-6 text-lg font-medium text-base-content">
            <DocumentTextIcon class="w-5 h-5" />
            文章设置
          </h2>
          <button @click="handleSave" class="btn btn-primary btn-sm" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-sm"></span>
            <CheckIcon v-else class="w-4 h-4" />
            {{ saving ? '保存中...' : '保存设置' }}
          </button>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- 每页文章数量 -->
          <label class="floating-label">
            <span>每页文章数量</span>
            <input
              v-model="settings.articlesPerPage"
              type="number"
              min="1"
              max="100"
              class="input input-bordered w-full peer"
              id="articlesPerPage"
              placeholder="每页文章数量"
            />
          </label>

          <!-- 文章摘要长度 -->
          <label class="floating-label">
            <span>文章摘要长度（字符）</span>
            <input
              v-model="settings.excerptLength"
              type="number"
              min="50"
              max="500"
              class="input input-bordered w-full peer"
              id="excerptLength"
              placeholder="文章摘要长度（字符）"
            />
          </label>

          <!-- 允许匿名评论 -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <input v-model="settings.allowAnonymousComments" type="checkbox" class="toggle toggle-primary" />
              <span class="label-text">允许匿名评论</span>
            </label>
          </div>

          <!-- 评论审核 -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <input v-model="settings.commentModeration" type="checkbox" class="toggle toggle-primary" />
              <span class="label-text">评论需要审核</span>
            </label>
          </div>
          <!-- 最大文件大小 -->
          <label class="floating-label">
            <span>最大文件大小（MB）</span>
            <input
              v-model="settings.maxFileSize"
              type="number"
              min="1"
              max="100"
              class="input input-bordered w-full peer"
              id="maxFileSize"
              placeholder="最大文件大小（MB）"
            />
            <div class="label">
              <span class="label-text-alt text-base-content/60">最大文件大小</span>
            </div>
          </label>

          <!-- 图片压缩质量 -->
          <label class="floating-label">
            <span>图片压缩质量（%）</span>
            <input
              v-model="settings.imageQuality"
              type="number"
              min="10"
              max="100"
              class="input input-bordered w-full peer"
              id="imageQuality"
              placeholder="图片压缩质量（%）"
            />

            <div class="label">
              <span class="label-text-alt text-base-content/60">图片压缩质量</span>
            </div>
          </label>

          <!-- 允许的文件类型 -->
          <div class="form-control">
            <label class="floating-label">
              <span>允许的文件类型</span>
              <input v-model="settings.allowedFileTypes" type="text" placeholder="请输入允许的文件类型" class="input input-bordered bg-base-200 w-full" />
            </label>
            <div class="label">
              <span class="label-text-alt text-base-content/60">用逗号分隔多个文件扩展名</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 分割线 -->
      <div class="divider"></div>

      <!-- 内容过滤 -->
      <div>
        <h2 class="flex items-center gap-2 mb-6 text-lg font-medium text-base-content">
          <ShieldCheckIcon class="w-5 h-5" />
          内容过滤
        </h2>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div class="space-y-4">
            <!-- 启用内容过滤 -->
            <div class="form-control">
              <label class="label cursor-pointer">
                <input v-model="settings.enableContentFilter" type="checkbox" class="toggle toggle-primary" />
                <span class="label-text">启用内容过滤</span>
              </label>
            </div>

            <!-- 敏感词列表 -->
            <label class="floating-label">
              <span>敏感词列表</span>
              <textarea
                v-model="settings.bannedWords"
                class="textarea textarea-bordered w-full peer h-32"
                id="bannedWords"
                placeholder="敏感词列表"
              ></textarea>
            </label>
            <div class="text-sm text-base-content/60">
              每行一个敏感词，支持通配符 *
            </div>
          </div>

          <div class="space-y-4">
            <!-- 自动替换敏感词 -->
            <div class="form-control">
              <label class="label cursor-pointer">
                <input v-model="settings.autoReplaceBannedWords" type="checkbox" class="toggle toggle-primary" />
                <span class="label-text">自动替换敏感词</span>
              </label>
            </div>

            <!-- 替换字符 -->
            <label class="floating-label">
              <span>替换字符</span>
              <input
                v-model="settings.replacementChar"
                type="text"
                maxlength="5"
                class="input input-bordered w-full peer"
                id="replacementChar"
                placeholder="替换字符"
              />
            </label>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  CheckIcon,
  CheckCircleIcon,
  XCircleIcon,
  DocumentTextIcon,
  CloudArrowUpIcon,
  ShieldCheckIcon, UserPlusIcon
} from '@heroicons/vue/24/outline'

// 设置数据
const settings = ref({
  // 文章设置
  articlesPerPage: 20,
  excerptLength: 200,
  allowAnonymousComments: true,
  commentModeration: false,

  // 上传设置
  maxFileSize: 10,
  allowedFileTypes: 'jpg,png,gif,pdf,doc,docx',
  imageQuality: 80,

  // 内容过滤
  enableContentFilter: true,
  bannedWords: '垃圾\n广告\n违法',
  autoReplaceBannedWords: true,
  replacementChar: '*'
})

// 状态变量
const saving = ref(false)
const showSuccessAlert = ref(false)
const showErrorAlert = ref(false)
const errorMessage = ref('')

// 保存设置
const handleSave = async () => {
  saving.value = true
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    showSuccess()
  } catch (error) {
    console.error('保存内容设置失败:', error)
    showError('保存设置失败，请重试')
  } finally {
    saving.value = false
  }
}

// 显示成功提示
const showSuccess = (message = '设置保存成功！') => {
  showSuccessAlert.value = true
  setTimeout(() => {
    showSuccessAlert.value = false
  }, 3000)
}

// 显示错误提示
const showError = (message: string) => {
  errorMessage.value = message
  showErrorAlert.value = true
  setTimeout(() => {
    showErrorAlert.value = false
  }, 5000)
}

// 组件挂载时加载设置
onMounted(() => {
  // 这里可以加载实际的设置数据
  console.log('内容设置页面已加载')
})
</script>

<style scoped>
/* 自定义样式 */
</style>
