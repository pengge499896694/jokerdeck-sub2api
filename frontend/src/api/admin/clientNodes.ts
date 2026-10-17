import { apiClient } from '../client'

export interface ClientNodePolicy {
  enabled: boolean
  subscription_url: string
  backup_subscription_urls: string[]
  disabled_nodes: string[]
  subscription_endpoint?: string
}

export interface ClientNodeEntry {
  name: string
  type: string
  enabled: boolean
  latency_ms?: number
  test_message?: string
}

export interface ClientNodeSubscription {
  id: string
  name: string
  url: string
  primary: boolean
  available: boolean
  error?: string
  nodes: ClientNodeEntry[]
}

export async function getPolicy(): Promise<ClientNodePolicy> {
  const { data } = await apiClient.get<ClientNodePolicy>('/admin/settings/client-node-policy')
  return data
}

export async function updatePolicy(policy: ClientNodePolicy): Promise<ClientNodePolicy> {
  const { data } = await apiClient.put<ClientNodePolicy>('/admin/settings/client-node-policy', policy)
  return data
}

export async function listNodes(): Promise<ClientNodeSubscription[]> {
  const { data } = await apiClient.get<ClientNodeSubscription[]>('/admin/settings/client-nodes')
  return data
}

export async function testNodes(names: string[] = [], subscriptionID = ''): Promise<Array<{ subscription_id: string; name: string; success: boolean; latency_ms?: number; message?: string }>> {
  const { data } = await apiClient.post('/admin/settings/client-nodes/test', { names, subscription_id: subscriptionID })
  return data
}

export default { getPolicy, updatePolicy, listNodes, testNodes }
