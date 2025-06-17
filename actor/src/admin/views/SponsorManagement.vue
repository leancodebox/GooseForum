<script setup lang="ts">
import {computed, h, onMounted, reactive, ref} from 'vue'
import {NButton, NSpace, NTag, useMessage, NTabs, NTabPane} from 'naive-ui'
import {AddOutline, SearchOutline, TrashOutline,CreateOutline} from '@vicons/ionicons5'
import type {Sponsor} from "../types/adminInterfaces.ts";

const message = useMessage()
const searchText = ref('')
const loading = ref(false)
const activeTab = ref('sponsors')

// 赞助商数据
const sponsors = ref<Sponsor[]>([])
const individualSponsors = ref<Sponsor[]>([])

// 分页设置
const pagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  onChange: (page: number) => {
    pagination.page = page
    fetchSponsors()
  }
})

// 赞助商等级选项
const sponsorLevels = [
  {label: '钻石', value: 1},
  {label: '白金', value: 2},
  {label: '黄金', value: 3},
  {label: '白银', value: 4}
]

// 获取赞助商列表
const fetchSponsors = async () => {
  loading.value = true
  try {
    // TODO: 调用API获取赞助商数据
    // const response = await getSponsorList()
    // sponsors.value = response.result.list || []
    // pagination.itemCount = response.result.total || 0
  } catch (error) {
    message.error('获取赞助商列表失败')
    console.error('获取赞助商列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 获取个人赞助者列表
const fetchIndividualSponsors = async () => {
  loading.value = true
  try {
    // TODO: 调用API获取个人赞助者数据
    // const response = await getIndividualSponsorList()
    // individualSponsors.value = response.result.list || []
  } catch (error) {
    message.error('获取个人赞助者列表失败')
    console.error('获取个人赞助者列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 表格列定义 - 赞助商
const sponsorColumns = [
  {
    title: 'ID',
    key: 'id'
  },
  {
    title: '名称',
    key: 'name'
  },
  {
    title: '等级',
    key: 'level',
    render(row: Sponsor) {
      const levelMap = {
        1: {text: '钻石', type: 'success'},
        2: {text: '白金', type: 'info'},
        3: {text: '黄金', type: 'warning'},
        4: {text: '白银', type: 'default'}
      }
      const level = levelMap[row.level] || {text: row.level, type: 'default'}
      return h(
          NTag,
          {type: level.type},
          {default: () => level.text}
      )
    }
  },
  {
    title: '赞助金额',
    key: 'amount'
  },
  {
    title: '赞助时间',
    key: 'createdAt'
  },
  {
    title: '操作',
    key: 'actions',
    render(row: Sponsor) {
      return h(
          NSpace,
          {},
          {
            default: () => [
              h(
                  NButton,
                  {
                    size: 'small',
                    quaternary: true,
                    onClick: () => handleEditSponsor(row)
                  },
                  {
                    default: () => '编辑',
                    icon: () => h(CreateOutline)
                  }
              ),
              h(
                  NButton,
                  {
                    size: 'small',
                    quaternary: true,
                    type: 'error',
                    onClick: () => handleDeleteSponsor(row)
                  },
                  {
                    default: () => '删除',
                    icon: () => h(TrashOutline)
                  }
              )
            ]
          }
      )
    }
  }
]

// 表格列定义 - 个人赞助者
const individualSponsorColumns = [
  {
    title: 'ID',
    key: 'id'
  },
  {
    title: '姓名',
    key: 'name'
  },
  {
    title: '赞助金额',
    key: 'amount'
  },
  {
    title: '赞助时间',
    key: 'createdAt'
  },
  {
    title: '操作',
    key: 'actions',
    render(row: Sponsor) {
      return h(
          NSpace,
          {},
          {
            default: () => [
              h(
                  NButton,
                  {
                    size: 'small',
                    quaternary: true,
                    type: 'error',
                    onClick: () => handleDeleteSponsor(row)
                  },
                  {
                    default: () => '删除',
                    icon: () => h(TrashOutline)
                  }
              )
            ]
          }
      )
    }
  }
]

// 初始化加载
onMounted(() => {
  fetchSponsors()
  fetchIndividualSponsors()
})

// 编辑赞助商
const handleEditSponsor = (sponsor: Sponsor) => {
  // TODO: 实现编辑逻辑
  message.warning(`此功能尚未实现：编辑赞助商 ${sponsor.name}`)
}

// 删除赞助商
const handleDeleteSponsor = (sponsor: Sponsor) => {
  // TODO: 实现删除逻辑
  message.warning(`此功能尚未实现：删除赞助商 ${sponsor.name}`)
}

// 添加赞助商
const handleAddSponsor = () => {
  // TODO: 实现添加逻辑
  message.warning('此功能尚未实现：添加赞助商')
}
</script>

<template>
  <div>
    <n-tabs v-model:value="activeTab" type="line">
      <n-tab-pane name="sponsors" tab="赞助商管理">
        <div class="action-bar">
          <n-space>
            <n-input
                v-model:value="searchText"
                placeholder="搜索赞助商"
                clearable
            >
              <template #prefix>
                <n-icon>
                  <SearchOutline/>
                </n-icon>
              </template>
            </n-input>
            <n-button type="primary" @click="fetchSponsors">
              搜索
            </n-button>
            <n-button type="primary" @click="handleAddSponsor">
              <template #icon>
                <n-icon>
                  <AddOutline/>
                </n-icon>
              </template>
              添加赞助商
            </n-button>
          </n-space>
        </div>

        <n-data-table
            :columns="sponsorColumns"
            :data="sponsors"
            :pagination="pagination"
            :bordered="true"
            :loading="loading"
            striped
        />
      </n-tab-pane>

      <n-tab-pane name="individual" tab="个人赞助者">
        <div class="action-bar">
          <n-space>
            <n-input
                v-model:value="searchText"
                placeholder="搜索个人赞助者"
                clearable
            >
              <template #prefix>
                <n-icon>
                  <SearchOutline/>
                </n-icon>
              </template>
            </n-input>
            <n-button type="primary" @click="fetchIndividualSponsors">
              搜索
            </n-button>
          </n-space>
        </div>

        <n-data-table
            :columns="individualSponsorColumns"
            :data="individualSponsors"
            :bordered="true"
            :loading="loading"
            striped
        />
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<style scoped>
.action-bar {
  margin-bottom: 16px;
}
</style>