<template>
  <AppLayout>
    <div class="space-y-6">
      <section v-if="isAdmin" class="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">记录数</p>
          <p class="mt-1 text-xl font-semibold text-gray-900 dark:text-gray-100">{{ formatNumber(storageStats?.records) }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">问题与回复文本</p>
          <p class="mt-1 text-xl font-semibold text-gray-900 dark:text-gray-100">{{ formatBytes(storageStats?.text_bytes) }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">数据库占用</p>
          <p class="mt-1 text-xl font-semibold text-gray-900 dark:text-gray-100">{{ formatBytes(storageStats?.table_bytes) }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">保留期限 / 队列丢弃</p>
          <p class="mt-1 text-xl font-semibold text-gray-900 dark:text-gray-100">{{ storageStats?.retention_days ?? '—' }} 天 / {{ formatNumber(storageStats?.queue_dropped) }}</p>
        </div>
      </section>

      <section class="card p-4 sm:p-6">
        <form class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4" @submit.prevent="applyFilters">
          <div>
            <label class="input-label">用户邮箱</label>
            <input v-model="filters.user" class="input" type="search" placeholder="输入邮箱片段" />
          </div>
          <div>
            <label class="input-label">密钥名称</label>
            <input v-model="filters.api_key" class="input" type="search" placeholder="输入密钥名称片段" />
          </div>
          <div>
            <label class="input-label">分组</label>
            <input v-model="filters.group" class="input" type="search" placeholder="输入分组名称片段" />
          </div>
          <div>
            <label class="input-label">模型</label>
            <input v-model="filters.model" class="input" type="search" placeholder="输入模型名称" />
          </div>
          <div>
            <label class="input-label">开始日期</label>
            <input v-model="filters.start_date" class="input" type="date" />
          </div>
          <div>
            <label class="input-label">结束日期</label>
            <input v-model="filters.end_date" class="input" type="date" />
          </div>
          <div>
            <label class="input-label">结果</label>
            <select v-model="filters.status" class="input">
              <option value="">全部</option>
              <option value="success">成功</option>
              <option value="failed">失败</option>
            </select>
          </div>
          <div class="flex items-end gap-2">
            <button type="submit" class="btn btn-primary flex-1" :disabled="loading">查询</button>
            <button type="button" class="btn btn-secondary" :disabled="loading" @click="resetFilters">重置</button>
          </div>
        </form>
      </section>

      <section class="card overflow-hidden">
        <div class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-200 px-4 py-3 dark:border-dark-700">
          <div class="text-sm text-gray-600 dark:text-gray-300">共 {{ formatNumber(total) }} 条记录</div>
          <div class="flex items-center gap-2">
            <button type="button" class="btn btn-secondary" :disabled="loading" @click="loadRecords">刷新</button>
            <button
              v-if="isAdmin"
              type="button"
              class="btn btn-danger"
              :disabled="selectedIds.length === 0 || deleting"
              @click="deleteSelected"
            >
              {{ deleting ? '删除中…' : `批量删除 (${selectedIds.length})` }}
            </button>
          </div>
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-[1050px] w-full text-left text-sm">
            <thead class="bg-gray-50 text-xs uppercase text-gray-500 dark:bg-dark-800 dark:text-gray-400">
              <tr>
                <th class="w-10 px-4 py-3">
                  <input
                    v-if="isAdmin"
                    type="checkbox"
                    aria-label="选择当前页全部记录"
                    :checked="allCurrentPageSelected"
                    :disabled="records.length === 0"
                    @change="toggleCurrentPageSelection"
                  />
                </th>
                <th class="px-4 py-3">时间 / 请求</th>
                <th class="px-4 py-3">用户</th>
                <th class="px-4 py-3">密钥 / 分组</th>
                <th class="px-4 py-3">模型</th>
                <th class="px-4 py-3">Token</th>
                <th class="px-4 py-3">结果</th>
                <th class="px-4 py-3">详情</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <template v-for="item in records" :key="item.id">
                <tr class="align-top hover:bg-gray-50/70 dark:hover:bg-dark-800/50">
                  <td class="px-4 py-3">
                    <input v-if="isAdmin" v-model="selectedIds" type="checkbox" :value="item.id" :aria-label="`选择记录 ${item.id}`" />
                  </td>
                  <td class="px-4 py-3">
                    <div class="whitespace-nowrap font-medium text-gray-900 dark:text-gray-100">{{ formatDate(item.created_at) }}</div>
                    <div class="mt-1 max-w-52 truncate font-mono text-xs text-gray-500" :title="item.request_id">{{ item.request_id }}</div>
                    <div class="mt-1 text-xs text-gray-500">{{ item.endpoint }} · {{ item.duration_ms }} ms</div>
                  </td>
                  <td class="px-4 py-3">
                    <div class="text-gray-900 dark:text-gray-100">{{ item.user_email || `用户 #${item.user_id}` }}</div>
                    <div class="mt-1 text-xs text-gray-500">用户 ID {{ item.user_id }}</div>
                  </td>
                  <td class="px-4 py-3">
                    <div class="text-gray-900 dark:text-gray-100">{{ item.api_key_name || `密钥 #${item.api_key_id}` }}</div>
                    <div class="mt-1 text-xs text-gray-500">{{ item.group_name || '无分组' }}</div>
                  </td>
                  <td class="px-4 py-3">
                    <div class="max-w-48 break-all text-gray-900 dark:text-gray-100">{{ item.model || '—' }}</div>
                    <div v-if="item.stream" class="mt-1 text-xs text-gray-500">流式</div>
                  </td>
                  <td class="whitespace-nowrap px-4 py-3 text-xs text-gray-600 dark:text-gray-300">
                    <div>输入 {{ formatNumber(item.input_tokens) }}</div>
                    <div class="mt-1">输出 {{ formatNumber(item.output_tokens) }}</div>
                  </td>
                  <td class="px-4 py-3">
                    <span class="inline-flex rounded-full px-2 py-1 text-xs font-medium" :class="item.success ? 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300' : 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300'">
                      {{ item.success ? '成功' : '失败' }} · {{ item.status_code || '—' }}
                    </span>
                  </td>
                  <td class="px-4 py-3">
                    <button type="button" class="text-primary-600 hover:text-primary-700 dark:text-primary-400" @click="toggleDetails(item)">
                      {{ expandedId === item.id ? '收起' : '查看问题与回复' }}
                    </button>
                  </td>
                </tr>
                <tr v-if="expandedId === item.id" class="bg-gray-50/80 dark:bg-dark-800/40">
                  <td colspan="8" class="px-4 py-4">
                    <div v-if="loadingDetailId === item.id" class="text-sm text-gray-500">正在加载详情…</div>
                    <div v-else-if="details[item.id]" class="grid gap-4 lg:grid-cols-2">
                      <div class="min-w-0 rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
                        <div class="mb-2 flex items-center justify-between gap-2 text-sm font-semibold text-gray-900 dark:text-gray-100">
                          <span>用户问题</span>
                          <span class="text-xs font-normal text-gray-500">{{ formatBytes(details[item.id].question_bytes) }}{{ details[item.id].question_truncated ? ' · 已截断' : '' }}</span>
                        </div>
                        <pre class="max-h-[32rem] overflow-auto whitespace-pre-wrap break-words text-sm text-gray-700 dark:text-gray-300">{{ details[item.id].question || '未提取到文本内容' }}</pre>
                      </div>
                      <div class="min-w-0 rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
                        <div class="mb-2 flex items-center justify-between gap-2 text-sm font-semibold text-gray-900 dark:text-gray-100">
                          <span>AI 回复</span>
                          <span class="text-xs font-normal text-gray-500">{{ formatBytes(details[item.id].answer_bytes) }}{{ details[item.id].answer_truncated ? ' · 已截断' : '' }}</span>
                        </div>
                        <pre class="max-h-[32rem] overflow-auto whitespace-pre-wrap break-words text-sm text-gray-700 dark:text-gray-300">{{ details[item.id].answer || '未提取到文本内容' }}</pre>
                      </div>
                    </div>
                    <p v-else class="text-sm text-red-600">详情加载失败，请重试。</p>
                  </td>
                </tr>
              </template>
              <tr v-if="!loading && records.length === 0">
                <td colspan="8" class="px-4 py-12 text-center text-sm text-gray-500">没有符合条件的对话记录</td>
              </tr>
              <tr v-if="loading">
                <td colspan="8" class="px-4 py-12 text-center text-sm text-gray-500">正在加载…</td>
              </tr>
            </tbody>
          </table>
        </div>
        <Pagination
          v-if="total > 0"
          :page="page"
          :total="total"
          :page-size="pageSize"
          @update:page="changePage"
          @update:pageSize="changePageSize"
        />
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useAppStore, useAuthStore } from '@/stores'
import { conversationLogsAPI } from '@/api'
import type { ConversationLog, ConversationLogFilters, ConversationLogStorageStats } from '@/api/conversationLogs'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'

const appStore = useAppStore()
const authStore = useAuthStore()
const isAdmin = computed(() => authStore.isAdmin)
const today = new Date()
function toDateInput(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
const initialStart = new Date(today)
initialStart.setDate(initialStart.getDate() - 6)
const filters = reactive({
  user: '',
  api_key: '',
  group: '',
  model: '',
  status: '' as '' | 'success' | 'failed',
  start_date: toDateInput(initialStart),
  end_date: toDateInput(today)
})
const records = ref<ConversationLog[]>([])
const details = ref<Record<number, ConversationLog>>({})
const selectedIds = ref<number[]>([])
const storageStats = ref<ConversationLogStorageStats | null>(null)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const loading = ref(false)
const deleting = ref(false)
const expandedId = ref<number | null>(null)
const loadingDetailId = ref<number | null>(null)
let listRequestSequence = 0

const allCurrentPageSelected = computed(() => records.value.length > 0 && records.value.every(item => selectedIds.value.includes(item.id)))

function formatNumber(value: number | undefined): string {
  return new Intl.NumberFormat().format(value ?? 0)
}

function formatBytes(value: number | undefined): string {
  if (value === undefined) return '—'
  if (value < 1024) return `${value} B`
  const units = ['KiB', 'MiB', 'GiB', 'TiB']
  let amount = value
  let unitIndex = -1
  do {
    amount /= 1024
    unitIndex++
  } while (amount >= 1024 && unitIndex < units.length - 1)
  return `${amount.toFixed(1)} ${units[unitIndex]}`
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'medium' }).format(new Date(value))
}

function buildFilters(): ConversationLogFilters {
  return {
    user: filters.user.trim() || undefined,
    api_key: filters.api_key.trim() || undefined,
    group: filters.group.trim() || undefined,
    model: filters.model.trim() || undefined,
    status: filters.status || undefined,
    start_date: filters.start_date || undefined,
    end_date: filters.end_date || undefined,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    page: page.value,
    page_size: pageSize.value
  }
}

async function loadRecords(): Promise<void> {
  const sequence = ++listRequestSequence
  loading.value = true
  try {
    const result = await conversationLogsAPI.list(buildFilters())
    if (sequence !== listRequestSequence) return
    records.value = result.items ?? []
    total.value = result.total ?? 0
    selectedIds.value = selectedIds.value.filter(id => records.value.some(item => item.id === id))
  } catch (error: any) {
    if (sequence === listRequestSequence) appStore.showError(error?.message || '加载对话记录失败')
  } finally {
    if (sequence === listRequestSequence) loading.value = false
  }
}

async function loadStorageStats(): Promise<void> {
  if (!isAdmin.value) return
  try {
    storageStats.value = await conversationLogsAPI.getStorageStats()
  } catch (error: any) {
    appStore.showError(error?.message || '加载存储统计失败')
  }
}

function applyFilters(): void {
  page.value = 1
  selectedIds.value = []
  expandedId.value = null
  void loadRecords()
}

function resetFilters(): void {
  filters.user = ''
  filters.api_key = ''
  filters.group = ''
  filters.model = ''
  filters.status = ''
  filters.start_date = ''
  filters.end_date = ''
  applyFilters()
}

function changePage(value: number): void {
  page.value = value
  expandedId.value = null
  void loadRecords()
}

function changePageSize(value: number): void {
  pageSize.value = value
  page.value = 1
  selectedIds.value = []
  expandedId.value = null
  void loadRecords()
}

function toggleCurrentPageSelection(event: Event): void {
  const checked = (event.target as HTMLInputElement).checked
  const currentIds = records.value.map(item => item.id)
  selectedIds.value = checked
    ? Array.from(new Set([...selectedIds.value, ...currentIds]))
    : selectedIds.value.filter(id => !currentIds.includes(id))
}

async function toggleDetails(item: ConversationLog): Promise<void> {
  if (expandedId.value === item.id) {
    expandedId.value = null
    return
  }
  expandedId.value = item.id
  if (details.value[item.id]) return
  loadingDetailId.value = item.id
  try {
    details.value[item.id] = await conversationLogsAPI.getById(item.id)
  } catch (error: any) {
    appStore.showError(error?.message || '加载对话详情失败')
  } finally {
    if (loadingDetailId.value === item.id) loadingDetailId.value = null
  }
}

async function deleteSelected(): Promise<void> {
  const ids = [...selectedIds.value]
  if (ids.length === 0 || !window.confirm(`确定批量删除选中的 ${ids.length} 条对话记录吗？`)) return
  deleting.value = true
  try {
    const result = await conversationLogsAPI.batchDelete(ids)
    appStore.showSuccess(`已删除 ${result.deleted} 条对话记录`)
    selectedIds.value = []
    await Promise.all([loadRecords(), loadStorageStats()])
  } catch (error: any) {
    appStore.showError(error?.message || '批量删除失败')
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  void loadRecords()
  void loadStorageStats()
})
</script>
