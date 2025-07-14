<template>
  <div class="space-y-6">
    <div class="card bg-base-100 shadow">
      <div class="card-body p-3">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
          <div class="form-control">
            <label class="floating-label">
              <input
                  v-model="searchParams.title"
                  type="text"
                  class="input input-sm input-bordered w-full"
                  placeholder="搜索工单标题"
                  @input="handleSearch"
              />
              <span>工单标题</span>
            </label>
          </div>

          <div class="form-control">
            <label class="floating-label">
              <select v-model="searchParams.type" class="select select-sm select-bordered w-full" @change="handleFilter">
                <option value="">全部类型</option>
                <option value="1">Bug反馈</option>
                <option value="2">功能建议</option>
                <option value="3">技术支持</option>
                <option value="4">账户问题</option>
                <option value="5">其他</option>
              </select>
              <span>工单类型</span>
            </label>
          </div>

          <div class="form-control">
            <label class="floating-label">
              <select v-model="searchParams.status" class="select select-sm select-bordered w-full" @change="handleFilter">
                <option value="">全部状态</option>
                <option value="1">待处理</option>
                <option value="2">处理中</option>
                <option value="3">已解决</option>
                <option value="4">已关闭</option>
              </select>
              <span>工单状态</span>
            </label>
          </div>

          <div class="form-control">
            <button class="btn btn-sm btn-outline" @click="resetSearch">
              <ArrowPathIcon class="w-4 h-4"/>
              重置
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 工单列表 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body p-0">
        <!-- 加载状态 -->
        <div v-if="loading" class="flex justify-center items-center py-12">
          <span class="loading loading-spinner loading-lg"></span>
        </div>

        <!-- 无数据状态 -->
        <div v-else-if="!applySheets.length" class="text-center py-12">
          <div class="text-base-content/50 mb-2">
            <DocumentTextIcon class="w-16 h-16 mx-auto mb-4 opacity-50"/>
          </div>
          <p class="text-base-content/70">暂无工单数据</p>
        </div>

        <!-- 工单列表 -->
        <div v-else class="bg-base-100 rounded-box p-4">
          <ul class="space-y-3">
            <li v-for="sheet in applySheets" :key="sheet.id"
                class="flex items-center justify-between p-4 hover:bg-base-200 rounded-lg transition-colors border border-base-300">
              <!-- 左侧：工单信息 -->
              <div class="flex items-center gap-3 flex-1 min-w-0">
                <!-- 工单图标 -->
                <div class="avatar flex-shrink-0">
                  <div class="w-10 h-10 rounded bg-base-200 flex items-center justify-center">
                    <DocumentTextIcon class="w-5 h-5 text-base-content/50"/>
                  </div>
                </div>

                <!-- 工单详情 -->
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2 mb-1">
                    <h4 class="font-semibold text-base truncate">#{{ sheet.id }} {{ sheet.title }}</h4>
                    <div class="flex items-center gap-1 flex-shrink-0">
                      <span class="badge badge-xs" :class="getTypeClass(sheet.type)">
                        {{ getTypeText(sheet.type) }}
                      </span>
                      <span class="badge badge-xs" :class="getStatusClass(sheet.status)">
                        {{ getStatusText(sheet.status) }}
                      </span>
                    </div>
                  </div>
                  <div class="flex items-center gap-2 text-sm text-base-content/70">
                    <span class="truncate">{{ sheet.applyUserInfo }}</span>
                    <span v-if="sheet.content" class="truncate hidden sm:block text-base-content/50">
                      · {{ sheet.content.substring(0, 50) }}{{ sheet.content.length > 50 ? '...' : '' }}
                    </span>
                  </div>
                  <div class="flex items-center gap-2 mt-1 text-xs text-base-content/50">
                    <span>用户ID: {{ sheet.userId }}</span>
                    <span class="hidden sm:block">创建时间: {{ formatDate(sheet.createTime) }}</span>
                    <span v-if="sheet.updateTime !== sheet.createTime" class="hidden sm:block">
                      · 更新时间: {{ formatDate(sheet.updateTime) }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- 右侧：操作按钮 -->
              <div class="flex gap-1 flex-shrink-0">
                <!-- 大屏幕显示完整按钮 -->
                <div class="hidden lg:flex gap-1">
                  <button class="btn btn-xs btn-ghost" @click="viewSheet(sheet)" title="查看详情">
                    <EyeIcon class="w-3 h-3"/>
                    <span class="ml-1">查看</span>
                  </button>
                  <button class="btn btn-xs btn-info" @click="copyContent(sheet)" title="复制内容">
                    <ClipboardDocumentIcon class="w-3 h-3"/>
                    <span class="ml-1">复制</span>
                  </button>
                </div>

                <!-- 中等屏幕显示图标按钮 -->
                <div class="hidden md:flex lg:hidden gap-1">
                  <button class="btn btn-xs btn-ghost" @click="viewSheet(sheet)" title="查看详情">
                    <EyeIcon class="w-3 h-3"/>
                  </button>
                  <button class="btn btn-xs btn-info" @click="copyContent(sheet)" title="复制内容">
                    <ClipboardDocumentIcon class="w-3 h-3"/>
                  </button>
                </div>

                <!-- 小屏幕显示下拉菜单 -->
                <div class="dropdown dropdown-end md:hidden">
                  <div tabindex="0" role="button" class="btn btn-xs btn-ghost">
                    <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M12 5v.01M12 12v.01M12 19v.01"></path>
                    </svg>
                  </div>
                  <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-32">
                    <li><a @click="viewSheet(sheet)">查看详情</a></li>
                    <li><a @click="copyContent(sheet)">复制内容</a></li>
                  </ul>
                </div>
              </div>
            </li>
          </ul>
        </div>

        <!-- 分页 -->
        <div v-if="totalPages > 1" class="flex justify-between items-center mt-4 pt-4 border-t border-base-200">
          <div class="text-xs text-base-content/60">
            共 {{ total }} 个工单
          </div>
          <div class="flex items-center gap-2">
            <button
                class="btn btn-sm"
                :disabled="currentPage <= 1"
                @click="changePage(currentPage - 1)"
            >
              <ChevronLeftIcon class="w-4 h-4"/>
              <span class="ml-1">上一页</span>
            </button>
            <span class="text-sm text-base-content/70 px-3">
              第 {{ currentPage }} 页 / 共 {{ totalPages }} 页
            </span>
            <button
                class="btn btn-sm"
                :disabled="currentPage >= totalPages"
                @click="changePage(currentPage + 1)"
            >
              <span class="mr-1">下一页</span>
              <ChevronRightIcon class="w-4 h-4"/>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 工单详情模态框 -->
    <dialog id="sheet_detail_modal" class="modal">
      <div class="modal-box w-11/12 max-w-2xl">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>

        <h3 class="font-normal text-lg mb-4">工单详情</h3>

        <div v-if="selectedSheet" class="grid grid-cols-2 gap-4">
          <div>
            <label class="floating-label">
              <input type="text" class="input input-bordered w-full" :value="'#' + selectedSheet.id" readonly />
              <span>工单ID</span>
            </label>
          </div>
          <div>
            <label class="floating-label">
              <input type="text" class="input input-bordered w-full" :value="selectedSheet.userId" readonly />
              <span>用户ID</span>
            </label>
          </div>

          <div class="col-span-2">
            <label class="floating-label">
              <input type="text" class="input input-bordered w-full" :value="selectedSheet.title" readonly />
              <span>工单标题</span>
            </label>
          </div>

          <div class="col-span-2">
            <label class="floating-label">
              <input type="text" class="input input-bordered w-full" :value="selectedSheet.applyUserInfo" readonly />
              <span>申请用户信息</span>
            </label>
          </div>

          <div>
            <label class="floating-label">
              <input type="text" class="input input-bordered w-full" :value="getTypeText(selectedSheet.type)" readonly />
              <span>工单类型</span>
            </label>
          </div>
          <div>
            <label class="floating-label">
              <input type="text" class="input input-bordered w-full" :value="getStatusText(selectedSheet.status)" readonly />
              <span>工单状态</span>
            </label>
          </div>

          <div class="col-span-2">
            <label class="floating-label">
              <textarea class="textarea textarea-bordered w-full h-32" :value="selectedSheet.content" readonly></textarea>
              <span>工单内容</span>
            </label>
          </div>

          <div>
            <label class="floating-label">
              <input type="text" class="input input-bordered w-full" :value="formatDate(selectedSheet.createTime)" readonly />
              <span>创建时间</span>
            </label>
          </div>
          <div>
            <label class="floating-label">
              <input type="text" class="input input-bordered w-full" :value="formatDate(selectedSheet.updateTime)" readonly />
              <span>更新时间</span>
            </label>
          </div>
        </div>

        <div class="modal-action">
          <button class="btn btn-ghost" onclick="sheet_detail_modal.close()">关闭</button>
          <button v-if="selectedSheet" class="btn btn-primary" @click="copyContent(selectedSheet)">
            <ClipboardDocumentIcon class="w-4 h-4"/>
            复制内容
          </button>
        </div>
      </div>
    </dialog>
  </div>
</template>

<script setup lang="ts">
import {computed, onMounted, reactive, ref} from 'vue'
import {
  ArrowPathIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ClipboardDocumentIcon,
  DocumentTextIcon,
  EyeIcon
} from '@heroicons/vue/24/outline'
import {applySheetList} from '../utils/adminService'
import type {ApplySheet, PageData, Result} from '../utils/adminInterfaces'

// 响应式数据
const loading = ref(false)
const applySheets = ref<ApplySheet[]>([])
const selectedSheet = ref<ApplySheet | null>(null)

// 搜索参数
const searchParams = reactive({
  title: '',
  type: '',
  status: ''
})

// 分页参数
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 计算属性
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

// 加载工单列表
const loadApplySheets = async () => {
  try {
    loading.value = true
    const response: Result<PageData<ApplySheet>> = await applySheetList(currentPage.value, pageSize.value)

    if (response.code === 0) {
      applySheets.value = response.result.list
      total.value = response.result.total
    } else {
      console.error('加载工单列表失败:', response.message)
    }
  } catch (error) {
    console.error('加载工单列表出错:', error)
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  currentPage.value = 1
  loadApplySheets()
}

// 筛选处理
const handleFilter = () => {
  currentPage.value = 1
  loadApplySheets()
}

// 重置搜索
const resetSearch = () => {
  searchParams.title = ''
  searchParams.type = ''
  searchParams.status = ''
  currentPage.value = 1
  loadApplySheets()
}

// 分页
const changePage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    loadApplySheets()
  }
}

// 查看工单详情
const viewSheet = (sheet: ApplySheet) => {
  selectedSheet.value = sheet
  const modal = document.getElementById('sheet_detail_modal') as HTMLDialogElement
  modal?.showModal()
}

// 复制内容
const copyContent = async (sheet: ApplySheet) => {
  try {
    const content = `工单 #${sheet.id}: ${sheet.title}\n\n用户信息: ${sheet.applyUserInfo}\n类型: ${getTypeText(sheet.type)}\n状态: ${getStatusText(sheet.status)}\n\n内容:\n${sheet.content}`
    await navigator.clipboard.writeText(content)
    // 这里可以添加成功提示
    console.log('内容已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
  }
}

// 获取类型文本
const getTypeText = (type: number): string => {
  const typeMap: Record<number, string> = {
    1: 'Bug反馈',
    2: '功能建议',
    3: '技术支持',
    4: '账户问题',
    5: '其他'
  }
  return typeMap[type] || '未知类型'
}

// 获取类型样式
const getTypeClass = (type: number): string => {
  const typeClassMap: Record<number, string> = {
    1: 'badge-error',
    2: 'badge-info',
    3: 'badge-warning',
    4: 'badge-primary',
    5: 'badge-neutral'
  }
  return typeClassMap[type] || 'badge-neutral'
}

// 获取状态文本
const getStatusText = (status: number): string => {
  const statusMap: Record<number, string> = {
    1: '待处理',
    2: '处理中',
    3: '已解决',
    4: '已关闭'
  }
  return statusMap[status] || '未知状态'
}

// 获取状态样式
const getStatusClass = (status: number): string => {
  const statusClassMap: Record<number, string> = {
    1: 'badge-warning',
    2: 'badge-info',
    3: 'badge-success',
    4: 'badge-neutral'
  }
  return statusClassMap[status] || 'badge-neutral'
}

// 格式化日期
const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 生命周期
onMounted(() => {
  loadApplySheets()
})
</script>