import { describe, expect, it } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import StoreView from '../StoreView.vue'

describe('StoreView', () => {
  it('embeds the unattended store and discloses the fee rate', () => {
    const wrapper = shallowMount(StoreView, {
      global: {
        stubs: {
          AppLayout: {
            template: '<div><slot /></div>',
          },
          Icon: true,
        },
      },
    })

    const frame = wrapper.get('iframe')
    const externalLink = wrapper.get('a[target="_blank"]')

    expect(frame.attributes('src')).toBe('https://catfk.com/shop/AYBIRLMN')
    expect(frame.attributes('title')).toContain('24')
    expect(externalLink.attributes('href')).toBe('https://catfk.com/shop/AYBIRLMN')
    expect(wrapper.text()).toContain('24h 无人值守')
    expect(wrapper.text()).toContain('1.8%')
  })
})
