<script setup>
import { NFlex } from 'naive-ui'
import { useUserStore } from '@/modules/user'
import UserInfoCard from '@/components/UserInfoCard.vue'

const userStore = useUserStore()
</script>

<template>
  <n-flex vertical>
    <user-info-card :user-id="userStore.userInfo.userId"/>
  </n-flex>
</template>
