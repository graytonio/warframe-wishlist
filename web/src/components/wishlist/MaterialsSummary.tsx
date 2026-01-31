import { useState, useEffect } from 'react'
import type { MaterialsResponse } from '@/lib/api'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ItemImage, getItemImageUrl } from '@/components/ui/item-image'
import { RefreshCw } from 'lucide-react'

export function MaterialsSummary() {
  const [materials, setMaterials] = useState<MaterialsResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchMaterials = async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await api.getWishlistMaterials()
      setMaterials(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load materials')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchMaterials()
  }, [])

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Total Materials Needed</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-4">
            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Total Materials Needed</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-destructive text-sm">{error}</p>
          <Button variant="outline" size="sm" onClick={fetchMaterials} className="mt-2">
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </CardContent>
      </Card>
    )
  }

  if (!materials || materials.materials.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Total Materials Needed</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-sm">
            Add items to your wishlist to see required materials
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Total Materials Needed</CardTitle>
        <Button variant="ghost" size="icon" onClick={fetchMaterials}>
          <RefreshCw className="h-4 w-4" />
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="p-3 bg-primary/10 rounded-lg">
          <p className="text-sm font-medium">Total Credits</p>
          <p className="text-2xl font-bold">
            {materials.totalCredits.toLocaleString()}
          </p>
        </div>

        <div className="space-y-2">
          <h4 className="text-sm font-medium">Materials</h4>
          <div className="grid gap-2 max-h-96 overflow-y-auto">
            {[...materials.materials].sort((a, b) => b.totalCount - a.totalCount).map((material) => (
              <div
                key={material.uniqueName}
                className="flex items-center gap-3 p-2 bg-muted rounded-md"
              >
                <ItemImage
                  src={getItemImageUrl(material.imageName)}
                  alt={material.name}
                  className="w-8 h-8"
                />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">{material.name}</p>
                </div>
                <span className="text-sm font-bold">
                  x{material.totalCount.toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
