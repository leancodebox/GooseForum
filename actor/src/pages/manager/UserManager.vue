<script setup lang="ts">
import {
  NButton,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSelect,
  NSpace,
  NSwitch,
  NTag,
  NCard,
  useMessage
} from 'naive-ui'
import {h, onMounted, ref} from 'vue'
import {editUser, getAllRoleItem, getUserList} from '@/service/request'
import {Ref, UnwrapRef} from "@vue/reactivity";

const message = useMessage()
type UserItem = {
  userId: number | null
  username: string | null
  email: string | null
  createTime: string | null
  roleList: any
  status: number | null
  validate: number | null
  prestige: number | null
}
const data: Ref<UnwrapRef<UserItem[]>> = ref([])
let isSmallScreen = ref(window.innerWidth < 800)

// 检查屏幕尺寸
function checkScreenSize() {
  isSmallScreen.value = window.innerWidth < 800
}

onMounted(() => {
  checkScreenSize()
  window.addEventListener('resize', checkScreenSize)
  getAllRoleItem().then(r => {
    roleOption.value = r.result
  })
  showUserList()
})

let columns = [
  {
    title: 'id',
    key: 'userId'
  },
  {
    title: 'username',
    key: 'username'
  },
  {
    title: 'email',
    key: 'email'
  },
  {
    title: '角色',
    key: 'role',
    render(row: UserItem) {
      if (!row.roleList) {
        return []
      }
      let res = []
      for (const rowKey in row.roleList) {
        res.push(h(
            NTag,
            {
              strong: true,
              tertiary: true,
              size: 'small',
            },
            {default: () => row.roleList[rowKey].name}
        ))
      }
      return res
    }
  },
  {
    title: '状态',
    key: 'status',
    render(row: UserItem) {
      if (row.status === 0) {
        return h(NTag, {type: 'success'}, () => "正常")
      } else {
        return h(NTag, {type: 'error'}, () => "冻结")
      }
    }
  },
  {
    title: '验证',
    key: 'validate',
    render(row: UserItem) {
      if (row.validate === 1) {
        return h(NTag, {type: 'success'}, () => "验证")
      } else {
        return h(NTag, {type: 'warning'}, () => "未验证")
      }
    }
  },
  {
    title: '声望',
    key: 'prestige'
  },
  {
    title: '创建时间',
    key: 'createTime'
  },
  {
    title: 'Action',
    key: 'actions',
    render(row: UserItem) {
      return h(NSpace, {}, {
        default: () => [
          h(
              NButton,
              {
                type: 'primary',
                size: 'small',
                onClick: () => handleEdit(row)
              },
              {default: () => '编辑'}
          ),
          h(
              NButton,
              {
                type: 'warning',
                size: 'small',
                onClick: () => {
                  message.info(`冻结用户 ${row.username}`)
                }
              },
              {default: () => '冻结'}
          )
        ]
      })
    }
  }
]

let roleOption = ref([])
let pagination = {
  page: 1,
  pageSize: 10,
  itemCount: 0,
  prefix({itemCount}) {
    return `共 ${itemCount} 条`
  }
}

let showModal = ref(false)
let userEntity = ref({
  userId: 0,
  username: "",
  roleId: [],
  status: 0,
  validate: 0
})

function handleEdit(row: UserItem) {
  showModal.value = true
  userEntity.value = {
    userId: row.userId,
    username: row.username,
    roleId: !!row.roleList ? row.roleList.map(item => {
      return item.value
    }) : [],
    status: row.status,
    validate: row.validate
  }
}

function showUserList() {
  getUserList().then(r => {
    data.value = r.result.list
  })
}

function userEdit4Role() {
  let req = userEntity.value
  editUser(req.userId, req.status, req.validate, req.roleId)
      .then(() => {
        showModal.value = false
        showUserList()
        message.success('更新成功')
      })
}
</script>

<template>
  <!-- 编辑用户弹窗 -->
  <n-modal
      v-model:show="showModal"
      style="width: 90%; max-width: 600px"
      preset="dialog"
      title="编辑用户"
      positive-text="确认"
      negative-text="取消"
      @positive-click="userEdit4Role"
      @negative-click="() => showModal = false"
  >
    <n-form
        ref="formRef"
        :model="userEntity"
        label-placement="left"
        label-width="auto"
        :style="{
          maxWidth: '640px'
        }"
    >
      <n-form-item label="用户名" path="username">
        <n-input v-model:value="userEntity.username" disabled></n-input>
      </n-form-item>
      <n-form-item label="是否冻结">
        <n-switch v-model:value="userEntity.status" :checked-value="1" :unchecked-value="0"/>
      </n-form-item>
      <n-form-item label="验证通过">
        <n-switch v-model:value="userEntity.validate" :checked-value="1" :unchecked-value="0"/>
      </n-form-item>
      <n-form-item label="角色" path="roleId">
        <n-select v-model:value="userEntity.roleId" multiple :options="roleOption"></n-select>
      </n-form-item>
    </n-form>
  </n-modal>

  <!-- PC端显示表格 -->
  <div v-if="!isSmallScreen" class="pc-view">
    <n-data-table
        :columns="columns"
        :data="data"
        :pagination="pagination"
        :bordered="false"
        striped
        flex-height
        style="height: calc(100vh - var(--header-height) - 28px);"
    />
  </div>

  <!-- 移动端显示卡片 -->
  <div v-else class="mobile-view">
    <n-space vertical>
      <n-card
          v-for="item in data"
          :key="item.userId"
          class="user-card"
          :bordered="false"
          size="small"
      >
        <n-space vertical>
          <div class="user-title">{{ item.username }}</div>
          <div class="user-email">{{ item.email }}</div>
          
          <n-space wrap>
            <template v-if="item.roleList">
              <n-tag
                  v-for="role in item.roleList"
                  :key="role.value"
                  size="small"
                  :bordered="false"
              >
                {{ role.name }}
              </n-tag>
            </template>
          </n-space>

          <n-space>
            <n-tag :type="item.status === 0 ? 'success' : 'error'">
              {{ item.status === 0 ? '正常' : '冻结' }}
            </n-tag>
            <n-tag :type="item.validate === 1 ? 'success' : 'warning'">
              {{ item.validate === 1 ? '已验证' : '未验证' }}
            </n-tag>
          </n-space>

          <div class="user-info">
            <span>声望: {{ item.prestige }}</span>
            <span class="user-time">创建: {{ item.createTime }}</span>
          </div>

          <n-space justify="end">
            <n-button size="small" type="primary" @click="handleEdit(item)">
              编辑
            </n-button>
            <n-button size="small" type="warning" @click="message.info(`冻结用户 ${item.username}`)">
              冻结
            </n-button>
          </n-space>
        </n-space>
      </n-card>
    </n-space>

    <!-- 移动端分页 -->
    <div class="mobile-pagination">
      <n-space justify="center" align="center">
        <n-button
            size="small"
            :disabled="pagination.page === 1"
            @click="pagination.page--"
        >
          上一页
        </n-button>
        <span>{{ pagination.page }}</span>
        <n-button
            size="small"
            @click="pagination.page++"
        >
          下一页
        </n-button>
      </n-space>
    </div>
  </div>
</template>

<style scoped>
.mobile-view {
  padding: 8px;
  padding-bottom: 60px;
}

.user-card {
  margin-bottom: 12px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.user-title {
  font-size: 16px;
  font-weight: bold;
  color: #333;
}

.user-email {
  font-size: 14px;
  color: #666;
}

.user-info {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #999;
}

.user-time {
  color: #999;
}

.mobile-pagination {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px;
  background: white;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.08);
  z-index: 1;
}
</style>
