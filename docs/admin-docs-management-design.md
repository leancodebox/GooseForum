# 后台文档管理系统设计方案

## 概述

基于现有的文档ORM模型（doc_projects、doc_versions、doc_contents、operation_logs），设计一套完整的后台文档管理系统。采用渐进式开发模式，按小模块逐步实现。

## 系统架构

### 数据模型关系
```
doc_projects (项目)
    ↓ 1:N
doc_versions (版本)
    ↓ 1:N  
doc_contents (内容)
    ↓ 记录到
operation_logs (操作日志)
```

### 技术栈
- **后端**: Go + Gin + GORM
- **前端**: TypeScript + Vue 3 + Element Plus
- **数据库**: MySQL
- **编辑器**: Monaco Editor (支持Markdown)

## 功能模块设计

### 1. 项目管理模块 (Project Management)

**功能特性:**
- 项目CRUD操作
- 项目状态管理（活跃/维护/废弃）
- 项目公开性控制
- 项目Logo上传
- 项目所有者管理

**API接口:**
```
GET    /api/admin/docs/projects          # 项目列表
POST   /api/admin/docs/projects          # 创建项目
GET    /api/admin/docs/projects/:id      # 项目详情
PUT    /api/admin/docs/projects/:id      # 更新项目
DELETE /api/admin/docs/projects/:id      # 删除项目（软删除）
```

### 2. 版本管理模块 (Version Management)

**功能特性:**
- 版本CRUD操作
- 默认版本设置
- 版本发布状态控制
- 版本排序管理
- 目录结构管理（JSON格式）

**API接口:**
```
GET    /api/admin/docs/versions                    # 版本列表
POST   /api/admin/docs/versions                    # 创建版本
GET    /api/admin/docs/versions/:id                # 版本详情
PUT    /api/admin/docs/versions/:id                # 更新版本
DELETE /api/admin/docs/versions/:id                # 删除版本
PUT    /api/admin/docs/versions/:id/set-default    # 设置默认版本
PUT    /api/admin/docs/versions/:id/directory      # 更新目录结构
```

### 3. 内容管理模块 (Content Management)

**功能特性:**
- 文档CRUD操作
- Markdown编辑器
- 实时预览
- 自动保存草稿
- 内容发布控制
- 文档排序管理

**API接口:**
```
GET    /api/admin/docs/contents                # 内容列表
POST   /api/admin/docs/contents                # 创建内容
GET    /api/admin/docs/contents/:id            # 内容详情
PUT    /api/admin/docs/contents/:id            # 更新内容
DELETE /api/admin/docs/contents/:id            # 删除内容
POST   /api/admin/docs/contents/:id/publish    # 发布内容
POST   /api/admin/docs/contents/:id/draft      # 设为草稿
POST   /api/admin/docs/contents/preview        # 预览内容
```

### 4. 操作日志模块 (Operation Logs)

**功能特性:**
- 操作历史查看
- 变更对比
- 用户操作追踪
- 数据恢复支持

**API接口:**
```
GET    /api/admin/docs/logs                    # 操作日志列表
GET    /api/admin/docs/logs/:id                # 日志详情
GET    /api/admin/docs/logs/entity/:type/:id   # 特定实体的日志
POST   /api/admin/docs/logs/compare            # 版本对比
```

## 渐进式开发计划

### 第一阶段：项目管理模块
1. **后端开发**
   - 创建 `adminDocsController.go`
   - 实现项目相关API接口
   - 注册路由到 `route4api.go`
   - 完善 `docProjects_rep.go` 的查询方法

2. **前端开发**
   - 创建 TypeScript 类型定义
   - 封装 API 调用函数
   - 实现项目列表页面
   - 实现项目编辑表单

3. **测试验证**
   - API 接口测试
   - 前端功能测试
   - 数据一致性验证

### 第二阶段：版本管理模块
1. **后端开发**
   - 扩展 `adminDocsController.go`
   - 实现版本相关API接口
   - 完善 `docVersions_rep.go` 的查询方法
   - 实现目录结构管理逻辑

2. **前端开发**
   - 扩展 TypeScript 类型定义
   - 封装版本管理 API
   - 实现版本列表页面
   - 实现目录结构编辑器

3. **测试验证**
   - 版本管理功能测试
   - 目录结构验证
   - 默认版本设置测试

### 第三阶段：内容管理模块
1. **后端开发**
   - 扩展 `adminDocsController.go`
   - 实现内容相关API接口
   - 完善 `docContents_rep.go` 的查询方法
   - 集成 Markdown 渲染器

2. **前端开发**
   - 集成 Monaco Editor
   - 实现 Markdown 编辑器
   - 实现实时预览功能
   - 实现自动保存机制

3. **测试验证**
   - 编辑器功能测试
   - 内容保存和发布测试
   - 预览功能验证

### 第四阶段：操作日志模块
1. **后端开发**
   - 实现操作日志记录中间件
   - 扩展 `adminDocsController.go`
   - 完善 `docOperationLogs_rep.go`
   - 实现变更对比功能

2. **前端开发**
   - 实现日志查看页面
   - 实现变更对比界面
   - 实现操作历史时间线

3. **测试验证**
   - 日志记录功能测试
   - 变更对比验证
   - 数据恢复测试

## 技术实现要点

### 1. 权限控制
- 使用现有的权限系统
- 新增文档管理权限：`permission.DocsManager`
- 在路由中添加权限检查中间件

### 2. 数据验证
- 使用 Go 的 validator 包进行数据验证
- 前端使用 Element Plus 的表单验证
- 确保 slug 的唯一性和格式正确性

### 3. 操作日志记录
- 在所有 CUD 操作中自动记录日志
- 记录字段级别的变更
- 支持批量操作的日志记录

### 4. 前端状态管理
- 使用 Pinia 进行状态管理
- 实现乐观更新和错误回滚
- 缓存常用数据减少API调用

### 5. 性能优化
- 实现分页查询
- 使用索引优化数据库查询
- 前端虚拟滚动处理大量数据

### 6. 用户体验
- 实现自动保存草稿
- 提供操作确认对话框
- 显示操作进度和结果反馈

## 文件结构规划

```
app/http/controllers/
├── adminDocsController.go          # 文档管理控制器

resource/src/admin/
├── docs/
│   ├── types.ts                     # TypeScript 类型定义
│   ├── api.ts                       # API 调用封装
│   ├── ProjectManagement.vue        # 项目管理页面
│   ├── VersionManagement.vue        # 版本管理页面
│   ├── ContentEditor.vue            # 内容编辑器
│   ├── OperationLogs.vue            # 操作日志页面
│   └── components/
│       ├── ProjectForm.vue          # 项目表单组件
│       ├── VersionForm.vue          # 版本表单组件
│       ├── DirectoryEditor.vue      # 目录编辑器
│       └── MarkdownEditor.vue       # Markdown编辑器
```

## 开发规范

### 1. 代码规范
- 遵循项目现有的代码风格
- 所有CRUD操作封装在 `*_rep.go` 文件中
- 使用统一的错误处理和响应格式
- 添加必要的代码注释

### 2. API设计规范
- 使用RESTful API设计
- 统一的响应格式
- 合理的HTTP状态码
- 完整的错误信息

### 3. 前端开发规范
- 使用TypeScript确保类型安全
- 组件化开发，提高复用性
- 统一的样式规范
- 完善的错误处理

### 4. 测试规范
- 每个模块开发完成后进行功能测试
- 确保API接口的正确性
- 验证前后端数据一致性
- 进行用户体验测试

## 总结

本设计方案采用渐进式开发模式，将复杂的文档管理系统分解为四个独立的模块，每个模块都可以独立开发和测试。这种方式可以：

1. **降低开发风险** - 每个小模块都可以独立验证
2. **提高开发效率** - 并行开发，快速迭代
3. **保证代码质量** - 及时发现和修复问题
4. **便于维护扩展** - 模块化设计，易于后续扩展

通过这套设计方案，我们可以构建一个功能完整、用户友好、易于维护的后台文档管理系统。