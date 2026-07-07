# GooseForum C 端 UI 设计系统长期规划

本文档记录 GooseForum 公共站点（`resource/src/site`）的长期 UI 统一方案。它不是一次性重构清单，而是后续迭代时的设计方向、判断标准和迁移路线。

## 背景

当前公共站点已经有一组主题变量，例如颜色、圆角、边框、阴影深度等，也已经通过 Tailwind CSS 暴露了一部分 token。但页面实现仍然大量直接书写 Tailwind 原子类，例如 `rounded-md border border-line bg-base-100 px-3 py-2`。

这导致几个问题：

- 变量已经定义，但没有被稳定消费，用户在主题预览里调整后不一定影响真实页面。
- 同类 UI 的视觉语义不统一，例如有的卡片很轻、有的带阴影，有的使用 `rounded-md`，有的使用 `rounded-lg`。
- 页面局部实现看起来都还可以，但细看会感到风格不完全一致。
- theme-preview 有时像演示页面，而不是对真实组件体系的验收页面。

长期目标是建立 GooseForum 自己的 C 端 UI 组件层：借鉴 daisyUI 的 token 和语义组件思想，但不照搬 daisyUI 的完整模型。

当前主题能力已经进入“可配置、可预发布、可发布”的阶段：管理员在主题预览设置页编辑 `gf-light` / `gf-dark`，保存为预发布配置，发布后才影响全站。后续设计系统建设应围绕这个流程继续增强真实组件覆盖，而不是扩展一套脱离真实页面的主题演示。

## 设计目标

- 主题变量必须真实影响公共站点的高频 UI，而不是只影响 theme-preview。
- 公共站点应保持内容优先、安静、可扫描、适合日常使用。
- Tailwind 继续用于布局、响应式、少量局部状态；稳定视觉风格由 `gf-*` 语义类承载。
- 可配置项必须可解释、可预览、可验收。不能长期保留“能配置但用户看不出效果”的变量。
- 逐页迁移，不为了抽象而一次性重写所有页面。

## 非目标

- 不引入完整 daisyUI 运行时或把页面改写成 daisyUI 组件。
- 不把后台管理端和 C 端强行统一成一套视觉组件。
- 不为了减少 class 数量而牺牲页面可读性。
- 不追求装饰性主题能力，例如复杂背景、花哨噪点或营销页式视觉。

## 分层模型

### 1. Theme Tokens

主题 token 是可配置的设计变量。只有满足以下条件的变量才应该暴露给管理员：

- 名称能被非设计系统专家理解。
- 在 theme-preview 中能直接看到效果。
- 在真实 C 端页面中有稳定消费点。
- 明暗主题都有合理默认值。

建议长期保留：

```text
color-base-100
color-base-200
color-base-300
color-base-content
color-icon-muted
color-line
color-primary
color-primary-content
color-secondary
color-secondary-content
color-accent
color-accent-content
color-neutral
color-neutral-content
color-info
color-info-content
color-success
color-success-content
color-warning
color-warning-content
color-error
color-error-content
radius-box
radius-field
radius-selector
border
depth
```

需要谨慎处理：

- `size-field`、`size-selector`：如果要保留，应先定义清楚它们影响的是控件高度、内部 padding，还是控件密度。
- `noise`：如果没有明确的站点背景策略，建议不暴露。噪点容易和内容优先的产品气质冲突。

当前公开 token 数量已经不算少，后续不应为了对齐 daisyUI 而机械增加变量。判断标准应始终是：变量是否有真实消费点、是否能在 preview 中清楚验收、是否能被管理员理解。

### 2. CSS Component Layer

在 `resource/src/styles/components.css` 中建立公共站点的语义类。页面优先使用这些类表达视觉组件，避免每个页面重复拼装视觉规则。

建议起步类：

```text
gf-card
gf-panel
gf-button
gf-button-primary
gf-button-secondary
gf-button-ghost
gf-input
gf-textarea
gf-select
gf-tab
gf-tabs
gf-menu
gf-menu-item
gf-dropdown
gf-badge
gf-pill
gf-topic-row
gf-topic-meta
gf-avatar-stack
gf-empty-state
gf-alert
```

这些类应直接消费主题变量，例如：

```css
.gf-card {
  border: var(--gf-border) solid var(--gf-color-line);
  border-radius: var(--gf-radius-box);
  background: var(--gf-color-base-100);
}

.gf-button {
  border-radius: var(--gf-radius-field);
}

.gf-badge {
  border-radius: var(--gf-radius-selector);
}
```

### 3. Vue Usage Layer

Vue 页面和组件仍然可以使用 Tailwind，但分工应清晰：

- Tailwind 负责布局：`grid`、`flex`、`gap`、`min-w-0`、响应式列数。
- `gf-*` 负责视觉语义：按钮、输入框、卡片、菜单、标签、列表行。
- 页面局部特殊样式必须有产品原因，不应只是“这里手感更好一点”。

## Tailwind 封装边界

`gf-*` 不是为了消灭 Tailwind，而是为了把稳定的视觉语义集中起来。封装太少会导致主题变量无效；封装太多会让 CSS 变成另一套难维护的框架。

建议保留在 Vue 模板里的 Tailwind：

```text
布局：grid, flex, block, hidden, min-w-0, overflow-*
响应式：sm:*, md:*, lg:*, xl:*
尺寸约束：w-*, max-w-*, min-h-*, aspect-*
间距布局：gap-*, space-y-*，仅限页面布局关系
状态编排：group, peer，少量局部交互
```

建议进入 `gf-*` 语义类的样式：

```text
颜色：背景、文字、边框、状态色
圆角：box / field / selector
边框：宽度、颜色、hover/focus 变化
阴影：浮层、卡片 depth
控件尺寸：按钮、输入框、菜单项、tab 的固定高度和 padding
可访问状态：focus-visible, disabled, aria-selected, aria-current
组件内部结构：菜单项、badge、topic row、empty state
```

判断规则：

- 如果一个 class 组合在 3 个以上页面重复出现，并且表达同一种 UI 语义，应提炼为 `gf-*`。
- 如果一个样式需要响应主题变量，应优先进入 `gf-*`。
- 如果一个样式只是当前页面的布局关系，应继续留在模板里。
- 不要创建只有一处使用、且没有状态/主题语义的 `gf-*` 类。
- 不要把复杂组件的业务结构藏进 CSS 类；CSS 类只表达视觉和交互状态。

示例：

```html
<!-- 推荐：布局留在模板，视觉语义交给 gf-card -->
<section class="gf-card grid gap-3 md:grid-cols-2">
  ...
</section>

<!-- 不推荐：每页重复拼装视觉规则 -->
<section class="grid gap-3 rounded-lg border border-line bg-base-100 p-4 shadow-sm md:grid-cols-2">
  ...
</section>
```

### Tailwind 和 `@apply`

可以在 `@layer components` 中用 `@apply` 组合 Tailwind 工具类，但需要节制：

- `@apply` 适合表达稳定组件的基础形态，例如 `.gf-button`。
- 主题相关值优先直接使用 CSS 变量，例如 `border-radius: var(--gf-radius-field)`。
- 不要在 `@apply` 中继续写大量一次性任意值，例如复杂的 `shadow-[...]`。
- 复杂状态应拆成明确类名，例如 `.gf-button-primary`、`.gf-button-ghost`，不要依赖页面到处传一串修饰 class。

## CSS 文件组织

当前公共站点 CSS 已经从单一 `resource.css` 拆分为多个职责文件。这个阶段已经完成，不应在后续路线中重复作为待办。

当前结构：

```text
resource/src/styles/
  resource.css          Tailwind 入口，只负责 import 顺序
  tokens.css            :root / [data-theme] / @theme inline
  base.css              html, body, app root, 全局基础元素
  motion.css            gf-page, gf-menu, gf-modal, reduced-motion
  prose.css             gf-prose, topic/post prose
  components.css        gf-card, gf-button, gf-input, gf-tab, gf-menu 等基础组件
  patterns.css          gf-topic-row, gf-home-topic-toolbar 等产品模式
  utilities.css         gf-scrollbar-none 等少量工具
```

`resource.css` 保持类似下面的 import 顺序：

```css
@import "tailwindcss";
@plugin "@tailwindcss/typography";
@source "...";

@import "./tokens.css";
@import "./base.css";
@import "./motion.css";
@import "./prose.css";
@import "./components.css";
@import "./patterns.css";
@import "./utilities.css";
```

拆分原则：

- `tokens.css` 是主题系统入口，必须保持小而稳定。
- `components.css` 放通用 UI 组件，不放页面专属布局。
- `patterns.css` 放产品级复合模式，例如 topic row、home toolbar。
- `prose.css` 单独维护，因为 topic 正文和 post 正文有独立排版规则。
- 新增 CSS 前先判断属于 token、component、pattern 还是 utility。

后续迁移时不再需要先做“拆文件”步骤，应直接从扩展 `components.css` 的最小基础语义类开始。

## 视觉决策

### Surface 和阴影

当前存在一些模糊地带：有的卡片只有边框，有的有轻阴影，有的区域看起来像卡片但实际是 section。这会让页面细看时产生不一致。

长期建议：

- 常规内容容器使用边框，不使用阴影。
- 浮层使用阴影，例如 dropdown、popover、modal、hover card。
- 可交互或被强调的面板可以使用非常轻的阴影，但必须由 `gf-card` 或 `gf-panel` 的变体统一控制。
- 列表行优先使用分割线和 hover 背景，不做成一张张重卡片。

建议语义：

```text
gf-panel     页面区域，边框为主，无阴影
gf-card      独立内容块，默认轻边框，可选 depth
gf-dropdown  浮层，允许阴影
gf-modal     浮层，允许更明确阴影
gf-topic-row 列表行，不使用卡片阴影
```

### Radius

圆角应由三个语义 token 驱动：

```text
radius-box       card, panel, modal, alert
radius-field     button, input, select, textarea, tab
radius-selector  badge, pill, checkbox, toggle
```

页面里应逐步减少裸写 `rounded-md`、`rounded-lg`，改用 `gf-*` 类或 `rounded-box`、`rounded-field`、`rounded-selector`。

### Density

C 端不是管理后台，但也不是营销页。它应该紧凑、可扫描。

建议：

- topic list 行保持密集，不做大卡片。
- settings、messages、notifications 等操作页可以稍微舒展，但控件尺寸应统一。
- header、sidebar、dropdown 的尺寸应稳定，避免每个功能入口都自己定义一套行高。

如果未来要暴露密度配置，应使用一个明确 token，例如：

```text
density: compact | comfortable
```

而不是直接暴露多个不清晰的 `size-*`。

## Theme Preview 的长期角色

theme-preview 应从“主题编辑页”升级为“真实组件验收页”。

当前实现应保持以下边界：

- 主题预览设置页仅管理员可见。
- 编辑态存入 `pageConfig.siteTheme.prepublish`，不直接污染已发布主题。
- 发布时将 `prepublish` 提升为正式主题配置，并清空预发布配置。
- 不保留历史版本列表；当前模型表达的是“一个待发布版本”，不是版本管理系统。
- 右侧预览应优先模拟真实 C 端组件，不应受后台管理端样式变量影响。

它需要覆盖：

- topic list 行。
- topic card / topic body。
- post stream。
- button、input、textarea、select。
- tab、badge、pill、alert。
- dropdown、menu、modal-like surface。
- empty、loading、disabled、error、success 状态。
- 明暗主题下的正文、弱文本、边框、hover、focus。

规则：

- theme-preview 使用和真实页面相同的 `gf-*` 类。
- 每个公开 token 都要在 preview 中有至少一个清晰样本。
- 如果某个 token 在真实页面没有消费点，就不要在 UI 中暴露。
- 右侧预览不只是“好看”，它应该是回归测试清单。

## 迁移路线

### 阶段一：收敛变量

- 保留当前有效的颜色变量。
- 保留 `radius-box`、`radius-field`、`radius-selector`。
- 明确 `border` 和 `depth` 的消费规则。
- 隐藏或暂缓 `noise`。
- 重新评估 `size-field`、`size-selector`，没有清晰语义前不扩大使用。
- 主题编辑流程保持“预发布 -> 发布”单向模型，不再引入历史回滚数组。

### 阶段二：扩展基础组件 CSS 文件

当前 CSS 职责拆分和独立 `components.css` 已经建立。下一步应继续补齐 `gf-input`、`gf-select`、`gf-alert` 等通用类，只放通用 UI 组件类，不放页面专属模式。

这一阶段继续小步扩展最小类集合，不一次性迁移所有页面。避免“扩展组件层”和“大面积改页面视觉”混在一个 diff 里。

### 阶段三：建立基础语义类

先在 `components.css` 中实现最少但稳定的一组类：

```text
gf-card
gf-panel
gf-button
gf-input
gf-tab
gf-menu
gf-badge
gf-alert
```

每个类必须包含：

- 默认状态。
- hover 或 active 状态。
- disabled 状态，如适用。
- focus-visible 状态，如适用。
- 明暗主题可用性。

### 阶段四：替换公共壳层

优先替换：

- header。
- sidebar。
- mobile drawer。
- user menu。
- language menu。
- flash message。

这些区域跨页面出现，收益最高。

### 阶段五：替换高频页面

优先顺序：

1. Home / Category topic list。
2. Topic detail / post stream。
3. Publish page。
4. Search page。
5. Messages / Notifications。
6. Settings / User page。
7. Links / Sponsors / Drafts。

每替换一页，同步更新 theme-preview 中对应的验收样本。

### 阶段五：清理遗留风格

逐步扫描并减少：

```text
rounded-md
rounded-lg
shadow-[...]
border border-line bg-base-100
bg-base-200 px-* py-* text-sm
```

这些不是禁止使用，而是需要判断是否应该由 `gf-*` 类表达。

## 验收标准

一个页面完成迁移后，应满足：

- 主要按钮、输入框、菜单、卡片使用 `gf-*` 语义类。
- 调整主题颜色后，页面主要层级清晰变化。
- 调整 radius 后，相关控件和卡片真实变化。
- 明暗主题都可读。
- 没有不必要的重阴影或页面级视觉噪音。
- 和 theme-preview 中对应样本一致。

## 已知风格债务

以下问题允许在迁移过程中逐步修复：

- 卡片/面板边界不统一：有的区域像卡片，有的像 section。
- 阴影使用不统一：部分卡片有阴影，部分只有边框。
- 圆角来源不统一：`rounded-md`、`rounded-lg`、`rounded-box` 混用。
- 表单控件没有统一组件层，按钮、输入框、textarea 的视觉细节依赖页面局部 class。
- 列表行、卡片、右侧 rail 的 surface 语义还不够明确。
- theme-preview 与真实页面的组件实现还没有完全共享。

## 工作原则

- 每次只迁移一个区域或一种组件，不做不可 review 的大爆炸式重构。
- 修改真实页面时，同步补充 theme-preview 验收样本。
- 新增公开 token 前，先证明它在真实 UI 中有消费点。
- 删除或隐藏无效配置优先于保留假能力。
- 文档、token、CSS 类、页面实现必须一起演进。
