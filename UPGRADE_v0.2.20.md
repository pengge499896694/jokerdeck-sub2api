# v0.2.20 二开保留升级（2026-11-11）

## 代码来源

- 线上基线：`14036135730ff39e31a5443c010d8792ae45b0de`（v0.2.19 二开）。
- GitHub 最新正式发布核验：v0.2.20，2026-11-07 发布。
- 上游正式 release commit：`dab829a4e9b7d1cc307189aea8b0319eb29e29d0`。
- 使用合并升级，而非部署现有同名 fork tag，避免丢失上线后的二开。

## 保留内容及合并决策

- 客户端节点、YAML 节点上传、节点/供应商策略及订阅端点。
- 财务页面与月份筛选、充值条款、支付配置、商店链接独立打开。
- 画布、独立优化画布入口、无限画布、网络中转配置、指南、进度展示。
- 生图没有产出不扣费；保留已有 OAuth 同账号重试五次后换号策略。
- 自助渠道监控、闲鱼充值订单、反馈提醒、会话日志与 ThreadID。
- Wire 同时保留二开 cleanup 和新 MCP/Digest/BalanceNotify providers。
- OpenAI 异步计费仍使用提前提取的 ChannelUsageFields，不读取复用 Gin context。
- 保留 VueFlow / SheetJS 0.20.3 及无限画布构建，升级 i18n 与安全依赖。
- MCP 默认关闭，新增迁移 `249_mcp_oauth.sql` 仅增加三张独立表。

## 验证与部署门槛

本地详细日志位于 `.local/releases/v0.2.20-20261111/`（忽略目录，不提交凭据）。

- 前端生产构建和类型检查；395 个文件、3124 项测试通过。
- 后端 unit 全包、默认全包、embed 包、integration 标签编译均通过（integration 仅编译，不宣称全套集成测试已执行）。
- 生图尺寸路由新测试正确包含已有重试次数，未绕过或更改收费语义。
- Go module verify、Wire 生成、现有 bluegreen 测试和现网专用发布脚本离线安全测试均通过。
- 核验全部 309 个既有迁移校验和一致，在隔离限额 PostgreSQL 恢复完整线上备份并演练第 310 个迁移。

## 现网专用发布方式

**不能直接使用默认宿主机 Nginx bluegreen 脚本处理此拓扑。**

现网 Nginx 在 `edge-nginx` 容器中，配置单文件 bind mount；所有配置写入必须原地保持 inode。专用脚本：

- `deploy/releases/v0.2.20-20261111/deploy.py`
- `deploy/releases/v0.2.20-20261111/rollback.py`

服务器 staging：`/opt/sub2api/staging/v0.2.20-20261111/`。
服务器备份：`/opt/sub2api/backups/pre-v0.2.20-20261111/`（数据库、配置、镜像/容器信息、迁移校验和）。

严格按 `start → verify → assets → switch → observe 600 → finalize → retire` 执行：

1. 发布二进制及运行时 resources SHA256 校验；将原镜像 digest 打上独立本地 base tag 并核验 digest 后，离线构建候选（不拉取不明确的同名远端镜像）。
2. 新迁移在事务中执行，锁超时 1 秒、SQL 超时 10 秒；绝不修改旧迁移记录。
3. 新容器克隆原环境/挂载/网络，旧容器仍承接请求。
4. 比较新旧 26 个 API，严格比较稳定二开配置、检查前端资源。检查只读，不进行真实扣费/购买。
5. 保留旧哈希静态资源，避免已打开客户端加载旧 chunk 失败。
6. 11 个 upstream 同时平滑切换，`nginx -t` 后 graceful reload，**不 restart Nginx**。
7. 连续观察 10 分钟公网/源站健康、真实网关状态、502/504、Nginx 上游错误、候选错误日志。失败自动切回健康旧容器，两端仍运行等待排空。
8. 同步 Compose / 部署状态但不 recreate 容器。
9. 只有切流前的 Nginx worker 全部退出才优雅停旧应用，否则保留运行，不强停长连接。保留旧镜像/容器回滚。

## 回滚

服务器执行：`python3 /opt/sub2api/staging/v0.2.20-20261111/rollback.py`。

先恢复旧应用健康，再校验并 reload 原 upstream；候选 worker 排空后再停候选。保留新增 MCP 表，不恢复旧数据库覆盖切流后的新业务数据。

无鉴权 401 探针只验证网关路由可达，不代表真实上游生成成功。升级完成状态、时间窗口与实际网关验收以独立部署记录为准；不承诺永久零 502。
