# 按需词典

- C 端词典的源数据位于 client 的 `src/i18n/messages/`，后台词典位于 admin 的 `messages/`。英文继承由显式 import/spread 保留，不复制继承词条。
- SDK 的原有 `*Resources` 导出保留兼容用途。React 运行时代码不要静态导入整套资源；语言元数据使用 `@gooseforum/client/i18n/locale`。
- 首屏和路由切换并行准备页面模块与所需词典，再显示目标页面。新增页面时同步更新 `goosePageNamespaces` 或 `adminPageNamespaces`。
- C 端公共 namespace 常驻，其他 namespace 按页面准备；中文回退随对应 namespace 准备。语言切换先准备已使用的 namespace，再修改语言及 cookie。
- 动态导入去重、缓存成功结果，失败可重试。C Provider 使用 `addResourceBundle` 同步已准备词典，不由词典触发 Suspense。
- 单元测试 setup 预加载完整资源以兼容直接渲染组件的 fixtures；生产入口不使用该 setup。loader 测试单独验证空缓存、按需范围和重试。
