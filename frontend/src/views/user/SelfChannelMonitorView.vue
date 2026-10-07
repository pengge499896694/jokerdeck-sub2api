<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="page-header">
        <h1 class="page-title">自助渠道监控</h1>
        <p class="page-description">查看共享检测结果。打开或刷新页面不会请求上游。</p>
        <div class="mt-4 flex items-center gap-3">
          <button class="btn btn-secondary" :disabled="loading || busy" @click="reload">刷新结果</button>
          <label v-if="auth.isAdmin" class="flex items-center gap-2">
            <input type="checkbox" :checked="enabled" :disabled="loading || busy" @change="toggle(0, !enabled)">开启自助渠道监控
          </label>
        </div>
      </header>
      <div v-if="error" role="alert" class="text-red-600">{{ error }}</div>
      <p v-if="!enabled" class="text-gray-500">自助渠道监控已关闭。</p>
      <section v-if="enabled" class="card space-y-4 p-5">
        <h2 class="font-semibold">我的分组检测</h2>
        <p class="text-sm text-gray-500">使用你自己的分组 Key 请求中转，测量该分组状态与完整请求延迟。按正常分组倍率计费，结果仅自己可见。</p>
        <div class="flex flex-wrap gap-3">
          <select v-model="selectedKey" class="input" :disabled="busy" @change="loadModels">
            <option :value="0">选择分组 Key</option>
            <option v-for="key in keys" :key="key.id" :value="key.id">{{ key.group?.name || `分组 ${key.group_id}` }} · {{ key.name }}</option>
          </select>
          <select v-model="selectedModel" class="input" :disabled="busy || modelsLoading">
            <option value="">{{ modelsLoading ? '读取可用模型…' : '选择检测模型' }}</option>
            <option v-for="model in models" :key="model" :value="model">{{ model }}</option>
          </select>
          <button class="btn btn-primary" :disabled="busy || !selectedModel || !selectedKey" @click="ask('group-one')">用我的 Key 检测所选分组</button>
          <button class="btn btn-primary" :disabled="busy || !keys.length" @click="ask('group-all')">用我的 Key 检测全部分组</button>
        </div>
        <p v-if="!keys.length" class="text-sm text-gray-500">没有有效分组 Key，请先在 API Key 页面创建。</p>
        <p v-if="busy" role="status" class="text-sm">{{ progress }}</p>
        <div v-for="result in groupResults" :key="result.groupId" class="rounded-xl border border-gray-200 p-3 dark:border-dark-700">
          <strong>{{ keys.find(key => key.group_id === result.groupId)?.group?.name || `分组 ${result.groupId}` }}</strong>
          <span class="ml-3" :class="result.status === 'operational' ? 'text-green-600' : 'text-red-600'">{{ result.status === 'operational' ? '可用' : '失败' }} · {{ result.latencyMs }} ms</span>
          <p class="text-xs text-gray-500">{{ result.model }} · {{ new Date(result.checkedAt).toLocaleString() }}</p>
          <p class="text-sm">{{ result.message }}</p>
          <p v-if="result.inputTokens !== undefined" class="text-xs">输入 {{ result.inputTokens }} / 输出 {{ result.outputTokens ?? '未知' }} token</p>
        </div>
      </section>
      <section v-if="auth.isAdmin && enabled" class="card space-y-3 p-5">
        <h2 class="font-semibold">管理员上游渠道检测</h2>
        <p class="text-sm text-gray-500">使用后台配置的上游 Key，消耗平台上游额度。检测结果保存后向用户共享。</p>
        <div class="flex flex-wrap gap-3">
          <select v-model="selectedChannel" class="input" :disabled="busy">
            <option :value="0">选择上游渠道</option>
            <option v-for="entry in items.filter(entry => entry.enabled)" :key="entry.monitor.id" :value="entry.monitor.id">{{ entry.monitor.name }}</option>
          </select>
          <button class="btn btn-primary" :disabled="busy || !selectedChannel" @click="ask('channel-one')">检测所选上游并发布结果</button>
          <button class="btn btn-primary" :disabled="busy || !items.some(entry => entry.enabled)" @click="ask('channel-all')">检测全部上游并发布结果</button>
        </div>
      </section>
      <h2 class="font-semibold">管理员共享的渠道状态</h2>
      <p v-if="!loading && !items.length" class="text-gray-500">暂无开放渠道。你仍可使用自己的分组 Key 检测。</p>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <section v-for="entry in items" :key="entry.monitor.id" class="space-y-2">
          <MonitorCard :item="entry.monitor" window="7d" :availability-value="entry.monitor.availability_7d" :countdown-seconds="0" />
          <label v-if="auth.isAdmin" class="flex items-center gap-2 text-sm">
            <input type="checkbox" :checked="entry.enabled" :disabled="loading || busy" @change="toggle(entry.monitor.id, !entry.enabled)">向用户开放此渠道
          </label>
        </section>
      </div>
    </div>
    <ConfirmDialog :show="pending !== null" title="确认检测费用" :message="confirmation" confirm-text="确认并检测" @cancel="pending = null" @confirm="execute" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { apiClient } from '@/api/client'
import type { UserMonitorView } from '@/api/channelMonitor'
import { useAuthStore } from '@/stores/auth'
import AppLayout from '@/components/layout/AppLayout.vue'
import MonitorCard from '@/components/user/monitor/MonitorCard.vue'
import { extractApiErrorMessage } from '@/utils/apiError'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import * as keysAPI from '@/api/keys'
import type { ApiKey } from '@/types'
import { groupProbeModels, probeGroup, type GroupProbeResult } from '@/api/selfMonitor'

const auth = useAuthStore()
const enabled = ref(true)
const loading = ref(false)
const error = ref('')
const items = ref<Array<{monitor: UserMonitorView; enabled: boolean}>>([])
const keys = ref<ApiKey[]>([])
const selectedKey = ref(0)
const selectedModel = ref('')
const models = ref<string[]>([])
const modelsLoading = ref(false)
const selectedChannel = ref(0)
const busy = ref(false)
const progress = ref('')
const groupResults = ref<GroupProbeResult[]>([])
type Action = 'group-one' | 'group-all' | 'channel-one' | 'channel-all'
const pending = ref<Action | null>(null)
const confirmation = computed(() => pending.value?.startsWith('channel')
  ? '将使用后台配置的上游 Key 请求渠道，消耗平台上游额度，并向所有用户发布检测结果。是否继续？'
  : `将使用你自己的分组 Key 发起${pending.value === 'group-all' ? '全部分组' : '所选分组'}模型请求，会消耗 token 并按分组正常计费。批量检测每个分组一次，失败不会自动重试。是否继续？`)
function ask(action: Action) { if (!busy.value) pending.value = action }
async function loadKeys() {
  const collected: ApiKey[] = []
  let page = 1
  while (true) {
    const response = await keysAPI.list(page++, 100)
    collected.push(...response.items)
    if (collected.length >= response.total || !response.items.length) break
  }
  keys.value = collected.filter(key => key.status === 'active' && key.group_id && (!key.expires_at || Date.parse(key.expires_at) > Date.now()))
}
async function loadModels() {
  const key = keys.value.find(key => key.id === selectedKey.value)
  selectedModel.value = ''
  models.value = []
  if (!key) return
  modelsLoading.value = true
  try {
    const list = await groupProbeModels(key)
    if (key.id !== selectedKey.value) return
    models.value = list
    selectedModel.value = list[0] || ''
  } catch (err) { error.value = extractApiErrorMessage(err, '模型读取失败') }
  finally { modelsLoading.value = false }
}
async function execute() {
  const action = pending.value
  if (!action || busy.value) return
  pending.value = null
  busy.value = true
  error.value = ''
  try {
    if (action.startsWith('channel')) {
      const targets = items.value.filter(entry => entry.enabled && (action === 'channel-all' || entry.monitor.id === selectedChannel.value))
      const failed: string[] = []
      for (const [index, entry] of targets.entries()) {
        progress.value = `正在检测上游 ${index + 1}/${targets.length}：${entry.monitor.name}…`
        try {
          const { data } = await apiClient.post<{results: Array<{id: number; ok: boolean}>}>('/self-channel-monitors/run', {id: entry.monitor.id, confirmed: true}, {timeout: 130000})
          if (data.results.some(result => !result.ok)) failed.push(entry.monitor.name)
        } catch { failed.push(entry.monitor.name) }
      }
      await reload()
      if (failed.length) error.value = `以下渠道未完成检测：${failed.join('、')}`
    } else {
      const unique = new Map<number, ApiKey>()
      for (const key of keys.value) if (!unique.has(key.group_id!)) unique.set(key.group_id!, key)
      const chosen = keys.value.find(key => key.id === selectedKey.value)
      if (chosen) unique.set(chosen.group_id!, chosen)
      const targets = action === 'group-one' ? (chosen ? [chosen] : []) : [...unique.values()]
      for (const [index, key] of targets.entries()) {
        progress.value = `正在检测分组 ${index + 1}/${targets.length}…`
        let result: GroupProbeResult
        try {
          const model = key.id === selectedKey.value && selectedModel.value ? selectedModel.value : (await groupProbeModels(key))[0]
          if (!model) throw new Error('分组没有可检测的文本模型')
          result = await probeGroup(key, model)
        } catch (err) {
          result = {groupId:key.group_id!, status:'failed', latencyMs:0, checkedAt:new Date().toISOString(), model:'', message:err instanceof Error ? err.message : '检测失败'}
        }
        groupResults.value = [...groupResults.value.filter(item => item.groupId !== key.group_id), result]
      }
    }
  } catch (err) { error.value = extractApiErrorMessage(err, '检测失败') }
  finally { busy.value = false; progress.value = '' }
}
async function reload() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await apiClient.get<{enabled: boolean; items: typeof items.value}>('/self-channel-monitors')
    enabled.value = data.enabled
    items.value = data.items
  } catch (err) { error.value = extractApiErrorMessage(err, '加载渠道失败') }
  finally { loading.value = false }
}
async function toggle(id: number, active: boolean) {
  loading.value = true
  try {
    await apiClient.put('/self-channel-monitors/switch', {id, enabled: active})
    await reload()
  } catch (err) { error.value = extractApiErrorMessage(err, '更新开关失败') }
  finally { loading.value = false }
}
onMounted(async () => {
  await reload()
  try { await loadKeys() } catch (err) { error.value = extractApiErrorMessage(err, '分组 Key 加载失败') }
})
</script>
