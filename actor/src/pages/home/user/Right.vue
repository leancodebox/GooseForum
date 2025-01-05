<script setup>
import { NCard, NFlex } from 'naive-ui'
import { ref, onMounted } from 'vue'
import { getUserInfoShow } from '@/service/request'
import { useUserStore } from '@/modules/user'
import UserInfoCard from '@/components/UserInfoCard.vue'

const userStore = useUserStore()
const userInfo = ref({
  username: '',
  userId: 0,
  avatarUrl: '',
  signature: '未填写',
  articleCount: 0,
  prestige: 0,
})

async function getUserInfo() {
  try {
    const res = await getUserInfoShow(userStore.userInfo.userId)
    if (res.code === 0 && res.result) {
      userInfo.value = {
        username: res.result.username || '',
        userId: res.result.userId || 0,
        avatarUrl: res.result.avatarUrl || '',
        signature: res.result.signature || '未填写',
        articleCount: res.result.articleCount || 0,
        prestige: res.result.prestige || 0,
      }
    }
  } catch (err) {
    console.error('获取用户信息失败:', err)
  }
}

onMounted(() => {
  getUserInfo()
})
</script>

<template>
  <n-flex vertical>
      <user-info-card :user-info="userInfo"/>
  </n-flex>
</template>
