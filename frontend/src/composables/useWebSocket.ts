import { ref, onUnmounted } from 'vue'

export type WSMessage = {
  type: string
  showtimeId?: string
  seatLabel?: string
  status?: string
  updatedBy?: string
}

export function useWebSocket(showtimeId: string, onMessage: (msg: WSMessage) => void) {
  const connected = ref(false)
  const error = ref<string | null>(null)
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectAttempts = 0
  const MAX_RECONNECT = 5

  function connect() {
    const wsBase =
      `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`

    const url = `${wsBase}/showtimes/${showtimeId}`

    ws = new WebSocket(url)

    ws.onopen = () => {
      connected.value = true
      error.value = null
      reconnectAttempts = 0
      console.log(`🔌 WebSocket connected: ${url}`)
    }

    ws.onmessage = (event) => {
      try {
        const lines = event.data.split('\n').filter(Boolean)
        for (const line of lines) {
          const msg: WSMessage = JSON.parse(line)
          onMessage(msg)
        }
      } catch (e) {
        console.warn('WebSocket message parse error:', e)
      }
    }

    ws.onerror = () => {
      error.value = 'WebSocket connection error'
    }

    ws.onclose = () => {
      connected.value = false
      // Attempt reconnect with exponential backoff
      if (reconnectAttempts < MAX_RECONNECT) {
        const delay = Math.min(1000 * 2 ** reconnectAttempts, 16000)
        reconnectAttempts++
        console.log(`🔄 WebSocket reconnecting in ${delay}ms (attempt ${reconnectAttempts})`)
        reconnectTimer = setTimeout(connect, delay)
      }
    }
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      ws.onclose = null // prevent reconnect
      ws.close()
      ws = null
    }
    connected.value = false
  }

  connect()
  onUnmounted(disconnect)

  return { connected, error, disconnect }
}
