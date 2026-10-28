import { apiClient } from '../client'

export interface CanvasNetworkSettings {
  enabled: boolean
  visible_to_users: boolean
  proxy_id: number | null
  allowed_hosts: string[]
}

export async function getSettings(): Promise<CanvasNetworkSettings> {
  const { data } = await apiClient.get<CanvasNetworkSettings>('/admin/settings/canvas-network', { headers: { 'Cache-Control': 'no-cache' } })
  return data
}

export async function updateSettings(settings: CanvasNetworkSettings): Promise<CanvasNetworkSettings> {
  const { data } = await apiClient.put<CanvasNetworkSettings>('/admin/settings/canvas-network', settings)
  return data
}

export default { getSettings, updateSettings }
