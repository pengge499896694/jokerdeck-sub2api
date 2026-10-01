<template>
  <AppLayout>
    <div class="mx-auto max-w-[950px]">
      <ProfileBalanceNotifyCard
        v-if="user && enabled"
        :enabled="user.balance_notify_enabled ?? true"
        :threshold="user.balance_notify_threshold"
        :extra-emails="user.balance_notify_extra_emails ?? []"
        :system-default-threshold="defaultThreshold"
        :user-email="user.email"
      />
      <p v-else-if="!loading" class="text-sm text-gray-500">当前站点未开启余额不足邮箱提醒。</p>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import ProfileBalanceNotifyCard from '@/components/user/profile/ProfileBalanceNotifyCard.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const appStore = useAppStore()
const authStore = useAuthStore()
const user = computed(() => authStore.user)
const enabled = ref(false)
const defaultThreshold = ref(0)
const loading = ref(true)

onMounted(async () => {
  try {
    const [settings] = await Promise.all([appStore.fetchPublicSettings(), authStore.refreshUser()])
    enabled.value = settings?.balance_low_notify_enabled ?? false
    defaultThreshold.value = settings?.balance_low_notify_threshold ?? 0
  } catch (error) {
    console.error('Failed to load balance notification settings:', error)
  } finally {
    loading.value = false
  }
})
</script>
