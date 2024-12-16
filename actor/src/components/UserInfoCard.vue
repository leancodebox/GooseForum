<script setup>

import {NAvatar, NButton, NDropdown, useMessage} from "naive-ui";
import {useUserStore} from "@/modules/user";
import router from "@/route/router"
import {computed} from "vue";

const userStore = useUserStore()
const message = useMessage()
let options = [
  {label: "个人中心", key: "userInfo"},
  {label: "编辑资料", key: "userEdit"},
  {label: "账号退出", key: "账号退出"},
]
let handleSelect = function (key) {
  message.info(String(key))
  if (key === "userInfo") {
    router.push({name: 'userInfo'})
  } else if (key === "userEdit"){
    router.push({name:"userEdit"})
  }
  //
}
let loginOrReg = function () {
  router.push({path: '/login'})
}
const truncateUsername = computed(() => {
  const maxLength = 3; // 设定一个最大长度
  const username = userStore.userInfo.username;
  return username.length > maxLength
      ? `${username.slice(0, maxLength)}...`
      : username;
});
</script>

<template>
  <n-button v-if="userStore.userInfo.username===''" @click="loginOrReg"> 登录/注册</n-button>
  <n-dropdown trigger="hover" :options="options" @select="handleSelect" v-else>
    <n-button text>
      <n-avatar
          round
          :size="32"
      >
        {{ truncateUsername }}
      </n-avatar>
      <div style="width: 5px"></div>
      {{ userStore.userInfo.username }}
    </n-button>
  </n-dropdown>
</template>
