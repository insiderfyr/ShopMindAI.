"use client"

import type React from "react"
import { useState, useRef, useEffect } from "react"
import Image from "next/image"
import Link from "next/link"
import { PlusCircle, Settings, Send, Menu, MessageSquare, Edit3, Trash2 } from "lucide-react"

type Message = {
  id: string
  content: string
  sender: "user" | "ai"
  timestamp: Date
}

type ChatHistory = {
  id: string
  title: string
  lastMessage: string
  date: Date
}

export default function ChatPage() {
  const [messages, setMessages] = useState<Message[]>([])
  const [inputValue, setInputValue] = useState("")
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  // Sample chat history organized by date
  const chatHistory: Record<string, ChatHistory[]> = {
    Today: [
      { id: "1", title: "Running shoes under $100", lastMessage: "What about Nike Air Zoom?", date: new Date() },
      { id: "2", title: "Skincare for sensitive skin", lastMessage: "I need gentle products", date: new Date() },
      { id: "3", title: "Best laptop for design work", lastMessage: "MacBook vs ThinkPad", date: new Date() },
    ],
    Yesterday: [
      {
        id: "4",
        title: "Gift ideas for anniversary",
        lastMessage: "Budget around $150",
        date: new Date(Date.now() - 86400000),
      },
      {
        id: "5",
        title: "Kitchen appliances comparison",
        lastMessage: "Looking for a new blender",
        date: new Date(Date.now() - 86400000),
      },
    ],
    "This Week": [
      {
        id: "6",
        title: "Winter jacket recommendations",
        lastMessage: "Something waterproof",
        date: new Date(Date.now() - 3 * 86400000),
      },
      {
        id: "7",
        title: "Home office setup",
        lastMessage: "Ergonomic chair options",
        date: new Date(Date.now() - 4 * 86400000),
      },
    ],
  }

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault()
    if (!inputValue.trim() || isLoading) return

    // Add user message
    const userMessage = {
      id: Date.now().toString(),
      content: inputValue,
      sender: "user",
      timestamp: new Date(),
    }
    setMessages((prev) => [...prev, userMessage])
    setInputValue("")
    setIsLoading(true)

    // Simulate AI response
    setTimeout(() => {
      const aiMessage = {
        id: (Date.now() + 1).toString(),
        content:
          "I'd be happy to help you find what you're looking for! Let me search through our partner stores to find the best options that match your criteria.",
        sender: "ai",
        timestamp: new Date(),
      }
      setMessages((prev) => [...prev, aiMessage])
      setIsLoading(false)
    }, 1000)
  }

  // Auto-scroll to bottom when messages change
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages])

  // Auto-focus input on load
  useEffect(() => {
    if (inputRef.current && messages.length === 0) {
      inputRef.current.focus()
    }
  }, [messages.length])

  // Handle responsive sidebar
  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth < 768) {
        setSidebarOpen(false)
      } else {
        setSidebarOpen(true)
      }
    }

    handleResize()
    window.addEventListener("resize", handleResize)
    return () => window.removeEventListener("resize", handleResize)
  }, [])

  const toggleSidebar = () => {
    setSidebarOpen(!sidebarOpen)
  }

  return (
    <div className="flex h-screen bg-gray-50 overflow-hidden">
      {/* Mobile sidebar backdrop */}
      {sidebarOpen && (
        <div className="fixed inset-0 bg-black/20 z-20 md:hidden" onClick={() => setSidebarOpen(false)} />
      )}

      {/* Sidebar */}
      <div
        className={`fixed md:relative z-30 h-full w-64 bg-white border-r border-gray-200 transition-transform duration-300 ${
          sidebarOpen ? "translate-x-0" : "-translate-x-full md:translate-x-0"
        }`}
      >
        {/* Logo */}
        <div className="p-4 border-b border-gray-100">
          <Link href="/" className="flex items-center gap-2">
            <div className="w-8 h-8 relative">
              <Image src="/logo.svg" alt="ShopMind Logo" fill className="object-contain" />
            </div>
            <h1 className="font-bold text-gray-800">ShopMindAI</h1>
          </Link>
        </div>

        {/* New Chat Button */}
        <div className="p-4">
          <button
            onClick={() => {
              setMessages([])
              setInputValue("")
              setSidebarOpen(false)
            }}
            className="w-full flex items-center justify-center gap-2 bg-blue-600 hover:bg-blue-700 text-white py-2 px-4 rounded-lg"
          >
            <PlusCircle size={18} />
            <span>New Chat</span>
          </button>
        </div>

        {/* Chat History */}
        <div className="p-4">
          <h3 className="text-xs font-medium text-gray-500 uppercase mb-2">Recent Chats</h3>
          <div className="space-y-1">
            {Object.entries(chatHistory).map(([date, chats]) => (
              <div key={date} className="mb-6">
                <h3 className="px-2 py-2 text-xs text-gray-500 uppercase tracking-wide font-medium">{date}</h3>
                <div className="space-y-1">
                  {chats.map((chat) => (
                    <div
                      key={chat.id}
                      className="group flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-gray-100 cursor-pointer"
                    >
                      <MessageSquare size={16} className="text-gray-400 flex-shrink-0" />
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-gray-700 truncate">{chat.title}</p>
                        <p className="text-xs text-gray-500 truncate">{chat.lastMessage}</p>
                      </div>
                      <div className="opacity-0 group-hover:opacity-100 flex items-center gap-1">
                        <button className="p-1 rounded hover:bg-gray-200">
                          <Edit3 size={14} className="text-gray-500" />
                        </button>
                        <button className="p-1 rounded hover:bg-gray-200">
                          <Trash2 size={14} className="text-gray-500" />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* User Profile */}
        <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-100">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white">U</div>
              <div>
                <p className="text-sm font-medium">User</p>
                <p className="text-xs text-gray-500">Free Plan</p>
              </div>
            </div>
            <button className="p-1 rounded-lg hover:bg-gray-100">
              <Settings size={16} className="text-gray-500" />
            </button>
          </div>
        </div>
      </div>

      {/* Main Chat Area */}
      <div className="flex-1 flex flex-col h-screen overflow-hidden">
        {/* Header */}
        <div className="bg-white border-b border-gray-200 p-4 flex items-center">
          <button onClick={() => setSidebarOpen(!sidebarOpen)} className="p-2 rounded-lg hover:bg-gray-100 md:hidden">
            <Menu size={20} />
          </button>
          <h1 className="ml-2 font-semibold text-gray-800">ShopMindAI Chat</h1>
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-4">
          {messages.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center">
              <div className="text-center mb-8">
                <div className="flex items-center justify-center gap-3 mb-4">
                  <div className="w-10 h-10 relative">
                    <Image src="/logo.svg" alt="ShopMind Logo" fill className="object-contain" />
                  </div>
                  <h1 className="text-2xl font-bold text-gray-800">Hi, I'm ShopMindAI</h1>
                </div>
                <p className="text-gray-600">How can I help you with your shopping today?</p>
              </div>

              <div className="w-full max-w-md grid grid-cols-1 gap-2">
                {[
                  "I need comfortable running shoes under $100",
                  "Find me skincare products for sensitive skin",
                  "What's the best laptop for graphic design?",
                ].map((example, i) => (
                  <button
                    key={i}
                    onClick={() => setInputValue(example)}
                    className="text-left p-3 bg-white border border-gray-200 rounded-lg hover:border-blue-300 transition-colors"
                  >
                    <p className="text-sm text-gray-800">{example}</p>
                  </button>
                ))}
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              {messages.map((message) => (
                <div key={message.id} className={`flex ${message.sender === "user" ? "justify-end" : "justify-start"}`}>
                  {message.sender === "ai" && (
                    <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center flex-shrink-0">
                      <div className="w-4 h-4 relative">
                        <Image src="/logo.svg" alt="AI" fill className="object-contain" />
                      </div>
                    </div>
                  )}
                  <div
                    className={`max-w-[75%] p-3 rounded-lg ${
                      message.sender === "user"
                        ? "bg-blue-600 text-white"
                        : "bg-white border border-gray-200 text-gray-800"
                    }`}
                  >
                    <p className="text-sm">{message.content}</p>
                    <p className={`text-xs mt-1 ${message.sender === "user" ? "text-blue-200" : "text-gray-500"}`}>
                      {message.timestamp.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
                    </p>
                  </div>
                  {message.sender === "user" && (
                    <div className="w-8 h-8 bg-gray-200 rounded-full flex items-center justify-center flex-shrink-0">
                      <span className="text-sm font-semibold text-gray-600">U</span>
                    </div>
                  )}
                </div>
              ))}

              {/* Loading indicator */}
              {isLoading && (
                <div className="flex justify-start">
                  <div className="bg-white border border-gray-200 p-3 rounded-lg">
                    <div className="flex items-center gap-2">
                      <div className="flex gap-1">
                        <div className="w-2 h-2 bg-gray-400 rounded-full animate-pulse"></div>
                        <div className="w-2 h-2 bg-gray-400 rounded-full animate-pulse delay-75"></div>
                        <div className="w-2 h-2 bg-gray-400 rounded-full animate-pulse delay-150"></div>
                      </div>
                      <span className="text-sm text-gray-500">Thinking...</span>
                    </div>
                  </div>
                </div>
              )}

              <div ref={messagesEndRef} />
            </div>
          )}
        </div>

        {/* Input Area */}
        <div className="bg-white border-t border-gray-200 p-4">
          <form onSubmit={handleSendMessage} className="flex items-center gap-2">
            <input
              ref={inputRef}
              type="text"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              placeholder="Type your message..."
              className="flex-1 p-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={isLoading}
            />
            <button
              type="submit"
              disabled={!inputValue.trim() || isLoading}
              className={`p-2 rounded-lg ${
                inputValue.trim() && !isLoading
                  ? "bg-blue-600 text-white hover:bg-blue-700"
                  : "bg-gray-200 text-gray-400 cursor-not-allowed"
              }`}
            >
              <Send size={20} />
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
