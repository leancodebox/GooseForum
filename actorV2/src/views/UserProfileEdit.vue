<script setup lang="ts">
import {ref, onMounted, watch, onUnmounted} from 'vue'
import { useUserStore } from '@/stores/userStore'
import axiosInstance from '@/utils/axiosInstance'
import { enqueueMessage } from '@/utils/messageManager'
import {NFormItem,NButton, NCard, NFlex, NImage, NInput, NList, NListItem, NModal, NSpace, NText, useMessage, NGrid, NGridItem} from "naive-ui"
import {VueCropper} from 'vue-cropper'
import 'vue-cropper/dist/index.css'
import {uploadAvatar} from "@/utils/articleService.ts";

// 定义文章接口
interface Article {
  id: number
  title: string
  content: string
  createTime: string
  viewCount: number
  commentCount: number
}

// 定义用户表单接口
interface UserForm {
  nickname: string
  email: string
  bio: string
  website: string
  signature: string
}

const message = useMessage()
const avatarUrl = ref<string>('')
const uploading = ref<boolean>(false)
const showCropModal = ref<boolean>(false)
const cropperRef = ref<any>(null)
const previewUrl = ref<string>('')
const cropImg = ref<string>('')
const fileInputRef = ref<HTMLInputElement | null>(null)
const userStore = useUserStore()
const activeTab = ref<'profile' | 'articles'>('profile')
const articles = ref<Article[]>([])
const isUploading = ref<boolean>(false)
const isSmallScreen = ref<boolean>(false)

// 用户信息表单
const userForm = ref<UserForm>({
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
      nickname: userStore.userInfo.nickname || '',
      email: userStore.userInfo.email || '',
      bio: userStore.userInfo.bio || '',
      website: userStore.userInfo.website || '',
      signature: userStore.userInfo.signature || '',
    }
    avatarUrl.value = userStore.userInfo.avatarUrl || ''
  }
  checkScreenSize();
  window.addEventListener('resize', checkScreenSize);
})

watch(activeTab, async (newTab) => {
  if (newTab === 'articles' && articles.value.length === 0) {
    await fetchUserArticles()
  }
})

// 获取用户文章列表
const fetchUserArticles = async (): Promise<void> => {
  try {
    const response = await axiosInstance.get('/bbs/get-user-articles')
    articles.value = response.data.result
  } catch (error) {
    enqueueMessage('获取文章列表失败')
  }
}

// 更新用户信息
const updateProfile = async (): Promise<void> => {
  try {
    await axiosInstance.post('/user/update-profile', userForm.value)
    await userStore.fetchUserInfo()
    enqueueMessage('个人信息更新成功')
  } catch (error) {
    enqueueMessage('更新失败，请重试')
  }
}

// 处理文件选择
function handleFileSelect(event: Event): void {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  // 验证文件类型
  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif']
  if (!allowedTypes.includes(file.type)) {
    message.error('只支持 jpg、png、gif 格式的图片')
    return
  }

  // 验证文件大小（2MB）
  if (file.size > 2 * 1024 * 1024) {
    message.error('图片大小不能超过 2MB')
    return
  }

  // 创建预览URL
  previewUrl.value = URL.createObjectURL(file)
  showCropModal.value = true
}

// 添加图片压缩函数
function compressImage(base64Data: string, maxWidth = 200): Promise<Blob> {
  return new Promise((resolve) => {
    const img = new Image()
    img.src = base64Data
    img.onload = () => {
      const canvas = document.createElement('canvas')
      const ctx = canvas.getContext('2d')
      if (ctx == null) {
        resolve(new Blob())
        return
      }

      // 计算压缩后的尺寸，保持宽高比
      let width = img.width
      let height = img.height
      if (width > maxWidth) {
        height = (maxWidth * height) / width
        width = maxWidth
      }

      canvas.width = width
      canvas.height = height

      // 绘制图片
      ctx.drawImage(img, 0, 0, width, height)

      // 转换为 blob，使用较低的质量值来减小文件大小
      canvas.toBlob(
          (blob) => resolve(blob || new Blob()),
          'image/jpeg',
          0.8  // 压缩质量，0.8通常是质量和大小的好平衡点
      )
    }
  })
}

// 修改裁切完成函数
async function handleCropFinish(): Promise<void> {
  if (!cropperRef.value) return

  try {
    cropperRef.value.getCropData(async (base64Data: string) => {
      uploading.value = true
      try {
        // 压缩裁切后的图片
        const compressedBlob = await compressImage(base64Data)

        // 检查压缩后的大小
        if (compressedBlob.size > 500 * 1024) { // 500KB 限制
          message.error('图片太大，请选择更小的区域或更小的图片')
          uploading.value = false
          return
        }

        const response = await uploadAvatar(compressedBlob)
        if (response.code === 0) {
          avatarUrl.value = response.result.avatarUrl
          await userStore.fetchUserInfo()
          message.success('头像上传成功')
          showCropModal.value = false
          // 重置文件输入框
          if (fileInputRef.value) {
            fileInputRef.value.value = ''
          }
          // 清理预览URL
          if (previewUrl.value && previewUrl.value.startsWith('blob:')) {
            URL.revokeObjectURL(previewUrl.value)
          }
        }
      } catch (error) {
        message.error('头像上传失败')
        console.error('上传失败:', error)
      } finally {
        uploading.value = false
      }
    })
  } catch (err) {
    console.error('裁切失败:', err)
    message.error('裁切失败')
  }
}

// 修改实时预览函数
function realTimePreview(): void {
  if (!cropperRef.value) return
  // 使用 getCropData 获取裁切后的图片数据
  cropperRef.value.getCropData((data: string) => {
    cropImg.value = data  // 使用单独的变量存储裁切后的预览图
  })
}

// 在组件卸载时清理预览URL
onUnmounted(() => {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
  }
  window.removeEventListener('resize', checkScreenSize);
})

function checkScreenSize(): void {
  isSmallScreen.value = window.innerWidth < 800;
}
</script>

<template>
  <div class="profile-container">
    <!-- 个人信息头部 -->
    <div class="profile-header">
      <div class="user-basic-info">
        <div class="avatar-section">
          <div class="avatar-wrapper">
            <img :src="avatarUrl || '/static/pic/default-avatar.png'" alt="用户头像" class="current-avatar">
            <div class="avatar-upload">
              <input type="file" accept="image/*" @change="handleFileSelect" id="avatar-input" :disabled="uploading">
              <label for="avatar-input" :class="{ 'uploading': uploading }">
                <i class="fas fa-camera"></i>
                <span v-if="uploading" class="upload-spinner"></span>
              </label>
            </div>
          </div>
          <div class="avatar-tip">点击头像上传新图片</div>
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

  <!-- 裁切头像的模态框 -->
  <n-modal
      v-model:show="showCropModal"
      preset="card"
      :style="isSmallScreen ? 'width: 95%' : 'width: 600px'"
      title="裁切头像"
      :mask-closable="false"
  >
    <n-space vertical size="large">
      <div class="cropper-container">
        <vue-cropper
            ref="cropperRef"
            :img="previewUrl"
            :autoCrop="true"
            :fixedBox="true"
            :centerBox="true"
            :fixed="true"
            :fixedNumber="[1, 1]"
            :canScale="true"
            :high="true"
            :maxImgSize="2048"
            :autoCropWidth="300"
            :autoCropHeight="300"
            :outputSize="1"
            :infoTrue="true"
            style="height: 400px"
            @realTime="realTimePreview"
        />
      </div>

      <n-flex justify="space-between" align="center" :wrap="isSmallScreen">
        <div class="preview-box">
          <n-text depth="3">预览效果</n-text>
          <div class="preview-container">
            <div class="preview-circle">
              <img
                  :src="cropImg"
                  style="width: 100%; height: 100%; object-fit: cover;"
              />
            </div>
          </div>
        </div>

        <n-space :style="isSmallScreen ? 'margin-top: 20px; width: 100%; justify-content: flex-end;' : ''">
          <n-button @click="showCropModal = false">取消</n-button>
          <n-button
              type="primary"
              :loading="uploading"
              @click="handleCropFinish"
          >
            确定
          </n-button>
        </n-space>
      </n-flex>
    </n-space>
  </n-modal>
</template>

<style scoped>
.profile-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 1.5rem;
}

.profile-header {
  background: var(--card-bg);
  border-radius: 8px;  /* 减小圆角 */
  padding: 1.5rem;  /* 减小内边距 */
  margin-bottom: 1.5rem;  /* 减小底部间距 */
  box-shadow: 0 2px 6px var(--shadow-color);  /* 调整阴影 */
}

.user-basic-info {
  display: flex;
  align-items: center;
  gap: 1.5rem;  /* 减小间距 */
  margin-bottom: 1.5rem;  /* 减小底部间距 */
}

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.3rem;  /* 减小间距 */
}

.avatar-wrapper {
  position: relative;
  width: 100px;  /* 减小头像尺寸 */
  height: 100px;  /* 减小头像尺寸 */
  border-radius: 50%;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 3px 8px rgba(0, 0, 0, 0.1);  /* 调整阴影 */
  transition: transform 0.3s ease;
}

.avatar-wrapper:hover {
  transform: scale(1.05);
}

.avatar-wrapper:hover .avatar-upload label {
  opacity: 1;
}

.current-avatar {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: filter 0.3s ease;
}

.avatar-wrapper:hover .current-avatar {
  filter: brightness(0.7);
}

.avatar-upload {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-upload input[type="file"] {
  display: none;
}

.avatar-upload label {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  color: white;
  opacity: 0;
  transition: opacity 0.3s ease;
  cursor: pointer;
}

.avatar-upload label i {
  font-size: 1.5rem;
}

.avatar-tip {
  font-size: 0.75rem;  /* 减小字体大小 */
  color: var(--text-color-light);
  text-align: center;
}

.upload-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s linear infinite;
  position: absolute;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.user-info {
  flex: 1;
}

.user-info h2 {
  margin: 0 0 0.3rem 0;  /* 减小底部间距 */
  color: var(--text-color);
  font-size: 1.5rem;  /* 减小字体大小 */
}

.tab-buttons {
  display: flex;
  gap: 0.8rem;  /* 减小间距 */
  border-top: 1px solid var(--border-color);
  padding-top: 1rem;  /* 减小顶部内边距 */
}

.tab-btn {
  padding: 0.4rem 1.5rem;  /* 减小内边距 */
  border: none;
  background: none;
  color: var(--text-color);
  font-weight: 500;
  cursor: pointer;
  position: relative;
  transition: color 0.3s ease;
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
  border-radius: 8px;  /* 减小圆角 */
  padding: 1.5rem;  /* 减小内边距 */
  box-shadow: 0 2px 6px var(--shadow-color);  /* 调整阴影 */
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));  /* 减小最小宽度 */
  gap: 1.5rem;  /* 增加列间距，确保输入框之间有足够间隙 */
  margin-bottom: 1rem;  /* 统一设置为1rem */
}

.form-group {
  margin-bottom: 1rem;  /* 统一设置为1rem */
}

.form-group label {
  display: block;
  margin-bottom: 0.4rem;  /* 稍微增加标签与输入框的间距 */
  color: var(--text-color);
  font-weight: 500;
  font-size: 0.9rem;
}

.form-group input,
.form-group textarea {
  width: calc(100% - 20px);
  padding: 0.6rem 0.8rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--input-bg-color);
  color: var(--text-color);
  transition: all 0.2s;
  font-size: 0.9rem;
  line-height: 1.4;  /* 添加行高控制 */
}

.form-group textarea {
  min-height: 80px;  /* 为文本域设置最小高度 */
  resize: vertical;  /* 允许垂直调整大小 */
}

.submit-btn {
  background: var(--primary-color);
  color: white;
  border: none;
  padding: 0.6rem 1.5rem;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
  font-size: 0.9rem;
  margin-top: 0.5rem;  /* 添加顶部间距 */
}

.article-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));  /* 减小最小宽度 */
  gap: 1.2rem;  /* 减小间距 */
}

.article-card {
  background: var(--bg-color);
  border-radius: 6px;  /* 减小圆角 */
  padding: 1.2rem;  /* 减小内边距 */
  transition: all 0.3s ease;
  border: 1px solid var(--border-color);
}

.article-title {
  margin: 0 0 0.8rem 0;  /* 减小底部间距 */
  font-size: 1.1rem;  /* 减小字体大小 */
}

.article-excerpt {
  color: var(--text-color-light);
  margin: 0 0 0.8rem 0;  /* 减小底部间距 */
  line-height: 1.4;  /* 减小行高 */
  font-size: 0.9rem;  /* 减小字体大小 */
}

.empty-state {
  text-align: center;
  padding: 3rem 1.5rem;  /* 减小内边距 */
  color: var(--text-color-light);
}

.empty-state i {
  font-size: 2.5rem;  /* 减小字体大小 */
  margin-bottom: 0.8rem;  /* 减小底部间距 */
  opacity: 0.5;
}

.create-post-btn {
  display: inline-block;
  margin-top: 1.2rem;  /* 减小顶部间距 */
  padding: 0.6rem 1.5rem;  /* 减小内边距 */
  background: var(--primary-color);
  color: white;
  border-radius: 6px;  /* 减小圆角 */
  text-decoration: none;
  transition: all 0.3s ease;
  font-weight: 500;
  font-size: 0.9rem;  /* 减小字体大小 */
}

/* 裁切相关样式 */
.cropper-container {
  border: 1px solid var(--border-color);
  border-radius: 6px;  /* 减小圆角 */
  overflow: hidden;
}

.preview-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.4rem;  /* 减小间距 */
}

.preview-container {
  width: 80px;  /* 减小尺寸 */
  height: 80px;  /* 减小尺寸 */
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-circle {
  width: 70px;  /* 减小尺寸 */
  height: 70px;  /* 减小尺寸 */
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid var(--primary-color);
}

@media (max-width: 768px) {
  .profile-container {
    padding: 0.8rem;  /* 减小内边距 */
  }

  .profile-header, .profile-content {
    padding: 1.2rem;  /* 减小内边距 */
  }

  .user-basic-info {
    flex-direction: column;
    text-align: center;
    gap: 0.8rem;  /* 减小间距 */
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .article-grid {
    grid-template-columns: 1fr;
  }

  .tab-buttons {
    justify-content: center;
  }

  .tab-btn {
    padding: 0.4rem 0.8rem;  /* 减小内边距 */
    font-size: 0.9rem;  /* 减小字体大小 */
  }
}
</style>
