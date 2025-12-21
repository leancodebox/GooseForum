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
        <span>站点名称</span>
        <input
            v-model="settings.siteName"

            type="text"
            id="siteName"
            placeholder="站点名称"
            class="input input-bordered w-full peer"
        />
      </div>

      <div class="floating-label">
        <span>站点URL</span>
        <input
            v-model="settings.siteUrl"

            type="url"
            id="siteUrl"
            placeholder="站点URL"
            class="input input-bordered w-full peer"
        />
      </div>
    </div>

    <div class="flex gap-3">
      <div class="floating-label flex-1">
        <span>Logo URL</span>
        <input
            v-model="settings.siteLogo"
            type="url"
            id="siteLogo"
            placeholder="Logo URL"
            class="input input-bordered w-full peer"
        />
      </div>
      
      <div class="floating-label flex-none">
        <span class="invisible">预览</span>
        <input
            ref="logoFileInput"
            type="file"
            accept="image/*"
            class="hidden"
            @change="handleLogoUpload"
        />
        <div 
          class="tooltip tooltip-bottom" 
          :data-tip="settings.siteLogo ? '点击更换 Logo' : '点击上传 Logo'"
        >
          <div 
            class="w-10 h-10 rounded-lg border-2 border-base-200 hover:border-primary cursor-pointer overflow-hidden flex items-center justify-center transition-colors bg-base-100"
            @click="logoFileInput?.click()"
          >
            <img 
              v-if="settings.siteLogo" 
              :src="settings.siteLogo" 
              alt="Logo" 
              class="w-full h-full object-cover"
            />
            <PhotoIcon v-else class="w-6 h-6 text-base-content/40"/>
          </div>
        </div>
      </div>
    </div>

    <div class="floating-label">
      <span>站点邮箱</span>
      <input
          v-model="settings.siteEmail"
          type="text"
          id="siteEmail"
          placeholder="站点邮箱"
          class="input input-bordered w-full peer"
      />
    </div>

    <div class="floating-label">
      <span>站点描述</span>
      <textarea
          v-model="settings.siteDescription"
          id="siteDescription"
          placeholder="站点描述"
          class="textarea textarea-bordered peer w-full"
          rows="3"
      ></textarea>
    </div>

    <div class="floating-label">
      <span>关键词（用逗号分隔）</span>
      <input
          v-model="settings.siteKeywords"
          type="text"
          id="siteKeywords"
          placeholder="关键词（用逗号分隔）"
          class="input input-bordered w-full peer"
      />
    </div>

    <div class="form-control">
      <div class="floating-label">
        <span>外部资源链接/Meta标签</span>
        <textarea
            v-model="settings.externalLinks"
            class="textarea textarea-bordered h-40 w-full font-mono text-sm"
            placeholder="请输入外部资源链接，例如：&#10;&lt;link rel=&quot;stylesheet&quot; href=&quot;https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&display=swap&quot;&gt;&#10;&lt;script src=&quot;https://cdn.jsdelivr.net/npm/chart.js&quot;&gt;&lt;/script&gt;&lt;meta name=&quot;author&quot; content=&quot;Your Name&quot;&gt;&#10;&lt;meta name=&quot;robots&quot; content=&quot;index,follow&quot;&gt;"
        ></textarea>
      </div>
      <div class="label">
        <span class="label-text-alt text-base-content/60">外部CSS和JS资源链接，每行一个链接</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue'
import {PhotoIcon} from '@heroicons/vue/24/outline'
import {
  getSiteSettings,
  saveSiteSettings,
  type SiteSettingsConfig,
  uploadImage as uploadImageApi
} from '@admin/utils/adminService'

// 本地数据状态
const settings = ref<SiteSettingsConfig>({
  siteName: '',
  siteLogo: '',
  siteDescription: '',
  siteKeywords: '',
  siteUrl: '',
  siteEmail: '',
  externalLinks: '',
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

// 验证外部链接内容
const validateExternalLinks = (content: string) => {
  if (!content) return true
  
  // 必须以 < 开头
  if (!content.trim().startsWith('<')) {
    return false
  }

  try {
    const parser = new DOMParser()
    const doc = parser.parseFromString(content, 'text/html')
    // 检查解析错误（虽然text/html模式很宽容，但某些严重错误仍可能产生parsererror）
    const parserError = doc.querySelector('parsererror')
    if (parserError) return false

    // 检查是否有非法文本节点直接出现在head或body中（意味着不是标签包裹的内容）
    // 注意：text/html解析器会将head中的非法内容移到body中
    const bodyChildren = Array.from(doc.body.childNodes)
    for (const node of bodyChildren) {
      // 如果是文本节点且不为空白，则说明有非法文本
      if (node.nodeType === Node.TEXT_NODE && node.textContent?.trim()) {
        return false
      }
    }
    
    return true
  } catch (e) {
    return false
  }
}

// 保存设置
const saveSettings = async () => {
  if (settings.value.externalLinks && !validateExternalLinks(settings.value.externalLinks)) {
    alert('外部资源链接格式错误：必须是有效的HTML标签（如 <link>, <script>, <meta>），不能包含纯文本。')
    return
  }

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
    const result = await uploadImageApi(formData)
    if (result.code === 0) {
      settings.value.siteLogo = result.result.url
      // 图片上传成功后自动保存
      await saveSettings()
    } else {
      alert(result.msg || 'Logo上传失败')
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
