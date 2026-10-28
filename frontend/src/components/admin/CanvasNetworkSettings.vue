<template>
  <section class="card p-6" aria-labelledby="canvas-network-title">
    <h2 id="canvas-network-title" class="text-lg font-semibold text-gray-900 dark:text-white">无限画布线路优化</h2>
    <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">
      画布仍使用原有功能；开启后，参照图和允许域名的媒体下载会通过服务器选择的代理线路访问。当前只对管理员生效，确认线路稳定后再开放给所有用户。
    </p>
    <div class="mt-5 space-y-4">
      <label class="flex items-start gap-3 text-sm text-gray-800 dark:text-gray-200">
        <input v-model="form.enabled" type="checkbox" class="mt-1" :disabled="loading || saving || !ready" />
        <span><strong>启用画布线路中转</strong><span class="mt-1 block text-xs text-gray-500">必须选择一个现有的活跃代理线路。</span></span>
      </label>
      <label class="flex items-start gap-3 text-sm text-gray-800 dark:text-gray-200">
        <input v-model="form.visible_to_users" type="checkbox" class="mt-1" :disabled="loading || saving || !ready" />
        <span><strong>开放给所有用户</strong><span class="mt-1 block text-xs text-gray-500">关闭时仅管理员账号测试使用，普通用户继续走原有画布请求。</span></span>
      </label>
      <div>
        <label for="canvas-network-proxy" class="input-label">代理线路</label>
        <select id="canvas-network-proxy" v-model="form.proxy_id" class="input w-full" :disabled="loading || saving || !ready">
          <option :value="null">请选择活跃代理线路</option>
          <option v-for="proxy in proxies" :key="proxy.id" :value="proxy.id">{{ proxy.name }} · {{ proxy.protocol }}://{{ proxy.host }}:{{ proxy.port }}</option>
        </select>
        <p v-if="!proxies.length && ready" class="mt-1 text-xs text-amber-600">暂无活跃代理，请先在代理管理中添加并测试线路。</p>
      </div>
      <div>
        <label for="canvas-network-hosts" class="input-label">允许中转的域名</label>
        <textarea id="canvas-network-hosts" v-model="hostsText" rows="4" class="input w-full font-mono text-sm" :disabled="loading || saving || !ready" />
        <p class="mt-1 text-xs text-gray-500">每行一个域名，默认包含 GitHub、GitHub Raw、GitHub CDN、jsDelivr 和 unpkg。只会匹配这些域名及其子域名。</p>
      </div>
    </div>
    <p v-if="error" role="alert" class="mt-4 text-sm text-red-600">{{ error }}</p>
    <p v-if="saved" role="status" class="mt-4 text-sm text-green-600">画布线路配置已保存。</p>
    <div class="mt-5 flex gap-3">
      <button type="button" class="btn btn-primary" :disabled="loading || saving || !ready" @click="save">{{ saving ? '保存中...' : '保存线路配置' }}</button>
      <button type="button" class="btn btn-secondary" :disabled="loading || saving" @click="load">{{ loading ? '读取中...' : '重新读取' }}</button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { adminAPI } from '@/api'
import type { Proxy } from '@/types'
import type { CanvasNetworkSettings } from '@/api/admin/canvasNetwork'

const form = reactive<CanvasNetworkSettings>({ enabled: false, visible_to_users: false, proxy_id: null, allowed_hosts: [] })
const hostsText = ref('')
const proxies = ref<Proxy[]>([])
const loading = ref(false)
const saving = ref(false)
const ready = ref(false)
const saved = ref(false)
const error = ref('')

async function load() {
  loading.value = true; error.value = ''; saved.value = false
  try {
    const [settings, activeProxies] = await Promise.all([adminAPI.canvasNetwork.getSettings(), adminAPI.proxies.getAll()])
    Object.assign(form, settings)
    hostsText.value = settings.allowed_hosts.join('\n')
    proxies.value = activeProxies.filter(proxy => proxy.status === 'active')
    ready.value = true
  } catch {
    ready.value = false
    error.value = '无法读取画布线路配置，请重试。'
  } finally { loading.value = false }
}

async function save() {
  if (!ready.value || saving.value) return
  saving.value = true; error.value = ''; saved.value = false
  try {
    const result = await adminAPI.canvasNetwork.updateSettings({
      enabled: form.enabled,
      visible_to_users: form.visible_to_users,
      proxy_id: form.proxy_id,
      allowed_hosts: hostsText.value.split(/\r?\n/).map(value => value.trim()).filter(Boolean),
    })
    Object.assign(form, result)
    hostsText.value = result.allowed_hosts.join('\n')
    saved.value = true
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : '画布线路保存失败，请重试。'
  } finally { saving.value = false }
}

onMounted(load)
</script>
