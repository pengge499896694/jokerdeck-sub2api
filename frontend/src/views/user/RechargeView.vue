<template>
  <AppLayout>
    <div class="mx-auto max-w-5xl space-y-5">
      <section class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
        <div class="grid gap-6 p-5 md:grid-cols-[1fr_auto] md:items-center md:p-7">
          <div class="flex min-w-0 items-start gap-4">
            <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
              <Icon name="creditCard" size="lg" />
            </div>
            <div class="min-w-0">
              <h1 class="text-xl font-semibold text-gray-900 dark:text-white">充值中心</h1>
              <p class="mt-1 text-sm leading-6 text-gray-600 dark:text-gray-300">
                在闲鱼完成付款后提交订单号。管理员核对订单并确认到账，处理结果会同步显示并发送邮件。
              </p>
            </div>
          </div>
          <a
            :href="xianyuLink"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-primary w-full justify-center md:w-auto"
          >
            <Icon name="externalLink" size="md" class="mr-2" />
            前往闲鱼下单
          </a>
        </div>
        <div class="grid border-t border-gray-100 bg-gray-50/70 sm:grid-cols-2 dark:border-dark-700 dark:bg-dark-800/50">
          <div class="px-5 py-3 text-sm text-gray-600 dark:text-gray-300">
            客服 QQ：<span class="font-medium text-gray-900 dark:text-white">3856645129</span>
          </div>
          <div class="border-t border-gray-100 px-5 py-3 text-sm text-gray-600 sm:border-l sm:border-t-0 dark:border-dark-700 dark:text-gray-300">
            客服邮箱：
            <a class="font-medium text-primary-600 hover:underline dark:text-primary-400" href="mailto:3856645129@qq.com">3856645129@qq.com</a>
          </div>
        </div>
      </section>

      <div class="grid gap-5 lg:grid-cols-[minmax(0,380px)_minmax(0,1fr)]">
        <section class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
          <div class="mb-5">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">提交充值订单</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">账号信息将从当前登录用户自动获取。</p>
          </div>

          <form class="space-y-4" @submit.prevent="submitOrder">
            <div>
              <label for="xianyu-order-no" class="input-label">闲鱼订单号</label>
              <input
                id="xianyu-order-no"
                v-model.trim="form.xianyu_order_no"
                class="input mt-1 w-full"
                type="text"
                maxlength="128"
                required
                autocomplete="off"
                placeholder="请输入闲鱼订单号"
              />
            </div>
            <div>
              <label for="paid-amount" class="input-label">实付金额（人民币）</label>
              <div class="relative mt-1">
                <span class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-gray-500">¥</span>
                <input
                  id="paid-amount"
                  v-model.number="form.paid_amount_cny"
                  class="input w-full pl-8"
                  type="number"
                  min="0.01"
                  max="1000000"
                  step="0.01"
                  required
                  placeholder="0.00"
                />
              </div>
            </div>
            <div>
              <label for="order-note" class="input-label">备注（选填）</label>
              <textarea
                id="order-note"
                v-model.trim="form.note"
                class="input mt-1 min-h-24 w-full resize-y"
                maxlength="1000"
                placeholder="例如：闲鱼昵称、特殊说明"
              ></textarea>
            </div>
            <button
              type="submit"
              class="btn btn-primary w-full justify-center"
              :disabled="submitting || !canSubmit"
            >
              <Icon name="checkCircle" size="md" class="mr-2" />
              {{ submitting ? '正在提交...' : '提交审核' }}
            </button>
          </form>
        </section>

        <section class="min-w-0 rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
          <div class="flex items-center justify-between border-b border-gray-100 px-5 py-4 dark:border-dark-700">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">我的充值记录</h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">管理员确认后余额会自动到账</p>
            </div>
            <button class="btn btn-secondary px-3" :disabled="loading" title="刷新" @click="loadOrders">
              <Icon name="refresh" size="md" :class="{ 'animate-spin': loading }" />
            </button>
          </div>

          <div v-if="loading && orders.length === 0" class="p-10 text-center text-sm text-gray-500">正在加载...</div>
          <div v-else-if="orders.length === 0" class="p-10 text-center">
            <Icon name="inbox" size="xl" class="mx-auto text-gray-300 dark:text-dark-600" />
            <p class="mt-3 text-sm text-gray-500 dark:text-gray-400">还没有提交过充值订单</p>
          </div>
          <div v-else class="divide-y divide-gray-100 dark:divide-dark-700">
            <article v-for="order in orders" :key="order.id" class="p-5">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div class="min-w-0">
                  <p class="break-all font-mono text-sm font-medium text-gray-900 dark:text-white">{{ order.xianyu_order_no }}</p>
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ formatDateTime(order.created_at) }}</p>
                </div>
                <span class="badge" :class="statusClass(order.status)">{{ statusLabel(order.status) }}</span>
              </div>
              <div class="mt-4 grid grid-cols-2 gap-3 text-sm">
                <div>
                  <p class="text-xs text-gray-500 dark:text-gray-400">实付金额</p>
                  <p class="mt-1 font-semibold text-gray-900 dark:text-white">¥{{ order.paid_amount_cny.toFixed(2) }}</p>
                </div>
                <div>
                  <p class="text-xs text-gray-500 dark:text-gray-400">到账金额</p>
                  <p class="mt-1 font-semibold text-gray-900 dark:text-white">
                    {{ order.credited_amount_usd != null ? `$${order.credited_amount_usd.toFixed(2)}` : '-' }}
                  </p>
                </div>
              </div>
              <p v-if="order.note" class="mt-3 break-words text-sm text-gray-600 dark:text-gray-300">备注：{{ order.note }}</p>
              <p v-if="order.admin_note" class="mt-3 rounded-md bg-gray-50 px-3 py-2 text-sm text-gray-700 dark:bg-dark-800 dark:text-gray-300">
                管理员备注：{{ order.admin_note }}
              </p>
            </article>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import {
  xianyuRechargeAPI,
  type XianyuRechargeOrder,
  type XianyuRechargeStatus
} from '@/api/xianyuRecharge'

const xianyuLink = 'https://m.tb.cn/h.8orR1uB?tk=qgivTTR6A22'
const appStore = useAppStore()
const orders = ref<XianyuRechargeOrder[]>([])
const loading = ref(false)
const submitting = ref(false)
const form = reactive({
  xianyu_order_no: '',
  paid_amount_cny: null as number | null,
  note: ''
})

const canSubmit = computed(() =>
  form.xianyu_order_no.length > 0 &&
  form.paid_amount_cny != null &&
  form.paid_amount_cny > 0
)

function errorMessage(error: unknown): string {
  if (error && typeof error === 'object' && 'message' in error) {
    return String((error as { message?: unknown }).message || '操作失败')
  }
  return '操作失败，请稍后重试'
}

function statusLabel(status: XianyuRechargeStatus): string {
  return { pending: '待审核', credited: '已到账', rejected: '已驳回' }[status]
}

function statusClass(status: XianyuRechargeStatus): string {
  return {
    pending: 'badge-warning',
    credited: 'badge-success',
    rejected: 'badge-danger'
  }[status]
}

async function loadOrders() {
  loading.value = true
  try {
    orders.value = await xianyuRechargeAPI.listMine()
  } catch (error) {
    appStore.showError(errorMessage(error))
  } finally {
    loading.value = false
  }
}

async function submitOrder() {
  if (!canSubmit.value || form.paid_amount_cny == null) return
  submitting.value = true
  try {
    await xianyuRechargeAPI.create({
      xianyu_order_no: form.xianyu_order_no,
      paid_amount_cny: form.paid_amount_cny,
      note: form.note || undefined
    })
    form.xianyu_order_no = ''
    form.paid_amount_cny = null
    form.note = ''
    appStore.showSuccess('充值订单已提交，请等待管理员审核')
    await loadOrders()
  } catch (error) {
    appStore.showError(errorMessage(error))
  } finally {
    submitting.value = false
  }
}

onMounted(loadOrders)
</script>
