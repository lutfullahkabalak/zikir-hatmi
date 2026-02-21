import { onBeforeUnmount, ref, watch, type Ref } from 'vue'

export type PresenceUser = { id: string; name: string }

type ServerMessage =
  | { type: 'state'; count: number; target: number }
  | { type: 'completed' }
  | { type: 'presence'; users: PresenceUser[] }

type ClientMessage =
  | { type: 'increment'; amount?: number }
  | { type: 'hello'; name?: string }
  | { type: 'set_name'; name?: string }

const getSocketUrl = (hatimId: string, token: string) => {
  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  return `${protocol}://${window.location.host}/ws/${hatimId}?token=${encodeURIComponent(token)}`
}

type ZikirSocketOptions = {
  hatimId: Ref<string | null | undefined>
  token: Ref<string | null | undefined>
  username?: Ref<string | null | undefined>
}

export const useZikirSocket = ({ hatimId, token, username }: ZikirSocketOptions) => {
  const count = ref(0)
  const target = ref(50)
  const connected = ref(false)
  const activeUsers = ref<PresenceUser[]>([])

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
      if (message.type === 'presence') {
        activeUsers.value = Array.isArray(message.users) ? message.users : []
      }
    } catch (error) {
      console.warn('WebSocket message parse error', error)
    }
  }

  const send = (payload: ClientMessage) => {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return
    }
    socket.send(JSON.stringify(payload))
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

      const name = username?.value?.trim()
      send({ type: 'hello', name: name || undefined })
    })

    socket.addEventListener('message', handleMessage)

    socket.addEventListener('close', () => {
      connected.value = false
      activeUsers.value = []
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

  const increment = (amount = 1) => {
    const nextAmount = Number.isFinite(amount) ? Math.floor(amount) : 1
    if (nextAmount <= 0) {
      return
    }
    send({ type: 'increment', amount: nextAmount })
  }

  const disconnect = () => {
    clearReconnect()
    if (socket) {
      manualClose = true
      socket.close()
      socket = null
    }
    connected.value = false
    activeUsers.value = []
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

  if (username) {
    watch(username, (next) => {
      if (!connected.value) {
        return
      }
      const name = next?.trim()
      send({ type: 'set_name', name: name || undefined })
    })
  }

  onBeforeUnmount(disconnect)

  return {
    count,
    target,
    connected,
    activeUsers,
    increment,
    setState,
    disconnect,
  }
}
