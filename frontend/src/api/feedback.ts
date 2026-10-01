import { apiClient } from './client'

export interface FeedbackMessage { id:number; ticket_id:number; author_type:'user'|'admin'; author_email?:string; body:string; created_at:string; context?:Record<string,unknown> }
export interface FeedbackTicket { id:number; user_id:number; user_email?:string; subject:string; category:string; status:string; priority:string; model?:string; error_code?:string; request_id?:string; context?:Record<string,unknown>; last_user_at:string; last_admin_at?:string; created_at:string; updated_at:string; messages?:FeedbackMessage[] }
export interface FeedbackPage { items: FeedbackTicket[]; total:number; page:number; page_size:number; pages:number }
export interface FeedbackPayload { subject:string; category?:string; body:string; model?:string; error_code?:string; request_id?:string; metadata?:Record<string,string> }
const feedbackAPI = {
  list: (params?: { page?:number; page_size?:number; status?:string }) => apiClient.get<FeedbackPage>('/feedback/tickets', { params }).then(r=>r.data),
  create: (payload:FeedbackPayload) => apiClient.post<FeedbackTicket>('/feedback/tickets', payload).then(r=>r.data),
  get: (id:number) => apiClient.get<FeedbackTicket>(`/feedback/tickets/${id}`).then(r=>r.data),
  reply: (id:number, payload:Partial<FeedbackPayload>) => apiClient.post<FeedbackTicket>(`/feedback/tickets/${id}/messages`, payload).then(r=>r.data),
  close: (id:number) => apiClient.post(`/feedback/tickets/${id}/close`).then(r=>r.data),
}
export default feedbackAPI
