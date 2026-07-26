import type { PropertyDetail as PropertyDetailData } from '../../api/client'
import { PropertySummary } from '../../components/marketplace/property-summary'
import { UnitTypeCard } from '../../components/marketplace/unit-type-card'
import { MediaPlaceholder } from '../../components/ui/media-placeholder'
import { ResponsiveImage } from '../../components/ui/responsive-image'

type PropertyDetailProps = {
  property: PropertyDetailData
}

export function PropertyDetail({ property }: PropertyDetailProps) {
  const propertyImages = property.media.filter((media) => media.kind === 'image')

  return (
    <article className="grid gap-7 lg:grid-cols-[minmax(0,1.1fr)_minmax(22rem,0.9fr)] lg:gap-10">
      <div className="min-w-0">
        <PropertyGallery propertyName={property.name} media={propertyImages} />
        <div className="mt-6">
          <PropertySummary property={property} />
        </div>
      </div>

      <section aria-labelledby="unit-types-heading" className="min-w-0">
        <div className="mb-4">
          <p className="section-label">Compare structure first</p>
          <h2
            id="unit-types-heading"
            className="font-display text-foreground text-2xl font-bold tracking-[-0.035em]"
          >
            Unit types
          </h2>
          <p className="font-body text-muted mt-2 text-sm leading-6">
            Choose a unit type to compare the independent agent offers attached to it.
          </p>
        </div>

        {property.unit_types.length > 0 ? (
          <div className="space-y-3">
            {property.unit_types.map((unit) => (
              <UnitTypeCard key={unit.id} propertyId={property.id} unit={unit} />
            ))}
          </div>
        ) : (
          <div className="bg-surface rounded-card px-5 py-8 text-center">
            <p className="font-display text-foreground text-base font-bold">
              No unit types are public yet
            </p>
            <p className="font-body text-muted mt-2 text-sm">
              Structural accommodation details will appear here when they are available.
            </p>
          </div>
        )}
      </section>
    </article>
  )
}

type PropertyGalleryProps = {
  propertyName: string
  media: PropertyDetailData['media']
}

function PropertyGallery({ propertyName, media }: PropertyGalleryProps) {
  if (media.length === 0) {
    return (
      <MediaPlaceholder
        label={`No property photos available for ${propertyName}`}
        className="rounded-media aspect-[4/3]"
      />
    )
  }

  const [hero, ...additional] = media
  return (
    <section aria-label={`${propertyName} property photos`}>
      <ResponsiveImage
        src={hero?.medium_url}
        alt={hero?.caption || `${propertyName} property`}
        loading="eager"
        className="bg-surface-strong rounded-media aspect-[4/3] w-full object-cover"
        sizes="(max-width: 1023px) 100vw, 55vw"
      />
      {additional.length > 0 ? (
        <div className="mt-2 grid grid-cols-3 gap-2">
          {additional.slice(0, 3).map((mediaItem) => (
            <ResponsiveImage
              key={mediaItem.id}
              src={mediaItem.thumbnail_url}
              alt={mediaItem.caption || `${propertyName} property`}
              className="bg-surface-strong rounded-control aspect-[4/3] w-full object-cover"
              sizes="180px"
            />
          ))}
        </div>
      ) : null}
    </section>
  )
}
