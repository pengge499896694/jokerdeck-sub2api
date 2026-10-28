import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import CanvasNetworkSettings from '../CanvasNetworkSettings.vue'

const mocks = vi.hoisted(() => ({ get: vi.fn(), update: vi.fn(), proxies: vi.fn() }))
vi.mock('@/api', () => ({ adminAPI: { canvasNetwork: { getSettings: mocks.get, updateSettings: mocks.update }, proxies: { getAll: mocks.proxies } } }))

describe('CanvasNetworkSettings', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    mocks.get.mockResolvedValue({ enabled: false, visible_to_users: false, proxy_id: null, allowed_hosts: ['github.com', 'raw.githubusercontent.com'] })
    mocks.proxies.mockResolvedValue([{ id: 4, name: '国内线路', protocol: 'http', host: 'proxy.example', port: 8080, status: 'active' }])
    mocks.update.mockImplementation(async value => value)
  })

  it('loads active proxy lines and saves rollout settings without nested form submission', async () => {
    const wrapper = mount(CanvasNetworkSettings)
    await flushPromises()
    expect(wrapper.get('#canvas-network-proxy').findAll('option')).toHaveLength(2)
    await wrapper.get('#canvas-network-proxy').setValue('4')
    await wrapper.get('input[type="checkbox"]').setValue(true)
    await wrapper.get('textarea').setValue('github.com\nraw.githubusercontent.com\n')
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(mocks.update).toHaveBeenCalledWith({ enabled: true, visible_to_users: false, proxy_id: 4, allowed_hosts: ['github.com', 'raw.githubusercontent.com'] })
    expect(wrapper.text()).toContain('画布线路配置已保存')
    wrapper.unmount()
  })

  it('saves the server default route when no managed proxies exist and shows API errors', async () => {
    mocks.proxies.mockResolvedValue([])
    const wrapper = mount(CanvasNetworkSettings)
    await flushPromises()
    await wrapper.get('input[type="checkbox"]').setValue(true)
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(mocks.update).toHaveBeenCalledWith(expect.objectContaining({ enabled: true, proxy_id: null, visible_to_users: false }))
    expect(wrapper.text()).toContain('画布线路配置已保存')
    mocks.update.mockRejectedValueOnce({ status: 400, message: '所选代理不可用或已过期' })
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('所选代理不可用或已过期')
    expect(wrapper.text()).not.toContain('画布线路配置已保存')
    wrapper.unmount()
  })
})
