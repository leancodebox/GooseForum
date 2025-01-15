<script setup>
import {VueCropper} from 'vue-cropper'
import 'vue-cropper/dist/index.css'
import {NFormItem,NButton, NCard, NFlex, NImage, NInput, NList, NListItem, NModal, NSpace, NText, useMessage} from "naive-ui"
import {onMounted, onUnmounted, ref} from "vue";
import {getUserProfile, updateUserProfile, uploadAvatar, changePassword} from "@/service/request";

let isSmallScreen = ref(false)

function checkScreenSize() {
  isSmallScreen.value = window.innerWidth < 800;
}

onMounted(() => {
  checkScreenSize();
  window.addEventListener('resize', checkScreenSize);
})
onUnmounted(() => {
  window.removeEventListener('resize', checkScreenSize);
})

const userInfo = ref({
  email: '',
  nickname: '',
  bio: '',
  signature: '',
  website: '',
});

const editing = ref({
  nickname: false,
  email: false,
  bio: false,
  signature: false,
  website: false,
});

const editValues = ref({
  nickname: '',
  email: '',
  bio: '',
  signature: '',
  website: '',
});

const message = useMessage()
const avatarUrl = ref('')
const uploading = ref(false)
const showCropModal = ref(false)
const cropperRef = ref(null)
const previewUrl = ref('')
const cropImg = ref('')
const fileInputRef = ref(null)
const activeTab = ref('profile') // 'profile' 或 'password'
const passwordForm = ref({
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
})

async function fetchUserInfo() {
  try {
    const response = await getUserProfile();
    if (response.code === 0) {
      userInfo.value = response.result;
      avatarUrl.value = response.result.avatarUrl || '';
    }
  } catch (error) {
    console.error('获取用户信息失败:', error);
    message.error('获取用户信息失败');
  }
}

async function saveEdit(field) {
  try {
    const updates = {
      [field]: editValues.value[field]
    };

    const response = await updateUserProfile(updates);
    if (response.code === 0) {
      userInfo.value[field] = editValues.value[field];
      editing.value[field] = false;
    }
  } catch (error) {
    console.error('保存失败:', error);
  }
}

function startEdit(field) {
  editValues.value[field] = userInfo.value[field];
  editing.value[field] = true;
}

// 处理文件选择
function handleFileSelect(event) {
  const file = event.target.files[0]
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
function compressImage(base64Data, maxWidth = 200) {
  return new Promise((resolve) => {
    const img = new Image()
    img.src = base64Data
    img.onload = () => {
      const canvas = document.createElement('canvas')
      const ctx = canvas.getContext('2d')

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
          (blob) => resolve(blob),
          'image/jpeg',
          0.8  // 压缩质量，0.8通常是质量和大小的好平衡点
      )
    }
  })
}

// 修改裁切完成函数
async function handleCropFinish() {
  if (!cropperRef.value) return

  try {
    cropperRef.value.getCropData(async (base64Data) => {
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
function realTimePreview() {
  if (!cropperRef.value) return
  // 使用 getCropData 获取裁切后的图片数据
  cropperRef.value.getCropData((data) => {
    cropImg.value = data  // 使用单独的变量存储裁切后的预览图
  })
}

// 在组件卸载时清理预览URL
onUnmounted(() => {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
  }
})

onMounted(() => {
  fetchUserInfo();
  checkScreenSize();
  window.addEventListener('resize', checkScreenSize);
})

async function handleChangePassword() {
    if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
        message.error('两次输入的新密码不一致')
        return
    }

    try {
        const response = await changePassword(
            passwordForm.value.oldPassword,
            passwordForm.value.newPassword
        )
        if (response.code === 0) {
            message.success('密码修改成功')
            // 清空表单
            passwordForm.value = {
                oldPassword: '',
                newPassword: '',
                confirmPassword: ''
            }
        }
    } catch (error) {
        console.error('修改密码失败:', error)
    }
}

</script>
<template>
  <n-card :bordered="false">
    <n-flex :justify="isSmallScreen ? 'start' : 'center'" :align-mid="true" :vertical="isSmallScreen">
      <n-flex vertical class="nav-buttons">
        <n-button
          class="nav-button"
          :type="activeTab === 'profile' ? 'primary' : 'default'"
          @click="activeTab = 'profile'"
        >
          个人资料
        </n-button>
        <n-button
          class="nav-button"
          :type="activeTab === 'password' ? 'primary' : 'default'"
          @click="activeTab = 'password'"
        >
          修改密码
        </n-button>
      </n-flex>

      <n-flex vertical class="list-component">
        <template v-if="activeTab === 'profile'">
          <n-card title="个人资料" :bordered="false">
            <n-list>
              <n-list-item>
                邮箱:
                <template v-if="!editing.email">
                  {{ userInfo.email }}
                  <n-button text type="primary" @click="startEdit('email')">编辑</n-button>
                </template>
                <template v-else>
                  <n-input v-model:value="editValues.email"/>
                  <n-button type="primary" @click="saveEdit('email')">保存</n-button>
                  <n-button @click="editing.email = false">取消</n-button>
                </template>
              </n-list-item>
              <n-list-item>
                昵称:
                <template v-if="!editing.nickname">
                  {{ userInfo.nickname }}
                  <n-button text type="primary" @click="startEdit('nickname')">编辑</n-button>
                </template>
                <template v-else>
                  <n-input v-model:value="editValues.nickname"/>
                  <n-button type="primary" @click="saveEdit('nickname')">保存</n-button>
                  <n-button @click="editing.nickname = false">取消</n-button>
                </template>
              </n-list-item>
              <n-list-item>
                个人简介:
                <template v-if="!editing.bio">
                  {{ userInfo.bio }}
                  <n-button text type="primary" @click="startEdit('bio')">编辑</n-button>
                </template>
                <template v-else>
                  <n-input v-model:value="editValues.bio" type="textarea" :autosize="{ minRows: 3, maxRows: 5 }"/>
                  <n-button type="primary" @click="saveEdit('bio')">保存</n-button>
                  <n-button @click="editing.bio = false">取消</n-button>
                </template>
              </n-list-item>

              <n-list-item>
                署名:
                <template v-if="!editing.signature">
                  {{ userInfo.signature }}
                  <n-button text type="primary" @click="startEdit('signature')">编辑</n-button>
                </template>
                <template v-else>
                  <n-input v-model:value="editValues.signature"/>
                  <n-button type="primary" @click="saveEdit('signature')">保存</n-button>
                  <n-button @click="editing.signature = false">取消</n-button>
                </template>
              </n-list-item>

              <n-list-item>
                个人网站:
                <template v-if="!editing.website">
                  {{ userInfo.website }}
                  <n-button text type="primary" @click="startEdit('website')">编辑</n-button>
                </template>
                <template v-else>
                  <n-input v-model:value="editValues.website"/>
                  <n-button type="primary" @click="saveEdit('website')">保存</n-button>
                  <n-button @click="editing.website = false">取消</n-button>
                </template>
              </n-list-item>
            </n-list>
          </n-card>
          <n-card title="头像设置" :bordered="false">
            <n-list>
              <n-space vertical>
                <n-image
                    width="100"
                    :src="avatarUrl || '/api/assets/default-avatar.png'"
                    :preview-disabled="!avatarUrl"
                    object-fit="cover"
                    :round="true"
                />

                <input
                    type="file"
                    accept="image/gif,image/jpeg,image/jpg,image/png"
                    style="display: none"
                    ref="fileInputRef"
                    @change="handleFileSelect"
                />

                <n-button :loading="uploading" @click="fileInputRef?.click()">
                  {{ uploading ? '上传中...' : '选择图片' }}
                </n-button>

                <span class="upload-tip">支持 jpg、png、gif 格式，建议选择小于 2MB 的图片，最终头像大小不超过 500KB</span>
              </n-space>
            </n-list>
          </n-card>
        </template>

        <template v-if="activeTab === 'password'">
          <n-card title="修改密码" :bordered="false">
            <n-list>
              <n-list-item>
                <n-form-item label="原始密码">
                  <n-input
                    type="password"
                    v-model:value="passwordForm.oldPassword"
                    placeholder="请输入原始密码"
                  />
                </n-form-item>
              </n-list-item>
              <n-list-item>
                <n-form-item label="新密码">
                  <n-input
                    type="password"
                    v-model:value="passwordForm.newPassword"
                    placeholder="请输入新密码"
                  />
                </n-form-item>
              </n-list-item>
              <n-list-item>
                <n-form-item label="确认新密码">
                  <n-input
                    type="password"
                    v-model:value="passwordForm.confirmPassword"
                    placeholder="请再次输入新密码"
                  />
                </n-form-item>
              </n-list-item>
              <n-list-item>
                <n-button type="primary" @click="handleChangePassword">
                  确认修改
                </n-button>
              </n-list-item>
            </n-list>
          </n-card>
        </template>
      </n-flex>
    </n-flex>
  </n-card>

  <!-- 裁切弹窗 -->
  <n-modal
      v-model:show="showCropModal"
      preset="card"
      style="width: 600px"
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

      <n-flex justify="space-between" align="center">
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

        <n-space>
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
.nav-buttons {
  min-width: 240px;
  gap: 8px;
}

.nav-button {
  width: 100%;
  margin-bottom: 8px;
}

.nav-button:last-child {
  margin-bottom: 0;
}

.menu-component {
  min-width: 180px;
  max-width: 240px;
  flex: 1; /* 让菜单在垂直布局时占据可用空间 */
}

.list-component {
  min-width: 460px;
  max-width: 900px;
  flex: 2; /* 让列表在垂直布局时占据更多空间 */
}

@media (max-width: 800px) {
  .menu-component, .list-component {
    min-width: 100%; /* 在小屏幕上，让菜单和列表都占据全部宽度 */
    max-width: none;
  }

  .n-flex.vertical {
    flex-direction: column; /* 确保在垂直模式下，元素是垂直排列的 */
  }
}

.upload-tip {
  color: #999;
  font-size: 12px;
}

.n-image {
  border-radius: 50%;
  overflow: hidden;
}

.cropper-container {
  width: 100%;
  background-color: #f5f5f5;
  border-radius: 4px;
}

.preview-box {
  padding: 16px;
  border-radius: 8px;
  background-color: #f9f9f9;
}

.preview-container {
  margin-top: 8px;
}

.preview-circle {
  width: 120px;
  height: 120px;
  overflow: hidden;
  border-radius: 50%;
  border: 2px solid #fff;
  box-shadow: 0 0 0 1px #ddd;
}

/* 调整裁切区域的样式 */
:deep(.cropper-view-box) {
  border-radius: 50%;
  outline: 2px solid #fff;
  outline-offset: -1px;
}

:deep(.cropper-face) {
  background-color: inherit !important;
}

:deep(.cropper-dashed) {
  border-color: #666;
}

:deep(.cropper-center) {
  opacity: 0.4;
}

:deep(.cropper-line) {
  background-color: #39f;
}

:deep(.cropper-point) {
  background-color: #39f;
  width: 8px;
  height: 8px;
  opacity: 0.8;
}

:deep(.cropper-point.point-se) {
  width: 8px;
  height: 8px;
}
</style>

