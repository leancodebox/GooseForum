<template>
  <div class=" min-h-[70vh] bg-base-200 flex items-center justify-center p-4">
    <div class="card w-full max-w-md bg-base-100 shadow-xl">
      <div class="card-body">
        <h1 class="text-3xl font-normal text-center mb-8">重置密码</h1>

        <div v-if="!tokenValid" class="alert alert-error">
          <span>重置链接无效或已过期，请重新申请密码重置</span>
          <div class="mt-4">
            <a href="/login" class="btn btn-primary">返回登录</a>
          </div>
        </div>

        <form v-else @submit.prevent="handleResetPassword" class="space-y-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">新密码</span>
            </label>
            <input
                v-model="resetForm.password"
                type="password"
                placeholder="请输入新密码"
                class="input input-bordered w-full"
                :class="{ 'input-error': resetErrors.password }"
                required
            />
            <label v-if="resetErrors.password" class="label">
              <span class="label-text-alt text-error">{{ resetErrors.password }}</span>
            </label>
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">确认新密码</span>
            </label>
            <input
                v-model="resetForm.confirmPassword"
                type="password"
                placeholder="请再次输入新密码"
                class="input input-bordered w-full"
                :class="{ 'input-error': resetErrors.confirmPassword }"
                required
            />
            <label v-if="resetErrors.confirmPassword" class="label">
              <span class="label-text-alt text-error">{{ resetErrors.confirmPassword }}</span>
            </label>
          </div>

          <div v-if="resetErrors.general" class="alert alert-error mb-4">
            <span>{{ resetErrors.general }}</span>
          </div>

          <div v-if="resetSuccess" class="alert alert-success mb-4">
            <span>{{ resetSuccess }}</span>
          </div>
          <div class="text-center mt-4" v-if="resetSuccess">
            <a href="/login" class="btn btn-primary w-full">前往登录</a>
          </div>

          <div class="form-control mt-6" v-if="!resetSuccess">
            <button type="submit" class="btn btn-primary w-full" :disabled="resetLoading">
              <span v-if="resetLoading" class="loading loading-spinner loading-sm"></span>
              {{ resetLoading ? '重置中...' : '重置密码' }}
            </button>
          </div>

          <div class="text-center mt-4" v-if="!resetSuccess">
            <a href="/login" class="link link-primary text-sm">返回登录</a>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'


// 重置表单数据
const resetForm = reactive({
  password: '',
  confirmPassword: ''
})

// 错误信息
const resetErrors = reactive({})

// 加载状态
const resetLoading = ref(false)

// 成功信息
const resetSuccess = ref('')

// token有效性
const tokenValid = ref(true)

// 重置token
const resetToken = ref('')

// 验证表单
const validateReset = () => {
  const errors = {}

  if (!resetForm.password.trim()) {
    errors.password = '请输入新密码'
  } else if (resetForm.password.length < 6) {
    errors.password = '密码长度至少6位'
  }

  if (!resetForm.confirmPassword.trim()) {
    errors.confirmPassword = '请确认新密码'
  } else if (resetForm.password !== resetForm.confirmPassword) {
    errors.confirmPassword = '两次输入的密码不一致'
  }

  Object.assign(resetErrors, errors)
  return Object.keys(errors).length === 0
}

// 处理密码重置
const handleResetPassword = async () => {
  // 清空之前的错误
  Object.keys(resetErrors).forEach(key => delete resetErrors[key])

  if (!validateReset()) {
    return
  }

  resetLoading.value = true

  try {
    const response = await fetch('/api/reset-password', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        token: resetToken.value,
        newPassword: resetForm.password
      })
    })

    const data = await response.json()

    if (data.code === 0) {
      resetSuccess.value = '密码重置成功！您现在可以使用新密码登录了'
      // 清空表单
      resetForm.password = ''
      resetForm.confirmPassword = ''
    } else {
      resetErrors.general = data.message || '密码重置失败，请稍后重试'
    }
  } catch (error) {
    console.error('密码重置请求失败:', error)
    resetErrors.general = '密码重置失败，请稍后重试'
  } finally {
    resetLoading.value = false
  }
}

// 页面加载时检查token
onMounted(() => {
  // 从URL路径中获取token参数
  const urlParams = new URLSearchParams(window.location.search)
  resetToken.value = urlParams.get('token') || ''
  if (!resetToken.value) {
    tokenValid.value = false
  }
})
</script>
