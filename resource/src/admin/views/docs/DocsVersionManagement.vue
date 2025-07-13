<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center mb-2">
      <h2 class="text-xl font-normal text-base-content">文档版本管理</h2>
      <button class="btn btn-primary btn-sm" @click="openCreateModal">
        <PlusIcon class="w-4 h-4"/>
        新建版本
      </button>
    </div>

    <!-- 搜索筛选 -->
    <div class="card bg-base-100 shadow-sm mb-6">
      <div class="card-body p-4">
        <div class="flex flex-wrap gap-3 items-end">
          <div class="flex-1 min-w-64">
            <label class="floating-label join w-full">
              <span>搜索版本名称或描述</span>
              <input
                  v-model="searchParams.keyword"
                  type="text"
                  placeholder="搜索版本名称或描述"
                  class="input input-bordered input-sm w-full"
                  @keyup.enter="loadVersions"
              />
            </label>
          </div>

          <div class="min-w-40">
            <label class="floating-label join w-full">
              <span>项目</span>
              <select v-model="searchParams.projectId" class="select select-bordered select-sm">
                <option value="">全部项目</option>
                <option v-for="project in projects" :key="project.id" :value="project.id">
                  {{ project.name }}
                </option>
              </select>
            </label>
          </div>

          <div class="min-w-32">
            <label class="floating-label join w-full">
              <span>状态</span>
              <select v-model="searchParams.status" class="select select-bordered select-sm">
                <option value="">全部状态</option>
                <option v-for="option in VERSION_STATUS_OPTIONS" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </label>
          </div>

          <button class="btn btn-neutral btn-sm" @click="loadVersions">
            <MagnifyingGlassIcon class="w-4 h-4"/>
            搜索
          </button>

          <button class="btn btn-ghost btn-sm" @click="resetSearch">
            <ArrowPathIcon class="w-4 h-4"/>
            重置
          </button>
        </div>
      </div>
    </div>

    <!-- 版本列表 -->
    <div>
      <div v-if="loading" class="flex justify-center items-center py-8">
        <span class="loading loading-spinner loading-lg"></span>
        <span class="ml-2 text-base-content">加载中...</span>
      </div>

      <div v-else-if="versions.length === 0" class="text-center py-8">
        <DocumentTextIcon class="w-12 h-12 text-base-content/50 mb-3 mx-auto"/>
        <p class="text-base-content/70">暂无版本数据</p>
      </div>

      <div v-else class="bg-base-100">
        <ul class="list shadow-md rounded-box ">
          <li v-for="version in versions" :key="version.id"
              class="flex items-center justify-between px-4 py-2 hover:bg-base-200 rounded-lg transition-colors list-row">
            <!-- 左侧：版本信息 -->
            <div class="flex items-center gap-3 flex-1 min-w-0">
              <!-- 版本详情 -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-1">
                  <h4 class="font-semibold text-base truncate">{{ version.name }}</h4>
                  <div class="flex items-center gap-1 flex-shrink-0">
                    <span v-if="version.isDefault" class="badge badge-primary badge-xs">默认</span>
                    <span class="badge badge-xs" :class="getStatusClass(version.status)">
                      {{ getStatusText(version.status) }}
                    </span>
                  </div>
                </div>
                <div class="flex items-center gap-2 text-sm text-base-content/70">
                  <span class="font-mono bg-base-200 px-2 py-0.5 rounded text-xs">
                    {{ version.slug }}
                  </span>
                  <span class="truncate">{{ version.projectName }}</span>
                  <span v-if="version.description" class="truncate hidden sm:block text-base-content/50">
                    · {{ version.description }}
                  </span>
                </div>
                <div class="flex items-center gap-2 mt-1 text-xs text-base-content/50">
                  <span>排序: #{{ version.sortOrder }}</span>
                  <span class="hidden sm:block">{{ formatDate(version.createdAt) }}</span>
                </div>
              </div>
            </div>

            <!-- 右侧：操作按钮 -->
            <div class="flex gap-1 flex-shrink-0">
              <!-- 大屏幕显示完整按钮 -->
              <div class="hidden lg:flex gap-1">
                <button class="btn btn-xs btn-ghost" @click="viewVersion(version)" title="查看">
                  <EyeIcon class="w-3 h-3"/>
                  <span class="ml-1">查看</span>
                </button>
                <button class="btn btn-xs btn-info" @click="editDirectory(version)" title="目录结构">
                  <FolderIcon class="w-3 h-3"/>
                  <span class="ml-1">目录</span>
                </button>
                <button
                    v-if="!version.isDefault"
                    class="btn btn-xs btn-success"
                    @click="setDefaultVersion(version)"
                    title="设为默认"
                >
                  <StarIcon class="w-3 h-3"/>
                  <span class="ml-1">设为默认</span>
                </button>
                <button class="btn btn-xs btn-warning" @click="editVersion(version)" title="编辑">
                  <PencilIcon class="w-3 h-3"/>
                  <span class="ml-1">编辑</span>
                </button>
                <button class="btn btn-xs btn-error" @click="deleteVersion(version)" title="删除">
                  <TrashIcon class="w-3 h-3"/>
                  <span class="ml-1">删除</span>
                </button>
              </div>

              <!-- 中等屏幕显示图标按钮 -->
              <div class="hidden md:flex lg:hidden gap-1">
                <button class="btn btn-xs btn-ghost" @click="viewVersion(version)" title="查看">
                  <EyeIcon class="w-3 h-3"/>
                </button>
                <button class="btn btn-xs btn-info" @click="editDirectory(version)" title="目录结构">
                  <FolderIcon class="w-3 h-3"/>
                </button>
                <button
                    v-if="!version.isDefault"
                    class="btn btn-xs btn-success"
                    @click="setDefaultVersion(version)"
                    title="设为默认"
                >
                  <StarIcon class="w-3 h-3"/>
                </button>
                <button class="btn btn-xs btn-warning" @click="editVersion(version)" title="编辑">
                  <PencilIcon class="w-3 h-3"/>
                </button>
                <button class="btn btn-xs btn-error" @click="deleteVersion(version)" title="删除">
                  <TrashIcon class="w-3 h-3"/>
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
                  <li><a @click="viewVersion(version)">查看</a></li>
                  <li><a @click="editDirectory(version)">目录</a></li>
                  <li v-if="!version.isDefault"><a @click="setDefaultVersion(version)">设为默认</a></li>
                  <li><a @click="editVersion(version)">编辑</a></li>
                  <li><a @click="deleteVersion(version)" class="text-error">删除</a></li>
                </ul>
              </div>
            </div>
          </li>
        </ul>
      </div>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="flex justify-between items-center mt-4 pt-4 border-t border-base-200">
        <div class="text-xs text-base-content/60">
          共 {{ total }} 个版本
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

    <!-- 创建/编辑版本模态框 -->
    <dialog id="version_modal" class="modal">
      <div class="modal-box w-11/12 max-w-2xl">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>

        <h3 class="font-normal text-lg mb-4">{{ isEditing ? '编辑版本' : '新建版本' }}</h3>

        <form @submit.prevent="submitForm" class="space-y-4">
          <!-- 所属项目 -->
          <label class="floating-label">
            <select v-model="formData.projectId" required class="select select-bordered w-full">
              <option value="">请选择项目</option>
              <option v-for="project in projects" :key="project.id" :value="project.id">
                {{ project.name }}
              </option>
            </select>
            <span>所属项目 *</span>
          </label>

          <!-- 版本名称 -->
          <label class="floating-label">
            <input
                v-model="formData.name"
                type="text"
                required
                placeholder="请输入版本名称"
                class="input input-bordered w-full"
            />
            <span>版本名称 *</span>
          </label>

          <!-- 版本标识 -->
          <fieldset class="fieldset">
            <label class="floating-label">
              <input
                  v-model="formData.slug"
                  type="text"
                  required
                  placeholder="请输入版本标识"
                  pattern="[a-zA-Z0-9.-_]+"
                  class="input input-bordered w-full"
              />
              <span>版本标识 *</span>
            </label>
            <span class="label text-xs text-base-content/60">用于URL路径，只能包含字母、数字、点号、连字符和下划线</span>
          </fieldset>

          <!-- 版本描述 -->
          <label class="floating-label">
            <textarea
                v-model="formData.description"
                placeholder="请输入版本描述"
                rows="3"
                class="textarea textarea-bordered w-full"
            ></textarea>
            <span>版本描述</span>
          </label>

          <!-- 状态和默认版本 -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <label class="select select-bordered">
              <span class="label">状态 *</span>
              <select v-model="formData.status" required>
                <option v-for="option in VERSION_STATUS_OPTIONS" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </label>

            <label class="select select-bordered">
              <span class="label">是否默认版本 *</span>
              <select v-model="formData.isDefault" required>
                <option v-for="option in DEFAULT_STATUS_OPTIONS" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </label>
          </div>

          <!-- 排序 -->
          <label class="floating-label">
            <input
                v-model.number="formData.sortOrder"
                type="number"
                required
                placeholder="请输入排序值"
                class="input input-bordered w-full"
            />
            <span>排序值 *</span>
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

    <!-- 目录结构编辑模态框 -->
    <dialog id="directory_modal" class="modal">
      <div class="modal-box w-11/12 max-w-4xl">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>

        <h3 class="font-normal text-lg mb-4">编辑目录结构 - {{ editingVersion?.name }}</h3>

        <div class="space-y-4">
          <div class="alert alert-info">
            <InformationCircleIcon class="w-5 h-5"/>
            <span>使用可视化编辑器编辑目录结构，支持拖拽排序和嵌套文件夹</span>
          </div>

          <!-- 树形目录编辑器 -->
          <DirectoryTreeEditor v-model="directoryStructure"/>
        </div>

        <div class="modal-action">
          <button class="btn" @click="closeDirectoryModal">取消</button>
          <button class="btn btn-primary" @click="saveDirectory" :disabled="submitting">
            <span v-if="submitting" class="loading loading-spinner loading-sm"></span>
            {{ submitting ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </dialog>
  </div>
</template>

<script setup lang="ts">
import {computed, onMounted, reactive, ref} from 'vue';
import {DocsProjectService} from '../../utils/docsService';
import type {
  DirectoryNode,
  DocsProjectItem,
  DocsVersionCreateReq,
  DocsVersionItem,
  DocsVersionListReq,
  DocsVersionUpdateReq
} from '../../utils/docsInterfaces';
import {DEFAULT_STATUS_OPTIONS, DefaultStatus, VERSION_STATUS_OPTIONS, VersionStatus} from '../../utils/docsInterfaces';
import {
  ArrowPathIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  DocumentTextIcon,
  EyeIcon,
  FolderIcon,
  InformationCircleIcon,
  MagnifyingGlassIcon,
  PencilIcon,
  PlusIcon,
  StarIcon,
  TrashIcon
} from '@heroicons/vue/24/outline';
import DirectoryTreeEditor from '../../components/DirectoryTreeEditor.vue';

// 响应式数据
const loading = ref(false);
const submitting = ref(false);
const versions = ref<DocsVersionItem[]>([]);
const projects = ref<DocsProjectItem[]>([]);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(20);

// 搜索参数
const searchParams = reactive<DocsVersionListReq>({
  page: 1,
  pageSize: 20,
  keyword: '',
  projectId: undefined,
  status: undefined
});

// 模态框状态
const isEditing = ref(false);
const editingVersion = ref<DocsVersionItem | null>(null);
const directoryStructure = ref<DirectoryNode[]>([]);

// 表单数据
const formData = reactive<DocsVersionCreateReq>({
  projectId: 0,
  name: '',
  slug: '',
  description: '',
  status: VersionStatus.DRAFT,
  isDefault: DefaultStatus.NOT_DEFAULT,
  sortOrder: 1
});

// 计算属性
const totalPages = computed(() => Math.ceil(total.value / pageSize.value));

// 方法
const loadVersions = async () => {
  loading.value = true;
  try {
    const params = {
      ...searchParams,
      page: currentPage.value,
      pageSize: pageSize.value
    };
    const response = await DocsProjectService.getVersionList(params);
    versions.value = response.list;
    total.value = response.total;
  } catch (error) {
    console.error('加载版本列表失败:', error);
    alert('加载版本列表失败，请重试');
  } finally {
    loading.value = false;
  }
};

const loadProjects = async () => {
  try {
    const response = await DocsProjectService.getProjectList({
      page: 1,
      pageSize: 1000 // 获取所有项目用于下拉选择
    });
    projects.value = response.list;
  } catch (error) {
    console.error('加载项目列表失败:', error);
  }
};

const resetSearch = () => {
  searchParams.keyword = '';
  searchParams.projectId = undefined;
  searchParams.status = undefined;
  currentPage.value = 1;
  loadVersions();
};

const changePage = (page: number) => {
  currentPage.value = page;
  loadVersions();
};

const viewVersion = (version: DocsVersionItem) => {
  // 可以跳转到版本详情页面或显示详情模态框
  console.log('查看版本:', version);
};

const openCreateModal = () => {
  isEditing.value = false;
  resetForm();
  const modal = document.getElementById('version_modal') as HTMLDialogElement;
  modal?.showModal();
};

const editVersion = (version: DocsVersionItem) => {
  isEditing.value = true;
  editingVersion.value = version;
  formData.projectId = version.projectId;
  formData.name = version.name;
  formData.slug = version.slug;
  formData.description = version.description;
  formData.status = version.status;
  formData.isDefault = version.isDefault;
  formData.sortOrder = version.sortOrder;
  const modal = document.getElementById('version_modal') as HTMLDialogElement;
  modal?.showModal();
};

const editDirectory = (version: DocsVersionItem) => {
  editingVersion.value = version;
  directoryStructure.value = version.directoryStructure || [];
  const modal = document.getElementById('directory_modal') as HTMLDialogElement;
  modal?.showModal();
};

const setDefaultVersion = async (version: DocsVersionItem) => {
  if (!confirm(`确定要将版本 "${version.name}" 设为默认版本吗？`)) {
    return;
  }

  try {
    await DocsProjectService.setDefaultVersion(version.id);
    alert('设置成功');
    loadVersions();
  } catch (error) {
    console.error('设置默认版本失败:', error);
    alert(error?.message ?? '设置默认版本失败，请重试');
  }
};

const deleteVersion = async (version: DocsVersionItem) => {
  if (!confirm(`确定要删除版本 "${version.name}" 吗？此操作不可恢复。`)) {
    return;
  }

  try {
    await DocsProjectService.deleteVersion(version.id);
    alert('删除成功');
    loadVersions();
  } catch (error) {
    console.error('删除版本失败:', error);
    alert('删除版本失败，请重试');
  }
};

const closeModal = () => {
  const modal = document.getElementById('version_modal') as HTMLDialogElement;
  modal?.close();
  isEditing.value = false;
  editingVersion.value = null;
  resetForm();
};

const closeDirectoryModal = () => {
  const modal = document.getElementById('directory_modal') as HTMLDialogElement;
  modal?.close();
  editingVersion.value = null;
  directoryStructure.value = [];
};

const resetForm = () => {
  formData.projectId = 0;
  formData.name = '';
  formData.slug = '';
  formData.description = '';
  formData.status = VersionStatus.DRAFT;
  formData.isDefault = DefaultStatus.NOT_DEFAULT;
  formData.sortOrder = 1;
};

const submitForm = async () => {
  submitting.value = true;
  try {
    if (!isEditing.value) {
      await DocsProjectService.createVersion(formData);
      alert('创建成功');
    } else if (editingVersion.value) {
      const updateData: DocsVersionUpdateReq = {
        id: editingVersion.value.id,
        ...formData
      };
      await DocsProjectService.updateVersion(editingVersion.value.id, updateData);
      alert('更新成功');
    }
    closeModal();
    loadVersions();
  } catch (error) {
    console.error('提交失败:', error);
    alert('操作失败，请重试');
  } finally {
    submitting.value = false;
  }
};

const saveDirectory = async () => {
  if (!editingVersion.value) return;

  submitting.value = true;
  try {
    await DocsProjectService.updateDirectory(editingVersion.value.id, directoryStructure.value);
    alert('保存成功');
    closeDirectoryModal();
    loadVersions();
  } catch (error) {
    console.error('保存目录结构失败:', error);
    alert('保存目录结构失败，请重试');
  } finally {
    submitting.value = false;
  }
};

// 状态相关方法
const getStatusText = (status: number) => {
  const option = VERSION_STATUS_OPTIONS.find(opt => opt.value === status);
  return option?.label || '未知';
};

const getStatusClass = (status: number) => {
  switch (status) {
    case VersionStatus.DRAFT:
      return 'badge-neutral';
    case VersionStatus.PUBLISHED:
      return 'badge-success';
    case VersionStatus.ARCHIVED:
      return 'badge-warning';
    default:
      return 'badge-neutral';
  }
};

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString('zh-CN');
};

// 生命周期
onMounted(() => {
  loadProjects();
  loadVersions();
});
</script>