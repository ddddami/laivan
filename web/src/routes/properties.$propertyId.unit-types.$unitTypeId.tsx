import { useQuery } from '@tanstack/react-query'
import { createFileRoute, useRouter } from '@tanstack/react-router'
import { ArrowLeft01Icon } from '@hugeicons/core-free-icons'

import { ApiError } from '../api/client'
import { propertyQueryOptions } from '../api/queries'
import { AppShell } from '../components/app-shell'
import { bathroomLabel, kitchenLabel, unitCategoryLabel } from '../components/marketplace/labels'
import { FeedbackState } from '../components/ui/feedback-state'
import { ProductIcon } from '../components/ui/product-icon'
import { Skeleton } from '../components/ui/skeleton'

export const Route = createFileRoute('/properties/$propertyId/unit-types/$unitTypeId')({
  component: UnitTypeTracer,
})

function UnitTypeTracer() {
  const { propertyId, unitTypeId } = Route.useParams()
  const propertyQuery = useQuery(propertyQueryOptions(propertyId))
  const router = useRouter()
  const propertyNotFound =
    propertyQuery.error instanceof ApiError && propertyQuery.error.status === 404
  const unit = propertyQuery.data?.unit_types.find((candidate) => candidate.id === unitTypeId)

  return (
    <AppShell>
      <button
        type="button"
        className="focus-ring font-body text-muted rounded-control mb-5 inline-flex min-h-11 items-center gap-2 pr-3 text-sm font-medium"
        onClick={() => router.history.back()}
      >
        <ProductIcon icon={ArrowLeft01Icon} size={18} />
        Back to property
      </button>

      {propertyQuery.isPending ? (
        <div className="space-y-4" aria-label="Loading unit type">
          <Skeleton className="h-9 w-2/3" />
          <Skeleton className="rounded-card h-40 w-full" />
        </div>
      ) : null}

      {propertyQuery.isError ? (
        <FeedbackState
          title={propertyNotFound ? 'Property not found' : 'Unit type unavailable'}
          description={
            propertyNotFound
              ? 'This property may no longer be part of the public FUTA marketplace.'
              : 'We could not load this unit type. Check your connection and try again.'
          }
          actionLabel={propertyNotFound ? undefined : 'Try again'}
          onAction={propertyNotFound ? undefined : () => void propertyQuery.refetch()}
        />
      ) : null}

      {propertyQuery.data && !unit ? (
        <FeedbackState
          title="Unit type not found"
          description={`This unit type is not part of ${propertyQuery.data.name}.`}
        />
      ) : null}

      {propertyQuery.data && unit ? (
        <article className="bg-surface rounded-card p-5 sm:p-8">
          <p className="section-label">{unitCategoryLabel(unit.category)}</p>
          <h1 className="font-display text-foreground text-3xl font-bold tracking-[-0.04em]">
            {unit.name || unitCategoryLabel(unit.category)}
          </h1>
          <p className="font-body text-muted mt-2 text-sm">
            {propertyQuery.data.name} · {propertyQuery.data.area}
          </p>
          <p className="font-body text-muted mt-5 text-sm">
            {bathroomLabel(unit.bathroom_type)} · {kitchenLabel(unit.kitchen_type)}
          </p>
        </article>
      ) : null}
    </AppShell>
  )
}
