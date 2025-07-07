# 用户通知系统使用指南

## 概述

用户通知系统是一个自动化的邮件提醒功能，用于向活跃用户发送未读通知的汇总邮件。系统会根据用户的活跃状态、未读通知情况以及上次发送时间来智能判断是否需要发送提醒。

### 🚀 性能优化特性

- **分批处理**：采用分批处理机制，每批处理50个用户，避免内存占用过高
- **增量查询**：使用 `QueryById` 方法进行增量查询，提高大数据量下的处理效率
- **内存友好**：不会一次性加载所有用户数据，适合大型用户群体

## 使用方法

### 基本命令

```bash
# 编译项目
go build -o gooseforum main.go

# 运行用户通知任务
./gooseforum userNotice
```

### 命令输出示例

```
开始执行用户通知任务...

处理第 1-7 个用户...

[1] 处理用户: abandon (ID: 1)
  ❌ 用户最近未活跃，跳过

[2] 处理用户: test (ID: 2)
  ❌ 用户最近未活跃，跳过

[3/7] 处理用户: admin1234 (ID: 3)
  ✅ 准备发送邮件通知
  📧 收件人: admin1234 <admin@example.com>
  📋 主题: GooseForum - 您有新的未读通知
  📊 通知统计:
     - 评论通知: 3条
     - 系统通知: 1条
  📝 邮件内容预览:
     亲爱的 admin1234，
     您在 GooseForum 有 4 条未读通知，请及时查看。
     访问链接: https://forum.example.com/notifications
  ⏰ 发送时间: 2024-01-15 10:30:45
  ✅ 已更新发送记录

=== 任务完成 ===
处理用户总数: 7
发送通知数量: 1
```

### 分批处理说明

- **批次大小**：默认每批处理50个用户
- **处理进度**：显示当前批次的用户范围（如：处理第 1-50 个用户...）
- **内存控制**：避免大量用户数据同时加载到内存中
- **性能优化**：适合处理数万甚至数十万用户的场景

## 筛选条件

系统会按照以下条件筛选需要发送通知的用户：

### 1. 用户活跃度检查
- **条件**：用户在过去7天内有活跃记录
- **数据源**：`user_statistics.last_active_time` 字段
- **逻辑**：如果用户超过7天未活跃，则跳过

### 2. 未读通知检查
- **条件**：用户存在7天内创建的未读通知
- **数据源**：`event_notification` 表
- **逻辑**：查询 `is_read = false` 且 `created_at` 在7天内的通知

### 3. 发送频率控制
- **条件**：距离上次发送邮件超过24小时
- **数据源**：`kv_store` 表，键格式为 `user_notice_last_send:{userId}`
- **逻辑**：防止频繁发送，提升用户体验

## 邮件内容构建

### 通知类型统计
- **评论通知**：文章被评论
- **回复通知**：评论被回复
- **系统通知**：系统公告等
- **关注通知**：被其他用户关注

### 邮件模板
```
主题：GooseForum - 您有新的未读通知

亲爱的 {用户名}，

您在 GooseForum 有 {数量} 条未读通知，请及时查看。

通知详情：
- 评论通知: {数量}条
- 回复通知: {数量}条
- 系统通知: {数量}条
- 关注通知: {数量}条

访问链接: https://forum.example.com/notifications

此致
GooseForum 团队
```

## 数据存储

### KV存储键值设计
- **键格式**：`user_notice_last_send:{userId}`
- **值格式**：RFC3339时间格式 (如: `2024-01-15T10:30:45Z`)
- **过期时间**：30天自动清理

### 示例数据
```
user_notice_last_send:1 -> "2024-01-15T10:30:45Z"
user_notice_last_send:2 -> "2024-01-14T15:20:30Z"
```

## 定时任务配置

### Cron 配置示例
```bash
# 每天上午9点执行
0 9 * * * /path/to/gooseforum userNotice

# 每12小时执行一次
0 */12 * * * /path/to/gooseforum userNotice

# 每周一上午9点执行
0 9 * * 1 /path/to/gooseforum userNotice
```

### systemd 定时器配置

创建服务文件 `/etc/systemd/system/gooseforum-notice.service`：
```ini
[Unit]
Description=GooseForum User Notice Service
After=network.target

[Service]
Type=oneshot
User=www-data
WorkingDirectory=/path/to/gooseforum
ExecStart=/path/to/gooseforum userNotice
```

创建定时器文件 `/etc/systemd/system/gooseforum-notice.timer`：
```ini
[Unit]
Description=Run GooseForum User Notice Daily
Requires=gooseforum-notice.service

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
```

启用定时器：
```bash
sudo systemctl enable gooseforum-notice.timer
sudo systemctl start gooseforum-notice.timer
```

## 监控和日志

### 日志输出
- 处理进度显示
- 用户筛选结果
- 邮件发送状态
- 错误信息记录

### 监控指标
- 处理用户总数
- 发送通知数量
- 执行时间
- 错误率

## 性能考虑

### 已实现的优化

- ✅ **分批处理**：每批处理50个用户，避免内存溢出
- ✅ **增量查询**：使用ID递增的方式进行分页查询
- ✅ **内存友好**：不会一次性加载所有用户数据

### 建议的运行时机

- 建议在用户较少的时间段执行（如凌晨2-4点）
- 可以通过调整筛选条件来控制处理量
- 适合处理大规模用户群体（支持数万用户）

### 可调整参数

```go
batchSize := 50 // 可根据服务器性能调整批次大小
```

## 故障排查

### 常见问题

1. **数据库连接失败**
   - 检查数据库配置
   - 确认数据库服务状态

2. **用户数据异常**
   - 检查用户表数据完整性
   - 确认统计表数据更新

3. **KV存储操作失败**
   - 检查存储空间
   - 确认权限设置

### 调试模式

可以通过修改代码添加更详细的调试信息：

```go
// 在 runUserNotice 函数开头添加
fmt.Printf("调试模式：当前时间 %s\n", time.Now().Format("2006-01-02 15:04:05"))
fmt.Printf("调试模式：7天前时间 %s\n", time.Now().AddDate(0, 0, -7).Format("2006-01-02 15:04:05"))
```

## 扩展功能

### 未来可能的改进

1. **邮件模板系统**
   - 支持HTML邮件模板
   - 多语言支持
   - 个性化内容

2. **通知渠道扩展**
   - 短信通知
   - 推送通知
   - 微信通知

3. **用户偏好设置**
   - 通知频率自定义
   - 通知类型选择
   - 免打扰时间设置

4. **统计分析**
   - 发送成功率统计
   - 用户参与度分析
   - A/B测试支持