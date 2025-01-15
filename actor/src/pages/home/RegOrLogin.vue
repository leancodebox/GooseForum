<script setup>
import {NButton, NCard, NForm, NFormItemRow, NFlex, NInput, NTabPane, NTabs, useMessage, NImage} from "naive-ui"
import {ref, onMounted} from "vue"
import {useUserStore} from "@/modules/user"
import {login, reg, getCaptcha, getUserInfo} from "@/service/request"
import router from "@/route/router"
import {useRoute} from 'vue-router'

const message = useMessage()
const route = useRoute()

let loginInfo = ref({
  email: "",
  password: "",
  captchaId: "",
  captchaCode: ""
})

let regInfo = ref({
  email: "",
  username: "",
  password: "",
  repeatPassword: "",
  captchaId: "",
  captchaCode: ""
})

// 验证码相关
const captchaImg = ref("")
const captchaId = ref("")

// 获取验证码
async function refreshCaptcha() {
  try {
    const res = await getCaptcha()
    captchaImg.value = res.result.captchaImg
    // 同时更新登录和注册表单的 captchaId
    loginInfo.value.captchaId = res.result.captchaId
    regInfo.value.captchaId = res.result.captchaId
  } catch (error) {
    console.error('Failed to get captcha:', error)
  }
}

// 组件挂载时获取验证码
onMounted(() => {
  refreshCaptcha()
})

const userStore = useUserStore()

async function loginAction() {
  try {
    if (!loginInfo.value.captchaId || !loginInfo.value.captchaCode) {
      message.error('请输入验证码')
      return
    }
    let res = await login(
      loginInfo.value.email,
      loginInfo.value.password,
      loginInfo.value.captchaId,
      loginInfo.value.captchaCode
    )

    // 只保存 token
    userStore.setToken(res.result.token)

    // 获取用户信息
    const userInfoRes = await getUserInfo()
    userStore.setUserInfo(userInfoRes.result)

    message.success('登录成功')

    const redirect = route.query.redirect || '/home/bbs/articlesList'
    router.push(redirect)
  } catch (error) {
    console.error('Login failed:', error)
    refreshCaptcha()
  }
}

async function regAction() {
  try {
    if (regInfo.value.password !== regInfo.value.repeatPassword) {
      message.error("两次密码不相等")
      return
    }
    if (!regInfo.value.captchaId || !regInfo.value.captchaCode) {
      message.error('请输入验证码')
      return
    }
    let res = await reg(
      regInfo.value.email,
      regInfo.value.username,
      regInfo.value.password,
      regInfo.value.captchaId,
      regInfo.value.captchaCode
    )
    message.success('注册成功')

    // 只保存 token
    userStore.setToken(res.result.token)

    // 获取用户信息
    const userInfoRes = await getUserInfo()
    userStore.setUserInfo(userInfoRes.result)

    const redirect = route.query.redirect || '/home/bbs/articlesList'
    router.push(redirect)
  } catch (error) {
    console.error('Registration failed:', error)
    refreshCaptcha()
  }
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
              <n-form-item-row label="验证码">
                <n-flex align="center" :wrap="false" style="width: 100%">
                  <n-input v-model:value="loginInfo.captchaCode" style="flex: 1"/>
                  <n-image
                    :src="captchaImg"
                    @click="refreshCaptcha"
                    style="cursor: pointer; flex-shrink: 0;"
                    :preview-disabled="true"
                  />
                </n-flex>
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
              <n-form-item-row label="验证码">
                <n-flex align="center" :wrap="false" style="width: 100%">
                  <n-input v-model:value="regInfo.captchaCode" style="flex: 1"/>
                  <n-image
                    :src="captchaImg"
                    @click="refreshCaptcha"
                    style="cursor: pointer; flex-shrink: 0;"
                    :preview-disabled="true"
                  />
                </n-flex>
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

<style scoped>
.n-image {
  width: 120px;
  height: 40px;
  margin-left: 12px;
  min-width: 120px;
}
</style>
