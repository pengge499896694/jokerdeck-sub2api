<template>
  <AppLayout>
    <RechargeTermsGate>
      <div class="mx-auto max-w-5xl space-y-5">
        <header class="card flex flex-wrap items-center justify-between gap-4 p-5">
          <div>
            <div class="flex items-center gap-2">
              <h1 class="text-lg font-semibold text-gray-900 dark:text-white">店铺销售</h1>
              <span class="badge badge-success">24h 无人值守</span>
            </div>
            <p class="mt-2 text-sm text-gray-500">在线选购并自动交付，平台手续费为订单金额的 1.8%。</p>
          </div>
          <button type="button" class="btn btn-secondary" :disabled="loading" @click="loadLinks">{{ loading ? '读取中...' : '更新店铺链接' }}</button>
        </header>
        <p v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:bg-red-900/20">{{ error }}</p>
        <section v-if="links" class="card grid gap-7 p-6 md:grid-cols-[1fr_260px]">
          <div class="space-y-4">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">打开独立店铺页面完成购买</h2>
            <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">店铺在新页面中打开，避免内嵌页面的登录和支付限制。如果二维码不显示，请用系统浏览器打开店铺后重新进入订单结算页。</p>
            <div v-if="links.store_backup_url" class="flex flex-wrap gap-2" aria-label="选择店铺入口">
              <button type="button" class="btn" :class="useBackup ? 'btn-secondary' : 'btn-primary'" :aria-pressed="!useBackup" @click="useBackup = false">主入口</button>
              <button type="button" class="btn" :class="useBackup ? 'btn-primary' : 'btn-secondary'" :aria-pressed="useBackup" @click="useBackup = true">备用入口</button>
            </div>
            <a :href="storeUrl" target="_blank" rel="noopener noreferrer" class="btn btn-primary inline-flex">
              <Icon name="externalLink" size="sm" class="mr-2" />打开{{ useBackup ? '备用' : '' }}店铺
            </a>
            <div class="flex items-center gap-2">
              <input :value="storeUrl" readonly aria-label="店铺链接" class="input min-w-0 flex-1" @focus="($event.target as HTMLInputElement).select()" />
              <button type="button" class="btn btn-secondary shrink-0" @click="copyLink">复制链接</button>
            </div>
            <p v-if="copyStatus" role="status" class="text-sm text-gray-600 dark:text-gray-300">{{ copyStatus }}</p>
            <div class="rounded-lg bg-gray-50 p-4 text-sm leading-6 text-gray-600 dark:bg-dark-800 dark:text-gray-300">
              <p>打不开时，可以复制链接到系统浏览器，或切换网络后重试。{{ links.store_backup_url ? '主入口不可用时可切换备用入口。' : '' }}</p>
              <p class="mt-2">下单和售后由店铺处理，实际支付金额以结算页为准。请勿在订单状态未确认时重复付款。</p>
            </div>
          </div>
          <div class="flex flex-col items-center gap-3">
            <h2 class="text-sm font-medium text-gray-900 dark:text-white">手机扫码打开店铺</h2>
            <div class="flex min-h-[224px] w-[224px] items-center justify-center rounded-lg border border-gray-200 bg-white p-2">
              <img v-if="qrCode" :src="qrCode" alt="店铺入口二维码" width="208" height="208" />
              <span v-else class="px-4 text-center text-sm text-gray-500">{{ qrBusy ? '正在生成二维码...' : '二维码未能生成，可复制左侧链接打开。' }}</span>
            </div>
            <p class="text-center text-xs leading-5 text-gray-500">这是店铺入口码，不是付款码。选购后在店铺结算页获取付款码。</p>
            <p v-if="qrError" role="alert" class="text-xs text-red-600">{{ qrError }}</p>
            <div class="flex gap-2">
              <button type="button" class="btn btn-secondary btn-sm" :disabled="qrBusy" @click="generateQR">重新生成</button>
              <a v-if="qrCode" :href="qrCode" download="jokerapi-store.png" class="btn btn-secondary btn-sm">保存二维码</a>
            </div>
          </div>
        </section>
        <div v-else-if="loading" class="card p-8 text-center text-sm text-gray-500" role="status">正在读取店铺入口...</div>
      </div>
    </RechargeTermsGate>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import QRCode from 'qrcode'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import RechargeTermsGate from '@/components/user/RechargeTermsGate.vue'
import { getCommerceLinks, type CommerceLinks } from '@/api/commerceLinks'

const links = ref<CommerceLinks | null>(null)
const loading = ref(false)
const error = ref('')
const useBackup = ref(false)
const storeUrl = computed(() => (useBackup.value ? links.value?.store_backup_url : links.value?.store_url) || '')
const qrCode = ref('')
const qrError = ref('')
const qrBusy = ref(false)
const copyStatus = ref('')
let qrRun = 0
let disposed = false

async function loadLinks() {
  if (loading.value) return
  loading.value = true; error.value = ''
  try {
    const data = await getCommerceLinks()
    if (disposed) return
    links.value = data
    if (!data.store_backup_url) useBackup.value = false
  } catch {
    // Never offer a stale destination after a requested refresh failed.
    links.value = null
    error.value = '无法读取店铺链接，请点击“更新店铺链接”重试，或联系管理员。'
  } finally { loading.value = false }
}
async function generateQR() {
  const run = ++qrRun
  const url = storeUrl.value
  qrCode.value = ''; qrError.value = ''; qrBusy.value = Boolean(url)
  if (!url) return
  try {
    const image = await QRCode.toDataURL(url, { width: 224, margin: 3, errorCorrectionLevel: 'M', color: { dark: '#000000', light: '#ffffff' } })
    if (run === qrRun && !disposed) qrCode.value = image
  } catch {
    if (run === qrRun && !disposed) qrError.value = '二维码生成失败，请重试或复制链接。'
  } finally { if (run === qrRun && !disposed) qrBusy.value = false }
}
async function copyLink() {
  copyStatus.value = ''
  try { await navigator.clipboard.writeText(storeUrl.value); copyStatus.value = '店铺链接已复制' }
  catch { copyStatus.value = '自动复制不可用，请选中链接后手动复制。' }
}
watch(storeUrl, () => { copyStatus.value = ''; void generateQR() })
onMounted(loadLinks)
onBeforeUnmount(() => { disposed = true; qrRun++ })
</script>
