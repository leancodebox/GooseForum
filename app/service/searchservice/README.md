# 搜索服务 (Search Service)

这个包提供了基于 Meilisearch 的文章搜索功能，包括索引构建和搜索查询。

## 功能特性

### 核心特性

- **高效搜索**: 直接从Meilisearch返回结果，无需额外数据库查询
- **全文搜索**: 利用Meilisearch的强大全文搜索能力
- **分类过滤**: 支持按文章分类筛选
- **分页支持**: 支持offset和limit参数
- **参数验证**: 完整的请求参数验证
- **错误处理**: 统一的错误处理机制

### 1. 搜索服务 (searchservice.go)

- **SearchArticles**: 通过关键词和分类搜索文章
  - 高效查询：直接从Meilisearch搜索结果返回ID和标题，无需额外数据库查询
  - 支持分页和分类过滤

#### 使用示例

```go
req := searchservice.SearchRequest{
    Query:      "Go语言",
    Categories: []uint64{1, 2}, // 可选的分类ID列表
    Limit:      20,
    Offset:     0,
}

result, err := searchservice.SearchArticles(req)
if err != nil {
    // 处理错误
}

// 使用搜索结果
for _, article := range result.Results {
    fmt.Printf("文章ID: %d, 标题: %s\n", article.ID, article.Title)
}
```

### 2. 索引服务 (indexservice.go)

- **BuildMeilisearchIndex**: 构建 Meilisearch 索引
  - 批量处理文章数据
  - 只索引已发布且正常状态的文章
  - 自动配置索引设置

#### 使用示例

```go
result, err := searchservice.BuildMeilisearchIndex()
if err != nil {
    log.Printf("构建索引失败: %v", err)
    return
}

log.Printf("索引构建完成，处理了 %d 篇文章", result.ProcessedCount)
```

## API 接口

### POST /api/forum/search-articles

搜索文章接口

#### 请求参数

```json
{
    "query": "搜索关键词",
    "categories": [1, 2, 3],  // 可选，分类ID列表
    "page": 1,                // 可选，页码，默认1
    "pageSize": 20            // 可选，每页数量，默认20，最大100
}
```

#### 响应格式

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "results": [
            {
                "id": 123,
                "title": "文章标题"
            }
        ],
        "total": 100,
        "page": 1,
        "pageSize": 20,
        "query": "搜索关键词",
        "categories": [1, 2, 3]
    }
}
```

## 命令行工具

原有的数据修复命令已经重构，现在使用新的搜索服务：

```bash
# 构建 Meilisearch 索引
./gooseforum checkAndRepairData --meilisearch

# 修复数据
./gooseforum checkAndRepairData --repair

# 同时执行两个操作
./gooseforum checkAndRepairData --meilisearch --repair
```

## 索引配置

### 可搜索字段
- `title`: 文章标题（权重最高）
- `searchContent`: 优化后的搜索内容

### 可过滤字段
- `type`: 文章类型
- `userId`: 用户ID
- `category`: 分类ID

### 可排序字段
- `createdAt`: 创建时间
- `updatedAt`: 更新时间

### 显示字段
- `id`: 文章ID
- `title`: 文章标题

## 注意事项

1. 搜索功能依赖 Meilisearch 服务，确保服务正常运行
2. 索引构建可能需要一些时间，取决于文章数量
3. 只有已发布且正常状态的文章会被索引
4. 搜索结果按相关性排序
5. 支持中文分词和模糊搜索

## 测试

运行测试：

```bash
go test ./app/service/searchservice/
```

测试覆盖了基础的数据结构和工具函数。