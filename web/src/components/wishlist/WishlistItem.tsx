import { useState, useEffect } from 'react'
import type { WishlistItem as WishlistItemType, Item } from '@/lib/api'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ItemImage, getItemImageUrl } from '@/components/ui/item-image'
import { Trash2, Minus, Plus } from 'lucide-react'

interface WishlistItemProps {
  item: WishlistItemType
  onUpdateQuantity: (uniqueName: string, quantity: number) => void
  onRemove: (uniqueName: string) => void
  onViewDetails?: (uniqueName: string) => void
}

export function WishlistItemRow({ item, onUpdateQuantity, onRemove, onViewDetails }: WishlistItemProps) {
  const [itemDetails, setItemDetails] = useState<Item | null>(null)
  const [localQuantity, setLocalQuantity] = useState(item.quantity)

  useEffect(() => {
    api.getItem(item.uniqueName)
      .then(setItemDetails)
      .catch(() => setItemDetails(null))
  }, [item.uniqueName])

  useEffect(() => {
    setLocalQuantity(item.quantity)
  }, [item.quantity])

  const handleQuantityChange = (newQuantity: number) => {
    if (newQuantity < 1) return
    setLocalQuantity(newQuantity)
    onUpdateQuantity(item.uniqueName, newQuantity)
  }

  // Get the best drop location (highest chance)
  const bestDrop = itemDetails?.drops?.length
    ? [...itemDetails.drops].sort((a, b) => (b.chance ?? 0) - (a.chance ?? 0))[0]
    : null

  return (
    <div className="flex items-center gap-4 p-4 border rounded-lg">
      <button
        type="button"
        onClick={() => onViewDetails?.(item.uniqueName)}
        className="flex items-center gap-4 flex-1 min-w-0 text-left hover:opacity-80 transition-opacity cursor-pointer"
      >
        <ItemImage
          src={getItemImageUrl(itemDetails?.imageName)}
          alt={itemDetails?.name || item.uniqueName}
          className="w-12 h-12"
        />
        <div className="flex-1 min-w-0">
          <h4 className="font-medium truncate">
            {itemDetails?.name || item.uniqueName}
          </h4>
          {bestDrop ? (
            <p className="text-sm text-muted-foreground truncate">
              {bestDrop.location}
              {bestDrop.chance && ` (${(bestDrop.chance * 100).toFixed(1)}%)`}
            </p>
          ) : itemDetails?.category ? (
            <p className="text-sm text-muted-foreground">{itemDetails.category}</p>
          ) : null}
          {itemDetails?.buildPrice && (
            <p className="text-xs text-muted-foreground">
              {(itemDetails.buildPrice * localQuantity).toLocaleString()} Credits
            </p>
          )}
        </div>
      </button>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="icon"
          onClick={() => handleQuantityChange(localQuantity - 1)}
          disabled={localQuantity <= 1}
        >
          <Minus className="h-4 w-4" />
        </Button>
        <Input
          type="number"
          min={1}
          value={localQuantity}
          onChange={(e) => handleQuantityChange(parseInt(e.target.value) || 1)}
          className="w-16 text-center"
        />
        <Button
          variant="outline"
          size="icon"
          onClick={() => handleQuantityChange(localQuantity + 1)}
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      <Button
        variant="ghost"
        size="icon"
        onClick={() => onRemove(item.uniqueName)}
        className="text-destructive hover:text-destructive"
      >
        <Trash2 className="h-4 w-4" />
      </Button>
    </div>
  )
}
