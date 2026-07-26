import { BathIcon, BedIcon, KitchenUtensilsIcon, Layout01Icon } from '@hugeicons/core-free-icons'

import type { UnitTypeDetail } from '../../api/client'
import { ProductIcon } from '../ui/product-icon'
import { bathroomLabel, kitchenLabel } from './labels'

type UnitFactsProps = {
  unit: UnitTypeDetail
}

export function UnitFacts({ unit }: UnitFactsProps) {
  const facts = [
    {
      label: 'Bedrooms',
      value: unit.bedroom_count === null ? 'Unknown' : String(unit.bedroom_count),
      icon: BedIcon,
    },
    {
      label: 'Bathroom',
      value: bathroomLabel(unit.bathroom_type).replace(' bathroom', ''),
      icon: BathIcon,
    },
    {
      label: 'Kitchen',
      value: kitchenLabel(unit.kitchen_type).replace(' kitchen', ''),
      icon: KitchenUtensilsIcon,
    },
    {
      label: 'Parlour',
      value: unit.has_parlour === null ? 'Unknown' : unit.has_parlour ? 'Yes' : 'No',
      icon: Layout01Icon,
    },
  ]

  return (
    <dl className="bg-surface rounded-card grid grid-cols-2 gap-3 p-4">
      {facts.map((fact) => (
        <div key={fact.label} className="flex min-w-0 items-center gap-2.5">
          <span className="bg-background text-muted rounded-control flex size-9 shrink-0 items-center justify-center">
            <ProductIcon icon={fact.icon} size={16} />
          </span>
          <div className="min-w-0">
            <dt className="font-body text-faint text-xs">{fact.label}</dt>
            <dd className="font-body text-foreground mt-0.5 truncate text-xs font-semibold capitalize">
              {fact.value}
            </dd>
          </div>
        </div>
      ))}
    </dl>
  )
}
