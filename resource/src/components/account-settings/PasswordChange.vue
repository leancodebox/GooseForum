<script setup lang="ts">
import { reactive, ref } from 'vue'
import { changePassword } from '@/utils/gooseForumService.ts'

// 密码修改表单
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 密码修改状态
const changingPassword = ref(false)

// 修改密码
const updatePassword = async () => {
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    alert('新密码和确认密码不一致')
    return
  }

  // 验证新密码长度
  if (passwordForm.newPassword.length < 6) {
    alert('新密码长度不能少于6位')
    return
  }

  try {
    changingPassword.value = true
    const response = await changePassword(
      passwordForm.currentPassword,
      passwordForm.newPassword
    )

    if (response.code === 0) {
      alert('密码修改成功')
      // 重置表单
      Object.assign(passwordForm, {
        currentPassword: '',
        newPassword: '',
        confirmPassword: ''
      })
    } else {
      alert(`密码修改失败: ${response.msg || '请重试'}`)
    }
  } catch (error) {
    console.error('修改密码失败:', error)
    alert('密码修改失败，请重试'+error)
  } finally {
    changingPassword.value = false
  }
}
</script>

<template>
  <div>
    <h3 class="card-title text-lg mb-6 border-b border-base-300 pb-3">修改密码</h3>
    <form @submit.prevent="updatePassword" class="grid grid-cols-1 gap-6">
      <div class="form-control">
        <label class="label">
          <span class="label-text font-normal">当前密码</span>
        </label>
        <input v-model="passwordForm.currentPassword" type="password" class="input input-bordered w-full"
          placeholder="请输入当前密码" required />
      </div>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div class="form-control">
          <label class="label">
            <span class="label-text font-normal">新密码</span>
          </label>
          <input v-model="passwordForm.newPassword" type="password" class="input input-bordered w-full"
            placeholder="请输入新密码" required />
          <label class="label">
            <span class="label-text-alt">密码长度至少8位，包含字母和数字</span>
          </label>
        </div>
        <div class="form-control">
          <label class="label">
            <span class="label-text font-normal">确认新密码</span>
          </label>
          <input v-model="passwordForm.confirmPassword" type="password" class="input input-bordered w-full"
            placeholder="请再次输入新密码" required />
        </div>
      </div>
      <div class="flex justify-end">
        <button type="submit" class="btn btn-secondary min-w-32" :disabled="changingPassword">
          <span v-if="changingPassword" class="loading loading-spinner loading-sm"></span>
          {{ changingPassword ? '修改中...' : '修改密码' }}
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped></style>
