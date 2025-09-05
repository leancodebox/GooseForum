# 未读消息通知

```sql
CREATE TABLE user_unread_counts (
    user_id INTEGER PRIMARY KEY,
    mention_count INTEGER DEFAULT 0,
    reply_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    system_count INTEGER DEFAULT 0,
    pm_count INTEGER DEFAULT 0,
    last_updated INTEGER,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);
```
user_unread_counts 的目的是为了小红点的设计。

# 是否执行check脚本 

需要一个 version 数据存入db，通过version和系统中的版本来判断是否需要执行数据迁移或者修复动作。
