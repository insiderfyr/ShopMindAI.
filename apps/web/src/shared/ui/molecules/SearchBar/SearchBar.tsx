import React, { useState, useRef, useEffect } from 'react'
import { Search, X, Loader2 } from 'lucide-react'
import { Button } from '../../atoms/Button/Button'
import { Input } from '../../atoms/Input/Input'
import { StoreSelector } from '../StoreSelector/StoreSelector'
import { cn } from '@/lib/utils'

interface SearchBarProps {
  placeholder?: string
  onSearch: (query: string, store: string) => void
  onClear?: () => void
  isLoading?: boolean
  suggestions?: string[]
  stores?: { id: string; name: string; icon: string }[]
  className?: string
}

export const SearchBar: React.FC<SearchBarProps> = ({
  placeholder = 'Search for any product...',
  onSearch,
  onClear,
  isLoading = false,
  suggestions = [],
  stores = [
    { id: 'all', name: 'All stores', icon: '🌐' },
    { id: 'amazon', name: 'Amazon', icon: '📦' },
    { id: 'bestbuy', name: 'Best Buy', icon: '🟡' },
    { id: 'walmart', name: 'Walmart', icon: '🔵' },
    { id: 'target', name: 'Target', icon: '🎯' },
    { id: 'ebay', name: 'eBay', icon: '🛒' }
  ],
  className
}) => {
  const [query, setQuery] = useState('')
  const [selectedStore, setSelectedStore] = useState('all')
  const [showSuggestions, setShowSuggestions] = useState(false)
  const searchRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(event.target as Node)) {
        setShowSuggestions(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleSearch = (e?: React.FormEvent) => {
    e?.preventDefault()
    if (query.trim()) {
      onSearch(query.trim(), selectedStore)
      setShowSuggestions(false)
    }
  }

  const handleClear = () => {
    setQuery('')
    onClear?.()
    setShowSuggestions(false)
  }

  const handleSuggestionClick = (suggestion: string) => {
    setQuery(suggestion)
    setShowSuggestions(false)
    onSearch(suggestion, selectedStore)
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setQuery(e.target.value)
    setShowSuggestions(e.target.value.length > 0 && suggestions.length > 0)
  }

  return (
    <div ref={searchRef} className={cn('relative w-full', className)}>
      <form onSubmit={handleSearch} className="flex items-center gap-2">
        <div className="relative flex-1">
          <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
            <Search size={20} />
          </div>
          
          <Input
            value={query}
            onChange={handleInputChange}
            onFocus={() => query && suggestions.length > 0 && setShowSuggestions(true)}
            placeholder={placeholder}
            className="pl-10 pr-10"
            data-testid="search-input"
          />
          
          {query && (
            <button
              type="button"
              onClick={handleClear}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 transition-colors"
              data-testid="clear-search"
            >
              <X size={18} />
            </button>
          )}
        </div>

        <StoreSelector
          stores={stores}
          selectedStore={selectedStore}
          onStoreChange={setSelectedStore}
          data-testid="store-selector"
        />

        <Button
          type="submit"
          loading={isLoading}
          disabled={!query.trim()}
          className="min-w-[100px]"
          data-testid="search-button"
        >
          {isLoading ? (
            <>
              <Loader2 className="animate-spin mr-2" size={16} />
              Searching...
            </>
          ) : (
            'Search'
          )}
        </Button>
      </form>

      {/* Suggestions Dropdown */}
      {showSuggestions && suggestions.length > 0 && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-50">
          <div className="p-2">
            <p className="text-xs text-gray-500 dark:text-gray-400 px-3 py-1">Suggestions</p>
            {suggestions.map((suggestion, index) => (
              <button
                key={index}
                onClick={() => handleSuggestionClick(suggestion)}
                className="w-full text-left px-3 py-2 rounded hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
              >
                <div className="flex items-center gap-2">
                  <Search size={14} className="text-gray-400" />
                  <span className="text-sm">{suggestion}</span>
                </div>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

export default SearchBar