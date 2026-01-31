import { useState, useCallback } from 'react'
import { cn } from '@/lib/utils'

// Simple SVG placeholder as a data URI - cannot fail to load
const PLACEHOLDER_SVG = `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='64' height='64' viewBox='0 0 64 64'%3E%3Crect width='64' height='64' fill='%23374151'/%3E%3Ctext x='32' y='36' text-anchor='middle' fill='%236b7280' font-size='12'%3E?%3C/text%3E%3C/svg%3E`

interface ItemImageProps {
  src: string | undefined
  alt: string
  className?: string
}

export function ItemImage({ src, alt, className }: ItemImageProps) {
  const [failed, setFailed] = useState(false)

  const handleError = useCallback(() => {
    setFailed(true)
  }, [])

  const imageSrc = failed || !src ? PLACEHOLDER_SVG : src

  return (
    <img
      src={imageSrc}
      alt={alt}
      className={cn('object-contain', className)}
      onError={handleError}
    />
  )
}

// Helper function to build CDN image URLs
export function getItemImageUrl(imageName: string | undefined): string {
  if (!imageName) return PLACEHOLDER_SVG
  return `https://cdn.warframestat.us/img/${imageName}`
}
