import { apiClient, buildGatewayUrl } from './client'
import type { ViewportTransform } from '@vue-flow/core'

export type CanvasKind = 'prompt' | 'chat' | 'image'
export interface CanvasData {
  kind: CanvasKind
  title: string
  text?: string
  model?: string
  assetId?: string
  referenceId?: string
  state?: 'idle' | 'running' | 'failed'
  error?: string
  answer?: string
}
export interface CanvasNode { id: string; type: string; position: { x: number; y: number }; data: CanvasData }
export interface CanvasEdge { id: string; source: string; target: string }
export interface CanvasDocument {
  nodes: CanvasNode[]
  edges: CanvasEdge[]
  viewport: ViewportTransform
}
export interface CanvasSummary { id: string; title: string; revision: number; updated_at?: string }
export interface CanvasBoard extends CanvasSummary { document: CanvasDocument }

// Keep [超分] as a requestable alias; account model mappings use the complete
// public ID to select the actual upstream resolution model.
export function normalizeGatewayModelId(value: string): string {
  let model = value.trim()
  let superResolution = false
  for (;;) {
    const tag = model.match(/^\[([^\]]+)\]\s*/)
    if (!tag) break
    superResolution ||= tag[1] === '超分'
    model = model.slice(tag[0].length).trimStart()
  }
  return `${superResolution ? '[超分] ' : ''}${model}`.trim()
}

export const canvasAPI = {
  async list(): Promise<CanvasSummary[]> { return (await apiClient.get('/canvas')).data },
  async create(title: string): Promise<CanvasSummary> { return (await apiClient.post('/canvas', { title })).data },
  async get(id: string): Promise<CanvasBoard> { return (await apiClient.get(`/canvas/${id}`)).data },
  async save(id: string, title: string, document: CanvasDocument, revision: number): Promise<{ revision: number }> {
    return (await apiClient.put(`/canvas/${id}`, { title, document, revision })).data
  },
  async remove(id: string): Promise<void> { await apiClient.delete(`/canvas/${id}`) },
  async upload(id: string, blob: Blob): Promise<{ id: string }> {
    const form = new FormData()
    form.append('image', blob, 'image.png')
    return (await apiClient.post(`/canvas/${id}/assets`, form, { headers: { 'Content-Type': undefined }, timeout: 60000 })).data
  },
  async asset(boardId: string, assetId: string): Promise<Blob> {
    return (await apiClient.get(`/canvas/${boardId}/assets/${assetId}`, { responseType: 'blob' })).data
  },
}

export async function gatewayRequest(key: string, path: string, payload: unknown): Promise<any> {
  const response = await fetch(buildGatewayUrl(path), {
    method: 'POST',
    headers: { Authorization: `Bearer ${key}`, 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  const body = await response.json().catch(() => ({}))
  if (!response.ok) throw new Error(body.error?.message || body.message || `请求失败 (${response.status})`)
  return body
}

export async function gatewayModels(key: string): Promise<string[]> {
  const response = await fetch(buildGatewayUrl('/v1/models'), { headers: { Authorization: `Bearer ${key}` } })
  if (!response.ok) throw new Error(`加载模型失败 (${response.status})`)
  const body = await response.json()
  const modelIDs = (body.data || [])
    .map((entry: { id?: string }): string => entry.id ? normalizeGatewayModelId(entry.id) : '')
    .filter((id: string): id is string => Boolean(id))
  return Array.from(new Set<string>(modelIDs))
}

export async function imageResultBlob(result: any): Promise<Blob> {
  const item = result.data?.[0]
  if (item?.b64_json) {
    const binary = atob(item.b64_json)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
    return new Blob([bytes], { type: 'image/png' })
  }
  if (item?.url) {
    const response = await fetch(item.url)
    if (!response.ok) throw new Error('无法下载生成结果')
    return response.blob()
  }
  throw new Error('模型没有返回图片')
}
