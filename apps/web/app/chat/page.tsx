'use client'

import { useState, useRef, useEffect } from 'react'
import { Send, Plus, User, Bot, Copy, Check, RotateCcw, ThumbsUp, ThumbsDown } from 'lucide-react'
import ReactMarkdown from 'react-markdown'

interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: Date
  isStreaming?: boolean
}

interface Conversation {
  id: string
  title: string
  messages: Message[]
  createdAt: Date
  updatedAt: Date
}

export default function ChatPage() {
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [currentConversationId, setCurrentConversationId] = useState<string | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isStreaming, setIsStreaming] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  // Get current conversation
  const currentConversation = conversations.find(c => c.id === currentConversationId)

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`
    }
  }, [input])

  // Scroll to bottom
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  // Create new conversation
  const createNewConversation = () => {
    const newConversation: Conversation = {
      id: Date.now().toString(),
      title: 'New conversation',
      messages: [],
      createdAt: new Date(),
      updatedAt: new Date()
    }
    setConversations(prev => [newConversation, ...prev])
    setCurrentConversationId(newConversation.id)
    setMessages([])
  }

  // Send message
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isStreaming) return

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: input,
      timestamp: new Date()
    }

    setMessages(prev => [...prev, userMessage])
    setInput('')
    setIsStreaming(true)

    // Update conversation title with first message
    if (messages.length === 0 && currentConversationId) {
      setConversations(prev => prev.map(conv => 
        conv.id === currentConversationId 
          ? { ...conv, title: input.slice(0, 30) + (input.length > 30 ? '...' : '') }
          : conv
      ))
    }

    // Simulate streaming response
    const assistantMessage: Message = {
      id: (Date.now() + 1).toString(),
      role: 'assistant',
      content: '',
      timestamp: new Date(),
      isStreaming: true
    }

    setMessages(prev => [...prev, assistantMessage])

    // Simulate typing
    const fullResponse = `I understand you're asking about "${input}". Let me help you with that.

This is a demonstration of the ChatGPT interface. In a real implementation, this would connect to an AI service like OpenAI's API or a self-hosted model.

Here are some key points:
- The interface supports markdown formatting
- Messages can be copied with the copy button
- You can give feedback with thumbs up/down
- Conversations are saved in the sidebar

Is there anything specific you'd like to know more about?`

    let index = 0
    const interval = setInterval(() => {
      if (index < fullResponse.length) {
        setMessages(prev => prev.map(msg => 
          msg.id === assistantMessage.id 
            ? { ...msg, content: fullResponse.slice(0, index + 1) }
            : msg
        ))
        index++
      } else {
        clearInterval(interval)
        setIsStreaming(false)
        setMessages(prev => prev.map(msg => 
          msg.id === assistantMessage.id 
            ? { ...msg, isStreaming: false }
            : msg
        ))
      }
    }, 10)
  }

  // Copy message
  const copyMessage = (content: string, messageId: string) => {
    navigator.clipboard.writeText(content)
    setCopiedId(messageId)
    setTimeout(() => setCopiedId(null), 2000)
  }

  // Handle key down
  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e as any)
    }
  }

  // Start with a new conversation if none exists
  useEffect(() => {
    if (conversations.length === 0) {
      createNewConversation()
    }
  }, [])

  return (
    <div className="flex h-screen bg-white dark:bg-gray-900">
      {/* Sidebar */}
      <div className="w-64 bg-gray-50 dark:bg-gray-950 border-r border-gray-200 dark:border-gray-800 flex flex-col">
        {/* New chat button */}
        <div className="p-3">
          <button
            onClick={createNewConversation}
            className="w-full flex items-center gap-3 px-3 py-3 rounded-md border border-gray-200 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          >
            <Plus size={16} />
            <span className="text-sm">New chat</span>
          </button>
        </div>

        {/* Conversations list */}
        <div className="flex-1 overflow-y-auto px-3">
          <div className="space-y-1">
            {conversations.map(conv => (
              <button
                key={conv.id}
                onClick={() => {
                  setCurrentConversationId(conv.id)
                  setMessages(conv.messages)
                }}
                className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
                  currentConversationId === conv.id
                    ? 'bg-gray-200 dark:bg-gray-800'
                    : 'hover:bg-gray-100 dark:hover:bg-gray-800'
                }`}
              >
                <div className="truncate">{conv.title}</div>
              </button>
            ))}
          </div>
        </div>

        {/* User section */}
        <div className="p-3 border-t border-gray-200 dark:border-gray-800">
          <div className="flex items-center gap-3 px-3 py-2">
            <div className="w-8 h-8 bg-gray-300 dark:bg-gray-700 rounded-full flex items-center justify-center">
              <User size={16} />
            </div>
            <span className="text-sm">User</span>
          </div>
        </div>
      </div>

      {/* Main chat area */}
      <div className="flex-1 flex flex-col">
        {/* Messages */}
        <div className="flex-1 overflow-y-auto">
          <div className="max-w-3xl mx-auto">
            {messages.length === 0 ? (
              <div className="h-full flex items-center justify-center">
                <div className="text-center">
                  <h1 className="text-3xl font-semibold mb-8">ChatGPT Clone</h1>
                  <p className="text-gray-500 dark:text-gray-400">Start a conversation</p>
                </div>
              </div>
            ) : (
              <div className="py-8">
                {messages.map((message) => (
                  <div
                    key={message.id}
                    className={`group ${
                      message.role === 'user' ? 'bg-white dark:bg-gray-900' : 'bg-gray-50 dark:bg-gray-800'
                    }`}
                  >
                    <div className="max-w-3xl mx-auto px-4 py-6">
                      <div className="flex gap-4">
                        {/* Avatar */}
                        <div className="flex-shrink-0">
                          {message.role === 'user' ? (
                            <div className="w-8 h-8 bg-gray-300 dark:bg-gray-700 rounded-sm flex items-center justify-center">
                              <User size={16} />
                            </div>
                          ) : (
                            <div className="w-8 h-8 bg-green-600 rounded-sm flex items-center justify-center">
                              <Bot size={16} className="text-white" />
                            </div>
                          )}
                        </div>

                        {/* Content */}
                        <div className="flex-1 space-y-2">
                          {message.isStreaming ? (
                            <div className="prose dark:prose-invert max-w-none">
                              {message.content}
                              <span className="inline-block w-2 h-4 bg-gray-800 dark:bg-gray-200 animate-pulse ml-1" />
                            </div>
                          ) : (
                            <div className="prose dark:prose-invert max-w-none">
                              <ReactMarkdown>{message.content}</ReactMarkdown>
                            </div>
                          )}

                          {/* Actions */}
                          {message.role === 'assistant' && !message.isStreaming && (
                            <div className="flex items-center gap-4 opacity-0 group-hover:opacity-100 transition-opacity">
                              <button
                                onClick={() => copyMessage(message.content, message.id)}
                                className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                              >
                                {copiedId === message.id ? (
                                  <Check size={16} />
                                ) : (
                                  <Copy size={16} />
                                )}
                              </button>
                              <button className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
                                <ThumbsUp size={16} />
                              </button>
                              <button className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
                                <ThumbsDown size={16} />
                              </button>
                              <button className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
                                <RotateCcw size={16} />
                              </button>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
                <div ref={messagesEndRef} />
              </div>
            )}
          </div>
        </div>

        {/* Input area */}
        <div className="border-t border-gray-200 dark:border-gray-800">
          <form onSubmit={handleSubmit} className="max-w-3xl mx-auto px-4 py-4">
            <div className="relative">
              <textarea
                ref={textareaRef}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Send a message..."
                rows={1}
                className="w-full resize-none rounded-md border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 px-4 py-3 pr-12 focus:outline-none focus:ring-2 focus:ring-blue-500"
                disabled={isStreaming}
              />
              <button
                type="submit"
                disabled={!input.trim() || isStreaming}
                className="absolute right-2 bottom-3 p-1 rounded-md text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 disabled:opacity-50"
              >
                <Send size={20} />
              </button>
            </div>
            <p className="text-xs text-center text-gray-400 dark:text-gray-500 mt-2">
              ChatGPT Clone can make mistakes. Consider checking important information.
            </p>
          </form>
        </div>
      </div>
    </div>
  )
}