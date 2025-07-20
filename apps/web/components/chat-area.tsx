'use client'

import { useEffect, useRef, useState } from 'react'
import { Send, Paperclip, Mic, StopCircle, ShoppingBag, Sparkles } from 'lucide-react'
import { ChatMessage } from '@/components/chat-message'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useChatStore } from '@/lib/store/chat'
import { useWebSocket } from '@/hooks/use-websocket'
import { cn } from '@/lib/utils'

interface ChatAreaProps {
  chatId: string | null
}

export function ChatArea({ chatId }: ChatAreaProps) {
  const [input, setInput] = useState('')
  const [isStreaming, setIsStreaming] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  
  const { messages, sendMessage, currentChat } = useChatStore()
  const { isConnected } = useWebSocket()

  const chatMessages = chatId ? messages[chatId] || [] : []

  useEffect(() => {
    // Auto-scroll to bottom when new messages arrive
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [chatMessages])

  useEffect(() => {
    // Auto-resize textarea
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`
    }
  }, [input])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!input.trim() || !chatId || isStreaming) return

    const userMessage = input.trim()
    setInput('')
    setIsStreaming(true)

    try {
      await sendMessage(chatId, userMessage)
    } catch (error) {
      console.error('Failed to send message:', error)
    } finally {
      setIsStreaming(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e as any)
    }
  }

  return (
    <div className="flex flex-col h-full">
      {/* Messages Area */}
      <ScrollArea 
        ref={scrollRef}
        className="flex-1 px-4 py-8"
      >
        <div className="max-w-3xl mx-auto">
          {chatMessages.length === 0 ? (
            <div className="text-center mt-32">
              <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-gradient-to-br from-[#4d8eff] to-[#3a6cd9] mb-6">
                <ShoppingBag className="w-10 h-10 text-white" />
              </div>
              <h1 className="text-4xl font-bold mb-4">
                <span className="text-black dark:text-white">ShopMind</span>
                <span className="text-[#4d8eff]">AI</span>
                <span className="text-gray-600 dark:text-gray-400 text-2xl font-normal ml-2">Chat</span>
              </h1>
              <p className="text-lg text-gray-600 dark:text-gray-400 mb-8">
                How can I help you shop smarter today?
              </p>
              
              {/* Suggestion chips */}
              <div className="flex flex-wrap gap-3 justify-center">
                {[
                  "Find me the best laptop under $1000",
                  "I need running shoes for marathons",
                  "Compare iPhone 15 vs Samsung S24",
                  "Gift ideas for my mom's birthday"
                ].map((suggestion, index) => (
                  <button
                    key={index}
                    onClick={() => setInput(suggestion)}
                    className="px-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-full text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-750 hover:border-[#4d8eff] transition-all duration-200"
                  >
                    {suggestion}
                  </button>
                ))}
              </div>
            </div>
          ) : (
            <div className="space-y-6">
              {chatMessages.map((message) => (
                <ChatMessage key={message.id} message={message} />
              ))}
              
              {isStreaming && (
                <ChatMessage 
                  message={{
                    id: 'streaming',
                    role: 'assistant',
                    content: '',
                    timestamp: new Date().toISOString(),
                    isStreaming: true
                  }} 
                />
              )}
            </div>
          )}
        </div>
      </ScrollArea>

      {/* Input Area */}
      <div className="border-t border-gray-200 dark:border-gray-800 px-4 py-4 bg-white dark:bg-gray-900">
        <form onSubmit={handleSubmit} className="max-w-3xl mx-auto">
          <div className="relative flex items-end gap-2">
            <div className="flex-1 relative">
              <Textarea
                ref={textareaRef}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Ask about products, compare prices, or get shopping advice..."
                className={cn(
                  "min-h-[24px] max-h-[200px] px-4 py-3 pr-12",
                  "resize-none overflow-hidden",
                  "bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700 rounded-xl",
                  "focus:outline-none focus:ring-2 focus:ring-[#4d8eff] focus:border-transparent",
                  "placeholder:text-gray-500 dark:placeholder:text-gray-400 text-gray-900 dark:text-gray-100"
                )}
                rows={1}
                disabled={isStreaming || !isConnected}
              />
              
              <div className="absolute right-2 bottom-2 flex items-center gap-1">
                <Button
                  type="button"
                  size="icon"
                  variant="ghost"
                  className="h-8 w-8 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                  disabled={isStreaming}
                >
                  <Paperclip size={18} />
                </Button>
                
                <Button
                  type="submit"
                  size="icon"
                  variant="ghost"
                  className={cn(
                    "h-8 w-8",
                    input.trim() 
                      ? "text-[#4d8eff] hover:bg-[#4d8eff]/10" 
                      : "text-gray-400 cursor-not-allowed"
                  )}
                  disabled={!input.trim() || isStreaming || !isConnected}
                >
                  {isStreaming ? (
                    <StopCircle size={18} />
                  ) : (
                    <Send size={18} />
                  )}
                </Button>
              </div>
            </div>
          </div>
          
          <div className="flex items-center justify-between mt-2">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              {isConnected ? (
                <span className="flex items-center gap-1">
                  <span className="w-2 h-2 bg-green-500 rounded-full"></span>
                  Connected to AI
                </span>
              ) : (
                <span className="flex items-center gap-1">
                  <span className="w-2 h-2 bg-red-500 rounded-full"></span>
                  Connecting...
                </span>
              )}
            </p>
            <p className="text-xs text-gray-500 dark:text-gray-400">
              Powered by advanced AI models
            </p>
          </div>
        </form>
      </div>
    </div>
  )
}