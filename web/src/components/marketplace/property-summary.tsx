import type { PropertyDetail } from '../../api/client'

type PropertySummaryProps = {
  property: PropertyDetail
}

export function PropertySummary({ property }: PropertySummaryProps) {
  const location = [property.area, property.landmark].filter(Boolean).join(' · ')

  return (
    <section aria-labelledby="property-name">
      <p className="section-label">Property</p>
      <h1
        id="property-name"
        className="font-display text-foreground text-3xl leading-none font-bold tracking-[-0.045em] sm:text-4xl"
      >
        {property.name}
      </h1>
      <p className="font-body text-muted mt-2 text-sm">
        {location || 'Approximate location unavailable'}
      </p>

      <dl className="mt-5 grid grid-cols-2 gap-3">
        <div className="bg-surface rounded-control px-3.5 py-3">
          <dt className="font-body text-faint text-xs">Area</dt>
          <dd className="font-body text-foreground mt-1 text-sm font-semibold">
            {property.area || 'Unknown'}
          </dd>
        </div>
        <div className="bg-surface rounded-control px-3.5 py-3">
          <dt className="font-body text-faint text-xs">Unit types</dt>
          <dd className="text-foreground mt-1 font-mono text-sm font-semibold">
            {property.unit_types.length}
          </dd>
        </div>
      </dl>

      <div className="border-border mt-5 border-t pt-5">
        <h2 className="font-display text-foreground text-base font-bold">About this property</h2>
        <p className="font-body text-muted mt-2 text-sm leading-6">
          {property.description || 'Building details are not available yet.'}
        </p>
      </div>
    </section>
  )
}
