<script setup lang="ts">
import {ref, reactive, computed, onMounted} from 'vue'
import {getFriendLinks, saveFriendLinks} from "@/admin/utils/adminService.ts";
import type {FriendLinksGroup, LinkItem} from "@/admin/utils/adminInterfaces.ts";
import draggable from 'vuedraggable'

// 响应式数据
const friendLinksGroups = ref<FriendLinksGroup[]>([])
const loading = ref(false)
const saving = ref(false)
const showAddGroupModal = ref(false)
const showAddLinkModal = ref(false)
const editingGroupIndex = ref(-1)
const editingLinkIndex = ref(-1)
const currentGroupIndex = ref(-1)

// 新增分组表单
const newGroup = reactive({
  name: ''
})

// 新增/编辑链接表单
const linkForm = reactive({
  name: '',
  desc: '',
  url: '',
  logoUrl: '',
  status: 1
})

// 加载友情链接数据
const loadFriendLinks = async () => {
  loading.value = true
  try {
    const response = await getFriendLinks()
    if (response.code === 0) {
      friendLinksGroups.value = response.result || []
    }
  } catch (error) {
    console.error('加载友情链接失败:', error)
  } finally {
    loading.value = false
  }
}

// 保存友情链接
const saveFriendLinksData = async () => {
  saving.value = true
  try {
    const response = await saveFriendLinks(friendLinksGroups.value)
    if (response.code === 0) {
      // 显示成功提示
      alert('保存成功！')
    } else {
      alert('保存失败：' + response.message)
    }
  } catch (error) {
    console.error('保存友情链接失败:', error)
    alert('保存失败，请重试')
  } finally {
    saving.value = false
  }
}



// 删除分组
const deleteGroup = (index: number) => {
  if (confirm('确定要删除这个分组吗？')) {
    friendLinksGroups.value.splice(index, 1)
  }
}

// 打开添加分组模态框
const openAddGroupModal = () => {
  editingGroupIndex.value = -1
  newGroup.name = ''
  showAddGroupModal.value = true
}

// 打开编辑分组模态框
const openEditGroupModal = (index: number) => {
  editingGroupIndex.value = index
  newGroup.name = friendLinksGroups.value[index].name
  showAddGroupModal.value = true
}

// 保存分组（添加或编辑）
const saveGroup = () => {
  if (newGroup.name.trim()) {
    if (editingGroupIndex.value >= 0) {
      // 编辑现有分组
      friendLinksGroups.value[editingGroupIndex.value].name = newGroup.name.trim()
    } else {
      // 添加新分组
      friendLinksGroups.value.push({
        name: newGroup.name.trim(),
        links: []
      })
    }
    newGroup.name = ''
    showAddGroupModal.value = false
  }
}

// 打开添加链接模态框
const openAddLinkModal = (groupIndex: number) => {
  currentGroupIndex.value = groupIndex
  editingLinkIndex.value = -1
  resetLinkForm()
  showAddLinkModal.value = true
}

// 打开编辑链接模态框
const openEditLinkModal = (groupIndex: number, linkIndex: number) => {
  currentGroupIndex.value = groupIndex
  editingLinkIndex.value = linkIndex
  const link = friendLinksGroups.value[groupIndex].links[linkIndex]
  Object.assign(linkForm, link)
  showAddLinkModal.value = true
}

// 重置链接表单
const resetLinkForm = () => {
  Object.assign(linkForm, {
    name: '',
    desc: '',
    url: '',
    logoUrl: '',
    status: 1
  })
}

// 保存链接
const saveLink = () => {
  if (!linkForm.name.trim() || !linkForm.url.trim()) {
    alert('请填写链接名称和URL')
    return
  }

  const linkData = { ...linkForm }
  
  if (editingLinkIndex.value >= 0) {
    // 编辑现有链接
    friendLinksGroups.value[currentGroupIndex.value].links[editingLinkIndex.value] = linkData
  } else {
    // 添加新链接
    friendLinksGroups.value[currentGroupIndex.value].links.push(linkData)
  }
  
  showAddLinkModal.value = false
  resetLinkForm()
}

// 删除链接
const deleteLink = (groupIndex: number, linkIndex: number) => {
  if (confirm('确定要删除这个链接吗？')) {
    friendLinksGroups.value[groupIndex].links.splice(linkIndex, 1)
  }
}

// 切换链接状态
const toggleLinkStatus = (groupIndex: number, linkIndex: number) => {
  const link = friendLinksGroups.value[groupIndex].links[linkIndex]
  link.status = link.status === 1 ? 0 : 1
}

// 组件挂载时加载数据
onMounted(() => {
  loadFriendLinks()
})
</script>

<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-normal">友情链接管理</h1>
      <div class="flex gap-2">
        <button 
          class="btn btn-primary btn-sm"
          @click="openAddGroupModal"
          :disabled="loading"
        >
          添加分组
        </button>
        <button 
          class="btn btn-success btn-sm"
          @click="saveFriendLinksData"
          :disabled="saving || loading"
        >
          <span v-if="saving" class="loading loading-spinner loading-sm"></span>
          {{ saving ? '保存中...' : '保存' }}
        </button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="flex justify-center items-center py-12">
      <span class="loading loading-spinner loading-lg"></span>
    </div>

    <!-- 友情链接分组列表 -->
    <div v-else class="space-y-3">
      <draggable 
        v-model="friendLinksGroups" 
        item-key="name"
        class="space-y-3"
        :animation="200"
        ghost-class="opacity-50"
      >
        <template #item="{ element: group, index: groupIndex }">
          <div class="card bg-base-100 shadow-md hover:shadow-xl transition-shadow duration-300">
            <!-- 分组标题 -->
            <div class="bg-gradient-to-r from-primary/10 to-primary/20 p-3 rounded-t-lg ">
              <div class="flex justify-between items-center">
                <div class="flex items-center gap-2 flex-1 min-w-0">
                  <div class="cursor-move text-base-content/60">
                    <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M7 2a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM7 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM7 14a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 2a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 14a2 2 0 1 1-4 0 2 2 0 0 1 4 0z"></path>
                    </svg>
                  </div>
                  <h2 class="text-lg font-normal truncate">{{ group.name }}</h2>
                  <div class="flex flex-col sm:flex-row gap-1 sm:gap-2 items-start sm:items-center">
                    <span class="badge badge-neutral text-xs whitespace-nowrap">{{ group.links.length }}</span>
                    <span class="text-xs text-base-content/60 hidden sm:inline">个链接</span>
                  </div>
                </div>
                <div class="flex gap-2 flex-shrink-0">
                  <button 
                    class="btn btn-sm btn-ghost" 
                    @click="openEditGroupModal(groupIndex)"
                  >
                    编辑
                  </button>
                  <button 
                    class="btn btn-sm btn-primary" 
                    @click="openAddLinkModal(groupIndex)"
                  >
                    添加链接
                  </button>
                  <button 
                    class="btn btn-sm btn-error" 
                    @click="deleteGroup(groupIndex)"
                  >
                    删除分组
                  </button>
                </div>
              </div>
            </div>

            <!-- 链接列表 -->
            <div class="card-body p-4">
              <div v-if="group.links.length === 0" class="text-center py-8 text-base-content/60">
                暂无链接，点击上方"添加链接"按钮添加
              </div>
              <draggable 
                v-else
                v-model="group.links" 
                item-key="name"
               <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-3"
                :animation="200"
                ghost-class="opacity-50"
              >
                <template #item="{ element: link, index: linkIndex }">
                  <div class="group relative">
                    <!-- 悬停时显示的操作按钮 -->
                    <div class="absolute top-1 right-1 opacity-0 group-hover:opacity-100 transition-opacity z-10 flex gap-1">
                      <button 
                        class="btn btn-xs btn-ghost bg-base-100/80 backdrop-blur-sm" 
                        @click="openEditLinkModal(groupIndex, linkIndex)"
                        title="编辑"
                      >
                        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                        </svg>
                      </button>
                      <button 
                        class="btn btn-xs btn-error bg-base-100/80 backdrop-blur-sm" 
                        @click="deleteLink(groupIndex, linkIndex)"
                        title="删除"
                      >
                        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                        </svg>
                      </button>
                    </div>
                    
                    <!-- 拖拽手柄 -->
                    <div class="absolute top-1 left-1 cursor-move text-base-content/40 opacity-0 group-hover:opacity-100 transition-opacity">
                      <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M7 2a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM7 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM7 14a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 2a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 14a2 2 0 1 1-4 0 2 2 0 0 1 4 0z"></path>
                      </svg>
                    </div>
                    
                    
                    
                    <!-- 链接内容 -->
                    <div class="bg-base-100 hover:bg-base-200 transition-colors rounded-lg p-3 border border-base-300 hover:border-primary/30">
                      <div class="flex flex-col items-center text-center space-y-2">
                        <img 
                          v-if="link.logoUrl" 
                          :src="link.logoUrl" 
                          :alt="link.name"
                          class="w-10 h-10 rounded object-cover"
                          @error="($event.target as HTMLImageElement).style.display='none'"
                        >
                        <div v-else class="w-10 h-10 bg-base-300 rounded flex items-center justify-center">
                          <svg class="w-5 h-5 text-base-content/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"></path>
                          </svg>
                        </div>
                        
                        <div class="w-full">
                          <h3 class="text-xs font-medium truncate" :title="link.name">{{ link.name }}</h3>
                          <p v-if="link.desc" class="text-xs text-base-content/60 truncate mt-1" :title="link.desc">{{ link.desc }}</p>
                        </div>
                        
                        <!-- 状态切换 -->
                        <div class="form-control">
                          <label class="cursor-pointer">
                            <input 
                              type="checkbox" 
                              class="toggle toggle-xs toggle-success" 
                              :checked="link.status === 1"
                              @change="toggleLinkStatus(groupIndex, linkIndex)"
                            >
                          </label>
                        </div>
                      </div>
                    </div>
                  </div>
                </template>
              </draggable>
            </div>
          </div>
        </template>
      </draggable>
    </div>

    <!-- 添加/编辑分组模态框 -->
    <dialog class="modal" :class="{ 'modal-open': showAddGroupModal }">
      <div class="modal-box w-96 max-w-sm">
        <h3 class="text-lg font-normal text-center mb-4">
          {{ editingGroupIndex >= 0 ? '编辑分组' : '添加分组' }}
        </h3>
        <fieldset class="fieldset bg-base-200 border-base-300 rounded-box border p-6">
          <label class="label" for="group-name">分组名称</label>
          <input 
            id="group-name"
            v-model="newGroup.name" 
            type="text" 
            placeholder="请输入分组名称" 
            class="input input-bordered mb-6"
            @keyup.enter="saveGroup"
          >
          
          <div class="flex justify-end gap-2">
            <button class="btn" @click="showAddGroupModal = false">取消</button>
            <button class="btn btn-primary" @click="saveGroup" :disabled="!newGroup.name.trim()">确定</button>
          </div>
        </fieldset>
      </div>
      <div class="modal-backdrop" @click="showAddGroupModal = false"></div>
    </dialog>

    <!-- 添加/编辑链接模态框 -->
    <dialog class="modal" :class="{ 'modal-open': showAddLinkModal }">
      <div class="modal-box w-full max-w-lg">
        <h3 class="text-lg font-normal text-center mb-4">
          {{ editingLinkIndex >= 0 ? '编辑链接' : '添加链接' }}
        </h3>
        <fieldset class="fieldset bg-base-200 border-base-300 rounded-box border p-6">
        
        <!-- 基本信息 -->
        <div class="space-y-4">
          <div>
            <label class="label" for="link-name">链接名称 <span class="text-error">*</span></label>
            <input 
              id="link-name"
              v-model="linkForm.name" 
              type="text" 
              placeholder="请输入链接名称" 
              class="input input-bordered w-full"
            >
          </div>
          
          <div>
            <label class="label" for="link-desc">链接描述</label>
            <textarea 
              id="link-desc"
              v-model="linkForm.desc" 
              placeholder="请输入链接描述" 
              class="textarea textarea-bordered w-full"
              rows="3"
            ></textarea>
          </div>
          
          <div>
            <label class="label" for="link-url">链接地址 <span class="text-error">*</span></label>
            <input 
              id="link-url"
              v-model="linkForm.url" 
              type="url" 
              placeholder="https://example.com" 
              class="input input-bordered w-full"
            >
          </div>
          
          <div>
            <label class="label" for="link-logo">Logo URL</label>
            <input 
              id="link-logo"
              v-model="linkForm.logoUrl" 
              type="url" 
              placeholder="https://example.com/logo.png" 
              class="input input-bordered w-full"
            >
          </div>
          
          <div v-if="linkForm.logoUrl" class="flex items-center gap-3">
            <span class="text-sm text-base-content/70">Logo预览:</span>
            <img 
               :src="linkForm.logoUrl" 
               alt="Logo预览" 
               class="w-8 h-8 rounded object-cover border border-base-300"
               @error="($event.target as HTMLImageElement).style.display='none'"
             >
          </div>
          
          <div>
            <label class="label cursor-pointer justify-start gap-3">
              <input 
                v-model="linkForm.status" 
                type="checkbox" 
                class="toggle toggle-success" 
                :true-value="1"
                :false-value="0"
              >
              <span class="label-text">启用状态</span>
            </label>
          </div>
        </div>
        
        <div class="flex justify-end gap-2 mt-6 pt-4 border-t border-base-300">
          <button class="btn" @click="showAddLinkModal = false">取消</button>
          <button class="btn btn-primary" @click="saveLink">保存</button>
        </div>
       </fieldset>
       </div>
       <div class="modal-backdrop" @click="showAddLinkModal = false"></div>
     </dialog>
  </div>
</template>

<style scoped>
</style>
