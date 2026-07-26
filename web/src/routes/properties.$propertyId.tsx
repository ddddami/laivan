import { useQuery } from '@tanstack/react-query'
import { Link, createFileRoute } from '@tanstack/react-router'
import { ArrowLeft01Icon } from '@hugeicons/core-free-icons'

import { ApiError } from '../api/client'
import { propertyQueryOptions } from '../api/queries'
import { AppShell } from '../components/app-shell'
import { Button } from '../components/ui/button'
import { FeedbackState } from '../components/ui/feedback-state'
import { ProductIcon } from '../components/ui/product-icon'
import { Skeleton } from '../components/ui/skeleton'

export const Route = createFileRoute('/properties/$propertyId')({
  component: PropertyTracer,
})

function PropertyTracer() {
  const { propertyId } = Route.useParams()
  const propertyQuery = useQuery(propertyQueryOptions(propertyId))
  const notFound = propertyQuery.error instanceof ApiError && propertyQuery.error.status === 404

  return (
    <AppShell>
      <Link
        to="/"
        search={{ availability: 'available', sort: 'recommended', page: 1 }}
        className="focus-ring font-body text-muted rounded-control mb-5 inline-flex min-h-11 items-center gap-2 pr-3 text-sm font-medium"
      >
        <ProductIcon icon={ArrowLeft01Icon} size={18} />
        Back to results
      </Link>

      {propertyQuery.isPending ? (
        <div className="space-y-3" aria-label="Loading property">
          <Skeleton className="rounded-media h-64 w-full" />
          <Skeleton className="h-8 w-2/3" />
          <Skeleton className="h-4 w-1/2" />
        </div>
      ) : null}

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

      {propertyQuery.data ? (
        <article className="bg-surface rounded-card p-5 sm:p-8">
          <p className="font-body text-faint text-xs font-semibold tracking-[0.1em] uppercase">
            Property
          </p>
          <h1 className="font-display text-foreground mt-2 text-3xl font-bold tracking-[-0.04em]">
            {propertyQuery.data.name}
          </h1>
          <p className="font-body text-muted mt-2 text-sm">
            {[propertyQuery.data.area, propertyQuery.data.landmark].filter(Boolean).join(' · ')}
          </p>
          <p className="font-body text-muted mt-5 max-w-2xl text-sm leading-6">
            {propertyQuery.data.description || 'Building details are not available yet.'}
          </p>
          <Button className="mt-6" disabled>
            Unit comparison arrives in the next slice
          </Button>
        </article>
      ) : null}
    </AppShell>
  )
}
