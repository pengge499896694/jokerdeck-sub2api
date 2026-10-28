import { describe, expect, it, vi } from 'vitest'
import { commerceURL, getCommerceLinks, updateCommerceLinks } from '../commerceLinks'

const mocks = vi.hoisted(() => ({ get: vi.fn(), put: vi.fn() }))
vi.mock('../client', () => ({ apiClient: mocks }))
describe('commerce links API', () => {
  it('rejects unsafe destinations', () => {
    for (const value of ['javascript:alert(1)', '//evil.example', 'https://user:pass@evil.example', 'https://shop.example/a b', 'https://shop.example\\@evil.example', 'https://']) {
      expect(commerceURL(value)).toBe('')
    }
    expect(commerceURL(' https://m.tb.cn/new?tk=123&x=2 ')).toBe('https://m.tb.cn/new?tk=123&x=2')
  })
  it('saves the complete configuration and reads back updated URLs', async () => {
    const links = { xianyu_order_url: 'https://m.tb.cn/new', store_url: 'https://store.example/shop', store_backup_url: '' }
    mocks.put.mockResolvedValue({ data: links })
    mocks.get.mockResolvedValue({ data: links })
    expect(await updateCommerceLinks(links)).toEqual(links)
    expect(mocks.put).toHaveBeenCalledWith('/admin/settings/commerce-links', links)
    expect(await getCommerceLinks()).toEqual(links)
    expect(mocks.get).toHaveBeenCalledWith('/commerce-links', expect.any(Object))
    mocks.get.mockResolvedValueOnce({ data: { ...links, store_url: 'javascript:alert(1)' } })
    await expect(getCommerceLinks()).rejects.toThrow('链接配置无效')
  })
})
