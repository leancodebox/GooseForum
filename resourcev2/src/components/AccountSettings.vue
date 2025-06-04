<script setup lang="ts">
import {onMounted, reactive, ref, watch} from 'vue'
import AvatarUpload from './AvatarUpload.vue'
import {changePassword, saveUserInfo} from '@/utils/articleService.ts'
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
    profileForm.value = { ...newUserInfo };
    console.log('AccountSettings: 接收到用户信息更新', newUserInfo);
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

// Tab 控制
const activeTab = ref('profile')

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
  <div class="w-full">
    <!-- Tab 导航 -->
    <div role="tablist" class="tabs tabs-lift ml-3">
      <a 
        role="tab" 
        class="tab tab-lg"
        :class="{ 'tab-active': activeTab === 'profile' }"
        @click="activeTab = 'profile'"
      >
        基本信息
      </a>
      <a 
        role="tab" 
        class="tab tab-lg"
        :class="{ 'tab-active': activeTab === 'password' }"
        @click="activeTab = 'password'"
      >
        修改密码
      </a>
      <a 
        role="tab" 
        class="tab tab-lg"
        :class="{ 'tab-active': activeTab === 'privacy' }"
        @click="activeTab = 'privacy'"
      >
        隐私设置
      </a>
    </div>

    <!-- Tab 内容 -->
    <div class="card bg-base-100 shadow-xl">
      <div class="card-body">
        <!-- 基本信息 Tab -->
        <div v-if="activeTab === 'profile'">
          <form @submit.prevent="updateProfile" class="grid grid-cols-1 gap-6">
            <div class="form-control">
              <label class="label">
                <span class="label-text font-medium">头像设置</span>
              </label>
              <AvatarUpload
                  :current-avatar="profileForm.avatarUrl"
                  @avatar-updated="handleAvatarUpdated"
              />
            </div>
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-medium">用户名</span>
                </label>
                <input v-model="profileForm.username" type="text" class="input input-bordered w-full"
                       placeholder="请输入用户名" disabled/>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-medium">邮箱</span>
                </label>
                <input v-model="profileForm.email" type="email" class="input input-bordered w-full"
                       placeholder="请输入邮箱" disabled/>
              </div>
            </div>

            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-medium">昵称</span>
                </label>
                <input v-model="profileForm.nickname" type="text" class="input input-bordered w-full"
                       placeholder="请输入昵称"/>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-medium">个性签名</span>
                </label>
                <input v-model="profileForm.signature" type="text" class="input input-bordered w-full"
                       placeholder="请输入个性签名"/>
              </div>
            </div>

            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-medium">网站名</span>
                </label>
                <input v-model="profileForm.websiteName" type="text" class="input input-bordered w-full"
                       placeholder="请输入网站名"/>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-medium">网站地址</span>
                </label>
                <input v-model="profileForm.website" type="text" class="input input-bordered w-full"
                       placeholder="请输入网站地址"/>
              </div>
            </div>
            <details tabindex="0" class="collapse collapse-arrow bg-base-100 border-base-300 border">
              <summary class="collapse-title font-semibold">外站配置</summary>
              <div class="collapse-content text-sm">

                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <div class="form-control">
                    <label class="label">
                      <span class="label-text font-medium">Github</span>
                    </label>
                    <input v-model="profileForm.externalInformation.github.link" type="text"
                           class="input input-bordered w-full"
                           placeholder="请输入Github地址"/>
                  </div>
                  <div class="form-control">
                    <label class="label">
                      <span class="label-text font-medium">BiliBili</span>
                    </label>
                    <input v-model="profileForm.externalInformation.bilibili.link" type="text"
                           class="input input-bordered w-full"
                           placeholder="请输入BiliBili地址"/>
                  </div>
                </div>


                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <div class="form-control">
                    <label class="label">
                      <span class="label-text font-medium">Weibo</span>
                    </label>
                    <input v-model="profileForm.externalInformation.weibo.link" type="text"
                           class="input input-bordered w-full"
                           placeholder="请输入Weibo地址"/>
                  </div>
                  <div class="form-control">
                    <label class="label">
                      <span class="label-text font-medium">Twitter</span>
                    </label>
                    <input v-model="profileForm.externalInformation.twitter.link" type="text"
                           class="input input-bordered w-full"
                           placeholder="请输入Twitter地址"/>
                  </div>
                </div>


                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <div class="form-control">
                    <label class="label">
                      <span class="label-text font-medium">zhihu</span>
                    </label>
                    <input v-model="profileForm.externalInformation.zhihu.link" type="text"
                           class="input input-bordered w-full"
                           placeholder="请输入Zhihu地址"/>
                  </div>
                  <div class="form-control">
                    <label class="label">
                      <span class="label-text font-medium">linkedIn</span>
                    </label>
                    <input v-model="profileForm.externalInformation.linkedIn.link" type="text"
                           class="input input-bordered w-full"
                           placeholder="请输入LinkedIn地址"/>
                  </div>
                </div>
              </div>
            </details>

            <div class="form-control">
              <label class="label">
                <span class="label-text font-medium">个人简介</span>
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

        <!-- 密码修改 Tab -->
        <div v-if="activeTab === 'password'">
          <h3 class="card-title text-lg mb-6 border-b border-base-300 pb-3">修改密码</h3>
          <form @submit.prevent="updatePassword" class="grid grid-cols-1 gap-6">
            <div class="form-control">
              <label class="label">
                <span class="label-text font-medium">当前密码</span>
              </label>
              <input v-model="passwordForm.currentPassword" type="password" class="input input-bordered w-full"
                     placeholder="请输入当前密码" required/>
            </div>
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-medium">新密码</span>
                </label>
                <input v-model="passwordForm.newPassword" type="password" class="input input-bordered w-full"
                       placeholder="请输入新密码" required/>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-medium">确认新密码</span>
                </label>
                <input v-model="passwordForm.confirmPassword" type="password" class="input input-bordered w-full"
                       placeholder="请再次输入新密码" required/>
              </div>
            </div>
            <div class="flex justify-end">
              <button type="submit" class="btn btn-primary" :disabled="changingPassword">
                <span v-if="changingPassword" class="loading loading-spinner loading-sm"></span>
                {{ changingPassword ? '修改中...' : '修改密码' }}
              </button>
            </div>
          </form>
        </div>

        <!-- 隐私设置 Tab -->
        <div v-if="activeTab === 'privacy'">
          <h3 class="card-title text-lg mb-6 border-b border-base-300 pb-3">隐私设置</h3>
          <div class="grid grid-cols-1 gap-6">
            <div class="form-control">
              <label class="label cursor-pointer justify-start gap-4">
                <input v-model="privacySettings.showArticles" type="checkbox" class="toggle toggle-primary"/>
                <div>
                  <span class="label-text font-medium">公开文章列表</span>
                  <div class="text-sm text-base-content/60">允许其他用户查看我发布的文章</div>
                </div>
              </label>
            </div>

            <div class="form-control">
              <label class="label cursor-pointer justify-start gap-4">
                <input v-model="privacySettings.showFollowing" type="checkbox" class="toggle toggle-primary"/>
                <div>
                  <span class="label-text font-medium">公开关注列表</span>
                  <div class="text-sm text-base-content/60">允许其他用户查看我的关注和粉丝</div>
                </div>
              </label>
            </div>

            <div class="form-control">
              <label class="label cursor-pointer justify-start gap-4">
                <input v-model="privacySettings.emailNotifications" type="checkbox"
                       class="toggle toggle-primary"/>
                <div>
                  <span class="label-text font-medium">邮件通知</span>
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
  </div>
</template>