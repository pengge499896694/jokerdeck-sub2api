import { beforeEach, describe, expect, it, vi } from 'vitest'
import { initializeServerCanvasRelay, withCanvasNetworkProxy } from '../../infinite-canvas/src/stores/use-config-store'
vi.mock('@/i18n', () => ({ default: { t: (key: string) => key } }))
describe('Canvas network entry isolation and authentication', () => {
  beforeEach(() => { window.name = ''; localStorage.clear(); vi.unstubAllGlobals() })
  it('does not load network settings or change asset URLs in the original canvas', async () => {
    const fetch = vi.fn()
    vi.stubGlobal('fetch', fetch)
    await initializeServerCanvasRelay()
    expect(fetch).not.toHaveBeenCalled()
    expect(withCanvasNetworkProxy('https://raw.githubusercontent.com/test/a.png')).toBe('https://raw.githubusercontent.com/test/a.png')
  })
  it('authenticates config in the optimized entry and relays only allowed asset hosts', async () => {
    window.name = 'jokerdeck-canvas-optimized'
    localStorage.setItem('auth_token', 'test-panel-token')
    const fetch = vi.fn().mockResolvedValue({ ok: true, json: async () => ({ enabled: true, endpoint: '/api/v1/canvas/network/fetch', allowed_hosts: ['raw.githubusercontent.com'] }) })
    vi.stubGlobal('fetch', fetch)
    await initializeServerCanvasRelay()
    expect(fetch).toHaveBeenCalledWith('/api/v1/canvas/network-config', expect.objectContaining({ headers: { Authorization: 'Bearer test-panel-token' }, credentials: 'same-origin' }))
    const url = 'https://raw.githubusercontent.com/test/a.png'
    expect(withCanvasNetworkProxy(url)).toBe('/api/v1/canvas/network/fetch?url=' + encodeURIComponent(url))
    expect(withCanvasNetworkProxy('https://api.example/v1/images')).toBe('https://api.example/v1/images')
    expect(withCanvasNetworkProxy('https://raw.githubusercontent.com.evil.test/a')).toBe('https://raw.githubusercontent.com.evil.test/a')
    window.name = ''
    expect(withCanvasNetworkProxy(url)).toBe(url)
  })
})
