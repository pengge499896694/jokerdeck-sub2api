# 无限画布线路优化

无限画布仍保留原有入口和本地存储方式。新增的线路优化只负责外部素材读取：
参照图、提示词源、插件清单/源码和生成结果下载在启用后可通过服务器中转。

后台入口为「系统设置 → 无限画布线路优化」。该入口位于管理员设置中，当前默认关闭；
选择现有活跃代理线路并保存后，管理员可以先测试。关闭「开放给所有用户」时，普通用户
继续使用原有直连路径；确认线路稳定后再打开该开关。

服务器中转只允许配置域名及其子域名，默认包括 GitHub、GitHub Raw、GitHub CDN、
jsDelivr 和 unpkg。请求必须是 HTTP(S)，不允许账号密码和未配置域名，因此不会成为
任意 URL 的 SSRF 代理。代理线路只在服务器端使用，账号密码不会返回浏览器。

首次加载速度同时做了两点优化：主站和画布入口响应会通过 `Link: rel=preload` 提前
拉取入口 JS，带 hash 的 JS/CSS 继续使用一年 immutable 缓存。当前 Nginx 关闭了 gzip，
部署时应在静态资源 location 单独开启 gzip（API 流式 location 继续关闭），以明显减少
首次 JS 传输体积。

接口：

- 登录用户：`GET /api/v1/canvas/network-config`。
- 画布素材中转：`GET /api/v1/canvas/network/fetch?url=...`。
- 管理员：`GET/PUT /api/v1/admin/settings/canvas-network`。

该设置使用现有 settings 表的 JSON 值，无新增数据库迁移。
