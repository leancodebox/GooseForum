<template>
  <div class="login-container">
    <n-card class="login-card" title="GooseForum 管理系统">
      <n-form
        ref="formRef"
        :model="formValue"
        :rules="rules"
        label-placement="left"
        label-width="auto"
        require-mark-placement="right-hanging"
      >
        <n-form-item path="username" label="用户名">
          <n-input v-model:value="formValue.username" placeholder="请输入用户名" />
        </n-form-item>
        <n-form-item path="password" label="密码">
          <n-input
            v-model:value="formValue.password"
            type="password"
            placeholder="请输入密码"
            @keydown.enter="handleLogin"
          />
        </n-form-item>
      </n-form>
      <div class="login-button-container">
        <n-button type="primary" block @click="handleLogin" :loading="loading">
          登录
        </n-button>
      </div>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'

const router = useRouter()
const message = useMessage()
const formRef = ref<FormInst | null>(null)
const loading = ref(false)

const formValue = reactive({
  username: '',
  password: ''
})

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

const handleLogin = () => {
  formRef.value?.validate((errors) => {
    if (!errors) {
      loading.value = true
      
      // 模拟登录请求
      setTimeout(() => {
        loading.value = false
        
        // 这里应该是实际的登录逻辑
        // localStorage.setItem('admin_token', 'your-token')
        
        message.success('登录成功')
        router.push('/')
      }, 1000)
    }
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f0f2f5;
}

.login-card {
  width: 400px;
}

.login-button-container {
  margin-top: 24px;
}
</style>