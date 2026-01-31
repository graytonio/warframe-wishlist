import { useState, useEffect, useCallback } from 'react'
import { api, type OwnedBlueprints, type ItemSearchResult } from '@/lib/api'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useToast } from '@/hooks/use-toast'
import { Search, Plus, Trash2, RefreshCw, AlertTriangle } from 'lucide-react'
import { ItemImage, getItemImageUrl } from '@/components/ui/item-image'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'

export default function ProfileSettings() {
  const [ownedBlueprints, setOwnedBlueprints] = useState<OwnedBlueprints | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<ItemSearchResult[]>([])
  const [searchLoading, setSearchLoading] = useState(false)
  const { toast } = useToast()

  const fetchOwnedBlueprints = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await api.getOwnedBlueprints()
      setOwnedBlueprints(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load owned blueprints')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchOwnedBlueprints()
  }, [fetchOwnedBlueprints])

  const searchReusableBlueprints = useCallback(async (query: string) => {
    if (!query.trim()) {
      setSearchResults([])
      return
    }

    setSearchLoading(true)
    try {
      const data = await api.searchReusableBlueprints(query)
      setSearchResults(data)
    } catch {
      toast({
        title: 'Search failed',
        description: 'Unable to search blueprints. Please try again.',
        variant: 'destructive',
      })
    } finally {
      setSearchLoading(false)
    }
  }, [toast])

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      searchReusableBlueprints(searchQuery)
    }, 300)

    return () => clearTimeout(timeoutId)
  }, [searchQuery, searchReusableBlueprints])

  const handleAddBlueprint = async (uniqueName: string) => {
    try {
      await api.addOwnedBlueprint(uniqueName)
      toast({
        title: 'Blueprint added',
        description: 'Blueprint has been added to your owned list.',
      })
      fetchOwnedBlueprints()
    } catch (err) {
      toast({
        title: 'Failed to add blueprint',
        description: err instanceof Error ? err.message : 'Please try again.',
        variant: 'destructive',
      })
    }
  }

  const handleRemoveBlueprint = async (uniqueName: string) => {
    try {
      await api.removeOwnedBlueprint(uniqueName)
      setOwnedBlueprints((prev) => {
        if (!prev) return prev
        return {
          ...prev,
          blueprints: prev.blueprints.filter((bp) => bp.uniqueName !== uniqueName),
        }
      })
      toast({
        title: 'Blueprint removed',
        description: 'Blueprint has been removed from your owned list.',
      })
    } catch (err) {
      toast({
        title: 'Failed to remove blueprint',
        description: err instanceof Error ? err.message : 'Please try again.',
        variant: 'destructive',
      })
    }
  }

  const handleClearAll = async () => {
    try {
      await api.clearOwnedBlueprints()
      setOwnedBlueprints((prev) => {
        if (!prev) return prev
        return { ...prev, blueprints: [] }
      })
      toast({
        title: 'All blueprints cleared',
        description: 'All owned blueprints have been removed.',
      })
    } catch (err) {
      toast({
        title: 'Failed to clear blueprints',
        description: err instanceof Error ? err.message : 'Please try again.',
        variant: 'destructive',
      })
    }
  }

  const isOwned = (uniqueName: string) => {
    return ownedBlueprints?.blueprints.some((bp) => bp.uniqueName === uniqueName) ?? false
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
          <Button onClick={fetchOwnedBlueprints}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    )
  }

  const blueprints = ownedBlueprints?.blueprints || []

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-bold">Profile Settings</h1>
        <Button variant="outline" size="icon" onClick={fetchOwnedBlueprints}>
          <RefreshCw className="h-4 w-4" />
        </Button>
      </div>

      <div className="grid lg:grid-cols-2 gap-8">
        <Card>
          <CardHeader>
            <CardTitle>Add Owned Blueprints</CardTitle>
            <CardDescription>
              Search for reusable blueprints you already own. These will be excluded from your materials summary.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="relative mb-4">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                type="search"
                placeholder="Search reusable blueprints..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>

            {searchLoading ? (
              <div className="flex items-center justify-center py-8">
                <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
              </div>
            ) : searchResults.length > 0 ? (
              <div className="space-y-2 max-h-96 overflow-y-auto">
                {searchResults.map((item) => (
                  <div
                    key={item.uniqueName}
                    className="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-accent/50 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <ItemImage
                        src={getItemImageUrl(item.imageName)}
                        alt={item.name}
                        className="w-8 h-8"
                      />
                      <div>
                        <p className="font-medium text-sm">{item.name}</p>
                        <p className="text-xs text-muted-foreground">{item._collection}</p>
                      </div>
                    </div>
                    <Button
                      size="sm"
                      variant={isOwned(item.uniqueName) ? 'secondary' : 'default'}
                      disabled={isOwned(item.uniqueName)}
                      onClick={() => handleAddBlueprint(item.uniqueName)}
                    >
                      {isOwned(item.uniqueName) ? 'Owned' : (
                        <>
                          <Plus className="h-4 w-4 mr-1" />
                          Add
                        </>
                      )}
                    </Button>
                  </div>
                ))}
              </div>
            ) : searchQuery ? (
              <p className="text-center text-muted-foreground py-8">
                No reusable blueprints found
              </p>
            ) : (
              <p className="text-center text-muted-foreground py-8">
                Search for blueprints to add
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Owned Blueprints ({blueprints.length})</CardTitle>
                <CardDescription>
                  These reusable blueprints will be excluded from your wishlist materials.
                </CardDescription>
              </div>
              {blueprints.length > 0 && (
                <Dialog>
                  <DialogTrigger asChild>
                    <Button variant="destructive" size="sm">
                      <Trash2 className="h-4 w-4 mr-1" />
                      Clear All
                    </Button>
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle className="flex items-center gap-2">
                        <AlertTriangle className="h-5 w-5 text-destructive" />
                        Clear All Owned Blueprints?
                      </DialogTitle>
                      <DialogDescription>
                        This will remove all {blueprints.length} blueprints from your owned list.
                        This action cannot be undone.
                      </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                      <Button variant="outline" onClick={() => {}}>Cancel</Button>
                      <Button variant="destructive" onClick={handleClearAll}>
                        Clear All
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              )}
            </div>
          </CardHeader>
          <CardContent>
            {blueprints.length === 0 ? (
              <p className="text-center text-muted-foreground py-8">
                No owned blueprints yet. Search and add some!
              </p>
            ) : (
              <div className="space-y-2 max-h-96 overflow-y-auto">
                {blueprints.map((bp) => (
                  <div
                    key={bp.uniqueName}
                    className="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-accent/50 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded bg-muted flex items-center justify-center">
                        <span className="text-xs font-medium">BP</span>
                      </div>
                      <div>
                        <p className="font-medium text-sm truncate max-w-[200px]">
                          {bp.uniqueName.split('/').pop()}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          Added {new Date(bp.addedAt).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => handleRemoveBlueprint(bp.uniqueName)}
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
