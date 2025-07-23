<template>
  <div class="space-y-6">
    <!-- 保存按钮 -->
    <div class="flex justify-between items-center">
      <h3 class="text-lg font-medium text-base-content">SEO设置</h3>
      <button 
        @click="saveSettings" 
        :disabled="saving"
        class="btn btn-primary btn-sm"
      >
        <span v-if="saving" class="loading loading-spinner loading-sm mr-2"></span>
        {{ saving ? '保存中...' : '保存设置' }}
      </button>
    </div>

    <div class="form-control">
      <div class="floating-label">
        <input
            v-model="settings.titleTemplate"
 
            type="text"
            id="titleTemplate"
            placeholder=""
            class="input input-bordered w-full peer"
        />
        <span>页面标题模板</span>
      </div>
      <div class="label">
        <span class="label-text-alt">可用变量: {title}, {siteName}</span>
      </div>
    </div>

    <div class="form-control">
      <div class="floating-label">
        <textarea
            v-model="settings.defaultDescription"
 
            id="defaultDescription"
            placeholder=""
            class="textarea textarea-bordered h-24 peer w-full"
        ></textarea>
        <span>默认描述</span>
      </div>
      <div class="label">
        <span class="label-text-alt">建议长度: 120-160个字符</span>
      </div>
    </div>

    <div class="form-control">
      <div class="floating-label">
        <input
            v-model="settings.icpNumber"
 
            type="text"
            id="icpNumber"
            placeholder=""
            class="input input-bordered w-full peer"
        />
        <span>ICP备案号</span>
      </div>
      <div class="label">
        <span class="label-text-alt">如果有ICP备案，请填写备案号</span>
      </div>
    </div>

    <div class="divider">高级SEO设置</div>

    <div class="alert alert-info">
      <div class="flex items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
        </svg>
        <span>高级SEO功能（如自定义Meta标签、统计代码）将在后续版本中添加</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue'
import {getSiteSettings, saveSiteSettings, type SiteSettingsConfig} from '../../utils/adminService'

// 本地数据状态
const settings = ref<SiteSettingsConfig>({
  siteName: '',
  siteLogo: '',
  siteDescription: '',
  siteKeywords: '',
  siteUrl: '',
  titleTemplate: '',
  defaultDescription: '',
  icpNumber: '',
  timezone: '',
  defaultLanguage: '',
  maintenanceMode: false,
  maintenanceMessage: ''
})

const saving = ref(false)

// 加载设置
const loadSettings = async () => {
  try {
    const response = await getSiteSettings()
    if (response.code === 0) {
      Object.assign(settings.value, response.result)
    }
  } catch (error) {
    console.error('加载设置失败:', error)
  }
}

// 保存设置
const saveSettings = async () => {
  saving.value = true
  try {
    const response = await saveSiteSettings(settings.value)
    if (response.code === 0) {
      // 可以添加成功提示
    }
  } catch (error) {
    console.error('保存设置失败:', error)
  } finally {
    saving.value = false
  }
}

// 手动保存功能
const handleSave = async () => {
  await saveSettings()
}

onMounted(() => {
  loadSettings()
})
</script>