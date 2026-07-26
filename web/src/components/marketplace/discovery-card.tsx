import { Link } from '@tanstack/react-router'

import type { DiscoveryResult } from '../../api/client'
import { Price } from '../ui/price'
import { ResponsiveImage } from '../ui/responsive-image'
import { bathroomLabel, kitchenLabel, unitCategoryLabel } from './labels'

type DiscoveryCardProps = {
  result: DiscoveryResult
}

export function DiscoveryCard({ result }: DiscoveryCardProps) {
  const location = [result.property.area, result.property.landmark].filter(Boolean).join(' · ')
  const offerCount = result.offer_summary.available_offer_count

  return (
    <article className="bg-surface rounded-card overflow-hidden">
      <Link
        to="/properties/$propertyId"
        params={{ propertyId: result.property.id }}
        className="focus-ring group rounded-card block"
      >
        <div className="bg-surface-strong relative aspect-[16/10] overflow-hidden">
          <ResponsiveImage
            src={result.thumbnail_url}
            alt={`${result.property.name} accommodation`}
            className="duration-standard h-full w-full object-cover transition-transform group-hover:scale-[1.015]"
          />
          <span className="font-body rounded-control absolute bottom-3 left-3 bg-black/80 px-2.5 py-1 text-xs font-medium text-white backdrop-blur-sm">
            {unitCategoryLabel(result.unit_type.category)}
          </span>
        </div>

        <div className="px-4 pt-3.5 pb-4">
          <h3 className="font-display text-foreground text-base leading-snug font-bold tracking-[-0.025em]">
            {result.property.name}
          </h3>
          <p className="font-body text-muted mt-1 text-xs">
            {location || 'Approximate location unavailable'}
          </p>

          <p className="font-body text-muted mt-3 flex flex-wrap gap-x-3 gap-y-1 text-xs">
            <span>{bathroomLabel(result.unit_type.bathroom_type)}</span>
            <span>{kitchenLabel(result.unit_type.kitchen_type)}</span>
            {result.unit_type.bedroom_count ? (
              <span>
                {result.unit_type.bedroom_count}{' '}
                {result.unit_type.bedroom_count === 1 ? 'bedroom' : 'bedrooms'}
              </span>
            ) : null}
          </p>

          <div className="border-border mt-3.5 flex items-end justify-between gap-4 border-t pt-3">
            <Price amount={result.pricing.lowest_price_naira} from />
            <div className="font-body text-right text-xs">
              <p className="text-foreground font-semibold">
                {offerCount} {offerCount === 1 ? 'offer' : 'offers'}
              </p>
              <p className={offerCount > 0 ? 'text-positive mt-1' : 'text-faint mt-1'}>
                {offerCount > 0 ? 'Available now' : 'No available offers'}
              </p>
            </div>
          </div>
        </div>
      </Link>
    </article>
  )
}
