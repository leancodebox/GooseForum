<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <h3 class="text-lg font-medium text-base-content">基本信息</h3>
      <button
          @click="saveSettings"
          :disabled="saving"
          class="btn btn-primary btn-sm"
      >
        <span v-if="saving" class="loading loading-spinner loading-sm mr-2"></span>
        {{ saving ? '保存中...' : '保存设置' }}
      </button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="floating-label">
        <input
            v-model="settings.siteName"

            type="text"
            id="siteName"
            placeholder=""
            class="input input-bordered w-full peer"
        />
        <span>站点名称</span>
      </div>

      <div class="floating-label">
        <input
            v-model="settings.siteUrl"

            type="url"
            id="siteUrl"
            placeholder=""
            class="input input-bordered w-full peer"
        />
        <span>站点URL</span>
      </div>
    </div>

    <div class="floating-label">
          <textarea
              v-model="settings.siteDescription"

              id="siteDescription"
              placeholder=""
              class="textarea textarea-bordered peer w-full"
              rows="3"
          ></textarea>
      <span>站点描述</span>
    </div>

    <div class="floating-label">
      <input
          v-model="settings.siteKeywords"

          type="text"
          id="siteKeywords"
          placeholder=""
          class="input input-bordered w-full peer"
      />
      <span>关键词（用逗号分隔）</span>
    </div>

    <div class="form-control">
      <div class="flex gap-4 items-start">
        <div class="flex-1">
          <div class="floating-label">
            <input
                v-model="settings.siteLogo"

                type="url"
                id="siteLogo"
                placeholder=""
                class="input input-bordered w-full peer"
            />
            <span>Logo URL</span>
          </div>
        </div>
        <div class="flex flex-col items-center gap-2">
          <input
              ref="logoFileInput"
              type="file"
              accept="image/*"
              class="hidden"
              @change="handleLogoUpload"
          />
          <button
              type="button"
              class="btn btn-outline btn-sm"
              @click="logoFileInput?.click()"
              :disabled="uploading"
          >
            <PhotoIcon class="w-4 h-4 mr-1"/>
            {{ uploading ? '上传中...' : '上传' }}
          </button>
          <div v-if="settings.siteLogo" class="w-16 h-16 border border-gray-200 rounded overflow-hidden">
            <img :src="settings.siteLogo" alt="Logo预览" class="w-full h-full object-cover"/>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue'
import {PhotoIcon} from '@heroicons/vue/24/outline'
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

const uploading = ref(false)
const saving = ref(false)
const logoFileInput = ref<HTMLInputElement>()

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

// 处理Logo上传
const handleLogoUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (!file.type.startsWith('image/')) {
    alert('请选择图片文件')
    return
  }

  if (file.size > 2 * 1024 * 1024) {
    alert('图片文件大小不能超过2MB')
    return
  }

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)

    const response = await fetch('/file/img-upload', {
      method: 'POST',
      body: formData,
      credentials: 'include'
    })

    const result = await response.json()
    if (result.code === 0) {
      settings.value.siteLogo = result.result.url
      // 图片上传成功后自动保存
      await saveSettings()
    } else {
      alert(result.msg || result.message || 'Logo上传失败')
    }
  } catch (error) {
    console.error('Logo上传失败:', error)
    alert('Logo上传失败')
  } finally {
    uploading.value = false
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
