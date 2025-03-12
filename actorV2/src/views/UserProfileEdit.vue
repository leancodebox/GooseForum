<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/userStore'
import type { UserInfo } from '@/stores/userStore'
import axiosInstance from '@/utils/axiosInstance'
import { enqueueMessage } from '@/utils/messageManager'

const userStore = useUserStore()
const activeTab = ref('profile') // 'profile' 或 'articles'
const articles = ref([])
const isUploading = ref(false)

// 用户信息表单
const userForm = ref({
  nickname: '',
  email: '',
  bio: '',
  website: '',
  signature: '',
})

// 初始化用户数据
onMounted(async () => {
  if (userStore.userInfo) {
    userForm.value = {
      nickname: userStore.userInfo.nickname,
      email: userStore.userInfo.email,
      bio: userStore.userInfo.bio,
      website: userStore.userInfo.website,
      signature: userStore.userInfo.signature,
    }
    await fetchUserArticles()
  }
})

// 获取用户文章列表
const fetchUserArticles = async () => {
  try {
    const response = await axiosInstance.get('/bbs/get-user-articles')
    articles.value = response.data.result
  } catch (error) {
    enqueueMessage('获取文章列表失败')
  }
}

// 更新用户信息
const updateProfile = async () => {
  try {
    await axiosInstance.post('/user/update-profile', userForm.value)
    await userStore.fetchUserInfo()
    enqueueMessage('个人信息更新成功')
  } catch (error) {
    enqueueMessage('更新失败，请重试')
  }
}

// 处理头像上传
const handleAvatarUpload = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  const formData = new FormData()
  formData.append('avatar', file)
  isUploading.value = true

  try {
    await axiosInstance.post('/user/update-avatar', formData)
    await userStore.fetchUserInfo()
    enqueueMessage('头像更新成功')
  } catch (error) {
    enqueueMessage('头像上传失败')
  } finally {
    isUploading.value = false
  }
}
</script>

<template>
  <div class="profile-container">
    <!-- 个人信息头部 -->
    <div class="profile-header">
      <div class="user-basic-info">
        <div class="avatar-section">
          <div class="avatar-wrapper">
            <img :src="userStore.userInfo?.avatarUrl" alt="用户头像" class="current-avatar">
            <div class="avatar-upload">
              <input type="file" accept="image/*" @change="handleAvatarUpload" id="avatar-input" :disabled="isUploading">
              <label for="avatar-input" :class="{ 'uploading': isUploading }">
                <i class="fas fa-camera"></i>
              </label>
            </div>
          </div>
        </div>
        <div class="user-info">
          <h2>{{ userStore.userInfo?.nickname || userStore.userInfo?.username }}</h2>
          <p class="user-bio">{{ userForm.bio || '这个人很懒，还没有写简介' }}</p>
        </div>
      </div>
      
      <div class="tab-buttons">
        <button :class="['tab-btn', { active: activeTab === 'profile' }]" @click="activeTab = 'profile'">
          个人资料
        </button>
        <button :class="['tab-btn', { active: activeTab === 'articles' }]" @click="activeTab = 'articles'">
          我的文章
        </button>
      </div>
    </div>

    <!-- 内容区域 -->
    <div class="profile-content">
      <!-- 个人资料编辑 -->
      <div v-if="activeTab === 'profile'" class="profile-edit">
        <form @submit.prevent="updateProfile" class="profile-form">
          <div class="form-row">
            <div class="form-group">
              <label for="nickname">昵称</label>
              <input type="text" id="nickname" v-model="userForm.nickname">
            </div>
            <div class="form-group">
              <label for="email">邮箱</label>
              <input type="email" id="email" v-model="userForm.email">
            </div>
          </div>

          <div class="form-group">
            <label for="bio">个人简介</label>
            <textarea id="bio" v-model="userForm.bio" rows="3"></textarea>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="website">个人网站</label>
              <input type="url" id="website" v-model="userForm.website">
            </div>
            <div class="form-group">
              <label for="signature">个性签名</label>
              <input type="text" id="signature" v-model="userForm.signature">
            </div>
          </div>

          <button type="submit" class="submit-btn">保存修改</button>
        </form>
      </div>

      <!-- 文章列表 -->
      <div v-else class="articles-list">
        <div v-if="articles.length === 0" class="empty-state">
          <i class="fas fa-file-alt"></i>
          <p>还没有发布过文章</p>
          <router-link to="/post-edit" class="create-post-btn">写第一篇文章</router-link>
        </div>
        <div v-else class="article-grid">
          <div v-for="article in articles" :key="article.id" class="article-card">
            <h3 class="article-title">
              <a :href="`/post/${article.id}`">{{ article.title }}</a>
            </h3>
            <p class="article-excerpt">{{ article.content.substring(0, 100) }}...</p>
            <div class="article-meta">
              <span class="publish-time">{{ article.createTime }}</span>
              <div class="stats">
                <span><i class="fas fa-eye"></i> {{ article.viewCount }}</span>
                <span><i class="fas fa-comment"></i> {{ article.commentCount }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.profile-header {
  background: var(--card-bg);
  border-radius: 12px;
  padding: 2rem;
  margin-bottom: 2rem;
  box-shadow: 0 2px 8px var(--shadow-color);
}

.user-basic-info {
  display: flex;
  align-items: center;
  gap: 2rem;
  margin-bottom: 2rem;
}

.avatar-wrapper {
  position: relative;
  width: 120px;
  height: 120px;
}

.current-avatar {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}

.avatar-upload {
  position: absolute;
  bottom: 0;
  right: 0;
}

.avatar-upload input[type="file"] {
  display: none;
}

.avatar-upload label {
  width: 32px;
  height: 32px;
  background: var(--primary-color);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.user-info h2 {
  margin: 0 0 0.5rem 0;
  color: var(--text-color);
}

.user-bio {
  color: var(--text-color-light);
  margin: 0;
}

.tab-buttons {
  display: flex;
  gap: 1rem;
  border-top: 1px solid var(--border-color);
  padding-top: 1.5rem;
}

.tab-btn {
  padding: 0.5rem 2rem;
  border: none;
  background: none;
  color: var(--text-color);
  font-weight: 500;
  cursor: pointer;
  position: relative;
}

.tab-btn.active {
  color: var(--primary-color);
}

.tab-btn.active::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  width: 100%;
  height: 2px;
  background: var(--primary-color);
}

.profile-content {
  background: var(--card-bg);
  border-radius: 12px;
  padding: 2rem;
  box-shadow: 0 2px 8px var(--shadow-color);
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 1.5rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: var(--text-color);
  font-weight: 500;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--input-bg-color);
  color: var(--text-color);
  transition: all 0.2s;
}

.form-group input:focus,
.form-group textarea:focus {
  border-color: var(--primary-color);
  outline: none;
}

.submit-btn {
  background: var(--primary-color);
  color: white;
  border: none;
  padding: 0.75rem 2rem;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
}

.article-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
}

.article-card {
  background: var(--bg-color);
  border-radius: 8px;
  padding: 1.5rem;
  transition: all 0.2s;
}

.article-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px var(--shadow-color);
}

.article-title {
  margin: 0 0 1rem 0;
}

.article-excerpt {
  color: var(--text-color-light);
  margin: 0 0 1rem 0;
  line-height: 1.5;
}

.article-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--text-color-light);
  font-size: 0.9rem;
}

.stats {
  display: flex;
  gap: 1rem;
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  color: var(--text-color-light);
}

.empty-state i {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.create-post-btn {
  display: inline-block;
  margin-top: 1rem;
  padding: 0.75rem 2rem;
  background: var(--primary-color);
  color: white;
  border-radius: 8px;
  text-decoration: none;
  transition: opacity 0.2s;
}

@media (max-width: 768px) {
  .profile-container {
    padding: 1rem;
  }

  .user-basic-info {
    flex-direction: column;
    text-align: center;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .article-grid {
    grid-template-columns: 1fr;
  }
}
</style>