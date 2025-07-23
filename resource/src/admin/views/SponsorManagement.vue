<template>
  <div class="">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-normal text-gray-800">赞助商管理</h1>
      <div class="flex gap-3">
        <button
          @click="addUserSponsor"
          class="btn btn-secondary btn-sm"
        >
          <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
          </svg>
          添加用户赞助
        </button>
        <button
          @click="saveSponsorsData"
          class="btn btn-success btn-sm"
          :class="{ 'loading': saving }"
          :disabled="saving"
        >
          <svg v-if="!saving" class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
          </svg>
          {{ saving ? '保存中...' : '保存' }}
        </button>
      </div>
    </div>

    <!-- 赞助商级别管理 -->
    <div class="card bg-base-100 shadow-xl mb-6">
      <div class="card-body">
        <h2 class="card-title text-lg mb-4  font-normal">赞助商级别管理</h2>

        <div class="space-y-4">
          <div v-for="(level, levelKey) in sponsorsData.sponsors" :key="levelKey" class="border border-gray-200 rounded-lg p-4">
            <div class="flex justify-between items-center mb-3">
              <h3 class="text-md font-semibold text-gray-700">{{ getLevelName(levelKey) }}</h3>
              <div class="flex gap-2">
                <button
                  @click="addSponsorToLevel(levelKey)"
                  class="btn btn-outline btn-xs"
                >
                  添加赞助商
                </button>
              </div>
            </div>

            <draggable
              v-model="sponsorsData.sponsors[levelKey]"
              group="sponsors"
              item-key="name"
              class="space-y-2"
            >
              <template #item="{ element: sponsor, index }">
                <div class="bg-gray-50 p-3 rounded border flex items-center justify-between hover:bg-gray-100 transition-colors">
                  <div class="flex items-center space-x-3">
                    <div class="cursor-move text-gray-400">
                      <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z"></path>
                      </svg>
                    </div>
                    <img v-if="sponsor.logo" :src="sponsor.logo" :alt="sponsor.name" class="w-8 h-8 rounded object-cover">
                    <div class="w-8 h-8 bg-gray-300 rounded flex items-center justify-center" v-else>
                      <svg class="w-4 h-4 text-gray-500" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clip-rule="evenodd"></path>
                      </svg>
                    </div>
                    <div>
                      <div class="font-medium text-gray-900">{{ sponsor.name || '未命名' }}</div>
                      <div class="text-sm text-gray-500">{{ sponsor.info || '暂无描述' }}</div>
                      <div class="flex flex-wrap gap-1 mt-1">
                        <span v-for="tag in sponsor.tag" :key="tag" class="badge badge-outline badge-xs">{{ tag }}</span>
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center space-x-2">
                    <button
                      @click="editSponsor(levelKey, index)"
                      class="btn btn-ghost btn-xs"
                    >
                      编辑
                    </button>
                    <button
                      @click="deleteSponsor(levelKey, index)"
                      class="btn btn-ghost btn-error btn-xs"
                    >
                      删除
                    </button>
                  </div>
                </div>
              </template>
            </draggable>

            <div v-if="level.length === 0" class="text-center py-3 text-gray-500">
              暂无赞助商，点击上方按钮添加
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 用户赞助管理 -->
    <div class="card bg-base-100 shadow-xl">
      <div class="card-body">
        <h2 class="card-title text-lg mb-4">用户赞助管理</h2>

        <draggable
          v-model="sponsorsData.users"
          group="users"
          item-key="name"
          class="space-y-2"
        >
          <template #item="{ element: user, index }">
            <div class="bg-gray-50 p-3 rounded border flex items-center justify-between hover:bg-gray-100 transition-colors">
              <div class="flex items-center space-x-3">
                <div class="cursor-move text-gray-400">
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z"></path>
                  </svg>
                </div>
                <img v-if="user.icon" :src="user.icon" :alt="user.name" class="w-8 h-8 rounded-full object-cover">
                <div class="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center" v-else>
                  <svg class="w-4 h-4 text-gray-500" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd"></path>
                  </svg>
                </div>
                <div>
                  <div class="font-medium text-gray-900">{{ user.name || '匿名用户' }}</div>
                  <div class="text-sm text-gray-500">赞助金额: {{ user.amount || '0' }} | 时间: {{ user.time || '未知' }}</div>
                </div>
              </div>
              <div class="flex items-center space-x-2">
                <button
                  @click="editUserSponsor(index)"
                  class="btn btn-ghost btn-xs"
                >
                  编辑
                </button>
                <button
                  @click="deleteUserSponsor(index)"
                  class="btn btn-ghost btn-error btn-xs"
                >
                  删除
                </button>
              </div>
            </div>
          </template>
        </draggable>

        <div v-if="sponsorsData.users.length === 0" class="text-center py-8 text-gray-500">
          暂无用户赞助记录，点击上方按钮添加
        </div>
      </div>
    </div>

    <!-- 编辑赞助商模态框 -->
    <div class="modal" :class="{ 'modal-open': showSponsorModal }">
      <div class="modal-box w-11/12 max-w-2xl">
        <h3 class="font-bold text-lg mb-4">{{ editingSponsorIndex === -1 ? '添加' : '编辑' }}赞助商</h3>

        <div class="space-y-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">赞助商名称 *</span>
            </label>
            <input
              v-model="currentSponsor.name"
              type="text"
              placeholder="请输入赞助商名称"
              class="input input-bordered w-full"
            >
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">Logo</span>
            </label>
            <div class="flex gap-2">
              <input
                v-model="currentSponsor.logo"
                type="text"
                placeholder="请输入Logo URL或上传图片"
                class="input input-bordered flex-1"
              >
              <input
                ref="logoFileInput"
                type="file"
                accept="image/*"
                class="hidden"
                @change="handleLogoUpload"
              >
              <button
                type="button"
                class="btn btn-outline btn-sm"
                @click="logoFileInput?.click()"
                :disabled="uploading"
              >
                {{ uploading ? '上传中...' : '上传' }}
              </button>
            </div>
            <div v-if="currentSponsor.logo" class="mt-2">
              <img :src="currentSponsor.logo" alt="Logo预览" class="w-16 h-16 object-cover rounded" />
            </div>
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">描述信息</span>
            </label>
            <textarea
              v-model="currentSponsor.info"
              placeholder="请输入赞助商描述"
              class="textarea textarea-bordered w-full"
              rows="3"
            ></textarea>
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">官网链接</span>
            </label>
            <input
              v-model="currentSponsor.url"
              type="text"
              placeholder="请输入官网链接"
              class="input input-bordered w-full"
            >
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">标签 (用逗号分隔)</span>
            </label>
            <input
              v-model="tagInput"
              type="text"
              placeholder="例如: 技术,开源,云服务"
              class="input input-bordered w-full"
            >
          </div>
        </div>

        <div class="modal-action">
          <button @click="closeSponsorModal" class="btn btn-ghost">取消</button>
          <button @click="saveSponsor" class="btn btn-primary">保存</button>
        </div>
      </div>
    </div>

    <!-- 编辑用户赞助模态框 -->
    <div class="modal" :class="{ 'modal-open': showUserModal }">
      <div class="modal-box w-11/12 max-w-2xl">
        <h3 class="font-bold text-lg mb-4">{{ editingUserIndex === -1 ? '添加' : '编辑' }}用户赞助</h3>

        <div class="space-y-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">用户名称 *</span>
            </label>
            <input
              v-model="currentUser.name"
              type="text"
              placeholder="请输入用户名称"
              class="input input-bordered w-full"
            >
          </div>

          <div class="form-control">
              <label class="label">
                <span class="label-text">Logo</span>
              </label>
              <div class="flex gap-2">
                <input
                  v-model="currentUser.logo"
                  type="text"
                  placeholder="请输入Logo URL或上传图片"
                  class="input input-bordered flex-1"
                >
                <input
                  ref="userLogoFileInput"
                  type="file"
                  accept="image/*"
                  class="hidden"
                  @change="handleUserLogoUpload"
                >
                <button
                  type="button"
                  class="btn btn-outline btn-sm"
                  @click="userLogoFileInput?.click()"
                  :disabled="uploading"
                >
                  {{ uploading ? '上传中...' : '上传' }}
                </button>
              </div>
              <div v-if="currentUser.logo" class="mt-2">
                 <img :src="currentUser.logo" alt="Logo预览" class="w-16 h-16 object-cover rounded" />
               </div>
            </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">赞助金额</span>
            </label>
            <input
              v-model="currentUser.amount"
              type="text"
              placeholder="请输入赞助金额"
              class="input input-bordered w-full"
            >
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">赞助时间</span>
            </label>
            <input
              v-model="currentUser.time"
              type="date"
              class="input input-bordered w-full"
            >
          </div>
        </div>

        <div class="modal-action">
          <button @click="closeUserModal" class="btn btn-ghost">取消</button>
          <button @click="saveUserSponsor" class="btn btn-primary">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import draggable from 'vuedraggable'
import { getSponsors, saveSponsors } from '../utils/adminService'
import type { SponsorsConfig, SponsorItem, UserSponsor } from '../utils/adminInterfaces'

// 响应式数据
const sponsorsData = reactive<SponsorsConfig>({
  sponsors: {
    level0: [],
    level1: [],
    level2: [],
    level3: []
  },
  users: []
})

const saving = ref(false)
const showSponsorModal = ref(false)
const showUserModal = ref(false)
const editingSponsorLevel = ref('')
const editingSponsorIndex = ref(-1)
const editingUserIndex = ref(-1)

const currentSponsor = reactive<SponsorItem>({
  name: '',
  logo: '',
  info: '',
  url: '',
  tag: []
})

const currentUser = reactive<UserSponsor>({
  name: '',
  logo: '',
  amount: '',
  time: ''
})

const tagInput = ref('')
const uploading = ref(false)
const logoFileInput = ref<HTMLInputElement>()
const userLogoFileInput = ref<HTMLInputElement>()

// 方法
const loadSponsors = async () => {
  try {
    const response = await getSponsors()
    if (response.code === 0) {
      Object.assign(sponsorsData, response.result)
    }
  } catch (error) {
    console.error('加载赞助商数据失败:', error)
  }
}

const saveSponsorsData = async () => {
  saving.value = true
  try {
    const response = await saveSponsors(sponsorsData)
    if (response.code === 0) {
      // 可以添加成功提示
      console.log('保存成功')
    }
  } catch (error) {
    console.error('保存失败:', error)
  } finally {
    saving.value = false
  }
}

const getLevelName = (levelKey: string) => {
  const levelNames: Record<string, string> = {
    level0: '特别赞助商 (Level 0)',
    level1: '金牌赞助商 (Level 1)',
    level2: '银牌赞助商 (Level 2)',
    level3: '铜牌赞助商 (Level 3)'
  }
  return levelNames[levelKey] || levelKey
}



const addSponsorToLevel = (levelKey: string) => {
  editingSponsorLevel.value = levelKey
  editingSponsorIndex.value = -1
  resetCurrentSponsor()
  showSponsorModal.value = true
}

const editSponsor = (levelKey: string, index: number) => {
  editingSponsorLevel.value = levelKey
  editingSponsorIndex.value = index
  const sponsor = sponsorsData.sponsors[levelKey as keyof typeof sponsorsData.sponsors][index]
  Object.assign(currentSponsor, sponsor)
  tagInput.value = sponsor.tag.join(', ')
  showSponsorModal.value = true
}

const deleteSponsor = (levelKey: string, index: number) => {
  if (confirm('确定要删除这个赞助商吗？')) {
    sponsorsData.sponsors[levelKey as keyof typeof sponsorsData.sponsors].splice(index, 1)
  }
}

const resetCurrentSponsor = () => {
  currentSponsor.name = ''
  currentSponsor.logo = ''
  currentSponsor.info = ''
  currentSponsor.url = ''
  currentSponsor.tag = []
  tagInput.value = ''
}

const closeSponsorModal = () => {
  showSponsorModal.value = false
  resetCurrentSponsor()
}

const saveSponsor = () => {
  if (!currentSponsor.name.trim()) {
    alert('请输入赞助商名称')
    return
  }

  // 处理标签
  currentSponsor.tag = tagInput.value.split(',').map(tag => tag.trim()).filter(tag => tag)

  if (editingSponsorIndex.value === -1) {
    // 添加新赞助商
    sponsorsData.sponsors[editingSponsorLevel.value as keyof typeof sponsorsData.sponsors].push({ ...currentSponsor })
  } else {
    // 编辑现有赞助商
    Object.assign(
      sponsorsData.sponsors[editingSponsorLevel.value as keyof typeof sponsorsData.sponsors][editingSponsorIndex.value],
      currentSponsor
    )
  }

  closeSponsorModal()
}

const addUserSponsor = () => {
  editingUserIndex.value = -1
  resetCurrentUser()
  showUserModal.value = true
}

const editUserSponsor = (index: number) => {
  editingUserIndex.value = index
  Object.assign(currentUser, sponsorsData.users[index])
  showUserModal.value = true
}

const deleteUserSponsor = (index: number) => {
  if (confirm('确定要删除这个用户赞助记录吗？')) {
    sponsorsData.users.splice(index, 1)
  }
}

const resetCurrentUser = () => {
  currentUser.name = ''
  currentUser.logo = ''
  currentUser.amount = ''
  currentUser.time = ''
}

const closeUserModal = () => {
  showUserModal.value = false
  resetCurrentUser()
}

const saveUserSponsor = () => {
  if (!currentUser.name.trim()) {
    alert('请输入用户名称')
    return
  }

  if (editingUserIndex.value === -1) {
    // 添加新用户赞助
    sponsorsData.users.push({ ...currentUser })
  } else {
    // 编辑现有用户赞助
    Object.assign(sponsorsData.users[editingUserIndex.value], currentUser)
  }

  closeUserModal()
}

// 文件上传处理
const handleLogoUpload = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)

    const response = await fetch('/file/img-upload', {
       method: 'POST',
       body: formData,
       credentials: 'include'
     })

    const result = await response.json()
     if (result.code === 0) {
       currentSponsor.logo = result.result.url
     } else {
       alert('上传失败: ' + (result.msg || result.message))
     }
  } catch (error) {
    console.error('上传失败:', error)
    alert('上传失败，请重试')
  } finally {
    uploading.value = false
  }
}

const handleUserLogoUpload = async (event: Event) => {
   const file = (event.target as HTMLInputElement).files?.[0]
   if (!file) return

   uploading.value = true
   try {
     const formData = new FormData()
     formData.append('file', file)

     const response = await fetch('/file/img-upload', {
       method: 'POST',
       body: formData,
       credentials: 'include'
     })

     const result = await response.json()
     if (result.code === 0) {
       currentUser.logo = result.result.url
     } else {
       alert('上传失败: ' + (result.msg || result.message))
     }
   } catch (error) {
     console.error('上传失败:', error)
     alert('上传失败，请重试')
   } finally {
     uploading.value = false
   }
 }

// 生命周期
onMounted(() => {
  loadSponsors()
})
</script>
