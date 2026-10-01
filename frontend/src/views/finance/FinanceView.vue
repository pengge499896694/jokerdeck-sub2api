<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">财务中心</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">按时间范围、用户和财务分组核对充值与使用明细。</p>
        </div>
        <div class="flex flex-wrap gap-2">
          <button class="btn btn-secondary" :disabled="loading" @click="loadAll">
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
            刷新
          </button>
          <button class="btn btn-primary" :disabled="loading" @click="downloadExport">
            <Icon name="download" size="sm" />
            导出{{ activeTab === 'recharge' ? '充值' : '使用' }}明细
          </button>
        </div>
      </div>

      <section class="card p-4">
        <div class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-6">
          <div>
            <label class="input-label">起始月份</label>
            <input v-model="filters.start_month" type="month" class="input" :max="filters.end_month || undefined" @change="search" />
          </div>
          <div>
            <label class="input-label">结束月份</label>
            <input v-model="filters.end_month" type="month" class="input" :min="filters.start_month || undefined" @change="search" />
          </div>
          <div>
            <label class="input-label">用户</label>
            <Select v-model="filters.user_id" :options="userOptions" searchable clearable @change="search" />
          </div>
          <div>
            <label class="input-label">财务分组</label>
            <Select v-model="filters.group_id" :options="groupOptions" clearable @change="search" />
          </div>
          <div>
            <label class="input-label">用户搜索</label>
            <input v-model.trim="filters.search" class="input" placeholder="邮箱或用户名" @keyup.enter="search" />
          </div>
          <div class="flex items-end">
            <button class="btn btn-secondary w-full" :disabled="loading" @click="resetFilters">重置筛选</button>
          </div>
        </div>
      </section>

      <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div class="card p-4">
          <div class="text-sm text-gray-500 dark:text-gray-400">{{ activeTab === 'recharge' ? '充值金额' : '使用费用' }}</div>
          <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">${{ formatMoney(page.total_amount) }}</div>
        </div>
        <div class="card p-4">
          <div class="text-sm text-gray-500 dark:text-gray-400">记录数</div>
          <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ page.total }}</div>
        </div>
        <div class="card p-4">
          <div class="text-sm text-gray-500 dark:text-gray-400">当前范围</div>
          <div class="mt-2 text-lg font-medium text-gray-900 dark:text-white">{{ rangeLabel }}</div>
        </div>
      </div>

      <div class="border-b border-gray-200 dark:border-dark-700">
        <button
          v-for="tab in tabs"
          :key="tab.value"
          class="mr-6 border-b-2 px-1 pb-3 text-sm font-medium"
          :class="activeTab === tab.value ? 'border-primary-500 text-primary-600 dark:text-primary-400' : 'border-transparent text-gray-500'"
          @click="switchTab(tab.value)"
        >
          {{ tab.label }}
        </button>
      </div>

      <div class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
        <div class="overflow-x-auto">
          <table class="min-w-full text-left text-sm">
            <thead class="bg-gray-50 text-xs uppercase text-gray-500 dark:bg-dark-800 dark:text-gray-400">
              <tr v-if="activeTab === 'recharge'">
                <th class="px-4 py-3">时间</th><th class="px-4 py-3">用户</th><th class="px-4 py-3">分组</th><th class="px-4 py-3">类型</th><th class="px-4 py-3 text-right">金额</th><th class="px-4 py-3">状态</th><th class="px-4 py-3">备注</th>
              </tr>
              <tr v-else>
                <th class="px-4 py-3">时间</th><th class="px-4 py-3">用户</th><th class="px-4 py-3">分组</th><th class="px-4 py-3">模型</th><th class="px-4 py-3">请求 ID</th><th class="px-4 py-3 text-right">Tokens</th><th class="px-4 py-3 text-right">费用</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
              <template v-if="activeTab === 'recharge'">
                <tr v-for="item in rechargeItems" :key="item.id">
                  <td class="whitespace-nowrap px-4 py-3 text-gray-500">{{ formatTime(item.created_at) }}</td>
                  <td class="px-4 py-3"><div class="font-medium">{{ item.username || '—' }}</div><div class="text-xs text-gray-500">{{ item.email }}</div></td>
                  <td class="px-4 py-3">{{ item.group_name || '未分组' }}</td><td class="px-4 py-3">{{ item.type }}</td>
                  <td class="px-4 py-3 text-right font-medium text-emerald-600">${{ formatMoney(item.value) }}</td><td class="px-4 py-3">{{ item.status }}</td><td class="max-w-[220px] truncate px-4 py-3">{{ item.notes || '—' }}</td>
                </tr>
              </template>
              <template v-else>
                <tr v-for="item in usageItems" :key="item.id">
                  <td class="whitespace-nowrap px-4 py-3 text-gray-500">{{ formatTime(item.created_at) }}</td>
                  <td class="px-4 py-3"><div class="font-medium">{{ item.username || '—' }}</div><div class="text-xs text-gray-500">{{ item.email }}</div></td>
                  <td class="px-4 py-3">{{ item.group_name || '未分组' }}</td><td class="max-w-[180px] truncate px-4 py-3">{{ item.model }}</td><td class="max-w-[180px] truncate px-4 py-3 font-mono text-xs">{{ item.request_id || '—' }}</td>
                  <td class="px-4 py-3 text-right">{{ item.total_tokens.toLocaleString() }}</td><td class="px-4 py-3 text-right font-medium">${{ formatMoney(item.actual_cost) }}</td>
                </tr>
              </template>
              <tr v-if="!loading && pageItems.length === 0"><td colspan="7" class="px-4 py-12 text-center text-gray-500">暂无记录</td></tr>
              <tr v-if="loading"><td colspan="7" class="px-4 py-12 text-center text-gray-500">加载中...</td></tr>
            </tbody>
          </table>
        </div>
        <Pagination
          v-if="page.total > 0"
          :total="page.total"
          :page="page.page"
          :page-size="page.page_size"
          @update:page="changePage"
          @update:pageSize="changePageSize"
        />
      </div>

      <section class="space-y-4">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">财务分组</h2>
            <p class="text-sm text-gray-500 dark:text-gray-400">拖动用户到分组，保存后用于筛选和报销。</p>
          </div>
          <button class="btn btn-primary" @click="openCreateGroup">新建分组</button>
        </div>
        <div class="grid gap-4 xl:grid-cols-[260px_1fr]">
          <div
            class="min-h-[180px] rounded-lg border border-dashed border-gray-300 bg-gray-50 p-3 dark:border-dark-600 dark:bg-dark-900"
            @dragover.prevent
            @drop="dropUser(null)"
          >
            <div class="mb-3 flex items-center justify-between"><h3 class="font-medium">未分组</h3><span class="text-xs text-gray-500">{{ ungroupedUsers.length }}</span></div>
            <div class="space-y-2">
              <div v-for="user in ungroupedUsers" :key="user.id" class="cursor-grab rounded border border-gray-200 bg-white p-2 text-sm dark:border-dark-700 dark:bg-dark-800" draggable="true" @dragstart="startDrag(user.id)">
                <div class="truncate font-medium">{{ user.username || user.email }}</div><div class="truncate text-xs text-gray-500">{{ user.email }}</div>
              </div>
            </div>
          </div>
          <div class="grid gap-4 md:grid-cols-2 2xl:grid-cols-3">
            <div v-for="group in groups" :key="group.id" class="rounded-lg border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-900" @dragover.prevent @drop="dropUser(group.id)">
              <div class="mb-3 flex items-start justify-between gap-2">
                <div class="min-w-0"><h3 class="truncate font-medium">{{ group.name }}</h3><p class="truncate text-xs text-gray-500">{{ group.description || '无描述' }}</p></div>
                <div class="flex shrink-0 gap-1"><button class="btn btn-ghost btn-sm" title="编辑" @click="openEditGroup(group)">编辑</button><button class="btn btn-ghost btn-sm text-red-600" title="删除" @click="askDeleteGroup(group)">删除</button></div>
              </div>
              <div class="min-h-[96px] space-y-2">
                <div v-for="user in group.members" :key="user.id" class="cursor-grab rounded border border-gray-200 p-2 text-sm dark:border-dark-700" draggable="true" @dragstart="startDrag(user.id)">
                  <div class="truncate font-medium">{{ user.username || user.email }}</div><div class="truncate text-xs text-gray-500">{{ user.email }}</div>
                </div>
                <div v-if="group.members.length === 0" class="py-7 text-center text-xs text-gray-400">拖动用户到这里</div>
              </div>
            </div>
            <div v-if="groups.length === 0" class="rounded-lg border border-dashed border-gray-300 p-8 text-center text-sm text-gray-500 dark:border-dark-600">暂无财务分组</div>
          </div>
        </div>
      </section>
    </div>

    <BaseDialog :show="groupDialog.open" :title="groupDialog.id ? '编辑分组' : '新建分组'" width="narrow" @close="groupDialog.open = false">
      <form id="finance-group-form" class="space-y-4" @submit.prevent="submitGroup">
        <div><label class="input-label">名称</label><input v-model.trim="groupDialog.name" class="input" maxlength="100" required /></div>
        <div><label class="input-label">描述</label><textarea v-model.trim="groupDialog.description" class="input" rows="3" maxlength="500" /></div>
      </form>
      <template #footer><div class="flex justify-end gap-2"><button class="btn btn-secondary" @click="groupDialog.open = false">取消</button><button class="btn btn-primary" form="finance-group-form" :disabled="savingGroupId === -1">保存</button></div></template>
    </BaseDialog>
    <ConfirmDialog :show="deleteDialog.open" title="删除分组" message="删除分组不会删除用户，但会清空该分组成员。确认继续吗？" confirm-text="确认删除" cancel-text="取消" :danger="true" @confirm="deleteGroup" @cancel="deleteDialog.open = false" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import financeAPI, { type FinanceFilters, type FinanceGroup, type FinanceRecharge, type FinanceUsage, type FinanceUser } from '@/api/finance'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

type Tab = 'recharge' | 'usage'
const appStore = useAppStore()
const activeTab = ref<Tab>('recharge')
const loading = ref(false)
const groups = ref<FinanceGroup[]>([])
const users = ref<FinanceUser[]>([])
const rechargeItems = ref<FinanceRecharge[]>([])
const usageItems = ref<FinanceUsage[]>([])
const draggedUserId = ref<number | null>(null)
const savingGroupId = ref<number | null>(null)
const deleteDialog = reactive({ open: false, group: null as FinanceGroup | null })
const groupDialog = reactive({ open: false, id: 0, name: '', description: '' })
const today = new Date()
const currentMonth = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}`
const filters = reactive<FinanceFilters>({ start_month: currentMonth, end_month: currentMonth, user_id: undefined, group_id: undefined, search: '', page: 1, page_size: 20 })
const page = reactive({ total: 0, page: 1, page_size: 20, pages: 0, total_amount: 0 })
const tabs = [{ value: 'recharge' as const, label: '充值记录' }, { value: 'usage' as const, label: '使用记录' }]

const groupOptions = computed(() => [{ value: undefined, label: '全部分组' }, ...groups.value.map((item) => ({ value: item.id, label: item.name }))])
const userOptions = computed(() => [{ value: undefined, label: '全部用户' }, ...users.value.map((item) => ({ value: item.id, label: `${item.username || item.email} (${item.email})` }))])
const pageItems = computed(() => activeTab.value === 'recharge' ? rechargeItems.value : usageItems.value)
const groupedUserIds = computed(() => new Set(groups.value.flatMap((group) => group.members.map((member) => member.id))))
const ungroupedUsers = computed(() => users.value.filter((user) => !groupedUserIds.value.has(user.id)))
const rangeLabel = computed(() => {
  if (filters.start_month && filters.end_month) return `${filters.start_month} ~ ${filters.end_month}`
  if (filters.start_month) return `${filters.start_month} 起`
  if (filters.end_month) return `截至 ${filters.end_month}`
  return '全部月份'
})

function formatMoney(value: number) { return Number(value || 0).toFixed(4) }
function formatTime(value: string) { return new Date(value).toLocaleString() }
function query(): FinanceFilters { return { start_month: filters.start_month, end_month: filters.end_month, user_id: filters.user_id, group_id: filters.group_id, search: filters.search, page: filters.page, page_size: filters.page_size } }
function validRange() {
  if (filters.start_month && filters.end_month && filters.start_month > filters.end_month) {
    appStore.showError('起始月份不能晚于结束月份')
    return false
  }
  return true
}

async function loadRecords() {
  const result = activeTab.value === 'recharge' ? await financeAPI.recharges(query()) : await financeAPI.usage(query())
  page.total = result.total; page.page = result.page; page.page_size = result.page_size; page.pages = result.pages; page.total_amount = result.total_amount
  if (activeTab.value === 'recharge') rechargeItems.value = result.items as FinanceRecharge[]
  else usageItems.value = result.items as FinanceUsage[]
}
async function loadAll() {
  try { loading.value = true; const [groupResult, userResult] = await Promise.all([financeAPI.groups(), financeAPI.users()]); groups.value = groupResult.items; users.value = userResult.items; await loadRecords() }
  catch (error) { appStore.showError(extractApiErrorMessage(error, '财务数据加载失败')) } finally { loading.value = false }
}
async function search() { if (!validRange()) return; filters.page = 1; await loadRecords().catch((error) => appStore.showError(extractApiErrorMessage(error, '明细加载失败'))) }
function resetFilters() { Object.assign(filters, { start_month: '', end_month: '', user_id: undefined, group_id: undefined, search: '', page: 1 }); void search() }
async function switchTab(tab: Tab) { activeTab.value = tab; filters.page = 1; await search() }
async function changePage(value: number) { filters.page = value; await loadRecords() }
async function changePageSize(value: number) { filters.page_size = value; filters.page = 1; await loadRecords() }
function startDrag(userId: number) { if (savingGroupId.value === null) draggedUserId.value = userId }
async function dropUser(groupId: number | null) {
  if (savingGroupId.value !== null) return
  if (!draggedUserId.value) return
  const user = users.value.find((item) => item.id === draggedUserId.value)
  if (!user) return
  const previousGroup = groups.value.find((group) => group.members.some((member) => member.id === user.id))
  draggedUserId.value = null
  if ((previousGroup?.id ?? null) === groupId) return
  const targetGroup = groupId === null ? null : groups.value.find((group) => group.id === groupId)
  if (groupId !== null && !targetGroup) return
  for (const group of groups.value) group.members = group.members.filter((member) => member.id !== user.id)
  if (targetGroup) targetGroup.members.push({ ...user, group_id: targetGroup.id, group_name: targetGroup.name })
  const changedGroup = targetGroup ?? previousGroup
  if (!changedGroup) return
  try {
    savingGroupId.value = changedGroup.id
    await financeAPI.replaceMembers(changedGroup.id, changedGroup.members.map((member) => member.id))
    await loadAll()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '分组成员保存失败'))
    await loadAll()
  } finally {
    savingGroupId.value = null
  }
}
function openCreateGroup() { Object.assign(groupDialog, { open: true, id: 0, name: '', description: '' }) }
function openEditGroup(group: FinanceGroup) { Object.assign(groupDialog, { open: true, id: group.id, name: group.name, description: group.description }) }
async function submitGroup() {
  try { savingGroupId.value = -1; if (groupDialog.id) await financeAPI.updateGroup(groupDialog.id, { name: groupDialog.name, description: groupDialog.description }); else await financeAPI.createGroup({ name: groupDialog.name, description: groupDialog.description }); groupDialog.open = false; await loadAll() }
  catch (error) { appStore.showError(extractApiErrorMessage(error, '分组保存失败')) } finally { savingGroupId.value = null }
}
function askDeleteGroup(group: FinanceGroup) { deleteDialog.group = group; deleteDialog.open = true }
async function deleteGroup() { if (!deleteDialog.group) return; try { savingGroupId.value = -1; await financeAPI.deleteGroup(deleteDialog.group.id); deleteDialog.open = false; await loadAll() } catch (error) { appStore.showError(extractApiErrorMessage(error, '分组删除失败')) } finally { savingGroupId.value = null } }
async function downloadExport() {
  if (!validRange()) return
  try { const blob = await financeAPI.export(activeTab.value, query()); const url = URL.createObjectURL(blob); const link = document.createElement('a'); link.href = url; link.download = `finance-${activeTab.value}-${new Date().toISOString().slice(0, 10)}.csv`; link.click(); URL.revokeObjectURL(url) }
  catch (error) { appStore.showError(extractApiErrorMessage(error, '导出失败')) }
}
onMounted(loadAll)
</script>
