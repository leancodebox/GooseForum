<script setup lang="ts">
import {onMounted, reactive, ref, watch} from 'vue'
import AvatarUpload from './AvatarUpload.vue'
import {changePassword, saveUserInfo, saveUserEmail, saveUserName} from '@/utils/articleService.ts'
import type {UserInfo} from "@/utils/articleInterfaces";

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
    github: {link: ''},
    weibo: {link: ''},
    bilibili: {link: ''},
    twitter: {link: ''},
    linkedIn: {link: ''},
    zhihu: {link: ''},
  },
})

// 监听props变化，同步更新表单数据
watch(() => props.userInfo, (newUserInfo) => {
  if (newUserInfo) {
    profileForm.value = {...newUserInfo};
  }
}, {immediate: true, deep: true})

onMounted(() => {
  // 初始化时如果有用户信息就同步到表单
  if (props.userInfo) {
    profileForm.value = {...props.userInfo};
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
</script>

<template>

  <!-- Tab 导航 -->
  <div class="w-full">
    <div class="tabs tabs-lift">
      <input type="radio" name="account_tabs" class="tab" aria-label="基本信息" checked/>
      <div class="tab-content bg-base-100 border-base-300 p-6">
        <form @submit.prevent="updateProfile" class="grid grid-cols-1 gap-6">
          <div class="form-control">
            <label class="label">
              <span class="label-text font-normal">头像设置</span>
            </label>
            <AvatarUpload
                :current-avatar="profileForm.avatarUrl"
                @avatar-updated="handleAvatarUpdated"
            />
          </div>
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div class="form-control ">
              <label class="floating-label join w-full">
                <span>用户名</span>
                <input v-model="profileForm.username" type="text" class="input input-bordered w-full"
                       placeholder="请输入用户名" :disabled="!usernameEditing"/>
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
                <input v-model="profileForm.email" type="email" class="input input-bordered w-full"
                       placeholder="请输入邮箱" :disabled="!emailEditing"/>
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
                       placeholder="请输入昵称"/>
              </label>
            </div>
            <div class="form-control">
              <label class="floating-label">
                <span>个性签名</span>
                <input v-model="profileForm.signature" type="text" class="input input-bordered w-full"
                       placeholder="请输入个性签名"/>
              </label>
            </div>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div class="form-control">
              <label class="floating-label">
                <span>网站名</span>
                <input v-model="profileForm.websiteName" type="text" class="input input-bordered w-full"
                       placeholder="请输入网站名"/>
              </label>
            </div>
            <div class="form-control">
              <label class="floating-label">
                <span>网站地址</span>
                <input v-model="profileForm.website" type="text" class="input input-bordered w-full"
                       placeholder="请输入网站地址"/>
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
                           class="input input-bordered w-full"
                           placeholder="请输入Github地址"/>
                  </label>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>BiliBili</span>
                    <input v-model="profileForm.externalInformation.bilibili.link" type="text"
                           class="input input-bordered w-full"
                           placeholder="请输入BiliBili地址"/>
                  </label>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>Weibo</span>
                    <input v-model="profileForm.externalInformation.weibo.link" type="text"
                           class="input input-bordered w-full"
                           placeholder="请输入Weibo地址"/>
                  </label>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>Twitter</span>
                    <input v-model="profileForm.externalInformation.twitter.link" type="text"
                           class="input input-bordered w-full"
                           placeholder="请输入Twitter地址"/>
                  </label>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>zhihu</span>
                    <input v-model="profileForm.externalInformation.zhihu.link" type="text"
                           class="input input-bordered w-full"
                           placeholder="请输入Zhihu地址"/>
                  </label>
                </div>
                <div class="form-control">
                  <label class="floating-label">
                    <span>linkedIn</span>
                    <input v-model="profileForm.externalInformation.linkedIn.link" type="text"
                           class="input input-bordered w-full"
                           placeholder="请输入LinkedIn地址"/>
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

      <input type="radio" name="account_tabs" class="tab" aria-label="修改密码"/>
      <div class="tab-content bg-base-100 border-base-300 p-6">
        <h3 class="card-title text-lg mb-6 border-b border-base-300 pb-3">修改密码</h3>
        <form @submit.prevent="updatePassword" class="grid grid-cols-1 gap-6">
          <div class="form-control">
            <label class="label">
              <span class="label-text font-normal">当前密码</span>
            </label>
            <input v-model="passwordForm.currentPassword" type="password" class="input input-bordered w-full"
                   placeholder="请输入当前密码" required/>
          </div>
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div class="form-control">
              <label class="label">
                <span class="label-text font-normal">新密码</span>
              </label>
              <input v-model="passwordForm.newPassword" type="password" class="input input-bordered w-full"
                     placeholder="请输入新密码" required/>
              <label class="label">
                <span class="label-text-alt">密码长度至少8位，包含字母和数字</span>
              </label>
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text font-normal">确认新密码</span>
              </label>
              <input v-model="passwordForm.confirmPassword" type="password" class="input input-bordered w-full"
                     placeholder="请再次输入新密码" required/>
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

      <input type="radio" name="account_tabs" class="tab" aria-label="隐私设置"/>
      <div class="tab-content bg-base-100 border-base-300 p-6">
        <h3 class="card-title text-lg mb-6 border-b border-base-300 pb-3">隐私设置</h3>
        <div class="grid grid-cols-1 gap-6">
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input v-model="privacySettings.showArticles" type="checkbox" class="toggle toggle-primary"/>
              <div>
                <span class="label-text font-normal">公开文章列表</span>
                <div class="text-sm text-base-content/60">允许其他用户查看我发布的文章</div>
              </div>
            </label>
          </div>

          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input v-model="privacySettings.showFollowing" type="checkbox" class="toggle toggle-primary"/>
              <div>
                <span class="label-text font-normal">公开关注列表</span>
                <div class="text-sm text-base-content/60">允许其他用户查看我的关注和粉丝</div>
              </div>
            </label>
          </div>

          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input v-model="privacySettings.emailNotifications" type="checkbox"
                     class="toggle toggle-primary"/>
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


<style scoped>
</style>
