<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-normal text-base-content">系统设置</h1>
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
      <!-- 基本设置 -->
      <div class="lg:col-span-2 space-y-6">
        <!-- 网站基本信息 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2 mb-6">
              <GlobeAltIcon class="w-5 h-5" />
              网站基本信息
            </h2>

            <div class="space-y-6">
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <label class="floating-label">
                  <span>网站名称</span>
                  <input v-model="settings.siteName" type="text" placeholder="请输入网站名称" class="input input-bordered w-full" />
                </label>
                <label class="floating-label">
                  <span>网站域名</span>
                  <input v-model="settings.siteUrl" type="url" placeholder="请输入网站域名" class="input input-bordered w-full" />
                </label>
              </div>

              <div class="form-control">
                <label class="floating-label">
                  <span>网站描述</span>
                  <textarea class="textarea textarea-bordered h-24 w-full" placeholder="请输入网站描述" v-model="settings.siteDescription"></textarea>
                </label>
                <div class="label">
                  <span class="label-text-alt text-base-content/60">用于SEO和网站介绍</span>
                </div>
              </div>

              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="form-control">
                  <label class="floating-label">
                    <span>网站关键词</span>
                    <input v-model="settings.siteKeywords" type="text" placeholder="请输入网站关键词" class="input input-bordered w-full" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">用逗号分隔多个关键词</span>
                  </div>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>ICP备案号</span>
                    <input v-model="settings.icpNumber" type="text" placeholder="请输入ICP备案号" class="input input-bordered w-full" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">可选，用于网站底部显示</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 用户设置 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2 mb-6">
              <UsersIcon class="w-5 h-5" />
              用户设置
            </h2>

            <div class="space-y-6">
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="form-control">
                  <label class="label cursor-pointer justify-start gap-4">
                    <input v-model="settings.allowRegistration" type="checkbox" class="toggle toggle-primary" />
                    <div class="flex flex-col">
                      <span class="label-text font-normal">允许用户注册</span>
                      <span class="label-text-alt text-base-content/60">开启后新用户可以注册账号</span>
                    </div>
                  </label>
                </div>
                <div class="form-control">
                  <label class="label cursor-pointer justify-start gap-4">
                    <input v-model="settings.requireEmailVerification" type="checkbox" class="toggle toggle-primary" />
                    <div class="flex flex-col">
                      <span class="label-text font-normal">需要邮箱验证</span>
                      <span class="label-text-alt text-base-content/60">新用户注册后需验证邮箱</span>
                    </div>
                  </label>
                </div>
              </div>

              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="form-control">
                  <label class="floating-label">
                    <span>默认用户角色</span>
                    <select v-model="settings.defaultUserRole" class="select select-bordered w-full">
                      <option value="" disabled>请选择默认用户角色</option>
                      <option value="user">普通用户</option>
                      <option value="vip">VIP用户</option>
                      <option value="moderator">版主</option>
                    </select>
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">新注册用户的默认权限级别</span>
                  </div>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>用户名最小长度</span>
                    <input v-model.number="settings.minUsernameLength" type="number" min="2" max="20" placeholder="请输入用户名最小长度" class="input input-bordered w-full" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">用户名允许的最少字符数</span>
                  </div>
                </div>
              </div>

              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="form-control">
                  <label class="floating-label">
                    <span>密码最小长度</span>
                    <input v-model.number="settings.minPasswordLength" type="number" min="6" max="50" placeholder="请输入密码最小长度" class="input input-bordered w-full" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">密码允许的最少字符数</span>
                  </div>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>会话过期时间(小时)</span>
                    <input v-model.number="settings.sessionTimeout" type="number" min="1" max="720" placeholder="请输入会话过期时间" class="input input-bordered w-full" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">用户登录状态保持时间</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 内容设置 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2 mb-6">
              <DocumentTextIcon class="w-5 h-5" />
              内容设置
            </h2>

            <div class="space-y-6">
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="form-control">
                  <label class="label cursor-pointer justify-start gap-4">
                    <input v-model="settings.requirePostApproval" type="checkbox" class="toggle toggle-primary" />
                    <div class="flex flex-col">
                      <span class="label-text font-normal">需要审核新帖子</span>
                      <span class="label-text-alt text-base-content/60">新发布的帖子需要管理员审核</span>
                    </div>
                  </label>
                </div>
                <div class="form-control">
                  <label class="label cursor-pointer justify-start gap-4">
                    <input v-model="settings.requireCommentApproval" type="checkbox" class="toggle toggle-primary" />
                    <div class="flex flex-col">
                      <span class="label-text font-normal">需要审核评论</span>
                      <span class="label-text-alt text-base-content/60">新发布的评论需要管理员审核</span>
                    </div>
                  </label>
                </div>
              </div>

              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="form-control">
                  <label class="floating-label">
                    <span>每页帖子数量</span>
                    <input v-model.number="settings.postsPerPage" type="number" min="5" max="100" placeholder="请输入每页帖子数量" class="input input-bordered w-full" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">列表页面每页显示的帖子数量</span>
                  </div>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>帖子标题最大长度</span>
                    <input v-model.number="settings.maxTitleLength" type="number" min="10" max="200" placeholder="请输入帖子标题最大长度" class="input input-bordered w-full" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">帖子标题允许的最大字符数</span>
                  </div>
                </div>
              </div>

              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="form-control">
                  <label class="floating-label">
                    <span>帖子内容最大长度</span>
                    <input v-model.number="settings.maxContentLength" type="number" min="100" max="50000" placeholder="请输入帖子内容最大长度" class="input input-bordered w-full" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">帖子内容允许的最大字符数</span>
                  </div>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>最大文件大小(MB)</span>
                    <input v-model.number="settings.maxFileSize" type="number" min="1" max="100" placeholder="请输入最大文件大小" class="input input-bordered w-full" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">单个文件上传的大小限制</span>
                  </div>
                </div>
              </div>

              <div class="form-control">
                <label class="floating-label">
                  <span>允许的文件类型</span>
                  <input v-model="settings.allowedFileTypes" type="text" placeholder="请输入允许的文件类型" class="input input-bordered bg-base-200 w-full" />
                </label>
                <div class="label">
                  <span class="label-text-alt text-base-content/60">用逗号分隔多个文件扩展名</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 邮件设置 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2 mb-6">
              <EnvelopeIcon class="w-5 h-5" />
              邮件设置
            </h2>

            <div class="space-y-6">
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="form-control">
                  <label class="label cursor-pointer justify-start gap-4">
                    <input v-model="settings.enableEmail" type="checkbox" class="toggle toggle-primary" />
                    <div class="flex flex-col">
                      <span class="label-text font-normal">启用邮件服务</span>
                      <span class="label-text-alt text-base-content/60">开启后系统可以发送邮件通知</span>
                    </div>
                  </label>
                </div>
                <div class="form-control">
                  <label class="label cursor-pointer justify-start gap-4">
                    <input v-model="settings.smtpSSL" type="checkbox" class="toggle toggle-primary"
                      :disabled="!settings.enableEmail" />
                    <div class="flex flex-col">
                      <span class="label-text font-normal">启用SSL/TLS</span>
                      <span class="label-text-alt text-base-content/60">推荐开启以提高安全性</span>
                    </div>
                  </label>
                </div>
              </div>

              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="form-control">
                  <label class="floating-label">
                    <span>SMTP服务器</span>
                    <input v-model="settings.smtpHost" type="text" placeholder="请输入SMTP服务器地址" class="input input-bordered w-full" :disabled="!settings.enableEmail" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">邮件服务器地址</span>
                  </div>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>SMTP端口</span>
                    <input v-model.number="settings.smtpPort" type="number" placeholder="请输入SMTP端口" class="input input-bordered w-full" :disabled="!settings.enableEmail" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">通常为25、465或587</span>
                  </div>
                </div>
              </div>

              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="form-control">
                  <label class="floating-label">
                    <span>发件人邮箱</span>
                    <input v-model="settings.smtpUsername" type="email" placeholder="请输入发件人邮箱" class="input input-bordered w-full" :disabled="!settings.enableEmail" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">用于发送邮件的邮箱地址</span>
                  </div>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>邮箱密码</span>
                    <input v-model="settings.smtpPassword" type="password" placeholder="请输入邮箱密码" class="input input-bordered w-full" :disabled="!settings.enableEmail" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">邮箱密码或应用专用密码</span>
                  </div>
                </div>
              </div>

              <div class="pt-4 border-t border-base-300">
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
      </div>

      <!-- 侧边栏设置 -->
      <div class="space-y-6">
        <!-- 系统状态 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2 mb-6">
              <ServerIcon class="w-5 h-5" />
              系统状态
            </h2>
            <div class="space-y-4">
              <div class="flex items-center justify-between p-3 bg-base-200 rounded-lg">
                <span class="text-sm font-normal">服务器状态</span>
                <div class="badge badge-success gap-2">
                  <div class="w-2 h-2 bg-success rounded-full"></div>
                  正常
                </div>
              </div>
              <div class="flex items-center justify-between p-3 bg-base-200 rounded-lg">
                <span class="text-sm font-normal">数据库</span>
                <div class="badge badge-success gap-2">
                  <div class="w-2 h-2 bg-success rounded-full"></div>
                  连接正常
                </div>
              </div>
              <div class="flex items-center justify-between p-3 bg-base-200 rounded-lg">
                <span class="text-sm font-normal">缓存服务</span>
                <div class="badge badge-warning gap-2">
                  <div class="w-2 h-2 bg-warning rounded-full"></div>
                  未启用
                </div>
              </div>
              <div class="flex items-center justify-between p-3 bg-base-200 rounded-lg">
                <span class="text-sm font-normal">邮件服务</span>
                <div :class="['badge gap-2', settings.enableEmail ? 'badge-success' : 'badge-error']">
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
            <h2 class="card-title flex items-center gap-2 mb-6">
              <ShieldCheckIcon class="w-5 h-5" />
              安全设置
            </h2>

            <div class="space-y-6">
              <div class="grid grid-cols-1 gap-6">
                <div class="form-control">
                  <label class="label cursor-pointer justify-start gap-4">
                    <input v-model="settings.enableCaptcha" type="checkbox" class="toggle toggle-primary" />
                    <div class="flex flex-col">
                      <span class="label-text font-normal">启用验证码</span>
                      <span class="label-text-alt text-base-content/60">登录和注册时需要验证码</span>
                    </div>
                  </label>
                </div>
                <div class="form-control">
                  <label class="label cursor-pointer justify-start gap-4">
                    <input v-model="settings.enableIPLimit" type="checkbox" class="toggle toggle-primary" />
                    <div class="flex flex-col">
                      <span class="label-text font-normal">启用IP限制</span>
                      <span class="label-text-alt text-base-content/60">限制可疑IP地址访问</span>
                    </div>
                  </label>
                </div>
              </div>

              <div class="grid grid-cols-1 gap-6">
                <div class="form-control">
                  <label class="floating-label">
                    <span>登录失败限制</span>
                    <input v-model.number="settings.maxLoginAttempts" type="number" min="3" max="10" placeholder="请输入登录失败限制次数" class="input input-bordered w-full" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">超过次数将锁定账户</span>
                  </div>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>锁定时间(分钟)</span>
                    <input v-model.number="settings.lockoutDuration" type="number" min="5" max="1440" placeholder="请输入锁定时间" class="input input-bordered w-full" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">账户被锁定的持续时间</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 缓存设置 -->
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title flex items-center gap-2 mb-6">
              <BoltIcon class="w-5 h-5" />
              缓存设置
            </h2>

            <div class="space-y-6">
              <div class="form-control">
                <label class="label cursor-pointer justify-start gap-4">
                  <input v-model="settings.enablePageCache" type="checkbox" class="toggle toggle-primary" />
                  <div class="flex flex-col">
                    <span class="label-text font-normal">启用页面缓存</span>
                    <span class="label-text-alt text-base-content/60">缓存页面内容以提高性能</span>
                  </div>
                </label>
              </div>

              <div class="form-control">
                <label class="floating-label">
                  <span>缓存过期时间(分钟)</span>
                  <input v-model.number="settings.cacheExpiration" type="number" min="1" max="1440" placeholder="请输入缓存过期时间" class="input input-bordered w-full" :disabled="!settings.enablePageCache" />
                </label>
                <div class="label">
                  <span class="label-text-alt text-base-content/60">缓存数据的有效期</span>
                </div>
              </div>

              <div class="pt-4 border-t border-base-300">
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
            <h2 class="card-title flex items-center gap-2 mb-6">
              <CloudArrowUpIcon class="w-5 h-5" />
              备份设置
            </h2>

            <div class="space-y-6">
              <div class="form-control">
                <label class="label cursor-pointer justify-start gap-4">
                  <input v-model="settings.enableAutoBackup" type="checkbox" class="toggle toggle-primary" />
                  <div class="flex flex-col">
                    <span class="label-text font-normal">启用自动备份</span>
                    <span class="label-text-alt text-base-content/60">定期自动备份系统数据</span>
                  </div>
                </label>
              </div>

              <div class="grid grid-cols-1 gap-6">
                <div class="form-control">
                  <label class="floating-label">
                    <span>备份频率</span>
                    <select v-model="settings.backupFrequency" class="select select-bordered w-full" :disabled="!settings.enableAutoBackup">
                      <option value="" disabled>请选择备份频率</option>
                      <option value="daily">每日</option>
                      <option value="weekly">每周</option>
                      <option value="monthly">每月</option>
                    </select>
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">自动备份的执行频率</span>
                  </div>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>保留备份数量</span>
                    <input v-model.number="settings.backupRetention" type="number" min="1" max="30" placeholder="请输入保留备份数量" class="input input-bordered w-full" :disabled="!settings.enableAutoBackup" />
                  </label>
                  <div class="label">
                    <span class="label-text-alt text-base-content/60">最多保留的备份文件数量</span>
                  </div>
                </div>
              </div>
              
              <div class="pt-4 border-t border-base-300">
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