<script setup>
import {NButton, NCard, NForm, NFormItemRow, NGrid, NGridItem, NInput, NTabPane, NTabs} from "naive-ui"
import {ref} from "vue"
import {useUserStore} from "@/modules/user";

let loginUsername = ref("")
let loginPassword = ref("")
const userStore = useUserStore()

function login() {
  userStore.login(loginUsername.value, loginPassword.value)
}
function regAction(){
    userStore.login(loginUsername.value, loginPassword.value)
}
</script>

<template>
  <n-grid :cols="3" item-responsive style="height: 100%">
    <n-grid-item span="0 660:1"/>
    <n-grid-item span="3 660:1" style="padding-top:64px ">
      <n-card style="">
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
                <n-input v-model:value="loginUsername"/>
              </n-form-item-row>
              <n-form-item-row label="密码">
                <n-input v-model:value="loginPassword"/>
              </n-form-item-row>
            </n-form>
            <n-button type="primary" @click="login" block secondary strong>
              登录
            </n-button>
          </n-tab-pane>
          <n-tab-pane name="signup" tab="注册">
            <n-form>
              <n-form-item-row label="用户名">
                <n-input/>
              </n-form-item-row>
              <n-form-item-row label="密码">
                <n-input/>
              </n-form-item-row>
              <n-form-item-row label="重复密码">
                <n-input/>
              </n-form-item-row>
            </n-form>
            <n-button type="primary" block secondary strong>
              注册
            </n-button>
          </n-tab-pane>
        </n-tabs>
      </n-card>
    </n-grid-item>
  </n-grid>
</template>
<style>

</style>