# 主题访问控制设计

> 状态：Issue [#12](https://github.com/leancodebox/GooseForum/issues/12) 的已实现架构契约。

## 结论

第一阶段采用“访问组授予分类能力”的模型，不提前把尚未收敛的“主题可见性”或“主分类”写入数据库：

```text
用户 -> 访问组成员关系 -> 分类能力 -> 主题 -> 帖子
```

规则如下：

1. 用户可以属于多个访问组。
2. 访问组对分类授予 `read/reply/create/manage`，能力逐级包含且取最高值。
3. `everyone`、`registered` 是隐式成员关系的系统组。
4. 第一阶段只有增式授权，不支持 `deny`；没有 grant 即无权限。
5. 公开主题可保留一至三个公开分类；受限主题只能属于一个受限分类。
6. 帖子和主题派生资源继承主题权限，不增加帖子级 ACL。
7. 全局管理员、TopicsManager、现有全局/分类版主继续作为 direct grant，由统一策略服务合并，不机械迁入访问组。
8. 游客访问“登录后可能可读”的资源跳转登录；已登录但无权的用户收到 404，避免泄露资源存在性。

这比通用 ACL 窄，但能完整交付最初的“游客/注册用户可见性”需求，并为自定义群组留出空间。

## 为什么先选择分类权限

Issue 讨论了三种模型：

| 模型 | 优点 | 尚未解决的问题 |
| --- | --- | --- |
| 分类权限 | 符合传统论坛板块和现有分类版主模型；分类列表可以继续共享缓存 | 第一阶段受限主题必须单分类 |
| 每主题可见性 | 跨列表可通过一个索引列过滤，分类保持纯归类 | 引入第二个近似“分类”的概念，也不能单独隐藏/保护分类 |
| 主分类 + 附加分类 | 主分类负责治理，附加分类扩大发现范围 | 可见性并集、只读回复、附加标签移除权和迁移规则都未确认 |

长篇设计稿和维护者随后确认的诉求都保留了分类级隐私。最后提出的“主分类 + 附加分类”还没有得到确认，不能把 issue 的最后一条建议直接变成核心 schema。

因此第一阶段实现已被充分推演的分类模型。未来若接受 `topics.visibility_id` 或 `topics.primary_category_id`，可以作为独立扩展，不需要现在预埋一个含义不确定的字段。

## 权限边界

| 等级 | 能力 |
| --- | --- |
| `read` | 查看分类、主题、帖子及其元数据 |
| `reply` | `read` + 回复 |
| `create` | `reply` + 在分类创建主题 |
| `manage` | `create` + 管理分类内容 |

策略固定按以下顺序解析：

```text
账号状态
-> 全局内容管理旁路
-> 主题发布/审核状态
-> 分类版主 direct grant
-> 访问组分类 grant
-> 作者和具体操作规则
```

- `Admin`：完整旁路。
- `TopicsManager`：全局内容管理旁路，但不因此获得站点、页面和角色管理。
- 全局版主：所有分类 `manage`。
- 分类版主：指定分类 `manage`。
- `RoleManager`：配置组与分类授权，但不因此读取私密内容。
- `UserManager`：维护成员，但不因此读取私密内容。
- 访问组经理：审核本组的入组申请；内容能力仍由组的分类 grant 决定。任意成员增删仍由 `RoleManager` 执行。

## 数据模型

实际实现使用 GORM Entity，并走当前 SQLite/MySQL `AutoMigrate` 与版本化数据迁移。

### `access_groups`

```sql
CREATE TABLE access_groups (
    id          BIGINT PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    system_key  VARCHAR(32) NULL,
    join_mode   VARCHAR(32) NOT NULL DEFAULT 'invite_only',
    status      TINYINT NOT NULL DEFAULT 1,
    created_by  BIGINT NOT NULL DEFAULT 0,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    UNIQUE (system_key)
);
```

`system_key` 为 `everyone`、`registered` 或 null。系统组不可删除，也不为每个用户写成员行。`join_mode` 第一阶段支持 `system/invite_only/application`。

### `access_group_members`

```sql
CREATE TABLE access_group_members (
    id               BIGINT PRIMARY KEY,
    access_group_id  BIGINT NOT NULL,
    user_id          BIGINT NOT NULL,
    member_role      VARCHAR(32) NOT NULL DEFAULT 'member',
    status           TINYINT NOT NULL DEFAULT 1,
    created_by       BIGINT NOT NULL DEFAULT 0,
    created_at       DATETIME NOT NULL,
    updated_at       DATETIME NOT NULL,
    UNIQUE (access_group_id, user_id)
);

CREATE INDEX idx_access_group_members_user_status_group
ON access_group_members (user_id, status, access_group_id);
```

`member_role` 为 `member/manager`；`status` 为 disabled/enabled/pending，其中 pending 用于入组申请。服务层限制每个用户最多 32 个 active 自定义组，数据库关系不写死此上限。

### `category_group_permissions`

```sql
CREATE TABLE category_group_permissions (
    id               BIGINT PRIMARY KEY,
    category_id      BIGINT NOT NULL,
    access_group_id  BIGINT NOT NULL,
    permission_level TINYINT NOT NULL,
    status           TINYINT NOT NULL DEFAULT 1,
    created_at       DATETIME NOT NULL,
    updated_at       DATETIME NOT NULL,
    UNIQUE (category_id, access_group_id)
);

CREATE INDEX idx_category_group_permissions_group_status_category
ON category_group_permissions (access_group_id, status, category_id);
```

迁移为每个现有分类写入：

```text
everyone   -> read
registered -> create
```

以此保持上线前后的默认行为。`everyone` 至少拥有 `read` 的分类是公开分类，否则是受限分类。

组 ID 不复制进 topics、posts、通知、文件或搜索文档；这些资源沿 `resource -> post/topic -> category_ids` 解析权限。

## 主题分类不变量

应用必须在同一事务中保证：

```text
topics.category_ids 去重后为 1..3 个
AND 有效 topic_category_index 与其完全一致
AND topics.main_category_id = category_ids[0]
```

新建、草稿和编辑都在服务端校验。前端选择限制只是交互帮助，不是授权。

**可见性只来自主分类。** `category_ids` 的第一项是主分类，写入 `topics.main_category_id`；其余分类是纯粹的辅助标签，既不扩大也不收窄可见范围。因此：

- 读判定在所有位置都是同一个条件——`topics.main_category_id IN (可读分类)`，作用在单个带索引的列上。列表与详情不可能给出不同答案。
- 受限分类不再限制主题能有几个分类。「选中受限分类就只能单选」这条不变量随之消失，不需要任何写路径、批处理或后续功能去维持它。
- 分类权限变化不会让任何既有主题失效：主分类是主题自己的属性，不随分类权限改写。转换机制、冲突预览、转换策略、10,000 条上限全部不再需要。
- 更换主分类（包括把原有的附加分类调到第一位）才是可见性变更，仍然要求操作者对新旧主分类都拥有 `manage`。因此分类集合的比较是**有序**的。

唯一新增的规则：按附加分类浏览列表时，仍要按主分类可读性过滤——被浏览的标签可读，不代表主题可读。这是 `main_category_id` 语义下唯一一处「标签 ≠ 可见性」的地方。

## 请求权限快照

热点主题列表不得 JOIN 成员表与权限表，也不把组 ID 写进 JWT。

权限服务维护两类小缓存：

1. `userID -> active custom group IDs`，成员变化时精确失效。
2. `groupID -> category capability map`，组授权变化时精确失效。

每个请求补充系统组，把多个 map 按最高等级合并一次，并把不可变快照放进 Gin Context。业务层只依赖：

```go
CanReadCategory(categoryID uint64) bool
CanReplyCategory(categoryID uint64) bool
CanCreateCategory(categoryID uint64) bool
CanManageCategory(categoryID uint64) bool
ReadableCategoryIDs() []uint64
```

不长期缓存 `userID -> merged map`，也不按 user/group 组合缓存主题列表。否则组授权变化需要反向清理全部成员，并产生大量低命中列表缓存 key。

## 查询规则

### 分类页和详情

分类路由先调用 `CanReadCategory`，通过后继续使用现有分类列表 SQL 和 `categoryID + sort + page` 共享缓存。

主题详情先按主键加载主题，在内存中检查最多三个分类 ID，通过后才读取帖子窗口。帖子窗口 SQL 不加 ACL：

```sql
SELECT * FROM posts
WHERE topic_id = ? AND post_no > ?
ORDER BY post_no ASC, id ASC
LIMIT 21;
```

所有从 `topic_id/post_id` 进入的 API 都必须调用同一策略，不能只保护 HTML 控制器。

### 跨分类列表

用户主页、动态、点赞、收藏、搜索等个性化跨分类列表必须在分页 SQL 内过滤：

```sql
AND EXISTS (
    SELECT 1 FROM topic_category_index idx
    WHERE idx.topic_id = topics.id
      AND idx.effective = 1
      AND idx.category_id IN (:readable_category_ids)
)
```

不能先 `LIMIT` 再在 Go 中删除无权项，否则会出现稀疏页、错误游标和摘要泄漏。仓储层接收排序去重后的 ID；空集合直接返回空结果，覆盖全部有效分类时省略 `EXISTS`。

会携带主题/回复预览的 `user_activities` 需要冗余 `topic_id`，使数据库能在游标分页前过滤；历史数据根据 subject 回填。

### 首页发现

第一阶段建议采用“成员首页混排其可读受限主题”（原讨论方案 A），因为此时不需要先开发另一套私密内容更新发现机制：

- 游客和普通注册用户可以继续共享稳定的基础列表缓存。
- 额外拥有受限分类读取权的用户执行带 `EXISTS` 的查询，不建立 audience 列表缓存。
- 分类页鉴权后仍按分类共享缓存。

若以后采用“首页只展示公开主题”，必须同时实现 `category.latest_activity_at`、`user_category.last_seen_at` 和侧边栏更新红点；否则成员无法可靠发现受限内容。

### 搜索、SEO 与 Feed

- 搜索文档保存可过滤的 `categoryIds`，查询使用当前请求的可读分类 ID。
- 爬虫等同游客。
- RSS、Sitemap、JSON-LD、Open Graph、no-js、统计和建议词只输出 `everyone` 可读内容。
- 自定义组授权变化无需重建搜索文档；主题换分类时才更新。

## 写操作

- 创建主题：对全部所选分类拥有 `create`，且满足分类不变量。
- 创建回复：对主题访问分类拥有 `reply`。
- 点赞、收藏、关注、举报：先校验 `read`。
- 编辑、删除：先 `read`，再执行现有作者/管理规则。
- 审核与批量接口：使用同一权限服务，不允许后台控制器旁路。

主题、首帖、`topics.category_ids` 和 `topic_category_index` 必须在一个数据库事务内保存；提交成功后再清缓存、更新搜索/文件引用并发布事件。

## 管理端与发布端

分类管理页由 `Admin/RoleManager` 配置：

- 由 `everyone` grant 推导出的公开/受限状态；
- 各访问组的分类能力；
- 无权用户是否看到申请入口；
- 申请加入哪个组。

访问组页管理组名、加入方式、经理、邀请、申请和成员。分类授权只做只读汇总，避免两处同时编辑同一策略。

发布 Payload 只下发当前用户拥有 `create` 的分类，并附服务端计算的 `isRestricted/allowMultipleCategories`。选择受限分类会清空其他分类并解释可见范围变化。

## 权限变更与迁移

- 入组/退组：只清该用户成员缓存，不改主题或搜索文档。
- 自定义组授权变化：只清该组 capability map，不改主题或搜索文档。
- 删除 `everyone -> read` 就是一次普通的权限写入：可见性从 `main_category_id` 读取，没有任何主题行需要改写，因此不需要冲突预览、转换策略和条数上限，也不存在改到一半的中间状态。
- 同理，搜索文档不需要因为权限变化而重建：文档里的 `mainCategory` 不会因为分类权限改变而失效。
- Role、版主和账号状态变化必须立即失效相关缓存。

## 全链路边界

功能只有覆盖以下入口才算完成：

- 首页所有排序、分类页/瀑布流、主题详情、帖子窗口；
- 用户主题、回复、点赞、收藏、关注、动态预览；
- 搜索、建议、通知、邮件、Webhook；
- RSS、Sitemap、JSON-LD、Open Graph、no-js、统计、榜单；
- 创建、回复、编辑、删除、互动、举报、审核 API；
- 图片与附件。

当前 `/file/img/*filename` 是公开路由。开放受限分类前必须明确选择：

1. 页面受保护，但已知文件直链仍可分享，并在产品中说明此限制；或
2. 帖子资源通过 `file_usage -> post -> topic` 鉴权或短时签名 URL 访问，头像和站点资源继续公开。

若目标包括公司内部论坛，方案 2 是上线条件，不是后续优化。

## 渐进交付

### Slice 0：契约与基线（完成）

- 评审本文并确定附件策略。
- 针对 10 万主题、95%/50%/5% 可见分布记录 SQLite P95 和缓存基线；MySQL 基线由部署环境的集成测试继续记录。

### Slice 1：增量数据与策略核心（完成）

- 增加三个模型和迁移。
- 写入系统组并为现有分类回填兼容 grant。
- 增加仓储、两类元数据缓存、请求快照和策略测试。
- 暂不开放受限分类配置。

### Slice 2：覆盖全部读取入口（完成）

- 保护详情、窗口和跨分类列表。
- 增加并回填 activity `topic_id`。
- 覆盖搜索、SEO、通知、Feed 和统计。
- 所有分类仍保持公开，先跑兼容与回归测试。

### Slice 3：写入和发布 UX（完成）

- 主题/首帖/分类持久化事务化。
- 覆盖所有写 API。
- 过滤发布 Payload 并实现受限分类选择交互。

### Slice 4：管理与启用（完成）

- 增加访问组、成员和分类授权管理。
- 增加邀请/申请流程。
- 增加公开转受限的影响预览与批量转换。
- 图片通过 `file_usage -> post/topic -> category` 鉴权；新上传图片在被正文认领前仅上传者可读。

性能基线与复现命令见 [访问控制性能基线](topic-access-control-performance.md)。

## 验收要求

自动化测试至少覆盖游客、普通注册用户、自定义组成员、分类版主、全局版主、TopicsManager、RoleManager、Admin，并组合以下场景：

- 公开、仅注册用户、自定义组分类；
- 直接 URL、所有列表、搜索、Feed；
- `read/reply/create/manage` 与无关后台权限隔离；
- 入退组和授权调整后的缓存实时性；
- 公开/受限转换及转换失败恢复；
- SQLite/MySQL schema 与索引兼容。

禁止列表先分页再在 Go 中删无权数据；禁止对每个主题单独查询一次权限。

## 明确延后

第一阶段 schema 不表示以下概念：

- 独立于分类的主题可见性组；
- 扩大可见范围的附加分类（附加分类**不影响**可见性，与该方案不同）；
- 显式 `deny`；
- 帖子级 ACL；
- audience 精确计数。

它们需要独立用户故事和已确认的冲突规则。延后能保持首轮迁移纯增量，也避免把 issue 尚未回答的最后一条评论写死进数据库。
