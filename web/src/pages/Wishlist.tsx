import { useState, useEffect, useCallback } from 'react'
import { api, type Wishlist as WishlistType } from '@/lib/api'
import { WishlistItemRow } from '@/components/wishlist/WishlistItem'
import { MaterialsSummary } from '@/components/wishlist/MaterialsSummary'
import { useToast } from '@/hooks/use-toast'
import { Button } from '@/components/ui/button'
import { Link } from 'react-router-dom'
import { RefreshCw, Plus } from 'lucide-react'

export default function Wishlist() {
  const [wishlist, setWishlist] = useState<WishlistType | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const { toast } = useToast()

  const fetchWishlist = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await api.getWishlist()
      setWishlist(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load wishlist')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchWishlist()
  }, [fetchWishlist])

  const handleUpdateQuantity = async (uniqueName: string, quantity: number) => {
    try {
      const updated = await api.updateWishlistQuantity(uniqueName, quantity)
      setWishlist(updated)
    } catch (err) {
      toast({
        title: 'Failed to update quantity',
        description: err instanceof Error ? err.message : 'Please try again.',
        variant: 'destructive',
      })
      fetchWishlist()
    }
  }

  const handleRemove = async (uniqueName: string) => {
    try {
      const updated = await api.removeFromWishlist(uniqueName)
      setWishlist(updated)
      toast({
        title: 'Item removed',
        description: 'Item has been removed from your wishlist.',
      })
    } catch (err) {
      toast({
        title: 'Failed to remove item',
        description: err instanceof Error ? err.message : 'Please try again.',
        variant: 'destructive',
      })
    }
  }

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-12">
          <p className="text-destructive mb-4">{error}</p>
          <Button onClick={fetchWishlist}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    )
  }

  const items = wishlist?.items || []

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-bold">My Wishlist</h1>
        <div className="flex gap-2">
          <Button variant="outline" size="icon" onClick={fetchWishlist}>
            <RefreshCw className="h-4 w-4" />
          </Button>
          <Button asChild>
            <Link to="/">
              <Plus className="h-4 w-4 mr-2" />
              Add Items
            </Link>
          </Button>
        </div>
      </div>

      {items.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-muted-foreground mb-4">Your wishlist is empty</p>
          <Button asChild>
            <Link to="/">
              <Plus className="h-4 w-4 mr-2" />
              Search for items to add
            </Link>
          </Button>
        </div>
      ) : (
        <div className="grid lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 space-y-4">
            <h2 className="text-lg font-medium">
              Items ({items.length})
            </h2>
            <div className="space-y-3">
              {items.map((item) => (
                <WishlistItemRow
                  key={item.uniqueName}
                  item={item}
                  onUpdateQuantity={handleUpdateQuantity}
                  onRemove={handleRemove}
                />
              ))}
            </div>
          </div>
          <div>
            <MaterialsSummary />
          </div>
        </div>
      )}
    </div>
  )
}
