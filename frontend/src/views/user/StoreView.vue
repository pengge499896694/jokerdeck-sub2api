<template>
  <AppLayout>
    <RechargeTermsGate>
    <div class="store-page">
      <header class="store-toolbar">
        <div class="store-heading">
          <div class="store-icon" aria-hidden="true">
            <Icon name="globe" size="lg" />
          </div>
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h1 class="text-lg font-semibold text-gray-900 dark:text-white">店铺销售</h1>
              <span class="badge badge-success">24h 无人值守</span>
            </div>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              在线选购并自动完成订单交付，平台手续费为订单金额的 1.8%。
            </p>
          </div>
        </div>

        <a
          :href="storeUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="btn btn-secondary flex-shrink-0"
        >
          <Icon name="externalLink" size="sm" class="mr-2" />
          新窗口打开
        </a>
      </header>

      <div class="store-notice" role="note">
        <Icon name="infoCircle" size="sm" class="flex-shrink-0" />
        <span>下单和售后由店铺页面处理。实际支付金额以店铺结算页为准，手续费率固定为 1.8%。</span>
      </div>

      <section class="store-frame-shell">
        <div v-if="loading" class="store-loading" aria-live="polite">
          <span class="store-spinner" aria-hidden="true"></span>
          <span>正在加载店铺...</span>
        </div>
        <iframe
          :src="storeUrl"
          title="24 小时无人值守店铺"
          class="store-frame"
          allow="clipboard-write; payment"
          referrerpolicy="strict-origin-when-cross-origin"
          @load="loading = false"
        ></iframe>
      </section>
    </div>
    </RechargeTermsGate>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import RechargeTermsGate from '@/components/user/RechargeTermsGate.vue'

const storeUrl = 'https://catfk.com/shop/AYBIRLMN'
const loading = ref(true)
</script>

<style scoped>
.store-page {
  @apply mx-auto flex max-w-[1600px] flex-col gap-3;
  height: calc(100vh - 64px - 2rem);
  min-height: 620px;
}

.store-toolbar {
  @apply flex flex-col gap-4 rounded-lg border border-gray-200 bg-white px-4 py-4 sm:flex-row sm:items-center sm:justify-between dark:border-dark-700 dark:bg-dark-900;
}

.store-heading {
  @apply flex min-w-0 items-start gap-3;
}

.store-icon {
  @apply flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-md bg-primary-50 text-primary-600 dark:bg-primary-900/20 dark:text-primary-400;
}

.store-notice {
  @apply flex items-start gap-2 rounded-md border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm text-amber-900 dark:border-amber-800/60 dark:bg-amber-900/20 dark:text-amber-200;
}

.store-frame-shell {
  @apply relative min-h-0 flex-1 overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-950;
}

.store-frame {
  display: block;
  width: 100%;
  height: 100%;
  min-height: 520px;
  border: 0;
  background: white;
}

.store-loading {
  @apply absolute inset-0 z-10 flex items-center justify-center gap-3 bg-white text-sm text-gray-500 dark:bg-dark-950 dark:text-gray-400;
}

.store-spinner {
  @apply h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent;
}

@media (max-width: 640px) {
  .store-page {
    height: calc(100vh - 64px - 1rem);
    min-height: 560px;
  }

  .store-frame {
    min-height: 460px;
  }
}
</style>
