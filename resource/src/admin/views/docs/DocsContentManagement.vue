<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center mb-2">
      <h2 class="text-xl font-normal text-base-content">文档内容管理</h2>
      <button class="btn btn-primary btn-sm" @click="openCreateModal">
        <PlusIcon class="w-4 h-4"/>
        新建内容
      </button>
    </div>

    <!-- 搜索筛选 -->
    <div class="card bg-base-100 shadow-sm mb-6">
      <div class="card-body p-4">
        <div class="flex flex-wrap gap-3 items-end">
          <div class="flex-1 min-w-64">
             <label class="floating-label join w-full">
               <span>搜索标题或内容</span>
               <input
                   v-model="searchParams.keyword"
                   type="text"
                   placeholder="搜索标题或内容"
                   class="input input-bordered input-sm w-full"
                   @keyup.enter="loadContents"
               />
             </label>
           </div>

          <div class="min-w-40">
             <label class="floating-label join w-full">
               <span>项目</span>
               <select v-model="selectedProjectId" class="select select-bordered select-sm" @change="onProjectChange">
                 <option value="">全部项目</option>
                 <option v-for="project in projects" :key="project.id" :value="project.id">
                   {{ project.name }}
                 </option>
               </select>
             </label>
           </div>

           <div class="min-w-40">
             <label class="floating-label join w-full">
               <span>版本</span>
               <select v-model="searchParams.versionId" class="select select-bordered select-sm" :disabled="!selectedProjectId">
                 <option value="">全部版本</option>
                 <option v-for="version in filteredVersions" :key="version.id" :value="version.id">
                   {{ version.name }}
                 </option>
               </select>
             </label>
           </div>

           <div class="min-w-32">
             <label class="floating-label join w-full">
               <span>状态</span>
               <select v-model="searchParams.status" class="select select-bordered select-sm">
                 <option value="">全部状态</option>
                 <option v-for="option in CONTENT_STATUS_OPTIONS" :key="option.value" :value="option.value">
                   {{ option.label }}
                 </option>
               </select>
             </label>
           </div>

          <button class="btn btn-neutral btn-sm" @click="loadContents">
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

    <!-- 内容列表 -->
    <div>
      <div v-if="loading" class="flex justify-center items-center py-8">
        <span class="loading loading-spinner loading-lg"></span>
        <span class="ml-2 text-base-content">加载中...</span>
      </div>

      <div v-else-if="!contents || contents.length === 0" class="text-center py-8">
        <DocumentTextIcon class="w-12 h-12 text-base-content/50 mb-3 mx-auto"/>
        <p class="text-base-content/70">暂无内容数据</p>
      </div>

      <div v-else-if="contents && contents.length > 0" class="space-y-2">
        <div
            v-for="content in contents"
            :key="content.id"
            class="bg-base-100 rounded-lg shadow-sm border border-base-300 p-3 hover:shadow-md transition-shadow"
        >
          <div class="flex items-center gap-3">
            <!-- 内容信息 -->
            <div class="flex-1 min-w-0">
              <!-- 第一行：标题 + 徽章 + 操作按钮 -->
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-2 flex-1 min-w-0">
                  <h4 class="font-semibold text-base-content text-base truncate">{{ content.title }}</h4>
                  <div class="flex gap-1 flex-shrink-0">
                    <span class="badge badge-xs" :class="getStatusClass(content.isPublished)">
                      {{ getStatusText(content.isPublished) }}
                    </span>
                    <span class="badge badge-xs badge-outline">
                      排序: {{ content.sortOrder }}
                    </span>
                  </div>
                </div>

                <!-- 操作按钮 - 响应式设计 -->
                <div class="flex gap-1 flex-shrink-0">
                  <!-- 大屏幕：显示完整按钮 -->
                  <div class="hidden lg:flex gap-1">
                    <button class="btn btn-xs btn-ghost" @click="viewContent(content)" title="查看">
                      <EyeIcon class="w-3 h-3"/>
                      <span class="ml-1">查看</span>
                    </button>
                    <button class="btn btn-xs btn-warning" @click="editContent(content)" title="编辑">
                      <PencilIcon class="w-3 h-3"/>
                      <span class="ml-1">编辑</span>
                    </button>
                    <button
                        v-if="content.isPublished === 0"
                        class="btn btn-xs btn-success"
                        @click="publishContent(content)"
                        title="发布"
                    >
                      <CheckIcon class="w-3 h-3"/>
                      <span class="ml-1">发布</span>
                    </button>
                    <button
                        v-else
                        class="btn btn-xs btn-neutral"
                        @click="draftContent(content)"
                        title="设为草稿"
                    >
                      <ArchiveBoxIcon class="w-3 h-3"/>
                      <span class="ml-1">草稿</span>
                    </button>
                    <button class="btn btn-xs btn-error" @click="deleteContent(content)" title="删除">
                      <TrashIcon class="w-3 h-3"/>
                      <span class="ml-1">删除</span>
                    </button>
                  </div>

                  <!-- 中等屏幕：只显示图标 -->
                  <div class="hidden md:flex lg:hidden gap-1">
                    <button class="btn btn-xs btn-ghost" @click="viewContent(content)" title="查看">
                      <EyeIcon class="w-3 h-3"/>
                    </button>
                    <button class="btn btn-xs btn-warning" @click="editContent(content)" title="编辑">
                      <PencilIcon class="w-3 h-3"/>
                    </button>
                    <button
                        v-if="content.isPublished === 0"
                        class="btn btn-xs btn-success"
                        @click="publishContent(content)"
                        title="发布"
                    >
                      <CheckIcon class="w-3 h-3"/>
                    </button>
                    <button
                        v-else
                        class="btn btn-xs btn-neutral"
                        @click="draftContent(content)"
                        title="设为草稿"
                    >
                      <ArchiveBoxIcon class="w-3 h-3"/>
                    </button>
                    <button class="btn btn-xs btn-error" @click="deleteContent(content)" title="删除">
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
                      <li><a @click="viewContent(content)"><EyeIcon class="w-3 h-3"/>查看</a></li>
                      <li><a @click="editContent(content)"><PencilIcon class="w-3 h-3"/>编辑</a></li>
                      <li v-if="content.isPublished === 0"><a @click="publishContent(content)"><CheckIcon class="w-3 h-3"/>发布</a></li>
                      <li v-else><a @click="draftContent(content)"><ArchiveBoxIcon class="w-3 h-3"/>草稿</a></li>
                      <li><a @click="deleteContent(content)"><TrashIcon class="w-3 h-3"/>删除</a></li>
                    </ul>
                  </div>
                </div>
              </div>

              <!-- 第二行：内容标识 + 版本信息 + 统计信息 + 更新时间 -->
              <div class="flex items-center gap-2 mt-1 text-xs text-base-content/70">
                <span class="font-mono bg-base-200 px-1.5 py-0.5 rounded flex-shrink-0">{{ content.slug }}</span>
                <span class="flex-shrink-0 hidden sm:block">{{ content.projectName }} - {{ content.versionName }}</span>
                <span class="flex-shrink-0 hidden md:block">浏览: {{ content.viewCount }}</span>
                <span class="flex-shrink-0 hidden md:block">点赞: {{ content.likeCount }}</span>
                <span class="flex-shrink-0 hidden lg:block">{{ formatDate(content.updatedAt) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="flex justify-between items-center mt-4 pt-4 border-t border-base-200">
        <div class="text-xs text-base-content/60">
          共 {{ total }} 个内容
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

    <!-- 创建/编辑内容模态框 -->
    <dialog id="content_modal" class="modal">
      <div class="modal-box w-11/12 max-w-4xl max-h-[90vh] overflow-y-auto">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
        <h3 class="font-bold text-lg mb-4">{{ isEditing ? '编辑内容' : '新建内容' }}</h3>

        <div class="space-y-4">
          <!-- 项目选择 -->
          <div class="flex flex-col space-y-4 sm:flex-row sm:space-x-4 sm:space-y-0">
            <div class="flex-1">
              <label class="floating-label">
                <span class="label-text">所属项目 <span class="text-error">*</span></span>
                <select v-model="formData.projectId" class="select select-bordered w-full" :disabled="isEditing" @change="onFormProjectChange">
                  <option value="">请选择项目</option>
                  <option v-for="project in projects" :key="project.id" :value="project.id">
                    {{ project.name }}
                  </option>
                </select>
              </label>
            </div>
            <div class="flex-1">
              <label class="floating-label">
                <span class="label-text">所属版本 <span class="text-error">*</span></span>
                <select v-model="formData.versionId" class="select select-bordered w-full" :disabled="isEditing || !formData.projectId">
                  <option value="0">请选择版本</option>
                  <option v-for="version in formFilteredVersions" :key="version.id" :value="version.id">
                    {{ version.name }}
                  </option>
                </select>
              </label>
            </div>
          </div>


          <!-- 标题 -->
          <div>
            <div class="relative">
              <input
                  v-model="formData.title"
                  type="text"
                  placeholder=" "
                  class="input input-bordered w-full"
                  maxlength="200"
                  required
              />
              <span class="absolute -top-2 left-3 bg-base-200 px-1 text-xs text-base-content/70 z-10">标题 *</span>
            </div>
          </div>

          <!-- Slug -->
          <div class="flex flex-col space-y-4 sm:flex-row sm:space-x-4 sm:space-y-0">
            <!-- Slug -->
            <div class="flex-1">
              <div class="relative">
                <input
                    v-model="formData.slug"
                    type="text"
                    placeholder=" "
                    class="input input-bordered w-full"
                    pattern="[a-zA-Z0-9-]+"
                    maxlength="100"
                    required
                />
                <span class="absolute -top-2 left-3 bg-base-200 px-1 text-xs text-base-content/70 z-10">URL标识 *</span>
              </div>
              <div class="text-xs text-base-content/60 mt-1">用于生成URL，只能包含字母、数字、连字符</div>
            </div>

            <!-- 排序权重 -->
            <div class="w-full sm:w-32">
              <div class="relative">
                <input
                    v-model.number="formData.sortOrder"
                    type="number"
                    placeholder=" "
                    class="input input-bordered w-full"
                    min="0"
                    max="9999"
                />
                <span class="absolute -top-2 left-3 bg-base-200 px-1 text-xs text-base-content/70 z-10">排序权重</span>
              </div>
              <div class="text-xs text-base-content/60 mt-1">数值越小排序越靠前</div>
            </div>
          </div>

          <!-- 内容编辑器 -->
          <div>
            <label class="label">
              <span class="label-text">内容</span>
              <span class="label-text-alt">支持Markdown格式</span>
            </label>
            <div class="border border-base-300 rounded-lg">
              <div class="flex border-b border-base-300 bg-base-200 rounded-t-lg">
                <button
                    type="button"
                    class="px-3 py-2 text-sm"
                    :class="editorTab === 'edit' ? 'bg-base-100 border-b-2 border-primary' : 'hover:bg-base-300'"
                    @click="editorTab = 'edit'"
                >
                  编辑
                </button>
                <button
                    type="button"
                    class="px-3 py-2 text-sm"
                    :class="editorTab === 'preview' ? 'bg-base-100 border-b-2 border-primary' : 'hover:bg-base-300'"
                    @click="previewContent"
                >
                  预览
                </button>
              </div>
              <div class="p-3">
                <textarea
                    v-if="editorTab === 'edit'"
                    v-model="formData.content"
                    class="textarea w-full h-64 resize-none border-0 focus:outline-none"
                    placeholder="请输入内容，支持Markdown格式"
                ></textarea>
                <div
                    v-else
                    class="prose max-w-none h-64 overflow-y-auto"
                    v-html="previewHtml"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-action">
          <button type="button" class="btn" @click="closeModal">取消</button>
          <button type="button" class="btn btn-primary" @click="saveContent" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-sm"></span>
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- 查看内容模态框 -->
    <dialog id="view_content_modal" class="modal">
      <div class="modal-box w-11/12 max-w-4xl max-h-[90vh] overflow-y-auto">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
        <div v-if="viewingContent">
          <div class="flex items-center gap-2 mb-4">
            <h3 class="font-bold text-lg">{{ viewingContent.title }}</h3>
            <span class="badge badge-sm" :class="getStatusClass(viewingContent.isPublished)">
              {{ getStatusText(viewingContent.isPublished) }}
            </span>
          </div>

          <div class="text-sm text-base-content/70 mb-4">
            <p>版本：{{ viewingContent.projectName }} - {{ viewingContent.versionName }}</p>
            <p>URL标识：{{ viewingContent.slug }}</p>
            <p>排序权重：{{ viewingContent.sortOrder }}</p>
            <p>浏览次数：{{ viewingContent.viewCount }}</p>
            <p>点赞次数：{{ viewingContent.likeCount }}</p>
            <p>创建时间：{{ formatDate(viewingContent.createdAt) }}</p>
            <p>更新时间：{{ formatDate(viewingContent.updatedAt) }}</p>
          </div>

          <div class="divider"></div>

          <div class="prose max-w-none" v-html="viewingContent.content"></div>
        </div>
      </div>
    </dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import {
  PlusIcon,
  MagnifyingGlassIcon,
  ArrowPathIcon,
  EyeIcon,
  PencilIcon,
  TrashIcon,
  CheckIcon,
  ArchiveBoxIcon,
  DocumentTextIcon,
  ChevronLeftIcon,
  ChevronRightIcon
} from '@heroicons/vue/24/outline';
import { DocsContentService, DocsProjectService } from '../../utils/docsService';
import type {
  DocsContentItem,
  DocsContentListReq,
  DocsContentCreateReq,
  DocsContentUpdateReq,
  DocsVersionItem,
  DocsProjectItem
} from '../../utils/docsInterfaces';
import { CONTENT_STATUS_OPTIONS } from '../../utils/docsInterfaces';
import { formatDate } from '../../utils/dateUtils';
import { showToast } from '../../utils/toast';

// 响应式数据
const loading = ref(false);
const saving = ref(false);
const contents = ref<DocsContentItem[]>([]);
const projects = ref<DocsProjectItem[]>([]);
const versions = ref<DocsVersionItem[]>([]);
const selectedProjectId = ref<number | string>('');
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(20);
const isEditing = ref(false);
const viewingContent = ref<DocsContentItem | null>(null);
const editorTab = ref<'edit' | 'preview'>('edit');
const previewHtml = ref('');

// 搜索参数
const searchParams = reactive<DocsContentListReq>({
  page: 1,
  pageSize: 20,
  keyword: '',
  versionId: undefined,
  status: undefined
});

// 表单数据
const formData = reactive<DocsContentCreateReq & { id?: number; projectId?: number | string }>({
  projectId: '',
  versionId: 0,
  title: '',
  slug: '',
  content: '',
  sortOrder: 0
});

// 计算属性
const totalPages = computed(() => Math.ceil(total.value / pageSize.value));

// 根据选中的项目过滤版本列表（筛选区域用）
const filteredVersions = computed(() => {
  if (!selectedProjectId.value) {
    return versions.value;
  }
  return versions.value.filter(version => version.projectId === Number(selectedProjectId.value));
});

// 根据表单中选中的项目过滤版本列表（新建/编辑表单用）
const formFilteredVersions = computed(() => {
  if (!formData.projectId) {
    return [];
  }
  return versions.value.filter(version => version.projectId === Number(formData.projectId));
});

// 状态样式
const getStatusClass = (isPublished: number) => {
  return isPublished === 1 ? 'badge-success' : 'badge-warning';
};

const getStatusText = (isPublished: number) => {
  return isPublished === 1 ? '已发布' : '草稿';
};

// 加载项目列表
const loadProjects = async () => {
  try {
    const response = await DocsProjectService.getProjectList({ page: 1, pageSize: 1000 });
    projects.value = response.list || [];
  } catch (error) {
    console.error('加载项目列表失败:', error);
    showToast('加载项目列表失败', 'error');
    // 确保在错误情况下projects仍然是空数组
    projects.value = [];
  }
};

// 加载版本列表
const loadVersions = async () => {
  try {
    const response = await DocsProjectService.getVersionList({ page: 1, pageSize: 1000 });
    versions.value = response.list || [];
  } catch (error) {
    console.error('加载版本列表失败:', error);
    showToast('加载版本列表失败', 'error');
    // 确保在错误情况下versions仍然是空数组
    versions.value = [];
  }
};

// 项目变化处理（筛选区域用）
const onProjectChange = () => {
  // 清空版本选择
  searchParams.versionId = undefined;
  // 重新加载内容列表
  currentPage.value = 1;
  loadContents();
};

// 表单项目变化处理
const onFormProjectChange = () => {
  // 清空版本选择
  formData.versionId = 0;
};

// 加载内容列表
const loadContents = async () => {
  loading.value = true;
  try {
    searchParams.page = currentPage.value;
    searchParams.pageSize = pageSize.value;
    const response = await DocsContentService.getContentList(searchParams);
    contents.value = response.list || [];
    total.value = response.total || 0;
  } catch (error) {
    console.error('加载内容列表失败:', error);
    showToast('加载内容列表失败', 'error');
    // 确保在错误情况下contents仍然是空数组
    contents.value = [];
    total.value = 0;
  } finally {
    loading.value = false;
  }
};

// 重置搜索
const resetSearch = () => {
  searchParams.keyword = '';
  searchParams.versionId = undefined;
  searchParams.status = undefined;
  selectedProjectId.value = '';
  currentPage.value = 1;
  loadContents();
};

// 分页
const changePage = (page: number) => {
  currentPage.value = page;
  loadContents();
};

// 打开创建模态框
const openCreateModal = () => {
  isEditing.value = false;
  resetFormData();
  editorTab.value = 'edit';
  (document.getElementById('content_modal') as HTMLDialogElement)?.showModal();
};

// 编辑内容
const editContent = async (content: DocsContentItem) => {
  isEditing.value = true;
  try {
    const detail = await DocsContentService.getContentDetail(content.id);
    formData.id = detail.id;
    formData.versionId = detail.versionId;
    formData.title = detail.title;
    formData.slug = detail.slug;
    formData.content = detail.content;
    formData.sortOrder = detail.sortOrder;

    // 根据版本ID找到对应的项目ID
    const version = versions.value.find(v => v.id === detail.versionId);
    formData.projectId = version ? version.projectId : '';

    editorTab.value = 'edit';
    (document.getElementById('content_modal') as HTMLDialogElement)?.showModal();
  } catch (error) {
    console.error('加载内容详情失败:', error);
    showToast('加载内容详情失败', 'error');
  }
};

// 查看内容
const viewContent = async (content: DocsContentItem) => {
  try {
    const detail = await DocsContentService.getContentDetail(content.id);
    viewingContent.value = detail;
    (document.getElementById('view_content_modal') as HTMLDialogElement)?.showModal();
  } catch (error) {
    console.error('加载内容详情失败:', error);
    showToast('加载内容详情失败', 'error');
  }
};

// 预览内容
const previewContent = async () => {
  if (!formData.content.trim()) {
    previewHtml.value = '<p class="text-base-content/50">暂无内容</p>';
    editorTab.value = 'preview';
    return;
  }

  try {
    const response = await DocsContentService.previewContent({ content: formData.content });
    previewHtml.value = response.html;
    editorTab.value = 'preview';
  } catch (error) {
    console.error('预览内容失败:', error);
    showToast('预览内容失败', 'error');
  }
};

// 保存内容
const saveContent = async () => {
  if (!formData.versionId || !formData.title.trim() || !formData.slug.trim()) {
    showToast('请填写必填字段', 'warning');
    return;
  }

  saving.value = true;
  try {
    if (isEditing.value && formData.id) {
      await DocsContentService.updateContent(formData.id, {
        id: formData.id,
        versionId: formData.versionId,
        title: formData.title,
        slug: formData.slug,
        content: formData.content,
        sortOrder: formData.sortOrder
      });
      showToast('内容更新成功', 'success');
    } else {
      await DocsContentService.createContent({
        versionId: formData.versionId,
        title: formData.title,
        slug: formData.slug,
        content: formData.content,
        sortOrder: formData.sortOrder
      });
      showToast('内容创建成功', 'success');
    }
    closeModal();
    loadContents();
  } catch (error) {
    console.error('保存内容失败:', error);
    showToast('保存内容失败', 'error');
  } finally {
    saving.value = false;
  }
};

// 发布内容
const publishContent = async (content: DocsContentItem) => {
  try {
    await DocsContentService.publishContent({ id: content.id });
    showToast('内容发布成功', 'success');
    loadContents();
  } catch (error) {
    console.error('发布内容失败:', error);
    showToast('发布内容失败', 'error');
  }
};

// 设为草稿
const draftContent = async (content: DocsContentItem) => {
  try {
    await DocsContentService.draftContent({ id: content.id });
    showToast('内容已设为草稿', 'success');
    loadContents();
  } catch (error) {
    console.error('设为草稿失败:', error);
    showToast('设为草稿失败', 'error');
  }
};

// 删除内容
const deleteContent = async (content: DocsContentItem) => {
  if (!confirm(`确定要删除内容「${content.title}」吗？此操作不可恢复。`)) {
    return;
  }

  try {
    await DocsContentService.deleteContent({ id: content.id });
    showToast('内容删除成功', 'success');
    loadContents();
  } catch (error) {
    console.error('删除内容失败:', error);
    showToast('删除内容失败', 'error');
  }
};

// 关闭模态框
const closeModal = () => {
  (document.getElementById('content_modal') as HTMLDialogElement)?.close();
};

// 重置表单数据
const resetFormData = () => {
  formData.id = undefined;
  formData.projectId = '';
  formData.versionId = 0;
  formData.title = '';
  formData.slug = '';
  formData.content = '';
  formData.sortOrder = 0;
};

// 组件挂载
onMounted(() => {
  loadProjects();
  loadVersions();
  loadContents();
});
</script>

<style scoped>
.prose {
  color: hsl(var(--bc));
}

.prose h1, .prose h2, .prose h3, .prose h4, .prose h5, .prose h6 {
  color: hsl(var(--bc));
}

.prose code {
  background-color: hsl(var(--b2));
  color: hsl(var(--bc));
}

.prose pre {
  background-color: hsl(var(--b2));
  color: hsl(var(--bc));
}

.prose blockquote {
  border-left-color: hsl(var(--p));
  color: hsl(var(--bc) / 0.8);
}
</style>
