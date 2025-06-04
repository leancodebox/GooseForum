<script setup>
import {reactive, ref} from 'vue'
import AccountSettings from "./components/AccountSettings.vue";
// 用户信息
const userInfo = reactive({
  id: 1,
  username: 'GooseUser',
  email: 'user@example.com',
  avatar: '/static/pic/default-avatar.png',
  bio: '热爱技术的开发者，专注于前端和全栈开发',
  articleCount: 15,
  followingCount: 42,
  followersCount: 128,
  joinDate: '2023-01-15'
})

// 标签页
const activeTab = ref('articles')
const tabs = [
  {key: 'articles', label: '我的文章'},
  {key: 'favorites', label: '我的收藏'},
  {key: 'comments', label: '我的评论'},
  {key: 'settings', label: '账户设置'}
]

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

// 我的收藏
const myFavorites = ref([
  {
    id: 1,
    title: 'React 18 新特性详解',
    summary: 'React 18 带来了许多新特性，包括并发渲染、自动批处理等...',
    author: {
      username: 'ReactDev',
      avatar: 'https://img.daisyui.com/images/profile/demo/3@94.webp'
    },
    publishTime: '2024-01-12'
  },
  {
    id: 2,
    title: 'Node.js 微服务架构实践',
    summary: '基于 Node.js 构建微服务架构的实践经验分享...',
    author: {
      username: 'NodeMaster',
      avatar: 'https://img.daisyui.com/images/profile/demo/4@94.webp'
    },
    publishTime: '2024-01-08'
  }
])

// 我的评论
const myComments = ref([
  {
    id: 1,
    articleTitle: 'JavaScript 异步编程最佳实践',
    content: '这篇文章写得很好，特别是关于 Promise 和 async/await 的部分，对我很有帮助！',
    createTime: '2024-01-14 10:30',
    likeCount: 5
  },
  {
    id: 2,
    articleTitle: 'CSS Grid 布局完全指南',
    content: '感谢分享，Grid 布局确实比 Flexbox 在某些场景下更适用。',
    createTime: '2024-01-13 15:20',
    likeCount: 3
  }
])

// 个人资料表单
const profileForm = reactive({
  username: userInfo.username,
  email: userInfo.email,
  bio: userInfo.bio,
  avatar: userInfo.avatar
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

// 编辑资料
const editProfile = () => {
  activeTab.value = 'settings'
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

// 移除收藏
const removeFavorite = (id) => {
  console.log('移除收藏:', id)
  myFavorites.value = myFavorites.value.filter(item => item.id !== id)
}

// 删除评论
const deleteComment = (id) => {
  console.log('删除评论:', id)
  myComments.value = myComments.value.filter(item => item.id !== id)
}

// 更新个人资料
const updateProfile = async () => {
  try {
    console.log('更新个人资料:', profileForm)
    // 这里应该调用实际的API
    Object.assign(userInfo, profileForm)
  } catch (error) {
    console.error('更新失败:', error)
  }
}

// 修改密码
const changePassword = async () => {
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    alert('新密码和确认密码不一致')
    return
  }

  try {
    console.log('修改密码')
    // 这里应该调用实际的API
    // 重置表单
    Object.assign(passwordForm, {
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    })
  } catch (error) {
    console.error('修改密码失败:', error)
  }
}

// 处理头像上传
const handleAvatarUpload = (event) => {
  const file = event.target.files[0]
  if (file) {
    // 这里应该上传到服务器，目前使用本地预览
    const reader = new FileReader()
    reader.onload = (e) => {
      profileForm.avatar = e.target.result
    }
    reader.readAsDataURL(file)
  }
}

// 保存隐私设置
const savePrivacySettings = async () => {
  try {
    console.log('保存隐私设置:', privacySettings)
    // 这里应该调用实际的API
  } catch (error) {
    console.error('保存失败:', error)
  }
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
              <div class="avatar mb-4 mx-auto">
                <div class="mask mask-squircle w-24 h-24">
                  <img :src="userInfo.avatar" :alt="userInfo.username"/>
                </div>
              </div>
              <h2 class="card-title justify-center text-xl">{{ userInfo.username }}</h2>
              <p class="text-base-content/70 text-sm mb-4">{{ userInfo.bio || '这个人很懒，什么都没留下' }}</p>

              <div class="grid grid-cols-3 gap-4 mb-4">
                <div class="text-center">
                  <div class="text-lg font-bold text-base-content">{{ userInfo.articleCount }}</div>
                  <div class="text-xs text-base-content/60">文章数</div>
                </div>
                <div class="text-center">
                  <div class="text-lg font-bold text-base-content">{{ userInfo.followingCount }}</div>
                  <div class="text-xs text-base-content/60">获赞数</div>
                </div>
                <div class="text-center">
                  <div class="text-lg font-bold text-base-content">{{ userInfo.followersCount }}</div>
                  <div class="text-xs text-base-content/60">粉丝数</div>
                </div>
              </div>
              <div class="mt-4">
                <button class="btn btn-primary btn-sm btn-block" @click="editProfile">编辑资料</button>
              </div>
            </div>
          </div>
        </div>

        <!-- 右侧主要内容区域 -->
        <div class="lg:col-span-3">
          <!-- 标签页导航 -->
          <div class="tabs tabs-boxed mb-6">
            <a v-for="tab in tabs" :key="tab.key" class="tab" :class="{ 'tab-active': activeTab === tab.key }"
               @click="activeTab = tab.key">
              {{ tab.label }}
            </a>
          </div>

          <!-- 我的文章 -->
          <div v-if="activeTab === 'articles'" class="space-y-4">
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

          <!-- 我的收藏 -->
          <div v-if="activeTab === 'favorites'" class="space-y-4">
            <h2 class="text-2xl font-bold">我的收藏</h2>
            <div class="space-y-3">
              <div v-for="favorite in myFavorites" :key="favorite.id"
                   class="card bg-base-100 shadow-sm hover:shadow-md transition-shadow">
                <div class="card-body p-4">
                  <div class="flex items-start gap-3">
                    <div class="avatar">
                      <div class="mask mask-squircle w-10 h-10">
                        <img :src="favorite.author.avatar" :alt="favorite.author.username"/>
                      </div>
                    </div>
                    <div class="flex-1">
                      <h3 class="card-title text-lg hover:text-primary cursor-pointer">{{ favorite.title }}</h3>
                      <p class="text-sm text-base-content/70 mt-1">by {{ favorite.author.username }} · {{
                          favorite.publishTime
                        }}</p>
                      <p class="text-base-content/70 text-sm mt-2 line-clamp-2">{{ favorite.summary }}</p>
                    </div>
                    <button class="btn btn-ghost btn-sm" @click="removeFavorite(favorite.id)">
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24"
                           stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                              d="M6 18L18 6M6 6l12 12"/>
                      </svg>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 我的评论 -->
          <div v-if="activeTab === 'comments'" class="space-y-4">
            <h2 class="text-2xl font-bold">我的评论</h2>
            <div class="space-y-3">
              <div v-for="comment in myComments" :key="comment.id" class="card bg-base-100 shadow-sm">
                <div class="card-body p-4">
                  <div class="text-sm text-base-content/70 mb-2">
                    评论于文章：<span class="text-primary hover:underline cursor-pointer">{{
                      comment.articleTitle
                    }}</span>
                  </div>
                  <p class="text-base-content mb-2">{{ comment.content }}</p>
                  <div class="flex justify-between items-center text-sm text-base-content/60">
                    <span>{{ comment.createTime }}</span>
                    <div class="flex gap-2">
                      <span>{{ comment.likeCount }} 点赞</span>
                      <button class="text-error hover:underline" @click="deleteComment(comment.id)">删除</button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 账户设置 -->
          <div v-if="activeTab === 'settings'" class="space-y-6">
            <AccountSettings/>
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