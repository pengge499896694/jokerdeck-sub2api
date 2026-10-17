import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ClientNodesPanel from '../ClientNodesPanel.vue'

const getPolicy = vi.hoisted(() => vi.fn())
const updatePolicy = vi.hoisted(() => vi.fn())
const listNodes = vi.hoisted(() => vi.fn())
const testNodes = vi.hoisted(() => vi.fn())

vi.mock('@/api', () => ({ adminAPI: { clientNodes: { getPolicy, updatePolicy, listNodes, testNodes } } }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ locale: { value: 'zh-CN' } }) }))

beforeEach(() => {
  vi.clearAllMocks()
  getPolicy.mockResolvedValue({ enabled: true, subscription_url: 'https://example.com/primary', backup_subscription_urls: ['https://example.com/backup'], disabled_nodes: [] })
  listNodes.mockResolvedValue([
    { id: 'primary', name: '主订阅', primary: true, available: false, error: '订阅读取失败或没有可用节点', nodes: [] },
    { id: 'backup-1', name: '备用订阅 1', primary: false, available: true, nodes: [{ name: 'SG-Backup', type: 'ss', enabled: true }] },
  ])
  updatePolicy.mockImplementation(async policy => policy)
  testNodes.mockResolvedValue([{ subscription_id: 'backup-1', name: 'SG-Backup', success: true, latency_ms: 27 }])
})

describe('ClientNodesPanel', () => {
  it('shows subscription groups and their nodes when primary is unavailable', async () => {
    const wrapper = mount(ClientNodesPanel)
    await flushPromises()
    expect(wrapper.findAll('details')).toHaveLength(2)
    expect(wrapper.text()).toContain('订阅读取失败')
    expect(wrapper.findAll('details')[1].text()).toContain('SG-Backup')
    expect(wrapper.text()).not.toContain('指定用户节点')
  })

  it('tests the backup group and displays latency', async () => {
    const wrapper = mount(ClientNodesPanel)
    await flushPromises()
    await wrapper.findAll('details button')[1].trigger('click')
    await flushPromises()
    expect(testNodes).toHaveBeenCalledWith(['SG-Backup'], 'backup-1')
    expect(wrapper.findAll('details')[1].text()).toContain('27ms')
  })

  it('allows a missing primary URL when a backup is configured', async () => {
    const wrapper = mount(ClientNodesPanel)
    await flushPromises()
    await wrapper.get('#client-subscription').setValue('')
    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()
    expect(updatePolicy).toHaveBeenCalledWith(expect.objectContaining({
      subscription_url: '', backup_subscription_urls: ['https://example.com/backup'],
    }))
  })

  it('saves and reloads multiple backup URLs', async () => {
    const wrapper = mount(ClientNodesPanel)
    await flushPromises()
    await wrapper.get('#client-backup-subscriptions').setValue('https://example.com/backup-1\nhttps://example.com/backup-2')
    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()
    expect(updatePolicy).toHaveBeenCalledWith(expect.objectContaining({
      backup_subscription_urls: ['https://example.com/backup-1', 'https://example.com/backup-2'],
    }))
    expect(wrapper.get('#client-backup-subscriptions').element.value).toBe('https://example.com/backup-1\nhttps://example.com/backup-2')
    expect(wrapper.text()).toContain('配置已保存')
  })

  it('displays the API error message instead of an object when saving fails', async () => {
    updatePolicy.mockRejectedValueOnce({ status: 400, message: 'subscription URLs must be unique' })
    const wrapper = mount(ClientNodesPanel)
    await flushPromises()
    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('subscription URLs must be unique')
    expect(wrapper.text()).not.toContain('[object Object]')
  })

  it('keeps the save confirmation when refreshing the node list fails', async () => {
    listNodes.mockResolvedValueOnce([]).mockRejectedValueOnce({ message: 'subscription fetch timed out' })
    const wrapper = mount(ClientNodesPanel)
    await flushPromises()
    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('配置已保存')
    expect(wrapper.get('[role="alert"]').text()).toBe('subscription fetch timed out')
  })
})
