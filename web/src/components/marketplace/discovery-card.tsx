import { Link } from '@tanstack/react-router'

import type { DiscoveryResult } from '../../api/client'
import { navigationEntryPoint } from '../../features/discovery/navigation-state'
import { Price } from '../ui/price'
import { ResponsiveImage } from '../ui/responsive-image'
import { bathroomLabel, kitchenLabel, unitCategoryLabel } from './labels'

type DiscoveryCardProps = {
  result: DiscoveryResult
}

export function DiscoveryCard({ result }: DiscoveryCardProps) {
  const location = [result.property.area, result.property.landmark].filter(Boolean).join(' · ')
  const offerCount = result.offer_summary.available_offer_count
  const category = unitCategoryLabel(result.unit_type.category)
  const unitName = result.unit_type.name.trim() || category
  const hasCustomName = unitName.toLocaleLowerCase() !== category.toLocaleLowerCase()

  return (
    <article className="bg-surface rounded-card h-full overflow-hidden">
      <Link
        to="/properties/$propertyId/unit-types/$unitTypeId"
        params={{ propertyId: result.property.id, unitTypeId: result.unit_type.id }}
        state={(current) => ({
          ...current,
          discoveryIndex: current.__TSR_index,
          entryPoint: navigationEntryPoint.discovery,
        })}
        className="focus-ring group rounded-card flex h-full flex-col"
      >
        <div className="bg-surface-strong relative aspect-[16/10] overflow-hidden">
          <ResponsiveImage
            src={result.thumbnail_url}
            alt={`${unitName} at ${result.property.name}`}
            className="duration-standard h-full w-full object-cover transition-transform group-hover:scale-[1.015]"
          />
        </div>
        {result.thumbnail_source === 'property' ? (
          <p className="font-body text-muted px-4 pt-2 text-xs">Property context photo</p>
        ) : null}

        <div className="flex flex-1 flex-col px-4 pt-3.5 pb-4">
          {hasCustomName ? <p className="section-label">{category}</p> : null}
          <h3 className="font-display text-foreground line-clamp-2 text-base leading-snug font-bold tracking-[-0.025em]">
            {unitName}
          </h3>
          <p className="font-body text-foreground mt-1 text-sm">
            <span className="text-muted">at </span>
            <span className="font-semibold">{result.property.name}</span>
          </p>
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

          <div className="border-border mt-auto flex items-end justify-between gap-4 border-t pt-3">
            <Price amount={result.pricing.lowest_price_naira} from />
            <p className="font-body text-foreground text-right text-xs font-semibold">
              {offerCount > 0
                ? `${offerCount} available ${offerCount === 1 ? 'offer' : 'offers'}`
                : 'No available offers'}
            </p>
          </div>
        </div>
      </Link>
    </article>
  )
}
