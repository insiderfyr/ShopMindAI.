'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import { useAuthStore } from '@/lib/store/auth'
import { useChatStore } from '@/lib/store/chat'

interface WebSocketMessage {
  id: string
  type: 'message:new' | 'message:update' | 'message:stream' | 'error' | 'ping'
  data: any
  timestamp: number
}

interface UseWebSocketReturn {
  isConnected: boolean
  connectionState: 'connecting' | 'connected' | 'disconnected' | 'error'
  sendMessage: (type: string, data: any) => void
  subscribe: (type: string, handler: (data: any) => void) => () => void
  metrics: {
    messagesSent: number
    messagesReceived: number
    reconnectAttempts: number
    latency: number
  }
}

const MAX_RECONNECT_ATTEMPTS = 10
const INITIAL_RECONNECT_DELAY = 1000
const MAX_RECONNECT_DELAY = 30000
const HEARTBEAT_INTERVAL = 30000
const MESSAGE_QUEUE_SIZE = 100

export function useWebSocket(): UseWebSocketReturn {
  const wsRef = useRef<WebSocket | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [connectionState, setConnectionState] = useState<UseWebSocketReturn['connectionState']>('disconnected')
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>()
  const heartbeatIntervalRef = useRef<NodeJS.Timeout>()
  const messageQueueRef = useRef<WebSocketMessage[]>([])
  const listenersRef = useRef<Map<string, Set<(data: any) => void>>>(new Map())
  const reconnectAttemptsRef = useRef(0)
  const metricsRef = useRef({
    messagesSent: 0,
    messagesReceived: 0,
    reconnectAttempts: 0,
    latency: 0,
  })
  
  const { token } = useAuthStore()
  const { addMessage, updateMessage } = useChatStore()

  // Calculate exponential backoff delay
  const getReconnectDelay = useCallback(() => {
    const delay = Math.min(
      INITIAL_RECONNECT_DELAY * Math.pow(2, reconnectAttemptsRef.current),
      MAX_RECONNECT_DELAY
    )
    return delay + Math.random() * 1000 // Add jitter
  }, [])

  // Send queued messages when reconnected
  const flushMessageQueue = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      while (messageQueueRef.current.length > 0) {
        const message = messageQueueRef.current.shift()
        if (message) {
          wsRef.current.send(JSON.stringify(message))
          metricsRef.current.messagesSent++
        }
      }
    }
  }, [])

  // Send heartbeat to keep connection alive
  const sendHeartbeat = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      const pingMessage: WebSocketMessage = {
        id: `ping-${Date.now()}`,
        type: 'ping',
        data: {},
        timestamp: Date.now(),
      }
      wsRef.current.send(JSON.stringify(pingMessage))
    }
  }, [])

  // Handle incoming messages
  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const message: WebSocketMessage = JSON.parse(event.data)
      metricsRef.current.messagesReceived++

      // Update latency for pong messages
      if (message.type === 'ping' && message.data.pong) {
        metricsRef.current.latency = Date.now() - message.timestamp
        return
      }

      // Handle different message types
      switch (message.type) {
        case 'message:new':
          addMessage(message.data.chatId, message.data.message)
          break
        case 'message:update':
          updateMessage(message.data.chatId, message.data.messageId, message.data.content)
          break
        case 'message:stream':
          updateMessage(message.data.chatId, message.data.messageId, message.data.chunk)
          break
        case 'error':
          console.error('WebSocket error:', message.data)
          break
      }

      // Notify subscribers
      const listeners = listenersRef.current.get(message.type)
      if (listeners) {
        listeners.forEach(handler => handler(message.data))
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error)
    }
  }, [addMessage, updateMessage])

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (!token || wsRef.current?.readyState === WebSocket.OPEN) {
      return
    }

    setConnectionState('connecting')

    // Build WebSocket URL with auth token
    const wsUrl = new URL(process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8001/ws')
    wsUrl.searchParams.set('token', token)

    try {
      const ws = new WebSocket(wsUrl.toString())
      wsRef.current = ws

      ws.onopen = () => {
        console.log('WebSocket connected')
        setIsConnected(true)
        setConnectionState('connected')
        reconnectAttemptsRef.current = 0
        metricsRef.current.reconnectAttempts = 0

        // Start heartbeat
        heartbeatIntervalRef.current = setInterval(sendHeartbeat, HEARTBEAT_INTERVAL)

        // Flush queued messages
        flushMessageQueue()
      }

      ws.onmessage = handleMessage

      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        setConnectionState('error')
      }

      ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason)
        setIsConnected(false)
        setConnectionState('disconnected')

        // Clear heartbeat
        if (heartbeatIntervalRef.current) {
          clearInterval(heartbeatIntervalRef.current)
        }

        // Attempt reconnection if not a normal closure
        if (event.code !== 1000 && reconnectAttemptsRef.current < MAX_RECONNECT_ATTEMPTS) {
          reconnectAttemptsRef.current++
          metricsRef.current.reconnectAttempts++
          
          const delay = getReconnectDelay()
          console.log(`Reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current})`)
          
          reconnectTimeoutRef.current = setTimeout(connect, delay)
        }
      }
    } catch (error) {
      console.error('Failed to create WebSocket:', error)
      setConnectionState('error')
    }
  }, [token, sendHeartbeat, flushMessageQueue, handleMessage, getReconnectDelay])

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    
    if (heartbeatIntervalRef.current) {
      clearInterval(heartbeatIntervalRef.current)
    }

    if (wsRef.current) {
      wsRef.current.close(1000, 'User disconnect')
      wsRef.current = null
    }

    setIsConnected(false)
    setConnectionState('disconnected')
    messageQueueRef.current = []
  }, [])

  // Send message with queuing support
  const sendMessage = useCallback((type: string, data: any) => {
    const message: WebSocketMessage = {
      id: `${type}-${Date.now()}-${Math.random()}`,
      type: type as any,
      data,
      timestamp: Date.now(),
    }

    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message))
      metricsRef.current.messagesSent++
    } else {
      // Queue message if not connected
      messageQueueRef.current.push(message)
      if (messageQueueRef.current.length > MESSAGE_QUEUE_SIZE) {
        messageQueueRef.current.shift() // Remove oldest message
      }
    }
  }, [])

  // Subscribe to message type
  const subscribe = useCallback((type: string, handler: (data: any) => void) => {
    if (!listenersRef.current.has(type)) {
      listenersRef.current.set(type, new Set())
    }
    listenersRef.current.get(type)!.add(handler)

    // Return unsubscribe function
    return () => {
      const listeners = listenersRef.current.get(type)
      if (listeners) {
        listeners.delete(handler)
        if (listeners.size === 0) {
          listenersRef.current.delete(type)
        }
      }
    }
  }, [])

  // Effect to manage connection lifecycle
  useEffect(() => {
    if (token) {
      connect()
    } else {
      disconnect()
    }

    return () => {
      disconnect()
    }
  }, [token, connect, disconnect])

  return {
    isConnected,
    connectionState,
    sendMessage,
    subscribe,
    metrics: metricsRef.current,
  }
}