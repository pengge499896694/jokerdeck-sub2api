<template>
  <div class="space-y-5">
    <div class="border-b border-gray-200 pb-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ text('客户端节点', 'Client nodes') }}</h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ text('主订阅优先，不可用时自动切到备用订阅。客户端按实际代理延迟自动选节点；此页测试的是节点入口的 TCP 延迟。', 'Primary nodes are preferred, with automatic backup failover. Clients test actual proxy latency; this panel measures TCP latency to node endpoints.') }}
      </p>
    </div>

    <p v-if="error" class="text-sm text-red-600" role="alert">{{ error }}</p>
    <p v-if="message" class="text-sm text-green-700" role="status">{{ message }}</p>

    <div class="flex flex-wrap items-center gap-3">
      <label class="flex items-center gap-2 text-sm text-gray-800 dark:text-gray-200">
        <input v-model="policy.enabled" type="checkbox" :disabled="loading" />
        {{ text('允许客户端连接节点', 'Allow client node connections') }}
      </label>
      <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="reload">{{ text('刷新', 'Refresh') }}</button>
      <button type="button" class="btn btn-secondary btn-sm" :disabled="loading || testing || !subscriptions.length" @click="testNodes()">{{ testing ? text('测速中...', 'Testing...') : text('测试全部节点', 'Test all nodes') }}</button>
    </div>

    <div class="space-y-3">
      <div>
        <label for="client-subscription" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ text('主订阅地址', 'Primary subscription URL') }}</label>
        <input id="client-subscription" v-model.trim="policy.subscription_url" type="url" class="input w-full" :disabled="loading" />
      </div>
      <div>
        <label for="client-backup-subscriptions" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ text('备用订阅地址', 'Backup subscription URLs') }}</label>
        <textarea id="client-backup-subscriptions" v-model="backupSubscriptionText" rows="3" class="input w-full font-mono text-sm" :disabled="loading" :placeholder="text('每行一个 HTTPS 地址', 'One HTTPS URL per line')" />
      </div>
      <div class="flex justify-end">
        <button type="button" class="btn btn-primary btn-sm" :disabled="loading" @click="savePolicy">{{ text('保存配置', 'Save settings') }}</button>
      </div>
    </div>

    <div class="flex items-center justify-between border-b border-gray-200 pb-2 dark:border-dark-700">
      <h3 class="font-medium text-gray-900 dark:text-white">{{ text('订阅节点树', 'Subscription node tree') }} <span class="text-sm text-gray-500">{{ totalNodes }}</span></h3>
      <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="refreshNodes">{{ text('读取订阅', 'Fetch subscriptions') }}</button>
    </div>

    <div v-if="subscriptions.length" class="space-y-2">
      <details v-for="subscription in subscriptions" :key="subscription.id" open class="border border-gray-200 dark:border-dark-700">
        <summary class="flex cursor-pointer list-none items-center justify-between gap-3 bg-gray-50 px-3 py-2 dark:bg-dark-800">
          <span class="min-w-0 truncate text-sm font-medium text-gray-800 dark:text-gray-100">
            <span class="mr-2 inline-block w-16 text-xs font-semibold" :class="subscription.primary ? 'text-primary-600' : 'text-gray-500'">{{ subscription.primary ? text('主订阅', 'Primary') : text('备用', 'Backup') }}</span>
            {{ subscription.name }}
            <span class="ml-2 text-xs text-gray-500">{{ subscription.nodes.length }}</span>
          </span>
          <button type="button" class="btn btn-secondary btn-sm shrink-0" :disabled="testing || !subscription.nodes.length" @click.prevent="testSubscription(subscription)">{{ text('测试本组', 'Test group') }}</button>
        </summary>
        <div v-if="subscription.error" class="px-3 py-2 text-sm text-red-600">{{ subscription.error }}</div>
        <div v-else class="divide-y divide-gray-100 dark:divide-dark-700">
          <div v-for="node in subscription.nodes" :key="`${subscription.id}-${node.name}`" class="flex items-center justify-between gap-3 px-3 py-2 text-sm">
            <div class="min-w-0 truncate">
              <span class="mr-2 inline-block h-2 w-2 rounded-full" :class="node.test_message ? 'bg-red-500' : node.enabled ? 'bg-green-500' : 'bg-gray-400'"></span>
              <span class="font-medium text-gray-800 dark:text-gray-200">{{ node.name }}</span>
              <span class="ml-2 text-gray-500">{{ node.type }}</span>
            </div>
            <div class="flex shrink-0 items-center gap-3">
              <span v-if="node.latency_ms != null" class="text-xs" :class="node.latency_ms === fastestLatency ? 'font-semibold text-green-600' : 'text-gray-500'">{{ node.latency_ms }}ms</span>
              <span v-else-if="node.test_message" class="max-w-32 truncate text-xs text-red-600" :title="node.test_message">{{ text('连接失败', 'Failed') }}</span>
              <label class="flex items-center gap-2 text-xs text-gray-600 dark:text-gray-300">
                <input type="checkbox" :checked="node.enabled" :disabled="loading" @change="toggleNode(node.name, ($event.target as HTMLInputElement).checked)" />
                {{ text('启用', 'Enabled') }}
              </label>
            </div>
          </div>
        </div>
      </details>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api'
import type { ClientNodeSubscription } from '@/api/admin/clientNodes'
import { extractApiErrorMessage } from '@/utils/apiError'

const { locale } = useI18n()
const text = (zh: string, en: string) => locale.value.startsWith('zh') ? zh : en
const loading = ref(false)
const testing = ref(false)
const error = ref('')
const message = ref('')
const subscriptions = ref<ClientNodeSubscription[]>([])
const policy = reactive({ enabled: true, subscription_url: '', backup_subscription_urls: [] as string[], disabled_nodes: [] as string[] })
const backupSubscriptionText = ref('')

const totalNodes = computed(() => subscriptions.value.reduce((total, group) => total + group.nodes.length, 0))
const fastestLatency = computed(() => {
  const values = subscriptions.value.flatMap(group => group.nodes.map(node => node.latency_ms).filter((value): value is number => typeof value === 'number' && value > 0))
  return values.length ? Math.min(...values) : null
})

async function run(action: () => Promise<void>) {
  loading.value = true
  error.value = ''
  message.value = ''
  try { await action() } catch (cause) { error.value = extractApiErrorMessage(cause) }
  finally { loading.value = false }
}

async function reload() {
  await run(async () => {
    const current = await adminAPI.clientNodes.getPolicy()
    policy.enabled = current.enabled
    policy.subscription_url = current.subscription_url
    policy.backup_subscription_urls = current.backup_subscription_urls || []
    backupSubscriptionText.value = policy.backup_subscription_urls.join('\n')
    policy.disabled_nodes = current.disabled_nodes || []
    subscriptions.value = await adminAPI.clientNodes.listNodes()
  })
}

async function refreshNodes() {
  await run(async () => { subscriptions.value = await adminAPI.clientNodes.listNodes() })
}

async function savePolicy() {
  await run(async () => {
    policy.backup_subscription_urls = backupSubscriptionText.value.split(/\r?\n/).map(url => url.trim()).filter(Boolean)
    const saved = await adminAPI.clientNodes.updatePolicy(policy)
    policy.enabled = saved.enabled
    policy.subscription_url = saved.subscription_url
    policy.backup_subscription_urls = saved.backup_subscription_urls || []
    backupSubscriptionText.value = policy.backup_subscription_urls.join('\n')
    policy.disabled_nodes = saved.disabled_nodes || []
    message.value = text('配置已保存', 'Settings saved')
  })
  if (!error.value) {
    try { subscriptions.value = await adminAPI.clientNodes.listNodes() }
    catch (cause) { error.value = extractApiErrorMessage(cause) }
  }
}

async function toggleNode(name: string, enabled: boolean) {
  const previous = [...policy.disabled_nodes]
  policy.disabled_nodes = enabled ? previous.filter(item => item !== name) : [...previous, name]
  await savePolicy()
  if (error.value) policy.disabled_nodes = previous
}

async function testNodes(names: string[] = [], subscriptionID = '') {
  testing.value = true
  error.value = ''
  try {
    const results = await adminAPI.clientNodes.testNodes(names, subscriptionID)
    const byName = new Map(results.map(result => [`${result.subscription_id}/${result.name}`, result]))
    subscriptions.value.forEach(group => group.nodes.forEach(node => {
      const result = byName.get(`${group.id}/${node.name}`)
      if (result) {
        node.latency_ms = result.success ? result.latency_ms : undefined
        node.test_message = result.success ? undefined : result.message || text('连接失败', 'Connection failed')
      }
    }))
    message.value = text('节点测速完成，最低延迟节点已标记。', 'Node tests completed; the lowest latency node is marked.')
  } catch (cause) { error.value = extractApiErrorMessage(cause) }
  finally { testing.value = false }
}

async function testSubscription(subscription: ClientNodeSubscription) {
  await testNodes(subscription.nodes.map(node => node.name), subscription.id)
}

onMounted(reload)
</script>
