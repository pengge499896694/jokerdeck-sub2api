<template>
  <div v-if="!accepted" class="mx-auto max-w-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
    <h2 class="text-lg font-semibold text-gray-900 dark:text-white">充值服务条款确认</h2>
    <p class="mt-2 text-sm text-gray-600 dark:text-gray-300">
      本服务不向中国大陆地区的个人或机构提供。继续购买前，请如实确认以下内容：
    </p>
    <div v-if="loading" class="mt-5 text-sm text-gray-500">正在检查确认状态...</div>
    <template v-else>
      <label class="mt-5 flex cursor-pointer gap-3 border border-gray-200 p-3 text-sm dark:border-dark-700">
        <input v-model="notMainlandSubject" type="checkbox" />
        <span>我确认自己不是中国大陆公民或中国大陆境内主体。</span>
      </label>
      <label class="mt-3 flex cursor-pointer gap-3 border border-gray-200 p-3 text-sm dark:border-dark-700">
        <input v-model="outsideMainland" type="checkbox" />
        <span>我确认当前所在地不在中国大陆，并自愿继续购买。</span>
      </label>
      <label class="mt-3 flex cursor-pointer gap-3 border border-gray-200 p-3 text-sm dark:border-dark-700">
        <input v-model="noInvoice" type="checkbox" />
        <span>本站不支持开具发票。如需发票或报销凭证，或不能接受本条款的任何内容，请不要向本站付款，也不要充值。</span>
      </label>
      <label for="recharge-terms-typed" class="mt-5 block text-sm text-gray-600 dark:text-gray-300">
        输入“我已知晓并确认”
      </label>
      <input id="recharge-terms-typed" v-model="typed" class="input mt-1 w-full" autocomplete="off" />
      <p class="mt-2 text-xs text-gray-500">确认仅记录您的声明，不代表平台已核验身份或所在地。</p>
      <p v-if="error" role="alert" class="mt-3 text-sm text-red-600">{{ error }}</p>
      <button class="btn btn-primary mt-5 w-full justify-center" :disabled="saving || !notMainlandSubject || !outsideMainland || !noInvoice || typed !== '我已知晓并确认'" @click="confirm">
        {{ saving ? '提交中...' : '确认并继续' }}
      </button>
    </template>
  </div>
  <slot v-else />
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { rechargeTermsAPI } from '@/api/rechargeTerms'

const accepted = ref(false)
const loading = ref(true)
const saving = ref(false)
const notMainlandSubject = ref(false)
const outsideMainland = ref(false)
const noInvoice = ref(false)
const typed = ref('')
const version = ref('')
const error = ref('')

onMounted(async () => {
  try {
    const status = await rechargeTermsAPI.status()
    version.value = status.version
    accepted.value = status.accepted
  } catch {
    error.value = '无法检查条款确认状态，请刷新页面重试。'
  } finally {
    loading.value = false
  }
})

async function confirm() {
  if (!notMainlandSubject.value || !outsideMainland.value || !noInvoice.value || typed.value !== '我已知晓并确认' || !version.value) return
  saving.value = true
  error.value = ''
  try {
    const status = await rechargeTermsAPI.accept(version.value)
    accepted.value = status.accepted
  } catch {
    error.value = '确认未保存，请检查网络后重试。'
  } finally {
    saving.value = false
  }
}
</script>
