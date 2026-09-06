<template>
  <AppLayout>
    <div class="space-y-4">
      <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
        <div class="flex flex-wrap items-center gap-3">
          <div class="min-w-56 flex-1 sm:max-w-80">
            <input
              v-model.trim="search"
              class="input w-full"
              type="search"
              placeholder="搜索订单号、邮箱或用户名"
              @keyup.enter="applyFilters"
            />
          </div>
          <Select v-model="status" :options="statusOptions" class="w-36" @change="applyFilters" />
          <button class="btn btn-secondary" :disabled="loading" title="刷新" @click="loadOrders">
            <Icon name="refresh" size="md" :class="{ 'animate-spin': loading }" />
          </button>
          <div class="ml-auto flex items-center gap-2 rounded-md bg-amber-50 px-3 py-2 text-sm text-amber-800 dark:bg-amber-900/20 dark:text-amber-300">
            <span>待审核</span>
            <span class="font-semibold">{{ pendingCount }}</span>
          </div>
        </div>
      </section>

      <section class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
        <DataTable :columns="columns" :data="orders" :loading="loading">
          <template #cell-user_email="{ row }">
            <div class="min-w-0">
              <p class="break-all text-sm font-medium text-gray-900 dark:text-white">{{ row.user_email || '-' }}</p>
              <p v-if="row.username" class="mt-0.5 text-xs text-gray-500">{{ row.username }}</p>
            </div>
          </template>
          <template #cell-xianyu_order_no="{ value }">
            <span class="break-all font-mono text-sm">{{ value }}</span>
          </template>
          <template #cell-paid_amount_cny="{ value }">
            <span class="font-semibold">¥{{ Number(value).toFixed(2) }}</span>
          </template>
          <template #cell-status="{ value }">
            <span class="badge" :class="statusClass(value)">{{ statusLabel(value) }}</span>
          </template>
          <template #cell-created_at="{ value }">
            <span class="whitespace-nowrap text-sm text-gray-500">{{ formatDateTime(value) }}</span>
          </template>
          <template #cell-actions="{ row }">
            <div v-if="row.status === 'pending'" class="flex justify-end gap-2">
              <button class="btn btn-primary px-3 py-1.5 text-xs" @click="openCredit(row)">确认到账</button>
              <button class="btn btn-danger px-3 py-1.5 text-xs" @click="openReject(row)">驳回</button>
            </div>
            <div v-else class="text-right text-xs text-gray-500">
              <p v-if="row.credited_amount_usd != null">到账 ${{ row.credited_amount_usd.toFixed(2) }}</p>
              <p v-if="row.admin_note" class="mt-1 max-w-64 break-words">{{ row.admin_note }}</p>
            </div>
          </template>
        </DataTable>
        <div v-if="pagination.total > 0" class="border-t border-gray-100 p-4 dark:border-dark-700">
          <Pagination
            :page="pagination.page"
            :page-size="pagination.page_size"
            :total="pagination.total"
            @update:page="changePage"
            @update:pageSize="changePageSize"
          />
        </div>
      </section>
    </div>

    <BaseDialog :show="!!creditTarget" title="确认充值到账" width="narrow" @close="closeDialogs">
      <div v-if="creditTarget" class="space-y-4">
        <div class="rounded-md bg-gray-50 p-3 text-sm dark:bg-dark-800">
          <p class="break-all font-mono text-gray-900 dark:text-white">{{ creditTarget.xianyu_order_no }}</p>
          <p class="mt-2 text-gray-600 dark:text-gray-300">
            用户实付：<strong>¥{{ creditTarget.paid_amount_cny.toFixed(2) }}</strong>
          </p>
        </div>
        <div>
          <label for="credit-amount" class="input-label">到账金额（美元）</label>
          <div class="relative mt-1">
            <span class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-gray-500">$</span>
            <input
              id="credit-amount"
              v-model.number="creditAmount"
              type="number"
              min="0.01"
              max="1000000"
              step="0.01"
              class="input w-full pl-8"
              placeholder="0.00"
            />
          </div>
        </div>
        <div>
          <label for="credit-note" class="input-label">管理员备注（选填）</label>
          <textarea id="credit-note" v-model.trim="adminNote" class="input mt-1 min-h-20 w-full resize-y" maxlength="1000"></textarea>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="closeDialogs">取消</button>
          <button class="btn btn-primary" :disabled="actionLoading || !creditAmount || creditAmount <= 0" @click="confirmCredit">
            {{ actionLoading ? '处理中...' : '确认并入账' }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="!!rejectTarget" title="驳回充值订单" width="narrow" @close="closeDialogs">
      <div v-if="rejectTarget" class="space-y-4">
        <p class="break-all rounded-md bg-gray-50 p-3 font-mono text-sm dark:bg-dark-800">{{ rejectTarget.xianyu_order_no }}</p>
        <div>
          <label for="reject-reason" class="input-label">驳回原因</label>
          <textarea
            id="reject-reason"
            v-model.trim="rejectReason"
            class="input mt-1 min-h-24 w-full resize-y"
            maxlength="1000"
            placeholder="用户会在充值中心和邮件中看到此原因"
          ></textarea>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="closeDialogs">取消</button>
          <button class="btn btn-danger" :disabled="actionLoading || !rejectReason" @click="confirmReject">
            {{ actionLoading ? '处理中...' : '确认驳回' }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import type { Column } from '@/components/common/types'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import {
  xianyuRechargeAPI,
  type XianyuRechargeOrder,
  type XianyuRechargeStatus
} from '@/api/xianyuRecharge'

const appStore = useAppStore()
const orders = ref<XianyuRechargeOrder[]>([])
const loading = ref(false)
const actionLoading = ref(false)
const pendingCount = ref(0)
const search = ref('')
const status = ref<XianyuRechargeStatus | ''>('pending')
const creditTarget = ref<XianyuRechargeOrder | null>(null)
const rejectTarget = ref<XianyuRechargeOrder | null>(null)
const creditAmount = ref<number | null>(null)
const adminNote = ref('')
const rejectReason = ref('')
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const columns = computed<Column[]>(() => [
  { key: 'user_email', label: '用户' },
  { key: 'xianyu_order_no', label: '闲鱼订单号' },
  { key: 'paid_amount_cny', label: '实付' },
  { key: 'status', label: '状态' },
  { key: 'created_at', label: '提交时间' },
  { key: 'actions', label: '操作', class: 'text-right' }
])

const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'pending', label: '待审核' },
  { value: 'credited', label: '已到账' },
  { value: 'rejected', label: '已驳回' }
]

function errorMessage(error: unknown): string {
  if (error && typeof error === 'object' && 'message' in error) {
    return String((error as { message?: unknown }).message || '操作失败')
  }
  return '操作失败，请稍后重试'
}

function statusLabel(value: XianyuRechargeStatus): string {
  return { pending: '待审核', credited: '已到账', rejected: '已驳回' }[value]
}

function statusClass(value: XianyuRechargeStatus): string {
  return { pending: 'badge-warning', credited: 'badge-success', rejected: 'badge-danger' }[value]
}

async function loadOrders() {
  loading.value = true
  try {
    const [result, count] = await Promise.all([
      xianyuRechargeAPI.listAdmin({
        page: pagination.page,
        page_size: pagination.page_size,
        status: status.value,
        search: search.value || undefined
      }),
      xianyuRechargeAPI.pendingCount()
    ])
    orders.value = result.items || []
    pagination.total = result.total || 0
    pendingCount.value = count
  } catch (error) {
    appStore.showError(errorMessage(error))
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  pagination.page = 1
  void loadOrders()
}

function changePage(page: number) {
  pagination.page = page
  void loadOrders()
}

function changePageSize(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadOrders()
}

function openCredit(order: XianyuRechargeOrder) {
  creditTarget.value = order
  creditAmount.value = null
  adminNote.value = ''
}

function openReject(order: XianyuRechargeOrder) {
  rejectTarget.value = order
  rejectReason.value = ''
}

function closeDialogs() {
  if (actionLoading.value) return
  creditTarget.value = null
  rejectTarget.value = null
}

async function confirmCredit() {
  if (!creditTarget.value || !creditAmount.value || creditAmount.value <= 0) return
  actionLoading.value = true
  try {
    await xianyuRechargeAPI.credit(creditTarget.value.id, creditAmount.value, adminNote.value || undefined)
    appStore.showSuccess('订单已确认，用户余额已到账')
    closeDialogs()
    await loadOrders()
  } catch (error) {
    appStore.showError(errorMessage(error))
  } finally {
    actionLoading.value = false
    creditTarget.value = null
  }
}

async function confirmReject() {
  if (!rejectTarget.value || !rejectReason.value) return
  actionLoading.value = true
  try {
    await xianyuRechargeAPI.reject(rejectTarget.value.id, rejectReason.value)
    appStore.showSuccess('订单已驳回')
    closeDialogs()
    await loadOrders()
  } catch (error) {
    appStore.showError(errorMessage(error))
  } finally {
    actionLoading.value = false
    rejectTarget.value = null
  }
}

onMounted(loadOrders)
</script>
