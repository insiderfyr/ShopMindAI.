'use client'

import { useState } from 'react'
import { Plus, Search, Trash2, Edit2, ShoppingBag, History, Star, TrendingUp, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useChatStore } from '@/lib/store/chat'
import { cn } from '@/lib/utils'

interface SidebarProps {
  isOpen: boolean
  onToggle: () => void
}

export function Sidebar({ isOpen, onToggle }: SidebarProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const { 
    chats, 
    currentChatId, 
    createChat, 
    selectChat, 
    deleteChat, 
    updateChat 
  } = useChatStore()

  const filteredChats = chats.filter(chat => 
    chat.title.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const categorizedChats = {
    today: filteredChats.filter(chat => {
      const chatDate = new Date(chat.createdAt)
      const today = new Date()
      return chatDate.toDateString() === today.toDateString()
    }),
    yesterday: filteredChats.filter(chat => {
      const chatDate = new Date(chat.createdAt)
      const yesterday = new Date()
      yesterday.setDate(yesterday.getDate() - 1)
      return chatDate.toDateString() === yesterday.toDateString()
    }),
    lastWeek: filteredChats.filter(chat => {
      const chatDate = new Date(chat.createdAt)
      const weekAgo = new Date()
      weekAgo.setDate(weekAgo.getDate() - 7)
      const yesterday = new Date()
      yesterday.setDate(yesterday.getDate() - 1)
      return chatDate > weekAgo && chatDate.toDateString() !== yesterday.toDateString() && chatDate.toDateString() !== new Date().toDateString()
    }),
    older: filteredChats.filter(chat => {
      const chatDate = new Date(chat.createdAt)
      const weekAgo = new Date()
      weekAgo.setDate(weekAgo.getDate() - 7)
      return chatDate <= weekAgo
    })
  }

  const handleNewChat = () => {
    const newChat = createChat()
    selectChat(newChat.id)
  }

  return (
    <div className="h-full flex flex-col bg-gray-50 dark:bg-gray-900">
      {/* Header */}
      <div className="p-4 border-b border-gray-200 dark:border-gray-800">
        <Button
          onClick={handleNewChat}
          className="w-full justify-start gap-2 bg-[#4d8eff] hover:bg-[#3a6cd9] text-white"
        >
          <Plus size={20} />
          New Shopping Chat
        </Button>
      </div>

      {/* Search */}
      <div className="p-4 border-b border-gray-200 dark:border-gray-800">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={16} />
          <Input
            type="search"
            placeholder="Search conversations..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700"
          />
        </div>
      </div>

      {/* Quick Actions */}
      <div className="px-4 py-3 border-b border-gray-200 dark:border-gray-800">
        <div className="grid grid-cols-2 gap-2">
          <Button
            variant="outline"
            size="sm"
            className="justify-start gap-2 text-xs border-gray-200 dark:border-gray-700"
          >
            <TrendingUp size={14} />
            Trending
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="justify-start gap-2 text-xs border-gray-200 dark:border-gray-700"
          >
            <Star size={14} />
            Saved
          </Button>
        </div>
      </div>

      {/* Chat List */}
      <ScrollArea className="flex-1">
        <div className="p-4 space-y-6">
          {/* Today */}
          {categorizedChats.today.length > 0 && (
            <div>
              <h3 className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                Today
              </h3>
              <div className="space-y-1">
                {categorizedChats.today.map(chat => (
                  <ChatItem
                    key={chat.id}
                    chat={chat}
                    isActive={chat.id === currentChatId}
                    onSelect={() => selectChat(chat.id)}
                    onDelete={() => deleteChat(chat.id)}
                    onUpdate={(title) => updateChat(chat.id, { title })}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Yesterday */}
          {categorizedChats.yesterday.length > 0 && (
            <div>
              <h3 className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                Yesterday
              </h3>
              <div className="space-y-1">
                {categorizedChats.yesterday.map(chat => (
                  <ChatItem
                    key={chat.id}
                    chat={chat}
                    isActive={chat.id === currentChatId}
                    onSelect={() => selectChat(chat.id)}
                    onDelete={() => deleteChat(chat.id)}
                    onUpdate={(title) => updateChat(chat.id, { title })}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Last Week */}
          {categorizedChats.lastWeek.length > 0 && (
            <div>
              <h3 className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                Last 7 Days
              </h3>
              <div className="space-y-1">
                {categorizedChats.lastWeek.map(chat => (
                  <ChatItem
                    key={chat.id}
                    chat={chat}
                    isActive={chat.id === currentChatId}
                    onSelect={() => selectChat(chat.id)}
                    onDelete={() => deleteChat(chat.id)}
                    onUpdate={(title) => updateChat(chat.id, { title })}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Older */}
          {categorizedChats.older.length > 0 && (
            <div>
              <h3 className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                Older
              </h3>
              <div className="space-y-1">
                {categorizedChats.older.map(chat => (
                  <ChatItem
                    key={chat.id}
                    chat={chat}
                    isActive={chat.id === currentChatId}
                    onSelect={() => selectChat(chat.id)}
                    onDelete={() => deleteChat(chat.id)}
                    onUpdate={(title) => updateChat(chat.id, { title })}
                  />
                ))}
              </div>
            </div>
          )}

          {filteredChats.length === 0 && (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
              <ShoppingBag className="w-12 h-12 mx-auto mb-3 opacity-50" />
              <p className="text-sm">No shopping conversations yet</p>
              <p className="text-xs mt-1">Start a new chat to explore products!</p>
            </div>
          )}
        </div>
      </ScrollArea>

      {/* Footer */}
      <div className="p-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900">
        <Button
          variant="ghost"
          size="sm"
          className="w-full justify-start gap-2 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100"
        >
          <History size={16} />
          Shopping History
        </Button>
      </div>
    </div>
  )
}

interface ChatItemProps {
  chat: any
  isActive: boolean
  onSelect: () => void
  onDelete: () => void
  onUpdate: (title: string) => void
}

function ChatItem({ chat, isActive, onSelect, onDelete, onUpdate }: ChatItemProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editTitle, setEditTitle] = useState(chat.title)

  const handleEdit = (e: React.MouseEvent) => {
    e.stopPropagation()
    setIsEditing(true)
  }

  const handleSave = () => {
    onUpdate(editTitle)
    setIsEditing(false)
  }

  const handleCancel = () => {
    setEditTitle(chat.title)
    setIsEditing(false)
  }

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    onDelete()
  }

  return (
    <div
      onClick={onSelect}
      className={cn(
        "group relative flex items-center gap-2 p-3 rounded-lg cursor-pointer transition-all duration-200",
        isActive 
          ? "bg-[#4d8eff]/10 text-gray-900 dark:text-gray-100" 
          : "hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300"
      )}
    >
      <ShoppingBag size={16} className={cn(
        "flex-shrink-0",
        isActive ? "text-[#4d8eff]" : "text-gray-400"
      )} />
      
      {isEditing ? (
        <div className="flex-1 flex items-center gap-2">
          <Input
            value={editTitle}
            onChange={(e) => setEditTitle(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleSave()
              if (e.key === 'Escape') handleCancel()
            }}
            onClick={(e) => e.stopPropagation()}
            className="h-7 text-sm"
            autoFocus
          />
          <Button size="sm" variant="ghost" onClick={handleSave}>
            Save
          </Button>
        </div>
      ) : (
        <>
          <span className="flex-1 truncate text-sm">{chat.title}</span>
          
          {isActive && <ChevronRight size={14} className="text-[#4d8eff]" />}
          
          <div className="hidden group-hover:flex items-center gap-1">
            <Button
              size="icon"
              variant="ghost"
              onClick={handleEdit}
              className="h-6 w-6 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            >
              <Edit2 size={12} />
            </Button>
            <Button
              size="icon"
              variant="ghost"
              onClick={handleDelete}
              className="h-6 w-6 text-gray-500 hover:text-red-600 dark:text-gray-400 dark:hover:text-red-500"
            >
              <Trash2 size={12} />
            </Button>
          </div>
        </>
      )}
    </div>
  )
}