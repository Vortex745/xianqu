# 经验教训集 (Lessons Learned)

## 2026-03-01 部署实战
### 1. Vercel Serverless (Go) 文件上传
- **问题**: Vercel 的运行时文件系统是只读的。`r.SaveUploadedFile()` 会报 `EPERM`。
- **对策**: 
  - 第一阶段：在内存中读取 `multipart.File`。
  - 第二阶段：使用 `disintegration/imaging` 压缩/调整尺寸。
  - 第三阶段：返回 Base64 Data URL (e.g. `data:image/jpeg;base64,...`)。
  - 优点：适配无持久存储的 Serverless，无需额外引入 S3/OSS。
  - 缺点：数据库体积增长。

### 2. 前端 BaseURL 与相对路径
- **问题**: `VITE_API_URL=""` 会导致 axios 发送请求到当前 window.location。如果前端部署在 CDN/Edge (Cloudflare Pages)，则会 404。
- **对策**:
  - 必须在 `.env.production` 中明确指定 `VITE_API_URL`。
  - 在 `request.js` 中使用 `import.meta.env.VITE_API_URL || '/'`。
  - 组件内的绝对链接（头像、预览等）必须统一通过 helper (如 `resolveUrl`) 转换，不能硬编码 `127.0.0.1`。

### 3. WebSocket 跨域与 Host
- **问题**: `wsHost = window.location.host` 仅在前后端合一时有效。
- **对策**:
  - 在生产环境，`wsHost` 应该指向 API 域名（如 `api.315279.xyz`）。
  - `const wsHost = apiBase ? apiBase.replace(/^https?:\/\//, '') : window.location.host`。

### 4. 前端表单静默失败
- **问题**: 在 Element Plus 中，给隐藏字段添加 `required: true` 会导致点击提交按钮无反应且不报错。
- **对策**: 
  - 审计 `formRules`，确保所有必填项在 UI 上都是可见且可交互的。
  - 优先处理表单提交逻辑中的 `validate` 回调，打印错误。
