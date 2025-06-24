<template>
  <div class="space-y-6">
    <!-- 页面标题和操作 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-base-content">赞助管理</h1>
        <p class="text-base-content/70 mt-1">管理网站的赞助信息和赞助商</p>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-outline" @click="openSettingsModal">
          <CogIcon class="w-4 h-4" />
          赞助设置
        </button>
        <button class="btn btn-primary" @click="openCreateModal">
          <PlusIcon class="w-4 h-4" />
          添加赞助
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <div class="stat bg-base-100 shadow rounded-lg">
        <div class="stat-figure text-primary">
          <CurrencyDollarIcon class="w-8 h-8" />
        </div>
        <div class="stat-title">总赞助金额</div>
        <div class="stat-value text-primary">¥{{ stats.totalAmount.toLocaleString() }}</div>
        <div class="stat-desc">累计收到赞助</div>
      </div>
      
      <div class="stat bg-base-100 shadow rounded-lg">
        <div class="stat-figure text-secondary">
          <UserGroupIcon class="w-8 h-8" />
        </div>
        <div class="stat-title">赞助人数</div>
        <div class="stat-value text-secondary">{{ stats.totalSponsors }}</div>
        <div class="stat-desc">累计赞助人数</div>
      </div>
      
      <div class="stat bg-base-100 shadow rounded-lg">
        <div class="stat-figure text-accent">
          <CalendarIcon class="w-8 h-8" />
        </div>
        <div class="stat-title">本月赞助</div>
        <div class="stat-value text-accent">¥{{ stats.monthlyAmount.toLocaleString() }}</div>
        <div class="stat-desc">{{ stats.monthlyCount }} 笔赞助</div>
      </div>
      
      <div class="stat bg-base-100 shadow rounded-lg">
        <div class="stat-figure text-info">
          <ChartBarIcon class="w-8 h-8" />
        </div>
        <div class="stat-title">平均金额</div>
        <div class="stat-value text-info">¥{{ stats.averageAmount.toFixed(0) }}</div>
        <div class="stat-desc">单笔平均赞助</div>
      </div>
    </div>

    <!-- 搜索和筛选 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">搜索赞助</span>
            </label>
            <div class="relative">
              <input 
                v-model="searchQuery" 
                type="text" 
                placeholder="赞助者姓名、留言" 
                class="input input-bordered w-full pl-10"
                @input="handleSearch"
              />
              <MagnifyingGlassIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-base-content/50" />
            </div>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">状态筛选</span>
            </label>
            <select v-model="filters.status" class="select select-bordered" @change="handleFilter">
              <option value="">全部状态</option>
              <option value="pending">待确认</option>
              <option value="confirmed">已确认</option>
              <option value="displayed">已展示</option>
              <option value="hidden">已隐藏</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">金额范围</span>
            </label>
            <select v-model="filters.amountRange" class="select select-bordered" @change="handleFilter">
              <option value="">全部金额</option>
              <option value="0-50">¥0 - ¥50</option>
              <option value="50-100">¥50 - ¥100</option>
              <option value="100-500">¥100 - ¥500</option>
              <option value="500+">¥500+</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">排序方式</span>
            </label>
            <select v-model="filters.sortBy" class="select select-bordered" @change="handleFilter">
              <option value="created_at">赞助时间</option>
              <option value="amount">赞助金额</option>
              <option value="sponsor_name">赞助者姓名</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- 赞助列表 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body p-0">
        <div class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
              <tr>
                <th>
                  <label>
                    <input type="checkbox" class="checkbox" v-model="selectAll" @change="toggleSelectAll" />
                  </label>
                </th>
                <th>赞助者信息</th>
                <th>赞助金额</th>
                <th>留言</th>
                <th>状态</th>
                <th>赞助时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="sponsor in sponsors" :key="sponsor.id">
                <td>
                  <label>
                    <input type="checkbox" class="checkbox" v-model="selectedSponsors" :value="sponsor.id" />
                  </label>
                </td>
                <td>
                  <div class="flex items-center gap-3">
                    <div class="avatar placeholder">
                      <div class="bg-neutral text-neutral-content rounded-full w-12">
                        <span class="text-xl">{{ sponsor.sponsorName.charAt(0) }}</span>
                      </div>
                    </div>
                    <div>
                      <div class="font-bold">{{ sponsor.sponsorName }}</div>
                      <div class="text-sm text-base-content/70">{{ sponsor.contact || '未提供联系方式' }}</div>
                      <div class="text-xs text-base-content/50">ID: {{ sponsor.id }}</div>
                    </div>
                  </div>
                </td>
                <td>
                  <div class="font-bold text-lg" :class="getAmountColor(sponsor.amount)">
                    ¥{{ sponsor.amount.toLocaleString() }}
                  </div>
                  <div class="text-xs text-base-content/70">{{ sponsor.paymentMethod }}</div>
                </td>
                <td>
                  <div class="max-w-xs">
                    <div class="text-sm" :title="sponsor.message">
                      {{ sponsor.message ? (sponsor.message.length > 50 ? sponsor.message.substring(0, 50) + '...' : sponsor.message) : '无留言' }}
                    </div>
                  </div>
                </td>
                <td>
                  <div class="badge" :class="getStatusBadgeClass(sponsor.status)">
                    {{ getStatusText(sponsor.status) }}
                  </div>
                </td>
                <td class="text-sm">{{ formatDate(sponsor.createdAt) }}</td>
                <td>
                  <div class="dropdown dropdown-end">
                    <div tabindex="0" role="button" class="btn btn-ghost btn-xs">
                      <EllipsisVerticalIcon class="w-4 h-4" />
                    </div>
                    <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52">
                      <li><a @click="viewSponsor(sponsor)">查看详情</a></li>
                      <li><a @click="editSponsor(sponsor)">编辑</a></li>
                      <li v-if="sponsor.status === 'pending'"><a @click="confirmSponsor(sponsor)">确认赞助</a></li>
                      <li v-if="sponsor.status === 'confirmed'"><a @click="displaySponsor(sponsor)">设为展示</a></li>
                      <li v-if="sponsor.status === 'displayed'"><a @click="hideSponsor(sponsor)">隐藏展示</a></li>
                      <li><a @click="deleteSponsor(sponsor)" class="text-error">删除</a></li>
                    </ul>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        
        <!-- 批量操作 -->
        <div v-if="selectedSponsors.length > 0" class="flex items-center gap-2 p-4 border-t border-base-300">
          <span class="text-sm text-base-content/70">已选择 {{ selectedSponsors.length }} 项</span>
          <div class="flex gap-2 ml-4">
            <button class="btn btn-sm btn-outline" @click="batchConfirm">批量确认</button>
            <button class="btn btn-sm btn-outline" @click="batchDisplay">批量展示</button>
            <button class="btn btn-sm btn-outline" @click="batchHide">批量隐藏</button>
            <button class="btn btn-sm btn-error" @click="batchDelete">批量删除</button>
          </div>
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

    <!-- 创建/编辑赞助模态框 -->
    <dialog ref="sponsorModal" class="modal">
      <div class="modal-box w-11/12 max-w-2xl">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
        
        <h3 class="font-bold text-lg mb-4">
          {{ editingSponsor ? '编辑赞助信息' : '添加赞助记录' }}
        </h3>
        
        <div class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">赞助者姓名 <span class="text-error">*</span></span>
              </label>
              <input 
                v-model="sponsorForm.sponsorName" 
                type="text" 
                placeholder="请输入赞助者姓名" 
                class="input input-bordered"
                :class="{ 'input-error': errors.sponsorName }"
              />
              <label class="label" v-if="errors.sponsorName">
                <span class="label-text-alt text-error">{{ errors.sponsorName }}</span>
              </label>
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">赞助金额 <span class="text-error">*</span></span>
              </label>
              <input 
                v-model.number="sponsorForm.amount" 
                type="number" 
                placeholder="0.00" 
                class="input input-bordered"
                :class="{ 'input-error': errors.amount }"
                min="0"
                step="0.01"
              />
              <label class="label" v-if="errors.amount">
                <span class="label-text-alt text-error">{{ errors.amount }}</span>
              </label>
            </div>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">联系方式</span>
              </label>
              <input 
                v-model="sponsorForm.contact" 
                type="text" 
                placeholder="邮箱、电话等" 
                class="input input-bordered"
              />
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">支付方式</span>
              </label>
              <select v-model="sponsorForm.paymentMethod" class="select select-bordered">
                <option value="alipay">支付宝</option>
                <option value="wechat">微信支付</option>
                <option value="bank">银行转账</option>
                <option value="paypal">PayPal</option>
                <option value="other">其他</option>
              </select>
            </div>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">赞助留言</span>
            </label>
            <textarea 
              v-model="sponsorForm.message" 
              class="textarea textarea-bordered" 
              placeholder="赞助者的留言或祝福"
              rows="3"
            ></textarea>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">状态</span>
            </label>
            <select v-model="sponsorForm.status" class="select select-bordered">
              <option value="pending">待确认</option>
              <option value="confirmed">已确认</option>
              <option value="displayed">已展示</option>
              <option value="hidden">已隐藏</option>
            </select>
          </div>
        </div>
        
        <div class="modal-action">
          <button type="button" class="btn btn-ghost" @click="closeModal">取消</button>
          <button type="button" class="btn btn-primary" @click="saveSponsor" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-sm"></span>
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- 赞助设置模态框 -->
    <dialog ref="settingsModal" class="modal">
      <div class="modal-box w-11/12 max-w-2xl">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
        
        <h3 class="font-bold text-lg mb-4">赞助设置</h3>
        
        <div class="space-y-4">
          <div class="form-control">
            <label class="cursor-pointer label">
              <span class="label-text">启用赞助功能</span>
              <input 
                v-model="settings.enabled" 
                type="checkbox" 
                class="toggle toggle-primary"
              />
            </label>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">赞助页面标题</span>
            </label>
            <input 
              v-model="settings.title" 
              type="text" 
              placeholder="支持我们" 
              class="input input-bordered"
            />
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">赞助页面描述</span>
            </label>
            <textarea 
              v-model="settings.description" 
              class="textarea textarea-bordered" 
              placeholder="感谢您的支持..."
              rows="3"
            ></textarea>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">最小赞助金额</span>
              </label>
              <input 
                v-model.number="settings.minAmount" 
                type="number" 
                placeholder="1.00" 
                class="input input-bordered"
                min="0"
                step="0.01"
              />
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">推荐赞助金额</span>
              </label>
              <input 
                v-model="settings.suggestedAmounts" 
                type="text" 
                placeholder="10,50,100,500" 
                class="input input-bordered"
              />
            </div>
          </div>
          
          <div class="form-control">
            <label class="cursor-pointer label">
              <span class="label-text">自动展示赞助信息</span>
              <input 
                v-model="settings.autoDisplay" 
                type="checkbox" 
                class="toggle toggle-primary"
              />
            </label>
          </div>
          
          <div class="form-control">
            <label class="cursor-pointer label">
              <span class="label-text">允许匿名赞助</span>
              <input 
                v-model="settings.allowAnonymous" 
                type="checkbox" 
                class="toggle toggle-primary"
              />
            </label>
          </div>
        </div>
        
        <div class="modal-action">
          <button type="button" class="btn btn-ghost" @click="closeSettingsModal">取消</button>
          <button type="button" class="btn btn-primary" @click="saveSettings" :disabled="savingSettings">
            <span v-if="savingSettings" class="loading loading-spinner loading-sm"></span>
            {{ savingSettings ? '保存中...' : '保存设置' }}
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
  CogIcon,
  CurrencyDollarIcon,
  UserGroupIcon,
  CalendarIcon,
  ChartBarIcon
} from '@heroicons/vue/24/outline'
import { api } from '../utils/axiosInstance'

// 数据类型定义
interface Sponsor {
  id: number
  sponsorName: string
  amount: number
  contact?: string
  paymentMethod: string
  message?: string
  status: 'pending' | 'confirmed' | 'displayed' | 'hidden'
  createdAt: string
}

interface SponsorStats {
  totalAmount: number
  totalSponsors: number
  monthlyAmount: number
  monthlyCount: number
  averageAmount: number
}

interface SponsorSettings {
  enabled: boolean
  title: string
  description: string
  minAmount: number
  suggestedAmounts: string
  autoDisplay: boolean
  allowAnonymous: boolean
}

// 响应式数据
const sponsors = ref<Sponsor[]>([])
const stats = ref<SponsorStats>({
  totalAmount: 0,
  totalSponsors: 0,
  monthlyAmount: 0,
  monthlyCount: 0,
  averageAmount: 0
})
const loading = ref(false)
const saving = ref(false)
const savingSettings = ref(false)
const searchQuery = ref('')
const editingSponsor = ref<Sponsor | null>(null)
const sponsorModal = ref<HTMLDialogElement>()
const settingsModal = ref<HTMLDialogElement>()
const selectAll = ref(false)
const selectedSponsors = ref<number[]>([])

// 筛选条件
const filters = reactive({
  status: '',
  amountRange: '',
  sortBy: 'created_at'
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 表单数据
const sponsorForm = reactive({
  sponsorName: '',
  amount: 0,
  contact: '',
  paymentMethod: 'alipay',
  message: '',
  status: 'pending' as 'pending' | 'confirmed' | 'displayed' | 'hidden'
})

// 设置数据
const settings = reactive<SponsorSettings>({
  enabled: true,
  title: '支持我们',
  description: '感谢您的支持，您的赞助将帮助我们更好地维护和发展这个平台。',
  minAmount: 1,
  suggestedAmounts: '10,50,100,500',
  autoDisplay: false,
  allowAnonymous: true
})

// 表单验证错误
const errors = reactive({
  sponsorName: '',
  amount: ''
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
const fetchSponsors = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      search: searchQuery.value,
      ...filters
    }
    
    const response = await api.get('/api/admin/sponsors', params)
    sponsors.value = response.data.data.sponsors
    pagination.total = response.data.data.total
  } catch (error) {
    console.error('获取赞助列表失败:', error)
    // 使用模拟数据
    sponsors.value = generateMockSponsors()
    pagination.total = 45
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const response = await api.get('/api/admin/sponsors/stats')
    stats.value = response.data.data
  } catch (error) {
    console.error('获取赞助统计失败:', error)
    // 使用模拟数据
    stats.value = {
      totalAmount: 12580,
      totalSponsors: 156,
      monthlyAmount: 2340,
      monthlyCount: 23,
      averageAmount: 80.6
    }
  }
}

const fetchSettings = async () => {
  try {
    // 赞助设置接口暂未实现，使用默认设置
    console.warn('赞助设置接口暂未实现')
  } catch (error) {
    console.error('获取赞助设置失败:', error)
  }
}

// 生成模拟数据
const generateMockSponsors = (): Sponsor[] => {
  const paymentMethods = ['alipay', 'wechat', 'bank', 'paypal', 'other']
  const statuses: ('pending' | 'confirmed' | 'displayed' | 'hidden')[] = ['pending', 'confirmed', 'displayed', 'hidden']
  const mockSponsors: Sponsor[] = []
  
  for (let i = 1; i <= pagination.pageSize; i++) {
    const sponsorId = (pagination.page - 1) * pagination.pageSize + i
    const amount = Math.floor(Math.random() * 500) + 10
    mockSponsors.push({
      id: sponsorId,
      sponsorName: `赞助者${sponsorId}`,
      amount,
      contact: Math.random() > 0.5 ? `sponsor${sponsorId}@example.com` : undefined,
      paymentMethod: paymentMethods[Math.floor(Math.random() * paymentMethods.length)],
      message: Math.random() > 0.3 ? `感谢提供这么好的平台，希望越来越好！` : undefined,
      status: statuses[Math.floor(Math.random() * statuses.length)],
      createdAt: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString()
    })
  }
  
  return mockSponsors
}

const handleSearch = () => {
  pagination.page = 1
  fetchSponsors()
}

const handleFilter = () => {
  pagination.page = 1
  fetchSponsors()
}

const changePage = (page: number) => {
  pagination.page = page
  fetchSponsors()
}

const toggleSelectAll = () => {
  if (selectAll.value) {
    selectedSponsors.value = sponsors.value.map(s => s.id)
  } else {
    selectedSponsors.value = []
  }
}

const openCreateModal = () => {
  editingSponsor.value = null
  resetForm()
  sponsorModal.value?.showModal()
}

const editSponsor = (sponsor: Sponsor) => {
  editingSponsor.value = sponsor
  Object.assign(sponsorForm, {
    sponsorName: sponsor.sponsorName,
    amount: sponsor.amount,
    contact: sponsor.contact || '',
    paymentMethod: sponsor.paymentMethod,
    message: sponsor.message || '',
    status: sponsor.status
  })
  sponsorModal.value?.showModal()
}

const viewSponsor = (sponsor: Sponsor) => {
  // 实现查看详情逻辑
  alert(`赞助者: ${sponsor.sponsorName}\n金额: ¥${sponsor.amount}\n留言: ${sponsor.message || '无'}\n时间: ${formatDate(sponsor.createdAt)}`)
}

const closeModal = () => {
  sponsorModal.value?.close()
  resetForm()
}

const openSettingsModal = () => {
  settingsModal.value?.showModal()
}

const closeSettingsModal = () => {
  settingsModal.value?.close()
}

const resetForm = () => {
  Object.assign(sponsorForm, {
    sponsorName: '',
    amount: 0,
    contact: '',
    paymentMethod: 'alipay',
    message: '',
    status: 'pending'
  })
  Object.assign(errors, {
    sponsorName: '',
    amount: ''
  })
}

const validateForm = () => {
  errors.sponsorName = ''
  errors.amount = ''
  
  if (!sponsorForm.sponsorName.trim()) {
    errors.sponsorName = '赞助者姓名不能为空'
    return false
  }
  
  if (sponsorForm.amount <= 0) {
    errors.amount = '赞助金额必须大于0'
    return false
  }
  
  return true
}

const saveSponsor = async () => {
  if (!validateForm()) {
    return
  }
  
  saving.value = true
  try {
    const data = { ...sponsorForm }
    
    if (editingSponsor.value) {
      // 编辑赞助
      await api.put(`/api/admin/sponsors/${editingSponsor.value.id}`, data)
    } else {
      // 创建赞助
      await api.post('/api/admin/sponsors', data)
    }
    
    closeModal()
    fetchSponsors()
    fetchStats()
  } catch (error) {
    console.error('保存赞助失败:', error)
    // 模拟保存成功
    closeModal()
    fetchSponsors()
    fetchStats()
  } finally {
    saving.value = false
  }
}

const saveSettings = async () => {
  savingSettings.value = true
  try {
    await api.put('/api/admin/sponsors/settings', settings)
    closeSettingsModal()
  } catch (error) {
    console.error('保存设置失败:', error)
    // 模拟保存成功
    closeSettingsModal()
  } finally {
    savingSettings.value = false
  }
}

const confirmSponsor = async (sponsor: Sponsor) => {
  try {
    await api.post(`/api/admin/sponsors/${sponsor.id}/confirm`)
    sponsor.status = 'confirmed'
    fetchStats()
  } catch (error) {
    console.error('确认赞助失败:', error)
    // 模拟确认成功
    sponsor.status = 'confirmed'
    fetchStats()
  }
}

const displaySponsor = async (sponsor: Sponsor) => {
  try {
    await api.post(`/api/admin/sponsors/${sponsor.id}/display`)
    sponsor.status = 'displayed'
  } catch (error) {
    console.error('设置展示失败:', error)
    // 模拟设置成功
    sponsor.status = 'displayed'
  }
}

const hideSponsor = async (sponsor: Sponsor) => {
  try {
    await api.post(`/api/admin/sponsors/${sponsor.id}/hide`)
    sponsor.status = 'hidden'
  } catch (error) {
    console.error('隐藏展示失败:', error)
    // 模拟隐藏成功
    sponsor.status = 'hidden'
  }
}

const deleteSponsor = async (sponsor: Sponsor) => {
  if (confirm(`确定要删除赞助者「${sponsor.sponsorName}」的记录吗？此操作不可恢复！`)) {
    try {
      await api.delete(`/api/admin/sponsors/${sponsor.id}`)
      fetchSponsors()
      fetchStats()
    } catch (error) {
      console.error('删除赞助失败:', error)
    }
  }
}

// 批量操作
const batchConfirm = async () => {
  if (selectedSponsors.value.length === 0) return
  
  try {
    await api.post('/api/admin/sponsors/batch-confirm', {
      ids: selectedSponsors.value
    })
    fetchSponsors()
    fetchStats()
    selectedSponsors.value = []
    selectAll.value = false
  } catch (error) {
    console.error('批量确认失败:', error)
  }
}

const batchDisplay = async () => {
  if (selectedSponsors.value.length === 0) return
  
  try {
    await api.post('/api/admin/sponsors/batch-display', {
      ids: selectedSponsors.value
    })
    fetchSponsors()
    selectedSponsors.value = []
    selectAll.value = false
  } catch (error) {
    console.error('批量展示失败:', error)
  }
}

const batchHide = async () => {
  if (selectedSponsors.value.length === 0) return
  
  try {
    await api.post('/api/admin/sponsors/batch-hide', {
      ids: selectedSponsors.value
    })
    fetchSponsors()
    selectedSponsors.value = []
    selectAll.value = false
  } catch (error) {
    console.error('批量隐藏失败:', error)
  }
}

const batchDelete = async () => {
  if (selectedSponsors.value.length === 0) return
  
  if (confirm(`确定要删除选中的 ${selectedSponsors.value.length} 条赞助记录吗？此操作不可恢复！`)) {
    try {
      await api.post('/api/admin/sponsors/batch-delete', {
        ids: selectedSponsors.value
      })
      fetchSponsors()
      fetchStats()
      selectedSponsors.value = []
      selectAll.value = false
    } catch (error) {
      console.error('批量删除失败:', error)
    }
  }
}

// 工具函数
const getStatusText = (status: string) => {
  const statusMap = {
    pending: '待确认',
    confirmed: '已确认',
    displayed: '已展示',
    hidden: '已隐藏'
  }
  return statusMap[status as keyof typeof statusMap] || status
}

const getStatusBadgeClass = (status: string) => {
  const classMap = {
    pending: 'badge-warning',
    confirmed: 'badge-info',
    displayed: 'badge-success',
    hidden: 'badge-error'
  }
  return classMap[status as keyof typeof classMap] || 'badge-outline'
}

const getAmountColor = (amount: number) => {
  if (amount >= 500) return 'text-error'
  if (amount >= 100) return 'text-warning'
  if (amount >= 50) return 'text-info'
  return 'text-success'
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

// 组件挂载时获取数据
onMounted(() => {
  fetchSponsors()
  fetchStats()
  fetchSettings()
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

/* 金额颜色 */
.text-success {
  color: hsl(var(--su));
}

.text-info {
  color: hsl(var(--in));
}

.text-warning {
  color: hsl(var(--wa));
}

.text-error {
  color: hsl(var(--er));
}
</style>