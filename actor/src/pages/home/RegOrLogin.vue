<script setup>
import {NButton, NCard, NForm, NFormItemRow, NGrid, NGridItem,NFlex, NInput, NTabPane, NTabs, useMessage} from "naive-ui"
import {ref} from "vue"
import {useUserStore} from "@/modules/user";
import {login, reg} from "@/service/request";
import router from "@/route/router"

const message = useMessage()

let loginInfo = ref({
  email: "",
  password: "",
})

let regInfo = ref({
  email: "",
  username: "",
  password: "",
  repeatPassword: "",
})

const userStore = useUserStore()

async function loginAction() {
  let res = await login(loginInfo.value.email, loginInfo.value.password)
  userStore.login(res.result)
  router.push({path: "/home/bbs"})
}

async function regAction() {
  if (regInfo.value.password !== regInfo.value.repeatPassword) {
    message.error("两次密码不相等")
    return
  }
  let res = await reg(regInfo.value.email, regInfo.value.username, regInfo.value.password)
  userStore.login(res.result)
  router.push({path: "/home/bbs"})
}
</script>

<template>
  <n-flex vertical style="margin-top: 120px">
    <n-flex justify="space-around">
      <n-card style="max-width: 400px;">
        <n-tabs
            class="card-tabs"
            default-value="signin"
            size="large"
            animated
            style="margin: 0 -4px"
            pane-style="padding-left: 4px; padding-right: 4px; box-sizing: border-box;"
        >
          <n-tab-pane name="signin" tab="登录">
            <n-form>
              <n-form-item-row label="用户名">
                <n-input v-model:value="loginInfo.email"/>
              </n-form-item-row>
              <n-form-item-row label="密码">
                <n-input v-model:value="loginInfo.password" type="password"/>
              </n-form-item-row>
            </n-form>
            <n-button type="primary" @click="loginAction" block secondary strong>
              登录
            </n-button>
          </n-tab-pane>
          <n-tab-pane name="signup" tab="注册">
            <n-form>
              <n-form-item-row label="邮箱">
                <n-input v-model:value="regInfo.email"/>
              </n-form-item-row>
              <n-form-item-row label="用户名">
                <n-input v-model:value="regInfo.username"/>
              </n-form-item-row>
              <n-form-item-row label="密码">
                <n-input v-model:value="regInfo.password" type="password"/>
              </n-form-item-row>
              <n-form-item-row label="重复密码">
                <n-input v-model:value="regInfo.repeatPassword" type="password"/>
              </n-form-item-row>
            </n-form>
            <n-button type="primary" @click="regAction" block secondary strong>
              注册
            </n-button>
          </n-tab-pane>
        </n-tabs>
      </n-card>
    </n-flex>
  </n-flex>
</template>
<style>

</style>
