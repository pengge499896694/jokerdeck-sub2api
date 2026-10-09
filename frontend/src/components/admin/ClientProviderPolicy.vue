<template>
  <section class="card p-6" aria-labelledby="client-provider-policy-title">
    <h2 id="client-provider-policy-title" class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.settings.clientProvider.title') }}</h2>
    <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.settings.clientProvider.description') }}</p>
    <p v-if="error" role="alert" class="mt-3 text-sm text-red-600">{{ error }}</p>
    <div class="mt-4 flex flex-wrap items-center gap-3">
      <button type="button" class="btn btn-primary" :disabled="busy || !policy" @click="toggle">
        {{ busy ? t('admin.settings.clientProvider.loading') : policy?.allow_provider_switch ? t('admin.settings.clientProvider.lock') : t('admin.settings.clientProvider.unlock') }}
      </button>
      <button type="button" class="btn btn-secondary" :disabled="busy" @click="load">{{ t('admin.settings.clientProvider.refresh') }}</button>
      <span class="text-sm text-gray-500">{{ policy?.allow_provider_switch ? t('admin.settings.clientProvider.allowed') : t('admin.settings.clientProvider.locked') }}</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getClientProviderPolicy, updateClientProviderPolicy, type ClientProviderPolicy } from '@/api/admin/clientProviderPolicy'

const { t } = useI18n()
const policy = ref<ClientProviderPolicy | null>(null)
const busy = ref(false)
const error = ref('')

async function load() {
  if (busy.value) return
  busy.value = true
  error.value = ''
  try { policy.value = await getClientProviderPolicy() }
  catch { policy.value = null; error.value = t('admin.settings.clientProvider.failed') }
  finally { busy.value = false }
}

async function toggle() {
  if (busy.value || !policy.value) return
  busy.value = true
  error.value = ''
  try { policy.value = await updateClientProviderPolicy(!policy.value.allow_provider_switch) }
  catch { error.value = t('admin.settings.clientProvider.failed') }
  finally { busy.value = false }
}

onMounted(load)
</script>
