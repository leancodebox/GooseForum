# Meilisearch 文章搜索使用示例

## 1. 构建索引

### 1.1 执行索引构建命令

```bash
# 只构建 Meilisearch 索引
./GooseForum checkAndRepairData --meilisearch

# 或者使用短参数
./GooseForum checkAndRepairData -m

# 同时执行数据修复和索引构建
./GooseForum checkAndRepairData --repair --meilisearch
```

### 1.2 索引构建过程

构建过程会：
1. 配置索引设置（可搜索字段、可过滤字段、可排序字段）
2. 批量处理文章（每批100篇）
3. 只索引已发布且正常状态的文章
4. 提取优化的搜索内容
5. 显示处理进度和结果统计

## 2. Go 代码中使用搜索

### 2.1 基础搜索

```go
package main

import (
    "fmt"
    "github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
    "github.com/meilisearch/meilisearch-go"
)

func basicSearch(query string) {
    client := meiliconnect.GetClient()
    index := client.Index("articles")
    
    searchRes, err := index.Search(query, &meilisearch.SearchRequest{
        Filter: "articleStatus = 1 AND processStatus = 0",
        AttributesToHighlight: []string{"title", "description", "searchContent"},
        Limit: 20,
        Offset: 0,
    })
    
    if err != nil {
        fmt.Printf("搜索失败: %v\n", err)
        return
    }
    
    fmt.Printf("找到 %d 篇文章\n", searchRes.EstimatedTotalHits)
    for _, hit := range searchRes.Hits {
        fmt.Printf("标题: %v\n", hit["title"])
        fmt.Printf("描述: %v\n", hit["description"])
        fmt.Println("---")
    }
}
```

### 2.2 按类型搜索

```go
func searchByType(query string, articleType int) {
    client := meiliconnect.GetClient()
    index := client.Index("articles")
    
    filter := fmt.Sprintf("type = %d AND articleStatus = 1 AND processStatus = 0", articleType)
    
    searchRes, err := index.Search(query, &meilisearch.SearchRequest{
        Filter: filter,
        AttributesToHighlight: []string{"title", "description"},
        Sort: []string{"likeCount:desc", "createdAt:desc"},
        Limit: 10,
    })
    
    if err != nil {
        fmt.Printf("搜索失败: %v\n", err)
        return
    }
    
    typeNames := map[int]string{0: "博文", 1: "教程", 2: "问答", 3: "分享"}
    fmt.Printf("在 %s 中找到 %d 篇文章\n", typeNames[articleType], searchRes.EstimatedTotalHits)
}
```

### 2.3 热门文章（无关键词搜索）

```go
func getPopularArticles() {
    client := meiliconnect.GetClient()
    index := client.Index("articles")
    
    searchRes, err := index.Search("", &meilisearch.SearchRequest{
        Filter: "articleStatus = 1 AND processStatus = 0",
        Sort: []string{"viewCount:desc"},
        Limit: 10,
    })
    
    if err != nil {
        fmt.Printf("获取热门文章失败: %v\n", err)
        return
    }
    
    fmt.Println("热门文章:")
    for _, hit := range searchRes.Hits {
        fmt.Printf("标题: %v (浏览量: %v)\n", hit["title"], hit["viewCount"])
    }
}
```

### 2.4 按作者搜索

```go
func searchByAuthor(userId uint64, query string) {
    client := meiliconnect.GetClient()
    index := client.Index("articles")
    
    filter := fmt.Sprintf("userId = %d AND articleStatus = 1 AND processStatus = 0", userId)
    
    searchRes, err := index.Search(query, &meilisearch.SearchRequest{
        Filter: filter,
        AttributesToHighlight: []string{"title", "description"},
        Sort: []string{"createdAt:desc"},
        Limit: 20,
    })
    
    if err != nil {
        fmt.Printf("搜索作者文章失败: %v\n", err)
        return
    }
    
    fmt.Printf("作者 %d 的相关文章: %d 篇\n", userId, searchRes.EstimatedTotalHits)
}
```

### 2.5 时间范围搜索

```go
func searchByTimeRange(query string, startTime, endTime int64) {
    client := meiliconnect.GetClient()
    index := client.Index("articles")
    
    filter := fmt.Sprintf("createdAt >= %d AND createdAt <= %d AND articleStatus = 1 AND processStatus = 0", 
        startTime, endTime)
    
    searchRes, err := index.Search(query, &meilisearch.SearchRequest{
        Filter: filter,
        AttributesToHighlight: []string{"title", "description"},
        Sort: []string{"createdAt:desc"},
        Limit: 20,
    })
    
    if err != nil {
        fmt.Printf("时间范围搜索失败: %v\n", err)
        return
    }
    
    fmt.Printf("时间范围内找到 %d 篇文章\n", searchRes.EstimatedTotalHits)
}
```

## 3. HTTP API 示例

### 3.1 创建搜索控制器

```go
package controllers

import (
    "strconv"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
    "github.com/meilisearch/meilisearch-go"
)

type SearchRequest struct {
    Query      string `json:"query" form:"query"`
    Type       *int8  `json:"type" form:"type"`
    Page       int    `json:"page" form:"page"`
    Limit      int    `json:"limit" form:"limit"`
    SortBy     string `json:"sortBy" form:"sortBy"`
    SortOrder  string `json:"sortOrder" form:"sortOrder"`
}

type SearchResponse struct {
    Hits              []map[string]interface{} `json:"hits"`
    EstimatedTotalHits int64                   `json:"estimatedTotalHits"`
    Page              int                      `json:"page"`
    Limit             int                      `json:"limit"`
    ProcessingTimeMs  int64                    `json:"processingTimeMs"`
}

func ArticleSearch(c *gin.Context) {
    var req SearchRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        c.JSON(400, gin.H{"error": "参数错误"})
        return
    }
    
    // 设置默认值
    if req.Page <= 0 {
        req.Page = 1
    }
    if req.Limit <= 0 || req.Limit > 100 {
        req.Limit = 20
    }
    
    client := meiliconnect.GetClient()
    index := client.Index("articles")
    
    // 构建过滤条件
    filter := "articleStatus = 1 AND processStatus = 0"
    if req.Type != nil {
        filter += " AND type = " + strconv.Itoa(int(*req.Type))
    }
    
    // 构建排序条件
    var sort []string
    if req.SortBy != "" {
        order := "desc"
        if req.SortOrder == "asc" {
            order = "asc"
        }
        sort = []string{req.SortBy + ":" + order}
    }
    
    offset := (req.Page - 1) * req.Limit
    
    searchRes, err := index.Search(req.Query, &meilisearch.SearchRequest{
        Filter: filter,
        AttributesToHighlight: []string{"title", "description", "searchContent"},
        Sort: sort,
        Limit: int64(req.Limit),
        Offset: int64(offset),
    })
    
    if err != nil {
        c.JSON(500, gin.H{"error": "搜索失败: " + err.Error()})
        return
    }
    
    response := SearchResponse{
        Hits:              searchRes.Hits,
        EstimatedTotalHits: searchRes.EstimatedTotalHits,
        Page:              req.Page,
        Limit:             req.Limit,
        ProcessingTimeMs:  searchRes.ProcessingTimeMs,
    }
    
    c.JSON(200, gin.H{"success": true, "data": response})
}
```

### 3.2 路由配置

```go
// 在路由文件中添加
api.GET("/search/articles", ArticleSearch)
```

### 3.3 前端调用示例

```javascript
// 基础搜索
fetch('/api/search/articles?query=Go语言&page=1&limit=20')
  .then(response => response.json())
  .then(data => {
    console.log('搜索结果:', data.data.hits);
    console.log('总数:', data.data.estimatedTotalHits);
  });

// 按类型搜索教程
fetch('/api/search/articles?query=数据库&type=1&sortBy=likeCount&sortOrder=desc')
  .then(response => response.json())
  .then(data => {
    console.log('教程搜索结果:', data.data.hits);
  });
```

## 4. 索引维护

### 4.1 增量更新单篇文章

```go
func updateArticleIndex(articleId uint64) error {
    client := meiliconnect.GetClient()
    index := client.Index("articles")
    
    article := articles.GetById(articleId)
    if article.Id == 0 {
        return fmt.Errorf("文章不存在")
    }
    
    // 如果文章已发布且正常，更新索引
    if article.ArticleStatus == 1 && article.ProcessStatus == 0 {
        doc := convertToSearchDocument(article)
        _, err := index.AddDocuments([]ArticleSearchDocument{doc})
        return err
    } else {
        // 如果文章状态变为草稿或被封禁，从索引中删除
        _, err := index.DeleteDocument(strconv.FormatUint(articleId, 10))
        return err
    }
}
```

### 4.2 删除文章索引

```go
func deleteArticleIndex(articleId uint64) error {
    client := meiliconnect.GetClient()
    index := client.Index("articles")
    
    _, err := index.DeleteDocument(strconv.FormatUint(articleId, 10))
    return err
}
```

### 4.3 清空并重建索引

```bash
# 清空现有索引并重建
./GooseForum checkAndRepairData --meilisearch
```

## 5. 性能优化建议

1. **搜索字段权重**: 标题权重最高，其次是描述和搜索内容
2. **过滤优化**: 始终包含状态过滤条件
3. **分页控制**: 限制每页最大结果数量
4. **缓存策略**: 对热门搜索词进行缓存
5. **监控指标**: 监控搜索响应时间和索引大小

## 6. 故障排除

### 6.1 常见问题

1. **索引构建失败**: 检查 Meilisearch 服务是否运行
2. **搜索无结果**: 确认文章状态和索引配置
3. **性能问题**: 检查索引大小和搜索复杂度

### 6.2 调试命令

```bash
# 查看索引统计
curl -X GET 'http://localhost:7700/indexes/articles/stats'

# 查看索引设置
curl -X GET 'http://localhost:7700/indexes/articles/settings'
```