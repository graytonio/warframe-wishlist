import { useState, useEffect, useCallback } from 'react'
import { api, type ItemSearchResult } from '@/lib/api'
import { Input } from '@/components/ui/input'
import { ItemCard } from '@/components/items/ItemCard'
import { ItemDetailDialog } from '@/components/items/ItemDetailDialog'
import { useToast } from '@/hooks/use-toast'
import { useAuth } from '@/contexts/AuthContext'
import { Search as SearchIcon } from 'lucide-react'

export default function Search() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<ItemSearchResult[]>([])
  const [loading, setLoading] = useState(false)
  const [selectedItem, setSelectedItem] = useState<string | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const { toast } = useToast()
  const { user } = useAuth()

  const searchItems = useCallback(async (searchQuery: string) => {
    if (!searchQuery.trim()) {
      setResults([])
      return
    }

    setLoading(true)
    try {
      const data = await api.searchItems(searchQuery)
      setResults(data)
    } catch {
      toast({
        title: 'Search failed',
        description: 'Unable to search items. Please try again.',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }, [toast])

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      searchItems(query)
    }, 300)

    return () => clearTimeout(timeoutId)
  }, [query, searchItems])

  const handleAddToWishlist = async (uniqueName: string, quantity = 1) => {
    if (!user) {
      toast({
        title: 'Sign in required',
        description: 'Please sign in to add items to your wishlist.',
        variant: 'destructive',
      })
      return
    }

    try {
      await api.addToWishlist(uniqueName, quantity)
      toast({
        title: 'Added to wishlist',
        description: 'Item has been added to your wishlist.',
      })
    } catch (err) {
      toast({
        title: 'Failed to add item',
        description: err instanceof Error ? err.message : 'Please try again.',
        variant: 'destructive',
      })
    }
  }

  const handleViewDetails = (uniqueName: string) => {
    setSelectedItem(uniqueName)
    setDialogOpen(true)
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-2xl mx-auto mb-8">
        <h1 className="text-3xl font-bold text-center mb-6">Warframe Item Search</h1>
        <div className="relative">
          <SearchIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Search for Warframes, weapons, mods..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="pl-10 h-12 text-lg"
          />
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : results.length > 0 ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {results.map((item) => (
            <ItemCard
              key={item.uniqueName}
              item={item}
              onAddToWishlist={user ? handleAddToWishlist : undefined}
              onViewDetails={handleViewDetails}
            />
          ))}
        </div>
      ) : query ? (
        <p className="text-center text-muted-foreground py-12">
          No items found for "{query}"
        </p>
      ) : (
        <p className="text-center text-muted-foreground py-12">
          Start typing to search for items
        </p>
      )}

      <ItemDetailDialog
        uniqueName={selectedItem}
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        onAddToWishlist={user ? handleAddToWishlist : undefined}
      />
    </div>
  )
}
