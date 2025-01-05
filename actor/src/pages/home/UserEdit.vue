<script setup>
import {NButton, NCard, NEllipsis, NFlex, NList, NListItem, NMenu, NInput, NSpace, NImage, NUpload, useMessage} from "naive-ui"
import {h, onMounted, onUnmounted, ref} from "vue";
import {getUserProfile, updateUserProfile, uploadAvatar} from "@/service/request";

let options = [
  {
    label: () =>
        h(NEllipsis, null, {default: () => '个人自料'}),
    key: '1'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '头像设置'}),
    key: '2'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '修改密码'}),
    key: '3'
  }
]
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
});

const editing = ref({
  nickname: false,
  email: false,
});

const editValues = ref({
  nickname: '',
  email: '',
});

const message = useMessage()
const avatarUrl = ref('')
const uploading = ref(false)

async function handleAvatarUpload({file}) {
  if (!file) return;
  
  // 验证文件类型
  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif'];
  if (!allowedTypes.includes(file.type)) {
    message.error('只支持 jpg、png、gif 格式的图片');
    return;
  }
  
  // 验证文件大小（2MB）
  if (file.size > 2 * 1024 * 1024) {
    message.error('图片大小不能超过 2MB');
    return;
  }
  
  uploading.value = true;
  try {
    const response = await uploadAvatar(file.file);
    if (response.code === 0) {
      avatarUrl.value = response.result.avatarUrl;
      message.success('头像上传成功');
    }
  } catch (error) {
    message.error('头像上传失败');
    console.error('上传失败:', error);
  } finally {
    uploading.value = false;
  }
}

async function fetchUserInfo() {
  try {
    const response = await getUserProfile();
    if (response.code === 0) {
      userInfo.value = response.result;
      avatarUrl.value = response.result.avatarUrl || '';
    }
  } catch (error) {
    console.error('获取用户信息失败:', error);
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

onMounted(() => {
  fetchUserInfo();
  checkScreenSize();
  window.addEventListener('resize', checkScreenSize);
})

</script>
<template>
  <n-card :bordered="false">
    <n-flex :justify="isSmallScreen ? 'start' : 'center'" :align-mid="true" :vertical="isSmallScreen">
      <n-menu :options="options" class="menu-component" default-value="1"/>
      <n-flex vertical class="list-component">
        <n-card title="个人资料" :bordered="false">
          <n-list>
            <n-list-item>
              邮箱:
              <template v-if="!editing.email">
                {{ userInfo.email }}
                <n-button text type="primary" @click="startEdit('email')">编辑</n-button>
              </template>
              <template v-else>
                <n-input v-model:value="editValues.email" />
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
                <n-input v-model:value="editValues.nickname" />
                <n-button type="primary" @click="saveEdit('nickname')">保存</n-button>
                <n-button @click="editing.nickname = false">取消</n-button>
              </template>
            </n-list-item>
          </n-list>
        </n-card>
        <n-card title="头像设置" :bordered="false">
          <n-list>
            <n-space vertical>
              <n-image
                  width="100"
                  :src="avatarUrl || 'https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg'"
                  :preview-disabled="!avatarUrl"
              />
              
              <n-upload
                  accept="image/gif,image/jpeg,image/jpg,image/png"
                  :max-size="2 * 1024 * 1024"
                  @change="handleAvatarUpload"
                  :show-file-list="false"
              >
                <n-button :loading="uploading">
                  {{ uploading ? '上传中...' : '上传图片' }}
                </n-button>
              </n-upload>
              
              <span class="upload-tip">支持 png、jpg、jpeg、gif 格式，文件大小不超过 2MB</span>
            </n-space>
          </n-list>
        </n-card>

        <n-card title="密码设置" :bordered="false">
          <n-list>
            <n-list-item>
              原始密码
              <n-input></n-input>
            </n-list-item>
            <n-list-item>
              新密码
            </n-list-item>
            <n-list-item>
              确认新密码
            </n-list-item>
          </n-list>
        </n-card>
      </n-flex>
    </n-flex>
  </n-card>

</template>

<style scoped>
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
</style>

