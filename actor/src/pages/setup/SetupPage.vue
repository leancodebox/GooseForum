<script setup>
import { ref, onMounted } from 'vue'
import {
  NCard,
  NButton,
  NSpace,
  NSteps,
  NStep,
  NForm,
  NFormItem,
  NInput,
  NResult,
  NSelect
} from 'naive-ui'
import { getSetupStatus, submitSetup } from '@/service/request4setup.js'
import { enqueueMessage } from '@/service/messageManager'

const isLoading = ref(false)
const currentStep = ref(0)
const isInitialized = ref(false)

const formData = ref({
  // 站点信息
  siteName: '',
  siteDesc: '',
  // 数据库配置
  dbType: 'mysql', // 默认数据库类型
  dbHost: 'localhost',
  dbPort: '3306',
  dbName: '',
  dbUser: '',
  dbPassword: '',
  sqlitePath: '', // SQLite 文件路径
  // 管理员账号
  adminUsername: '',
  adminPassword: '',
  adminEmail: ''
})

const dbOptions = [
  { label: 'MySQL', value: 'mysql' },
  { label: 'SQLite', value: 'sqlite' }
]

const onDbTypeChange = (value) => {
  // 根据选择的数据库类型，重置相关字段
  if (value === 'sqlite') {
    formData.value.dbHost = ''
    formData.value.dbPort = ''
    formData.value.dbName = ''
    formData.value.dbUser = ''
    formData.value.dbPassword = ''
  } else {
    formData.value.sqlitePath = ''
  }
}

const steps = [
  { title: '欢迎', description: '开始设置您的网站' },
  { title: '站点信息', description: '设置基本信息' },
  { title: '数据库配置', description: '配置数据库连接' },
  { title: '管理员账号', description: '创建管理员账号' }
]

onMounted(async () => {
  let resp  = await getSetupStatus()

  isInitialized.value = resp.result.isInit
  if (isInitialized.value) {
    enqueueMessage('网站已经完成初始化', 'warning')
  }
})

const handleSubmit = async () => {
  try {
    isLoading.value = true
    const response = await submitSetup(formData.value)
    if (response.code === 0) {
      enqueueMessage('初始化成功', 'success')
      window.location.href = '/'
    } else {
      enqueueMessage(response.msg || '初始化失败', 'error')
    }
  } catch (error) {
    enqueueMessage('系统错误：' + error.message, 'error')
  } finally {
    isLoading.value = false
  }
}

const nextStep = () => {
  if (currentStep.value < steps.length - 1) {
    currentStep.value++
  }
}

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

const getStepStatus = (index) => {
  if (index < currentStep.value) {
    return 'finish'
  }
  if (index === currentStep.value) {
    return isLoading.value ? 'process' : 'process'
  }
  return 'wait'
}
</script>

<template>
  <div class="setup-container">
    <n-card class="setup-card">
      <div class="steps-wrapper">
        <n-steps
          :current="currentStep"
        >
          <n-step
            v-for="(step, index) in steps"
            :key="index"
            :title="step.title"
            :description="step.description"
            :status="getStepStatus(index)"
          />
        </n-steps>
      </div>

      <div class="setup-content" v-if="!isInitialized">
        <!-- 欢迎页面 -->
        <div v-if="currentStep === 0" class="step-content">
          <n-result
            status="success"
            title="欢迎使用"
            description="让我们开始配置您的网站"
          >
            <template #footer>
              <n-button type="primary" @click="nextStep">
                开始设置
              </n-button>
            </template>
          </n-result>
        </div>

        <!-- 站点信息 -->
        <div v-if="currentStep === 1" class="step-content">
          <n-form>
            <n-form-item label="站点名称" required>
              <n-input v-model:value="formData.siteName" placeholder="请输入站点名称"/>
            </n-form-item>
            <n-form-item label="站点描述">
              <n-input
                v-model:value="formData.siteDesc"
                type="textarea"
                placeholder="请输入站点描述"
              />
            </n-form-item>
          </n-form>
        </div>

        <!-- 数据库配置 -->
        <div v-if="currentStep === 2" class="step-content">
          <n-form>
            <n-form-item label="数据库类型" required>
              <n-select
                v-model:value="formData.dbType"
                :options="dbOptions"
                placeholder="选择数据库类型"
                @change="onDbTypeChange"
              />
            </n-form-item>

            <n-form-item v-if="formData.dbType === 'mysql'" label="数据库主机" required>
              <n-input v-model:value="formData.dbHost" placeholder="请输入数据库主机"/>
            </n-form-item>
            <n-form-item v-if="formData.dbType === 'mysql'" label="数据库端口" required>
              <n-input v-model:value="formData.dbPort" placeholder="请输入数据库端口"/>
            </n-form-item>
            <n-form-item v-if="formData.dbType === 'mysql'" label="数据库名称" required>
              <n-input v-model:value="formData.dbName" placeholder="请输入数据库名称"/>
            </n-form-item>
            <n-form-item v-if="formData.dbType === 'mysql'" label="数据库用户名" required>
              <n-input v-model:value="formData.dbUser" placeholder="请输入数据库用户名"/>
            </n-form-item>
            <n-form-item v-if="formData.dbType === 'mysql'" label="数据库密码" required>
              <n-input
                v-model:value="formData.dbPassword"
                type="password"
                show-password-on="click"
                placeholder="请输入数据库密码"
              />
            </n-form-item>

            <n-form-item v-if="formData.dbType === 'sqlite'" label="SQLite 数据库文件路径" required>
              <n-input v-model:value="formData.sqlitePath" placeholder="请输入 SQLite 数据库文件路径"/>
            </n-form-item>
          </n-form>
        </div>

        <!-- 管理员账号 -->
        <div v-if="currentStep === 3" class="step-content">
          <n-form>
            <n-form-item label="管理员用户名" required>
              <n-input v-model:value="formData.adminUsername"/>
            </n-form-item>
            <n-form-item label="管理员密码" required>
              <n-input
                v-model:value="formData.adminPassword"
                type="password"
                show-password-on="click"
              />
            </n-form-item>
            <n-form-item label="管理员邮箱" required>
              <n-input v-model:value="formData.adminEmail"/>
            </n-form-item>
          </n-form>
        </div>

        <!-- 步骤控制按钮 -->
        <div class="step-actions">
          <n-space justify="center">
            <n-button
              v-if="currentStep > 0"
              @click="prevStep"
              :disabled="isLoading"
            >
              上一步
            </n-button>
            <n-button
              v-if="currentStep < steps.length - 1"
              type="primary"
              @click="nextStep"
              :disabled="isLoading"
            >
              下一步
            </n-button>
            <n-button
              v-if="currentStep === steps.length - 1"
              type="primary"
              @click="handleSubmit"
              :loading="isLoading"
            >
              完成设置
            </n-button>
          </n-space>
        </div>
      </div>

      <!-- 已初始化提示 -->
      <div v-else class="initialized-notice">
        <n-result
          status="info"
          title="网站已初始化"
          description="网站已经完成初始化设置，如需重新设置请联系管理员"
        >
          <template #footer>
            <n-button type="primary" @click="() => window.location.href = '/'">
              返回首页
            </n-button>
          </template>
        </n-result>
      </div>
    </n-card>
  </div>
</template>

<style scoped>
.setup-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #f0f2f5;
  padding: 20px;
}

.setup-card {
  width: 100%;
  max-width: 800px;
}

.setup-content {
  margin-top: 48px;
}

.step-content {
  margin: 24px 0;
  min-height: 200px;
}

.step-actions {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #f0f0f0;
}

.initialized-notice {
  margin: 48px 0;
}

.steps-wrapper {
  width: 80%;
  margin: 0 auto;
  padding: 20px 0;
}
</style>
