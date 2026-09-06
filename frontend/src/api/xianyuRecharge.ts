import { apiClient } from './client'
import type { PaginatedResponse } from '@/types'

export type XianyuRechargeStatus = 'pending' | 'credited' | 'rejected'

export interface XianyuRechargeOrder {
  id: number
  user_id: number
  user_email?: string
  username?: string
  xianyu_order_no: string
  paid_amount_cny: number
  note?: string
  status: XianyuRechargeStatus
  credited_amount_usd?: number
  admin_note?: string
  reviewed_by?: number
  reviewed_at?: string
  created_at: string
  updated_at: string
}

export interface CreateXianyuRechargeOrderRequest {
  xianyu_order_no: string
  paid_amount_cny: number
  note?: string
}

export async function createXianyuRechargeOrder(
  payload: CreateXianyuRechargeOrderRequest
): Promise<XianyuRechargeOrder> {
  const { data } = await apiClient.post<XianyuRechargeOrder>('/xianyu-recharge-orders', payload)
  return data
}

export async function listMyXianyuRechargeOrders(): Promise<XianyuRechargeOrder[]> {
  const { data } = await apiClient.get<XianyuRechargeOrder[]>('/xianyu-recharge-orders')
  return data
}

export async function listXianyuRechargeOrders(params: {
  page?: number
  page_size?: number
  status?: XianyuRechargeStatus | ''
  search?: string
}): Promise<PaginatedResponse<XianyuRechargeOrder>> {
  const { data } = await apiClient.get<PaginatedResponse<XianyuRechargeOrder>>(
    '/admin/xianyu-recharge-orders',
    { params }
  )
  return data
}

export async function getXianyuRechargePendingCount(): Promise<number> {
  const { data } = await apiClient.get<{ count: number }>(
    '/admin/xianyu-recharge-orders/pending-count'
  )
  return data.count
}

export async function creditXianyuRechargeOrder(
  id: number,
  amountUSD: number,
  note?: string
): Promise<XianyuRechargeOrder> {
  const { data } = await apiClient.post<XianyuRechargeOrder>(
    `/admin/xianyu-recharge-orders/${id}/credit`,
    { amount_usd: amountUSD, note }
  )
  return data
}

export async function rejectXianyuRechargeOrder(
  id: number,
  reason: string
): Promise<XianyuRechargeOrder> {
  const { data } = await apiClient.post<XianyuRechargeOrder>(
    `/admin/xianyu-recharge-orders/${id}/reject`,
    { reason }
  )
  return data
}

export const xianyuRechargeAPI = {
  create: createXianyuRechargeOrder,
  listMine: listMyXianyuRechargeOrders,
  listAdmin: listXianyuRechargeOrders,
  pendingCount: getXianyuRechargePendingCount,
  credit: creditXianyuRechargeOrder,
  reject: rejectXianyuRechargeOrder
}

export default xianyuRechargeAPI
