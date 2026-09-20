<template>
  <BaseDialog :show="show" :title="t('admin.users.bonusReclaim.title')" width="normal" @close="$emit('close')">
    <div class="space-y-5">
      <!-- 说明 -->
      <div class="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-200">
        {{ t('admin.users.bonusReclaim.desc') }}
      </div>

      <div class="flex justify-end">
        <button type="button" :disabled="loadingPreview" class="btn btn-secondary" @click="loadPreview">
          {{ loadingPreview ? t('common.loading') : t('admin.users.bonusReclaim.preview') }}
        </button>
      </div>

      <!-- 预览结果 -->
      <div v-if="preview" class="rounded-xl border border-gray-200 dark:border-dark-600">
        <div class="flex items-center justify-between border-b border-gray-200 p-4 dark:border-dark-600">
          <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('admin.users.bonusReclaim.hitCount') }}</span>
          <span class="font-bold text-gray-900 dark:text-gray-100">{{ preview.count }}</span>
        </div>
        <div class="flex items-center justify-between border-b border-gray-200 p-4 dark:border-dark-600">
          <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('admin.users.bonusReclaim.totalAmount') }}</span>
          <span class="font-bold text-gray-900 dark:text-gray-100">${{ formatBalance(preview.total_amount) }}</span>
        </div>
        <div v-if="preview.samples.length" class="max-h-56 overflow-y-auto p-2">
          <div
            v-for="s in preview.samples"
            :key="s.user_id"
            class="flex items-center justify-between px-2 py-1 text-xs"
          >
            <span class="truncate text-gray-700 dark:text-gray-300">{{ s.email }}</span>
            <span class="ml-2 shrink-0 text-gray-500 dark:text-gray-400">${{ formatBalance(s.balance) }} · {{ formatDateTime(s.created_at) }}</span>
          </div>
          <p v-if="preview.count > preview.samples.length" class="px-2 py-1 text-xs text-gray-400">
            {{ t('admin.users.bonusReclaim.sampleTruncated', { shown: preview.samples.length, total: preview.count }) }}
          </p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button class="btn btn-secondary" @click="$emit('close')">{{ t('common.cancel') }}</button>
        <button
          class="btn btn-danger"
          :disabled="!preview || preview.count === 0 || executing"
          @click="handleExecute"
        >
          {{ executing ? t('common.processing') : t('admin.users.bonusReclaim.execute', { count: preview?.count ?? 0 }) }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { BonusReclaimPreview } from '@/api/admin/users'
import { formatDateTime } from '@/utils/format'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits(['close', 'success'])
const { t } = useI18n()
const appStore = useAppStore()

const loadingPreview = ref(false)
const executing = ref(false)
const preview = ref<BonusReclaimPreview | null>(null)

watch(
  () => props.show,
  async (v) => {
    if (v) {
      preview.value = null
      await loadPreview()
    }
  }
)

const formatBalance = (value: number) => {
  if (!value) return '0.00'
  const formatted = value.toFixed(8).replace(/\.?0+$/, '')
  const parts = formatted.split('.')
  if (parts.length === 1) return formatted + '.00'
  if (parts[1].length === 1) return formatted + '0'
  return formatted
}

const loadPreview = async () => {
  loadingPreview.value = true
  try {
    preview.value = await adminAPI.users.previewBonusReclaim()
  } catch (e: any) {
    console.error('Failed to preview bonus reclaim:', e)
    appStore.showError(e.response?.data?.detail || t('common.error'))
  } finally {
    loadingPreview.value = false
  }
}

const handleExecute = async () => {
  if (!preview.value || preview.value.count === 0) return
  const ok = window.confirm(
    t('admin.users.bonusReclaim.confirm', {
      count: preview.value.count,
      amount: formatBalance(preview.value.total_amount)
    })
  )
  if (!ok) return
  executing.value = true
  try {
    const res = await adminAPI.users.executeBonusReclaim()
    appStore.showSuccess(
      t('admin.users.bonusReclaim.done', {
        count: res.reclaimed_count,
        amount: formatBalance(res.total_amount)
      })
    )
    emit('success')
    emit('close')
  } catch (e: any) {
    console.error('Failed to execute bonus reclaim:', e)
    appStore.showError(e.response?.data?.detail || t('common.error'))
  } finally {
    executing.value = false
  }
}
</script>
