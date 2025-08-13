# 全局通知函数使用指南

## 概述

全局通知函数提供了一个统一的方式来在整个应用中显示通知消息，支持在Vue组件和普通HTML页面中使用。

## 功能特性

- 🎨 **DaisyUI集成**：完全基于DaisyUI的alert和toast组件
- 🌈 **主题适配**：自动适配DaisyUI主题色彩，无硬编码颜色
- 📍 **灵活定位**：支持6种显示位置，使用DaisyUI的toast定位系统
- ⏰ **自定义时长**：可设置显示时长，支持永久显示
- 🔘 **手动关闭**：可选择是否显示关闭按钮
- 📝 **标题支持**：可选择添加标题和描述文本
- 🎭 **图标控制**：可选择是否显示SVG图标
- 🎪 **批量管理**：支持关闭所有通知
- 🎪 **动画效果**：平滑的显示和隐藏动画
- 📱 **响应式设计**：适配不同屏幕尺寸
- 🔄 **自动堆叠**：多个通知自动垂直排列
- 🌐 **全局可用**：Vue组件和HTML页面均可使用

## 在Vue组件中使用

### 导入方式

```typescript
// 导入通知函数
import { notification, NotificationType } from '@/utils/notification'

// 或者导入默认导出
import notification from '@/utils/notification'
```

### 基础用法

```typescript
// 显示成功消息
notification.success('操作成功！')

// 显示错误消息
notification.error('操作失败，请重试')

// 显示警告消息
notification.warning('请注意检查输入内容')

// 显示信息消息
notification.info('这是一条提示信息')
```

### 高级配置

```typescript
// 自定义配置
notification.show({
  message: '这是一条自定义通知',
  type: NotificationType.SUCCESS,
  duration: 5000, // 5秒后自动关闭
  position: 'top-center', // 顶部居中显示
  showClose: true // 显示关闭按钮
})

// 带标题的通知
notification.show({
  title: '新消息',
  message: '您有一条未读消息',
  type: NotificationType.INFO,
  duration: 5000
})

// 无图标通知
notification.show({
  message: '这是一个简洁的通知',
  type: NotificationType.SUCCESS,
  icon: false
})

// 不自动关闭的通知
const notificationId = notification.show({
  message: '这条通知需要手动关闭',
  type: NotificationType.WARNING,
  duration: 0, // 0表示不自动关闭
  position: 'bottom-right'
})

// 手动关闭特定通知
notification.close(notificationId)

// 关闭所有通知
notification.closeAll()
```

### 在Vue组件中的完整示例

```vue
<template>
  <div>
    <button @click="showSuccess">显示成功消息</button>
    <button @click="showError">显示错误消息</button>
    <button @click="showCustom">显示自定义通知</button>
    <button @click="closeAll">关闭所有通知</button>
  </div>
</template>

<script setup lang="ts">
import { notification, NotificationType } from '@/utils/notification'

const showSuccess = () => {
  notification.success('数据保存成功！')
}

const showError = () => {
  notification.error('网络连接失败，请检查网络设置')
}

const showCustom = () => {
  notification.show({
    message: '这是一条持续显示的重要通知',
    type: NotificationType.WARNING,
    duration: 0,
    position: 'top-center',
    showClose: true
  })
}

const closeAll = () => {
  notification.closeAll()
}
</script>
```

## 在HTML页面中使用

由于通知函数已经挂载到全局 `window` 对象上，可以直接在HTML页面的JavaScript中使用：

```html
<!DOCTYPE html>
<html>
<head>
    <title>通知示例</title>
</head>
<body>
    <button onclick="showSuccess()">显示成功消息</button>
    <button onclick="showError()">显示错误消息</button>
    <button onclick="showCustomNotification()">显示自定义通知</button>
    <button onclick="showAdvancedNotification()">显示带标题通知</button>
    <button onclick="showNoIconNotification()">显示无图标通知</button>
    <button onclick="closeAllNotifications()">关闭所有通知</button>

    <script>
        function showSuccess() {
            window.notification.success('操作成功完成！')
        }

        function showError() {
            window.notification.error('发生了一个错误')
        }

        function showCustomNotification() {
            window.notification.show({
                message: '这是一条自定义通知消息',
                type: window.NotificationType.INFO,
                duration: 4000,
                position: 'bottom-center'
            })
        }

        function showAdvancedNotification() {
            window.notification.show({
                title: '系统通知',
                message: '您的操作已完成',
                type: window.NotificationType.SUCCESS,
                duration: 5000,
                position: 'top-center',
                showClose: true
            })
        }

        function showNoIconNotification() {
            window.notification.show({
                message: '简洁的提示信息',
                type: window.NotificationType.INFO,
                icon: false
            })
        }

        function closeAllNotifications() {
            window.notification.closeAll()
        }
    </script>
</body>
</html>
```

## API 参考

### NotificationType 枚举

```typescript
enum NotificationType {
  SUCCESS = 'success',
  ERROR = 'error', 
  WARNING = 'warning',
  INFO = 'info'
}
```

### NotificationConfig 接口

```typescript
interface NotificationConfig {
  message: string;                    // 通知消息内容
  type?: NotificationType;            // 通知类型，默认为 INFO
  duration?: number;                  // 显示时长（毫秒），0表示不自动关闭，默认3000
  position?: 'top-right' |            // 显示位置，默认为 'top-right'
            'top-left' | 
            'bottom-right' | 
            'bottom-left' | 
            'top-center' | 
            'bottom-center';
  showClose?: boolean;                // 是否显示关闭按钮，默认为 true
  title?: string;                     // 可选标题
  icon?: boolean;                     // 是否显示图标，默认 true
}
```

### 方法列表

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `notification.show(config)` | `NotificationConfig` | `string` | 显示通知，返回通知ID |
| `notification.success(message, duration?)` | `string, number?` | `string` | 显示成功通知 |
| `notification.error(message, duration?)` | `string, number?` | `string` | 显示错误通知 |
| `notification.warning(message, duration?)` | `string, number?` | `string` | 显示警告通知 |
| `notification.info(message, duration?)` | `string, number?` | `string` | 显示信息通知 |
| `notification.close(id)` | `string` | `void` | 关闭指定ID的通知 |
| `notification.closeAll()` | - | `void` | 关闭所有通知 |

## 样式自定义

通知组件完全基于DaisyUI的原生组件进行设计：

### DaisyUI集成
- **Toast容器**：使用DaisyUI的`toast`组件作为容器
- **Alert组件**：每个通知都是一个DaisyUI的`alert`组件
- **主题适配**：自动适配当前DaisyUI主题的色彩方案
- **位置系统**：使用DaisyUI的toast定位类（`toast-top`、`toast-end`等）

### 主要组件结构
```html
<!-- 容器 -->
<div class="toast toast-top toast-end">
  <!-- 通知项 -->
  <div class="alert alert-success">
    <svg>...</svg>
    <div>
      <div class="font-bold">标题</div>
      <div class="text-xs">消息内容</div>
    </div>
  </div>
</div>
```

### 自定义主题
通过DaisyUI的主题系统自定义颜色：

```css
[data-theme="custom"] {
  --su: 84 100% 59%;     /* success */
  --er: 0 91% 71%;       /* error */
  --wa: 54 91% 68%;      /* warning */
  --in: 198 93% 60%;     /* info */
}
```

### 位置配置
- `top-right`: `toast-top toast-end`
- `top-left`: `toast-top toast-start`
- `bottom-right`: `toast-bottom toast-end`
- `bottom-left`: `toast-bottom toast-start`
- `top-center`: `toast-top toast-center`
- `bottom-center`: `toast-bottom toast-center`

## 注意事项

1. 通知函数会自动创建DOM容器，无需手动初始化
2. 通知会自动堆叠显示，新通知会出现在旧通知的下方
3. 建议在生产环境中合理设置通知的显示时长，避免过多通知堆积
4. 通知组件具有最高的z-index值(9999)，确保始终显示在最顶层
5. 通知消息支持换行，长文本会自动换行显示