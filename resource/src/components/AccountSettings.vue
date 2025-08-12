<script setup lang="ts">
import { ref, onMounted } from 'vue'
import BasicInfo from './account-settings/BasicInfo.vue'
import PasswordChange from './account-settings/PasswordChange.vue'
import AccountBinding from './account-settings/AccountBinding.vue'
import PrivacySettings from './account-settings/PrivacySettings.vue'
import type { UserInfo } from "@/utils/articleInterfaces";

const accountTabsRef = ref(null)

// 定义props
const props = defineProps<{
  userInfo: UserInfo
}>()

// 定义emits
const emit = defineEmits<{
  'user-info-updated': []
}>()

// 处理用户信息更新
const handleUserInfoUpdated = () => {
  emit('user-info-updated')
}

// 检查URL参数中的tab设置
const checkUrlMessages = () => {
  const urlParams = new URLSearchParams(window.location.search);
  const tab = urlParams.get('setting-tab')
  if (tab === 'account') {
    if (accountTabsRef.value) {
      accountTabsRef.value.checked = true
    }
  }
}

onMounted(() => {
  checkUrlMessages()
})
</script>

<template>
  <!-- Tab 导航 -->
  <div class="w-full">
    <div class="tabs tabs-lift">
      <input type="radio" name="account_tabs" class="tab" aria-label="基本信息" checked />
      <div class="tab-content bg-base-100 border-base-300 p-6">
        <BasicInfo :user-info="userInfo" @user-info-updated="handleUserInfoUpdated" />
      </div>

      <input type="radio" name="account_tabs" class="tab" aria-label="修改密码" />
      <div class="tab-content bg-base-100 border-base-300 p-6">
        <PasswordChange />
      </div>

      <input ref="accountTabsRef" type="radio" name="account_tabs" class="tab" aria-label="账号绑定" />
      <div class="tab-content bg-base-100 border-base-300 p-6">
        <AccountBinding />
      </div>

      <input type="radio" name="account_tabs" class="tab" aria-label="隐私设置" />
      <div class="tab-content bg-base-100 border-base-300 p-6">
        <PrivacySettings />
      </div>
    </div>
  </div>
</template>


<style scoped></style>
