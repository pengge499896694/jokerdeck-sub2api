<template>
  <AppLayout>
    <div class="space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-gray-200 p-3 text-sm dark:border-dark-700">
        <div>
          <strong>无限画布（国内优化）</strong>
          <span v-if="access && !access.enabled" class="ml-3 text-amber-600">线路中转尚未启用，当前使用原始资源地址。</span>
          <span v-else-if="access?.enabled" class="ml-3 text-gray-500">允许的外部素材由服务器代取。</span>
        </div>
        <RouterLink v-if="auth.isAdmin" to="/admin/settings?tab=canvasRoute" class="btn btn-secondary btn-sm">配置画布线路</RouterLink>
      </div>
      <p v-if="error" role="alert" class="text-sm text-red-600">{{ error }} <button type="button" class="underline" @click="load">重试</button></p>
      <p v-else-if="!access" role="status" class="p-8 text-center text-gray-500">正在读取画布线路...</p>
      <iframe v-else-if="access.can_access" name="jokerdeck-canvas-optimized" src="/infinite-canvas/?embedded=1&amp;entry=canvas&amp;network=optimized" title="无限画布（国内优化）" class="h-[calc(100dvh-210px)] min-h-[520px] w-full border-0" allow="clipboard-read; clipboard-write" />
      <p v-else class="p-8 text-center text-gray-500">此入口暂未开放，请使用原无限画布。</p>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { getCanvasNetworkAccess, type CanvasNetworkAccess } from '@/api/canvasNetwork'
const auth = useAuthStore()
const access = ref<CanvasNetworkAccess | null>(null)
const error = ref('')
async function load() {
  error.value = ''
  try { access.value = await getCanvasNetworkAccess() }
  catch { error.value = '画布线路读取失败，请重试。' }
}
onMounted(load)
</script>
