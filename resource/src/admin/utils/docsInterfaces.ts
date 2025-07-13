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

// ===== 内容管理相关接口 =====

// 内容列表请求参数
export interface DocsContentListReq {
  page: number;
  pageSize: number;
  versionId?: number;
  keyword?: string;
  status?: number; // 0:草稿 1:已发布
}

// 创建内容请求参数
export interface DocsContentCreateReq {
  versionId: number;
  title: string;
  slug: string;
  content?: string;
  sortOrder?: number;
}

// 更新内容请求参数
export interface DocsContentUpdateReq {
  id: number;
  versionId: number;
  title: string;
  slug: string;
  content?: string;
  sortOrder?: number;
}

// 删除内容请求参数
export interface DocsContentDeleteReq {
  id: number;
}

// 发布内容请求参数
export interface DocsContentPublishReq {
  id: number;
}

// 设为草稿请求参数
export interface DocsContentDraftReq {
  id: number;
}

// 预览内容请求参数
export interface DocsContentPreviewReq {
  content: string;
}

// 内容列表项
export interface DocsContentItem {
  id: number;
  versionId: number;
  versionName: string;
  projectId: number;
  projectName: string;
  title: string;
  slug: string;
  content: string;
  isPublished: number;
  sortOrder: number;
  viewCount: number;
  likeCount: number;
  createdAt: string;
  updatedAt: string;
}

// 预览响应
export interface DocsContentPreviewResponse {
  html: string;
  toc: string;
}

// 内容发布状态枚举
export enum ContentStatus {
  DRAFT = 0,
  PUBLISHED = 1
}

// 内容状态选项
export const CONTENT_STATUS_OPTIONS = [
  { label: '草稿', value: ContentStatus.DRAFT },
  { label: '已发布', value: ContentStatus.PUBLISHED }
];

// 内容排序选项
export const CONTENT_SORT_OPTIONS = [
  { label: '创建时间（最新）', value: 'created_at_desc' },
  { label: '创建时间（最早）', value: 'created_at_asc' },
  { label: '更新时间（最新）', value: 'updated_at_desc' },
  { label: '更新时间（最早）', value: 'updated_at_asc' },
  { label: '排序权重（升序）', value: 'sort_order_asc' },
  { label: '排序权重（降序）', value: 'sort_order_desc' },
  { label: '标题（A-Z）', value: 'title_asc' },
  { label: '标题（Z-A）', value: 'title_desc' }
];