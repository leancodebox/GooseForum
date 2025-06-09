<script setup lang="ts">
import {onMounted, reactive, ref} from 'vue'
import AccountSettings from "./components/AccountSettings.vue";
import {getUserInfo} from "@/utils/articleService.ts";
import type {UserInfo} from "@/utils/articleInterfaces.ts";

// 加载状态
const isLoading = ref(true)

// 用户信息 - 根据UserInfo接口定义
const userInfo = reactive<UserInfo>({
  avatarUrl: '/static/pic/default-avatar.png',
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

// 额外的统计信息（不在UserInfo接口中）
const userStats = reactive({
  articleCount: 0,
  followingCount: 0,
  followersCount: 0,
  joinDate: ''
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
  }
}

// 处理账户设置更新事件
const handleUserInfoUpdated = () => {
  console.log('收到用户信息更新通知，重新加载用户信息');
  loadUserInfo();
}

onMounted(() => {
  loadUserInfo();
})


// 我的文章
const myArticles = ref([
  {
    id: 1,
    title: 'Vue 3 组合式 API 深度解析',
    summary: '详细介绍 Vue 3 组合式 API 的使用方法和最佳实践，包括 setup 函数、响应式系统等核心概念。',
    publishTime: '2024-01-15',
    viewCount: 1234,
    likeCount: 89,
    commentCount: 23,
    status: '已发布'
  },
  {
    id: 2,
    title: 'Nuxt.js 性能优化实战指南',
    summary: '从多个维度分析 Nuxt.js 应用的性能优化策略，包括代码分割、懒加载、缓存策略等。',
    publishTime: '2024-01-10',
    viewCount: 856,
    likeCount: 67,
    commentCount: 15,
    status: '已发布'
  },
  {
    id: 3,
    title: 'TypeScript 进阶技巧分享',
    summary: '分享一些 TypeScript 的高级用法和技巧，帮助开发者更好地使用类型系统。',
    publishTime: '2024-01-05',
    viewCount: 432,
    likeCount: 34,
    commentCount: 8,
    status: '草稿'
  }
])




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
const editProfile = () => {
  // 切换到账户设置tab
  const settingsTab = document.querySelector('input[aria-label="账户设置"]')
  if (settingsTab) {
    settingsTab.checked = true
  }
}

// 编辑文章
const editArticle = (id) => {
  console.log('编辑文章:', id)
  // 跳转到编辑页面
}

// 查看文章数据
const viewStats = (id) => {
  console.log('查看文章数据:', id)
  // 显示文章统计数据
}

// 删除文章
const deleteArticle = (id) => {
  console.log('删除文章:', id)
  // 确认删除文章
}


</script>
<template>
  <div class="container mx-auto px-4 py-8">
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
                  <div class="mask mask-squircle w-24 h-24" >
                    <img :src="userInfo.avatarUrl || '/static/pic/default-avatar.png'"
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
                    <div class="text-lg font-bold text-base-content">{{ userStats.articleCount }}</div>
                    <div class="text-xs text-base-content/60">文章数</div>
                  </div>
                  <div class="text-center">
                    <div class="text-lg font-bold text-base-content">{{ userStats.followingCount }}</div>
                    <div class="text-xs text-base-content/60">获赞数</div>
                  </div>
                  <div class="text-center">
                    <div class="text-lg font-bold text-base-content">{{ userStats.followersCount }}</div>
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
              <input type="radio" name="my_tabs_3" class="tab" aria-label="我的文章" checked="checked"/>
              <div class="tab-content space-y-4 mt-3">
                <div class="flex justify-between items-center">
                  <h2 class="text-2xl font-bold">我的文章</h2>
                  <a href="/publish" class="btn btn-primary btn-sm">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24"
                         stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                    </svg>
                    写文章
                  </a>
                </div>
                <div class="space-y-3">
                  <div v-for="article in myArticles" :key="article.id"
                       class="card bg-base-100 shadow-sm hover:shadow-md transition-shadow">
                    <div class="card-body p-4">
                      <div class="flex justify-between items-start">
                        <div class="flex-1">
                          <h3 class="card-title text-lg hover:text-primary cursor-pointer">{{ article.title }}</h3>
                          <p class="text-base-content/70 text-sm mt-2 line-clamp-2">{{ article.summary }}</p>
                          <div class="flex items-center gap-4 mt-3 text-sm text-base-content/60">
                            <span>{{ article.publishTime }}</span>
                            <span>{{ article.viewCount }} 阅读</span>
                            <span>{{ article.likeCount }} 点赞</span>
                            <span>{{ article.commentCount }} 评论</span>
                            <div class="badge badge-outline badge-sm">{{ article.status }}</div>
                          </div>
                        </div>
                        <div class="dropdown dropdown-end">
                          <div tabindex="0" role="button" class="btn btn-ghost btn-sm btn-circle">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24"
                                 stroke="currentColor">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                    d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z"/>
                            </svg>
                          </div>
                          <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-32">
                            <li><a @click="editArticle(article.id)">编辑</a></li>
                            <li><a @click="viewStats(article.id)">数据</a></li>
                            <li><a @click="deleteArticle(article.id)" class="text-error">删除</a></li>
                          </ul>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                <!-- 分页 -->
                <div class="flex justify-center mt-8">
                  <div class="join bg-base-100 rounded-lg shadow-sm">
                    <button class="join-item btn btn-sm bg-base-100 border-base-300">«</button>
                    <button class="join-item btn btn-sm bg-primary text-primary-content border-primary">1</button>
                    <button class="join-item btn btn-sm bg-base-100 border-base-300">2</button>
                    <button class="join-item btn btn-sm bg-base-100 border-base-300">3</button>
                    <button class="join-item btn btn-sm bg-base-100 border-base-300">»</button>
                  </div>
                </div>
              </div>

              <input type="radio" name="my_tabs_3" class="tab" aria-label="我的收藏"/>
              <div class="tab-content space-y-4 mt-3">
                <div class="space-y-3">
                  <div class="card bg-base-100 shadow-sm w-full">
                    <div class="card-body">
                      <p>开发中开发中开发中ing</p>
                    </div>
                  </div>
                </div>
              </div>

              <input type="radio" name="my_tabs_3" class="tab" aria-label="我的评论"/>
              <div class="tab-content space-y-4 mt-3">
                <div class="space-y-3">
                  <div class="card bg-base-100 shadow-sm w-full">
                    <div class="card-body">
                      <p>寻找中寻找中寻找中ing</p>
                    </div>
                  </div>
                </div>
              </div>

              <input type="radio" name="my_tabs_3" class="tab" aria-label="账户设置"/>
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
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>