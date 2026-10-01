import { apiClient } from './client'

export interface FinanceUser {
  id: number
  email: string
  username: string
  group_id?: number | null
  group_name?: string
}

export interface FinanceGroup {
  id: number
  name: string
  description: string
  sort_order: number
  members: FinanceUser[]
}

export interface FinanceRecharge {
  id: number
  user_id: number
  email: string
  username: string
  group_name: string
  type: string
  value: number
  status: string
  notes: string
  used_at?: string
  created_at: string
}

export interface FinanceUsage {
  id: number
  user_id: number
  email: string
  username: string
  group_name: string
  model: string
  request_id: string
  input_tokens: number
  output_tokens: number
  total_tokens: number
  actual_cost: number
  created_at: string
}

export interface FinancePage<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  pages: number
  total_amount: number
  kind: 'recharge' | 'usage'
}

export interface FinanceFilters {
  start_month?: string
  end_month?: string
  user_id?: number
  group_id?: number
  search?: string
  page?: number
  page_size?: number
}

const financeAPI = {
  users: (search?: string) =>
    apiClient.get<{ items: FinanceUser[] }>('/finance/users', { params: { search } }).then((r) => r.data),
  groups: () => apiClient.get<{ items: FinanceGroup[] }>('/finance/groups').then((r) => r.data),
  createGroup: (payload: { name: string; description?: string }) =>
    apiClient.post<FinanceGroup>('/finance/groups', payload).then((r) => r.data),
  updateGroup: (id: number, payload: { name: string; description?: string }) =>
    apiClient.put<FinanceGroup>(`/finance/groups/${id}`, payload).then((r) => r.data),
  deleteGroup: (id: number) => apiClient.delete(`/finance/groups/${id}`).then((r) => r.data),
  replaceMembers: (id: number, userIds: number[]) =>
    apiClient.put(`/finance/groups/${id}/members`, { user_ids: userIds }).then((r) => r.data),
  recharges: (params: FinanceFilters) =>
    apiClient.get<FinancePage<FinanceRecharge>>('/finance/recharges', { params }).then((r) => r.data),
  usage: (params: FinanceFilters) =>
    apiClient.get<FinancePage<FinanceUsage>>('/finance/usage', { params }).then((r) => r.data),
  export: (type: 'recharge' | 'usage', params: FinanceFilters) => {
    const query = { ...params, type }
    return apiClient.get<Blob>('/finance/export', { params: query, responseType: 'blob' }).then((r) => r.data)
  }
}

export default financeAPI
