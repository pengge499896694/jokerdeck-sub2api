import { afterEach, describe, expect, it, vi } from 'vitest'
import { groupProbeModels, probeGroup } from '../selfMonitor'
import type { ApiKey } from '@/types'
vi.mock('../client', () => ({buildGatewayUrl: (path: string) => path}))
const key = {key:'group-secret', group_id:7, group:{platform:'openai'}} as ApiKey
afterEach(() => vi.unstubAllGlobals())
describe('self monitor normal gateway billing path', () => {
 it('uses the group Key and a bounded model request without panel auth', async () => {
  const fetch = vi.fn().mockResolvedValue({ok:true, json:async () => ({usage:{input_tokens:4, output_tokens:2}})})
  vi.stubGlobal('fetch', fetch)
  const result = await probeGroup(key, 'gpt-test')
  expect(fetch).toHaveBeenCalledTimes(1)
  const [path, request] = fetch.mock.calls[0]
  expect(path).toBe('/v1/responses')
  expect(request.headers.Authorization).toBe('Bearer group-secret')
  expect(JSON.parse(request.body).max_output_tokens).toBe(16)
  expect(result.status).toBe('operational')
  expect(result.inputTokens).toBe(4)
 })
 it('does not retry failed billable requests', async () => {
  const fetch = vi.fn().mockResolvedValue({ok:false, status:403, json:async () => ({error:{message:'group denied'}})})
  vi.stubGlobal('fetch', fetch)
  expect((await probeGroup(key, 'gpt-test')).status).toBe('failed')
  expect(fetch).toHaveBeenCalledTimes(1)
 })
 it('model discovery never posts a billable model request', async () => {
  const fetch = vi.fn().mockResolvedValue({ok:true, json:async () => ({data:[{id:'gpt-test'},{id:'dall-e-3'}]})})
  vi.stubGlobal('fetch', fetch)
  expect(await groupProbeModels(key)).toEqual(['gpt-test'])
  expect(fetch.mock.calls[0][1].method).toBe('GET')
 })
})
