'use client'

import { useState } from 'react'
import { Copy, Check, User, ShoppingBag, RefreshCw, ThumbsUp, ThumbsDown } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import remarkGfm from 'remark-gfm'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
  isStreaming?: boolean
  productRecommendations?: any[]
}

interface ChatMessageProps {
  message: Message
}

export function ChatMessage({ message }: ChatMessageProps) {
  const [copied, setCopied] = useState(false)
  const [feedback, setFeedback] = useState<'up' | 'down' | null>(null)
  const isUser = message.role === 'user'

  const handleCopy = () => {
    navigator.clipboard.writeText(message.content)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleFeedback = (type: 'up' | 'down') => {
    setFeedback(type)
    // Send feedback to backend
    console.log(`Feedback: ${type} for message ${message.id}`)
  }

  return (
    <div className={cn(
      "group relative flex gap-4",
      isUser ? "justify-end" : "justify-start"
    )}>
      {/* Avatar */}
      {!isUser && (
        <div className="flex-shrink-0">
          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-[#4d8eff] to-[#3a6cd9] flex items-center justify-center shadow-md">
            <ShoppingBag size={20} className="text-white" />
          </div>
        </div>
      )}

      {/* Message Content */}
      <div className={cn(
        "flex flex-col max-w-[70%]",
        isUser && "items-end"
      )}>
        <div className={cn(
          "rounded-2xl px-4 py-3 shadow-sm",
          isUser 
            ? "bg-[#4d8eff] text-white" 
            : "bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700"
        )}>
          {message.isStreaming ? (
            <div className="flex items-center gap-2">
              <div className="flex gap-1">
                <div className="w-2 h-2 bg-[#4d8eff] rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                <div className="w-2 h-2 bg-[#4d8eff] rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                <div className="w-2 h-2 bg-[#4d8eff] rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
              </div>
              <span className="text-sm text-gray-600 dark:text-gray-400">ShopMindAI is thinking...</span>
            </div>
          ) : (
            <div className={cn(
              "prose max-w-none",
              isUser ? "prose-invert" : "prose-gray dark:prose-invert"
            )}>
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                components={{
                  code({ node, inline, className, children, ...props }) {
                    const match = /language-(\w+)/.exec(className || '')
                    const language = match ? match[1] : ''
                    
                    if (!inline && language) {
                      return (
                        <div className="relative my-4">
                          <div className="flex items-center justify-between bg-gray-900 rounded-t-md px-4 py-2">
                            <span className="text-xs text-gray-400">{language}</span>
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => {
                                navigator.clipboard.writeText(String(children))
                              }}
                              className="h-6 px-2 text-xs text-gray-400 hover:text-white"
                            >
                              Copy
                            </Button>
                          </div>
                          <SyntaxHighlighter
                            style={oneDark}
                            language={language}
                            PreTag="div"
                            customStyle={{
                              margin: 0,
                              borderRadius: '0 0 0.375rem 0.375rem',
                              fontSize: '0.875rem',
                            }}
                            {...props}
                          >
                            {String(children).replace(/\n$/, '')}
                          </SyntaxHighlighter>
                        </div>
                      )
                    }
                    
                    return (
                      <code className="bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded text-sm" {...props}>
                        {children}
                      </code>
                    )
                  },
                  p({ children }) {
                    return <p className="mb-4 last:mb-0 leading-relaxed">{children}</p>
                  },
                  ul({ children }) {
                    return <ul className="list-disc pl-6 mb-4 space-y-1">{children}</ul>
                  },
                  ol({ children }) {
                    return <ol className="list-decimal pl-6 mb-4 space-y-1">{children}</ol>
                  },
                  h1({ children }) {
                    return <h1 className="text-2xl font-bold mb-4 text-gray-900 dark:text-gray-100">{children}</h1>
                  },
                  h2({ children }) {
                    return <h2 className="text-xl font-bold mb-3 text-gray-900 dark:text-gray-100">{children}</h2>
                  },
                  h3({ children }) {
                    return <h3 className="text-lg font-bold mb-2 text-gray-900 dark:text-gray-100">{children}</h3>
                  },
                  blockquote({ children }) {
                    return (
                      <blockquote className="border-l-4 border-[#4d8eff] pl-4 my-4 text-gray-600 dark:text-gray-400 italic">
                        {children}
                      </blockquote>
                    )
                  },
                  a({ href, children }) {
                    return (
                      <a
                        href={href}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-[#4d8eff] hover:underline font-medium"
                      >
                        {children}
                      </a>
                    )
                  },
                }}
              >
                {message.content}
              </ReactMarkdown>
            </div>
          )}
        </div>

        {/* Product Recommendations (if any) */}
        {message.productRecommendations && message.productRecommendations.length > 0 && (
          <div className="mt-3 grid grid-cols-1 sm:grid-cols-2 gap-2">
            {message.productRecommendations.map((product, index) => (
              <div key={index} className="bg-gray-50 dark:bg-gray-800 rounded-lg p-3 border border-gray-200 dark:border-gray-700">
                <h4 className="font-medium text-sm text-gray-900 dark:text-gray-100">{product.name}</h4>
                <p className="text-sm text-[#4d8eff] font-semibold">{product.price}</p>
                <p className="text-xs text-gray-600 dark:text-gray-400 mt-1">{product.store}</p>
              </div>
            ))}
          </div>
        )}

        {/* Actions */}
        {!isUser && !message.isStreaming && (
          <div className="flex items-center gap-2 mt-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <Button
              size="icon"
              variant="ghost"
              onClick={handleCopy}
              className="h-7 w-7 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            >
              {copied ? <Check size={14} /> : <Copy size={14} />}
            </Button>
            
            <Button
              size="icon"
              variant="ghost"
              className="h-7 w-7 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            >
              <RefreshCw size={14} />
            </Button>

            <div className="flex items-center gap-1 ml-2">
              <Button
                size="icon"
                variant="ghost"
                onClick={() => handleFeedback('up')}
                className={cn(
                  "h-7 w-7",
                  feedback === 'up' 
                    ? "text-green-600 hover:text-green-700" 
                    : "text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                )}
              >
                <ThumbsUp size={14} />
              </Button>
              
              <Button
                size="icon"
                variant="ghost"
                onClick={() => handleFeedback('down')}
                className={cn(
                  "h-7 w-7",
                  feedback === 'down' 
                    ? "text-red-600 hover:text-red-700" 
                    : "text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                )}
              >
                <ThumbsDown size={14} />
              </Button>
            </div>
          </div>
        )}

        {/* Timestamp */}
        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
          {new Date(message.timestamp).toLocaleTimeString()}
        </p>
      </div>

      {/* User Avatar */}
      {isUser && (
        <div className="flex-shrink-0">
          <div className="w-10 h-10 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
            <User size={20} className="text-gray-600 dark:text-gray-300" />
          </div>
        </div>
      )}
    </div>
  )
}