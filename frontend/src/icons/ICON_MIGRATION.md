# Icon Migration（Element Icons -> Tailwind 兼容 SVG）

## 1) 问题定位

- 影响范围：`frontend/src` 中共 `21` 个文件直接从 `@element-plus/icons-vue` 导入图标。
- 受影响图标：共 `48` 个命名图标（见下方映射表）。
- 现象：编译通常不报错，但运行时出现图标空白、尺寸异常、旋转类不生效（`is-loading`）等视觉问题。
- 根因：项目已去 Element Plus 组件运行时，但图标仍依赖 Element 图标实现与样式语义，导致渲染链不一致。

## 2) 方案选型

选型：**Lucide 风格原生 SVG 组件（Vue）**  
实现方式：本地适配层 `src/icons/tw-icons.js`，并通过 Vite alias 替换 `@element-plus/icons-vue`。

理由：

1. SVG 纯组件，天然兼容 Tailwind 的 `w-* / h-* / text-*`（`currentColor`）。
2. 无额外运行时依赖，避免三方库升级带来的 API 变化。
3. 支持保留原有业务代码导入写法，迁移成本最低。
4. 和现有“去 Element 组件化”架构一致。

## 3) 新旧图标映射表

| Element 图标 | 新图标（Lucide 风格） | 备注 |
|---|---|---|
| Aim | Aim | 同语义 |
| ArrowLeft | ArrowLeft | 同语义 |
| ArrowRight | ArrowRight | 同语义 |
| Bell | Bell | 同语义 |
| Calendar | Calendar | 同语义 |
| Camera | Camera | 同语义 |
| CaretBottom | CaretBottom | 同语义 |
| CaretTop | CaretTop | 同语义 |
| ChatDotRound | ChatDotRound | 近似（聊天气泡） |
| Check | Check | 同语义 |
| CircleCheck | CircleCheck | 近似（圈选） |
| CircleCloseFilled | CircleCloseFilled | 近似（实心关闭） |
| Close | Close | 同语义 |
| CloseBold | CloseBold | 近似（加粗关闭） |
| Delete | Delete | 近似（垃圾桶） |
| Edit | Edit | 同语义 |
| Filter | Filter | 同语义 |
| Goods | Goods | 近似（包裹） |
| HomeFilled | HomeFilled | 近似（实心首页） |
| Iphone | Iphone | 近似（手机） |
| List | List | 同语义 |
| Loading | Loading | 旋转加载 |
| Location | Location | 同语义 |
| LocationInformation | LocationInformation | 近似（定位+信息） |
| Lock | Lock | 同语义 |
| Money | Money | 近似（币种） |
| MoreFilled | MoreFilled | 三点 |
| Odometer | Odometer | 近似（仪表盘） |
| Operation | Operation | 近似（操作/滑杆） |
| Picture | Picture | 同语义 |
| Plus | Plus | 同语义 |
| Promotion | Promotion | 近似（发送） |
| Refresh | Refresh | 同语义 |
| Right | Right | 同语义（尖角右） |
| Scissor | Scissor | 同语义 |
| Search | Search | 同语义 |
| Select | Select | 同语义（勾） |
| ShoppingCart | ShoppingCart | 同语义 |
| Star | Star | 同语义 |
| StarFilled | StarFilled | 同语义（实心） |
| Switch | Switch | 近似（切换） |
| SwitchButton | SwitchButton | 同语义（电源） |
| Top | Top | 近似（向上） |
| User | User | 同语义 |
| Van | Van | 同语义 |
| View | View | 同语义（眼睛） |
| Wallet | Wallet | 同语义 |
| ZoomIn | ZoomIn | 同语义 |

## 4) 替换后代码示例

### 示例 A：业务文件保持原写法（推荐）

```vue
<template>
  <el-icon class="text-gray-500">
    <Search class="w-4 h-4" />
  </el-icon>
</template>

<script setup>
import { Search } from '@element-plus/icons-vue'
</script>
```

说明：通过 alias，这里实际加载的是 `src/icons/tw-icons.js`。

### 示例 B：直接使用本地图标模块

```vue
<template>
  <Search class="w-5 h-5 text-amber-500" />
</template>

<script setup>
import { Search } from '@/icons/tw-icons'
</script>
```

### 尺寸/颜色逻辑

- 默认尺寸：`1em`（跟随字体）。
- 颜色：`currentColor`（跟随 `text-*` 或父级 `color`）。
- 支持：`w-*` / `h-*` / `text-*`（Tailwind 工具类）。
- 旋转：`.el-icon.is-loading` 对 `Loading` 生效。

## 5) 批量验证清单

1. 构建检查：
   - `npm run build`
2. 导入检查：
   - `rg -n \"from '@element-plus/icons-vue'\" frontend/src -S`
   - 结果应全部可编译且无丢失命名导出。
3. 视觉冒烟页：
   - 首页、商品详情、购物车、消息、发布页、后台四页（仪表盘/用户/商品/订单）。
4. 状态类检查：
   - `is-loading` 旋转是否生效。
   - 深色/浅色背景下图标对比度是否正常。
5. Tailwind 兼容检查（若项目已启用 Tailwind）：
   - 在任一图标组件上加 `w-6 h-6 text-red-500` 验证尺寸与颜色。

