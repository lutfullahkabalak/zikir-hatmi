import { onBeforeUnmount, ref, watch, type Ref } from 'vue'

type ServerMessage =
  | { type: 'state'; count: number; target: number }
  | { type: 'completed' }

type ClientMessage = { type: 'increment' }

const getSocketUrl = (hatimId: string, token: string) => {
  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  return `${protocol}://${window.location.host}/ws/${hatimId}?token=${encodeURIComponent(token)}`
}

type ZikirSocketOptions = {
  hatimId: Ref<string | null | undefined>
  token: Ref<string | null | undefined>
}

export const useZikirSocket = ({ hatimId, token }: ZikirSocketOptions) => {
  const count = ref(0)
  const target = ref(50)
  const connected = ref(false)

  let socket: WebSocket | null = null
  let reconnectTimer: number | null = null
  let retryAttempt = 0
  let manualClose = false

  const clearReconnect = () => {
    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
  }

  const scheduleReconnect = () => {
    clearReconnect()
    const delay = Math.min(1000 * 2 ** retryAttempt, 20000)
    retryAttempt += 1
    reconnectTimer = window.setTimeout(() => {
      connect()
    }, delay)
  }

  const handleMessage = (event: MessageEvent) => {
    try {
      const message = JSON.parse(event.data) as ServerMessage
      if (message.type === 'state') {
        count.value = message.count
        target.value = message.target
      }
      if (message.type === 'completed') {
        count.value = target.value
      }
    } catch (error) {
      console.warn('WebSocket message parse error', error)
    }
  }

  const connect = () => {
    if (!hatimId.value || !token.value) {
      return
    }
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
      return
    }

    socket = new WebSocket(getSocketUrl(hatimId.value, token.value))

    socket.addEventListener('open', () => {
      connected.value = true
      retryAttempt = 0
      clearReconnect()
    })

    socket.addEventListener('message', handleMessage)

    socket.addEventListener('close', () => {
      connected.value = false
      if (manualClose) {
        manualClose = false
        return
      }
      scheduleReconnect()
    })

    socket.addEventListener('error', () => {
      connected.value = false
      socket?.close()
    })
  }

  const increment = () => {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return
    }
    const payload: ClientMessage = { type: 'increment' }
    socket.send(JSON.stringify(payload))
  }

  const disconnect = () => {
    clearReconnect()
    if (socket) {
      manualClose = true
      socket.close()
      socket = null
    }
    connected.value = false
  }

  const setState = (nextCount: number, nextTarget: number) => {
    count.value = nextCount
    target.value = nextTarget
  }

  watch([hatimId, token], () => {
    disconnect()
    if (hatimId.value && token.value) {
      connect()
    }
  }, { immediate: true })

  onBeforeUnmount(disconnect)

  return {
    count,
    target,
    connected,
    increment,
    setState,
    disconnect,
  }
}
