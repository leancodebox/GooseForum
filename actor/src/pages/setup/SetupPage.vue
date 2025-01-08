<script setup>
import {ref, onMounted, watch} from 'vue'
import {useMessage} from 'naive-ui'
import {
  create,
  NButton,
  NCard,
  NForm,
  NFormItem,
  NInput,
  NResult,
  NStep,
  NSteps,
  NIcon,
  NSpace,
  NRadio,
  NRadioGroup,
} from 'naive-ui'
import {getSetupStatus, submitSetup} from '@/service/request4setup.js'
import {
  SettingsSharp,
  ServerSharp,
  PersonSharp,
  CheckmarkCircleSharp,
  ArrowBack,
  ArrowForward,
  Checkmark,
  Enter
} from '@vicons/ionicons5'

const message = useMessage()

const currentStep = ref(0)
const stepStatus = ref('process')
const isCompleted = ref(false)
const submitting = ref(false)

// 表单引用
const formRef1 = ref(null)
const formRef2 = ref(null)
const formRef3 = ref(null)

const setupForm = ref({
  siteName: '',
  siteDesc: '',
  dbType: 'mysql',
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

const nextStep = async () => {
  if (currentStep.value === 0) {
    try {
      await formRef1.value?.validate()
      currentStep.value++
    } catch (errors) {
      message.error('请完善基本信息')
    }
  } else if (currentStep.value === 1) {
    try {
      await formRef2.value?.validate()
      currentStep.value++
    } catch (errors) {
      message.error('请完善数据库配置')
    }
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
    // message.error('初始化失败：' + error.message)
  } finally {
    submitting.value = false
  }
}

const goToHome = () => {
  window.location.href = '/actor/'
}

// 监听数据库类型变化，重置相关字段
watch(() => setupForm.value.dbType, (newType) => {
  if (newType === 'sqlite') {
    setupForm.value.dbHost = ''
    setupForm.value.dbPort = ''
    setupForm.value.dbUser = ''
    setupForm.value.dbPassword = ''
    setupForm.value.dbName = 'storage/database.db'
  } else {
    setupForm.value.dbHost = 'localhost'
    setupForm.value.dbPort = '3306'
    setupForm.value.dbName = ''
  }
})
</script>
<template>
  <div class="setup-container">
    <n-card title="网站初始化设置" class="setup-card">
      <div class="setup-header">
        <n-steps :current="currentStep" :status="stepStatus" class="setup-steps">
          <n-step title="基本信息" description="设置网站基本信息"/>
          <n-step title="数据库配置" description="配置数据库连接"/>
          <n-step title="管理员账号" description="创建管理员账号"/>
        </n-steps>
      </div>

      <div class="setup-content">
        <div class="setup-form" v-if="!isCompleted">
          <!-- 第一步：基本信息 -->
          <n-form v-if="currentStep === 0" ref="formRef1" :model="setupForm" :rules="basicRules">
            <h3 class="form-title">基本信息设置</h3>
            <n-form-item label="网站名称" path="siteName">
              <n-input v-model:value="setupForm.siteName" placeholder="请输入网站名称"/>
            </n-form-item>
            <n-form-item label="网站描述" path="siteDesc">
              <n-input v-model:value="setupForm.siteDesc" type="textarea" placeholder="请输入网站描述"/>
            </n-form-item>
          </n-form>

          <!-- 第二步：数据库配置 -->
          <div v-if="currentStep === 1" class="scrollable-form">
            <n-form ref="formRef2" :model="setupForm" :rules="dbRules">
              <h3 class="form-title">数据库配置</h3>
              <n-form-item label="数据库类型" path="dbType">
                <n-radio-group v-model:value="setupForm.dbType" name="dbType">
                  <n-space>
                    <n-radio value="mysql">
                      <n-space align="center">
                        <span>MySQL</span>
                      </n-space>
                    </n-radio>
                    <n-radio value="sqlite">
                      <n-space align="center">
                        <span>SQLite</span>
                      </n-space>
                    </n-radio>
                  </n-space>
                </n-radio-group>
              </n-form-item>

              <template v-if="setupForm.dbType === 'mysql'">
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
              </template>

              <template v-else>
                <n-form-item label="数据库文件路径" path="dbName">
                  <n-input v-model:value="setupForm.dbName" placeholder="storage/database.db"/>
                </n-form-item>
              </template>
            </n-form>
          </div>

          <!-- 第三步：管理员账号 -->
          <n-form v-if="currentStep === 2" ref="formRef3" :model="setupForm" :rules="adminRules">
            <h3 class="form-title">管理员账号设置</h3>
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
            <n-button v-if="currentStep > 0" @click="prevStep" :disabled="submitting">
              <template #icon>
                <n-icon><arrow-back/></n-icon>
              </template>
              上一步
            </n-button>
            <n-button v-if="currentStep < 2" type="primary" @click="nextStep">
              下一步
              <template #icon>
                <n-icon><arrow-forward/></n-icon>
              </template>
            </n-button>
            <n-button v-if="currentStep === 2" type="primary" @click="handleSubmit" :loading="submitting">
              完成设置
              <template #icon>
                <n-icon><checkmark/></n-icon>
              </template>
            </n-button>
          </div>
        </div>

        <!-- 完成提示 -->
        <div v-else class="setup-complete">
          <n-result status="success" title="初始化完成" description="网站已经完成初始化设置">
            <template #icon>
              <div class="success-icon">
                <n-icon><checkmark-circle-sharp/></n-icon>
              </div>
            </template>
            <template #footer>
              <n-button type="primary" size="large" @click="goToHome">
                进入首页
                <template #icon>
                  <n-icon><enter/></n-icon>
                </template>
              </n-button>
            </template>
          </n-result>
        </div>
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
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  padding: 20px;
}

.setup-card {
  width: 100%;
  max-width: 800px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
  border-radius: 16px;
  overflow: hidden;
}

.setup-header {
  margin-bottom: 40px;
  padding: 0 20px;
  display: flex;
  justify-content: center;
}

.setup-steps {
  width: 80%;
  max-width: 600px;
}

.setup-content {
  padding: 0 40px;
}

.form-title {
  text-align: center;
  color: #18a058;
  font-size: 20px;
  margin: 0 0 32px;
  font-weight: 500;
}

.setup-form {
  margin-top: 20px;
  position: relative;
}

.scrollable-form {
  max-height: calc(100vh - 400px);
  overflow-y: auto;
  padding-right: 20px;
}

.form-header {
  text-align: center;
  margin-bottom: 32px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.header-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 16px;
  background: #18a058;
  color: white;
  margin-bottom: 16px;
}

.form-header h2 {
  margin: 0;
  color: #18a058;
  font-weight: 500;
}

.setup-actions {
  margin-top: 32px;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  position: sticky;
  bottom: 0;
  background: white;
  padding: 16px 0;
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.05);
}

.setup-complete {
  margin: 40px 0;
  text-align: center;
}

.success-icon {
  font-size: 64px;
  color: #18a058;
}

:deep(.n-form-item) {
  max-width: 600px;
  margin: 0 auto 24px;
}

:deep(.n-card-header) {
  text-align: center;
  font-size: 24px;
  padding: 24px;
  border-bottom: 1px solid #eee;
  background: #f9f9f9;
}

:deep(.n-steps) {
  justify-content: center;
}

:deep(.n-step) {
  flex: none !important;
  min-width: 120px;
}

:deep(.n-button .n-icon) {
  margin: 0 4px;
}

:deep(.n-radio-group .n-icon) {
  margin-right: 4px;
}

/* 自定义滚动条样式 */
.scrollable-form::-webkit-scrollbar {
  width: 8px;
}

.scrollable-form::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.scrollable-form::-webkit-scrollbar-thumb {
  background: #888;
  border-radius: 4px;
}

.scrollable-form::-webkit-scrollbar-thumb:hover {
  background: #555;
}

/* 调整表单项间距 */
:deep(.n-form-item:not(:last-child)) {
  margin-bottom: 24px;
}
</style>
