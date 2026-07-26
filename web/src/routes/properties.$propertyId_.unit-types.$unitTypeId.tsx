import { useQuery } from '@tanstack/react-query'
import { createFileRoute, useLocation, useNavigate, useRouter } from '@tanstack/react-router'
import { ArrowLeft01Icon } from '@hugeicons/core-free-icons'

import { ApiError } from '../api/client'
import { propertyQueryOptions } from '../api/queries'
import { AppShell } from '../components/app-shell'
import { FeedbackState } from '../components/ui/feedback-state'
import { ProductIcon } from '../components/ui/product-icon'
import { Skeleton } from '../components/ui/skeleton'
import { UnitDetail } from '../features/unit/unit-detail'

export const Route = createFileRoute('/properties/$propertyId_/unit-types/$unitTypeId')({
  component: UnitTypePage,
})

function UnitTypePage() {
  const { propertyId, unitTypeId } = Route.useParams()
  const location = useLocation()
  const propertyQuery = useQuery(propertyQueryOptions(propertyId))
  const router = useRouter()
  const navigate = useNavigate()
  const propertyNotFound =
    propertyQuery.error instanceof ApiError && propertyQuery.error.status === 404
  const unit = propertyQuery.data?.unit_types.find((candidate) => candidate.id === unitTypeId)

  function backToProperty() {
    if (location.state.__TSR_index > 0) {
      router.history.back()
      return
    }
    void navigate({
      to: '/properties/$propertyId',
      params: { propertyId },
      replace: true,
    })
  }

  return (
    <AppShell>
      <button
        type="button"
        className="focus-ring font-body text-muted rounded-control mb-5 inline-flex min-h-11 items-center gap-2 pr-3 text-sm font-medium"
        onClick={backToProperty}
      >
        <ProductIcon icon={ArrowLeft01Icon} size={18} />
        Back to property
      </button>

      {propertyQuery.isPending ? <UnitSkeleton /> : null}

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

      {propertyQuery.data && unit ? <UnitDetail property={propertyQuery.data} unit={unit} /> : null}
    </AppShell>
  )
}

function UnitSkeleton() {
  return (
    <div aria-label="Loading unit type">
      <Skeleton className="h-9 w-2/3" />
      <Skeleton className="mt-3 h-4 w-1/2" />
      <div className="mt-6 grid gap-8 lg:grid-cols-[minmax(20rem,0.8fr)_minmax(0,1.2fr)]">
        <div className="space-y-4">
          <Skeleton className="rounded-media aspect-[4/3] w-full" />
          <Skeleton className="rounded-card h-40 w-full" />
        </div>
        <div className="space-y-3">
          <Skeleton className="h-8 w-1/2" />
          <Skeleton className="rounded-card h-72 w-full" />
          <Skeleton className="rounded-card h-72 w-full" />
        </div>
      </div>
    </div>
  )
}
