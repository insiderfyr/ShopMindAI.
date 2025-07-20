'use client'

import { useEffect, useRef, useState } from 'react'
import { io, Socket } from 'socket.io-client'
import { useAuthStore } from '@/lib/store/auth'
import { useChatStore } from '@/lib/store/chat'

interface UseWebSocketReturn {
  isConnected: boolean
  socket: Socket | null
  sendMessage: (event: string, data: any) => void
  onMessage: (event: string, handler: (data: any) => void) => void
  offMessage: (event: string, handler?: (data: any) => void) => void
}

export function useWebSocket(): UseWebSocketReturn {
  const socketRef = useRef<Socket | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const { token } = useAuthStore()
  const { addMessage, updateMessage } = useChatStore()

  useEffect(() => {
    if (!token) {
      if (socketRef.current) {
        socketRef.current.disconnect()
        socketRef.current = null
      }
      return
    }

    // Initialize socket connection
    socketRef.current = io(process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080', {
      auth: { token },
      transports: ['websocket', 'polling'],
      reconnection: true,
      reconnectionAttempts: 5,
      reconnectionDelay: 1000,
    })

    const socket = socketRef.current

    // Connection handlers
    socket.on('connect', () => {
      console.log('WebSocket connected')
      setIsConnected(true)
    })

    socket.on('disconnect', () => {
      console.log('WebSocket disconnected')
      setIsConnected(false)
    })

    socket.on('connect_error', (error) => {
      console.error('WebSocket connection error:', error)
      setIsConnected(false)
    })

    // Message handlers
    socket.on('message:new', (data) => {
      addMessage(data.chatId, data.message)
    })

    socket.on('message:update', (data) => {
      updateMessage(data.chatId, data.messageId, data.content)
    })

    socket.on('message:stream', (data) => {
      updateMessage(data.chatId, data.messageId, data.chunk)
    })

    socket.on('error', (error) => {
      console.error('WebSocket error:', error)
    })

    return () => {
      socket.disconnect()
    }
  }, [token, addMessage, updateMessage])

  const sendMessage = (event: string, data: any) => {
    if (socketRef.current && socketRef.current.connected) {
      socketRef.current.emit(event, data)
    }
  }

  const onMessage = (event: string, handler: (data: any) => void) => {
    if (socketRef.current) {
      socketRef.current.on(event, handler)
    }
  }

  const offMessage = (event: string, handler?: (data: any) => void) => {
    if (socketRef.current) {
      if (handler) {
        socketRef.current.off(event, handler)
      } else {
        socketRef.current.off(event)
      }
    }
  }

  return {
    isConnected,
    socket: socketRef.current,
    sendMessage,
    onMessage,
    offMessage,
  }
}