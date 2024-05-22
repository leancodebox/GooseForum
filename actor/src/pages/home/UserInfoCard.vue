<script setup>

import {NAvatar, NButton, NDropdown, useMessage} from "naive-ui";
import {useUserStore} from "@/modules/user";
import router from "@/route/router"

const userStore = useUserStore()
const message = useMessage()
let options = [
  {label: "我的发布", key: "我的发布"},
  {label: "个人中心", key: "个人中心"},
  {label: "编辑资料", key: "编辑资料"},
  {label: "账号退出", key: "账号退出"},
]
let handleSelect = function (key) {
  message.info(String(key))
}
let loginOrReg = function (){
  router.push({ path: '/login' })
}
</script>

<template>
  <n-button v-if="userStore.userInfo.username===''" @click="loginOrReg"> 登录/注册</n-button>
  <n-dropdown trigger="hover" :options="options" @select="handleSelect" v-else>
    <n-button text>
      <n-avatar
          round
          :size="36"
      >
        {{ userStore.userInfo.username }}
      </n-avatar>
      <div style="width: 5px"></div>
      {{ userStore.userInfo.username }}
    </n-button>
  </n-dropdown>
</template>
