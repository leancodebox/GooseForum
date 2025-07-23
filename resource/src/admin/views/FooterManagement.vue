<template>
  <div class="space-y-6">
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-4">
      <h1 class="text-2xl font-normal text-base-content">页脚管理</h1>
      <button 
        @click="saveFooterConfig" 
        class="btn btn-primary btn-sm"
        :disabled="loading"
      >
        <span v-if="loading" class="loading loading-spinner loading-sm"></span>
        {{ loading ? '保存中...' : '保存配置' }}
      </button>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <!-- HTML内容部分 -->
      <div class="card bg-base-100 shadow-sm border border-base-300">
        <div class="card-body p-4">
          <div class="flex justify-between items-center mb-3">
            <h2 class="card-title text-lg font-normal">HTML内容列表</h2>
            <button @click="addPItem" class="btn btn-success btn-xs">
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
              </svg>
              添加HTML项
            </button>
          </div>
          
          <div class="space-y-3 max-h-96 overflow-y-auto">
            <draggable
              v-model="footerConfig.primary"
              :item-key="(item, index) => `html-${index}`"
              class="space-y-3"
              :animation="200"
              ghost-class="opacity-50"
            >
              <template #item="{ element: htmlItem, index: htmlIndex }">
                 <div class="card bg-base-200 border border-base-300">
                   <div class="card-body p-3">
                     <div class="flex justify-between items-center gap-2">
                       <div class="cursor-move text-base-content/60">
                         <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                           <path d="M7 2a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM7 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM7 14a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 2a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 14a2 2 0 1 1-4 0 2 2 0 0 1 4 0z"></path>
                         </svg>
                       </div>
                       <input 
                         v-model="htmlItem.content"
                         placeholder="请输入HTML项内容"
                         class="input input-bordered input-sm flex-1"
                       />
                       <button 
                         @click="removePItem(htmlIndex)"
                         class="btn btn-error btn-xs"
                       >
                         删除
                       </button>
                     </div>
                   </div>
                 </div>
               </template>
            </draggable>
            
            <div v-if="footerConfig.primary.length === 0" class="text-center py-8 text-base-content/60">
              暂无HTML项，点击上方按钮添加
            </div>
          </div>
        </div>
      </div>

      <!-- 链接分组部分 -->
      <div class="card bg-base-100 shadow-sm border border-base-300">
        <div class="card-body p-4">
          <div class="flex justify-between items-center mb-3">
            <h2 class="card-title text-lg font-normal">链接分组</h2>
            <button @click="addGroup" class="btn btn-success btn-xs">
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
              </svg>
              添加分组
            </button>
          </div>
          
          <div class="space-y-3">
            <draggable
              v-model="footerConfig.list"
              :item-key="(item, index) => `group-${index}`"
              class="space-y-3"
              :animation="200"
              ghost-class="opacity-50"
            >
              <template #item="{ element: group, index: groupIndex }">
                <div class="card bg-base-200 border border-base-300">
                  <div class="card-body p-3">
                    <div class="flex justify-between items-center mb-2">
                      <div class="flex items-center gap-2 flex-1">
                        <div class="cursor-move text-base-content/60">
                          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                            <path d="M7 2a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM7 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM7 14a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 2a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 14a2 2 0 1 1-4 0 2 2 0 0 1 4 0z"></path>
                          </svg>
                        </div>
                        <input 
                          v-model="group.name" 
                          placeholder="分组名称"
                          class="input input-bordered input-sm flex-1"
                        />
                      </div>
                      <button 
                        @click="removeGroup(groupIndex)" 
                        class="btn btn-error btn-xs"
                      >
                        删除
                      </button>
                    </div>
                 
                    <div class="space-y-2">
                      <draggable
                        v-model="group.children"
                        :item-key="(item, index) => `child-${groupIndex}-${index}`"
                        class="space-y-2"
                        :animation="200"
                        ghost-class="opacity-50"
                      >
                        <template #item="{ element: child, index: childIndex }">
                          <div class="flex gap-2 items-center">
                            <div class="cursor-move text-base-content/40">
                              <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                                <path d="M7 2a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM7 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM7 14a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 2a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0zM17 14a2 2 0 1 1-4 0 2 2 0 0 1 4 0z"></path>
                              </svg>
                            </div>
                            <input 
                              v-model="child.name" 
                              placeholder="链接名称"
                              class="input input-bordered input-xs flex-1"
                            />
                            <input 
                              v-model="child.url" 
                              placeholder="链接地址"
                              class="input input-bordered input-xs flex-1"
                            />
                            <button 
                              @click="removeChild(groupIndex, childIndex)" 
                              class="btn btn-error btn-xs"
                            >
                              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                              </svg>
                            </button>
                          </div>
                        </template>
                      </draggable>
                      <button 
                        @click="addChild(groupIndex)" 
                        class="btn btn-success btn-xs w-full"
                      >
                        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
                        </svg>
                        添加链接
                      </button>
                    </div>
                  </div>
                </div>
              </template>
            </draggable>
            
            <div v-if="footerConfig.list.length === 0" class="text-center py-8 text-base-content/60">
              暂无分组，点击上方按钮添加
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getFooterLinks, saveFooterLinks, type FooterConfig, type FooterGroup, type FooterItem, type HtmlItem } from '../utils/adminService'
import draggable from 'vuedraggable'

const loading = ref(false)
const footerConfig = ref<FooterConfig>({
  primary: [],
  list: []
})

// 加载页脚配置
const loadFooterConfig = async () => {
  try {
    const response = await getFooterLinks()
    if (response.code === 0) {
      footerConfig.value = {
        primary: response.result.primary || [],
        list: response.result.list || []
      }
    }
  } catch (error) {
    console.error('加载页脚配置失败:', error)
    alert('加载页脚配置失败')
  }
}

// 保存页脚配置
const saveFooterConfig = async () => {
  loading.value = true
  try {
    const response = await saveFooterLinks(footerConfig.value)
    if (response.code === 0) {
      alert('保存成功')
    } else {
      alert('保存失败')
    }
  } catch (error) {
    console.error('保存页脚配置失败:', error)
    alert('保存失败')
  } finally {
    loading.value = false
  }
}

// 添加HTML项
const addPItem = () => {
  footerConfig.value.primary.push({
    content: ''
  })
}

// 删除HTML项
const removePItem = (index: number) => {
  footerConfig.value.primary.splice(index, 1)
}

// 添加分组
const addGroup = () => {
  footerConfig.value.list.push({
    name: '',
    children: []
  })
}

// 删除分组
const removeGroup = (groupIndex: number) => {
  footerConfig.value.list.splice(groupIndex, 1)
}

// 添加子链接
const addChild = (groupIndex: number) => {
  footerConfig.value.list[groupIndex].children.push({
    name: '',
    url: ''
  })
}

// 删除子链接
const removeChild = (groupIndex: number, childIndex: number) => {
  footerConfig.value.list[groupIndex].children.splice(childIndex, 1)
}

onMounted(() => {
  loadFooterConfig()
})
</script>

<style scoped>
/* 使用 DaisyUI 主题，无需自定义样式 */
</style>