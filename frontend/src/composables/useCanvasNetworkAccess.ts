import { computed, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { getCanvasNetworkAccess } from '@/api/canvasNetwork'

const publicAccess = ref(false)
let currentToken = ''
let pending: Promise<void> | null = null

export function useCanvasNetworkAccess() {
  const auth = useAuthStore()
  const canAccessOptimizedCanvas = computed(() => auth.isAdmin || (auth.token === currentToken && publicAccess.value))
  async function refreshCanvasNetworkAccess() {
    const token = auth.token || ''
    if (!token) { publicAccess.value = false; currentToken = ''; return }
    if (pending && token === currentToken) return pending
    currentToken = token
    publicAccess.value = false
    const request = getCanvasNetworkAccess()
      .then(data => { if (auth.token === token) publicAccess.value = data.can_access })
      .catch(() => { if (auth.token === token) publicAccess.value = false })
      .finally(() => { if (pending === request) pending = null })
    pending = request
    return request
  }
  return { canAccessOptimizedCanvas, refreshCanvasNetworkAccess }
}
