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

// ===== 版本管理相关接口 =====

// 版本列表请求参数
export interface DocsVersionListReq {
  page: number;
  pageSize: number;
  projectId?: number;
  keyword?: string;
  status?: number;
}

// 创建版本请求参数
export interface DocsVersionCreateReq {
  projectId: number;
  name: string;
  slug: string;
  description?: string;
  status: number;
  isDefault: number;
  sortOrder: number;
  directoryStructure?: string; // JSON格式的目录结构
}

// 更新版本请求参数
export interface DocsVersionUpdateReq {
  id: number;
  projectId: number;
  name: string;
  slug: string;
  description?: string;
  status: number;
  isDefault: number;
  sortOrder: number;
  directoryStructure?: string;
}

// 版本列表项
export interface DocsVersionItem {
  id: number;
  projectId: number;
  projectName: string;
  name: string;
  slug: string;
  description: string;
  status: number;
  isDefault: number;
  sortOrder: number;
  directoryStructure: DirectoryNode[];
  createdAt: string;
  updatedAt: string;
}

// 版本状态枚举
export enum VersionStatus {
  DRAFT = 1,
  PUBLISHED = 2,
  ARCHIVED = 3
}

// 默认版本枚举
export enum DefaultStatus {
  NOT_DEFAULT = 0,
  IS_DEFAULT = 1
}

// 版本状态选项
export const VERSION_STATUS_OPTIONS = [
  { label: '草稿', value: VersionStatus.DRAFT },
  { label: '已发布', value: VersionStatus.PUBLISHED },
  { label: '归档', value: VersionStatus.ARCHIVED }
];

// 默认版本选项
export const DEFAULT_STATUS_OPTIONS = [
  { label: '否', value: DefaultStatus.NOT_DEFAULT },
  { label: '是', value: DefaultStatus.IS_DEFAULT }
];

// 目录结构节点
export interface DirectoryNode {
  title: string;
  slug: string;
  description?: string;
  children?: DirectoryNode[];
}