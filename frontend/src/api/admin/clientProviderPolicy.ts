import { apiClient } from '../client'

export interface ClientProviderPolicy {
  allow_provider_switch: boolean
  require_recharge: boolean
  external_provider_support: boolean
}

export async function getClientProviderPolicy(): Promise<ClientProviderPolicy> {
  const { data } = await apiClient.get<ClientProviderPolicy>('/admin/settings/client-provider-policy')
  return data
}

export async function updateClientProviderPolicy(allowed: boolean): Promise<ClientProviderPolicy> {
  const { data } = await apiClient.put<ClientProviderPolicy>('/admin/settings/client-provider-policy', { allow_provider_switch: allowed })
  return data
}
