import { useState, useEffect, useMemo } from 'react'
import type { Item, Component } from '@/lib/api'
import { api } from '@/lib/api'

interface RawMaterial {
  uniqueName: string
  name: string
  imageName?: string
  totalCount: number
}

// Recursively resolve raw materials (components with no sub-components)
function resolveRawMaterials(components: Component[] | undefined, multiplier = 1): Map<string, RawMaterial> {
  const materials = new Map<string, RawMaterial>()

  if (!components) return materials

  for (const comp of components) {
    const count = comp.itemCount * multiplier

    if (comp.components && comp.components.length > 0) {
      // Has sub-components, recurse
      const subMaterials = resolveRawMaterials(comp.components, count)
      for (const [key, mat] of subMaterials) {
        const existing = materials.get(key)
        if (existing) {
          existing.totalCount += mat.totalCount
        } else {
          materials.set(key, { ...mat })
        }
      }
    } else {
      // Leaf material (no sub-components)
      const existing = materials.get(comp.uniqueName)
      if (existing) {
        existing.totalCount += count
      } else {
        materials.set(comp.uniqueName, {
          uniqueName: comp.uniqueName,
          name: comp.name,
          imageName: comp.imageName,
          totalCount: count,
        })
      }
    }
  }

  return materials
}
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ItemImage, getItemImageUrl } from '@/components/ui/item-image'
import { Plus, Minus, ExternalLink, ArrowLeft } from 'lucide-react'

interface ItemDetailDialogProps {
  uniqueName: string | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onAddToWishlist?: (uniqueName: string, quantity: number) => void
}

export function ItemDetailDialog({
  uniqueName,
  open,
  onOpenChange,
  onAddToWishlist,
}: ItemDetailDialogProps) {
  const [item, setItem] = useState<Item | null>(null)
  const [loading, setLoading] = useState(false)
  const [quantity, setQuantity] = useState(1)
  const [itemStack, setItemStack] = useState<string[]>([])

  // Track the current item being viewed (either from stack or initial uniqueName)
  const currentUniqueName = itemStack.length > 0 ? itemStack[itemStack.length - 1] : uniqueName

  useEffect(() => {
    if (currentUniqueName && open) {
      setLoading(true)
      setQuantity(1)
      api.getItem(currentUniqueName)
        .then(setItem)
        .catch(() => setItem(null))
        .finally(() => setLoading(false))
    }
  }, [currentUniqueName, open])

  // Reset stack when dialog closes or initial uniqueName changes
  useEffect(() => {
    if (!open) {
      setItemStack([])
    }
  }, [open, uniqueName])

  const handleComponentClick = (componentUniqueName: string) => {
    setItemStack([...itemStack, componentUniqueName])
  }

  const handleBack = () => {
    setItemStack(itemStack.slice(0, -1))
  }

  // Compute raw materials (leaf components with no sub-components)
  const rawMaterials = useMemo(() => {
    if (!item?.components) return []
    const materialsMap = resolveRawMaterials(item.components)
    // Sort by quantity (descending), then by name
    return Array.from(materialsMap.values()).sort((a, b) => {
      if (b.totalCount !== a.totalCount) {
        return b.totalCount - a.totalCount
      }
      return a.name.localeCompare(b.name)
    })
  }, [item])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl lg:max-w-4xl max-h-[80vh] lg:max-h-[90vh] overflow-y-auto">
        {loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        ) : item ? (
          <>
            <DialogHeader>
              <div className="flex items-center gap-4">
                {itemStack.length > 0 && (
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={handleBack}
                    className="shrink-0"
                  >
                    <ArrowLeft className="h-4 w-4" />
                  </Button>
                )}
                <ItemImage
                  src={getItemImageUrl(item.imageName)}
                  alt={item.name}
                  className="w-16 h-16"
                />
                <div>
                  <DialogTitle className="text-xl">{item.name}</DialogTitle>
                  <DialogDescription>
                    {item.category}
                    {item.isPrime && ' • Prime'}
                    {item.masteryReq ? ` • MR ${item.masteryReq}` : ''}
                  </DialogDescription>
                </div>
              </div>
            </DialogHeader>

            <div className="space-y-4">
              {item.description && (
                <p className="text-sm text-muted-foreground">{item.description}</p>
              )}

              {item.buildPrice && (
                <div className="flex items-center gap-2 text-sm">
                  <span className="font-medium">Build Cost:</span>
                  <span>{item.buildPrice.toLocaleString()} Credits</span>
                </div>
              )}

              {item.components && item.components.length > 0 && (
                <div>
                  <h4 className="font-medium mb-2">Components Required</h4>
                  <div className="grid grid-cols-2 gap-2">
                    {item.components.map((comp) =>
                      comp.hasOwnPage ? (
                        <button
                          key={comp.uniqueName}
                          type="button"
                          onClick={() => handleComponentClick(comp.uniqueName)}
                          className="flex items-center gap-2 p-2 bg-muted rounded-md hover:bg-muted/80 transition-colors cursor-pointer text-left"
                        >
                          <ItemImage
                            src={getItemImageUrl(comp.imageName)}
                            alt={comp.name}
                            className="w-8 h-8"
                          />
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium truncate">{comp.name}</p>
                            <p className="text-xs text-muted-foreground">
                              x{comp.itemCount}
                            </p>
                          </div>
                        </button>
                      ) : (
                        <div
                          key={comp.uniqueName}
                          className="flex items-center gap-2 p-2 bg-muted rounded-md"
                        >
                          <ItemImage
                            src={getItemImageUrl(comp.imageName)}
                            alt={comp.name}
                            className="w-8 h-8"
                          />
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium truncate">{comp.name}</p>
                            <p className="text-xs text-muted-foreground">
                              x{comp.itemCount}
                            </p>
                          </div>
                        </div>
                      )
                    )}
                  </div>
                </div>
              )}

              {rawMaterials.length > 0 && (
                <div>
                  <h4 className="font-medium mb-2">Raw Materials Needed</h4>
                  <div className="grid grid-cols-2 lg:grid-cols-3 gap-2">
                    {rawMaterials.map((material) => (
                      <div
                        key={material.uniqueName}
                        className="flex items-center gap-2 p-2 bg-muted rounded-md"
                      >
                        <ItemImage
                          src={getItemImageUrl(material.imageName)}
                          alt={material.name}
                          className="w-8 h-8"
                        />
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium truncate">{material.name}</p>
                          <p className="text-xs text-muted-foreground">
                            x{material.totalCount.toLocaleString()}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {item.drops && item.drops.length > 0 && (
                <div>
                  <h4 className="font-medium mb-2">Drop Locations</h4>
                  <div className="space-y-1 max-h-40 overflow-y-auto">
                    {[...item.drops]
                      .sort((a, b) => (b.chance ?? 0) - (a.chance ?? 0))
                      .slice(0, 10)
                      .map((drop, idx) => (
                        <div
                          key={idx}
                          className="text-sm p-2 bg-muted rounded-md flex justify-between"
                        >
                          <span>{drop.location}</span>
                          {drop.chance && (
                            <span className="text-muted-foreground">
                              {(drop.chance * 100).toFixed(1)}%
                            </span>
                          )}
                        </div>
                      ))}
                    {item.drops.length > 10 && (
                      <p className="text-xs text-muted-foreground text-center">
                        +{item.drops.length - 10} more locations
                      </p>
                    )}
                  </div>
                </div>
              )}

              {item.wikiaUrl && (
                <a
                  href={item.wikiaUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
                >
                  View on Wiki <ExternalLink className="h-3 w-3" />
                </a>
              )}

              {onAddToWishlist && (
                <div className="flex items-end gap-4 pt-4 border-t">
                  <div className="space-y-2">
                    <Label htmlFor="quantity">Quantity</Label>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="icon"
                        onClick={() => setQuantity(Math.max(1, quantity - 1))}
                      >
                        <Minus className="h-4 w-4" />
                      </Button>
                      <Input
                        id="quantity"
                        type="number"
                        min={1}
                        value={quantity}
                        onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value) || 1))}
                        className="w-20 text-center"
                      />
                      <Button
                        variant="outline"
                        size="icon"
                        onClick={() => setQuantity(quantity + 1)}
                      >
                        <Plus className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                  <Button
                    className="flex-1"
                    onClick={() => {
                      onAddToWishlist(item.uniqueName, quantity)
                      onOpenChange(false)
                    }}
                  >
                    <Plus className="h-4 w-4 mr-2" />
                    Add to Wishlist
                  </Button>
                </div>
              )}
            </div>
          </>
        ) : (
          <div className="py-8 text-center text-muted-foreground">
            Item not found
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
