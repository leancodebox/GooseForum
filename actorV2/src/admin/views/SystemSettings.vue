<template>
  <div>
    <n-tabs type="line" animated>
      <n-tab-pane name="basic" tab="基本设置">
        <n-form
          ref="basicFormRef"
          :model="basicForm"
          :rules="basicRules"
          label-placement="left"
          label-width="auto"
          require-mark-placement="right-hanging"
        >
          <n-form-item path="siteName" label="网站名称">
            <n-input v-model:value="basicForm.siteName" placeholder="请输入网站名称" />
          </n-form-item>
          <n-form-item path="siteDescription" label="网站描述">
            <n-input
              v-model:value="basicForm.siteDescription"
              type="textarea"
              placeholder="请输入网站描述"
            />
          </n-form-item>
          <n-form-item path="siteKeywords" label="网站关键词">
            <n-input v-model:value="basicForm.siteKeywords" placeholder="请输入网站关键词，用逗号分隔" />
          </n-form-item>
          <n-form-item path="siteUrl" label="网站地址">
            <n-input v-model:value="basicForm.siteUrl" placeholder="请输入网站地址" />
          </n-form-item>
          <n-form-item path="siteIcp" label="ICP备案号">
            <n-input v-model:value="basicForm.siteIcp" placeholder="请输入ICP备案号" />
          </n-form-item>
          <n-form-item path="siteCopyright" label="版权信息">
            <n-input v-model:value="basicForm.siteCopyright" placeholder="请输入版权信息" />
          </n-form-item>
          <n-form-item path="footer" label="Footer设置">
            <n-input v-model:value="basicForm.footer" placeholder="请输入版权信息" />
          </n-form-item>
        </n-form>
        <div class="action-buttons">
          <n-button type="primary" @click="handleSaveBasic">保存设置</n-button>
        </div>
      </n-tab-pane>

      <n-tab-pane name="content" tab="内容设置">
        <n-form
          ref="contentFormRef"
          :model="contentForm"
          label-placement="left"
          label-width="auto"
          require-mark-placement="right-hanging"
        >
          <n-form-item path="postReview" label="帖子审核">
            <n-switch v-model:value="contentForm.postReview" />
          </n-form-item>
          <n-form-item path="commentReview" label="评论审核">
            <n-switch v-model:value="contentForm.commentReview" />
          </n-form-item>
          <n-form-item path="postsPerPage" label="每页帖子数">
            <n-input-number v-model:value="contentForm.postsPerPage" :min="1" :max="100" />
          </n-form-item>
          <n-form-item path="commentsPerPage" label="每页评论数">
            <n-input-number v-model:value="contentForm.commentsPerPage" :min="1" :max="100" />
          </n-form-item>
          <n-form-item path="allowRegistration" label="允许注册">
            <n-switch v-model:value="contentForm.allowRegistration" />
          </n-form-item>
          <n-form-item path="allowAnonymousComment" label="允许匿名评论">
            <n-switch v-model:value="contentForm.allowAnonymousComment" />
          </n-form-item>
        </n-form>
        <div class="action-buttons">
          <n-button type="primary" @click="handleSaveContent">保存设置</n-button>
        </div>
      </n-tab-pane>

      <n-tab-pane name="upload" tab="上传设置">
        <n-form
          ref="uploadFormRef"
          :model="uploadForm"
          label-placement="left"
          label-width="auto"
          require-mark-placement="right-hanging"
        >
          <n-form-item path="uploadMaxSize" label="最大上传大小(MB)">
            <n-input-number v-model:value="uploadForm.uploadMaxSize" :min="1" :max="50" />
          </n-form-item>
          <n-form-item path="allowedFileTypes" label="允许的文件类型">
            <n-select
              v-model:value="uploadForm.allowedFileTypes"
              multiple
              :options="fileTypeOptions"
              placeholder="请选择允许上传的文件类型"
            />
          </n-form-item>
          <n-form-item path="imageMaxWidth" label="图片最大宽度">
            <n-input-number v-model:value="uploadForm.imageMaxWidth" :min="100" :max="5000" />
          </n-form-item>
          <n-form-item path="imageMaxHeight" label="图片最大高度">
            <n-input-number v-model:value="uploadForm.imageMaxHeight" :min="100" :max="5000" />
          </n-form-item>
          <n-form-item path="imageCompression" label="图片压缩质量">
            <n-slider
              v-model:value="uploadForm.imageCompression"
              :min="0"
              :max="100"
              :step="1"
              :tooltip="true"
            />
          </n-form-item>
        </n-form>
        <div class="action-buttons">
          <n-button type="primary" @click="handleSaveUpload">保存设置</n-button>
        </div>
      </n-tab-pane>

      <n-tab-pane name="security" tab="安全设置">
        <n-form
          ref="securityFormRef"
          :model="securityForm"
          label-placement="left"
          label-width="auto"
          require-mark-placement="right-hanging"
        >
          <n-form-item path="enableCaptcha" label="启用验证码">
            <n-switch v-model:value="securityForm.enableCaptcha" />
          </n-form-item>
          <n-form-item path="loginAttempts" label="最大登录尝试次数">
            <n-input-number v-model:value="securityForm.loginAttempts" :min="1" :max="10" />
          </n-form-item>
          <n-form-item path="lockTime" label="锁定时间(分钟)">
            <n-input-number v-model:value="securityForm.lockTime" :min="1" :max="1440" />
          </n-form-item>
          <n-form-item path="sessionTimeout" label="会话超时时间(分钟)">
            <n-input-number v-model:value="securityForm.sessionTimeout" :min="1" :max="1440" />
          </n-form-item>
          <n-form-item path="enableSensitiveWordFilter" label="启用敏感词过滤">
            <n-switch v-model:value="securityForm.enableSensitiveWordFilter" />
          </n-form-item>
          <n-form-item path="sensitiveWords" label="敏感词列表">
            <n-input
              v-model:value="securityForm.sensitiveWords"
              type="textarea"
              placeholder="请输入敏感词，一行一个"
              :autosize="{ minRows: 3, maxRows: 10 }"
            />
          </n-form-item>
        </n-form>
        <div class="action-buttons">
          <n-button type="primary" @click="handleSaveSecurity">保存设置</n-button>
        </div>
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useMessage } from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'

const message = useMessage()
const basicFormRef = ref<FormInst | null>(null)
const contentFormRef = ref<FormInst | null>(null)
const uploadFormRef = ref<FormInst | null>(null)
const securityFormRef = ref<FormInst | null>(null)

// 基本设置表单
const basicForm = reactive({
  siteName: 'GooseForum',
  siteDescription: '一个简单而强大的论坛系统',
  siteKeywords: '论坛,社区,讨论,分享',
  siteUrl: 'https://example.com',
  siteIcp: '京ICP备12345678号',
  siteCopyright: 'Copyright © 2023 GooseForum. All Rights Reserved.',
  footer :[]
})

// 基本设置验证规则
const basicRules: FormRules = {
  siteName: [
    { required: true, message: '请输入网站名称', trigger: 'blur' }
  ],
  siteUrl: [
    { required: true, message: '请输入网站地址', trigger: 'blur' },
    { type: 'url', message: '请输入有效的网址', trigger: 'blur' }
  ]
}

// 内容设置表单
const contentForm = reactive({
  postReview: true,
  commentReview: true,
  postsPerPage: 20,
  commentsPerPage: 30,
  allowRegistration: true,
  allowAnonymousComment: false
})

// 上传设置表单
const uploadForm = reactive({
  uploadMaxSize: 10,
  allowedFileTypes: ['image/jpeg', 'image/png', 'image/gif'],
  imageMaxWidth: 1920,
  imageMaxHeight: 1080,
  imageCompression: 80
})

// 文件类型选项
const fileTypeOptions = [
  { label: 'JPEG图片', value: 'image/jpeg' },
  { label: 'PNG图片', value: 'image/png' },
  { label: 'GIF图片', value: 'image/gif' },
  { label: 'PDF文档', value: 'application/pdf' },
  { label: 'Word文档', value: 'application/msword' },
  { label: 'Excel表格', value: 'application/vnd.ms-excel' },
  { label: 'ZIP压缩包', value: 'application/zip' }
]

// 安全设置表单
const securityForm = reactive({
  enableCaptcha: true,
  loginAttempts: 5,
  lockTime: 30,
  sessionTimeout: 120,
  enableSensitiveWordFilter: true,
  sensitiveWords: '敏感词1\n敏感词2\n敏感词3'
})

// 保存基本设置
const handleSaveBasic = () => {
  basicFormRef.value?.validate((errors) => {
    if (!errors) {
      message.success('基本设置保存成功')
      // 实际应用中应该发送请求保存设置
    }
  })
}

// 保存内容设置
const handleSaveContent = () => {
  message.success('内容设置保存成功')
  // 实际应用中应该发送请求保存设置
}

// 保存上传设置
const handleSaveUpload = () => {
  message.success('上传设置保存成功')
  // 实际应用中应该发送请求保存设置
}

// 保存安全设置
const handleSaveSecurity = () => {
  message.success('安全设置保存成功')
  // 实际应用中应该发送请求保存设置
}
</script>

<style scoped>
.action-buttons {
  margin-top: 24px;
}
</style>
