<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-base-content">系统设置</h1>
        <p class="text-base-content/70 mt-1">配置和管理系统参数</p>
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

    <!-- 设置表单 -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- 基本设置 -->
      <div class="lg:col-span-2 space-y-6">
        <!-- 网站基本信息 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <div class="flex justify-between">
              <h2 class="card-title flex items-center gap-2">
                <GlobeAltIcon class="w-5 h-5" />
                网站基本信息
              </h2>
               <button class="btn btn-primary ml-auto btn-sm">保存</button>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="form-control">
                <label class="label">
                  <span class="label-text">网站名称</span>
                </label>
                <input v-model="settings.siteName" type="text" placeholder="请输入网站名称" class="input input-bordered" />
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">网站域名</span>
                </label>
                <input v-model="settings.siteUrl" type="url" placeholder="https://example.com"
                  class="input input-bordered" />
              </div>
              <div class="form-control md:col-span-2">
                <fieldset class="fieldset">
                  <legend class="fieldset-legend">Your bio</legend>
                  <textarea class="textarea h-24" placeholder="Bio" v-model="settings.siteDescription"></textarea>
                  <div class="label">Optional</div>
                </fieldset>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">网站关键词</span>
                </label>
                <input v-model="settings.siteKeywords" type="text" placeholder="关键词1,关键词2,关键词3"
                  class="input input-bordered" />
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">ICP备案号</span>
                </label>
                <input v-model="settings.icpNumber" type="text" placeholder="请输入ICP备案号" class="input input-bordered" />
              </div>
            </div>
          </div>
        </div>

        <!-- 用户设置 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2">
              <UsersIcon class="w-5 h-5" />
              用户设置
            </h2>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="form-control">
                <label class="label cursor-pointer">
                  <span class="label-text">允许用户注册</span>
                  <input v-model="settings.allowRegistration" type="checkbox" class="toggle toggle-primary" />
                </label>
              </div>
              <div class="form-control">
                <label class="label cursor-pointer">
                  <span class="label-text">需要邮箱验证</span>
                  <input v-model="settings.requireEmailVerification" type="checkbox" class="toggle toggle-primary" />
                </label>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">默认用户角色</span>
                </label>
                <select v-model="settings.defaultUserRole" class="select select-bordered">
                  <option value="user">普通用户</option>
                  <option value="vip">VIP用户</option>
                  <option value="moderator">版主</option>
                </select>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">用户名最小长度</span>
                </label>
                <input v-model.number="settings.minUsernameLength" type="number" min="2" max="20"
                  class="input input-bordered" />
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">密码最小长度</span>
                </label>
                <input v-model.number="settings.minPasswordLength" type="number" min="6" max="50"
                  class="input input-bordered" />
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">会话过期时间(小时)</span>
                </label>
                <input v-model.number="settings.sessionTimeout" type="number" min="1" max="720"
                  class="input input-bordered" />
              </div>
            </div>
          </div>
        </div>

        <!-- 内容设置 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2">
              <DocumentTextIcon class="w-5 h-5" />
              内容设置
            </h2>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="form-control">
                <label class="label cursor-pointer">
                  <span class="label-text">需要审核新帖子</span>
                  <input v-model="settings.requirePostApproval" type="checkbox" class="toggle toggle-primary" />
                </label>
              </div>
              <div class="form-control">
                <label class="label cursor-pointer">
                  <span class="label-text">需要审核评论</span>
                  <input v-model="settings.requireCommentApproval" type="checkbox" class="toggle toggle-primary" />
                </label>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">每页帖子数量</span>
                </label>
                <input v-model.number="settings.postsPerPage" type="number" min="5" max="100"
                  class="input input-bordered" />
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">帖子标题最大长度</span>
                </label>
                <input v-model.number="settings.maxTitleLength" type="number" min="10" max="200"
                  class="input input-bordered" />
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">帖子内容最大长度</span>
                </label>
                <input v-model.number="settings.maxContentLength" type="number" min="100" max="50000"
                  class="input input-bordered" />
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">允许的文件类型</span>
                </label>
                <input v-model="settings.allowedFileTypes" type="text" placeholder="jpg,png,gif,pdf,doc"
                  class="input input-bordered" />
              </div>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">

              <div class="form-control">
                <label class="label">
                  <span class="label-text">最大文件大小(MB)</span>
                </label>
                <input v-model.number="settings.maxFileSize" type="number" min="1" max="100"
                  class="input input-bordered" />
              </div>
            </div>
          </div>
        </div>

        <!-- 邮件设置 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2">
              <EnvelopeIcon class="w-5 h-5" />
              邮件设置
            </h2>
            <div class="space-y-4">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="form-control">
                  <label class="label cursor-pointer">
                    <span class="label-text">启用邮件服务</span>
                    <input v-model="settings.enableEmail" type="checkbox" class="toggle toggle-primary" />
                  </label>
                </div>
                <div class="form-control">
                  <label class="label cursor-pointer">
                    <span class="label-text">启用SSL/TLS</span>
                    <input v-model="settings.smtpSSL" type="checkbox" class="toggle toggle-primary"
                      :disabled="!settings.enableEmail" />
                  </label>
                </div>
              </div>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="form-control">
                  <label class="label">
                    <span class="label-text">SMTP服务器</span>
                  </label>
                  <input v-model="settings.smtpHost" type="text" placeholder="smtp.example.com"
                    class="input input-bordered" :disabled="!settings.enableEmail" />
                </div>
                <div class="form-control">
                  <label class="label">
                    <span class="label-text">SMTP端口</span>
                  </label>
                  <input v-model.number="settings.smtpPort" type="number" placeholder="587" class="input input-bordered"
                    :disabled="!settings.enableEmail" />
                </div>
                <div class="form-control">
                  <label class="label">
                    <span class="label-text">发件人邮箱</span>
                  </label>
                  <input v-model="settings.smtpUsername" type="email" placeholder="noreply@example.com"
                    class="input input-bordered" :disabled="!settings.enableEmail" />
                </div>
                <div class="form-control">
                  <label class="label">
                    <span class="label-text">邮箱密码</span>
                  </label>
                  <input v-model="settings.smtpPassword" type="password" placeholder="请输入邮箱密码"
                    class="input input-bordered" :disabled="!settings.enableEmail" />
                </div>
              </div>
            </div>
            <div class="mt-4">
              <button @click="testEmail" class="btn btn-outline btn-info btn-sm"
                :disabled="!settings.enableEmail || testingEmail">
                <span v-if="testingEmail" class="loading loading-spinner loading-sm"></span>
                <PaperAirplaneIcon v-else class="w-4 h-4" />
                测试邮件发送
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 侧边栏设置 -->
      <div class="space-y-6">
        <!-- 系统状态 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2">
              <ServerIcon class="w-5 h-5" />
              系统状态
            </h2>
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <span class="text-sm">服务器状态</span>
                <div class="badge badge-success gap-1">
                  <div class="w-2 h-2 bg-success rounded-full"></div>
                  正常
                </div>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm">数据库</span>
                <div class="badge badge-success gap-1">
                  <div class="w-2 h-2 bg-success rounded-full"></div>
                  连接正常
                </div>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm">缓存服务</span>
                <div class="badge badge-warning gap-1">
                  <div class="w-2 h-2 bg-warning rounded-full"></div>
                  未启用
                </div>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm">邮件服务</span>
                <div :class="['badge gap-1', settings.enableEmail ? 'badge-success' : 'badge-error']">
                  <div :class="['w-2 h-2 rounded-full', settings.enableEmail ? 'bg-success' : 'bg-error']"></div>
                  {{ settings.enableEmail ? '已启用' : '未启用' }}
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 安全设置 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2">
              <ShieldCheckIcon class="w-5 h-5" />
              安全设置
            </h2>
            <div class="space-y-4">
              <div class="form-control">
                <label class="label cursor-pointer">
                  <span class="label-text">启用验证码</span>
                  <input v-model="settings.enableCaptcha" type="checkbox" class="toggle toggle-primary" />
                </label>
              </div>
              <div class="form-control">
                <label class="label cursor-pointer">
                  <span class="label-text">启用IP限制</span>
                  <input v-model="settings.enableIPLimit" type="checkbox" class="toggle toggle-primary" />
                </label>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">登录失败限制</span>
                </label>
                <input v-model.number="settings.maxLoginAttempts" type="number" min="3" max="10"
                  class="input input-bordered input-sm" />
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">锁定时间(分钟)</span>
                </label>
                <input v-model.number="settings.lockoutDuration" type="number" min="5" max="1440"
                  class="input input-bordered input-sm" />
              </div>
            </div>
          </div>
        </div>

        <!-- 缓存设置 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2">
              <BoltIcon class="w-5 h-5" />
              缓存设置
            </h2>
            <div class="space-y-4">
              <div class="form-control">
                <label class="label cursor-pointer">
                  <span class="label-text">启用页面缓存</span>
                  <input v-model="settings.enablePageCache" type="checkbox" class="toggle toggle-primary" />
                </label>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">缓存过期时间(分钟)</span>
                </label>
                <input v-model.number="settings.cacheExpiration" type="number" min="1" max="1440"
                  class="input input-bordered input-sm" :disabled="!settings.enablePageCache" />
              </div>
              <div class="mt-4">
                <button @click="clearCache" class="btn btn-outline btn-warning btn-sm w-full" :disabled="clearingCache">
                  <span v-if="clearingCache" class="loading loading-spinner loading-sm"></span>
                  <TrashIcon v-else class="w-4 h-4" />
                  清空缓存
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 备份设置 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2">
              <CloudArrowUpIcon class="w-5 h-5" />
              备份设置
            </h2>
            <div class="space-y-4">
              <div class="form-control">
                <label class="label cursor-pointer">
                  <span class="label-text">自动备份</span>
                  <input v-model="settings.enableAutoBackup" type="checkbox" class="toggle toggle-primary" />
                </label>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">备份频率</span>
                </label>
                <select v-model="settings.backupFrequency" class="select select-bordered select-sm"
                  :disabled="!settings.enableAutoBackup">
                  <option value="daily">每日</option>
                  <option value="weekly">每周</option>
                  <option value="monthly">每月</option>
                </select>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">保留备份数量</span>
                </label>
                <input v-model.number="settings.backupRetention" type="number" min="1" max="30"
                  class="input input-bordered input-sm" :disabled="!settings.enableAutoBackup" />
              </div>
              <div class="mt-4">
                <button @click="createBackup" class="btn btn-outline btn-info btn-sm w-full" :disabled="creatingBackup">
                  <span v-if="creatingBackup" class="loading loading-spinner loading-sm"></span>
                  <CloudArrowUpIcon v-else class="w-4 h-4" />
                  立即备份
                </button>
              </div>
            </div>
          </div>
        </div>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  GlobeAltIcon,
  UsersIcon,
  DocumentTextIcon,
  EnvelopeIcon,
  ServerIcon,
  ShieldCheckIcon,
  BoltIcon,
  CloudArrowUpIcon,
  CheckIcon,
  ArrowPathIcon,
  PaperAirplaneIcon,
  TrashIcon,
  CheckCircleIcon,
  XCircleIcon
} from '@heroicons/vue/24/outline'
import { api } from '../utils/axiosInstance'

// 设置数据
const settings = ref({
  // 网站基本信息
  siteName: 'GooseForum',
  siteUrl: 'https://localhost:3000',
  siteDescription: '一个现代化的论坛系统',
  siteKeywords: 'forum,discussion,community',
  icpNumber: '',

  // 用户设置
  allowRegistration: true,
  requireEmailVerification: false,
  defaultUserRole: 'user',
  minUsernameLength: 3,
  minPasswordLength: 6,
  sessionTimeout: 168, // 7天

  // 内容设置
  requirePostApproval: false,
  requireCommentApproval: false,
  postsPerPage: 20,
  maxTitleLength: 100,
  maxContentLength: 10000,
  allowedFileTypes: 'jpg,jpeg,png,gif,pdf,doc,docx',
  maxFileSize: 10,
  enableRichEditor: true,

  // 邮件设置
  enableEmail: false,
  smtpHost: '',
  smtpPort: 587,
  smtpUsername: '',
  smtpPassword: '',
  smtpSSL: true,

  // 安全设置
  enableCaptcha: true,
  enableIPLimit: false,
  maxLoginAttempts: 5,
  lockoutDuration: 30,

  // 缓存设置
  enablePageCache: false,
  cacheExpiration: 60,

  // 备份设置
  enableAutoBackup: true,
  backupFrequency: 'daily',
  backupRetention: 7
})

// 状态变量
const saving = ref(false)
const testingEmail = ref(false)
const clearingCache = ref(false)
const creatingBackup = ref(false)
const showSuccessAlert = ref(false)
const showErrorAlert = ref(false)
const errorMessage = ref('')

// 加载设置
const loadSettings = async () => {
  try {
    const response = await api.get('/admin/settings')
    if (response.data.success) {
      Object.assign(settings.value, response.data.data)
    }
  } catch (error) {
    console.error('加载设置失败:', error)
    showError('加载设置失败')
  }
}

// 保存设置
const saveSettings = async () => {
  saving.value = true
  try {
    const response = await api.post('/admin/settings', settings.value)
    if (response.data.success) {
      showSuccess()
    } else {
      showError(response.data.message || '保存失败')
    }
  } catch (error) {
    console.error('保存设置失败:', error)
    showError('保存设置失败')
  } finally {
    saving.value = false
  }
}

// 重置为默认值
const resetToDefaults = () => {
  if (confirm('确定要重置为默认设置吗？这将覆盖当前所有设置。')) {
    // 重置为默认值
    Object.assign(settings.value, {
      siteName: 'GooseForum',
      siteUrl: 'https://localhost:3000',
      siteDescription: '一个现代化的论坛系统',
      siteKeywords: 'forum,discussion,community',
      icpNumber: '',
      allowRegistration: true,
      requireEmailVerification: false,
      defaultUserRole: 'user',
      minUsernameLength: 3,
      minPasswordLength: 6,
      sessionTimeout: 168,
      requirePostApproval: false,
      requireCommentApproval: false,
      postsPerPage: 20,
      maxTitleLength: 100,
      maxContentLength: 10000,
      allowedFileTypes: 'jpg,jpeg,png,gif,pdf,doc,docx',
      maxFileSize: 10,
      enableRichEditor: true,
      enableEmail: false,
      smtpHost: '',
      smtpPort: 587,
      smtpUsername: '',
      smtpPassword: '',
      smtpSSL: true,
      enableCaptcha: true,
      enableIPLimit: false,
      maxLoginAttempts: 5,
      lockoutDuration: 30,
      enablePageCache: false,
      cacheExpiration: 60,
      enableAutoBackup: true,
      backupFrequency: 'daily',
      backupRetention: 7
    })
  }
}

// 测试邮件发送
const testEmail = async () => {
  testingEmail.value = true
  try {
    const response = await api.post('/admin/test-email', {
      smtpHost: settings.value.smtpHost,
      smtpPort: settings.value.smtpPort,
      smtpUsername: settings.value.smtpUsername,
      smtpPassword: settings.value.smtpPassword,
      smtpSSL: settings.value.smtpSSL
    })
    if (response.data.success) {
      showSuccess('测试邮件发送成功！')
    } else {
      showError(response.data.message || '测试邮件发送失败')
    }
  } catch (error) {
    console.error('测试邮件失败:', error)
    showError('测试邮件发送失败')
  } finally {
    testingEmail.value = false
  }
}

// 清空缓存
const clearCache = async () => {
  clearingCache.value = true
  try {
    const response = await api.post('/admin/clear-cache')
    if (response.data.success) {
      showSuccess('缓存清空成功！')
    } else {
      showError(response.data.message || '清空缓存失败')
    }
  } catch (error) {
    console.error('清空缓存失败:', error)
    showError('清空缓存失败')
  } finally {
    clearingCache.value = false
  }
}

// 创建备份
const createBackup = async () => {
  creatingBackup.value = true
  try {
    const response = await api.post('/admin/create-backup')
    if (response.data.success) {
      showSuccess('备份创建成功！')
    } else {
      showError(response.data.message || '创建备份失败')
    }
  } catch (error) {
    console.error('创建备份失败:', error)
    showError('创建备份失败')
  } finally {
    creatingBackup.value = false
  }
}

// 显示成功提示
const showSuccess = (message = '操作成功！') => {
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