<template>
  <AppLayout>
    <div class="space-y-6">
      <h1 class="text-2xl font-semibold">工单管理</h1>
      <div class="grid gap-4 lg:grid-cols-[360px_1fr]">
        <div class="space-y-2">
          <button
            v-for="ticket in tickets"
            :key="ticket.id"
            class="card block w-full p-4 text-left"
            :disabled="loading"
            @click="open(ticket.id)"
          >
            <div class="font-medium">#{{ ticket.id }} {{ ticket.subject }}</div>
            <div class="mt-1 text-xs text-gray-500">{{ ticket.user_email }} · {{ ticket.status }}</div>
          </button>
        </div>

        <div v-if="selected" class="card p-5">
          <h2 class="text-lg font-semibold">#{{ selected.id }} {{ selected.subject }}</h2>
          <div v-if="selected.model || selected.error_code || selected.request_id" class="mt-2 text-xs text-gray-500">
            {{ selected.model }} {{ selected.error_code }} {{ selected.request_id }}
          </div>
          <div class="mt-4 space-y-3">
            <div v-for="message in selected.messages" :key="message.id" class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
              <div class="text-xs text-gray-500">{{ message.author_type }} · {{ new Date(message.created_at).toLocaleString() }}</div>
              <div class="whitespace-pre-wrap">{{ message.body }}</div>
            </div>
          </div>
          <div class="mt-4 flex gap-2">
            <textarea v-model="reply" class="input flex-1" rows="3" :disabled="loading" />
            <button class="btn btn-primary self-end" :disabled="loading" @click="send">回复</button>
          </div>
          <div class="mt-3 flex flex-wrap gap-2">
            <button class="btn btn-secondary" :disabled="loading" @click="setStatus('pending_user')">等待用户</button>
            <button class="btn btn-secondary" :disabled="loading" @click="setStatus('resolved')">标记已解决</button>
            <button class="btn btn-danger" :disabled="loading" @click="setStatus('closed')">关闭</button>
            <button class="btn btn-danger" :disabled="loading" @click="showDeleteDialog = true">删除工单</button>
          </div>
        </div>
      </div>
    </div>

    <ConfirmDialog
      :show="showDeleteDialog"
      title="删除工单"
      message="删除后工单及全部回复不可恢复，确认继续吗？"
      confirm-text="确认删除"
      cancel-text="取消"
      :danger="true"
      @confirm="remove"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import adminFeedbackAPI from '@/api/admin/feedback'
import type { FeedbackTicket } from '@/api/feedback'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const appStore = useAppStore()
const tickets = ref<FeedbackTicket[]>([])
const selected = ref<FeedbackTicket | null>(null)
const reply = ref('')
const loading = ref(false)
const showDeleteDialog = ref(false)

async function load() {
  try {
    tickets.value = (await adminFeedbackAPI.list({ page: 1, page_size: 100 })).items
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '工单列表加载失败'))
  }
}

async function open(id: number) {
  try {
    loading.value = true
    selected.value = await adminFeedbackAPI.get(id)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '工单加载失败'))
  } finally {
    loading.value = false
  }
}

async function send() {
  if (!selected.value || !reply.value.trim()) return
  try {
    loading.value = true
    selected.value = await adminFeedbackAPI.reply(selected.value.id, { body: reply.value })
    reply.value = ''
    appStore.showSuccess('回复已发送')
    await load()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '回复发送失败'))
  } finally {
    loading.value = false
  }
}

async function setStatus(value: string) {
  if (!selected.value) return
  try {
    loading.value = true
    await adminFeedbackAPI.setStatus(selected.value.id, value)
    selected.value = await adminFeedbackAPI.get(selected.value.id)
    appStore.showSuccess('工单状态已更新')
    await load()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '工单状态更新失败'))
  } finally {
    loading.value = false
  }
}

async function remove() {
  if (!selected.value) return
  try {
    loading.value = true
    await adminFeedbackAPI.remove(selected.value.id)
    showDeleteDialog.value = false
    selected.value = null
    reply.value = ''
    appStore.showSuccess('工单已删除')
    await load()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '工单删除失败'))
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
