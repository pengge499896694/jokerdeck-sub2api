import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RechargeTermsGate from '../RechargeTermsGate.vue'

const { status, accept } = vi.hoisted(() => ({ status: vi.fn(), accept: vi.fn() }))
vi.mock('@/api/rechargeTerms', () => ({
  rechargeTermsAPI: { status, accept }
}))

describe('RechargeTermsGate', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    status.mockResolvedValue({ version: '2026-10-04', accepted: false })
    accept.mockResolvedValue({ version: '2026-10-04', accepted: true })
  })

  it('blocks purchase content until all three declarations are confirmed', async () => {
    const wrapper = mount(RechargeTermsGate, {
      slots: { default: '<button id="purchase">购买</button>' }
    })
    await flushPromises()
    expect(wrapper.find('#purchase').exists()).toBe(false)
    const button = wrapper.get('button')
    expect(button.attributes('disabled')).toBeDefined()

    const checkboxes = wrapper.findAll('input[type="checkbox"]')
    expect(checkboxes).toHaveLength(3)
    for (const checkbox of checkboxes) await checkbox.setValue(true)
    await wrapper.get('#recharge-terms-typed').setValue('我已知晓并确认')
    expect(button.attributes('disabled')).toBeUndefined()
    await button.trigger('click')
    await flushPromises()

    expect(accept).toHaveBeenCalledWith('2026-10-04')
    expect(wrapper.find('#purchase').exists()).toBe(true)
  })

  it('keeps purchase content hidden when the status check fails', async () => {
    status.mockRejectedValue(new Error('network'))
    const wrapper = mount(RechargeTermsGate, {
      slots: { default: '<button id="purchase">购买</button>' }
    })
    await flushPromises()
    expect(wrapper.find('#purchase').exists()).toBe(false)
    expect(wrapper.get('[role="alert"]').text()).toContain('无法检查')
  })
})
