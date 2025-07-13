import axiosInstance from './axiosInstance';
import type {
  DocsProjectListReq,
  DocsProjectCreateReq,
  DocsProjectUpdateReq,
  DocsProjectItem,
  DocsVersionListReq,
  DocsVersionCreateReq,
  DocsVersionUpdateReq,
  DocsVersionItem,
  DocsContentListReq,
  DocsContentCreateReq,
  DocsContentUpdateReq,
  DocsContentDeleteReq,
  DocsContentPublishReq,
  DocsContentDraftReq,
  DocsContentPreviewReq,
  DocsContentItem,
  DocsContentPreviewResponse,
  DirectoryNode,
  PageResponse,
  ApiResponse
} from './docsInterfaces';

// 文档项目管理API服务
export class DocsProjectService {
  // 获取项目列表
  static async getProjectList(params: DocsProjectListReq): Promise<PageResponse<DocsProjectItem>> {
    const response = await axiosInstance.post(
      '/api/admin/docs/projects/list',
      params
    ) as any;
    return response.result;
  }

  // 获取项目详情
  static async getProjectDetail(id: number): Promise<DocsProjectItem> {
    const response = await axiosInstance.get(
      `/api/admin/docs/projects/${id}`
    ) as any;
    return response.result;
  }

  // 创建项目
  static async createProject(data: DocsProjectCreateReq): Promise<DocsProjectItem> {
    const response = await axiosInstance.post(
      '/api/admin/docs/projects',
      data
    ) as any;
    return response.result;
  }

  // 更新项目
  static async updateProject(id: number, data: DocsProjectUpdateReq): Promise<DocsProjectItem> {
    const response = await axiosInstance.put(
      `/api/admin/docs/projects/${id}`,
      data
    ) as any;
    return response.result;
  }

  // 删除项目
  static async deleteProject(id: number): Promise<void> {
    await axiosInstance.delete(
      `/api/admin/docs/projects/${id}`
    );
  }

  // 文档版本管理API服务
  // 获取版本列表
  static async getVersionList(params: DocsVersionListReq): Promise<PageResponse<DocsVersionItem>> {
    const response = await axiosInstance.post(
      '/api/admin/docs/versions/list',
      params
    ) as any;
    return response.result;
  }

  // 获取版本详情
  static async getVersionDetail(id: number): Promise<DocsVersionItem> {
    const response = await axiosInstance.get(
      `/api/admin/docs/versions/${id}`
    ) as any;
    return response.result;
  }

  // 创建版本
  static async createVersion(data: DocsVersionCreateReq): Promise<DocsVersionItem> {
    const response = await axiosInstance.post(
      '/api/admin/docs/versions',
      data
    ) as any;
    return response.result;
  }

  // 更新版本
  static async updateVersion(id: number, data: DocsVersionUpdateReq): Promise<DocsVersionItem> {
    const response = await axiosInstance.put(
      `/api/admin/docs/versions/${id}`,
      data
    ) as any;
    return response.result;
  }

  // 删除版本
  static async deleteVersion(id: number): Promise<void> {
    await axiosInstance.delete(
      `/api/admin/docs/versions/${id}`
    );
  }

  // 设置默认版本
  static async setDefaultVersion(id: number): Promise<void> {
    await axiosInstance.put(
      `/api/admin/docs/versions/${id}/set-default`
    );
  }

  // 更新目录结构
  static async updateDirectory(id: number, directoryStructure: DirectoryNode[]): Promise<void> {
    await axiosInstance.put(
      `/api/admin/docs/versions/${id}/directory`,
      { directoryStructure }
    );
  }

  // 获取项目的版本列表（用于下拉选择）
  static async getProjectVersions(projectId: number): Promise<DocsVersionItem[]> {
    const response = await axiosInstance.get(
      `/api/admin/docs/projects/${projectId}/versions`
    ) as any;
    return response.result;
  }
}

// 文档内容管理API服务
export class DocsContentService {
  // 获取内容列表
  static async getContentList(params: DocsContentListReq): Promise<PageResponse<DocsContentItem>> {
    const response = await axiosInstance.post(
      '/api/admin/docs/contents/list',
      params
    ) as any;
    return response.result;
  }

  // 获取内容详情
  static async getContentDetail(id: number): Promise<DocsContentItem> {
    const response = await axiosInstance.get(
      `/api/admin/docs/contents/${id}`
    ) as any;
    return response.result;
  }

  // 创建内容
  static async createContent(data: DocsContentCreateReq): Promise<DocsContentItem> {
    const response = await axiosInstance.post(
      '/api/admin/docs/contents',
      data
    ) as any;
    return response.result;
  }

  // 更新内容
  static async updateContent(id: number, data: DocsContentUpdateReq): Promise<DocsContentItem> {
    const response = await axiosInstance.put(
      `/api/admin/docs/contents/${id}`,
      data
    ) as any;
    return response.result;
  }

  // 删除内容
  static async deleteContent(data: DocsContentDeleteReq): Promise<void> {
    await axiosInstance.delete(
      `/api/admin/docs/contents/${data.id}`
    );
  }

  // 发布内容
  static async publishContent(data: DocsContentPublishReq): Promise<void> {
    await axiosInstance.put(
      `/api/admin/docs/contents/${data.id}/publish`
    );
  }

  // 设为草稿
  static async draftContent(data: DocsContentDraftReq): Promise<void> {
    await axiosInstance.put(
      `/api/admin/docs/contents/${data.id}/draft`
    );
  }

  // 预览内容
  static async previewContent(data: DocsContentPreviewReq): Promise<DocsContentPreviewResponse> {
    const response = await axiosInstance.post(
      '/api/admin/docs/contents/preview',
      data
    ) as any;
    return response.result;
  }

  // 获取版本的内容列表（用于下拉选择）
  static async getVersionContents(versionId: number): Promise<DocsContentItem[]> {
    const response = await axiosInstance.get(
      `/api/admin/docs/versions/${versionId}/contents`
    ) as any;
    return response.result;
  }
}

export default DocsProjectService;
