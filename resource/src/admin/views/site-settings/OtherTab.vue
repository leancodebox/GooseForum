<template>
  <div class="space-y-6">
    <!-- 保存按钮 -->
    <div class="flex justify-between items-center">
      <h3 class="text-lg font-medium text-base-content">其他设置</h3>
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
      <div class="form-control">
        <div class="floating-label">
          <select 
             v-model="settings.timezone"
             id="timezone"
             class="select select-bordered w-full peer"
           >
            <option value="Asia/Shanghai">Asia/Shanghai (北京时间)</option>
            <option value="UTC">UTC (协调世界时)</option>
            <option value="America/New_York">America/New_York (东部时间)</option>
            <option value="America/Los_Angeles">America/Los_Angeles (太平洋时间)</option>
            <option value="Europe/London">Europe/London (格林威治时间)</option>
            <option value="Asia/Tokyo">Asia/Tokyo (日本时间)</option>
          </select>
          <span>时区</span>
        </div>
      </div>

      <div class="form-control">
        <div class="floating-label">
          <select 
             v-model="settings.defaultLanguage"
             id="defaultLanguage"
             class="select select-bordered w-full peer"
           >
            <option value="zh-CN">简体中文</option>
            <option value="zh-TW">繁体中文</option>
            <option value="en-US">English</option>
            <option value="ja-JP">日本語</option>
            <option value="ko-KR">한국어</option>
          </select>
          <span>默认语言</span>
        </div>
      </div>
    </div>

    <div class="divider">维护模式</div>

    <div class="form-control">
      <label class="label cursor-pointer">
        <span class="label-text">启用维护模式</span>
        <input 
           v-model="settings.maintenanceMode" 

           type="checkbox" 
           class="toggle toggle-primary"
         />
      </label>
      <div class="label">
        <span class="label-text-alt">启用后，普通用户将无法访问网站</span>
      </div>
    </div>

    <div v-if="settings.maintenanceMode" class="form-control">
      <div class="floating-label">
        <textarea 
           v-model="settings.maintenanceMessage"
           id="maintenanceMessage"
           placeholder="" 
           class="textarea textarea-bordered h-20 peer"
         ></textarea>
        <span>维护提示信息</span>
      </div>
      <div class="label">
        <span class="label-text-alt">用户在维护模式下看到的提示信息</span>
      </div>
    </div>

    <div class="divider">系统信息</div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="stat bg-base-200 rounded-lg">
        <div class="stat-title">系统版本</div>
        <div class="stat-value text-lg">GooseForum v1.0</div>
        <div class="stat-desc">当前运行版本</div>
      </div>

      <div class="stat bg-base-200 rounded-lg">
        <div class="stat-title">运行时间</div>
        <div class="stat-value text-lg">{{ formatUptime() }}</div>
        <div class="stat-desc">系统运行时长</div>
      </div>
    </div>

    <div class="alert alert-warning">
      <div class="flex items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" /></svg>
        <div>
          <h3 class="font-bold">注意事项</h3>
          <div class="text-xs">修改时区和语言设置可能需要重启服务才能生效</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getSiteSettings, saveSiteSettings, type SiteSettingsConfig } from '../../utils/adminService'

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
const uptime = ref(0)

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

const updateValue = (key: keyof SiteSettingsConfig, value: any) => {
  (settings.value as any)[key] = value
}

const formatUptime = () => {
  const days = Math.floor(uptime.value / (24 * 60 * 60))
  const hours = Math.floor((uptime.value % (24 * 60 * 60)) / (60 * 60))
  const minutes = Math.floor((uptime.value % (60 * 60)) / 60)
  
  if (days > 0) {
    return `${days}天 ${hours}小时`
  } else if (hours > 0) {
    return `${hours}小时 ${minutes}分钟`
  } else {
    return `${minutes}分钟`
  }
}

onMounted(() => {
  loadSettings()
  // 模拟系统运行时间
  uptime.value = Math.floor(Math.random() * 86400 * 7) // 0-7天的随机时间
})
</script>