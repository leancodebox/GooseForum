<script setup>
import {ref, onMounted} from 'vue'
import {useMessage} from 'naive-ui'
import {getSetupStatus, submitSetup} from '@/service/request'

const message = useMessage()

const currentStep = ref(0)
const stepStatus = ref('process')
const isCompleted = ref(false)
const submitting = ref(false)

const setupForm = ref({
  siteName: '',
  siteDesc: '',
  dbHost: 'localhost',
  dbPort: '3306',
  dbName: '',
  dbUser: '',
  dbPassword: '',
  adminUsername: '',
  adminPassword: '',
  adminEmail: ''
})

// 表单验证规则
const basicRules = {
  siteName: {required: true, message: '请输入网站名称', trigger: 'blur'}
}

const dbRules = {
  dbHost: {required: true, message: '请输入数据库主机', trigger: 'blur'},
  dbPort: {required: true, message: '请输入数据库端口', trigger: 'blur'},
  dbName: {required: true, message: '请输入数据库名', trigger: 'blur'},
  dbUser: {required: true, message: '请输入数据库用户名', trigger: 'blur'},
  dbPassword: {required: true, message: '请输入数据库密码', trigger: 'blur'}
}

const adminRules = {
  adminUsername: {required: true, message: '请输入管理员用户名', trigger: 'blur'},
  adminPassword: {required: true, message: '请输入管理员密码', trigger: 'blur'},
  adminEmail: {required: true, type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur'}
}

// 检查初始化状态
onMounted(async () => {
  try {
    const res = await getSetupStatus()
    if (res.result.isInit) {
      window.location.href = '/actor/'
    }
  } catch (error) {
    console.error('Failed to check setup status:', error)
  }
})

const nextStep = () => {
  if (currentStep.value < 2) {
    currentStep.value++
  }
}

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    await submitSetup(setupForm.value)
    isCompleted.value = true
    message.success('初始化成功')
  } catch (error) {
    message.error('初始化失败：' + error.message)
  } finally {
    submitting.value = false
  }
}

const goToHome = () => {
  window.location.href = '/actor/'
}
</script>
<template>
  <div class="setup-container">
    <n-card title="网站初始化设置" class="setup-card">
      <n-steps :current="currentStep" :status="stepStatus">
        <n-step title="基本信息" description="设置网站基本信息"/>
        <n-step title="数据库配置" description="配置数据库连接"/>
        <n-step title="管理员账号" description="创建管理员账号"/>
      </n-steps>

      <div class="setup-form" v-if="!isCompleted">
        <!-- 第一步：基本信息 -->
        <n-form v-if="currentStep === 0" ref="formRef1" :model="setupForm" :rules="basicRules">
          <n-form-item label="网站名称" path="siteName">
            <n-input v-model:value="setupForm.siteName" placeholder="请输入网站名称"/>
          </n-form-item>
          <n-form-item label="网站描述" path="siteDesc">
            <n-input v-model:value="setupForm.siteDesc" type="textarea" placeholder="请输入网站描述"/>
          </n-form-item>
        </n-form>

        <!-- 第二步：数据库配置 -->
        <n-form v-if="currentStep === 1" ref="formRef2" :model="setupForm" :rules="dbRules">
          <n-form-item label="数据库主机" path="dbHost">
            <n-input v-model:value="setupForm.dbHost" placeholder="localhost"/>
          </n-form-item>
          <n-form-item label="数据库端口" path="dbPort">
            <n-input v-model:value="setupForm.dbPort" placeholder="3306"/>
          </n-form-item>
          <n-form-item label="数据库名" path="dbName">
            <n-input v-model:value="setupForm.dbName"/>
          </n-form-item>
          <n-form-item label="用户名" path="dbUser">
            <n-input v-model:value="setupForm.dbUser"/>
          </n-form-item>
          <n-form-item label="密码" path="dbPassword">
            <n-input v-model:value="setupForm.dbPassword" type="password"/>
          </n-form-item>
        </n-form>

        <!-- 第三步：管理员账号 -->
        <n-form v-if="currentStep === 2" ref="formRef3" :model="setupForm" :rules="adminRules">
          <n-form-item label="管理员用户名" path="adminUsername">
            <n-input v-model:value="setupForm.adminUsername"/>
          </n-form-item>
          <n-form-item label="管理员密码" path="adminPassword">
            <n-input v-model:value="setupForm.adminPassword" type="password"/>
          </n-form-item>
          <n-form-item label="管理员邮箱" path="adminEmail">
            <n-input v-model:value="setupForm.adminEmail"/>
          </n-form-item>
        </n-form>

        <!-- 操作按钮 -->
        <div class="setup-actions">
          <n-button v-if="currentStep > 0" @click="prevStep">上一步</n-button>
          <n-button v-if="currentStep < 2" type="primary" @click="nextStep">下一步</n-button>
          <n-button v-if="currentStep === 2" type="primary" @click="handleSubmit" :loading="submitting">
            完成设置
          </n-button>
        </div>
      </div>

      <!-- 完成提示 -->
      <div v-else class="setup-complete">
        <n-result status="success" title="初始化完成" description="网站已经完成初始化设置">
          <template #footer>
            <n-button type="primary" @click="goToHome">进入首页</n-button>
          </template>
        </n-result>
      </div>
    </n-card>
  </div>
</template>


<style scoped>
.setup-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f5f7fa;
  padding: 20px;
}

.setup-card {
  width: 100%;
  max-width: 800px;
}

.setup-form {
  margin-top: 40px;
}

.setup-actions {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.setup-complete {
  margin-top: 40px;
  text-align: center;
}
</style>
