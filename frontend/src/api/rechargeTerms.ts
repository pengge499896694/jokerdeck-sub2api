import { apiClient } from './client'

export interface RechargeTermsStatus {
  version: string
  accepted: boolean
}

export const rechargeTermsAPI = {
  async status(): Promise<RechargeTermsStatus> {
    const { data } = await apiClient.get<RechargeTermsStatus>('/recharge-terms')
    return data
  },
  async accept(version: string): Promise<RechargeTermsStatus> {
    const { data } = await apiClient.post<RechargeTermsStatus>('/recharge-terms/accept', {
      version, not_mainland_subject: true, outside_mainland: true,
      no_invoice: true, typed: '我已知晓并确认'
    })
    return data
  }
}
