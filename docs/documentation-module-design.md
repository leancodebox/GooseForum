# 文档教程模块设计方案（简化版）

## 概述

本文档设计一个简洁实用的文档教程模块，采用服务端渲染方式，基于用户提出的三表结构实现。

## 数据库设计

### 设计原则

1. **简洁性**：保持四表结构的简洁性，避免过度设计
2. **实用性**：满足基本的文档管理需求和编辑追踪
3. **扩展性**：预留必要的扩展字段，便于后续功能增加
4. **松耦合**：不使用外键约束，通过应用层维护数据一致性
5. **JSON存储**：使用JSON格式存储目录结构，简化层级管理

### 核心四表设计

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
    origin_content LONGTEXT NOT NULL COMMENT '原始文档文档内容（Markdown格式），一些官方文档的原始版本',
    content LONGTEXT NOT NULL COMMENT '文档内容（Markdown格式）',
    view_count INT DEFAULT 0 COMMENT '浏览次数',
    like_count INT DEFAULT 0 COMMENT '点赞次数',
    status TINYINT DEFAULT 1 COMMENT '状态：1-正常，0-禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_version_slug (version_id, slug),
    INDEX idx_version_parent (version_id, parent_id),
    INDEX idx_status (status)
);
```

### 4. 通用操作日志表 (operation_logs)

```sql
CREATE TABLE doc_operation_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    entity_type VARCHAR(50) NOT NULL COMMENT '实体类型：project, version, content',
    entity_id BIGINT NOT NULL COMMENT '实体ID',
    user_id BIGINT NOT NULL COMMENT '操作用户ID',
    action VARCHAR(50) NOT NULL COMMENT '操作类型：create, update, delete, publish, archive',
    field_name VARCHAR(100) COMMENT '变更字段名',
    old_value LONGTEXT COMMENT '变更前值',
    new_value LONGTEXT COMMENT '变更后值',
    change_summary TEXT COMMENT '变更摘要',
    metadata JSON COMMENT '扩展元数据（如IP、User-Agent等）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_entity (entity_type, entity_id),
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at)
);
```

## 功能特性设计

### 1. 核心功能
- **项目管理**：创建、编辑、删除文档项目
- **版本控制**：支持多版本并存，版本切换
- **内容管理**：支持层级结构的文档内容
- **Markdown支持**：内容使用Markdown格式编写
- **目录导航**：基于JSON存储的层级目录结构
- **操作日志**：记录所有实体的操作历史，支持变更追踪
- **版本历史**：通过操作日志实现内容版本回溯

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
1. 创建四个核心数据表（项目、版本、内容、操作日志）
2. 实现对应的Go模型结构
3. 创建基础的数据访问层（Repository）
4. 编写基础的CRUD操作
5. 实现通用操作日志记录机制

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
5. 实现操作日志查看和管理
6. 添加内容版本对比功能

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
- 移除level字段，通过JSON结构维护层级关系

### 数据一致性
- 不使用外键约束，通过应用层逻辑维护数据一致性
- 实现软删除机制，避免数据丢失
- 操作日志提供数据变更追踪和恢复能力

### 通用操作日志机制
- 记录所有实体（项目、版本、内容）的增删改操作
- 支持字段级别的变更记录和对比
- 使用JSON元数据存储扩展信息（IP、User-Agent等）
- 提供操作审计和安全追踪
- 可扩展到其他业务模块使用

## 总结

这个文档教程模块设计采用了简洁实用的四表结构，能够满足文档管理、展示和编辑追踪的需求。设计重点关注：

1. **简单易用**：四表结构清晰，易于理解和维护
2. **服务端渲染**：符合项目现有架构，SEO友好
3. **松耦合设计**：移除外键约束，通过应用层维护数据一致性
4. **JSON目录结构**：使用JSON存储层级目录，简化层级管理
5. **操作追踪**：通用的操作日志机制，支持版本回溯和审计
6. **版本管理**：支持文档的多版本并存
7. **集成友好**：可以很好地集成到现有的GooseForum项目中

建议按照四个阶段逐步实现，先搭建基础架构，再完善功能特性，确保系统的稳定性和可扩展性。新增的通用操作日志表为系统提供了强大的变更追踪能力，不仅适用于文档模块，还可扩展到其他业务模块，同时松耦合的设计提高了系统的灵活性和可维护性。