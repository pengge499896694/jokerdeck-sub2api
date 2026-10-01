import { apiClient } from '../client'
import type { FeedbackPage, FeedbackTicket, FeedbackPayload } from '../feedback'
const adminFeedbackAPI = {
  list: (params?: { page?:number; page_size?:number; status?:string }) => apiClient.get<FeedbackPage>('/admin/feedback/tickets', { params }).then(r=>r.data),
  get: (id:number) => apiClient.get<FeedbackTicket>(`/admin/feedback/tickets/${id}`).then(r=>r.data),
  reply: (id:number, payload:Partial<FeedbackPayload>) => apiClient.post<FeedbackTicket>(`/admin/feedback/tickets/${id}/messages`, payload).then(r=>r.data),
  setStatus: (id:number, status:string) => apiClient.put(`/admin/feedback/tickets/${id}/status`, { status }).then(r=>r.data),
  remove: (id:number) => apiClient.delete(`/admin/feedback/tickets/${id}`).then(r=>r.data),
}
export default adminFeedbackAPI
