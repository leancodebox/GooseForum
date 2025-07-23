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
              v-model="settings.smtpHost"
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
              v-model="settings.smtpPort"
              type="number"
              min="1"
              max="65535"
              class="input input-bordered w-full peer"
              id="smtpPort"
              placeholder="SMTP端口"
            />
          </label>

          <!-- 加密方式 -->
          <label class="floating-label">
            <span>加密方式</span>
            <select
              v-model="settings.smtpEncryption"
              class="select select-bordered w-full peer"
              id="smtpEncryption"
            >
              <option value="none">无加密</option>
              <option value="tls">TLS</option>
              <option value="ssl">SSL</option>
            </select>
          </label>

          <!-- 用户名 -->
          <label class="floating-label">
            <span>用户名（邮箱）</span>
            <input
              v-model="settings.smtpUsername"
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
              v-model="settings.smtpPassword"
              type="password"
              class="input input-bordered w-full peer"
              id="smtpPassword"
              placeholder="密码/授权码"
            />
          </label>

          <!-- 发件人名称 -->
          <label class="floating-label">
            <span>发件人名称</span>
            <input
              v-model="settings.fromName"
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
              v-model="settings.fromEmail"
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

      <!-- 通知设置 -->
      <div>
        <h2 class="flex items-center gap-2 mb-6 text-lg font-medium text-base-content">
          <BellIcon class="w-5 h-5" />
          通知设置
        </h2>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <div class="space-y-4">
            <h3 class="font-medium text-base-content">用户通知</h3>

            <!-- 新用户注册通知管理员 -->
            <div class="form-control">
              <label class="label cursor-pointer">
                <span class="label-text">新用户注册通知管理员</span>
                <input v-model="settings.notifyAdminOnRegistration" type="checkbox" class="toggle toggle-primary" />
              </label>
            </div>

            <!-- 新帖子通知管理员 -->
            <div class="form-control">
              <label class="label cursor-pointer">
                <span class="label-text">新帖子通知管理员</span>
                <input v-model="settings.notifyAdminOnNewPost" type="checkbox" class="toggle toggle-primary" />
              </label>
            </div>

            <!-- 新评论通知作者 -->
            <div class="form-control">
              <label class="label cursor-pointer">
                <span class="label-text">新评论通知作者</span>
                <input v-model="settings.notifyAuthorOnComment" type="checkbox" class="toggle toggle-primary" />
              </label>
            </div>

            <!-- 私信通知 -->
            <div class="form-control">
              <label class="label cursor-pointer">
                <span class="label-text">私信通知</span>
                <input v-model="settings.notifyOnPrivateMessage" type="checkbox" class="toggle toggle-primary" />
              </label>
            </div>
          </div>

          <div class="space-y-4">
            <h3 class="font-medium text-base-content">系统通知</h3>

            <!-- 系统维护通知 -->
            <div class="form-control">
              <label class="label cursor-pointer">
                <span class="label-text">系统维护通知</span>
                <input v-model="settings.notifyOnMaintenance" type="checkbox" class="toggle toggle-primary" />
              </label>
            </div>

            <!-- 安全警告通知 -->
            <div class="form-control">
              <label class="label cursor-pointer">
                <span class="label-text">安全警告通知</span>
                <input v-model="settings.notifyOnSecurityAlert" type="checkbox" class="toggle toggle-primary" />
              </label>
            </div>

            <!-- 管理员邮箱 -->
            <label class="floating-label">
              <span >管理员邮箱</span>
              <input
                v-model="settings.adminEmail"
                type="email"
                class="input input-bordered w-full peer"
                id="adminEmail"
                placeholder=" "
              />
            </label>

            <!-- 邮件发送频率限制 -->
            <label class="floating-label">
              <span>邮件发送频率限制</span>
              <select
                v-model="settings.emailRateLimit"
                class="select select-bordered w-full peer"
                id="emailRateLimit"
              >
                <option value="none">无限制</option>
                <option value="1min">每分钟1封</option>
                <option value="5min">每5分钟1封</option>
                <option value="15min">每15分钟1封</option>
                <option value="1hour">每小时1封</option>
              </select>
            </label>
          </div>
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
            <button @click="sendTestEmail" class="btn btn-outline btn-primary" :disabled="sendingTest || !testEmail">
              <span v-if="sendingTest" class="loading loading-spinner loading-sm"></span>
              <PaperAirplaneIcon v-else class="w-4 h-4" />
              发送测试邮件
            </button>
          </div>

          <div class="space-y-2">
            <h3 class="font-medium text-base-content">常见邮箱配置</h3>
            <div class="text-sm text-base-content/70 space-y-1">
              <div><strong>Gmail:</strong> smtp.gmail.com:587 (TLS)</div>
              <div><strong>QQ邮箱:</strong> smtp.qq.com:587 (TLS)</div>
              <div><strong>163邮箱:</strong> smtp.163.com:25 (无加密)</div>
              <div><strong>126邮箱:</strong> smtp.126.com:25 (无加密)</div>
              <div><strong>Outlook:</strong> smtp-mail.outlook.com:587 (TLS)</div>
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

// 设置数据
const settings = ref({
  // SMTP服务器设置
  enableMail: true,
  smtpHost: 'smtp.gmail.com',
  smtpPort: 587,
  smtpEncryption: 'tls',
  smtpUsername: 'noreply@example.com',
  smtpPassword: '',
  fromName: 'GooseForum',
  fromEmail: 'noreply@example.com',

  // 邮件模板设置
  verificationSubject: '请验证您的邮箱 - {site_name}',
  verificationContent: '亲爱的 {username}，\n\n感谢您注册 {site_name}！请点击以下链接验证您的邮箱：\n\n{verification_link}\n\n如果您没有注册账户，请忽略此邮件。\n\n祝好！\n{site_name} 团队',
  resetPasswordSubject: '重置您的密码 - {site_name}',
  resetPasswordContent: '亲爱的 {username}，\n\n我们收到了重置您密码的请求。请点击以下链接重置密码：\n\n{reset_link}\n\n如果您没有请求重置密码，请忽略此邮件。\n\n祝好！\n{site_name} 团队',

  // 通知设置
  notifyAdminOnRegistration: true,
  notifyAdminOnNewPost: false,
  notifyAuthorOnComment: true,
  notifyOnPrivateMessage: true,
  notifyOnMaintenance: true,
  notifyOnSecurityAlert: true,
  adminEmail: 'admin@example.com',
  emailRateLimit: '5min'
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
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    showSuccess()
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
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 2000))
    showSuccess('测试邮件发送成功！请检查您的邮箱。')
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

// 组件挂载时加载设置
onMounted(() => {
  // 这里可以加载实际的设置数据
  console.log('邮件设置页面已加载')
})
</script>

<style scoped>
/* 自定义样式 */
</style>
