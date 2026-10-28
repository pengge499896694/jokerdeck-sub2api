import { apiClient } from './client'

export interface CommerceLinks {
  xianyu_order_url: string
  store_url: string
  store_backup_url: string
}

export function commerceURL(value: string): string {
  const raw = value.trim()
  if (!/^https?:\/\//i.test(raw) || /[\s\\]/.test(raw) || raw.length > 2048) return ''
  try {
    const url = new URL(raw)
    return url.hostname && !url.username && !url.password ? raw : ''
  } catch { return '' }
}

function checkedLinks(links: CommerceLinks): CommerceLinks {
  const result = {
    xianyu_order_url: commerceURL(links.xianyu_order_url || ''),
    store_url: commerceURL(links.store_url || ''),
    store_backup_url: commerceURL(links.store_backup_url || '')
  }
  if (!result.xianyu_order_url || !result.store_url || (links.store_backup_url && !result.store_backup_url)) {
    throw new Error('链接配置无效，请联系管理员更新')
  }
  return result
}

export async function getCommerceLinks(admin = false): Promise<CommerceLinks> {
  const { data } = await apiClient.get<CommerceLinks>(admin ? '/admin/settings/commerce-links' : '/commerce-links', { headers: { 'Cache-Control': 'no-cache' } })
  return checkedLinks(data)
}

export async function updateCommerceLinks(links: CommerceLinks): Promise<CommerceLinks> {
  const { data } = await apiClient.put<CommerceLinks>('/admin/settings/commerce-links', checkedLinks(links))
  return checkedLinks(data)
}
