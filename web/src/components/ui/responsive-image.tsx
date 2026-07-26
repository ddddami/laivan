import { useState } from 'react'

import { MediaPlaceholder } from './media-placeholder'

type ResponsiveImageProps = {
  src?: string | null
  alt: string
  className?: string
  loading?: 'eager' | 'lazy'
  sizes?: string
}

export function ResponsiveImage({
  src,
  alt,
  className = '',
  loading = 'lazy',
  sizes = '(max-width: 767px) 100vw, 50vw',
}: ResponsiveImageProps) {
  const [failed, setFailed] = useState(false)

  if (!src || failed) {
    return <MediaPlaceholder className={className} />
  }

  return (
    <img
      src={src}
      alt={alt}
      className={className}
      loading={loading}
      decoding="async"
      sizes={sizes}
      onError={() => setFailed(true)}
    />
  )
}
