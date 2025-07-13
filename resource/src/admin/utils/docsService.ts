import axiosInstance from './axiosInstance';
import type {
  DocsProjectListReq,
  DocsProjectCreateReq,
  DocsProjectUpdateReq,
  DocsProjectItem,
  PageResponse,
  ApiResponse
} from './docsInterfaces';

// 文档项目管理API服务
export class DocsProjectService {
  // 获取项目列表
  static async getProjectList(params: DocsProjectListReq): Promise<PageResponse<DocsProjectItem>> {
    const response = await axiosInstance.post<ApiResponse<PageResponse<DocsProjectItem>>>(
      '/api/admin/docs/projects/list',
      params
    );
    return response.data.result;
  }

  // 获取项目详情
  static async getProjectDetail(id: number): Promise<DocsProjectItem> {
    const response = await axiosInstance.get<ApiResponse<DocsProjectItem>>(
      `/api/admin/docs/projects/${id}`
    );
    return response.data.result;
  }

  // 创建项目
  static async createProject(data: DocsProjectCreateReq): Promise<DocsProjectItem> {
    const response = await axiosInstance.post<ApiResponse<DocsProjectItem>>(
      '/api/admin/docs/projects',
      data
    );
    return response.data.result;
  }

  // 更新项目
  static async updateProject(id: number, data: DocsProjectUpdateReq): Promise<DocsProjectItem> {
    const response = await axiosInstance.put<ApiResponse<DocsProjectItem>>(
      `/api/admin/docs/projects/${id}`,
      data
    );
    return response.data.result;
  }

  // 删除项目
  static async deleteProject(id: number): Promise<void> {
    await axiosInstance.delete<ApiResponse<void>>(
      `/api/admin/docs/projects/${id}`
    );
  }
}

export default DocsProjectService;