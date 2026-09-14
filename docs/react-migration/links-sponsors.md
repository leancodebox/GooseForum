# Links 与 Sponsors React 迁移记录

## 保留的功能契约

- 统一主站顶栏：品牌、配置化头部链接、搜索、主题、语言与登录/用户入口。
- 桌面左侧导航：主题排序、目录、登录态入口、资源、自定义分组、分类和页脚。
- 移动端使用带焦点管理的 Sheet 展示完整导航，并提供本地化标题与关闭按钮。
- 登录态保留未读提示、30 秒轮询、发布、设置、访问组、主题预览和管理后台入口。
- Links 保留总数、分组、外链安全属性、Logo fallback、空状态、申请入口和收录原则。
- Sponsors 保留内容标题、等级分组与密度、头像、可选外链、默认留言、联系方式、展示规则和空状态。

## 响应式与无障碍

- 主站模式边界统一为 `lg`：小于 1024px 完整采用无侧栏、无主区横向 padding 的 mobile/tablet 模式；达到 1024px 后一次性进入桌面布局。
- `lg` 以下隐藏桌面侧栏，使用移动菜单；Links 在移动端保持两列卡片。
- `xl` 起页面主体与 260px 辅助栏并排，较窄屏幕按文档流顺序排列。
- 页面主标题使用 `h1`，分组和辅助卡片使用 `h2`。
- Sheet 含必需的 title/description；外链保留 `target="_blank"` 与 `rel="noopener noreferrer"`。
- 菜单、主题、语言和搜索使用可访问名称；Avatar 始终提供 fallback。

## 有意差异

- 使用 React 官方 shadcn/ui 的 Sheet、DropdownMenu、Avatar、Badge、Button 和 Empty 组合，不复刻 Vue/Reka DOM。
- 保留旧版信息密度与列宽，但统一使用 React shadcn 的语义色和 focus ring。
- 移动菜单从旧版异步自定义 Drawer 换为官方 Sheet，键盘焦点约束更明确，内容顺序不变。

## 后续关联工作

- 通知与消息页面迁移时补充未读状态的文档标题联动和缓存策略。
- Topic 页面迁移时再接入动态 header title、tags 和右侧 rail；本次不提前抽象这些高风险行为。
