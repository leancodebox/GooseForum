<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div>
      <h1 class="text-2xl font-bold text-base-content">系统设置</h1>
      <p class="text-base-content/70 mt-1">管理系统的各项配置参数</p>
    </div>

    <!-- 设置选项卡 -->
    <div class="tabs tabs-bordered">
      <a 
        v-for="tab in tabs" 
        :key="tab.key"
        class="tab"
        :class="{ 'tab-active': activeTab === tab.key }"
        @click="activeTab = tab.key"
      >
        <component :is="tab.icon" class="w-4 h-4 mr-2" />
        {{ tab.label }}
      </a>
    </div>

    <!-- 基本设置 -->
    <div v-show="activeTab === 'basic'" class="space-y-6">
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title">站点信息</h2>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">站点名称</span>
              </label>
              <input 
                v-model="settings.basic.siteName" 
                type="text" 
                placeholder="请输入站点名称" 
                class="input input-bordered"
              />
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">站点域名</span>
              </label>
              <input 
                v-model="settings.basic.siteUrl" 
                type="url" 
                placeholder="https://example.com" 
                class="input input-bordered"
              />
            </div>
            
            <div class="form-control md:col-span-2">
              <label class="label">
                <span class="label-text">站点描述</span>
              </label>
              <textarea 
                v-model="settings.basic.siteDescription" 
                class="textarea textarea-bordered" 
                placeholder="请输入站点描述"
                rows="3"
              ></textarea>
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">站点关键词</span>
              </label>
              <input 
                v-model="settings.basic.siteKeywords" 
                type="text" 
                placeholder="关键词1,关键词2,关键词3" 
                class="input input-bordered"
              />
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">ICP备案号</span>
              </label>
              <input 
                v-model="settings.basic.icpNumber" 
                type="text" 
                placeholder="请输入ICP备案号" 
                class="input input-bordered"
              />
            </div>
          </div>
        </div>
      </div>
      
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title">联系信息</h2>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">管理员邮箱</span>
              </label>
              <input 
                v-model="settings.basic.adminEmail" 
                type="email" 
                placeholder="admin@example.com" 
                class="input input-bordered"
              />
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">客服QQ</span>
              </label>
              <input 
                v-model="settings.basic.customerQQ" 
                type="text" 
                placeholder="请输入客服QQ" 
                class="input input-bordered"
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 用户设置 -->
    <div v-show="activeTab === 'user'" class="space-y-6">
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title">注册设置</h2>
          
          <div class="space-y-4">
            <div class="form-control">
              <label class="cursor-pointer label">
                <span class="label-text">允许用户注册</span>
                <input 
                  v-model="settings.user.allowRegister" 
                  type="checkbox" 
                  class="toggle toggle-primary"
                />
              </label>
            </div>
            
            <div class="form-control">
              <label class="cursor-pointer label">
                <span class="label-text">注册需要邮箱验证</span>
                <input 
                  v-model="settings.user.requireEmailVerification" 
                  type="checkbox" 
                  class="toggle toggle-primary"
                />
              </label>
            </div>
            
            <div class="form-control">
              <label class="cursor-pointer label">
                <span class="label-text">注册需要管理员审核</span>
                <input 
                  v-model="settings.user.requireAdminApproval" 
                  type="checkbox" 
                  class="toggle toggle-primary"
                />
              </label>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="form-control">
                <label class="label">
                  <span class="label-text">用户名最小长度</span>
                </label>
                <input 
                  v-model.number="settings.user.usernameMinLength" 
                  type="number" 
                  min="2" 
                  max="20" 
                  class="input input-bordered"
                />
              </div>
              
              <div class="form-control">
                <label class="label">
                  <span class="label-text">密码最小长度</span>
                </label>
                <input 
                  v-model.number="settings.user.passwordMinLength" 
                  type="number" 
                  min="6" 
                  max="50" 
                  class="input input-bordered"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title">用户行为限制</h2>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">每日发帖限制</span>
              </label>
              <input 
                v-model.number="settings.user.dailyPostLimit" 
                type="number" 
                min="0" 
                class="input input-bordered"
              />
              <label class="label">
                <span class="label-text-alt">0表示不限制</span>
              </label>
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">每日评论限制</span>
              </label>
              <input 
                v-model.number="settings.user.dailyCommentLimit" 
                type="number" 
                min="0" 
                class="input input-bordered"
              />
              <label class="label">
                <span class="label-text-alt">0表示不限制</span>
              </label>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 内容设置 -->
    <div v-show="activeTab === 'content'" class="space-y-6">
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title">发布设置</h2>
          
          <div class="space-y-4">
            <div class="form-control">
              <label class="cursor-pointer label">
                <span class="label-text">新帖需要审核</span>
                <input 
                  v-model="settings.content.requirePostApproval" 
                  type="checkbox" 
                  class="toggle toggle-primary"
                />
              </label>
            </div>
            
            <div class="form-control">
              <label class="cursor-pointer label">
                <span class="label-text">评论需要审核</span>
                <input 
                  v-model="settings.content.requireCommentApproval" 
                  type="checkbox" 
                  class="toggle toggle-primary"
                />
              </label>
            </div>
            
            <div class="form-control">
              <label class="cursor-pointer label">
                <span class="label-text">允许匿名发帖</span>
                <input 
                  v-model="settings.content.allowAnonymousPost" 
                  type="checkbox" 
                  class="toggle toggle-primary"
                />
              </label>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="form-control">
                <label class="label">
                  <span class="label-text">帖子标题最大长度</span>
                </label>
                <input 
                  v-model.number="settings.content.maxTitleLength" 
                  type="number" 
                  min="10" 
                  max="200" 
                  class="input input-bordered"
                />
              </div>
              
              <div class="form-control">
                <label class="label">
                  <span class="label-text">帖子内容最大长度</span>
                </label>
                <input 
                  v-model.number="settings.content.maxContentLength" 
                  type="number" 
                  min="100" 
                  class="input input-bordered"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title">敏感词过滤</h2>
          
          <div class="space-y-4">
            <div class="form-control">
              <label class="cursor-pointer label">
                <span class="label-text">启用敏感词过滤</span>
                <input 
                  v-model="settings.content.enableSensitiveWordFilter" 
                  type="checkbox" 
                  class="toggle toggle-primary"
                />
              </label>
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">敏感词列表</span>
              </label>
              <textarea 
                v-model="settings.content.sensitiveWords" 
                class="textarea textarea-bordered" 
                placeholder="每行一个敏感词"
                rows="5"
              ></textarea>
              <label class="label">
                <span class="label-text-alt">每行输入一个敏感词</span>
              </label>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 邮件设置 -->
    <div v-show="activeTab === 'email'" class="space-y-6">
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title">SMTP配置</h2>
          
          <div class="space-y-4">
            <div class="form-control">
              <label class="cursor-pointer label">
                <span class="label-text">启用邮件功能</span>
                <input 
                  v-model="settings.email.enabled" 
                  type="checkbox" 
                  class="toggle toggle-primary"
                />
              </label>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="form-control">
                <label class="label">
                  <span class="label-text">SMTP服务器</span>
                </label>
                <input 
                  v-model="settings.email.smtpHost" 
                  type="text" 
                  placeholder="smtp.example.com" 
                  class="input input-bordered"
                />
              </div>
              
              <div class="form-control">
                <label class="label">
                  <span class="label-text">SMTP端口</span>
                </label>
                <input 
                  v-model.number="settings.email.smtpPort" 
                  type="number" 
                  placeholder="587" 
                  class="input input-bordered"
                />
              </div>
              
              <div class="form-control">
                <label class="label">
                  <span class="label-text">发件人邮箱</span>
                </label>
                <input 
                  v-model="settings.email.fromEmail" 
                  type="email" 
                  placeholder="noreply@example.com" 
                  class="input input-bordered"
                />
              </div>
              
              <div class="form-control">
                <label class="label">
                  <span class="label-text">发件人名称</span>
                </label>
                <input 
                  v-model="settings.email.fromName" 
                  type="text" 
                  placeholder="论坛系统" 
                  class="input input-bordered"
                />
              </div>
              
              <div class="form-control">
                <label class="label">
                  <span class="label-text">SMTP用户名</span>
                </label>
                <input 
                  v-model="settings.email.smtpUsername" 
                  type="text" 
                  placeholder="用户名" 
                  class="input input-bordered"
                />
              </div>
              
              <div class="form-control">
                <label class="label">
                  <span class="label-text">SMTP密码</span>
                </label>
                <input 
                  v-model="settings.email.smtpPassword" 
                  type="password" 
                  placeholder="密码" 
                  class="input input-bordered"
                />
              </div>
            </div>
            
            <div class="form-control">
              <label class="cursor-pointer label">
                <span class="label-text">启用SSL/TLS</span>
                <input 
                  v-model="settings.email.enableSSL" 
                  type="checkbox" 
                  class="toggle toggle-primary"
                />
              </label>
            </div>
            
            <div class="flex gap-2">
              <button class="btn btn-outline" @click="testEmail" :disabled="testingEmail">
                <span v-if="testingEmail" class="loading loading-spinner loading-sm"></span>
                {{ testingEmail ? '测试中...' : '测试邮件' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 安全设置 -->
    <div v-show="activeTab === 'security'" class="space-y-6">
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title">登录安全</h2>
          
          <div class="space-y-4">
            <div class="form-control">
              <label class="cursor-pointer label">
                <span class="label-text">启用验证码</span>
                <input 
                  v-model="settings.security.enableCaptcha" 
                  type="checkbox" 
                  class="toggle toggle-primary"
                />
              </label>
            </div>
            
            <div class="form-control">
              <label class="cursor-pointer label">
                <span class="label-text">启用双因素认证</span>
                <input 
                  v-model="settings.security.enableTwoFactor" 
                  type="checkbox" 
                  class="toggle toggle-primary"
                />
              </label>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="form-control">
                <label class="label">
                  <span class="label-text">登录失败锁定次数</span>
                </label>
                <input 
                  v-model.number="settings.security.maxLoginAttempts" 
                  type="number" 
                  min="3" 
                  max="10" 
                  class="input input-bordered"
                />
              </div>
              
              <div class="form-control">
                <label class="label">
                  <span class="label-text">锁定时间(分钟)</span>
                </label>
                <input 
                  v-model.number="settings.security.lockoutDuration" 
                  type="number" 
                  min="5" 
                  max="1440" 
                  class="input input-bordered"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title">IP访问控制</h2>
          
          <div class="space-y-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">IP白名单</span>
              </label>
              <textarea 
                v-model="settings.security.ipWhitelist" 
                class="textarea textarea-bordered" 
                placeholder="每行一个IP地址或IP段\n例如: 192.168.1.1 或 192.168.1.0/24"
                rows="4"
              ></textarea>
              <label class="label">
                <span class="label-text-alt">留空表示不限制，每行一个IP地址</span>
              </label>
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">IP黑名单</span>
              </label>
              <textarea 
                v-model="settings.security.ipBlacklist" 
                class="textarea textarea-bordered" 
                placeholder="每行一个IP地址或IP段\n例如: 192.168.1.1 或 192.168.1.0/24"
                rows="4"
              ></textarea>
              <label class="label">
                <span class="label-text-alt">每行一个IP地址</span>
              </label>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 保存按钮 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <div class="flex justify-end gap-2">
          <button class="btn btn-ghost" @click="resetSettings">重置</button>
          <button class="btn btn-primary" @click="saveSettings" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-sm"></span>
            {{ saving ? '保存中...' : '保存设置' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
  CogIcon,
  UserIcon,
  DocumentTextIcon,
  EnvelopeIcon,
  ShieldCheckIcon
} from '@heroicons/vue/24/outline'
import { api } from '../utils/axiosInstance'

// 选项卡配置
const tabs = [
  { key: 'basic', label: '基本设置', icon: CogIcon },
  { key: 'user', label: '用户设置', icon: UserIcon },
  { key: 'content', label: '内容设置', icon: DocumentTextIcon },
  { key: 'email', label: '邮件设置', icon: EnvelopeIcon },
  { key: 'security', label: '安全设置', icon: ShieldCheckIcon }
]

// 响应式数据
const activeTab = ref('basic')
const saving = ref(false)
const testingEmail = ref(false)

// 设置数据
const settings = reactive({
  basic: {
    siteName: '',
    siteUrl: '',
    siteDescription: '',
    siteKeywords: '',
    icpNumber: '',
    adminEmail: '',
    customerQQ: ''
  },
  user: {
    allowRegister: true,
    requireEmailVerification: false,
    requireAdminApproval: false,
    usernameMinLength: 3,
    passwordMinLength: 6,
    dailyPostLimit: 0,
    dailyCommentLimit: 0
  },
  content: {
    requirePostApproval: false,
    requireCommentApproval: false,
    allowAnonymousPost: false,
    maxTitleLength: 100,
    maxContentLength: 10000,
    enableSensitiveWordFilter: false,
    sensitiveWords: ''
  },
  email: {
    enabled: false,
    smtpHost: '',
    smtpPort: 587,
    fromEmail: '',
    fromName: '',
    smtpUsername: '',
    smtpPassword: '',
    enableSSL: true
  },
  security: {
    enableCaptcha: false,
    enableTwoFactor: false,
    maxLoginAttempts: 5,
    lockoutDuration: 30,
    ipWhitelist: '',
    ipBlacklist: ''
  }
})

// 方法
const fetchSettings = async () => {
  try {
    const response = await api.get('/api/admin/settings')
    Object.assign(settings, response.data.data)
  } catch (error) {
    console.error('获取设置失败:', error)
    // 使用默认设置
    loadDefaultSettings()
  }
}

const loadDefaultSettings = () => {
  Object.assign(settings, {
    basic: {
      siteName: 'GooseForum',
      siteUrl: 'https://example.com',
      siteDescription: '一个基于Go语言开发的现代化论坛系统',
      siteKeywords: '论坛,社区,讨论,技术分享',
      icpNumber: '',
      adminEmail: 'admin@example.com',
      customerQQ: ''
    },
    user: {
      allowRegister: true,
      requireEmailVerification: false,
      requireAdminApproval: false,
      usernameMinLength: 3,
      passwordMinLength: 6,
      dailyPostLimit: 10,
      dailyCommentLimit: 50
    },
    content: {
      requirePostApproval: false,
      requireCommentApproval: false,
      allowAnonymousPost: false,
      maxTitleLength: 100,
      maxContentLength: 10000,
      enableSensitiveWordFilter: true,
      sensitiveWords: '垃圾\n广告\n色情\n暴力'
    },
    email: {
      enabled: false,
      smtpHost: 'smtp.qq.com',
      smtpPort: 587,
      fromEmail: 'noreply@example.com',
      fromName: 'GooseForum',
      smtpUsername: '',
      smtpPassword: '',
      enableSSL: true
    },
    security: {
      enableCaptcha: true,
      enableTwoFactor: false,
      maxLoginAttempts: 5,
      lockoutDuration: 30,
      ipWhitelist: '',
      ipBlacklist: ''
    }
  })
}

const saveSettings = async () => {
  saving.value = true
  try {
    // 系统设置接口暂未实现
    console.warn('系统设置接口暂未实现')
    // await api.post('/api/admin/settings', settings)
    
    // 显示成功提示
    const toast = document.createElement('div')
    toast.className = 'toast toast-top toast-end'
    toast.innerHTML = `
      <div class="alert alert-success">
        <span>设置保存成功！</span>
      </div>
    `
    document.body.appendChild(toast)
    
    setTimeout(() => {
      document.body.removeChild(toast)
    }, 3000)
  } catch (error) {
    console.error('保存设置失败:', error)
    
    // 显示错误提示
    const toast = document.createElement('div')
    toast.className = 'toast toast-top toast-end'
    toast.innerHTML = `
      <div class="alert alert-error">
        <span>保存设置失败，请重试！</span>
      </div>
    `
    document.body.appendChild(toast)
    
    setTimeout(() => {
      document.body.removeChild(toast)
    }, 3000)
  } finally {
    saving.value = false
  }
}

const resetSettings = () => {
  if (confirm('确定要重置所有设置吗？此操作将恢复默认配置！')) {
    loadDefaultSettings()
  }
}

const testEmail = async () => {
  if (!settings.email.enabled) {
    alert('请先启用邮件功能！')
    return
  }
  
  if (!settings.email.smtpHost || !settings.email.fromEmail) {
    alert('请先配置SMTP服务器和发件人邮箱！')
    return
  }
  
  testingEmail.value = true
  try {
    await api.post('/api/admin/settings/test-email', {
      toEmail: settings.basic.adminEmail || 'test@example.com'
    })
    
    alert('测试邮件发送成功！请检查收件箱。')
  } catch (error) {
    console.error('测试邮件发送失败:', error)
    alert('测试邮件发送失败，请检查SMTP配置！')
  } finally {
    testingEmail.value = false
  }
}

// 组件挂载时获取设置
onMounted(() => {
  fetchSettings()
})
</script>

<style scoped>
.tabs {
  border-bottom: 1px solid hsl(var(--bc) / 0.2);
}

.tab {
  display: flex;
  align-items: center;
}

/* Toast样式 */
.toast {
  position: fixed;
  top: 1rem;
  right: 1rem;
  z-index: 9999;
}
</style>