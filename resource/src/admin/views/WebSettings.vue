
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  CodeBracketIcon,
  MagnifyingGlassIcon,
  EyeIcon,
  InformationCircleIcon,
  CheckIcon,
  ArrowPathIcon,
  CheckCircleIcon,
  XCircleIcon
} from '@heroicons/vue/24/outline'
import { getWebSettings, saveWebSettings, type WebSettingsConfig } from '../utils/adminService'

// 设置数据
const settings = ref({
  // Meta标签
  metaTags: '',
  // 自定义CSS
  customCSS: '',
  // 自定义JavaScript
  customJS: '',
  // 外部资源链接
  externalLinks: '',
  // 网站图标
  favicon: '',
  appleTouchIcon: '',
  // SEO设置
  ogImage: '',
  twitterCardType: 'summary',
  structuredData: ''
})

// 状态变量
const saving = ref(false)
const showSuccessAlert = ref(false)
const showErrorAlert = ref(false)
const errorMessage = ref('')

// 计算属性
const metaTagsCount = computed(() => {
  return settings.value.metaTags.split('\n').filter(line => line.trim()).length
})

const cssLinesCount = computed(() => {
  return settings.value.customCSS.split('\n').filter(line => line.trim()).length
})

const jsLinesCount = computed(() => {
  return settings.value.customJS.split('\n').filter(line => line.trim()).length
})

const externalLinksCount = computed(() => {
  return settings.value.externalLinks.split('\n').filter(line => line.trim()).length
})

// 加载设置
const loadSettings = async () => {
  try {
    const response = await getWebSettings()
    if (response.code === 0) {
      Object.assign(settings.value, response.result)
    } else {
      showError(response.message || '加载设置失败')
    }
  } catch (error) {
    console.error('加载网页设置失败:', error)
    showError('加载设置失败，请重试')
  }
}

// 保存设置
const saveSettings = async () => {
  saving.value = true
  try {
    const response = await saveWebSettings(settings.value as WebSettingsConfig)
    if (response.code === 0) {
      showSuccess()
    } else {
      showError(response.message || '保存设置失败')
    }
  } catch (error) {
    console.error('保存网页设置失败:', error)
    showError('保存设置失败，请重试')
  } finally {
    saving.value = false
  }
}

// 重置为默认值
const resetToDefaults = () => {
  if (confirm('确定要重置为默认设置吗？这将清空当前所有自定义内容。')) {
    Object.assign(settings.value, {
      metaTags: '',
      customCSS: '',
      customJS: '',
      externalLinks: '',
      favicon: '',
      appleTouchIcon: '',
      ogImage: '',
      twitterCardType: 'summary',
      structuredData: ''
    })
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
  loadSettings()
})
</script>

<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-normal text-base-content">网页设置</h1>
        <p class="text-base-content/70 mt-1">配置网站公共头部内容和SEO设置</p>
      </div>
      <div class="flex gap-2">
        <button @click="resetToDefaults" class="btn btn-outline btn-warning btn-sm">
          <ArrowPathIcon class="w-4 h-4" />
          重置默认
        </button>
        <button @click="saveSettings" class="btn btn-primary btn-sm" :disabled="saving">
          <span v-if="saving" class="loading loading-spinner loading-sm"></span>
          <CheckIcon v-else class="w-4 h-4" />
          保存设置
        </button>
      </div>
    </div>

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

    <!-- 设置表单 -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- 主要设置 -->
      <div class="lg:col-span-2 space-y-6">
        <!-- 公共Head设置 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2 mb-6">
              <CodeBracketIcon class="w-5 h-5" />
              公共Head设置
            </h2>

            <div class="space-y-6">
              <!-- Meta标签 -->
              <div class="form-control">
                <label class="floating-label">
                  <span>Meta标签</span>
                  <textarea 
                    v-model="settings.metaTags" 
                    class="textarea textarea-bordered h-32 w-full font-mono text-sm" 
                    placeholder="请输入Meta标签，例如：&#10;&lt;meta name=&quot;author&quot; content=&quot;Your Name&quot;&gt;&#10;&lt;meta name=&quot;robots&quot; content=&quot;index,follow&quot;&gt;"
                  ></textarea>
                </label>
                <div class="label">
                  <span class="label-text-alt text-base-content/60">用于SEO优化的Meta标签，每行一个标签</span>
                </div>
              </div>

              <!-- 自定义CSS -->
              <div class="form-control">
                <label class="floating-label">
                  <span>自定义CSS</span>
                  <textarea 
                    v-model="settings.customCSS" 
                    class="textarea textarea-bordered h-40 w-full font-mono text-sm" 
                    placeholder="请输入自定义CSS代码，例如：&#10;body { font-family: 'Arial', sans-serif; }&#10;.custom-class { color: #333; }"
                  ></textarea>
                </label>
                <div class="label">
                  <span class="label-text-alt text-base-content/60">自定义样式代码，将被包含在&lt;style&gt;标签中</span>
                </div>
              </div>

              <!-- 自定义JavaScript -->
              <div class="form-control">
                <label class="floating-label">
                  <span>自定义JavaScript</span>
                  <textarea 
                    v-model="settings.customJS" 
                    class="textarea textarea-bordered h-40 w-full font-mono text-sm" 
                    placeholder="请输入自定义JavaScript代码，例如：&#10;console.log('Custom script loaded');&#10;// 你的自定义代码"
                  ></textarea>
                </label>
                <div class="label">
                  <span class="label-text-alt text-base-content/60">自定义脚本代码，将被包含在&lt;script&gt;标签中</span>
                </div>
              </div>

              <!-- 外部资源链接 -->
              <div class="form-control">
                <label class="floating-label">
                  <span>外部资源链接</span>
                  <textarea 
                    v-model="settings.externalLinks" 
                    class="textarea textarea-bordered h-32 w-full font-mono text-sm" 
                    placeholder="请输入外部资源链接，例如：&#10;&lt;link rel=&quot;stylesheet&quot; href=&quot;https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&display=swap&quot;&gt;&#10;&lt;script src=&quot;https://cdn.jsdelivr.net/npm/chart.js&quot;&gt;&lt;/script&gt;"
                  ></textarea>
                </label>
                <div class="label">
                  <span class="label-text-alt text-base-content/60">外部CSS和JS资源链接，每行一个链接</span>
                </div>
              </div>

              <!-- 网站图标设置 -->
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="form-control">
                  <label class="floating-label">
                    <span>网站图标(Favicon)</span>
                    <input 
                      v-model="settings.favicon" 
                      type="url" 
                      placeholder="https://example.com/favicon.ico" 
                      class="input input-bordered w-full" 
                    />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">网站图标URL，支持.ico、.png格式</span>
                  </div>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>Apple Touch图标</span>
                    <input 
                      v-model="settings.appleTouchIcon" 
                      type="url" 
                      placeholder="https://example.com/apple-touch-icon.png" 
                      class="input input-bordered w-full" 
                    />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">iOS设备添加到主屏幕时的图标</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

      </div>

      <!-- 侧边栏 -->
      <div class="space-y-6">
        <!-- 预览 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2 mb-6">
              <EyeIcon class="w-5 h-5" />
              预览
            </h2>
            <div class="space-y-4">
              <div class="p-3 bg-base-200 rounded-lg">
                <div class="text-sm font-normal mb-2">Meta标签数量</div>
                <div class="text-lg font-semibold">{{ metaTagsCount }}</div>
              </div>
              <div class="p-3 bg-base-200 rounded-lg">
                <div class="text-sm font-normal mb-2">CSS代码行数</div>
                <div class="text-lg font-semibold">{{ cssLinesCount }}</div>
              </div>
              <div class="p-3 bg-base-200 rounded-lg">
                <div class="text-sm font-normal mb-2">JS代码行数</div>
                <div class="text-lg font-semibold">{{ jsLinesCount }}</div>
              </div>
              <div class="p-3 bg-base-200 rounded-lg">
                <div class="text-sm font-normal mb-2">外部资源数量</div>
                <div class="text-lg font-semibold">{{ externalLinksCount }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 帮助信息 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2 mb-6">
              <InformationCircleIcon class="w-5 h-5" />
              使用说明
            </h2>
            <div class="space-y-3 text-sm text-base-content/70">
              <div class="p-3 bg-info/10 rounded-lg">
                <div class="font-medium text-info mb-1">Meta标签</div>
                <div>用于SEO优化，每行输入一个完整的meta标签</div>
              </div>
              <div class="p-3 bg-warning/10 rounded-lg">
                <div class="font-medium text-warning mb-1">自定义CSS</div>
                <div>直接输入CSS代码，无需包含&lt;style&gt;标签</div>
              </div>
              <div class="p-3 bg-success/10 rounded-lg">
                <div class="font-medium text-success mb-1">自定义JS</div>
                <div>直接输入JavaScript代码，无需包含&lt;script&gt;标签</div>
              </div>
              <div class="p-3 bg-error/10 rounded-lg">
                <div class="font-medium text-error mb-1">注意事项</div>
                <div>请确保代码语法正确，错误的代码可能影响网站正常运行</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>