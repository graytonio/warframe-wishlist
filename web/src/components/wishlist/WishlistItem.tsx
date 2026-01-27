import { useState, useEffect } from 'react'
import type { WishlistItem as WishlistItemType, Item } from '@/lib/api'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Trash2, Minus, Plus } from 'lucide-react'

interface WishlistItemProps {
  item: WishlistItemType
  onUpdateQuantity: (uniqueName: string, quantity: number) => void
  onRemove: (uniqueName: string) => void
}

export function WishlistItemRow({ item, onUpdateQuantity, onRemove }: WishlistItemProps) {
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

  const imageUrl = itemDetails?.imageName
    ? `https://cdn.warframestat.us/img/${itemDetails.imageName}`
    : '/placeholder.png'

  return (
    <div className="flex items-center gap-4 p-4 border rounded-lg">
      <img
        src={imageUrl}
        alt={itemDetails?.name || item.uniqueName}
        className="w-12 h-12 object-contain"
        onError={(e) => {
          e.currentTarget.src = '/placeholder.png'
        }}
      />
      <div className="flex-1 min-w-0">
        <h4 className="font-medium truncate">
          {itemDetails?.name || item.uniqueName}
        </h4>
        {itemDetails?.category && (
          <p className="text-sm text-muted-foreground">{itemDetails.category}</p>
        )}
        {itemDetails?.buildPrice && (
          <p className="text-xs text-muted-foreground">
            {(itemDetails.buildPrice * localQuantity).toLocaleString()} Credits
          </p>
        )}
      </div>
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
