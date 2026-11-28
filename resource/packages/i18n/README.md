# 前端国际化

主站和管理后台共用一个词条目录。中文是原文，词条按业务放入 site、admin 或 shared 命名空间，ID 形如 `site.topic:replyCount`。原文、翻译上下文、各语言译文及对应的版本记录跟随同一个 key。所有写入通过 resource/ 下的 pnpm i18n 命令完成。

## 日常操作

1. `pnpm i18n add site.topic:replyCount --source '{count, number} 条回复' --context '主题页上的回复总数'`
2. 主站调用 `useTranslation("site.topic").t("replyCount", { count: 3 })`。后台使用 `adminText(locale, namespace)` 获取带类型的翻译函数。路由声明自己需要的命名空间。
3. `pnpm i18n translate site.topic:replyCount --locale en --text '{count, plural, one {# reply} other {# replies}}'` 保存草稿。
4. `pnpm i18n status site.topic:replyCount --locale en` 查看草稿正文和 reviewRevision。核对后运行 `pnpm i18n review site.topic:replyCount --locale en --revision <reviewRevision>`。
5. `pnpm i18n update site.topic:replyCount --source '共有 {count, number} 条回复'` 修改原文。原文或上下文变化时，已有译文保留在目录并自动进入 stale 状态。
6. 删除或重命名前先更新调用处。`pnpm i18n remove <id>` 和 `pnpm i18n rename <id> <new-id>` 用候选目录生成类型，再通过代码类型和 API 消息码检查验证变更。重命名保留翻译及审核状态。

也可以直接运行 `pnpm i18n --help` 查看参数。源码里的业务 key 可由调用方维护，翻译目录和生成文件统一通过上述命令写入。

## 批量翻译

运行 `pnpm i18n translate --locale ja`，从标准输出获取该语言缺失和过期的条目。交给译者或翻译工具后，通过 `pnpm i18n translate --import <file>` 导入 JSON 数组。每条包含 id、locale、sourceRevision 和 text。源版本变化后旧批次会被拒绝，需重新导出。导入会校验整批消息，然后保存草稿供审核。

需要撤下某种语言的译文时，运行 `pnpm i18n translate <id> --locale ja --clear`。

审核者可用 `pnpm i18n review --import <file>` 一次提交多条审核结果。每条包含 id、locale 和 reviewRevision，签的是当前草稿正文和对应原文版本。

## 生成、回退与检查

- 生成器根据原文和上下文计算版本，导出每个语言、每个命名空间的 JSON 和静态可分析的动态导入表。i18next 官方 CLI 生成资源类型，TypeScript 检查业务调用的 key 和 ICU 参数。构建、类型检查和测试会先生成，开发服务会监听目录更新。复杂 ICU 消息会随词典加载格式化模块。generated/ 属于本地构建产物。
- 运行时加载当前路由及公共命名空间，缓存已完成的加载并允许重试。译文要经过审核且与当前原文版本一致才进入目标语言词典。日语和意大利语优先回退到英语，其余缺译回退到中文。
- `pnpm check:i18n` 检查目录格式、完整性、ICU 语法和参数，以及前端 API 消息码与服务端契约的一致性。当前完整发布语言为中文、英文和俄文。迁移时导入各语言已有的译文，原先继承英文的条目计为缺译。`pnpm i18n:check --strict` 用于要求所有五种语言完整的检查。
