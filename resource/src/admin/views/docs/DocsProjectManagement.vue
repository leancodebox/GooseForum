<template>
  <div class="p-4">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center mb-2">
      <h2 class="text-xl font-bold text-base-content">文档项目管理</h2>
      <button class="btn btn-primary btn-sm" @click="openCreateModal">
        <PlusIcon class="w-4 h-4"/>
        新建项目
      </button>
    </div>

    <!-- 搜索筛选 -->
    <div class="bg-base-200 p-3 rounded-lg mb-4">
      <div class="grid grid-cols-1 md:grid-cols-4 lg:grid-cols-6 gap-3 items-end">
        <div class="md:col-span-2">
          <label class="floating-label join w-full">
            <span>搜索项目名称或描述</span>
            <input
                v-model="searchParams.keyword"
                type="text"
                placeholder="搜索项目名称或描述"
                class="input input-bordered input-sm w-full"
                @keyup.enter="loadProjects"
            />
          </label>
        </div>

        <label class="floating-label join w-full">
          <span>状态</span>
          <select v-model="searchParams.status" class="select select-bordered select-sm">
            <option value="">全部状态</option>
            <option v-for="option in PROJECT_STATUS_OPTIONS" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </label>

        <label class="floating-label join w-full">
          <span>公开性</span>
          <select v-model="searchParams.isPublic" class="select select-bordered select-sm">
            <option value="">全部公开性</option>
            <option v-for="option in PUBLIC_STATUS_OPTIONS" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </label>

        <button class="btn btn-neutral btn-sm" @click="loadProjects">
          <MagnifyingGlassIcon class="w-4 h-4"/>
          搜索
        </button>

        <button class="btn btn-ghost btn-sm" @click="resetSearch">
          <ArrowPathIcon class="w-4 h-4"/>
          重置
        </button>
      </div>
    </div>

    <!-- 项目列表 -->
    <div>
      <div v-if="loading" class="flex justify-center items-center py-8">
        <span class="loading loading-spinner loading-lg"></span>
        <span class="ml-2 text-base-content">加载中...</span>
      </div>

      <div v-else-if="projects.length === 0" class="text-center py-8">
        <FolderOpenIcon class="w-12 h-12 text-base-content/50 mb-3 mx-auto"/>
        <p class="text-base-content/70">暂无项目数据</p>
      </div>

      <div v-else class="space-y-2">
        <div
            v-for="project in projects"
            :key="project.id"
            class="bg-base-100 rounded-lg shadow-sm border border-base-300 p-3 hover:shadow-md transition-shadow cursor-move"
        >
          <div class="flex items-center gap-3">
            <!-- 拖拽手柄 -->
            <div class="flex-shrink-0">
              <Bars3Icon class="w-4 h-4 text-base-content/30 cursor-grab"/>
            </div>

            <!-- 项目Logo -->
            <div class="avatar flex-shrink-0 hidden sm:block">
              <div class="w-8 h-8 rounded-lg">
                <img v-if="project.logoUrl" :src="project.logoUrl" :alt="project.name" class="object-cover"/>
                <div v-else class="bg-base-200 flex items-center justify-center w-full h-full">
                  <BookOpenIcon class="w-4 h-4 text-base-content/50"/>
                </div>
              </div>
            </div>

            <!-- 项目信息 -->
            <div class="flex-1 min-w-0">
              <!-- 第一行：项目名称 + 徽章 + 操作按钮 -->
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-2 flex-1 min-w-0">
                  <h4 class="font-semibold text-base-content text-base truncate">{{ project.name }}</h4>
                  <div class="flex gap-1 flex-shrink-0">
                    <span class="badge badge-xs" :class="getStatusClass(project.status)">
                      {{ getStatusText(project.status) }}
                    </span>
                    <span class="badge badge-xs" :class="getPublicClass(project.isPublic)">
                      {{ getPublicText(project.isPublic) }}
                    </span>
                  </div>
                </div>

                <!-- 操作按钮 - 响应式设计 -->
                <div class="flex gap-1 flex-shrink-0">
                  <!-- 大屏幕：显示完整按钮 -->
                  <div class="hidden lg:flex gap-1">
                    <button class="btn btn-xs btn-ghost" @click="viewProject(project)" title="查看">
                      <EyeIcon class="w-3 h-3"/>
                      <span class="ml-1">查看</span>
                    </button>
                    <button class="btn btn-xs btn-warning" @click="editProject(project)" title="编辑">
                      <PencilIcon class="w-3 h-3"/>
                      <span class="ml-1">编辑</span>
                    </button>
                    <button class="btn btn-xs btn-error" @click="deleteProject(project)" title="删除">
                      <TrashIcon class="w-3 h-3"/>
                      <span class="ml-1">删除</span>
                    </button>
                  </div>

                  <!-- 中等屏幕：只显示图标 -->
                  <div class="hidden md:flex lg:hidden gap-1">
                    <button class="btn btn-xs btn-ghost" @click="viewProject(project)" title="查看">
                      <EyeIcon class="w-3 h-3"/>
                    </button>
                    <button class="btn btn-xs btn-warning" @click="editProject(project)" title="编辑">
                      <PencilIcon class="w-3 h-3"/>
                    </button>
                    <button class="btn btn-xs btn-error" @click="deleteProject(project)" title="删除">
                      <TrashIcon class="w-3 h-3"/>
                    </button>
                  </div>

                  <!-- 小屏幕：下拉菜单 -->
                  <div class="dropdown dropdown-end md:hidden">
                    <div tabindex="0" role="button" class="btn btn-xs btn-ghost">
                      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01"></path>
                      </svg>
                    </div>
                    <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-32">
                      <li><a @click="viewProject(project)"><EyeIcon class="w-3 h-3"/>查看</a></li>
                      <li><a @click="editProject(project)"><PencilIcon class="w-3 h-3"/>编辑</a></li>
                      <li><a @click="deleteProject(project)"><TrashIcon class="w-3 h-3"/>删除</a></li>
                    </ul>
                  </div>
                </div>
              </div>

              <!-- 第二行：项目标识 + 描述 + 所有者 + 创建时间 -->
              <div class="flex items-center gap-2 mt-1 text-xs text-base-content/70">
                <span class="font-mono bg-base-200 px-1.5 py-0.5 rounded flex-shrink-0">{{ project.slug }}</span>
                <span class="truncate flex-1 hidden sm:block">{{ project.description || '暂无描述' }}</span>
                <span class="flex-shrink-0 hidden md:block">{{ project.ownerName || '未知用户' }}</span>
                <span class="flex-shrink-0 hidden lg:block">{{ formatDate(project.createdAt) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="flex justify-between items-center mt-4 pt-4 border-t border-base-200">
        <div class="text-xs text-base-content/60">
          共 {{ total }} 个项目
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

    <!-- 创建/编辑项目模态框 -->
    <dialog id="project_modal" class="modal">
      <div class="modal-box w-11/12 max-w-2xl">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>

        <h3 class="font-bold text-lg mb-4">{{ isEditing ? '编辑项目' : '新建项目' }}</h3>

        <form @submit.prevent="submitForm" class="space-y-4">
          <!-- 项目名称 -->
          <label class="floating-label">
            <input
                v-model="formData.name"
                type="text"
                required
                placeholder="请输入项目名称"
                class="input input-bordered w-full"
            />
            <span>项目名称 *</span>
          </label>

          <!-- 项目标识 -->
          <fieldset class="fieldset">
            <label class="floating-label">
              <input
                  v-model="formData.slug"
                  type="text"
                  required
                  placeholder="请输入项目标识"
                  pattern="[a-zA-Z0-9-_]+"
                  class="input input-bordered w-full"
              />
              <span>项目标识 *</span>
            </label>
            <span class="label text-xs text-base-content/60">用于URL路径，只能包含字母、数字、连字符和下划线</span>
          </fieldset>

          <!-- 项目描述 -->
          <label class="floating-label">
            <textarea
                v-model="formData.description"
                placeholder="请输入项目描述"
                rows="3"
                class="textarea textarea-bordered w-full"
            ></textarea>
            <span>项目描述</span>
          </label>

          <!-- Logo URL -->
          <label class="floating-label">
            <input
                v-model="formData.logoUrl"
                type="url"
                placeholder="请输入Logo图片URL"
                class="input input-bordered w-full"
            />
            <span>Logo URL</span>
          </label>

          <!-- 状态和公开性 -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <label class="select select-bordered">
              <span class="label">状态 *</span>
              <select v-model="formData.status" required>
                <option v-for="option in PROJECT_STATUS_OPTIONS" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </label>

            <label class="select select-bordered">
              <span class="label">公开性 *</span>
              <select v-model="formData.isPublic" required>
                <option v-for="option in PUBLIC_STATUS_OPTIONS" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </label>
          </div>

          <!-- 所有者ID -->
          <label class="floating-label">
            <input
                v-model.number="formData.ownerId"
                type="number"
                required
                placeholder="请输入所有者用户ID"
                class="input input-bordered w-full"
            />
            <span>所有者ID *</span>
          </label>
        </form>

        <div class="modal-action">
          <button class="btn" @click="closeModal">取消</button>
          <button class="btn btn-primary" @click="submitForm" :disabled="submitting">
            <span v-if="submitting" class="loading loading-spinner loading-sm"></span>
            {{ submitting ? '提交中...' : (isEditing ? '更新' : '创建') }}
          </button>
        </div>
      </div>
    </dialog>
  </div>
</template>

<script setup lang="ts">
import {computed, onMounted, reactive, ref} from 'vue';
import DocsProjectService from '../../utils/docsService';
import type {
  DocsProjectCreateReq,
  DocsProjectItem,
  DocsProjectListReq,
  DocsProjectUpdateReq
} from '../../utils/docsInterfaces';
import {PROJECT_STATUS_OPTIONS, ProjectStatus, PUBLIC_STATUS_OPTIONS, PublicStatus} from '../../utils/docsInterfaces';
import {
  ArrowPathIcon,
  Bars3Icon,
  BookOpenIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  EyeIcon,
  FolderOpenIcon,
  MagnifyingGlassIcon,
  PencilIcon,
  PlusIcon,
  TrashIcon
} from '@heroicons/vue/24/outline';

// 响应式数据
const loading = ref(false);
const submitting = ref(false);
const projects = ref<DocsProjectItem[]>([]);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(20);

// 搜索参数
const searchParams = reactive<DocsProjectListReq>({
  page: 1,
  pageSize: 20,
  keyword: '',
  status: undefined,
  isPublic: undefined
});

// 模态框状态
const isEditing = ref(false);
const editingProject = ref<DocsProjectItem | null>(null);

// 表单数据
const formData = reactive<DocsProjectCreateReq>({
  name: '',
  slug: '',
  description: '',
  logoUrl: '',
  status: ProjectStatus.DRAFT,
  isPublic: PublicStatus.PRIVATE,
  ownerId: 0
});

// 计算属性
const totalPages = computed(() => Math.ceil(total.value / pageSize.value));

// 方法
const loadProjects = async () => {
  loading.value = true;
  try {
    const params = {
      ...searchParams,
      page: currentPage.value,
      pageSize: pageSize.value
    };
    const response = await DocsProjectService.getProjectList(params);
    projects.value = response.list;
    total.value = response.total;
  } catch (error) {
    console.error('加载项目列表失败:', error);
    alert('加载项目列表失败，请重试');
  } finally {
    loading.value = false;
  }
};

const resetSearch = () => {
  searchParams.keyword = '';
  searchParams.status = undefined;
  searchParams.isPublic = undefined;
  currentPage.value = 1;
  loadProjects();
};

const changePage = (page: number) => {
  currentPage.value = page;
  loadProjects();
};

const viewProject = (project: DocsProjectItem) => {
  // 可以跳转到项目详情页面或显示详情模态框
  console.log('查看项目:', project);
};

const openCreateModal = () => {
  isEditing.value = false;
  resetForm();
  const modal = document.getElementById('project_modal') as HTMLDialogElement;
  modal?.showModal();
};

const editProject = (project: DocsProjectItem) => {
  isEditing.value = true;
  editingProject.value = project;
  formData.name = project.name;
  formData.slug = project.slug;
  formData.description = project.description;
  formData.logoUrl = project.logoUrl;
  formData.status = project.status;
  formData.isPublic = project.isPublic;
  formData.ownerId = project.ownerId;
  const modal = document.getElementById('project_modal') as HTMLDialogElement;
  modal?.showModal();
};

const deleteProject = async (project: DocsProjectItem) => {
  if (!confirm(`确定要删除项目 "${project.name}" 吗？此操作不可恢复。`)) {
    return;
  }

  try {
    await DocsProjectService.deleteProject(project.id);
    alert('删除成功');
    loadProjects();
  } catch (error) {
    console.error('删除项目失败:', error);
    alert('删除项目失败，请重试');
  }
};

const closeModal = () => {
  const modal = document.getElementById('project_modal') as HTMLDialogElement;
  modal?.close();
  isEditing.value = false;
  editingProject.value = null;
  resetForm();
};

const resetForm = () => {
  formData.name = '';
  formData.slug = '';
  formData.description = '';
  formData.logoUrl = '';
  formData.status = ProjectStatus.DRAFT;
  formData.isPublic = PublicStatus.PRIVATE;
  formData.ownerId = 0;
};

const submitForm = async () => {
  submitting.value = true;
  try {
    if (!isEditing.value) {
      await DocsProjectService.createProject(formData);
      alert('创建成功');
    } else if (editingProject.value) {
      const updateData: DocsProjectUpdateReq = {
        id: editingProject.value.id,
        ...formData
      };
      await DocsProjectService.updateProject(editingProject.value.id, updateData);
      alert('更新成功');
    }
    closeModal();
    loadProjects();
  } catch (error) {
    console.error('提交失败:', error);
    alert('操作失败，请重试');
  } finally {
    submitting.value = false;
  }
};

// 状态相关方法
const getStatusText = (status: number) => {
  const option = PROJECT_STATUS_OPTIONS.find(opt => opt.value === status);
  return option?.label || '未知';
};

const getStatusClass = (status: number) => {
  switch (status) {
    case ProjectStatus.DRAFT:
      return 'badge-neutral';
    case ProjectStatus.ACTIVE:
      return 'badge-success';
    case ProjectStatus.ARCHIVED:
      return 'badge-warning';
    default:
      return 'badge-neutral';
  }
};

const getPublicText = (isPublic: number) => {
  return isPublic === PublicStatus.PUBLIC ? '公开' : '私有';
};

const getPublicClass = (isPublic: number) => {
  return isPublic === PublicStatus.PUBLIC ? 'badge-info' : 'badge-ghost';
};

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString('zh-CN');
};

// 生命周期
onMounted(() => {
  loadProjects();
});
</script>