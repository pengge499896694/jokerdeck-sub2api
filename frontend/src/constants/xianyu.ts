/**
 * 二开模块开关（闲鱼充值 / 店铺销售）。
 *
 * 这两个模块是相对官方上游的二开增量，不由 public settings 驱动，
 * 因此用编译期常量控制，改一行即可整体开关。关闭时：
 *
 *   - 侧边栏对应菜单项不渲染（AppSidebar.vue 的 featureFlag）
 *   - 对应路由不注册，直接访问 URL 落到 NotFound（router/index.ts）
 *
 * 后端接口（/xianyu-recharge-orders、/admin/xianyu-recharge-orders）与
 * xianyu_recharge_orders 表**保持不动**，历史订单数据不受影响；重新开启
 * 只需把开关改回 true 并重建镜像。
 */

/** 闲鱼充值：用户端「充值中心」与管理端「充值订单」。 */
export const XIANYU_RECHARGE_ENABLED = true

/** 店铺销售：用户端「店铺销售」（catfk.com iframe），保留作为自动充值渠道。 */
export const XIANYU_STORE_ENABLED = true
