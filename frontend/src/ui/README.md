# Native UI 组件说明

本目录提供一套原生 Vue 组件，用于替换 Element UI 组件，视觉变量与主页风格对齐（暖黄主色、圆角、玻璃感面板、轻阴影）。

## 接入方式

在 `main.js` 中注册：

```js
import NativeUI from '@/ui/native-components'
import '@/ui/native-ui.css'

app.use(NativeUI)
```

并通过 Vite alias 将业务里 `import { ElMessage } from 'element-plus'` 自动映射到本地实现：

```js
// vite.config.js
alias: {
  'element-plus': fileURLToPath(new URL('./src/ui/feedback.js', import.meta.url))
}
```

## 已覆盖组件

- 基础：`el-button` `el-input` `el-form` `el-form-item` `el-icon`
- 反馈：`ElMessage` `ElMessageBox.confirm` `ElLoading.service` `v-loading`
- 弹层：`el-dialog` `el-drawer` `el-dropdown` `el-dropdown-menu` `el-dropdown-item`
- 展示：`el-avatar` `el-image` `el-empty` `el-badge` `el-tag` `el-tooltip`
- 数据：`el-table` `el-table-column` `el-pagination`
- 布局/导航：`el-row` `el-col` `el-menu` `el-menu-item` `el-breadcrumb` `el-breadcrumb-item` `el-backtop`
- 表单增强：`el-input-number` `el-select` `el-option` `el-radio-group` `el-radio-button`
- 切换：`el-tabs` `el-tab-pane` `el-collapse-transition`
- 上传：`el-upload`

## 使用示例

### 1. 按钮与消息

```vue
<el-button type="primary" @click="submit">保存</el-button>
```

```js
import { ElMessage } from 'element-plus'

ElMessage.success('保存成功')
ElMessage.error('保存失败')
```

### 2. 确认弹窗

```js
import { ElMessageBox } from 'element-plus'

await ElMessageBox.confirm('确定删除吗？', '提示', {
  confirmButtonText: '删除',
  cancelButtonText: '取消'
})
```

### 3. 加载服务

```js
import { ElLoading } from 'element-plus'

const loading = ElLoading.service({ text: '处理中...' })
// ...
loading.close()
```

### 4. 表格

```vue
<el-table :data="rows" stripe>
  <el-table-column prop="name" label="名称" />
  <el-table-column label="操作">
    <template #default="{ row }">
      <el-button link @click="edit(row)">编辑</el-button>
    </template>
  </el-table-column>
</el-table>
```

## 兼容说明

- 为减少业务改动，组件名保持 `el-*` 不变，但底层已完全替换为本地原生实现。
- 现有页面中的交互事件与数据流可继续沿用。
- 如需进一步增强（例如分页器数字页码、表格排序、上传进度），可在当前组件基础上增量扩展。
