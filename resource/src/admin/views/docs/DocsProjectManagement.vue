<template>
  <div class="docs-project-management">
    <!-- 页面标题 -->
    <div class="page-header">
      <h2>文档项目管理</h2>
      <button class="btn btn-primary" @click="showCreateModal = true">
        <i class="fas fa-plus"></i> 新建项目
      </button>
    </div>

    <!-- 搜索筛选 -->
    <div class="search-filters">
      <div class="filter-row">
        <div class="filter-item">
          <label>关键词:</label>
          <input 
            v-model="searchParams.keyword" 
            type="text" 
            placeholder="搜索项目名称或描述"
            @keyup.enter="loadProjects"
          />
        </div>
        <div class="filter-item">
          <label>状态:</label>
          <select v-model="searchParams.status">
            <option value="">全部状态</option>
            <option v-for="option in PROJECT_STATUS_OPTIONS" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </div>
        <div class="filter-item">
          <label>公开性:</label>
          <select v-model="searchParams.isPublic">
            <option value="">全部</option>
            <option v-for="option in PUBLIC_STATUS_OPTIONS" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </div>
        <div class="filter-actions">
          <button class="btn btn-secondary" @click="loadProjects">
            <i class="fas fa-search"></i> 搜索
          </button>
          <button class="btn btn-light" @click="resetSearch">
            <i class="fas fa-undo"></i> 重置
          </button>
        </div>
      </div>
    </div>

    <!-- 项目列表 -->
    <div class="project-list">
      <div v-if="loading" class="loading">
        <i class="fas fa-spinner fa-spin"></i> 加载中...
      </div>
      
      <div v-else-if="projects.length === 0" class="empty-state">
        <i class="fas fa-folder-open"></i>
        <p>暂无项目数据</p>
      </div>

      <div v-else class="table-container">
        <table class="table">
          <thead>
            <tr>
              <th>项目信息</th>
              <th>状态</th>
              <th>所有者</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="project in projects" :key="project.id">
              <td>
                <div class="project-info">
                  <div class="project-logo">
                    <img v-if="project.logoUrl" :src="project.logoUrl" :alt="project.name" />
                    <i v-else class="fas fa-book"></i>
                  </div>
                  <div class="project-details">
                    <h4>{{ project.name }}</h4>
                    <p class="slug">{{ project.slug }}</p>
                    <p class="description">{{ project.description || '暂无描述' }}</p>
                  </div>
                </div>
              </td>
              <td>
                <div class="status-badges">
                  <span class="badge" :class="getStatusClass(project.status)">
                    {{ getStatusText(project.status) }}
                  </span>
                  <span class="badge" :class="getPublicClass(project.isPublic)">
                    {{ getPublicText(project.isPublic) }}
                  </span>
                </div>
              </td>
              <td>
                <div class="owner-info">
                  <span>{{ project.ownerName || '未知用户' }}</span>
                  <small>(ID: {{ project.ownerId }})</small>
                </div>
              </td>
              <td>
                <div class="time-info">
                  <div>{{ formatDate(project.createdAt) }}</div>
                  <small>更新: {{ formatDate(project.updatedAt) }}</small>
                </div>
              </td>
              <td>
                <div class="actions">
                  <button class="btn btn-sm btn-info" @click="viewProject(project)">
                    <i class="fas fa-eye"></i>
                  </button>
                  <button class="btn btn-sm btn-warning" @click="editProject(project)">
                    <i class="fas fa-edit"></i>
                  </button>
                  <button class="btn btn-sm btn-danger" @click="deleteProject(project)">
                    <i class="fas fa-trash"></i>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="pagination">
        <button 
          class="btn btn-sm" 
          :disabled="currentPage <= 1" 
          @click="changePage(currentPage - 1)"
        >
          上一页
        </button>
        <span class="page-info">
          第 {{ currentPage }} 页，共 {{ totalPages }} 页，总计 {{ total }} 条
        </span>
        <button 
          class="btn btn-sm" 
          :disabled="currentPage >= totalPages" 
          @click="changePage(currentPage + 1)"
        >
          下一页
        </button>
      </div>
    </div>

    <!-- 创建/编辑项目模态框 -->
    <div v-if="showCreateModal || showEditModal" class="modal-overlay" @click="closeModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>{{ showCreateModal ? '新建项目' : '编辑项目' }}</h3>
          <button class="close-btn" @click="closeModal">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="submitForm">
            <div class="form-group">
              <label>项目名称 *</label>
              <input 
                v-model="formData.name" 
                type="text" 
                required 
                placeholder="请输入项目名称"
              />
            </div>
            <div class="form-group">
              <label>项目标识 *</label>
              <input 
                v-model="formData.slug" 
                type="text" 
                required 
                placeholder="请输入项目标识（英文字母、数字）"
                pattern="[a-zA-Z0-9-_]+"
              />
              <small>用于URL路径，只能包含字母、数字、连字符和下划线</small>
            </div>
            <div class="form-group">
              <label>项目描述</label>
              <textarea 
                v-model="formData.description" 
                placeholder="请输入项目描述"
                rows="3"
              ></textarea>
            </div>
            <div class="form-group">
              <label>Logo URL</label>
              <input 
                v-model="formData.logoUrl" 
                type="url" 
                placeholder="请输入Logo图片URL"
              />
            </div>
            <div class="form-row">
              <div class="form-group">
                <label>状态 *</label>
                <select v-model="formData.status" required>
                  <option v-for="option in PROJECT_STATUS_OPTIONS" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
              </div>
              <div class="form-group">
                <label>公开性 *</label>
                <select v-model="formData.isPublic" required>
                  <option v-for="option in PUBLIC_STATUS_OPTIONS" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
              </div>
            </div>
            <div class="form-group">
              <label>所有者ID *</label>
              <input 
                v-model.number="formData.ownerId" 
                type="number" 
                required 
                placeholder="请输入所有者用户ID"
              />
            </div>
          </form>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeModal">取消</button>
          <button class="btn btn-primary" @click="submitForm" :disabled="submitting">
            <i v-if="submitting" class="fas fa-spinner fa-spin"></i>
            {{ submitting ? '提交中...' : (showCreateModal ? '创建' : '更新') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import DocsProjectService from '../../utils/docsService';
import {
  PROJECT_STATUS_OPTIONS,
  PUBLIC_STATUS_OPTIONS,
  ProjectStatus,
  PublicStatus
} from '../../utils/docsInterfaces';
import type {
  DocsProjectItem,
  DocsProjectListReq,
  DocsProjectCreateReq,
  DocsProjectUpdateReq
} from '../../utils/docsInterfaces';

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
const showCreateModal = ref(false);
const showEditModal = ref(false);
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

const editProject = (project: DocsProjectItem) => {
  editingProject.value = project;
  formData.name = project.name;
  formData.slug = project.slug;
  formData.description = project.description;
  formData.logoUrl = project.logoUrl;
  formData.status = project.status;
  formData.isPublic = project.isPublic;
  formData.ownerId = project.ownerId;
  showEditModal.value = true;
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
  showCreateModal.value = false;
  showEditModal.value = false;
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
    if (showCreateModal.value) {
      await DocsProjectService.createProject(formData);
      alert('创建成功');
    } else if (showEditModal.value && editingProject.value) {
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
    case ProjectStatus.DRAFT: return 'badge-secondary';
    case ProjectStatus.ACTIVE: return 'badge-success';
    case ProjectStatus.ARCHIVED: return 'badge-warning';
    default: return 'badge-secondary';
  }
};

const getPublicText = (isPublic: number) => {
  return isPublic === PublicStatus.PUBLIC ? '公开' : '私有';
};

const getPublicClass = (isPublic: number) => {
  return isPublic === PublicStatus.PUBLIC ? 'badge-info' : 'badge-dark';
};

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString('zh-CN');
};

// 生命周期
onMounted(() => {
  loadProjects();
});
</script>

<style scoped>
.docs-project-management {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  color: #333;
}

.search-filters {
  background: #f8f9fa;
  padding: 15px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.filter-row {
  display: flex;
  gap: 15px;
  align-items: end;
  flex-wrap: wrap;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.filter-item label {
  font-weight: 500;
  color: #555;
}

.filter-item input,
.filter-item select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  min-width: 150px;
}

.filter-actions {
  display: flex;
  gap: 10px;
}

.loading {
  text-align: center;
  padding: 40px;
  color: #666;
}

.empty-state {
  text-align: center;
  padding: 60px;
  color: #999;
}

.empty-state i {
  font-size: 48px;
  margin-bottom: 15px;
}

.table-container {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.table {
  width: 100%;
  border-collapse: collapse;
}

.table th,
.table td {
  padding: 12px 15px;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.table th {
  background: #f8f9fa;
  font-weight: 600;
  color: #555;
}

.project-info {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.project-logo {
  width: 40px;
  height: 40px;
  border-radius: 6px;
  overflow: hidden;
  background: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.project-logo img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.project-logo i {
  color: #999;
  font-size: 18px;
}

.project-details h4 {
  margin: 0 0 4px 0;
  font-size: 16px;
  color: #333;
}

.project-details .slug {
  margin: 0 0 4px 0;
  font-size: 12px;
  color: #666;
  font-family: monospace;
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  display: inline-block;
}

.project-details .description {
  margin: 0;
  font-size: 13px;
  color: #777;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-badges {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
  text-align: center;
  min-width: 50px;
}

.badge-success { background: #d4edda; color: #155724; }
.badge-warning { background: #fff3cd; color: #856404; }
.badge-secondary { background: #e2e3e5; color: #383d41; }
.badge-info { background: #d1ecf1; color: #0c5460; }
.badge-dark { background: #d6d8db; color: #1b1e21; }

.owner-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.owner-info small {
  color: #999;
  font-size: 11px;
}

.time-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 13px;
}

.time-info small {
  color: #999;
  font-size: 11px;
}

.actions {
  display: flex;
  gap: 5px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary { background: #007bff; color: white; }
.btn-primary:hover:not(:disabled) { background: #0056b3; }

.btn-secondary { background: #6c757d; color: white; }
.btn-secondary:hover:not(:disabled) { background: #545b62; }

.btn-info { background: #17a2b8; color: white; }
.btn-info:hover:not(:disabled) { background: #117a8b; }

.btn-warning { background: #ffc107; color: #212529; }
.btn-warning:hover:not(:disabled) { background: #e0a800; }

.btn-danger { background: #dc3545; color: white; }
.btn-danger:hover:not(:disabled) { background: #c82333; }

.btn-light { background: #f8f9fa; color: #212529; border: 1px solid #dee2e6; }
.btn-light:hover:not(:disabled) { background: #e2e6ea; }

.btn-sm {
  padding: 4px 8px;
  font-size: 12px;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 15px;
  margin-top: 20px;
  padding: 20px;
}

.page-info {
  color: #666;
  font-size: 14px;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 8px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #eee;
}

.modal-header h3 {
  margin: 0;
  color: #333;
}

.close-btn {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: #999;
  padding: 5px;
}

.close-btn:hover {
  color: #333;
}

.modal-body {
  padding: 20px;
}

.form-group {
  margin-bottom: 15px;
}

.form-row {
  display: flex;
  gap: 15px;
}

.form-row .form-group {
  flex: 1;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
  color: #555;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0,123,255,0.25);
}

.form-group small {
  display: block;
  margin-top: 5px;
  color: #666;
  font-size: 12px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 20px;
  border-top: 1px solid #eee;
}
</style>