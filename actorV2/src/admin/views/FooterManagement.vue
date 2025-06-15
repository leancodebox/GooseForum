<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NButton, NCard, NInput, NList, NListItem, NModal, NSpace, NThing, useMessage } from 'naive-ui'
import draggable from 'vuedraggable'
import type { FooterGroup, FooterItem } from '../types/adminInterfaces.ts'

const message = useMessage()
const footerGroups = ref<FooterGroup[]>([])
const loading = ref(false)
const startEdit = ref(false)
const showItemModal = ref(false)
const modalMode = ref('add') // 'add' | 'edit'
const editingItem = ref({
  id: 0,
  title: '',
  url: '',
  sort: 1,
  status: 1,
  createTime: '',
  groupIndex: 0,
  itemIndex: 0
})

// 获取Footer分组列表
const fetchFooterGroups = async () => {
  loading.value = true
  try {
    // 模拟API调用 - 实际项目中应该调用真实的API
    const mockData: FooterGroup[] = [
      {
        name: 'SERVICES',
        items: [
          {
            id: 1,
            title: 'Github',
            url: 'https://github.com',
            sort: 1,
            status: 1,
            createTime: '2024-01-01 10:00:00'
          }
        ]
      },
      {
        name: 'LEGAL',
        items: [
          {
            id: 2,
            title: 'LICENSE',
            url: '/license',
            sort: 1,
            status: 1,
            createTime: '2024-01-01 10:05:00'
          }
        ]
      },
      {
        name: 'TEAM',
        items: [
          {
            id: 3,
            title: 'About',
            url: '/about',
            sort: 1,
            status: 1,
            createTime: '2024-01-01 10:10:00'
          }
        ]
      }
    ]
    
    footerGroups.value = mockData
    message.success('获取Footer分组列表成功')
  } catch (error) {
    message.error('获取Footer分组列表失败')
    console.error('获取Footer分组列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 添加/编辑Footer项目
function showAddItem(groupIndex: number) {
  modalMode.value = 'add'
  editingItem.value = {
    id: 0,
    title: '',
    url: '',
    sort: footerGroups.value[groupIndex].items.length + 1,
    status: 1,
    createTime: '',
    groupIndex,
    itemIndex: 0
  }
  showItemModal.value = true
}

function showEditItem(groupIndex: number, itemIndex: number, item: FooterItem) {
  modalMode.value = 'edit'
  editingItem.value = {
    id: item.id,
    title: item.title,
    url: item.url,
    sort: item.sort,
    status: item.status,
    createTime: item.createTime,
    groupIndex,
    itemIndex
  }
  showItemModal.value = true
}

function handleSaveItem() {
  if (!editingItem.value.title || !editingItem.value.url) {
    message.warning('请填写完整的Footer项目信息')
    return
  }

  const { groupIndex, itemIndex } = editingItem.value
  const newItem: FooterItem = {
    id: editingItem.value.id || Date.now(),
    title: editingItem.value.title,
    url: editingItem.value.url,
    sort: editingItem.value.sort,
    status: editingItem.value.status,
    createTime: editingItem.value.createTime || new Date().toLocaleString()
  }

  if (modalMode.value === 'add') {
    footerGroups.value[groupIndex].items.push(newItem)
    message.success('添加成功')
  } else {
    footerGroups.value[groupIndex].items[itemIndex] = newItem
    message.success('修改成功')
  }

  showItemModal.value = false
}

// 拖拽排序处理
const handleItemDrag = (groupIndex: number, event: any) => {
  const { removed, added } = event
  if (!removed || !added) return

  const items = [...footerGroups.value[groupIndex].items]
  const [movedItem] = items.splice(removed.oldIndex, 1)
  items.splice(added.newIndex, 0, movedItem)
  footerGroups.value[groupIndex].items = items
}

const handleGroupDrag = (event: any) => {
  const { removed, added } = event
  if (!removed || !added) return

  const groups = [...footerGroups.value]
  const [movedItem] = groups.splice(removed.oldIndex, 1)
  groups.splice(added.newIndex, 0, movedItem)
  footerGroups.value = groups
}

// 删除操作
function handleDeleteItem(groupIndex: number, itemIndex: number) {
  footerGroups.value[groupIndex].items.splice(itemIndex, 1)
  message.success('删除成功')
}

// 保存配置
async function saveConfig() {
  startEdit.value = !startEdit.value
  if (startEdit.value) return

  try {
    // 调用API保存配置
    // await saveFooterGroups(footerGroups.value)
    message.success('保存成功')
  } catch (error) {
    message.error('保存失败')
    console.error(error)
  }
}

// 添加新分组
function handleAddGroup() {
  footerGroups.value.push({
    name: '新分组',
    items: []
  })
  message.success('分组添加成功')
}

// 删除分组
function handleDeleteGroup(groupIndex: number) {
  footerGroups.value.splice(groupIndex, 1)
  message.success('分组删除成功')
}

// 初始化加载
onMounted(() => {
  fetchFooterGroups()
})


</script>

<template>
  <div class="footer-management">
    <div class="header-section">
      <h2>Footer管理</h2>
      <p class="description">管理网站Footer区域的分组和链接项目，支持拖拽排序。</p>
    </div>

    <!-- 操作栏 -->
    <div class="toolbar">
      <n-space justify="space-between">
        <n-button @click="handleAddGroup">
          添加分组
        </n-button>
        
        <n-button type="primary" @click="saveConfig">
          {{ startEdit ? '保存配置' : '编辑配置' }}
        </n-button>
      </n-space>
    </div>

    <!-- Footer分组列表 -->
    <div class="groups-container">
      <draggable
        v-model="footerGroups"
        :disabled="!startEdit"
        item-key="name"
        @change="handleGroupDrag"
        class="groups-list"
      >
        <template #item="{ element: group, index: groupIndex }">
          <n-card class="group-card" :class="{ 'edit-mode': startEdit }">
            <template #header>
              <div class="group-header">
                <n-input
                  v-if="startEdit"
                  v-model:value="group.name"
                  placeholder="分组名称"
                  style="max-width: 200px"
                />
                <span v-else class="group-name">{{ group.name }}</span>
                
                <n-space v-if="startEdit">
                  <n-button size="small" @click="showAddItem(groupIndex)">
                    添加项目
                  </n-button>
                  <n-button size="small" type="error" @click="handleDeleteGroup(groupIndex)">
                    删除分组
                  </n-button>
                </n-space>
              </div>
            </template>
            
            <div class="items-container">
              <draggable
                v-model="group.items"
                :disabled="!startEdit"
                item-key="id"
                @change="(event) => handleItemDrag(groupIndex, event)"
                class="items-list"
              >
                <template #item="{ element: item, index: itemIndex }">
                  <div class="item-card" :class="{ 'edit-mode': startEdit }">
                    <div class="item-content">
                      <div class="item-info">
                        <h4 class="item-title">{{ item.title }}</h4>
                        <a :href="item.url" target="_blank" class="item-url">{{ item.url }}</a>
                      </div>
                      
                      <div v-if="startEdit" class="item-actions">
                        <n-button size="small" @click="showEditItem(groupIndex, itemIndex, item)">
                          编辑
                        </n-button>
                        <n-button size="small" type="error" @click="handleDeleteItem(groupIndex, itemIndex)">
                          删除
                        </n-button>
                      </div>
                    </div>
                  </div>
                </template>
              </draggable>
              
              <div v-if="group.items.length === 0" class="empty-items">
                <p>暂无项目</p>
                <n-button v-if="startEdit" size="small" @click="showAddItem(groupIndex)">
                  添加第一个项目
                </n-button>
              </div>
            </div>
          </n-card>
        </template>
      </draggable>
    </div>

    <!-- 添加/编辑项目模态框 -->
    <n-modal
      v-model:show="showItemModal"
      preset="dialog"
      :title="modalMode === 'add' ? '添加Footer项目' : '编辑Footer项目'"
      positive-text="保存"
      negative-text="取消"
      @positive-click="handleSaveItem"
    >
      <n-space vertical>
        <n-input
          v-model:value="editingItem.title"
          placeholder="项目标题"
        />
        <n-input
          v-model:value="editingItem.url"
          placeholder="链接地址"
        />
      </n-space>
    </n-modal>
  </div>
</template>

<style scoped>
.footer-management {
  padding: 20px;
}

.header-section {
  margin-bottom: 24px;
}

.header-section h2 {
  margin: 0 0 8px 0;
  color: #333;
}

.description {
  color: #666;
  margin: 0;
}

.toolbar {
  margin-bottom: 24px;
}

.groups-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.group-card {
  border: 1px solid #e0e0e0;
  transition: all 0.3s ease;
}

.group-card.edit-mode {
  border-color: #18a058;
  box-shadow: 0 2px 8px rgba(24, 160, 88, 0.1);
}

.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.group-name {
  font-weight: 600;
  font-size: 16px;
  color: #333;
}

.items-container {
  min-height: 60px;
}

.items-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.item-card {
  padding: 12px;
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  background: #fafafa;
  transition: all 0.3s ease;
}

.item-card.edit-mode {
  border-color: #d9d9d9;
  background: #fff;
}

.item-card:hover {
  border-color: #18a058;
}

.item-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.item-info {
  flex: 1;
}

.item-title {
  margin: 0 0 4px 0;
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.item-url {
  color: #18a058;
  text-decoration: none;
  font-size: 12px;
}

.item-url:hover {
  text-decoration: underline;
}

.item-actions {
  display: flex;
  gap: 8px;
}

.empty-items {
  text-align: center;
  padding: 40px 20px;
  color: #999;
}

.empty-items p {
  margin: 0 0 12px 0;
}

/* 拖拽样式 */
.sortable-ghost {
  opacity: 0.5;
}

.sortable-chosen {
  transform: scale(1.02);
}

.sortable-drag {
  transform: rotate(5deg);
}
</style>