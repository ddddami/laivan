import type { Media } from '../../api/client'
import { MediaPlaceholder } from '../ui/media-placeholder'
import { ResponsiveImage } from '../ui/responsive-image'

type MediaGalleryProps = {
  media: readonly Media[]
  emptyLabel: string
  ariaLabel: string
  altFallback: string
  heroSizes: string
  thumbnailSizes: string
  notice?: {
    label: string
    description: string
  }
}

export function MediaGallery({
  media,
  emptyLabel,
  ariaLabel,
  altFallback,
  heroSizes,
  thumbnailSizes,
  notice,
}: MediaGalleryProps) {
  if (media.length === 0) {
    return <MediaPlaceholder label={emptyLabel} className="rounded-media aspect-[4/3]" />
  }

  const [hero, ...additional] = media
  return (
    <section aria-label={ariaLabel}>
      {notice ? (
        <>
          <p className="section-label">{notice.label}</p>
          <p className="font-body text-muted mt-1 text-xs leading-5">{notice.description}</p>
        </>
      ) : null}
      <ResponsiveImage
        src={hero?.medium_url}
        alt={hero?.caption || altFallback}
        loading="eager"
        className="bg-surface-strong rounded-media aspect-[4/3] w-full object-cover"
        sizes={heroSizes}
      />
      {additional.length > 0 ? (
        <div className="mt-2 grid grid-cols-3 gap-2">
          {additional.slice(0, 3).map((mediaItem) => (
            <ResponsiveImage
              key={mediaItem.id}
              src={mediaItem.thumbnail_url}
              alt={mediaItem.caption || altFallback}
              className="bg-surface-strong rounded-control aspect-[4/3] w-full object-cover"
              sizes={thumbnailSizes}
            />
          ))}
        </div>
      ) : null}
    </section>
  )
}
