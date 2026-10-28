import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import OptimizedCanvasView from '../OptimizedCanvasView.vue'
const mocks = vi.hoisted(() => ({ get: vi.fn(), auth: { isAdmin: true } }))
vi.mock('@/api/canvasNetwork', () => ({ getCanvasNetworkAccess: mocks.get }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => mocks.auth }))
function mountView() {
  return shallowMount(OptimizedCanvasView, { global: { stubs: { AppLayout: { template: '<div><slot /></div>' }, RouterLink: { template: '<a><slot /></a>' } } } })
}
describe('OptimizedCanvasView', () => {
  beforeEach(() => { vi.resetAllMocks(); mocks.auth.isAdmin = true })
  it('allows an admin to preview before enabling relay and provides its separate entry', async () => {
    mocks.get.mockResolvedValue({ can_access: true, enabled: false })
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.get('iframe').attributes('name')).toBe('jokerdeck-canvas-optimized')
    expect(wrapper.get('iframe').attributes('src')).toContain('network=optimized')
    expect(wrapper.text()).toContain('线路中转尚未启用')
    expect(wrapper.text()).toContain('配置画布线路')
    wrapper.unmount()
  })
  it('does not open the iframe for users without rollout access', async () => {
    mocks.auth.isAdmin = false
    mocks.get.mockResolvedValue({ can_access: false, enabled: false })
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(false)
    expect(wrapper.text()).toContain('暂未开放')
    wrapper.unmount()
  })
  it('does not open the iframe after a config failure and supports retry', async () => {
    mocks.get.mockRejectedValueOnce(new Error('offline'))
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(false)
    mocks.get.mockResolvedValueOnce({ can_access: true, enabled: true })
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('线路中转尚未启用')
    wrapper.unmount()
  })
})
