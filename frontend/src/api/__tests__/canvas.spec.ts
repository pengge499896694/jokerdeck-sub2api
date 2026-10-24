import { describe, expect, it } from 'vitest'
import { normalizeGatewayModelId } from '@/api/canvas'

describe('normalizeGatewayModelId', () => {
  it('keeps the super-resolution alias for model mapping', () => {
    expect(normalizeGatewayModelId('[超分] gpt-image-2')).toBe('[超分] gpt-image-2')
    expect(normalizeGatewayModelId('[超分] [超分] gpt-image-2.5-sunburst')).toBe(
      '[超分] gpt-image-2.5-sunburst'
    )
  })

  it('removes unrelated display tags', () => {
    expect(normalizeGatewayModelId('[官方] gpt-image-2')).toBe('gpt-image-2')
  })
})
