import { apiClient } from './client'

export interface ConversationLog {
  id: number
  request_id: string
  user_id: number
  user_email: string
  api_key_id: number
  api_key_name: string
  group_id: number | null
  group_name: string
  model: string
  endpoint: string
  stream: boolean
  status_code: number
  success: boolean
  question?: string
  answer?: string
  question_bytes: number
  answer_bytes: number
  question_truncated: boolean
  answer_truncated: boolean
  input_tokens: number
  output_tokens: number
  duration_ms: number
  created_at: string
}

export interface ConversationLogFilters {
  user_id?: number
  api_key_id?: number
  group_id?: number
  user?: string
  api_key?: string
  group?: string
  model?: string
  status?: '' | 'success' | 'failed'
  start_date?: string
  end_date?: string
  timezone?: string
  page?: number
  page_size?: number
}

export interface ConversationLogPage {
  items: ConversationLog[]
  total: number
  page: number
  page_size: number
  pages: number
}

export interface ConversationLogStorageStats {
  records: number
  text_bytes: number
  table_bytes: number
  queue_dropped: number
  retention_days: number
}

export const CONVERSATION_LOGS_VIEW_ALL = 'conversation_logs.view_all'

export async function list(filters: ConversationLogFilters): Promise<ConversationLogPage> {
  const { data } = await apiClient.get<ConversationLogPage>('/conversation-logs', { params: filters })
  return data
}

export async function getById(id: number): Promise<ConversationLog> {
  const { data } = await apiClient.get<ConversationLog>(`/conversation-logs/${id}`)
  return data
}

export async function getMyMenuPermissions(): Promise<{ menu_permissions: string[]; can_view_all_conversation_logs: boolean }> {
  const { data } = await apiClient.get<{ menu_permissions: string[]; can_view_all_conversation_logs: boolean }>(
    '/conversation-logs/access'
  )
  return data
}

export async function getStorageStats(): Promise<ConversationLogStorageStats> {
  const { data } = await apiClient.get<ConversationLogStorageStats>('/admin/conversation-logs/storage')
  return data
}

export async function batchDelete(ids: number[]): Promise<{ deleted: number }> {
  const { data } = await apiClient.post<{ deleted: number }>('/admin/conversation-logs/batch-delete', { ids })
  return data
}
