<template>
  <AppLayout>
    <div class="mx-auto max-w-5xl space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-semibold">反馈/工单</h1>
          <p class="text-sm text-gray-500">提交问题并跟进管理员回复</p>
        </div>
        <button class="btn btn-primary" :disabled="loading" @click="showCreate = true">新建工单</button>
      </div>

      <div class="grid gap-4 md:grid-cols-2">
        <button
          v-for="ticket in tickets"
          :key="ticket.id"
          class="card p-4 text-left"
          :disabled="loading"
          @click="open(ticket.id)"
        >
          <div class="flex justify-between">
            <span class="font-medium">#{{ ticket.id }} {{ ticket.subject }}</span>
            <span class="badge badge-gray">{{ ticket.status }}</span>
          </div>
          <div class="mt-2 text-xs text-gray-500">{{ ticket.category }} · {{ format(ticket.updated_at) }}</div>
          <div v-if="ticket.model || ticket.error_code || ticket.request_id" class="mt-2 text-xs text-gray-500">
            {{ ticket.model }} {{ ticket.error_code }} {{ ticket.request_id }}
          </div>
        </button>
      </div>

      <div v-if="selected" class="card p-5">
        <h2 class="text-lg font-semibold">#{{ selected.id }} {{ selected.subject }}</h2>
        <div class="mt-4 space-y-3">
          <div v-for="message in selected.messages" :key="message.id" class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
            <div class="text-xs text-gray-500">{{ message.author_type }} · {{ format(message.created_at) }}</div>
            <div class="mt-1 whitespace-pre-wrap">{{ message.body }}</div>
          </div>
        </div>
        <div class="mt-4 flex gap-2">
          <textarea v-model="reply" class="input flex-1" rows="3" placeholder="继续回复" />
          <button class="btn btn-primary self-end" :disabled="loading" @click="sendReply">发送</button>
        </div>
      </div>

      <div v-if="showCreate" class="card p-5">
        <div class="grid gap-3">
          <input v-model="form.subject" class="input" placeholder="主题" />
          <select v-model="form.category" class="input">
            <option value="other">其他</option>
            <option value="bug">Bug</option>
            <option value="billing">账单</option>
            <option value="account">账号</option>
          </select>

          <select v-model="selectedUsageID" class="input" @change="applyUsage">
            <option value="">选择账单记录（可选）</option>
            <option v-for="usage in usages" :key="usage.id" :value="String(usage.id)">
              {{ format(usage.created_at) }} · {{ usage.model }} · {{ usage.request_id || '无 Request ID' }}
            </option>
          </select>

          <select v-model="selectedErrorID" class="input" @change="applyError">
            <option value="">选择错误请求（可选）</option>
            <option v-for="error in errorRequests" :key="error.id" :value="String(error.id)">
              {{ format(error.created_at) }} · {{ error.error_code || error.category }} · {{ error.model }}
            </option>
          </select>

          <div class="grid gap-3 sm:grid-cols-3">
            <input v-model="form.model" class="input" list="feedback-models" placeholder="模型（可选择或填写）" />
            <input v-model="form.error_code" class="input" list="feedback-error-codes" placeholder="错误码（可选择或填写）" />
            <input v-model="form.request_id" class="input" list="feedback-request-ids" placeholder="Request ID（可选择或填写）" />
          </div>
          <datalist id="feedback-models">
            <option v-for="model in modelOptions" :key="model" :value="model" />
          </datalist>
          <datalist id="feedback-error-codes">
            <option v-for="code in errorCodeOptions" :key="code" :value="code" />
          </datalist>
          <datalist id="feedback-request-ids">
            <option v-for="requestID in requestIDOptions" :key="requestID" :value="requestID" />
          </datalist>

          <textarea v-model="form.body" class="input" rows="6" placeholder="请描述问题" />
          <div class="flex gap-2">
            <button class="btn btn-primary" :disabled="loading" @click="create">{{ loading ? '提交中...' : '提交' }}</button>
            <button class="btn btn-secondary" :disabled="loading" @click="showCreate = false">取消</button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import feedbackAPI, { type FeedbackTicket } from '@/api/feedback'
import usageAPI from '@/api/usage'
import type { UsageLog, UserErrorRequest } from '@/types'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const route = useRoute()
const appStore = useAppStore()
const tickets = ref<FeedbackTicket[]>([])
const usages = ref<UsageLog[]>([])
const errorRequests = ref<UserErrorRequest[]>([])
const selected = ref<FeedbackTicket | null>(null)
const showCreate = ref(false)
const loading = ref(false)
const reply = ref('')
const selectedUsageID = ref('')
const selectedErrorID = ref('')
const form = ref({
  subject: '',
  category: 'other',
  body: '',
  model: String(route.query.model || ''),
  error_code: String(route.query.error_code || ''),
  request_id: String(route.query.request_id || '')
})

const format = (value: string) => new Date(value).toLocaleString()
const modelOptions = computed(() => unique([
  ...usages.value.map(item => item.model),
  ...errorRequests.value.map(item => item.model)
]))
const errorCodeOptions = computed(() => unique(errorRequests.value.map(item => item.error_code || item.category)))
const requestIDOptions = computed(() => unique([
  ...usages.value.map(item => item.request_id),
  ...errorRequests.value.map(item => item.request_id || '')
]))

function unique(values: string[]): string[] {
  return [...new Set(values.map(value => value.trim()).filter(Boolean))]
}

async function load() {
  try {
    const [ticketResult, usageResult, errorResult] = await Promise.allSettled([
      feedbackAPI.list({ page: 1, page_size: 50 }),
      usageAPI.list(1, 100),
      usageAPI.listMyErrorRequests({ page: 1, page_size: 100 })
    ])
    if (ticketResult.status === 'fulfilled') tickets.value = ticketResult.value.items
    if (usageResult.status === 'fulfilled') usages.value = usageResult.value.items
    if (errorResult.status === 'fulfilled') errorRequests.value = errorResult.value.items
    if (ticketResult.status === 'rejected') throw ticketResult.reason
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '反馈数据加载失败'))
  }
}

function applyUsage() {
  const usage = usages.value.find(item => String(item.id) === selectedUsageID.value)
  if (!usage) return
  form.value.model = usage.model || form.value.model
  form.value.request_id = usage.request_id || form.value.request_id
}

function applyError() {
  const error = errorRequests.value.find(item => String(item.id) === selectedErrorID.value)
  if (!error) return
  form.value.model = error.model || form.value.model
  form.value.error_code = error.error_code || error.category || form.value.error_code
  form.value.request_id = error.request_id || form.value.request_id
}

async function open(id: number) {
  try {
    loading.value = true
    selected.value = await feedbackAPI.get(id)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '工单加载失败'))
  } finally {
    loading.value = false
  }
}

async function create() {
  if (!form.value.subject.trim() || !form.value.body.trim()) {
    appStore.showError('请填写主题和问题描述')
    return
  }
  try {
    loading.value = true
    await feedbackAPI.create(form.value)
    form.value = { subject: '', category: 'other', body: '', model: '', error_code: '', request_id: '' }
    selectedUsageID.value = ''
    selectedErrorID.value = ''
    showCreate.value = false
    appStore.showSuccess('工单已提交')
    await load()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '工单提交失败'))
  } finally {
    loading.value = false
  }
}

async function sendReply() {
  if (!selected.value || !reply.value.trim()) return
  try {
    loading.value = true
    selected.value = await feedbackAPI.reply(selected.value.id, { body: reply.value })
    reply.value = ''
    appStore.showSuccess('回复已发送')
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '回复发送失败'))
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
