'use client'

import { useState } from 'react'
import { Sidebar } from '@/components/sidebar'
import { ChatArea } from '@/components/chat-area'
import { useChatStore } from '@/lib/store/chat'
import { useAuthStore } from '@/lib/store/auth'

export default function ChatPage() {
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const { currentChatId } = useChatStore()
  const { user } = useAuthStore()

  return (
    <div className="flex h-full w-full">
      {/* Sidebar */}
      <Sidebar 
        isOpen={sidebarOpen} 
        onToggle={() => setSidebarOpen(!sidebarOpen)}
      />

      {/* Main Chat Area */}
      <main className="flex-1 relative flex flex-col bg-gpt-dark">
        {/* Mobile sidebar toggle */}
        {!sidebarOpen && (
          <button
            onClick={() => setSidebarOpen(true)}
            className="absolute left-4 top-4 z-10 p-2 rounded-md hover:bg-gpt-hover md:hidden"
          >
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M3 12H21M3 6H21M3 18H21" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </button>
        )}

        {/* Chat Content */}
        <ChatArea chatId={currentChatId} />
      </main>
    </div>
  )
}