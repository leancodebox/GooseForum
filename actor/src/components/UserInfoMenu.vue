<script setup>

import {NAvatar, NButton, NDropdown, useMessage} from "naive-ui";
import {useUserStore} from "@/modules/user";
import router from "@/route/router"
import {computed} from "vue";

const userStore = useUserStore()
const message = useMessage()

const options = [
  {label: "个人中心", key: "userInfo"},
  {label: "编辑资料", key: "userEdit"},
  {label: "退出登录", key: "logout"},
]

const handleSelect = function (key) {
  switch(key) {
    case "userInfo":
      router.push({name: 'userInfo'})
      break
    case "userEdit":
      router.push({name: "userEdit"})
      break
    case "logout":
      handleLogout()
      break
  }
}

const handleLogout = () => {
  userStore.clearUserInfo()
  message.success('已退出登录')
  
  router.push('/home/bbs/bbs')
}

const handleLoginOrRegister = () => {
  router.push({
    path: '/home/regOrLogin',
    query: { redirect: router.currentRoute.value.fullPath }
  })
}

const truncateUsername = computed(() => {
  const maxLength = 3;
  const username = userStore.userInfo.username;
  return username.length > maxLength
      ? `${username.slice(0, maxLength)}...`
      : username;
})
</script>

<template>
  <n-button v-if="userStore.userInfo.username===''" @click="handleLoginOrRegister">
    登录/注册
  </n-button>
  <n-dropdown trigger="hover" :options="options" @select="handleSelect" v-else>
    <n-button text>
      <n-avatar
          round
          :size="32"
          v-if="!userStore.userInfo.avatarUrl"
      >
        {{ truncateUsername }}
      </n-avatar>
      <n-avatar
          round
          v-else
          :size="32"
          :src="userStore.userInfo.avatarUrl || '/api/assets/default-avatar.png'"
      />
      <div style="width: 5px"></div>
      {{ userStore.userInfo.username }}
    </n-button>
  </n-dropdown>
</template>
