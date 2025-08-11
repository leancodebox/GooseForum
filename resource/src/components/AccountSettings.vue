<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import AvatarUpload from './AvatarUpload.vue'
import { changePassword, saveUserInfo, saveUserEmail, saveUserName, getOAuthBindings, unbindOAuth } from '@/utils/articleService.ts'
import type { UserInfo, OAuthBindings } from "@/utils/articleInterfaces";

// 提示消息状态
const showSuccessMessage = ref(false)
const showErrorMessage = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const accountTabsRef = ref(null)

// 检查URL参数中的消息
const checkUrlMessages = () => {
  const urlParams = new URLSearchParams(window.location.search);
  const success = urlParams.get('success');
  const error = urlParams.get('error');
  const tab = urlParams.get('setting-tab')
  if (tab === 'account') {
    if (accountTabsRef.value) {
      accountTabsRef.value.checked = true
    }
  }
  if (success === 'bind_success') {
    successMessage.value = 'OAuth账户绑定成功！'
    showSuccessMessage.value = true
    // 3秒后自动隐藏
    setTimeout(() => {
      showSuccessMessage.value = false
    }, 3000)
    // 重新获取绑定状态
    fetchOAuthBindings();
    // 清理success参数，保留其他参数
    urlParams.delete('success')
    const newUrl = urlParams.toString() ? `${window.location.pathname}?${urlParams.toString()}` : window.location.pathname
    window.history.replaceState({}, '', newUrl)
  } else if (error) {
    errorMessage.value = decodeURIComponent(error)
    showErrorMessage.value = true
    // 5秒后自动隐藏
    setTimeout(() => {
      showErrorMessage.value = false
    }, 5000)
    // 清理error参数，保留其他参数
    urlParams.delete('error')
    const newUrl = urlParams.toString() ? `${window.location.pathname}?${urlParams.toString()}` : window.location.pathname
    window.history.replaceState({}, '', newUrl)
  }
}

// 定义props
const props = defineProps<{
  userInfo: UserInfo
}>()

// 定义emits
const emit = defineEmits<{
  'user-info-updated': []
}>()

// 个人资料表单
const profileForm = ref<UserInfo>({
  avatarUrl: "",
  username: "",
  bio: "",
  email: "",
  nickname: "",
  signature: "",
  website: "",
  websiteName: "",
  externalInformation: {
    github: { link: '' },
    weibo: { link: '' },
    bilibili: { link: '' },
    twitter: { link: '' },
    linkedIn: { link: '' },
    zhihu: { link: '' },
  },
})

// 监听props变化，同步更新表单数据
watch(() => props.userInfo, (newUserInfo) => {
  if (newUserInfo) {
    profileForm.value = { ...newUserInfo };
  }
}, { immediate: true, deep: true })

onMounted(() => {
  // 初始化时如果有用户信息就同步到表单
  if (props.userInfo) {
    profileForm.value = { ...props.userInfo };
  }
})


// 密码修改表单
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 隐私设置
const privacySettings = reactive({
  showArticles: true,
  showFollowing: true,
  emailNotifications: true
})

// 密码修改状态
const changingPassword = ref(false)

// 用户名和邮箱编辑状态
const usernameEditing = ref(false)
const emailEditing = ref(false)
const usernameUpdating = ref(false)
const emailUpdating = ref(false)

// AccountSettings组件现在使用radio button tabs，不需要activeTab状态

// 更新个人资料
const updateProfile = async () => {
  try {
    const response = await saveUserInfo(
      profileForm.value.nickname,
      profileForm.value.email,
      profileForm.value.bio,
      profileForm.value.signature,
      profileForm.value.website,
      profileForm.value.websiteName,
      profileForm.value.externalInformation,
    )

    if (response.code === 0) {
      alert('个人资料更新成功');
      // 通知父组件重新获取用户信息
      emit('user-info-updated');
    } else {
      alert(`更新失败: ${response.message || '请重试'}`)
    }
  } catch (error) {
    console.error('更新失败:', error)
    alert('更新失败，请重试')
  }
}

// 修改密码
const updatePassword = async () => {
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    alert('新密码和确认密码不一致')
    return
  }

  // 验证新密码长度
  if (passwordForm.newPassword.length < 6) {
    alert('新密码长度不能少于6位')
    return
  }

  try {
    changingPassword.value = true
    const response = await changePassword(
      passwordForm.currentPassword,
      passwordForm.newPassword
    )

    if (response.code === 0) {
      alert('密码修改成功')
      // 重置表单
      Object.assign(passwordForm, {
        currentPassword: '',
        newPassword: '',
        confirmPassword: ''
      })
    } else {
      alert(`密码修改失败: ${response.message || '请重试'}`)
    }
  } catch (error) {
    console.error('修改密码失败:', error)
    alert('密码修改失败，请重试')
  } finally {
    changingPassword.value = false
  }
}

// 头像更新回调
const handleAvatarUpdated = (newAvatarUrl) => {
  profileForm.value.avatarUrl = newAvatarUrl
  // 头像更新后也通知父组件刷新用户信息
  emit('user-info-updated');
}

// 用户名编辑相关函数
const toggleUsernameEdit = () => {
  usernameEditing.value = !usernameEditing.value
  if (!usernameEditing.value) {
    // 取消编辑时恢复原值
    profileForm.value.username = props.userInfo.username
  }
}

const saveUsername = async () => {
  if (!profileForm.value.username.trim()) {
    alert('用户名不能为空')
    return
  }

  try {
    usernameUpdating.value = true
    const response = await saveUserName(profileForm.value.username)

    if (response.code === 0) {
      alert('用户名更新成功')
      usernameEditing.value = false
      emit('user-info-updated')
    } else {
      alert(`更新失败: ${response.message || '请重试'}`)
    }
  } catch (error) {
    console.error('更新用户名失败:', error)
    alert(`更新失败: ${error.message || '请重试'}`)
  } finally {
    usernameUpdating.value = false
  }
}

// 邮箱编辑相关函数
const toggleEmailEdit = () => {
  emailEditing.value = !emailEditing.value
  if (!emailEditing.value) {
    // 取消编辑时恢复原值
    profileForm.value.email = props.userInfo.email
  }
}

const saveEmail = async () => {
  if (!profileForm.value.email.trim()) {
    alert('邮箱不能为空')
    return
  }

  // 简单的邮箱格式验证
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(profileForm.value.email)) {
    alert('请输入有效的邮箱地址')
    return
  }

  try {
    emailUpdating.value = true
    const response = await saveUserEmail(profileForm.value.email)

    if (response.code === 0) {
      alert('邮箱更新成功')
      emailEditing.value = false
      emit('user-info-updated')
    } else {
      alert(`更新失败: ${response.message || '请重试'}`)
    }
  } catch (error) {
    console.error('更新邮箱失败:', error)
    alert(`更新失败: ${error.message || '请重试'}`)
  } finally {
    emailUpdating.value = false
  }
}

// 保存隐私设置
const savePrivacySettings = async () => {
  try {
    console.log('保存隐私设置:', privacySettings)
    // 这里应该调用实际的API
    alert('隐私设置保存成功')
  } catch (error) {
    console.error('保存失败:', error)
    alert('保存失败，请重试')
  }
}

// OAuth绑定相关状态
const oauthBindings = ref<OAuthBindings>({})
const loadingBindings = ref(false)

// 获取OAuth绑定状态
const fetchOAuthBindings = async () => {
  try {
    loadingBindings.value = true
    const response = await getOAuthBindings()
    if (response && response.result) {
      oauthBindings.value = response.result
    }
  } catch (error) {
    console.error('获取OAuth绑定状态失败:', error)
    showErrorMessage.value = true
    errorMessage.value = '获取绑定状态失败，请重试'
    setTimeout(() => {
      showErrorMessage.value = false
    }, 5000)
  } finally {
    loadingBindings.value = false
  }
}

// 绑定OAuth账户
const bindOAuth = (provider: string) => {
  // 直接跳转到登录URL（已登录状态下会自动进入绑定模式）
  window.location.href = `/api/auth/${provider}`
}

// 解绑OAuth账户
const unbindOAuthAccount = async (provider: string) => {
  if (!confirm(`确定要解绑${provider.toUpperCase()}账户吗？`)) {
    return
  }

  try {
    const response = await unbindOAuth(provider)
    if (response.code === 0) {
       successMessage.value = 'OAuth账户解绑成功！'
       showSuccessMessage.value = true
       setTimeout(() => {
         showSuccessMessage.value = false
       }, 3000)
      await fetchOAuthBindings() // 重新获取绑定状态
    } else {
      showErrorMessage.value = true
      errorMessage.value = response.message || '解绑失败，请重试'
      setTimeout(() => {
        showErrorMessage.value = false
      }, 5000)
    }
  } catch (error: any) {
    console.error('解绑失败:', error)
    showErrorMessage.value = true
    errorMessage.value = error.message || '解绑失败，请重试'
    setTimeout(() => {
      showErrorMessage.value = false
    }, 5000)
  }
}

// 组件挂载时获取OAuth绑定状态
onMounted(() => {
  // 获取OAuth绑定状态
  fetchOAuthBindings()
  // 检查URL消息
  checkUrlMessages()
})
</script>

<template>
  <!-- 提示消息 -->
  <div class="fixed top-20 right-4 z-50 space-y-2">
    <!-- 成功消息 -->
    <div v-if="showSuccessMessage" class="alert alert-success shadow-lg max-w-sm">
      <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
          d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <span>{{ successMessage }}</span>
    </div>

    <!-- 错误消息 -->
    <div v-if="showErrorMessage" class="alert alert-error shadow-lg max-w-sm">
      <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
          d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <span>绑定失败: {{ errorMessage }}</span>
    </div>
  </div>

  <!-- Tab 导航 -->
  <div class="w-full">
    <div class="tabs tabs-lift">
      <input type="radio" name="account_tabs" class="tab" aria-label="基本信息" checked />
      <div class="tab-content bg-base-100 border-base-300 p-6">
        <form @submit.prevent="updateProfile" class="grid grid-cols-1 gap-6">
          <div class="form-control">
            <label class="label">
              <span class="label-text font-normal">头像设置</span>
            </label>
            <AvatarUpload :current-avatar="profileForm.avatarUrl" @avatar-updated="handleAvatarUpdated" />
          </div>
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div class="form-control ">
              <label class="floating-label join w-full">
                <span>用户名</span>
                <input v-model="profileForm.username" type="text" class="input input-bordered w-full"
                  placeholder="请输入用户名" :disabled="!usernameEditing" />
                <button v-if="!usernameEditing" @click="toggleUsernameEdit" class="btn btn-primary join-item">更改
                </button>
                <button v-else @click="saveUsername" class="btn btn-success join-item" :disabled="usernameUpdating">
                  <span v-if="usernameUpdating" class="loading loading-spinner loading-sm"></span>
                  {{ usernameUpdating ? '保存中...' : '保存' }}
                </button>
                <button v-if="usernameEditing && !usernameUpdating" @click="toggleUsernameEdit"
                  class="btn btn-warning join-item">取消
                </button>
              </label>
            </div>
            <div class="form-control">
              <label class="floating-label join w-full">
                <span>邮箱</span>
                <input v-model="profileForm.email" type="email" class="input input-bordered w-full" placeholder="请输入邮箱"
                  :disabled="!emailEditing" />
                <button v-if="!emailEditing" @click="toggleEmailEdit" class="btn btn-primary join-item">更改</button>
                <button v-else @click="saveEmail" class="btn btn-success join-item" :disabled="emailUpdating">
                  <span v-if="emailUpdating" class="loading loading-spinner loading-sm"></span>
                  {{ emailUpdating ? '保存中...' : '保存' }}
                </button>
                <button v-if="emailEditing && !emailUpdating" @click="toggleEmailEdit"
                  class="btn  btn-warning  join-item">取消
                </button>
              </label>
            </div>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div class="form-control">
              <label class="floating-label">
                <span>昵称</span>
                <input v-model="profileForm.nickname" type="text" class="input input-bordered w-full"
                  placeholder="请输入昵称" />
              </label>
            </div>
            <div class="form-control">
              <label class="floating-label">
                <span>个性签名</span>
                <input v-model="profileForm.signature" type="text" class="input input-bordered w-full"
                  placeholder="请输入个性签名" />
              </label>
            </div>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div class="form-control">
              <label class="floating-label">
                <span>网站名</span>
                <input v-model="profileForm.websiteName" type="text" class="input input-bordered w-full"
                  placeholder="请输入网站名" />
              </label>
            </div>
            <div class="form-control">
              <label class="floating-label">
                <span>网站地址</span>
                <input v-model="profileForm.website" type="text" class="input input-bordered w-full"
                  placeholder="请输入网站地址" />
              </label>
            </div>
          </div>
          <details tabindex="0" class="collapse collapse-arrow bg-base-100 border-base-300 border">
            <summary class="collapse-title font-normal">外站配置</summary>
            <div class="collapse-content text-sm">
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div class="form-control">
                  <label class="floating-label">
                    <span>Github</span>
                    <input v-model="profileForm.externalInformation.github.link" type="text"
                      class="input input-bordered w-full" placeholder="请输入Github地址" />
                  </label>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>BiliBili</span>
                    <input v-model="profileForm.externalInformation.bilibili.link" type="text"
                      class="input input-bordered w-full" placeholder="请输入BiliBili地址" />
                  </label>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>Weibo</span>
                    <input v-model="profileForm.externalInformation.weibo.link" type="text"
                      class="input input-bordered w-full" placeholder="请输入Weibo地址" />
                  </label>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>Twitter</span>
                    <input v-model="profileForm.externalInformation.twitter.link" type="text"
                      class="input input-bordered w-full" placeholder="请输入Twitter地址" />
                  </label>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>zhihu</span>
                    <input v-model="profileForm.externalInformation.zhihu.link" type="text"
                      class="input input-bordered w-full" placeholder="请输入Zhihu地址" />
                  </label>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>linkedIn</span>
                    <input v-model="profileForm.externalInformation.linkedIn.link" type="text"
                      class="input input-bordered w-full" placeholder="请输入LinkedIn地址" />
                  </label>
                </div>
              </div>
            </div>
          </details>

          <div class="form-control">
            <label class="label">
              <span class="label-text font-normal">个人简介</span>
              <span class="label-text-alt">{{ profileForm.bio?.length || 0 }}/200</span>
            </label>
            <textarea v-model="profileForm.bio" class="textarea textarea-bordered w-full" rows="4"
              placeholder="介绍一下自己..." maxlength="200"></textarea>
          </div>

          <div class="flex justify-end">
            <button type="submit" class="btn btn-primary">保存基本信息</button>
          </div>
        </form>
      </div>

      <input type="radio" name="account_tabs" class="tab" aria-label="修改密码" />
      <div class="tab-content bg-base-100 border-base-300 p-6">
        <h3 class="card-title text-lg mb-6 border-b border-base-300 pb-3">修改密码</h3>
        <form @submit.prevent="updatePassword" class="grid grid-cols-1 gap-6">
          <div class="form-control">
            <label class="label">
              <span class="label-text font-normal">当前密码</span>
            </label>
            <input v-model="passwordForm.currentPassword" type="password" class="input input-bordered w-full"
              placeholder="请输入当前密码" required />
          </div>
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div class="form-control">
              <label class="label">
                <span class="label-text font-normal">新密码</span>
              </label>
              <input v-model="passwordForm.newPassword" type="password" class="input input-bordered w-full"
                placeholder="请输入新密码" required />
              <label class="label">
                <span class="label-text-alt">密码长度至少8位，包含字母和数字</span>
              </label>
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text font-normal">确认新密码</span>
              </label>
              <input v-model="passwordForm.confirmPassword" type="password" class="input input-bordered w-full"
                placeholder="请再次输入新密码" required />
            </div>
          </div>
          <div class="flex justify-end">
            <button type="submit" class="btn btn-secondary min-w-32" :disabled="changingPassword">
              <span v-if="changingPassword" class="loading loading-spinner loading-sm"></span>
              {{ changingPassword ? '修改中...' : '修改密码' }}
            </button>
          </div>
        </form>
      </div>

      <input ref="accountTabsRef" type="radio" name="account_tabs" class="tab" aria-label="账号绑定" />
      <div class="tab-content bg-base-100 border-base-300 p-6">
        <h3 class="card-title text-lg mb-6 border-b border-base-300 pb-3">第三方账号绑定</h3>
        <div v-if="loadingBindings" class="flex justify-center py-8">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
        <div v-else class="grid grid-cols-1 gap-6">
          <!-- GitHub绑定 -->
          <div class="card bg-base-200 shadow-sm">
            <div class="card-body p-4">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 bg-gray-900 rounded-full flex items-center justify-center">
                    <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd"
                        d="M10 0C4.477 0 0 4.484 0 10.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0110 4.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.203 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.942.359.31.678.921.678 1.856 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0020 10.017C20 4.484 15.522 0 10 0z"
                        clip-rule="evenodd" />
                    </svg>
                  </div>
                  <div>
                    <h4 class="font-medium">GitHub</h4>
                    <p class="text-sm text-base-content/60">绑定GitHub账户</p>
                  </div>
                </div>
                <div v-if="oauthBindings.github?.bound" class="flex items-center gap-2">
                  <span class="badge badge-success">已绑定</span>
                  <button @click="unbindOAuthAccount('github')" class="btn btn-sm btn-error">解绑</button>
                </div>
                <button v-else @click="bindOAuth('github')" class="btn btn-sm btn-primary">绑定</button>
              </div>
            </div>
          </div>

          <!-- Google绑定 -->
          <div class="card bg-base-200 shadow-sm" v-if="false">
            <div class="card-body p-4">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 bg-white rounded-full flex items-center justify-center border">
                    <svg class="w-5 h-5" viewBox="0 0 24 24">
                      <path fill="#4285F4"
                        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
                      <path fill="#34A853"
                        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                      <path fill="#FBBC05"
                        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
                      <path fill="#EA4335"
                        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
                    </svg>
                  </div>
                  <div>
                    <h4 class="font-medium">Google</h4>
                    <p class="text-sm text-base-content/60">绑定Google账户</p>
                  </div>
                </div>
                <div v-if="oauthBindings.google?.bound" class="flex items-center gap-2">
                  <span class="badge badge-success">已绑定</span>
                  <button @click="unbindOAuthAccount('google')" class="btn btn-sm btn-error">解绑</button>
                </div>
                <button v-else @click="bindOAuth('google')" class="btn btn-sm btn-primary">绑定</button>
              </div>
            </div>
          </div>

          <div class="alert alert-info">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
              class="stroke-current shrink-0 w-6 h-6">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <span>绑定第三方账户后，您可以使用这些账户快速登录。解绑后将无法使用对应账户登录。</span>
          </div>
        </div>
      </div>

      <input type="radio" name="account_tabs" class="tab" aria-label="隐私设置" />
      <div class="tab-content bg-base-100 border-base-300 p-6">
        <h3 class="card-title text-lg mb-6 border-b border-base-300 pb-3">隐私设置</h3>
        <div class="grid grid-cols-1 gap-6">
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input v-model="privacySettings.showArticles" type="checkbox" class="toggle toggle-primary" />
              <div>
                <span class="label-text font-normal">公开文章列表</span>
                <div class="text-sm text-base-content/60">允许其他用户查看我发布的文章</div>
              </div>
            </label>
          </div>

          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input v-model="privacySettings.showFollowing" type="checkbox" class="toggle toggle-primary" />
              <div>
                <span class="label-text font-normal">公开关注列表</span>
                <div class="text-sm text-base-content/60">允许其他用户查看我的关注和粉丝</div>
              </div>
            </label>
          </div>

          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input v-model="privacySettings.emailNotifications" type="checkbox" class="toggle toggle-primary" />
              <div>
                <span class="label-text font-normal">邮件通知</span>
                <div class="text-sm text-base-content/60">接收评论、点赞等活动的邮件提醒</div>
              </div>
            </label>
          </div>

          <div class="flex justify-end">
            <button @click="savePrivacySettings" class="btn btn-primary">保存隐私设置</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>


<style scoped></style>
