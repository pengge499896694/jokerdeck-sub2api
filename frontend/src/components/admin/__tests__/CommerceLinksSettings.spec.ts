import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import CommerceLinksSettings from '../CommerceLinksSettings.vue'

const mocks = vi.hoisted(() => ({ get: vi.fn(), update: vi.fn() }))
vi.mock('@/api/commerceLinks', () => ({ getCommerceLinks: mocks.get, updateCommerceLinks: mocks.update }))
const links = { xianyu_order_url: 'https://m.tb.cn/old', store_url: 'https://store.example/shop', store_backup_url: '' }
describe('CommerceLinksSettings', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    mocks.get.mockResolvedValue({ ...links })
    mocks.update.mockImplementation(async value => value)
  })
  it('reads server settings, saves all URLs with a separate button and reloads', async () => {
    const wrapper = mount(CommerceLinksSettings)
    await flushPromises()
    expect(mocks.get).toHaveBeenCalledWith(true)
    expect((wrapper.get('#commerce-xianyu').element as HTMLInputElement).value).toBe(links.xianyu_order_url)
    await wrapper.get('#commerce-xianyu').setValue('https://m.tb.cn/new')
    await wrapper.get('#commerce-backup').setValue('https://backup.example/shop')
    expect(wrapper.get('button').attributes('type')).toBe('button')
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(mocks.update).toHaveBeenCalledWith({ ...links, xianyu_order_url: 'https://m.tb.cn/new', store_backup_url: 'https://backup.example/shop' })
    expect(wrapper.text()).toContain('链接已保存')
    mocks.get.mockResolvedValueOnce({ ...links, xianyu_order_url: 'https://m.tb.cn/new' })
    await wrapper.findAll('button')[1]!.trigger('click')
    await flushPromises()
    expect((wrapper.get('#commerce-backup').element as HTMLInputElement).value).toBe('')
    wrapper.unmount()
  })
  it('prevents overwriting unknown configuration after load failure and retains edits on save failure', async () => {
    mocks.get.mockRejectedValueOnce(new Error('offline'))
    const wrapper = mount(CommerceLinksSettings)
    await flushPromises()
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('无法读取链接配置')
    await wrapper.findAll('button')[1]!.trigger('click')
    await flushPromises()
    await wrapper.get('#commerce-xianyu').setValue('https://m.tb.cn/new')
    mocks.update.mockRejectedValueOnce(new Error('保存失败'))
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.text()).not.toContain('链接已保存')
    expect(wrapper.text()).toContain('保存失败')
    expect((wrapper.get('#commerce-xianyu').element as HTMLInputElement).value).toBe('https://m.tb.cn/new')
    wrapper.unmount()
  })
})
