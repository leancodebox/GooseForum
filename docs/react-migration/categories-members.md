# Categories 与 Members React 迁移记录

## 保留的功能契约

- 两页均完全采用 Go payload 给出的顺序，不在浏览器重新排序或筛选。
- Categories 保留分类颜色、emoji/图片图标、描述 fallback、主题数缩写、URL 和空状态。
- Members 保留头像 fallback、昵称与用户名、加入时间、简介 fallback、声望/主题/回复统计和空状态。
- Members 分页保留服务端 `previousUrl`、`nextUrl`，并输出 `rel="prev"` 与 `rel="next"`。

## 响应式与无障碍

- 主站模式边界统一为 `lg`：小于 1024px 保持边到边目录列表，达到 1024px 后再出现主区 padding、圆角卡片和桌面多列。
- 手机端目录项保持边到边、分隔线列表；桌面端恢复圆角边框和多列网格。
- Categories 在 `sm` 起两列；Members 在 `md` 起两列、`xl` 起三列。
- 页面主标题使用 `h1`，目录项名称使用 `h2`，分页包含可访问名称。
- 分类图片图标为空 alt，避免与分类名称产生重复朗读；成员头像始终提供 fallback。

## 有意差异

- 数字缩写逻辑下沉至 `@gooseforum/client`，Vue 与 React 后续可共同复用。
- 采用 React shadcn 的 Avatar、Badge、Button 和 Empty，不复制 Vue/Reka DOM。
