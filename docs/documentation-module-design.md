# 文档教程模块设计方案（简化版）

## 概述

本文档设计一个简洁实用的文档教程模块，采用服务端渲染方式，基于用户提出的三表结构实现。

## 数据库设计

### 设计原则

1. **简洁性**：保持三表结构的简洁性，避免过度设计
2. **实用性**：满足基本的文档管理需求
3. **扩展性**：预留必要的扩展字段，便于后续功能增加

### 核心三表设计

### 1. 文档项目表 (doc_projects)

```sql
CREATE TABLE doc_projects (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL COMMENT '项目名称',
    slug VARCHAR(255) UNIQUE NOT NULL COMMENT 'URL友好标识',
    description TEXT COMMENT '项目描述',
    logo_url VARCHAR(500) COMMENT '项目Logo',
    sort_order INT DEFAULT 0 COMMENT '排序权重',
    status TINYINT DEFAULT 1 COMMENT '状态：1-正常，0-禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_slug (slug),
    INDEX idx_status (status)
);
```

### 2. 文档版本表 (doc_versions)

```sql
CREATE TABLE doc_versions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT NOT NULL COMMENT '所属项目ID',
    version VARCHAR(100) NOT NULL COMMENT '版本号',
    name VARCHAR(255) NOT NULL COMMENT '版本名称',
    description TEXT COMMENT '版本描述',
    directory TEXT COMMENT '目录结构（JSON格式）',
    is_default BOOLEAN DEFAULT FALSE COMMENT '是否为默认版本',
    sort_order INT DEFAULT 0 COMMENT '排序权重',
    status TINYINT DEFAULT 1 COMMENT '状态：1-正常，0-禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES doc_projects(id) ON DELETE CASCADE,
    UNIQUE KEY uk_project_version (project_id, version),
    INDEX idx_project_status (project_id, status)
);
```

### 3. 文档内容表 (doc_contents)

```sql
CREATE TABLE doc_contents (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    version_id BIGINT NOT NULL COMMENT '所属版本ID',
    title VARCHAR(255) NOT NULL COMMENT '文档标题',
    slug VARCHAR(255) NOT NULL COMMENT 'URL友好标识',
    content LONGTEXT NOT NULL COMMENT '文档内容（Markdown格式）',
    sort_order INT DEFAULT 0 COMMENT '排序权重',
    parent_id BIGINT DEFAULT 0 COMMENT '父文档ID（支持层级结构）',
    level INT DEFAULT 1 COMMENT '层级深度',
    view_count INT DEFAULT 0 COMMENT '浏览次数',
    status TINYINT DEFAULT 1 COMMENT '状态：1-正常，0-禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (version_id) REFERENCES doc_versions(id) ON DELETE CASCADE,
    UNIQUE KEY uk_version_slug (version_id, slug),
    INDEX idx_version_parent (version_id, parent_id),
    INDEX idx_status (status)
);
```

## 功能特性设计

### 1. 核心功能
- **项目管理**：创建、编辑、删除文档项目
- **版本控制**：支持多版本并存，版本切换
- **内容管理**：支持层级结构的文档内容
- **Markdown支持**：内容使用Markdown格式编写
- **目录导航**：基于层级结构的侧边栏导航

### 2. 页面路由设计（服务端渲染）

```
# 文档首页
GET /docs                           # 文档项目列表页

# 项目相关页面
GET /docs/{project_slug}            # 项目首页（默认版本）
GET /docs/{project_slug}/{version}  # 指定版本首页

# 文档内容页面
GET /docs/{project_slug}/{version}/{content_slug}  # 具体文档页面

# 管理页面（需要权限）
GET /admin/docs                     # 文档管理首页
GET /admin/docs/projects            # 项目管理
GET /admin/docs/projects/create     # 创建项目
GET /admin/docs/projects/{id}/edit  # 编辑项目
GET /admin/docs/versions            # 版本管理
GET /admin/docs/contents            # 内容管理
GET /admin/docs/contents/create     # 创建内容
GET /admin/docs/contents/{id}/edit  # 编辑内容
```

## 实现计划

### 第一阶段：数据库和基础模型
1. 创建三个核心数据表
2. 实现对应的Go模型结构
3. 创建基础的数据访问层（Repository）
4. 编写基础的CRUD操作

### 第二阶段：控制器和路由
1. 实现文档相关的控制器
2. 配置路由规则
3. 实现基础的页面渲染逻辑
4. 创建简单的HTML模板

### 第三阶段：前端页面
1. 设计文档展示页面
2. 实现Markdown内容渲染
3. 添加侧边栏目录导航
4. 实现版本切换功能

### 第四阶段：管理功能
1. 实现文档管理后台
2. 添加内容编辑功能
3. 实现项目和版本管理
4. 添加权限控制

## 技术实现要点

### 当前项目集成
- **框架**：基于现有的Go项目结构
- **数据库**：使用项目现有的数据库连接
- **模板引擎**：使用项目现有的HTML模板系统
- **路由**：集成到现有的路由配置中

### Markdown处理
- 使用Go的Markdown解析库（如blackfriday）
- 支持代码高亮和表格渲染
- 实现目录自动生成

### 目录结构
- 使用JSON格式存储层级目录
- 支持拖拽排序（前端实现）
- 自动生成面包屑导航

## 总结

这个文档教程模块设计采用了简洁实用的三表结构，能够满足基本的文档管理和展示需求。设计重点关注：

1. **简单易用**：三表结构清晰，易于理解和维护
2. **服务端渲染**：符合项目现有架构，SEO友好
3. **层级支持**：通过parent_id和directory字段支持目录结构
4. **版本管理**：支持文档的多版本并存
5. **集成友好**：可以很好地集成到现有的GooseForum项目中

建议按照四个阶段逐步实现，先搭建基础架构，再完善功能特性，确保系统的稳定性和可扩展性。