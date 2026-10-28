import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import RechargeView from '../RechargeView.vue'

const mocks = vi.hoisted(() => ({ get: vi.fn(), list: vi.fn() }))
vi.mock('@/api/commerceLinks', () => ({ getCommerceLinks: mocks.get }))
vi.mock('@/api/xianyuRecharge', () => ({ xianyuRechargeAPI: { listMine: mocks.list } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: vi.fn() }) }))
describe('RechargeView purchase links', () => {
  beforeEach(() => { vi.resetAllMocks(); mocks.list.mockResolvedValue([]) })
  it('reads the current destination from the server and refreshes without a page reload', async () => {
    mocks.get.mockResolvedValueOnce({ xianyu_order_url: 'https://m.tb.cn/old' })
    const wrapper = shallowMount(RechargeView, { global: { stubs: {
      AppLayout: { template: '<div><slot /></div>' }, RechargeTermsGate: { template: '<div><slot /></div>' }, Icon: true
    } } })
    await flushPromises()
    expect(wrapper.get('a[target="_blank"]').attributes('href')).toBe('https://m.tb.cn/old')
    const refresh = () => wrapper.findAll('button').find(b => b.text() === '更新下单链接')!
    mocks.get.mockResolvedValueOnce({ xianyu_order_url: 'https://m.tb.cn/new' })
    await refresh().trigger('click')
    await flushPromises()
    expect(wrapper.get('a[target="_blank"]').attributes('href')).toBe('https://m.tb.cn/new')
    mocks.get.mockRejectedValueOnce(new Error('offline'))
    await refresh().trigger('click')
    await flushPromises()
    expect(wrapper.find('a[target="_blank"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('下单链接读取失败')
    expect(mocks.list).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })
})
