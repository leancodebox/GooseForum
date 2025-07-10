<script setup lang="ts">
import {onMounted, reactive, ref, nextTick} from 'vue'
import AccountSettings from "./components/AccountSettings.vue";
import MyArticles from "./components/MyArticles.vue";
import {getUserInfo} from "@/utils/articleService.ts";
import type {UserInfo} from "@/utils/articleInterfaces.ts";

// 加载状态
const isLoading = ref(true)

// 标签页 refs
const articlesTabRef = ref<HTMLInputElement>()
const collectionsTabRef = ref<HTMLInputElement>()
const commentsTabRef = ref<HTMLInputElement>()
const settingsTabRef = ref<HTMLInputElement>()

// 用户信息 - 根据UserInfo接口定义
const userInfo = reactive<UserInfo>({
  avatarUrl: '/static/pic/default-avatar.webp',
  username: '',
  nickname: '',
  email: '',
  bio: '',
  website: '',
  websiteName: '',
  signature: '',
  externalInformation: {
    github: {link: ''},
    weibo: {link: ''},
    bilibili: {link: ''},
    twitter: {link: ''},
    linkedIn: {link: ''},
    zhihu: {link: ''}
  }
})


// 加载用户信息的函数
const loadUserInfo = async () => {
  try {
    const res = await getUserInfo();
    if (res.code === 0 && res.result) {
      // 更新用户信息
      Object.assign(userInfo, res.result);
      // 同步更新表单数据
      Object.assign(profileForm, {
        username: res.result.username,
        nickname: res.result.nickname,
        email: res.result.email,
        bio: res.result.bio,
        signature: res.result.signature,
        website: res.result.website,
        websiteName: res.result.websiteName,
        avatarUrl: res.result.avatarUrl
      });
      console.log('用户信息加载成功:', res.result);
    } else {
      console.error('获取用户信息失败:', res.message);
    }
  } catch (error) {
    console.error('获取用户信息出错:', error);
  } finally {
    isLoading.value = false;
    // 在加载完成后检查URL参数
    await nextTick();
    checkUrlParams();
  }
}

// 处理账户设置更新事件
const handleUserInfoUpdated = () => {
  console.log('收到用户信息更新通知，重新加载用户信息');
  loadUserInfo();
}

// 检查URL参数并切换到对应标签页
const checkUrlParams = () => {
  const urlParams = new URLSearchParams(window.location.search);
  const tab = urlParams.get('tab');
  
  if (tab === 'settings') {
    if (settingsTabRef.value) {
      settingsTabRef.value.checked = true;
    }
  } else {
    // 默认选中"我的文章"标签页
    if (articlesTabRef.value) {
      articlesTabRef.value.checked = true;
    }
  }
}

onMounted(() => {
  loadUserInfo();
})


// 个人资料表单
const profileForm = reactive({
  username: userInfo.username,
  nickname: userInfo.nickname,
  email: userInfo.email,
  bio: userInfo.bio,
  signature: userInfo.signature,
  website: userInfo.website,
  websiteName: userInfo.websiteName,
  avatarUrl: userInfo.avatarUrl
})


// 编辑资料
const editProfile = async () => {
  // 等待DOM渲染完成后切换到账户设置tab
  await nextTick();
  if (settingsTabRef.value) {
    settingsTabRef.value.checked = true;
  }
}

</script>
<template>
  <div class="container mx-auto px-4 py-4">
    <div class="max-w-6xl mx-auto">
      <div class="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <!-- 左侧用户信息卡片 -->
        <div class="lg:col-span-1">
          <div class="card bg-base-100 shadow-xl sticky top-24">
            <div class="card-body text-center">
              <!-- 加载状态 -->
              <div v-if="isLoading" class="flex justify-center items-center py-8">
                <span class="loading loading-spinner loading-lg"></span>
              </div>
              <!-- 用户信息 -->
              <div v-else>
                <div class="avatar mb-4 mx-auto">
                  <div class="mask mask-squircle w-24 h-24">
                    <img :src="userInfo.avatarUrl || '/static/pic/default-avatar.webp'"
                         :alt="userInfo.nickname || userInfo.username"/>
                  </div>
                </div>
                <h2 class="card-title justify-center text-xl">{{
                    userInfo.nickname || userInfo.username || '用户'
                  }}</h2>
                <p class="text-base-content/70 text-sm mb-4">{{
                    userInfo.bio || userInfo.signature || '这个人很懒，什么都没留下'
                  }}</p>
                <div class="grid grid-cols-3 gap-4 mb-4">
                  <div class="text-center">
                    <div class="text-lg font-normal text-base-content">
                      {{ userInfo?.authorInfoStatistics?.articleCount ?? 0 }}
                    </div>
                    <div class="text-xs text-base-content/60">文章数</div>
                  </div>
                  <div class="text-center">
                    <div class="text-lg font-normal text-base-content">
                      {{ userInfo?.authorInfoStatistics?.likeGivenCount ?? 0 }}
                    </div>
                    <div class="text-xs text-base-content/60">获赞数</div>
                  </div>
                  <div class="text-center">
                    <div class="text-lg font-normal text-base-content">
                      {{ userInfo?.authorInfoStatistics?.followerCount ?? 0 }}
                    </div>
                    <div class="text-xs text-base-content/60">粉丝数</div>
                  </div>
                </div>
                <div class="mt-4">
                  <button class="btn btn-primary btn-sm btn-block" @click="editProfile">编辑资料</button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 右侧主要内容区域 -->
        <div class="lg:col-span-3">
          <!-- 加载状态 -->
          <div v-if="isLoading" class="flex justify-center items-center py-16">
            <div class="text-center">
              <span class="loading loading-spinner loading-lg mb-4"></span>
              <p class="text-base-content/70">正在加载用户信息...</p>
            </div>
          </div>

          <!-- 主要内容 -->
          <div v-else>
            <div class="tabs">
              <input ref="articlesTabRef" type="radio" name="my_tabs_3" class="tab" aria-label="我的文章"/>
              <div class="tab-content space-y-4 mt-3">
                <MyArticles :user-info="userInfo"></MyArticles>
              </div>

              <input ref="collectionsTabRef" type="radio" name="my_tabs_3" class="tab" aria-label="我的收藏"/>
              <div class="tab-content space-y-4 mt-3">
                <div class="space-y-3">
                  <div class="card bg-base-100 shadow-sm w-full">
                    <div class="card-body">
                      <p>开发中开发中开发中ing</p>
                    </div>
                  </div>
                </div>
              </div>

              <input ref="commentsTabRef" type="radio" name="my_tabs_3" class="tab" aria-label="我的评论"/>
              <div class="tab-content space-y-4 mt-3">
                <div class="space-y-3">
                  <div class="card bg-base-100 shadow-sm w-full">
                    <div class="card-body">
                      <p>寻找中寻找中寻找中ing</p>
                    </div>
                  </div>
                </div>
              </div>

              <input ref="settingsTabRef" type="radio" name="my_tabs_3" class="tab" aria-label="账户设置"/>
              <div class="tab-content space-y-4 mt-2">
                <AccountSettings
                    :user-info="userInfo"
                    @user-info-updated="handleUserInfoUpdated"
                />
              </div>
            </div>

          </div>
        </div>
      </div>
    </div>
  </div>
</template>


<style scoped>
</style>