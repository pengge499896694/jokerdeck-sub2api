import { buildGatewayUrl } from './client'
import type { ApiKey } from '@/types'

export interface GroupProbeResult {
  groupId: number
  status: 'operational' | 'failed'
  latencyMs: number
  checkedAt: string
  model: string
  message: string
  inputTokens?: number
  outputTokens?: number
}

async function gatewayJSON(path: string, key: string, body?: object): Promise<Record<string, any>> {
  const controller = new AbortController()
  const timer = window.setTimeout(() => controller.abort(), 60000)
  try {
    // Fetch avoids the panel interceptor replacing the user's group Key with a JWT.
    const response = await fetch(buildGatewayUrl(path), {
      method: body ? 'POST' : 'GET', signal: controller.signal,
      headers: { Authorization: `Bearer ${key}`, 'x-api-key': key, 'anthropic-version': '2023-06-01', 'Content-Type': 'application/json' },
      body: body ? JSON.stringify(body) : undefined,
    })
    const data = await response.json()
    if (!response.ok || data.error) throw new Error(String(data.error?.message || data.message || `HTTP ${response.status}`).slice(0, 200))
    return data
  } finally { window.clearTimeout(timer) }
}

export async function groupProbeModels(key: ApiKey): Promise<string[]> {
  const result = await gatewayJSON('/v1/models', key.key)
  return (result.data || result.models || []).map((model: {id?: string; slug?: string}) => model.id || model.slug)
    .filter((id: unknown): id is string => typeof id === 'string' && !/image|dall-e|embedding|whisper|tts|sora|video/i.test(id))
}

export async function probeGroup(key: ApiKey, model: string): Promise<GroupProbeResult> {
  const start = performance.now()
  const result: GroupProbeResult = {groupId: key.group_id!, status: 'failed', latencyMs: 0, checkedAt: '', model, message: ''}
  try {
    const platform = key.group?.platform
    const anthropic = platform === 'anthropic' || model.startsWith('claude')
    const openai = platform === 'openai' || platform === 'composite'
    const path = anthropic ? '/v1/messages' : openai ? '/v1/responses' : '/v1/chat/completions'
    const body = anthropic
      ? {model, max_tokens: 16, messages: [{role: 'user', content: 'Reply OK.'}]}
      : openai ? {model, input: 'Reply OK.', max_output_tokens: 16, stream: false}
      : {model, max_tokens: 16, messages: [{role: 'user', content: 'Reply OK.'}], stream: false}
    const response = await gatewayJSON(path, key.key, body)
    result.status = 'operational'
    result.message = '请求成功，费用按分组正常计费，可在使用记录查看'
    result.inputTokens = response.usage?.input_tokens ?? response.usage?.prompt_tokens
    result.outputTokens = response.usage?.output_tokens ?? response.usage?.completion_tokens
  } catch (err) { result.message = err instanceof Error ? err.message : '请求失败' }
  result.latencyMs = Math.round(performance.now() - start)
  result.checkedAt = new Date().toISOString()
  return result
}
