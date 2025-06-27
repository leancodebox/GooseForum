<template>
  <div class="space-y-6">
    <!-- 页面标题和操作 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-base-content">工单管理</h1>
        <p class="text-base-content/70 mt-1">处理用户反馈和技术支持请求</p>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-outline" @click="openTemplateModal">
          <DocumentTextIcon class="w-4 h-4" />
          回复模板
        </button>
        <button class="btn btn-primary" @click="openCreateModal">
          <PlusIcon class="w-4 h-4" />
          创建工单
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <div class="stat bg-base-100 shadow rounded-lg">
        <div class="stat-figure text-warning">
          <ClockIcon class="w-8 h-8" />
        </div>
        <div class="stat-title">待处理</div>
        <div class="stat-value text-warning">{{ stats.pending }}</div>
        <div class="stat-desc">需要及时处理</div>
      </div>
      
      <div class="stat bg-base-100 shadow rounded-lg">
        <div class="stat-figure text-info">
          <ChatBubbleLeftRightIcon class="w-8 h-8" />
        </div>
        <div class="stat-title">处理中</div>
        <div class="stat-value text-info">{{ stats.inProgress }}</div>
        <div class="stat-desc">正在跟进处理</div>
      </div>
      
      <div class="stat bg-base-100 shadow rounded-lg">
        <div class="stat-figure text-success">
          <CheckCircleIcon class="w-8 h-8" />
        </div>
        <div class="stat-title">已解决</div>
        <div class="stat-value text-success">{{ stats.resolved }}</div>
        <div class="stat-desc">本月已解决</div>
      </div>
      
      <div class="stat bg-base-100 shadow rounded-lg">
        <div class="stat-figure text-primary">
          <ChartBarIcon class="w-8 h-8" />
        </div>
        <div class="stat-title">解决率</div>
        <div class="stat-value text-primary">{{ stats.resolveRate }}%</div>
        <div class="stat-desc">本月解决率</div>
      </div>
    </div>

    <!-- 搜索和筛选 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3 sm:gap-4">
          <div class="form-control relative">
            <label class="floating-label">
              <span>搜索工单</span>
              <input 
                v-model="searchQuery" 
                type="text" 
                class="input input-bordered w-full"
                placeholder="标题、用户名"
                @input="handleSearch"
              />
            </label>
          </div>
          
          <div class="form-control">
            <label class="floating-label">
              <span>状态筛选</span>
              <select v-model="filters.status" class="select select-bordered w-full" @change="handleFilter">
                <option value="">全部状态</option>
                <option value="open">待处理</option>
                <option value="in_progress">处理中</option>
                <option value="resolved">已解决</option>
                <option value="closed">已关闭</option>
              </select>
            </label>
          </div>
          
          <div class="form-control">
            <label class="floating-label">
              <span>优先级</span>
              <select v-model="filters.priority" class="select select-bordered w-full" @change="handleFilter">
                <option value="">全部优先级</option>
                <option value="low">低</option>
                <option value="normal">普通</option>
                <option value="high">高</option>
                <option value="urgent">紧急</option>
              </select>
            </label>
          </div>
          
          <div class="form-control">
            <label class="floating-label">
              <span>分类筛选</span>
              <select v-model="filters.category" class="select select-bordered w-full" @change="handleFilter">
                <option value="">全部分类</option>
                <option value="bug">Bug反馈</option>
                <option value="feature">功能建议</option>
                <option value="support">技术支持</option>
                <option value="account">账户问题</option>
                <option value="other">其他</option>
              </select>
            </label>
          </div>
          
          <div class="form-control">
            <label class="floating-label">
              <span>排序方式</span>
              <select v-model="filters.sortBy" class="select select-bordered w-full" @change="handleFilter">
                <option value="created_at">创建时间</option>
                <option value="updated_at">更新时间</option>
                <option value="priority">优先级</option>
                <option value="status">状态</option>
              </select>
            </label>
          </div>
        </div>
      </div>
    </div>

    <!-- 工单列表 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body p-0">
        <div class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
              <tr>
                <th class="w-1/3 min-w-[250px]">工单信息</th>
                <th class="w-40 min-w-[150px] hidden md:table-cell">用户</th>
                <th class="w-24 min-w-[80px] hidden lg:table-cell">分类</th>
                <th class="w-20 min-w-[80px] hidden sm:table-cell">优先级</th>
                <th class="w-20 min-w-[80px]">状态</th>
                <th class="w-32 min-w-[120px] hidden lg:table-cell">处理人</th>
                <th class="w-32 min-w-[120px] hidden sm:table-cell">创建时间</th>
                <th class="w-20 min-w-[80px]">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="ticket in tickets" :key="ticket.id" class="hover">
                <td>
                  <div class="space-y-1 min-w-0">
                    <div class="font-bold text-base cursor-pointer hover:text-primary break-words" @click="viewTicket(ticket)">
                      #{{ ticket.id }} {{ ticket.title }}
                    </div>
                    <div class="text-sm text-base-content/70 line-clamp-2 break-words">
                      <span class="block sm:hidden">{{ ticket.description.substring(0, 60) }}{{ ticket.description.length > 60 ? '...' : '' }}</span>
                      <span class="hidden sm:block">{{ ticket.description.substring(0, 120) }}{{ ticket.description.length > 120 ? '...' : '' }}</span>
                    </div>
                    <div class="flex flex-wrap items-center gap-1 sm:gap-2 text-xs text-base-content/50">
                      <span>{{ ticket.replies }} 回复</span>
                      <span class="hidden sm:inline">•</span>
                      <span class="hidden sm:inline">最后更新: {{ formatRelativeTime(ticket.updatedAt) }}</span>
                      <!-- 在小屏幕显示用户信息 -->
                      <span class="md:hidden">•</span>
                      <span class="md:hidden">{{ ticket.user.name }}</span>
                      <!-- 在小屏幕显示优先级 -->
                      <span class="sm:hidden">•</span>
                      <div class="sm:hidden badge badge-xs" :class="getPriorityBadgeClass(ticket.priority)">
                        {{ getPriorityText(ticket.priority) }}
                      </div>
                    </div>
                  </div>
                </td>
                <td class="hidden md:table-cell">
                  <div class="flex items-center gap-3 min-w-0">
                    <div class="avatar placeholder flex-shrink-0">
                      <div class="bg-neutral text-neutral-content rounded-full w-10">
                        <span class="text-sm">{{ ticket.user.name.charAt(0) }}</span>
                      </div>
                    </div>
                    <div class="min-w-0">
                      <div class="font-medium truncate">{{ ticket.user.name }}</div>
                      <div class="text-sm text-base-content/70 truncate">{{ ticket.user.email }}</div>
                    </div>
                  </div>
                </td>
                <td class="hidden lg:table-cell">
                  <div class="badge badge-outline badge-sm whitespace-nowrap">{{ getCategoryText(ticket.category) }}</div>
                </td>
                <td class="hidden sm:table-cell">
                  <div class="badge badge-sm whitespace-nowrap" :class="getPriorityBadgeClass(ticket.priority)">
                    {{ getPriorityText(ticket.priority) }}
                  </div>
                </td>
                <td>
                  <div class="badge badge-sm whitespace-nowrap" :class="getStatusBadgeClass(ticket.status)">
                    {{ getStatusText(ticket.status) }}
                  </div>
                </td>
                <td class="hidden lg:table-cell">
                  <div v-if="ticket.assignee" class="flex items-center gap-2 min-w-0">
                    <div class="avatar placeholder flex-shrink-0">
                      <div class="bg-primary text-primary-content rounded-full w-8">
                        <span class="text-xs">{{ ticket.assignee.name.charAt(0) }}</span>
                      </div>
                    </div>
                    <span class="text-sm truncate">{{ ticket.assignee.name }}</span>
                  </div>
                  <span v-else class="text-sm text-base-content/50">未分配</span>
                </td>
                <td class="text-sm whitespace-nowrap hidden sm:table-cell">{{ formatDate(ticket.createdAt) }}</td>
                <td>
                  <div class="dropdown dropdown-end">
                    <div tabindex="0" role="button" class="btn btn-ghost btn-xs">
                      <EllipsisVerticalIcon class="w-4 h-4" />
                    </div>
                    <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52">
                      <li><a @click="viewTicket(ticket)">查看详情</a></li>
                      <li><a @click="viewTicket(ticket)">回复</a></li>
                      <li><a>分配处理人</a></li>
                      <li><a @click="changeStatus(ticket, 'in_progress')" v-if="ticket.status === 'open'">标记为处理中</a></li>
                      <li><a @click="changeStatus(ticket, 'resolved')" v-if="ticket.status === 'in_progress'">标记为已解决</a></li>
                      <li><a @click="changeStatus(ticket, 'closed')" v-if="ticket.status === 'resolved'">关闭工单</a></li>
                      <li><a>修改优先级</a></li>
                      <li><a @click="deleteTicket(ticket)" class="text-error">删除</a></li>
                    </ul>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        
        <!-- 分页 -->
        <div class="flex justify-between items-center p-4 border-t border-base-300">
          <div class="text-sm text-base-content/70">
            显示 {{ (pagination.page - 1) * pagination.pageSize + 1 }} - 
            {{ Math.min(pagination.page * pagination.pageSize, pagination.total) }} 
            共 {{ pagination.total }} 条
          </div>
          <div class="join">
            <button 
              class="join-item btn btn-sm" 
              :disabled="pagination.page <= 1"
              @click="changePage(pagination.page - 1)"
            >
              上一页
            </button>
            <button 
              v-for="page in visiblePages" 
              :key="page"
              class="join-item btn btn-sm"
              :class="{ 'btn-active': page === pagination.page }"
              @click="changePage(page)"
            >
              {{ page }}
            </button>
            <button 
              class="join-item btn btn-sm" 
              :disabled="pagination.page >= totalPages"
              @click="changePage(pagination.page + 1)"
            >
              下一页
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建工单模态框 -->
    <dialog ref="ticketModal" class="modal">
      <div class="modal-box w-full max-w-3xl">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
        
        <h3 class="font-bold text-lg mb-4">创建工单</h3>
        
        <div class="space-y-6">
          <div class="form-control">
            <label class="floating-label" :class="{ 'input-error': errors.title }">
              <span>工单标题 *</span>
              <input 
                v-model="ticketForm.title" 
                type="text" 
                class="input input-bordered w-full"
                placeholder="请输入工单标题"
              />
            </label>
            <div class="label" v-if="errors.title">
              <span class="label-text-alt text-error">{{ errors.title }}</span>
            </div>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div class="form-control">
              <label class="floating-label">
                <span>分类</span>
                <select v-model="ticketForm.category" class="select select-bordered w-full">
                  <option value="bug">Bug反馈</option>
                  <option value="feature">功能建议</option>
                  <option value="support">技术支持</option>
                  <option value="account">账户问题</option>
                  <option value="other">其他</option>
                </select>
              </label>
            </div>
            
            <div class="form-control">
              <label class="floating-label">
                <span>优先级</span>
                <select v-model="ticketForm.priority" class="select select-bordered w-full">
                  <option value="low">低</option>
                  <option value="normal">普通</option>
                  <option value="high">高</option>
                  <option value="urgent">紧急</option>
                </select>
              </label>
            </div>
            
            <div class="form-control">
              <label class="floating-label">
                <span>处理人</span>
                <select v-model="ticketForm.assigneeId" class="select select-bordered w-full">
                  <option value="">未分配</option>
                  <option v-for="admin in admins" :key="admin.id" :value="admin.id">
                    {{ admin.name }}
                  </option>
                </select>
              </label>
            </div>
          </div>
          
          <div class="form-control">
            <label class="floating-label">
              <span>用户邮箱</span>
              <input 
                v-model="ticketForm.userEmail" 
                type="email" 
                class="input input-bordered w-full"
                placeholder="user@example.com"
              />
            </label>
          </div>
          
          <div class="form-control">
            <label class="floating-label" :class="{ 'textarea-error': errors.description }">
              <span>问题描述 <span class="text-error">*</span></span>
              <textarea 
                v-model="ticketForm.description" 
                class="textarea textarea-bordered w-full" 
                placeholder="请详细描述问题..."
                rows="6"
              ></textarea>
            </label>
            <div class="label" v-if="errors.description">
              <span class="label-text-alt text-error">{{ errors.description }}</span>
            </div>
          </div>
        </div>
        
        <div class="modal-action">
          <button type="button" class="btn btn-ghost" @click="closeModal">取消</button>
          <button type="button" class="btn btn-primary" @click="saveTicket" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-sm"></span>
            {{ saving ? '创建中...' : '创建工单' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- 工单详情模态框 -->
    <dialog ref="detailModal" class="modal">
      <div class="modal-box w-11/12 max-w-4xl max-h-[90vh]">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
        
        <div v-if="selectedTicket" class="space-y-6">
          <!-- 工单头部信息 -->
          <div class="border-b border-base-300 pb-4">
            <div class="flex items-start justify-between">
              <div>
                <h3 class="font-bold text-xl">#{{ selectedTicket.id }} {{ selectedTicket.title }}</h3>
                <div class="flex items-center gap-4 mt-2 text-sm text-base-content/70">
                  <span>创建时间: {{ formatDate(selectedTicket.createdAt) }}</span>
                  <span>最后更新: {{ formatDate(selectedTicket.updatedAt) }}</span>
                </div>
              </div>
              <div class="flex gap-2">
                <div class="badge" :class="getStatusBadgeClass(selectedTicket.status)">
                  {{ getStatusText(selectedTicket.status) }}
                </div>
                <div class="badge" :class="getPriorityBadgeClass(selectedTicket.priority)">
                  {{ getPriorityText(selectedTicket.priority) }}
                </div>
              </div>
            </div>
          </div>
          
          <!-- 工单内容 -->
          <div class="space-y-4">
            <div class="card bg-base-200">
              <div class="card-body">
                <div class="flex items-center gap-3 mb-3">
                  <div class="avatar placeholder">
                    <div class="bg-neutral text-neutral-content rounded-full w-10">
                      <span class="text-sm">{{ selectedTicket.user.name.charAt(0) }}</span>
                    </div>
                  </div>
                  <div>
                    <div class="font-medium">{{ selectedTicket.user.name }}</div>
                    <div class="text-sm text-base-content/70">{{ selectedTicket.user.email }}</div>
                  </div>
                  <div class="ml-auto text-sm text-base-content/70">
                    {{ formatDate(selectedTicket.createdAt) }}
                  </div>
                </div>
                <div class="prose max-w-none">
                  <p>{{ selectedTicket.description }}</p>
                </div>
              </div>
            </div>
            
            <!-- 回复列表 -->
            <div v-for="reply in selectedTicket.replies" :key="reply.id" class="card bg-base-100">
              <div class="card-body">
                <div class="flex items-center gap-3 mb-3">
                  <div class="avatar placeholder">
                    <div class="bg-primary text-primary-content rounded-full w-10">
                      <span class="text-sm">{{ reply.author.name.charAt(0) }}</span>
                    </div>
                  </div>
                  <div>
                    <div class="font-medium">{{ reply.author.name }}</div>
                    <div class="text-sm text-base-content/70">{{ reply.author.role === 'admin' ? '管理员' : '用户' }}</div>
                  </div>
                  <div class="ml-auto text-sm text-base-content/70">
                    {{ formatDate(reply.createdAt) }}
                  </div>
                </div>
                <div class="prose max-w-none">
                  <p>{{ reply.content }}</p>
                </div>
              </div>
            </div>
          </div>
          
          <!-- 回复表单 -->
          <div class="border-t border-base-300 pt-4">
            <div class="form-control">
              <label class="floating-label">
                <span>添加回复</span>
                <textarea 
                  v-model="replyContent" 
                  class="textarea textarea-bordered" 
                  placeholder="输入回复内容..."
                  rows="4"
                ></textarea>
              </label>
            </div>
            <div class="flex justify-between items-center mt-4">
              <div class="flex gap-2">
                <select v-model="replyStatus" class="select select-bordered select-sm">
                  <option value="">保持当前状态</option>
                  <option value="in_progress">标记为处理中</option>
                  <option value="resolved">标记为已解决</option>
                  <option value="closed">关闭工单</option>
                </select>
              </div>
              <button class="btn btn-primary" @click="submitReply" :disabled="!replyContent.trim() || replying">
                <span v-if="replying" class="loading loading-spinner loading-sm"></span>
                {{ replying ? '发送中...' : '发送回复' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </dialog>

    <!-- 回复模板模态框 -->
    <dialog ref="templateModal" class="modal">
      <div class="modal-box w-11/12 max-w-2xl">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
        
        <h3 class="font-bold text-lg mb-4">回复模板管理</h3>
        
        <div class="space-y-4">
          <div v-for="template in replyTemplates" :key="template.id" class="card bg-base-200">
            <div class="card-body">
              <div class="flex justify-between items-start">
                <div class="flex-1">
                  <h4 class="font-medium">{{ template.title }}</h4>
                  <p class="text-sm text-base-content/70 mt-1">{{ template.content.substring(0, 100) }}...</p>
                </div>
                <div class="flex gap-2">
                  <button class="btn btn-ghost btn-xs" @click="useTemplate(template)">
                    使用
                  </button>
                  <button class="btn btn-ghost btn-xs">
                    编辑
                  </button>
                  <button class="btn btn-ghost btn-xs text-error">
                    删除
                  </button>
                </div>
              </div>
            </div>
          </div>
          
          <button class="btn btn-outline w-full">
            <PlusIcon class="w-4 h-4" />
            添加新模板
          </button>
        </div>
      </div>
    </dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import {
  PlusIcon,
  MagnifyingGlassIcon,
  EllipsisVerticalIcon,
  DocumentTextIcon,
  ClockIcon,
  ChatBubbleLeftRightIcon,
  CheckCircleIcon,
  ChartBarIcon
} from '@heroicons/vue/24/outline'
import { api } from '../utils/axiosInstance'

// 数据类型定义
interface User {
  id: number
  name: string
  email: string
  role?: string
}

interface TicketReply {
  id: number
  content: string
  author: User
  createdAt: string
}

interface Ticket {
  id: number
  title: string
  description: string
  category: string
  priority: 'low' | 'normal' | 'high' | 'urgent'
  status: 'open' | 'in_progress' | 'resolved' | 'closed'
  user: User
  assignee?: User
  replies: TicketReply[]
  createdAt: string
  updatedAt: string
}

interface TicketStats {
  pending: number
  inProgress: number
  resolved: number
  resolveRate: number
}

interface ReplyTemplate {
  id: number
  title: string
  content: string
}

// 响应式数据
const tickets = ref<Ticket[]>([])
const stats = ref<TicketStats>({
  pending: 0,
  inProgress: 0,
  resolved: 0,
  resolveRate: 0
})
const admins = ref<User[]>([])
const replyTemplates = ref<ReplyTemplate[]>([])
const loading = ref(false)
const saving = ref(false)
const replying = ref(false)
const searchQuery = ref('')
const selectedTicket = ref<Ticket | null>(null)
const ticketModal = ref<HTMLDialogElement>()
const detailModal = ref<HTMLDialogElement>()
const templateModal = ref<HTMLDialogElement>()
const replyContent = ref('')
const replyStatus = ref('')

// 筛选条件
const filters = reactive({
  status: '',
  priority: '',
  category: '',
  sortBy: 'created_at'
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 表单数据
const ticketForm = reactive({
  title: '',
  description: '',
  category: 'support',
  priority: 'normal' as 'low' | 'normal' | 'high' | 'urgent',
  userEmail: '',
  assigneeId: ''
})

// 表单验证错误
const errors = reactive({
  title: '',
  description: ''
})

// 计算属性
const totalPages = computed(() => {
  return Math.ceil(pagination.total / pagination.pageSize)
})

const visiblePages = computed(() => {
  const current = pagination.page
  const total = totalPages.value
  const pages = []
  
  let start = Math.max(1, current - 2)
  let end = Math.min(total, current + 2)
  
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  
  return pages
})

// 方法
const fetchTickets = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      search: searchQuery.value,
      ...filters
    }
    
    const response = await api.get('/api/admin/tickets', params)
    tickets.value = response.data.data.tickets
    pagination.total = response.data.data.total
  } catch (error) {
    console.error('获取工单列表失败:', error)
    // 使用模拟数据
    tickets.value = generateMockTickets()
    pagination.total = 35
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const response = await api.get('/api/admin/tickets/stats')
    stats.value = response.data.data
  } catch (error) {
    console.error('获取工单统计失败:', error)
    // 使用模拟数据
    stats.value = {
      pending: 12,
      inProgress: 8,
      resolved: 45,
      resolveRate: 85
    }
  }
}

const fetchAdmins = async () => {
  try {
    // 管理员列表接口暂未实现，使用模拟数据
    console.warn('管理员列表接口暂未实现')
    // 使用模拟数据
    admins.value = [
      { id: 1, name: '管理员1', email: 'admin1@example.com' },
      { id: 2, name: '管理员2', email: 'admin2@example.com' }
    ]
  } catch (error) {
    console.error('获取管理员列表失败:', error)
  }
}

const fetchReplyTemplates = async () => {
  try {
    // 回复模板接口暂未实现，使用模拟数据
    console.warn('回复模板接口暂未实现')
    // 使用模拟数据
    replyTemplates.value = [
      {
        id: 1,
        title: '问题已收到',
        content: '感谢您的反馈，我们已经收到您的问题，将尽快为您处理。'
      },
      {
        id: 2,
        title: '问题已解决',
        content: '您的问题已经解决，如果还有其他疑问，请随时联系我们。'
      }
    ]
  } catch (error) {
    console.error('获取回复模板失败:', error)
  }
}

// 生成模拟数据
const generateMockTickets = (): Ticket[] => {
  const categories = ['bug', 'feature', 'support', 'account', 'other']
  const priorities: ('low' | 'normal' | 'high' | 'urgent')[] = ['low', 'normal', 'high', 'urgent']
  const statuses: ('open' | 'in_progress' | 'resolved' | 'closed')[] = ['open', 'in_progress', 'resolved', 'closed']
  const mockTickets: Ticket[] = []
  
  for (let i = 1; i <= pagination.pageSize; i++) {
    const ticketId = (pagination.page - 1) * pagination.pageSize + i
    const status = statuses[Math.floor(Math.random() * statuses.length)]
    mockTickets.push({
      id: ticketId,
      title: `工单标题 ${ticketId}`,
      description: `这是工单 ${ticketId} 的详细描述，用户反馈了一个问题需要处理。`,
      category: categories[Math.floor(Math.random() * categories.length)],
      priority: priorities[Math.floor(Math.random() * priorities.length)],
      status,
      user: {
        id: ticketId,
        name: `用户${ticketId}`,
        email: `user${ticketId}@example.com`
      },
      assignee: Math.random() > 0.5 ? {
        id: 1,
        name: '管理员1',
        email: 'admin1@example.com'
      } : undefined,
      replies: [],
      createdAt: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString(),
      updatedAt: new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString()
    })
  }
  
  return mockTickets
}

const handleSearch = () => {
  pagination.page = 1
  fetchTickets()
}

const handleFilter = () => {
  pagination.page = 1
  fetchTickets()
}

const changePage = (page: number) => {
  pagination.page = page
  fetchTickets()
}

const openCreateModal = () => {
  resetForm()
  ticketModal.value?.showModal()
}

const closeModal = () => {
  ticketModal.value?.close()
  resetForm()
}

const openTemplateModal = () => {
  templateModal.value?.showModal()
}

const resetForm = () => {
  Object.assign(ticketForm, {
    title: '',
    description: '',
    category: 'support',
    priority: 'normal',
    userEmail: '',
    assigneeId: ''
  })
  Object.assign(errors, {
    title: '',
    description: ''
  })
}

const validateForm = () => {
  errors.title = ''
  errors.description = ''
  
  if (!ticketForm.title.trim()) {
    errors.title = '工单标题不能为空'
    return false
  }
  
  if (!ticketForm.description.trim()) {
    errors.description = '问题描述不能为空'
    return false
  }
  
  return true
}

const saveTicket = async () => {
  if (!validateForm()) {
    return
  }
  
  saving.value = true
  try {
    const data = { ...ticketForm }
    await api.post('/api/admin/tickets', data)
    closeModal()
    fetchTickets()
    fetchStats()
  } catch (error) {
    console.error('创建工单失败:', error)
    // 模拟创建成功
    closeModal()
    fetchTickets()
    fetchStats()
  } finally {
    saving.value = false
  }
}

const viewTicket = async (ticket: Ticket) => {
  try {
    const response = await api.get(`/api/admin/tickets/${ticket.id}`)
    selectedTicket.value = response.data.data
  } catch (error) {
    console.error('获取工单详情失败:', error)
    // 使用模拟数据
    selectedTicket.value = {
      ...ticket,
      replies: [
        {
          id: 1,
          content: '我们已经收到您的反馈，正在处理中。',
          author: { id: 1, name: '管理员1', email: 'admin1@example.com', role: 'admin' },
          createdAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString()
        }
      ]
    }
  }
  detailModal.value?.showModal()
}

const submitReply = async () => {
  if (!replyContent.value.trim() || !selectedTicket.value) return
  
  replying.value = true
  try {
    const data = {
      content: replyContent.value,
      status: replyStatus.value || undefined
    }
    
    await api.post(`/api/admin/tickets/${selectedTicket.value.id}/replies`, data)
    
    // 更新本地数据
    if (replyStatus.value) {
      selectedTicket.value.status = replyStatus.value as any
    }
    
    replyContent.value = ''
    replyStatus.value = ''
    
    // 重新获取工单详情
    viewTicket(selectedTicket.value)
    fetchTickets()
    fetchStats()
  } catch (error) {
    console.error('回复工单失败:', error)
    // 模拟回复成功
    replyContent.value = ''
    replyStatus.value = ''
  } finally {
    replying.value = false
  }
}

const changeStatus = async (ticket: Ticket, status: string) => {
  try {
    await api.put(`/api/admin/tickets/${ticket.id}/status`, { status })
    ticket.status = status as any
    fetchStats()
  } catch (error) {
    console.error('修改状态失败:', error)
    // 模拟修改成功
    ticket.status = status as any
    fetchStats()
  }
}

const deleteTicket = async (ticket: Ticket) => {
  if (confirm(`确定要删除工单「${ticket.title}」吗？此操作不可恢复！`)) {
    try {
      await api.delete(`/api/admin/tickets/${ticket.id}`)
      fetchTickets()
      fetchStats()
    } catch (error) {
      console.error('删除工单失败:', error)
    }
  }
}

const useTemplate = (template: ReplyTemplate) => {
  replyContent.value = template.content
  templateModal.value?.close()
}

// 工具函数
const getCategoryText = (category: string) => {
  const categoryMap = {
    bug: 'Bug反馈',
    feature: '功能建议',
    support: '技术支持',
    account: '账户问题',
    other: '其他'
  }
  return categoryMap[category as keyof typeof categoryMap] || category
}

const getPriorityText = (priority: string) => {
  const priorityMap = {
    low: '低',
    normal: '普通',
    high: '高',
    urgent: '紧急'
  }
  return priorityMap[priority as keyof typeof priorityMap] || priority
}

const getStatusText = (status: string) => {
  const statusMap = {
    open: '待处理',
    in_progress: '处理中',
    resolved: '已解决',
    closed: '已关闭'
  }
  return statusMap[status as keyof typeof statusMap] || status
}

const getPriorityBadgeClass = (priority: string) => {
  const classMap = {
    low: 'badge-info',
    normal: 'badge-outline',
    high: 'badge-warning',
    urgent: 'badge-error'
  }
  return classMap[priority as keyof typeof classMap] || 'badge-outline'
}

const getStatusBadgeClass = (status: string) => {
  const classMap = {
    open: 'badge-warning',
    in_progress: 'badge-info',
    resolved: 'badge-success',
    closed: 'badge-outline'
  }
  return classMap[status as keyof typeof classMap] || 'badge-outline'
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatRelativeTime = (dateString: string) => {
  const now = new Date()
  const date = new Date(dateString)
  const diff = now.getTime() - date.getTime()
  
  const minutes = Math.floor(diff / (1000 * 60))
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  
  if (minutes < 60) {
    return `${minutes}分钟前`
  } else if (hours < 24) {
    return `${hours}小时前`
  } else {
    return `${days}天前`
  }
}

// 组件挂载时获取数据
onMounted(() => {
  fetchTickets()
  fetchStats()
  fetchAdmins()
  fetchReplyTemplates()
})
</script>

<style scoped>
.table th {
  background-color: hsl(var(--b2));
  font-weight: 600;
}

/* 统计卡片样式 */
.stat {
  padding: 1.5rem;
}

.stat-figure {
  opacity: 0.8;
}

/* 文本截断 */
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* 工单行悬停效果 */
.table tbody tr:hover {
  background-color: hsl(var(--b2));
}

/* 回复区域样式 */
.prose {
  color: hsl(var(--bc));
}

.prose p {
  margin: 0;
  white-space: pre-wrap;
}
</style>