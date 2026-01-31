import type { ItemSearchResult } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ItemImage, getItemImageUrl } from '@/components/ui/item-image'
import { Plus } from 'lucide-react'

interface ItemCardProps {
  item: ItemSearchResult
  onAddToWishlist?: (uniqueName: string) => void
  onViewDetails?: (uniqueName: string) => void
}

export function ItemCard({ item, onAddToWishlist, onViewDetails }: ItemCardProps) {
  return (
    <Card data-testid="item-card" className="overflow-hidden hover:shadow-lg transition-shadow cursor-pointer">
      <div onClick={() => onViewDetails?.(item.uniqueName)}>
        <CardHeader className="p-4 pb-2">
          <div className="flex items-center gap-3">
            <ItemImage
              src={getItemImageUrl(item.imageName)}
              alt={item.name}
              className="w-12 h-12"
            />
            <div className="flex-1 min-w-0">
              <CardTitle className="text-base truncate">{item.name}</CardTitle>
              {item.category && (
                <p className="text-xs text-muted-foreground">{item.category}</p>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-4 pt-0">
          {item.description && (
            <p className="text-sm text-muted-foreground line-clamp-2">
              {item.description}
            </p>
          )}
        </CardContent>
      </div>
      {onAddToWishlist && (
        <div className="px-4 pb-4">
          <Button
            size="sm"
            className="w-full"
            onClick={(e) => {
              e.stopPropagation()
              onAddToWishlist(item.uniqueName)
            }}
          >
            <Plus className="h-4 w-4 mr-1" />
            Add to Wishlist
          </Button>
        </div>
      )}
    </Card>
  )
}
