// 文档管理相关接口类型定义

// 项目列表请求参数
export interface DocsProjectListReq {
  page: number;
  pageSize: number;
  keyword?: string;
  status?: number;
  isPublic?: number;
}

// 创建项目请求参数
export interface DocsProjectCreateReq {
  name: string;
  slug: string;
  description?: string;
  logoUrl?: string;
  status: number;
  isPublic: number;
  ownerId: number;
}

// 更新项目请求参数
export interface DocsProjectUpdateReq {
  id: number;
  name: string;
  slug: string;
  description?: string;
  logoUrl?: string;
  status: number;
  isPublic: number;
  ownerId: number;
}

// 项目列表项
export interface DocsProjectItem {
  id: number;
  name: string;
  slug: string;
  description: string;
  logoUrl: string;
  status: number;
  isPublic: number;
  ownerId: number;
  ownerName: string;
  createdAt: string;
  updatedAt: string;
}

// 分页响应
export interface PageResponse<T> {
  list: T[];
  page: number;
  size: number;
  total: number;
}

// API响应格式
export interface ApiResponse<T = any> {
  code: number;
  msg: string | null;
  result: T;
}

// 项目状态枚举
export enum ProjectStatus {
  DRAFT = 1,
  ACTIVE = 2,
  ARCHIVED = 3
}

// 公开状态枚举
export enum PublicStatus {
  PRIVATE = 0,
  PUBLIC = 1
}

// 项目状态选项
export const PROJECT_STATUS_OPTIONS = [
  { label: '草稿', value: ProjectStatus.DRAFT },
  { label: '活跃', value: ProjectStatus.ACTIVE },
  { label: '归档', value: ProjectStatus.ARCHIVED }
];

// 公开状态选项
export const PUBLIC_STATUS_OPTIONS = [
  { label: '私有', value: PublicStatus.PRIVATE },
  { label: '公开', value: PublicStatus.PUBLIC }
];