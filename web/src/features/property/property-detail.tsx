import type { PropertyDetail as PropertyDetailData } from '../../api/client'
import { MediaGallery } from '../../components/marketplace/media-gallery'
import { PropertySummary } from '../../components/marketplace/property-summary'
import { UnitTypeCard } from '../../components/marketplace/unit-type-card'

type PropertyDetailProps = {
  property: PropertyDetailData
}

export function PropertyDetail({ property }: PropertyDetailProps) {
  const propertyImages = property.media.filter((media) => media.kind === 'image')

  return (
    <article className="grid gap-7 lg:grid-cols-[minmax(0,1.1fr)_minmax(22rem,0.9fr)] lg:gap-10">
      <div className="min-w-0">
        <MediaGallery
          media={propertyImages}
          emptyLabel={`No property photos available for ${property.name}`}
          ariaLabel={`${property.name} property photos`}
          altFallback={`${property.name} property`}
          heroSizes="(max-width: 1023px) 100vw, 55vw"
          thumbnailSizes="180px"
        />
        <div className="mt-6">
          <PropertySummary property={property} />
        </div>
      </div>

      <section aria-labelledby="accommodation-options-heading" className="min-w-0">
        <div className="mb-4">
          <p className="section-label">Compare accommodation first</p>
          <h2
            id="accommodation-options-heading"
            className="font-display text-foreground text-2xl font-bold tracking-[-0.035em]"
          >
            Accommodation options at {property.name}
          </h2>
          <p className="font-body text-muted mt-2 text-sm leading-6">
            Choose an accommodation type to compare its independent agent offers.
          </p>
        </div>

        {property.unit_types.length > 0 ? (
          <div className="space-y-3">
            {property.unit_types.map((unit) => (
              <UnitTypeCard
                key={unit.id}
                propertyId={property.id}
                unit={unit}
              />
            ))}
          </div>
        ) : (
          <div className="bg-surface rounded-card px-5 py-8 text-center">
            <p className="font-display text-foreground text-base font-bold">
              No accommodation types are public yet
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
