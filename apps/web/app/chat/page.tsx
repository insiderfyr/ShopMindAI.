'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import { Send, Plus, User, Bot, Copy, Check, ThumbsUp, ThumbsDown, RotateCcw, ChevronDown } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'

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

const AVAILABLE_STORES = [
  { id: 'all', name: 'All stores', icon: '🌐' },
  { id: 'amazon', name: 'Amazon', icon: '📦' },
  { id: 'bestbuy', name: 'Best Buy', icon: '🟡' },
  { id: 'walmart', name: 'Walmart', icon: '🔵' },
  { id: 'target', name: 'Target', icon: '🎯' },
  { id: 'ebay', name: 'eBay', icon: '🛒' },
  { id: 'newegg', name: 'Newegg', icon: '🥚' },
  { id: 'costco', name: 'Costco', icon: '🏪' },
]

export default function ChatPage() {
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [currentConversationId, setCurrentConversationId] = useState<string | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isStreaming, setIsStreaming] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const [selectedStore, setSelectedStore] = useState('all')
  const [showStoreDropdown, setShowStoreDropdown] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const dropdownRef = useRef<HTMLDivElement>(null)

  // Get current conversation
  const currentConversation = conversations.find(c => c.id === currentConversationId)

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`
    }
  }, [input])

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowStoreDropdown(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

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
      title: 'New shopping search',
      messages: [],
      createdAt: new Date(),
      updatedAt: new Date()
    }
    setConversations(prev => [newConversation, ...prev])
    setCurrentConversationId(newConversation.id)
    setMessages([])
  }

  // Copy message
  const copyMessage = (content: string, messageId: string) => {
    navigator.clipboard.writeText(content)
    setCopiedId(messageId)
    setTimeout(() => setCopiedId(null), 2000)
  }

  // Handle submit
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!input.trim() || isStreaming) return

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: input.trim(),
      timestamp: new Date()
    }

    // Update messages
    setMessages(prev => [...prev, userMessage])
    setInput('')
    setIsStreaming(true)

    // Create conversation if needed
    if (!currentConversationId) {
      createNewConversation()
    }

    // Simulate AI response
    const assistantMessage: Message = {
      id: (Date.now() + 1).toString(),
      role: 'assistant',
      content: '',
      timestamp: new Date(),
      isStreaming: true
    }

    setMessages(prev => [...prev, assistantMessage])

    // Simulate streaming response
    const searchingStore = selectedStore === 'all' ? 'all major retailers' : AVAILABLE_STORES.find(s => s.id === selectedStore)?.name
    const fullResponse = `I'll search ${searchingStore} for "${userMessage.content}". Here's what I found:\n\n**Best Deals:**\n\n1. **Product Name** - $XX.99\n   • Key feature 1\n   • Key feature 2\n   • ${searchingStore === 'all major retailers' ? 'Available at multiple stores' : `Available at ${searchingStore}`}\n   • ⭐ 4.5/5 (1,234 reviews)\n\n2. **Alternative Product** - $YY.99\n   • Different feature\n   • Another benefit\n   • Free shipping available\n\nWould you like me to search for specific features or compare prices across different stores?`

    let currentText = ''
    const words = fullResponse.split(' ')
    
    for (let i = 0; i < words.length; i++) {
      await new Promise(resolve => setTimeout(resolve, 50))
      currentText += words[i] + ' '
      
      setMessages(prev => prev.map(msg => 
        msg.id === assistantMessage.id 
          ? { ...msg, content: currentText.trim(), isStreaming: i < words.length - 1 }
          : msg
      ))
    }

    setIsStreaming(false)

    // Update conversation title
    if (currentConversation && currentConversation.messages.length === 0) {
      setConversations(prev => prev.map(conv => 
        conv.id === currentConversationId
          ? { ...conv, title: userMessage.content.slice(0, 30) + '...', messages: [...messages, userMessage, { ...assistantMessage, content: fullResponse, isStreaming: false }] }
          : conv
      ))
    }
  }

  // Handle key down
  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e as any)
    }
  }

  // Group conversations by date
  const groupConversationsByDate = () => {
    const today = new Date()
    const yesterday = new Date(today)
    yesterday.setDate(yesterday.getDate() - 1)
    const lastWeek = new Date(today)
    lastWeek.setDate(lastWeek.getDate() - 7)

    const groups: { [key: string]: Conversation[] } = {
      Today: [],
      Yesterday: [],
      'Previous 7 days': [],
      Older: []
    }

    conversations.forEach(conv => {
      const convDate = new Date(conv.createdAt)
      if (convDate.toDateString() === today.toDateString()) {
        groups.Today.push(conv)
      } else if (convDate.toDateString() === yesterday.toDateString()) {
        groups.Yesterday.push(conv)
      } else if (convDate > lastWeek) {
        groups['Previous 7 days'].push(conv)
      } else {
        groups.Older.push(conv)
      }
    })

    return groups
  }

  // Initial conversation
  useEffect(() => {
    if (conversations.length === 0) {
      createNewConversation()
    }
  }, [])

  const conversationGroups = groupConversationsByDate()

  return (
    <div className="flex h-screen bg-white dark:bg-gray-900">
      {/* Sidebar - Exactly like ChatGPT */}
      <div className="w-64 bg-gray-50 dark:bg-gray-950 border-r border-gray-200 dark:border-gray-800 flex flex-col">
        {/* New chat button */}
        <div className="p-3">
          <button
            onClick={createNewConversation}
            className="w-full flex items-center gap-3 px-3 py-3 rounded-md border border-gray-200 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          >
            <Plus size={16} />
            <span className="text-sm font-medium">New chat</span>
          </button>
        </div>

        {/* Conversations list */}
        <div className="flex-1 overflow-y-auto px-3">
          <div className="space-y-2">
            {Object.entries(conversationGroups).map(([dateGroup, convs]) => (
              convs.length > 0 && (
                <div key={dateGroup}>
                  <div className="text-xs font-semibold text-gray-500 dark:text-gray-400 px-3 py-2">
                    {dateGroup}
                  </div>
                  {convs.map(conv => (
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
              )
            ))}
          </div>
        </div>

        {/* User section */}
        <div className="p-3 border-t border-gray-200 dark:border-gray-800">
          <div className="flex items-center gap-3 px-3 py-2 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 cursor-pointer">
            <div className="w-8 h-8 bg-gray-300 dark:bg-gray-700 rounded-full flex items-center justify-center">
              <User size={16} />
            </div>
            <span className="text-sm font-medium">User</span>
          </div>
        </div>
      </div>

      {/* Main chat area - Exactly like ChatGPT */}
      <div className="flex-1 flex flex-col">
        {/* Messages */}
        <div className="flex-1 overflow-y-auto">
          <div className="pb-32">
            {messages.length === 0 ? (
              <div className="h-full flex items-center justify-center min-h-[calc(100vh-200px)]">
                <div className="text-center max-w-2xl px-4">
                  <h1 className="text-3xl font-semibold mb-8 text-gray-900 dark:text-gray-100">ShopGPT</h1>
                  <p className="text-lg text-gray-500 dark:text-gray-400 mb-8">
                    Hi! I'm your shopping assistant. I search across {selectedStore === 'all' ? '1000+ online stores' : AVAILABLE_STORES.find(s => s.id === selectedStore)?.name} to find the best deals for you.
                  </p>
                  
                  {/* Example prompts */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-xl mx-auto">
                    {[
                      "Find me the best gaming laptop under $1500",
                      "Compare iPhone 15 Pro prices",
                      "I need running shoes for marathons",
                      "Search for 4K TVs with HDR"
                    ].map((prompt, index) => (
                      <button
                        key={index}
                        onClick={() => setInput(prompt)}
                        className="text-left p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
                      >
                        <p className="text-sm text-gray-700 dark:text-gray-300">{prompt}</p>
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            ) : (
              <div>
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
                            <div className="w-8 h-8 bg-gray-600 dark:bg-gray-400 rounded-sm flex items-center justify-center">
                              <User size={20} className="text-white" />
                            </div>
                          ) : (
                            <div className="w-8 h-8 bg-green-600 rounded-sm flex items-center justify-center text-white font-bold">
                              🛍️
                            </div>
                          )}
                        </div>

                        {/* Content */}
                        <div className="flex-1 space-y-2">
                          <div className="font-semibold text-gray-900 dark:text-gray-100">
                            {message.role === 'user' ? 'You' : 'ShopGPT'}
                          </div>
                          
                          {message.isStreaming ? (
                            <div className="prose dark:prose-invert max-w-none">
                              <ReactMarkdown>{message.content}</ReactMarkdown>
                              <span className="inline-block w-2 h-4 bg-gray-800 dark:bg-gray-200 animate-pulse ml-1" />
                            </div>
                          ) : (
                            <div className="prose dark:prose-invert max-w-none">
                              <ReactMarkdown
                                components={{
                                  code({node, inline, className, children, ...props}) {
                                    const match = /language-(\w+)/.exec(className || '')
                                    return !inline && match ? (
                                      <SyntaxHighlighter
                                        style={oneDark}
                                        language={match[1]}
                                        PreTag="div"
                                        {...props}
                                      >
                                        {String(children).replace(/\n$/, '')}
                                      </SyntaxHighlighter>
                                    ) : (
                                      <code className={className} {...props}>
                                        {children}
                                      </code>
                                    )
                                  }
                                }}
                              >
                                {message.content}
                              </ReactMarkdown>
                            </div>
                          )}

                          {/* Actions */}
                          {message.role === 'assistant' && !message.isStreaming && (
                            <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                              <button
                                onClick={() => copyMessage(message.content, message.id)}
                                className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
                              >
                                {copiedId === message.id ? (
                                  <Check size={16} className="text-gray-500" />
                                ) : (
                                  <Copy size={16} className="text-gray-500" />
                                )}
                              </button>
                              <button className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700">
                                <ThumbsUp size={16} className="text-gray-500" />
                              </button>
                              <button className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700">
                                <ThumbsDown size={16} className="text-gray-500" />
                              </button>
                              <button className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700">
                                <RotateCcw size={16} className="text-gray-500" />
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

        {/* Input area - Fixed at bottom */}
        <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-white via-white dark:from-gray-900 dark:via-gray-900 pt-6">
          <form onSubmit={handleSubmit} className="max-w-3xl mx-auto px-4 pb-6">
            <div className="relative">
              <textarea
                ref={textareaRef}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Search for any product..."
                rows={1}
                className="w-full resize-none rounded-md border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 px-4 py-3 pr-24 focus:outline-none focus:ring-2 focus:ring-green-500 dark:focus:ring-green-600 shadow-[0_0_10px_rgba(0,0,0,0.1)] dark:shadow-[0_0_10px_rgba(0,0,0,0.5)]"
                disabled={isStreaming}
              />
              
              {/* Store selector dropdown */}
              <div className="absolute right-12 bottom-3" ref={dropdownRef}>
                <button
                  type="button"
                  onClick={() => setShowStoreDropdown(!showStoreDropdown)}
                  className="flex items-center gap-1 px-2 py-1 text-xs bg-gray-100 dark:bg-gray-700 rounded hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                >
                  <span>{AVAILABLE_STORES.find(s => s.id === selectedStore)?.icon}</span>
                  <span>{selectedStore === 'all' ? 'All' : AVAILABLE_STORES.find(s => s.id === selectedStore)?.name}</span>
                  <ChevronDown size={12} />
                </button>
                
                {showStoreDropdown && (
                  <div className="absolute bottom-full right-0 mb-2 w-48 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-md shadow-lg py-1">
                    {AVAILABLE_STORES.map(store => (
                      <button
                        key={store.id}
                        type="button"
                        onClick={() => {
                          setSelectedStore(store.id)
                          setShowStoreDropdown(false)
                        }}
                        className="w-full text-left px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
                      >
                        <span>{store.icon}</span>
                        <span>{store.name}</span>
                      </button>
                    ))}
                  </div>
                )}
              </div>

              <button
                type="submit"
                disabled={!input.trim() || isStreaming}
                className={`absolute right-2 bottom-3 p-1 rounded-md transition-colors ${
                  input.trim() && !isStreaming
                    ? 'text-white bg-green-600 hover:bg-green-700'
                    : 'text-gray-400 dark:text-gray-500'
                }`}
              >
                <Send size={20} />
              </button>
            </div>
            
            <p className="text-xs text-center text-gray-400 dark:text-gray-500 mt-2">
              ShopGPT can make mistakes. Verify important product information and prices.
            </p>
          </form>
        </div>
      </div>
    </div>
  )
}