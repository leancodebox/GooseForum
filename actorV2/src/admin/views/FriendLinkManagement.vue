<script setup lang="ts">
import {onMounted, ref} from 'vue'
import {NButton, NCard, NInput, NList, NListItem, NModal, NSpace, NThing, useMessage} from 'naive-ui'
import draggable from 'vuedraggable'
import {getFriendLinks} from "@/admin/utils/authService.ts";
import type {FriendLinksGroup} from "@/admin/types/adminInterfaces.ts";

const message = useMessage()
const friendLinks = ref<FriendLinksGroup[]>([])
const loading = ref(false)
const startEdit = ref(false)
const showLinkModal = ref(false)
const modalMode = ref('add') // 'add' | 'edit'
const editingLink = ref({
  name: '',
  desc: '',
  url: '',
  logoUrl: '',
  status: 1,
  groupIndex: 0,
  linkIndex: 0
})

// 获取友情链接列表
const fetchFriendLinks = async () => {
  loading.value = true
  try {
    const res = await getFriendLinks()
    if (res.code === 0) {
      friendLinks.value = res.result // 直接使用接口返回的数据结构
      message.success('获取友情链接成功')
    } else {
      message.error(res.message || '获取友情链接失败')
    }
  } catch (error) {
    message.error('获取友情链接失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

// 添加/编辑链接
function showAddLink(groupIndex) {
  modalMode.value = 'add'
  editingLink.value = {
    name: '',
    desc: '',
    url: '',
    status: 1,
    logoUrl: '',
    groupIndex,
    linkIndex: 0
  }
  showLinkModal.value = true
}

function showEditLink(groupIndex, linkIndex, link) {
  modalMode.value = 'edit'
  editingLink.value = {
    name: link.name,
    desc: link.desc,
    url: link.url,
    logoUrl: link.logoUrl,
    status: link.status,
    groupIndex,
    linkIndex
  }
  showLinkModal.value = true
}

function handleSaveLink() {
  if (!editingLink.value.name || !editingLink.value.url) {
    message.warning('请填写完整的链接信息')
    return
  }

  const {groupIndex, linkIndex} = editingLink.value
  const newLink = {
    name: editingLink.value.name,
    desc: editingLink.value.desc,
    url: editingLink.value.url,
    logoUrl: editingLink.value.logoUrl,
    status: editingLink.value.status,
  }

  if (modalMode.value === 'add') {
    friendLinks.value[groupIndex].links.push(newLink)
    message.success('添加成功')
  } else {
    friendLinks.value[groupIndex].links[linkIndex] = newLink
    message.success('修改成功')
  }

  showLinkModal.value = false
}

// 拖拽排序处理
const handleLinkDrag = (groupIndex, event) => {
  const {removed, added} = event
  if (!removed || !added) return

  const links = [...friendLinks.value[groupIndex].links]
  const [movedItem] = links.splice(removed.oldIndex, 1)
  links.splice(added.newIndex, 0, movedItem)
  friendLinks.value[groupIndex].links = links
}

const handleGroupDrag = (event) => {
  const {removed, added} = event
  if (!removed || !added) return

  const groups = [...friendLinks.value]
  const [movedItem] = groups.splice(removed.oldIndex, 1)
  groups.splice(added.newIndex, 0, movedItem)
  friendLinks.value = groups
}

// 删除操作
function handleDeleteLink(groupIndex, linkIndex) {
  friendLinks.value[groupIndex].links.splice(linkIndex, 1)
  message.success('删除成功')
}

// 保存配置
async function saveConfig() {
  startEdit.value = !startEdit.value
  if (startEdit.value) return

  try {
    // 调用API保存配置
    // await saveFriendLinks(friendLinks.value)
    message.success('保存成功')
  } catch (error) {
    message.error('保存失败')
    console.error(error)
  }
}

// 添加新分组
function handleAddGroup() {
  friendLinks.value.push({
    name: '新分组',
    links: []
  })
  message.success('分组添加成功')
}

// 删除分组
function handleDeleteGroup(groupIndex) {
  friendLinks.value.splice(groupIndex, 1)
  message.success('分组删除成功')
}

onMounted(() => {
  fetchFriendLinks()
})
</script>

<template>
  <n-modal v-model:show="showLinkModal">
    <n-card
        style="width: 520px; border-radius: 16px; box-shadow: 0 4px 24px rgba(0,0,0,0.08);"
        :title="modalMode === 'add' ? '添加友情链接' : '编辑友情链接'"
        :bordered="false"
        size="huge"
    >
      <n-space vertical size="large">
        <n-grid cols="12" x-gap="16" y-gap="12">
          <n-grid-item span="3">
            <n-avatar
              :src="editingLink.logoUrl || 'https://naive-ui.oss-cn-beijing.aliyuncs.com/logo.png'"
              size="64"
              style="border-radius: 8px; border: 1px solid #eee; background: #fafbfc;"
            />
          </n-grid-item>
          <n-grid-item span="9">
            <n-form label-placement="left" label-width="60">
              <n-form-item label="名称">
                <n-input
                    v-model:value="editingLink.name"
                    placeholder="请输入链接名称"
                    clearable
                    maxlength="20"
                    show-count
                />
              </n-form-item>
              <n-form-item label="描述">
                <n-input
                    v-model:value="editingLink.desc"
                    placeholder="简要描述（可选）"
                    clearable
                    maxlength="40"
                    show-count
                />
              </n-form-item>
              <n-form-item label="地址">
                <n-input
                    v-model:value="editingLink.url"
                    placeholder="请输入链接地址"
                    clearable
                />
              </n-form-item>
              <n-form-item label="Logo">
                <n-input
                    v-model:value="editingLink.logoUrl"
                    placeholder="Logo图片URL（可选）"
                    clearable
                />
              </n-form-item>
              <n-form-item label="状态">
                <n-switch v-model:value="editingLink.status" :checked-value="1" :unchecked-value="0"/>
              </n-form-item>
            </n-form>
          </n-grid-item>
        </n-grid>
        <n-divider style="margin: 0 0 8px 0;"/>
        <n-space justify="end">
          <n-button @click="showLinkModal = false">取消</n-button>
          <n-button type="primary" @click="handleSaveLink">
            {{ modalMode === 'add' ? '添加' : '保存' }}
          </n-button>
        </n-space>
      </n-space>
    </n-card>
  </n-modal>

  <n-list>
    <n-list-item>
      <n-space>
        <n-button type="primary" @click="fetchFriendLinks" :size="'small'">
          刷新
        </n-button>
        <n-button type="primary" :size="'small'" @click="saveConfig">
          {{ !startEdit ? '启动编辑' : '保存' }}
        </n-button>
        <n-button type="success" :size="'small'" @click="handleAddGroup" disabled>
          添加分组
        </n-button>
      </n-space>
    </n-list-item>

    <draggable
        v-model="friendLinks"
        v-if="startEdit"
        @change="handleGroupDrag"
        item-key="title"
        handle=".group-drag-handle"
    >
      <template #item="{element: groupItem, index: groupIndex}">
        <n-list-item>
          <n-thing>
            <template #header>
              <n-space v-if="startEdit">
                <n-button circle size="small" class="group-drag-handle">
                  ≡
                </n-button>
                <n-input v-model:value="groupItem.name"></n-input>
                <n-button
                    type="error"
                    size="small"
                    @click="handleDeleteGroup(groupIndex)"
                >
                  删除
                </n-button>
              </n-space>
              <span v-else>{{ groupItem.name }}</span>
            </template>
            <n-space>
              <n-card
                v-for="(item, linkIndex) in groupItem.links"
                :title="item.name"
                size="small"
                :bordered="false"
                class="link-item"
                @click="showEditLink(groupIndex, linkIndex, item)"
                style="cursor:pointer;position:relative;"
              >
                <template #header-extra>
                  <n-button
                    v-if="startEdit"
                    circle
                    type="error"
                    size="small"
                    style="position:absolute;top:8px;right:8px;z-index:2;"
                    @click.stop="handleDeleteLink(groupIndex, linkIndex)"
                  >删</n-button>
                </template>
                <n-space vertical>
                  <n-space align="center">
                    <n-image
                      v-if="item.logoUrl"
                      :src="item.logoUrl"
                      width="50"
                      height="50"
                      object-fit="contain"
                    />
                    <n-space vertical>
                      <n-text>{{ item.desc }}</n-text>
                      <n-text depth="3">{{ item.url }}</n-text>
                    </n-space>
                  </n-space>
                  <n-button
                    tag="a"
                    :href="item.url"
                    target="_blank"
                    type="primary"
                    size="small"
                    @click.stop
                  >访问链接</n-button>
                </n-space>
              </n-card>
              <n-button
                  dashed
                  type="primary"
                  :size="'small'"
                  v-if="startEdit"
                  @click="showAddLink(groupIndex)"
              >
                添加链接
              </n-button>
            </n-space>
          </n-thing>
        </n-list-item>
      </template>
    </draggable>

    <template v-if="!startEdit">
      <n-list-item v-for="groupItem in friendLinks">
        <n-thing>
          <template #header>
            <span>{{ groupItem.name }}</span>
          </template>
          <n-space>
            <n-card
              v-for="item in groupItem.links"
              :title="item.name"
              size="small"
              :bordered="false"
            >
              <n-space vertical>
                <n-space align="center">
                  <n-image
                    v-if="item.logoUrl"
                    :src="item.logoUrl"
                    width="50"
                    height="50"
                    object-fit="contain"
                  />
                  <n-space vertical>
                    <n-text>{{ item.desc }}</n-text>
                    <n-text depth="3">{{ item.url }}</n-text>
                  </n-space>
                </n-space>
                <n-button
                  tag="a"
                  :href="item.url"
                  target="_blank"
                  type="primary"
                  size="small"
                >
                  访问链接
                </n-button>
              </n-space>
            </n-card>
          </n-space>
        </n-thing>
      </n-list-item>
    </template>
  </n-list>
</template>

<style scoped>
.sortable-ghost {
  opacity: 0.5;
  background: #c8ebfb;
}

.sortable-drag {
  opacity: 0.9;
}

.links-container {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.link-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background: #fff;
  transition: all 0.2s ease;
}

.link-item:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.link-drag-handle,
.group-drag-handle {
  cursor: move;
  user-select: none;
}

.n-space {
  align-items: center;
}

/* 卡片样式优化 */
.n-card {
  margin-bottom: 12px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
}
</style>
