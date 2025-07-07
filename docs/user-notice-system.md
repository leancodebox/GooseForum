# 用户通知系统设计文档

## 概述

用户通知系统是一个定时任务，用于检查用户的活跃状态和未读通知，并向符合条件的用户发送邮件提醒。

## 功能需求

### 核心功能
1. **用户活跃度检查**：根据 `userStatistics` 表的 `last_active_time` 字段判断用户最后活跃时间
2. **未读通知检查**：检查 `eventNotification` 表中用户是否存在7天内的未读通知
3. **发送频率控制**：通过 `kvstore` 记录上次发送时间，避免频繁发送
4. **邮件内容合并**：综合判断后合并需要提醒的内容
5. **模拟发送**：打印需要发送的邮件内容（不实际发送）

## 数据模型分析

### 1. userStatistics 表
```go
type Entity struct {
    UserId            uint64     `gorm:"primaryKey;column:user_id;autoIncrement;not null;" json:"userId"`
    ArticleCount      uint       `gorm:"column:article_count;type:int unsigned;not null;default:0;" json:"articleCount"`
    ReplyCount        uint       `gorm:"column:reply_count;type:int unsigned;not null;default:0;" json:"replyCount"`
    FollowerCount     uint       `gorm:"column:follower_count;type:int unsigned;not null;default:0;" json:"followerCount"`
    FollowingCount    uint       `gorm:"column:following_count;type:int unsigned;not null;default:0;" json:"followingCount"`
    LikeReceivedCount uint       `gorm:"column:like_received_count;type:int unsigned;not null;default:0;" json:"likeReceivedCount"`
    LikeGivenCount    uint       `gorm:"column:like_given_count;type:int unsigned;not null;default:0;" json:"likeGivenCount"`
    CollectionCount   uint       `gorm:"column:collection_count;type:int unsigned;not null;default:0;" json:"collectionCount"`
    LastActiveTime    *time.Time `gorm:"column:last_active_time;type:datetime;" json:"lastActiveTime"` // 关键字段
    CreatedAt         time.Time  `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`
    UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}
```

### 2. eventNotification 表
```go
type Entity struct {
    Id        uint64              `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
    UserId    uint64              `gorm:"column:user_id;type:bigint;index;" json:"userId"`
    Payload   NotificationPayload `gorm:"column:payload;type:json;" json:"payload"`
    EventType string              `gorm:"column:event_type;type:varchar(50);index;" json:"eventType"`
    IsRead    bool                `gorm:"column:is_read;type:boolean;default:false;index;" json:"isRead"`
    ReadAt    *time.Time          `gorm:"column:read_at;type:timestamp;null;" json:"readAt"`
    CreatedAt time.Time           `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`
    UpdatedAt time.Time           `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}
```

### 3. kvstore 表
```go
type Entity struct {
    Key       string     `gorm:"primaryKey;column:key;not null;default:'';" json:"key"`
    Value     string     `gorm:"column:value;type:text;;" json:"value"`
    TTL       int        `gorm:"column:ttl;not null;default:0" json:"ttl"`
    ExpiresAt *time.Time `gorm:"column:expires_at;" json:"expiresAt"`
    CreatedAt time.Time  `gorm:"column:created_at;index;autoCreateTime;<-:create;" json:"createdAt"`
}
```

## 业务逻辑设计

### 1. 用户筛选条件
- **活跃度检查**：`last_active_time` 在7天内的用户
- **未读通知检查**：存在7天内创建且未读的通知
- **发送频率控制**：距离上次发送邮件超过24小时

### 2. 邮件内容构建
根据用户的未读通知类型和数量，构建个性化的邮件内容：
- 评论通知
- 回复通知
- 系统通知
- 关注通知

### 3. KV存储键设计
- 键格式：`user_notice_last_send:{userId}`
- 值：上次发送时间的时间戳
- TTL：30天（自动清理）

## 实现流程

### 主要步骤

1. **分批获取用户列表**
   - 使用 `users.QueryById(startId, limit)` 分批获取用户
   - 每批处理50个用户，避免内存占用过高
   - 使用增量ID查询，确保不遗漏用户

2. **用户活跃度检查**
   - 检查用户统计信息中的最后活跃时间
   - 查询7天内的未读通知
   - 检查KV存储中的上次发送时间
   - 综合判断是否需要发送通知

3. **构建邮件内容**
4. **模拟发送**（打印内容）
5. **更新发送记录**

### 核心算法（分批处理版本）

```go
func runUserNotice() {
    batchSize := 50
    lastUserId := uint64(0)
    
    // 分批处理用户
    for {
        // 1. 获取一批用户
        userBatch := getUserBatch(lastUserId, batchSize)
        if len(userBatch) == 0 {
            break // 没有更多用户
        }
        
        // 2. 遍历当前批次的用户
        for _, user := range userBatch {
            lastUserId = user.Id
            
            // 3. 检查活跃度
            if !isActiveRecently(user) {
                continue
            }
            
            // 4. 检查未读通知
            notifications := getUnreadNotifications(user)
            if len(notifications) == 0 {
                continue
            }
            
            // 5. 检查发送频率
            if !shouldSend(user) {
                continue
            }
            
            // 6. 发送通知
            sendNotification(user, notifications)
        }
    }
}

// 分批获取用户
func getUserBatch(startId uint64, limit int) []*users.Entity {
    return users.QueryById(startId, limit)
}
```

## 错误处理

- 数据库连接异常
- 用户数据不完整
- KV存储操作失败
- 邮件内容构建异常

## 性能考虑

### 分批处理机制
- ✅ **已实现**：每批处理50个用户，可根据服务器性能调整
- ✅ **增量查询**：使用 `QueryById(startId, limit)` 进行ID递增查询
- ✅ **内存友好**：避免一次性加载所有用户数据
- ✅ **可扩展性**：支持处理数万甚至数十万用户

### 数据库查询优化
- 使用索引优化查询性能（ID主键索引）
- 分批查询减少内存占用
- 合理设置批次大小（默认50）

### 内存管理
- ✅ **已优化**：分批处理避免大量数据同时加载
- ✅ **增量处理**：处理完一批后自动释放内存
- ✅ **可控制**：通过 `batchSize` 参数控制内存使用

### 执行时间控制
- 设置合理的超时时间
- 避免在高峰期执行
- 可配置的执行频率
- 分批处理降低单次执行时间

## 扩展性

- 支持不同类型的通知模板
- 支持用户自定义通知频率
- 支持多种发送渠道（邮件、短信、推送）
- 支持A/B测试