# 无限画布线路与首屏加载优化部署记录

北京时间 2026-10-29 07:07 平滑切流，07:09 完成最终验收。

## 功能

- 保留现有无限画布入口和本地数据存储方式。
- 新增「系统设置 → 无限画布线路优化」菜单。
- 可选择现有活跃代理线路，配置 GitHub、GitHub Raw、GitHub CDN、jsDelivr、unpkg 等允许域名。
- 默认关闭；关闭「开放给所有用户」时仅管理员可测试，普通用户继续走原有路径。
- 服务器中转拒绝非 HTTP(S)、账号密码、未配置域名和未启用状态，避免任意 URL 代理。
- 画布入口和主站入口增加 JS preload；指纹化 JS/CSS 继续 immutable 缓存。
- Nginx 静态资源 location 开启 gzip，API 流式 location 保持 gzip 关闭。

## 线上状态

- 新容器：`sub2api-canvas-network-20261029`，镜像 `local/sub2api:v0.2.19-canvas-network-20261029`，版本 `0.2.19`，重启 0 次。
- 旧容器：`sub2api-commerce-links-20261029` 已优雅停止，restart policy 为 `no`。
- Nginx 平滑 reload，旧 worker 已排空；保留 172 个旧画布资源供已打开页面继续加载。
- 数据库无新增迁移，309 个迁移 checksum 与线上一致。
- 备份：`/opt/sub2api/backups/pre-canvas-network-20261029/`。

## 验证

- 画布线路 service/handler 安全校验、前端管理组件、SettingsView、画布类型检查、完整前端构建通过。
- 候选与现网二开接口响应一致；画布配置默认关闭，普通用户配置可读但中转关闭，未授权中转返回 404。
- 三个公网域名健康检查通过；切流后 Nginx 315 次 200、1 次预期 404，真实网关请求 19 次 200，无 5xx/502。
- 切流后 60 次公网健康检查全部 200。
- 静态 JS 已确认返回 `Content-Encoding: gzip`。
