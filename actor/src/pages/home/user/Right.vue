<script setup>
import {
  NAvatar,
  NCard,
  NDivider,
  NFlex,
  NStatistic,
  NButton
} from 'naive-ui'
import { ref, onMounted } from 'vue'
import { getUserInfoShow } from '@/service/request'
import { useUserStore } from '@/modules/user'

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
    <n-card style="min-width: 180px;">
      <n-flex vertical>
        <n-flex align="center" :wrap="false">
          <n-avatar
              round
              :size="60"
              v-if="!userInfo.avatarUrl"
          >
            {{ userInfo.username?.charAt(0) }}
          </n-avatar>
          <n-avatar
              round
              :size="60"
              v-else
              :src="userInfo.avatarUrl || '/api/assets/default-avatar.png'"
          />
          <n-flex vertical style="margin-left: 12px; flex: 1;">
            <span class="username">{{ userInfo.username }}</span>
            <span class="signature">{{ userInfo.signature }}</span>
          </n-flex>
        </n-flex>
        <n-divider style="margin: 16px 0"/>
        <n-flex justify="space-around">
          <n-statistic :value="userInfo.prestige">
            <template #label>
              <div class="stat-label">声望</div>
            </template>
          </n-statistic>
          <n-statistic :value="userInfo.articleCount">
            <template #label>
              <div class="stat-label">文章</div>
            </template>
          </n-statistic>
        </n-flex>
        <n-divider style="margin: 16px 0"/>
        <n-flex justify="space-around">
          <n-button secondary size="small">关注</n-button>
          <n-button secondary size="small">私信</n-button>
        </n-flex>
      </n-flex>
    </n-card>
  </n-flex>
</template>

<style scoped>
.username {
  font-size: 16px;
  font-weight: 500;
  line-height: 1.5;
  color: var(--n-text-color);
}

.signature {
  font-size: 13px;
  color: var(--n-text-color-3);
  line-height: 1.5;
}

.stat-label {
  font-size: 13px;
  color: var(--n-text-color-3);
}

:deep(.n-statistic-value) {
  font-size: 20px;
  font-weight: 500;
}
</style>
