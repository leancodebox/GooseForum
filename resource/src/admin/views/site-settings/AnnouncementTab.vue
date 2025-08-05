<template>
  <div class="space-y-6">
    <!-- 成功提示 -->
    <div v-if="showSuccessAlert" class="alert alert-success">
      <CheckCircleIcon class="w-6 h-6"/>
      <span>公告设置保存成功！</span>
    </div>

    <!-- 错误提示 -->
    <div v-if="showErrorAlert" class="alert alert-error">
      <XCircleIcon class="w-6 h-6"/>
      <span>{{ errorMessage }}</span>
    </div>

    <!-- 基本设置 -->
    <div>
      <div class="flex justify-between items-center">
        <h2 class="flex items-center gap-2 mb-6 text-lg font-medium text-base-content">
          <BellIcon class="w-5 h-5"/>
          公告基本设置
        </h2>

        <div class="join">
          <button
              class="btn btn-primary btn-sm join-item"
              @click="saveSettings"
              :disabled="saving"
          >
            <ArrowPathIcon v-if="saving" class="w-4 h-4 animate-spin"/>
            <CheckIcon v-else class="w-4 h-4"/>
            {{ saving ? '保存中...' : '保存设置' }}
          </button>

          <button
              class="btn btn-outline btn-sm join-item"
              @click="resetToDefaults"
              :disabled="saving"
          >
            重置为默认值
          </button>
        </div>
      </div>

      <div class="space-y-4">
        <!-- 启用公告 -->
        <div class="form-control">
          <label class="label cursor-pointer justify-start gap-3">
            <input
                type="checkbox"
                class="toggle toggle-primary"
                v-model="settings.enabled"
            />
            <span class="label-text font-medium">启用公告</span>
          </label>
          <br>
          <div class="label">
            <span class="label-text-alt text-base-content/60">开启后，公告将在网站前台显示</span>
          </div>
        </div>

        <!-- 公告标题 -->
        <div class="form-control">
          <label class="floating-label">
            <span>公告标题</span>
            <input
                type="text"
                class="input input-bordered w-full"
                v-model="settings.title"
                placeholder="请输入公告标题"
                :disabled="!settings.enabled"
            />
          </label>
        </div>

        <!-- 公告内容 -->
        <div class="form-control">
          <label class="floating-label">
            <span>公告内容</span>
            <textarea
                class="textarea textarea-bordered h-32 w-full"
                v-model="settings.content"
                placeholder="请输入公告内容"
                :disabled="!settings.enabled"
            ></textarea>
          </label>
          <div class="label">
            <span class="label-text-alt text-base-content/60">支持HTML标签和Markdown语法</span>
          </div>
        </div>

        <!-- 跳转链接 -->
        <div class="form-control">
          <label class="floating-label">
            <span>跳转链接（可选）</span>

            <input
                type="url"
                class="input input-bordered w-full"
                v-model="settings.link"
                placeholder="https://example.com"
                :disabled="!settings.enabled"
            />
          </label>
          <div class="label">
            <span class="label-text-alt text-base-content/60">点击公告时跳转到指定链接</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue'
import {ArrowPathIcon, BellIcon, CheckCircleIcon, CheckIcon, XCircleIcon} from '@heroicons/vue/24/outline'
import {type AnnouncementConfig, getAnnouncement, saveAnnouncement} from '../../utils/adminService'

// 设置数据
const settings = ref({
  enabled: false,
  title: '',
  content: '',
  link: ''
})

// 状态变量
const saving = ref(false)
const showSuccessAlert = ref(false)
const showErrorAlert = ref(false)
const errorMessage = ref('')

// 加载设置
const loadSettings = async () => {
  try {
    const response = await getAnnouncement()
    if (response.code === 0) {
      Object.assign(settings.value, response.result)
    } else {
      showError(response.msg || '加载设置失败')
    }
  } catch (error) {
    console.error('加载公告设置失败:', error)
    showError('加载设置失败，请重试')
  }
}

// 保存设置
const saveSettings = async () => {
  saving.value = true
  try {
    const response = await saveAnnouncement(settings.value as AnnouncementConfig)
    if (response.code === 0) {
      showSuccess()
    } else {
      showError(response.msg || '保存设置失败')
    }
  } catch (error) {
    console.error('保存公告设置失败:', error)
    showError('保存设置失败，请重试')
  } finally {
    saving.value = false
  }
}

// 重置为默认值
const resetToDefaults = () => {
  settings.value = {
    enabled: false,
    title: '',
    content: '',
    link: ''
  }
}

// 显示成功提示
const showSuccess = () => {
  showSuccessAlert.value = true
  showErrorAlert.value = false
  setTimeout(() => {
    showSuccessAlert.value = false
  }, 3000)
}

// 显示错误提示
const showError = (message: string) => {
  errorMessage.value = message
  showErrorAlert.value = true
  showSuccessAlert.value = false
  setTimeout(() => {
    showErrorAlert.value = false
  }, 5000)
}

// 组件挂载时加载设置
onMounted(() => {
  loadSettings()
})
</script>

<style scoped>

</style>
