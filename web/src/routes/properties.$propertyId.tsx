import { useQuery } from '@tanstack/react-query'
import { createFileRoute, useNavigate, useRouter } from '@tanstack/react-router'
import { ArrowLeft01Icon } from '@hugeicons/core-free-icons'

import { ApiError } from '../api/client'
import { propertyQueryOptions } from '../api/queries'
import { AppShell } from '../components/app-shell'
import { FeedbackState } from '../components/ui/feedback-state'
import { ProductIcon } from '../components/ui/product-icon'
import { Skeleton } from '../components/ui/skeleton'
import { PropertyDetail } from '../features/property/property-detail'

export const Route = createFileRoute('/properties/$propertyId')({
  component: PropertyPage,
})

function PropertyPage() {
  const { propertyId } = Route.useParams()
  const propertyQuery = useQuery(propertyQueryOptions(propertyId))
  const router = useRouter()
  const navigate = useNavigate()
  const notFound = propertyQuery.error instanceof ApiError && propertyQuery.error.status === 404

  function backToResults() {
    if (window.history.length > 1) {
      router.history.back()
      return
    }
    void navigate({
      to: '/',
      search: { availability: 'available', sort: 'recommended', page: 1 },
    })
  }

  return (
    <AppShell>
      <button
        type="button"
        className="focus-ring font-body text-muted rounded-control mb-5 inline-flex min-h-11 items-center gap-2 pr-3 text-sm font-medium"
        onClick={backToResults}
      >
        <ProductIcon icon={ArrowLeft01Icon} size={18} />
        Back to results
      </button>

      {propertyQuery.isPending ? <PropertySkeleton /> : null}

      {propertyQuery.isError ? (
        <FeedbackState
          title={notFound ? 'Property not found' : 'Property unavailable'}
          description={
            notFound
              ? 'This property may no longer be part of the public FUTA marketplace.'
              : 'We could not load this property. Check your connection and try again.'
          }
          actionLabel={notFound ? undefined : 'Try again'}
          onAction={notFound ? undefined : () => void propertyQuery.refetch()}
        />
      ) : null}

      {propertyQuery.data ? <PropertyDetail property={propertyQuery.data} /> : null}
    </AppShell>
  )
}

function PropertySkeleton() {
  return (
    <div
      className="grid gap-7 lg:grid-cols-[minmax(0,1.1fr)_minmax(22rem,0.9fr)]"
      aria-label="Loading property"
    >
      <div className="space-y-4">
        <Skeleton className="rounded-media aspect-[4/3] w-full" />
        <Skeleton className="h-9 w-2/3" />
        <Skeleton className="h-4 w-1/2" />
        <Skeleton className="h-24 w-full" />
      </div>
      <div className="space-y-3">
        <Skeleton className="h-8 w-1/2" />
        <Skeleton className="rounded-card h-40 w-full" />
        <Skeleton className="rounded-card h-40 w-full" />
      </div>
    </div>
  )
}
