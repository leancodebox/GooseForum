<script setup>
import {NButton, NCard, NEllipsis, NFlex} from "naive-ui"
import {h, ref} from 'vue'
import UserInfoCard from "@/components/UserInfoCard.vue";
import {useUserStore} from "@/modules/user.js";

let options = [
  {
    label: () =>
        h(NEllipsis, null, {default: () => '我的博文'}),
    key: '2'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '我的回复'}),
    key: '3'
  }
]
const userStore = useUserStore()
</script>

<template>
  <n-card :bordered="false">
    <n-flex :justify="'center'" class="responsive-container">
      <!-- 左侧导航按钮 -->
      <n-flex vertical class="nav-buttons">
        <n-button class="nav-button">我的博文</n-button>
        <n-button class="nav-button">我的回复</n-button>
      </n-flex>
      
      <!-- 中间内容区 -->
      <n-card class="content-area">
        <router-view></router-view>
      </n-card>
      
      <!-- 右侧用户信息卡片 -->
      <n-flex vertical class="user-info-section">
        <user-info-card :user-id="userStore.userInfo.userId"/>
      </n-flex>
    </n-flex>
  </n-card>
</template>

<style scoped>
.responsive-container {
  gap: 16px;
  flex-wrap: wrap;
}

.nav-buttons {
  min-width: 140px;
  gap: 8px;
}

.nav-button {
  width: 100%;
}

.content-area {
  flex: 1;
  min-width: 300px;
  max-width: 800px;
}

.user-info-section {
  min-width: 240px;
}

/* 平板设备 */
@media screen and (max-width: 1024px) {
  .responsive-container {
    flex-direction: column;
  }
  
  .nav-buttons {
    flex-direction: row;
    min-width: 100%;
    justify-content: center;
    gap: 16px;
  }
  
  .nav-button {
    width: auto;
  }
  
  .content-area {
    min-width: 100%;
  }
  
  .user-info-section {
    min-width: 100%;
  }
}

/* 移动设备 */
@media screen and (max-width: 640px) {
  .nav-buttons {
    flex-direction: column;
  }
  
  .nav-button {
    width: 100%;
  }
}
</style>
