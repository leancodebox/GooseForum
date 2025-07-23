<template>
  <div class="space-y-6">
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

    <!-- 用户设置表单 -->
    <div class="space-y-8">
      <!-- 注册设置 -->
      <div>
        <div class="flex justify-between items-center">
          <h2 class="flex items-center gap-2 mb-6 text-lg font-medium text-base-content">
            <UserPlusIcon class="w-5 h-5" />
            注册设置
          </h2>
          <button @click="handleSave" class="btn btn-primary btn-sm" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-sm"></span>
            <CheckIcon v-else class="w-4 h-4" />
            {{ saving ? '保存中...' : '保存设置' }}
          </button>
        </div>


        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- 允许用户注册 -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <input v-model="settings.allowRegistration" type="checkbox" class="toggle toggle-primary" />
              <span class="label-text">允许用户注册</span>
            </label>
          </div>

          <!-- 注册需要邮箱验证 -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <input v-model="settings.requireEmailVerification" type="checkbox" class="toggle toggle-primary" />
              <span class="label-text">注册需要邮箱验证</span>
            </label>
          </div>

          <!-- 注册需要管理员审核 -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <input v-model="settings.requireAdminApproval" type="checkbox" class="toggle toggle-primary" />
              <span class="label-text">注册需要管理员审核</span>
            </label>
          </div>

          <!-- 默认用户组 -->
          <label class="floating-label">
            <span>默认用户组</span>
            <select
              v-model="settings.defaultUserGroup"
              class="select select-bordered w-full peer"
              id="defaultUserGroup"
            >
              <option value="user">普通用户</option>
              <option value="vip">VIP用户</option>
              <option value="contributor">贡献者</option>
            </select>
          </label>

          <!-- 用户名最小长度 -->
          <label class="floating-label">
            <span>用户名最小长度</span>
            <input
              v-model="settings.minUsernameLength"
              type="number"
              min="2"
              max="20"
              class="input input-bordered w-full peer"
              id="minUsernameLength"
              placeholder=" "
            />
          </label>

          <!-- 用户名最大长度 -->
          <label class="floating-label">
            <span>用户名最大长度</span>
            <input
              v-model="settings.maxUsernameLength"
              type="number"
              min="5"
              max="50"
              class="input input-bordered w-full peer"
              id="maxUsernameLength"
              placeholder=" "
            />
          </label>
        </div>
      </div>

      <!-- 分割线 -->
      <div class="divider"></div>

      <!-- 登录设置 -->
      <div>
        <h2 class="flex items-center gap-2 mb-6 text-lg font-medium text-base-content">
          <KeyIcon class="w-5 h-5" />
          登录设置
        </h2>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- 允许记住登录状态 -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <span class="label-text">允许记住登录状态</span>
              <input v-model="settings.allowRememberMe" type="checkbox" class="toggle toggle-primary" />
            </label>
          </div>

          <!-- 登录失败锁定 -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <span class="label-text">启用登录失败锁定</span>
              <input v-model="settings.enableLoginLockout" type="checkbox" class="toggle toggle-primary" />
            </label>
          </div>

          <!-- 最大登录失败次数 -->
          <label class="floating-label">
            <span>最大登录失败次数</span>
            <input
              v-model="settings.maxLoginAttempts"
              type="number"
              min="3"
              max="10"
              class="input input-bordered w-full peer"
              id="maxLoginAttempts"
              placeholder=" "
            />
          </label>

          <!-- 锁定时间（分钟） -->
          <label class="floating-label">
            <span>锁定时间（分钟）</span>
            <input
              v-model="settings.lockoutDuration"
              type="number"
              min="5"
              max="1440"
              class="input input-bordered w-full peer"
              id="lockoutDuration"
              placeholder=" "
            />
          </label>

          <!-- 会话超时时间 -->
          <label class="floating-label">
            <span>会话超时时间（分钟）</span>
            <input
              v-model="settings.sessionTimeout"
              type="number"
              min="30"
              max="10080"
              class="input input-bordered w-full peer"
              id="sessionTimeout"
              placeholder=" "
            />
          </label>
        </div>
      </div>

      <!-- 分割线 -->
      <div class="divider"></div>

      <!-- 用户权限 -->
      <div>
        <h2 class="flex items-center gap-2 mb-6 text-lg font-medium text-base-content">
          <ShieldCheckIcon class="w-5 h-5" />
          用户权限
        </h2>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- 允许用户修改个人资料 -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <span class="label-text">允许用户修改个人资料</span>
              <input v-model="settings.allowProfileEdit" type="checkbox" class="toggle toggle-primary" />
            </label>
          </div>

          <!-- 允许用户上传头像 -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <span class="label-text">允许用户上传头像</span>
              <input v-model="settings.allowAvatarUpload" type="checkbox" class="toggle toggle-primary" />
            </label>
          </div>

          <!-- 允许用户删除自己的帖子 -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <span class="label-text">允许用户删除自己的帖子</span>
              <input v-model="settings.allowSelfPostDelete" type="checkbox" class="toggle toggle-primary" />
            </label>
          </div>

          <!-- 允许用户编辑自己的帖子 -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <span class="label-text">允许用户编辑自己的帖子</span>
              <input v-model="settings.allowSelfPostEdit" type="checkbox" class="toggle toggle-primary" />
            </label>
          </div>

          <!-- 帖子编辑时间限制 -->
          <label class="floating-label">
            <input
              v-model="settings.postEditTimeLimit"
              type="number"
              min="0"
              max="1440"
              class="input input-bordered w-full peer"
              id="postEditTimeLimit"
              placeholder=" "
            />
            <label for="postEditTimeLimit" class="floating-label-text">帖子编辑时间限制（分钟，0为无限制）</label>
          </label>
        </div>
      </div>

      <!-- 分割线 -->
      <div class="divider"></div>

      <!-- 积分设置 -->
      <div>
        <h2 class="flex items-center gap-2 mb-6 text-lg font-medium text-base-content">
          <StarIcon class="w-5 h-5" />
          积分设置
        </h2>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- 启用积分系统 -->
          <div class="form-control">
            <label class="label cursor-pointer">
              <input v-model="settings.enablePointSystem" type="checkbox" class="toggle toggle-primary" />
              <span class="label-text">启用积分系统</span>
            </label>
          </div>

          <!-- 发帖获得积分 -->
          <label class="floating-label">
            <span>发帖获得积分</span>
            <input
              v-model="settings.pointsForPost"
              type="number"
              min="0"
              max="100"
              class="input input-bordered w-full peer"
              id="pointsForPost"
              placeholder=" "
            />
          </label>

          <!-- 回复获得积分 -->
          <label class="floating-label">
            <span>回复获得积分</span>
            <input
              v-model="settings.pointsForReply"
              type="number"
              min="0"
              max="50"
              class="input input-bordered w-full peer"
              id="pointsForReply"
              placeholder=" "
            />
          </label>

          <!-- 每日登录获得积分 -->
          <label class="floating-label">
            <span>每日登录获得积分</span>
            <input
              v-model="settings.pointsForDailyLogin"
              type="number"
              min="0"
              max="20"
              class="input input-bordered w-full peer"
              id="pointsForDailyLogin"
              placeholder=" "
            />
          </label>

          <!-- 新用户初始积分 -->
          <label class="floating-label">
            <span>新用户初始积分</span>
            <input
              v-model="settings.initialPoints"
              type="number"
              min="0"
              max="1000"
              class="input input-bordered w-full peer"
              id="initialPoints"
              placeholder=" "
            />
          </label>
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
  UserPlusIcon,
  KeyIcon,
  ShieldCheckIcon,
  StarIcon, ServerIcon
} from '@heroicons/vue/24/outline'

// 设置数据
const settings = ref({
  // 注册设置
  allowRegistration: true,
  requireEmailVerification: true,
  requireAdminApproval: false,
  defaultUserGroup: 'user',
  minUsernameLength: 3,
  maxUsernameLength: 20,

  // 登录设置
  allowRememberMe: true,
  enableLoginLockout: true,
  maxLoginAttempts: 5,
  lockoutDuration: 30,
  sessionTimeout: 1440,

  // 用户权限
  allowProfileEdit: true,
  allowAvatarUpload: true,
  allowSelfPostDelete: true,
  allowSelfPostEdit: true,
  postEditTimeLimit: 60,

  // 积分设置
  enablePointSystem: true,
  pointsForPost: 10,
  pointsForReply: 5,
  pointsForDailyLogin: 2,
  initialPoints: 100
})

// 状态变量
const saving = ref(false)
const showSuccessAlert = ref(false)
const showErrorAlert = ref(false)
const errorMessage = ref('')

// 保存设置
const handleSave = async () => {
  saving.value = true
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    showSuccess()
  } catch (error) {
    console.error('保存用户设置失败:', error)
    showError('保存设置失败，请重试')
  } finally {
    saving.value = false
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
  console.log('用户设置页面已加载')
})
</script>

<style scoped>
/* 自定义样式 */
</style>
