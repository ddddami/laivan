import { Link } from '@tanstack/react-router'
import { ArrowRight01Icon } from '@hugeicons/core-free-icons'

import type { UnitTypeDetail } from '../../api/client'
import { Price } from '../ui/price'
import { ProductIcon } from '../ui/product-icon'
import { ResponsiveImage } from '../ui/responsive-image'
import { bathroomLabel, kitchenLabel, unitCategoryLabel } from './labels'

type UnitTypeCardProps = {
  propertyId: string
  unit: UnitTypeDetail
}

export function UnitTypeCard({ propertyId, unit }: UnitTypeCardProps) {
  const availableOffers = unit.agent_offers.filter((offer) => offer.status === 'available')
  const lowestPrice =
    availableOffers.length > 0
      ? Math.min(...availableOffers.map((offer) => offer.price_naira))
      : null
  const category = unitCategoryLabel(unit.category)
  const unitImage = unit.media.find((media) => media.kind === 'image')

  return (
    <article className="bg-surface rounded-card overflow-hidden">
      <Link
        to="/properties/$propertyId/unit-types/$unitTypeId"
        params={{ propertyId, unitTypeId: unit.id }}
        className="focus-ring group rounded-card grid min-h-40 grid-cols-[6.5rem_minmax(0,1fr)] sm:grid-cols-[9rem_minmax(0,1fr)]"
      >
        <ResponsiveImage
          src={unitImage?.thumbnail_url}
          alt={`${unit.name || category} at this property`}
          className="bg-surface-strong h-full min-h-40 w-full object-cover"
          sizes="144px"
        />

        <div className="flex min-w-0 flex-col p-4">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="font-body text-faint text-xs">{category}</p>
              <h3 className="font-display text-foreground mt-1 text-base leading-tight font-bold tracking-[-0.025em]">
                {unit.name || category}
              </h3>
            </div>
            <span className="bg-background text-muted rounded-control duration-standard flex size-8 shrink-0 items-center justify-center transition-transform group-hover:translate-x-0.5">
              <ProductIcon icon={ArrowRight01Icon} size={15} />
            </span>
          </div>

          <p className="font-body text-muted mt-3 flex flex-wrap gap-x-3 gap-y-1 text-xs">
            <span>{bathroomLabel(unit.bathroom_type)}</span>
            <span>{kitchenLabel(unit.kitchen_type)}</span>
            {unit.has_parlour !== null ? (
              <span>{unit.has_parlour ? 'Has parlour' : 'No parlour'}</span>
            ) : null}
          </p>

          <div className="border-border mt-auto flex items-end justify-between gap-3 border-t pt-3">
            <Price amount={lowestPrice} from />
            <p className="font-body text-foreground text-right text-xs font-semibold">
              {availableOffers.length}{' '}
              {availableOffers.length === 1 ? 'available offer' : 'available offers'}
            </p>
          </div>
        </div>
      </Link>
    </article>
  )
}
