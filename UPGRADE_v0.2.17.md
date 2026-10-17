# v0.2.17 升级 & 无感部署说明

## 2026-10-17 16:35 最终上线记录（覆盖下方早期回滚记录）

完整修正版现已上线。发现旧二开的客户端节点后台与订阅合并逻辑未合入，已从
`v0216-docs` 工作区恢复原有功能，替换此前仅支持静态 YAML 的不完整实现。
保留原 `client_node_global_policy` 数据格式、主备订阅、禁用节点过滤、订阅合并、
分组列表和 TCP 测速，以及系统设置中的客户端节点面板。

- 镜像：`local/sub2api:v0.2.17-full-20261017`
- 活动容器：`sub2api-v0217-full-20261017`，healthy，检查时零重启。
- nginx 容器内 10 处 upstream 全部指向新容器，原地写配置并平滑 reload。
- 旧容器 `sub2api-v0216-nodes-20261015` 保持运行，未删除。
- 备份：`/opt/sub2api/backups/pre-v0217-full-20261017/`（数据库 dump 已用 pg_restore -l 检查）。
- 迁移：部署前后均 301，启动未新增迁移。
- 二进制 SHA256：`3d60d425a15f0c8c3144e30f9d9a8e0e5c54056ddd9138941fc47c1a66f5af4c`。

验证结果：

- 后端 service、handler、server/middleware/routes 测试通过。
- 前端类型检查、完整构建通过；店铺、充值条款、节点面板共 9 项测试通过。
- 使用现有管理员正常登录，候选与旧版的财务、充值、自助监控、支付配置接口均 200。
- 客户端节点管理员策略、客户端策略逐项一致；28,546 字节订阅逐字节一致。
- 管理端以原值保存节点策略成功，重新读取一致。
- 切流后公网登录接口、上述模块接口全部 200；StoreView、RechargeView、
  SelfChannelMonitorView、SettingsView 构建资源全部 200。
- 切流后观察到真实 `/v1/responses` 请求完成并返回 200，使用记录持续写入。
- 未执行真实购买、充值或扣款操作。

回滚（保持配置文件 inode，避免 bind mount 读到旧文件）：

```bash
python3 /opt/sub2api/staging/v0217-full-20261017/rollback.py
```

以下为历史升级过程，静态 YAML 兼容接口说明及早期部署状态不代表最终实现。


> 2026-10-17 更正：首次切换 v0.2.17 后，客户端节点设置不兼容，已按要求回滚。
> 本文以下首次部署的“成功”记录只是当时的检查结果，不代表客户端节点功能已验证。
> 当前实际 upstream 应以 `docker exec edge-nginx nginx -T` 和线上请求为准，
> 不能仅凭宿主机配置文件或容器健康状态判断。
> v0.2.17 修正版增加了 `client_node_global_policy`（启用、禁用节点、主/备订阅地址）
> 到新接口的只读兼容，新版静态 YAML 设置仍优先；旧容器必须保留供回滚。
> 修正版启动时额外执行了 `188_add_usage_log_thread_id.sql`（迁移总数 301）。
> 兼容修正版 `e1a7b293a` 已本地编译并上传候选目录，尚未确认最终镜像或切流：
> SSH 在密钥交换阶段超时。此时仍由 v0.2.16 处理线上请求。

## 一、本次升级做了什么

把上游框架从 **v0.2.14 升到 v0.2.17**，二开功能全部保留。

| 项目 | 状态 |
|---|---|
| 上游版本 | `v0.2.17` (060d0160) |
| 后端编译 | ✅ `go build ./...` 通过 |
| 后端测试 | ✅ `go test ./...` 全绿 |
| 前端类型检查 | ✅ `vue-tsc --noEmit` 通过 |
| 前端构建 | ✅ `pnpm build` 通过（含 i18n 校验） |
| 前端测试 | ⚠️ 19 个失败，**均为升级前已存在**，与本次无关（见第四节） |
| 冲突文件 | 41 个，全部已解决 |

### 上游 v0.2.17 的主要变化

- **平台清单化**：平台列表不再硬编码，改由后端 `domain/platforms.go` +
  前端 `platformCatalog.builtin.json` 驱动，新增 `command_code`、`cline` 两个平台。
- **Gemini 缓存内容服务**：新增 `GeminiCachedContentService` 与计费。
- **复合路由按 endpoint 解析**：`CompositeRouteResolver` 支持按入站端点分流。
- **安全修复**：EasyPay 回调参数加固（拒绝含 `&=` 的值）、支付 `return_url`
  剥离 query（防签名重放，issue #7881）、Stripe 跳转不再带 `client_secret`。

### 保留的二开功能

| 功能 | 关键文件 / 路由 |
|---|---|
| 闲鱼充值订单 | `xianyu_recharge_handler.go`、migration 192 |
| 充值条款确认 | `recharge_terms.go`、`consentGuard`、migration 243 |
| 反馈工单 | `feedback*.go`、migration 241 |
| 财务 / 分组 | `finance_handler.go`、migration 242 |
| 自助分组 & 渠道监控 | `channel_monitor_user_handler.go`、`self_monitor_settings.go` |
| 客户端供应商策略 | `client_provider_policy.go` |
| 注册赠送回收 | `signup_bonus_reclaim_service.go` |

> 关于 `consentGuard`：`routes/payment.go` 与 `router.go` 的冲突特意保留了我方
> 参数（上游无此参数），确保充值条款仍在下单前强制校验。
> `cmd/server/wire.go` 与 `wire_gen.go` 同时接入了上游新增的
> `geminiCachedContentService` / `compositeRouteResolver` 和我方的
> `feedbackReminderService` / `signupBonusReclaimService`。

### 迁移文件说明（重要）

我方与上游各自新增了 **241 / 242 / 243 同号** 的迁移文件：

```
241_add_payment_order_bonus_amount.sql   (上游)
241_add_typesafe_platform.sql            (上游)
241_feedback_tickets.sql                 (我方)
242_drop_platform_check_constraints.sql  (上游)
242_finance_groups.sql                   (我方)
242_group_i18n.sql                       (上游)
243_gemini_cached_contents.sql           (上游)
243_recharge_terms_acceptances.sql       (我方)
```

迁移执行器按 **文件名字符串排序**，并由 `schema_migrations.filename` 记录，
因此同号不会冲突、重复执行会被跳过。已确认三方迁移操作的是**互不相交的表**，
先后顺序无依赖问题。同号在此前 192 就已存在，属既有现象，本次未新增风险。

---

## 二、无感（零停机）部署方案

采用**蓝绿部署 + nginx upstream 切换**：

```
         nginx (80/443)
              │  upstream sub2api_backend
   ┌──────────┴──────────┐
   │                     │
 blue:127.0.0.1:18080   green:127.0.0.1:18081
 (旧版本，在跑)          (新版本，健康后接管)
```

流程：**起新容器 → `/health` 健康检查 → nginx reload 切流 → 停旧容器**。
任何一步失败都会中止，旧版本继续服务，**中转站不中断**。

### 前置条件

1. 服务器上已有 nginx，且站点 `reverse_proxy` 指向 `upstream sub2api_backend`。
   如果你的 nginx 目前是 `proxy_pass http://127.0.0.1:8080;`，改成：

   ```nginx
   location / {
       proxy_pass http://sub2api_backend;
       proxy_http_version 1.1;
       proxy_set_header Connection "";
       proxy_set_header Host $host;
       proxy_set_header X-Real-IP $remote_addr;
       proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
       proxy_set_header X-Forwarded-Proto $scheme;
   }
   ```

   > 若你用的是 **Caddy**（仓库 `deploy/Caddyfile`）或 Docker 容器版 nginx，
   > 把 `HAS_NGINX=0`，改为手动切端口，或告诉我实际结构我改脚本。

2. 两台容器共享同一数据卷 `sub2api_data`，`deploy/.env` 就位。

### 执行步骤（在服务器上）

```bash
cd /path/to/jokerdeck-sub2api/deploy
chmod +x bluegreen-deploy.sh

# 1) 构建新镜像（不影响线上）
./bluegreen-deploy.sh build 0.2.17

# 2) 零停机发布
./bluegreen-deploy.sh deploy 0.2.17

# 3) 查看状态
./bluegreen-deploy.sh status

# 出问题一键回滚（旧容器仍保留，未删除）
./bluegreen-deploy.sh rollback
```

### 首次迁移到蓝绿架构

如果线上现在是单个 `sub2api` 容器直接映射 8080：

```bash
# 记录当前为 blue，并把它移到 18080
docker stop sub2api
docker rm sub2api
# 用脚本起 blue（会读取 deploy/.env）
./bluegreen-deploy.sh deploy 0.2.14   # 先跑旧版本，确认 nginx 通
# 再发布新版本
./bluegreen-deploy.sh deploy 0.2.17
```

> ⚠️ 这一步会有**一次短暂中断**（迁移到蓝绿架构的一次性成本）。
> 之后所有发布都是零停机。

---

## 三、部署后验证

```bash
# 健康检查
curl -s https://your-domain/health

# 版本
curl -s https://your-domain/api/v1/version   # 或后台页面查看

# 二开功能自检
curl -s -o /dev/null -w '%{http_code}\n' https://your-domain/api/v1/recharge-terms        # 期望 401/200
curl -s -o /dev/null -w '%{http_code}\n' https://your-domain/api/v1/xianyu-recharge-orders # 期望 401
curl -s -o /dev/null -w '%{http_code}\n' https://your-domain/api/v1/feedback/tickets       # 期望 401

# 确认数据库迁移已应用
docker exec sub2api-postgres psql -U sub2api -d sub2api \
  -c "SELECT filename FROM schema_migrations ORDER BY filename DESC LIMIT 12;"
```

---

## 四、关于那 19 个前端测试失败

**结论：与本次升级无关，是二开引入的既有失败。**

- 合并后：`19 failed | 26 passed`
- 升级前 `main`（HEAD 93cdc43d5）：`19 failed | 26 passed` —— **完全一致**
- 上游 v0.2.17 干净代码：`41 passed` —— 说明失败来自二开改动

原因：二开用 `RechargeTermsGate` 包裹了整个 `PaymentView`：

```vue
<RechargeTermsGate>
  <div class="mx-auto max-w-4xl space-y-6"> ... </div>
</RechargeTermsGate>
```

`PaymentView.spec.ts` 直接挂载该组件断言内部 DOM/文本，外层多了一个 gate
（且 gate 在 `onMounted` 拉状态前会渲染条款表单）后选择器失配。
另一个失败文件 `ccswitchImport.spec.ts` 是二开自己新增的用例。

**不影响线上安全**：条款由后端 `RequireTerms` 中间件在下单时强制校验
（挂在 `orders.POST("")` 上），前端 gate 只是 UX 层。

如果希望这 19 个测试恢复，需要更新 `PaymentView.spec.ts` 的挂载辅助函数
（stub 掉 `rechargeTermsAPI.status` 并等待 settle）。这属于独立改动，
本次为「保证二开不受影响」未擅自修改测试预期。

---

## 五、本次提交

```
upgrade/v0.2.17
├─ feat: upgrade upstream framework to v0.2.17   (454cfa0aa)
└─ chore: regenerate wire_gen.go for v0.2.17     (8ac427aa3)
```

`backend/internal/web/dist` 已用 v0.2.17 前端重新构建并入库。

---

## 六、首次生产部署记录（已回滚）

**站点：** https://jokerdeck.de5.net
**时间：** 2026-10-17
**方式：** 蓝绿切换（edge-nginx upstream），**零停机**

### 实际拓扑（与预想不同，已适配）

```
edge-nginx (容器, :80/:443)
    └─ upstream sub2api-next:8080
         └─ sub2api-next 容器 (v0.2.17)
            网络：sub2api_private + public-proxy
```

- 生产原版本为 **v0.2.16**（不是 v0.2.14）
- 前端代理是 **`edge-nginx` 容器**，配置挂在 `/opt/edge/nginx.conf`
- 发布方式是**换二进制**：`FROM local/sub2api:<base>` + `COPY sub2api /app/sub2api`
- 服务器架构 **x86_64**，二进制需交叉编译 `GOOS=linux GOARCH=amd64`，
  且**必须带 `-tags embed`**（否则前端不嵌入，页面空白）

### 执行步骤

| 步骤 | 结果 |
|---|---|
| 交叉编译 v0.2.17 (linux/amd64, `-tags embed`) | ✅ 118MB，md5 `5241ac1f…` 双向校验一致 |
| 上传 + 构建镜像 `local/sub2api:v0.2.17-jokerdeck-20261017` | ✅ 镜像内二进制 md5 一致 |
| 备份 nginx.conf / env / pg_dump | ✅ `backups/pre-v0217-20261017-040256/db.sql` (101M, 107 表) |
| 迁移集比对（生产 vs 新二进制） | ✅ **300/300 完全一致**，部署时不执行任何迁移 |
| 起 `sub2api-next` + 健康检查 | ✅ healthy，0 重启 |
| 切换 nginx upstream + 平滑 reload | ✅ 未重启 nginx，无丢连接 |
| 验证后停旧容器 | ✅ 旧镜像保留可回滚 |

### 首次部署时的验证（未覆盖客户端节点功能）

```
/health                        -> 200
/                              -> 200
/login                         -> 200
/api/v1/settings/public        -> 200
/api/v1/recharge-terms         -> 401   (二开：充值条款)
/api/v1/finance/users          -> 401   (二开：财务)
/api/v1/self-channel-monitors  -> 401   (二开：自助监控)
/api/v1/xianyu-recharge-orders -> 401   (二开：闲鱼充值)
/api/v1/feedback/tickets       -> 401   (二开：反馈工单)
/v1/messages                   -> 401   (中转网关正常，非 502)

前端懒加载分块（二开页面）：
FinanceView / SelfChannelMonitorView / StoreView /
XianyuRechargeOrdersView / FeedbackView  -> 全部 200

nginx 5xx 计数：0      应用 ERROR/panic：0
```

### 回滚方式

```bash
docker start sub2api-v0216-nodes-20261015
# /opt/edge/nginx.conf 是文件 bind mount；不能用 cp / sed -i 替换宿主机 inode。
# 先备份并确认配置，再原地写入 upstream；验证容器内的实际内容。
# 修改前先确认 /opt/edge/nginx.conf 的真实路径、旧值和目标容器名。
# docker exec edge-nginx nginx -t && docker exec edge-nginx nginx -s reload
# docker exec edge-nginx nginx -T | grep sub2api_upstream
# 如果挂载 inode 已经与宿主机脱节，重启 edge-nginx 后再验证真实 upstream。
```

或直接用服务器上已安装的通用脚本：

```bash
/opt/sub2api/deploy.sh <新二进制路径> <版本标签>
```

### 客户端接口验收项

- 使用正常登录的客户端会话核对 `GET /api/v1/client/node-policy` 返回旧配置的
  `enabled` 与 `disabled_nodes`，并核对 `GET /api/v1/client/node-subscription` 返回可解析的原始 YAML。
- 使用管理员会话核对 `GET /api/v1/admin/client/node-subscription` 的 `content`；
  `PUT /api/v1/admin/client/nodes` 写入新格式后应立即反映在客户端读取结果中。
- 未登录请求返回 401 只能证明授权门禁存在，**不能**证明旧配置被正确下发。
- 切流时必须核对 nginx 容器内 10 处 upstream、公开站点健康检查及真实中转请求，
  并保留旧容器运行，直至客户端节点验收完成。

### 遗留说明

- 首次切换后由于客户端节点配置不兼容，线上已回滚至
  `sub2api-v0216-nodes-20261015`；不能以首次验证结果宣称新版已上线。
- 修正版必须在预发布容器验证旧策略和订阅读取后才能切流；
  `sub2api-v0216-nodes-20261015` 必须保持运行以便即时回滚。
- 旧配置的主/备订阅 URL 仅在新版静态订阅尚未设置时生效；
  管理员在新版保存静态 YAML 后，旧版不会读取该新键。回滚前应确认配置兼容性。
- 上述早期 `300/300` 迁移对比仅适用于首版二进制；修正版包含额外迁移。
