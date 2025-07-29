<template>
  <div class="space-y-6">
    <!-- 保存按钮 -->
    <div class="flex justify-end">

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

    <!-- 邮件设置表单 -->
    <div class="space-y-8">
      <!-- SMTP服务器设置 -->
      <div>
        <div class="flex justify-between items-center">
          <h2 class="flex items-center gap-2 mb-6 text-lg font-medium text-base-content">
            <ServerIcon class="w-5 h-5" />
            SMTP服务器设置
          </h2>
          <button @click="handleSave" class="btn btn-primary btn-sm" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-sm"></span>
            <CheckIcon v-else class="w-4 h-4" />
            {{ saving ? '保存中...' : '保存设置' }}
          </button>
        </div>


        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- 启用邮件服务 -->
          <div class="form-control lg:col-span-2">
            <label class="label cursor-pointer">
              <span class="label-text">启用邮件服务</span>
              <input v-model="settings.enableMail" type="checkbox" class="toggle toggle-primary" />
            </label>
          </div>

          <!-- SMTP主机 -->
          <label class="floating-label">
            <span>SMTP主机</span>
            <input
              v-model="settings.smtpHost" :disabled="!settings.enableMail"
              type="text"
              class="input input-bordered w-full peer"
              id="smtpHost"
              placeholder="SMTP主机"
            />
          </label>

          <!-- SMTP端口 -->
          <label class="floating-label">

            <span>SMTP端口</span>
            <input
              v-model="settings.smtpPort" :disabled="!settings.enableMail"
              type="number"
              min="1"
              max="65535"
              class="input input-bordered w-full peer"
              id="smtpPort"
              placeholder="SMTP端口"
            />
          </label>


          <!-- 用户名 -->
          <label class="floating-label">
            <span>用户名（邮箱）</span>
            <input
              v-model="settings.smtpUsername" :disabled="!settings.enableMail"
              type="email"
              class="input input-bordered w-full peer"
              id="smtpUsername"
              placeholder="用户名（邮箱）"
            />
          </label>

          <!-- 密码 -->
          <label class="floating-label">
            <span>密码/授权码</span>
            <input
              v-model="settings.smtpPassword" :disabled="!settings.enableMail"
              type="password"
              class="input input-bordered w-full peer"
              id="smtpPassword"
              placeholder="密码/授权码"
            />
          </label>

          <!-- 是否使用SSL -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <span class="label-text">使用SSL加密</span>
              <input v-model="settings.useSSL" :disabled="!settings.enableMail" type="checkbox" class="toggle toggle-primary" />
            </label>
          </div>

          <label class="floating-label">
          </label>

          <!-- 发件人名称 -->
          <label class="floating-label">
            <span>发件人名称</span>
            <input
              v-model="settings.fromName" :disabled="!settings.enableMail"
              type="text"
              class="input input-bordered w-full peer"
              id="fromName"
              placeholder="发件人名称"
            />
          </label>

          <!-- 发件人邮箱 -->
          <label class="floating-label">
            <span>发件人邮箱</span>
            <input
              v-model="settings.fromEmail" :disabled="!settings.enableMail"
              type="email"
              class="input input-bordered w-full peer"
              id="fromEmail"
              placeholder="发件人邮箱"
            />
          </label>
        </div>
      </div>

      <!-- 分割线 -->
      <div class="divider"></div>

      <!-- 测试邮件 -->
      <div>
        <h2 class="flex items-center gap-2 mb-6 text-lg font-medium text-base-content">
          <PaperAirplaneIcon class="w-5 h-5" />
          测试邮件
        </h2>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div class="space-y-4">
            <!-- 测试邮箱 -->
            <label class="floating-label">
              <span>测试邮箱</span>
              <input
                v-model="testEmail"
                type="email"
                class="input input-bordered w-full peer"
                id="testEmail"
                placeholder="测试邮箱:example@example.com"
              />
            </label>

            <!-- 发送测试邮件按钮 -->
            <button @click="sendTestEmail" class="btn btn-sm btn-outline btn-primary" :disabled="sendingTest || !testEmail">
              <span v-if="sendingTest" class="loading loading-spinner loading-sm"></span>
              <PaperAirplaneIcon v-else class="w-4 h-4" />
              发送测试邮件
            </button>
          </div>

          <div class="space-y-2">
            <h3 class="font-medium text-base-content">常见邮箱配置</h3>
            <div class="text-sm text-base-content/70 space-y-1">
              <div><strong>Gmail:</strong> smtp.gmail.com:587 (不使用SSL)</div>
              <div><strong>QQ邮箱:</strong> smtp.qq.com:587 (不使用SSL)</div>
              <div><strong>163邮箱:</strong> smtp.163.com:25 (不使用SSL)</div>
              <div><strong>126邮箱:</strong> smtp.126.com:25 (不使用SSL)</div>
              <div><strong>Outlook:</strong> smtp-mail.outlook.com:587 (不使用SSL)</div>
              <div><strong>Gmail SSL:</strong> smtp.gmail.com:465 (使用SSL)</div>
            </div>
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
  ServerIcon,
  DocumentTextIcon,
  BellIcon,
  PaperAirplaneIcon
} from '@heroicons/vue/24/outline'
import { getMailSettings, saveMailSettings, testMailConnection, type MailSettingsConfig } from '../../utils/adminService'

// 设置数据
const settings = ref({
  // SMTP服务器设置
  enableMail: true,
  smtpHost: 'smtp.gmail.com',
  smtpPort: 587,
  useSSL: false,
  smtpUsername: 'noreply@example.com',
  smtpPassword: '',
  fromName: 'GooseForum',
  fromEmail: 'noreply@example.com',

})

// 状态变量
const saving = ref(false)
const sendingTest = ref(false)
const showSuccessAlert = ref(false)
const showErrorAlert = ref(false)
const errorMessage = ref('')
const testEmail = ref('')

// 保存设置
const handleSave = async () => {
  saving.value = true
  try {
    const response = await saveMailSettings(settings.value as MailSettingsConfig)

    if (response.code === 0) {
      showSuccess('邮件设置保存成功！')
    } else {
      showError(response.msg || '保存设置失败')
    }
  } catch (error) {
    console.error('保存邮件设置失败:', error)
    showError('保存设置失败，请重试')
  } finally {
    saving.value = false
  }
}

// 发送测试邮件
const sendTestEmail = async () => {
  if (!testEmail.value) {
    showError('请输入测试邮箱')
    return
  }

  sendingTest.value = true
  try {
    const response = await testMailConnection(settings.value as MailSettingsConfig, testEmail.value)

    if (response.code === 0) {
      showSuccess(response.msg || '测试邮件发送成功')
    } else {
      showError(response.msg || '测试邮件发送失败')
    }
  } catch (error) {
    console.error('发送测试邮件失败:', error)
    showError('发送测试邮件失败，请检查SMTP配置')
  } finally {
    sendingTest.value = false
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

// 加载邮件设置
const loadMailSettings = async () => {
  try {
    const response = await getMailSettings()
    if (response.code === 0) {
      Object.assign(settings.value, response.result)
    } else {
      console.error('加载邮件设置失败:', response.msg)
    }
  } catch (error) {
    console.error('加载邮件设置失败:', error)
  }
}

// 组件挂载时加载设置
onMounted(() => {
  loadMailSettings()
})
</script>

<style scoped>
/* 自定义样式 */
</style>
