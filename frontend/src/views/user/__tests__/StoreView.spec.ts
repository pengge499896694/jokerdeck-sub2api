import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import StoreView from '../StoreView.vue'

const mocks = vi.hoisted(() => ({ get: vi.fn(), qr: vi.fn() }))
vi.mock('@/api/commerceLinks', () => ({ getCommerceLinks: mocks.get }))
vi.mock('qrcode', () => ({ default: { toDataURL: mocks.qr } }))
const links = { xianyu_order_url: 'https://m.tb.cn/new', store_url: 'https://shop.example/main', store_backup_url: 'https://backup.example/shop' }
function mountView() {
  return shallowMount(StoreView, { global: { stubs: {
    AppLayout: { template: '<div><slot /></div>' },
    RechargeTermsGate: { template: '<div><slot /></div>' }, Icon: true
  } } })
}
describe('StoreView', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    mocks.get.mockResolvedValue({ ...links })
    mocks.qr.mockResolvedValue('data:image/png;base64,main')
  })
  it('opens the configured store independently and generates an entry QR locally', async () => {
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(false)
    expect(wrapper.get('a[target="_blank"]').attributes('href')).toBe(links.store_url)
    expect(wrapper.get('a[target="_blank"]').attributes('rel')).toContain('noopener')
    expect(mocks.qr).toHaveBeenCalledWith(links.store_url, expect.objectContaining({ margin: 3 }))
    expect(wrapper.get('img').attributes('src')).toBe('data:image/png;base64,main')
    expect(wrapper.text()).toContain('不是付款码')
    expect(wrapper.text()).toContain('1.8%')
    wrapper.unmount()
  })
  it('keeps the QR aligned with the selected backup even if an older generation finishes later', async () => {
    let finishMain!: (value: string) => void
    mocks.qr.mockImplementationOnce(() => new Promise(resolve => { finishMain = resolve }))
    mocks.qr.mockResolvedValueOnce('data:image/png;base64,backup')
    const wrapper = mountView()
    await flushPromises()
    await wrapper.findAll('button').find(b => b.text() === '备用入口')!.trigger('click')
    await flushPromises()
    finishMain('data:image/png;base64,stale')
    await flushPromises()
    expect(wrapper.get('a[target="_blank"]').attributes('href')).toBe(links.store_backup_url)
    expect(wrapper.get('img').attributes('src')).toBe('data:image/png;base64,backup')
    expect(wrapper.get('a[download]').attributes('href')).toBe('data:image/png;base64,backup')
    wrapper.unmount()
  })
  it('offers QR retry while keeping the direct link usable', async () => {
    mocks.qr.mockRejectedValueOnce(new Error('canvas unavailable'))
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain('二维码生成失败')
    expect(wrapper.get('a[target="_blank"]').attributes('href')).toBe(links.store_url)
    await wrapper.findAll('button').find(b => b.text() === '重新生成')!.trigger('click')
    await flushPromises()
    expect(wrapper.find('img').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('二维码生成失败')
    wrapper.unmount()
  })
  it('reloads changed links, clears stale destinations on errors and recovers', async () => {
    const wrapper = mountView()
    await flushPromises()
    mocks.get.mockRejectedValueOnce(new Error('offline'))
    await wrapper.findAll('button').find(b => b.text() === '更新店铺链接')!.trigger('click')
    await flushPromises()
    expect(wrapper.find('a[target="_blank"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('无法读取店铺链接')
    mocks.get.mockResolvedValueOnce({ ...links, store_url: 'https://new.example/shop', store_backup_url: '' })
    await wrapper.findAll('button').find(b => b.text() === '更新店铺链接')!.trigger('click')
    await flushPromises()
    expect(wrapper.get('a[target="_blank"]').attributes('href')).toBe('https://new.example/shop')
    expect(wrapper.text()).not.toContain('备用入口')
    expect(mocks.qr).toHaveBeenLastCalledWith('https://new.example/shop', expect.any(Object))
    wrapper.unmount()
  })
})
