import { apiClient } from './client'

export interface CanvasNetworkAccess {
  can_access: boolean
  enabled: boolean
  endpoint: string
  allowed_hosts: string[]
}

export async function getCanvasNetworkAccess(): Promise<CanvasNetworkAccess> {
  return (await apiClient.get<CanvasNetworkAccess>('/canvas/network-config')).data
}
