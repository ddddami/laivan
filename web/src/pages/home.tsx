import { useQuery } from '@tanstack/react-query'

import { campusQueryOptions, discoveryQueryOptions } from '../api/queries'
import { AppShell } from '../components/app-shell'
import { DiscoveryCard } from '../components/marketplace/discovery-card'
import { FeedbackState } from '../components/ui/feedback-state'
import { Skeleton } from '../components/ui/skeleton'

export function Home() {
  const campusQuery = useQuery(campusQueryOptions('futa'))
  const discoveryQuery = useQuery({
    ...discoveryQueryOptions({
      campus_id: campusQuery.data?.id ?? '',
      availability: 'available',
      sort: 'recommended',
      page: 1,
      page_size: 20,
    }),
    enabled: campusQuery.isSuccess,
  })

  return (
    <AppShell>
      <section className="border-border border-b pb-5 sm:pb-7">
        <p className="font-body text-faint mb-2 text-xs font-semibold tracking-[0.12em] uppercase">
          FUTA student accommodation
        </p>
        <h1 className="font-display text-foreground max-w-2xl text-[2rem] leading-[0.98] font-bold tracking-[-0.045em] sm:text-5xl">
          Find accommodation
          <br />
          near <span className="text-muted">FUTA, Akure.</span>
        </h1>
      </section>

      <section className="pt-5 sm:pt-7" aria-labelledby="discovery-heading">
        <div className="mb-3 flex items-end justify-between gap-4">
          <div>
            <p className="font-body text-faint text-xs">
              {discoveryQuery.data
                ? resultCountLabel(discoveryQuery.data.metadata.total_records)
                : 'Available around campus'}
            </p>
            <h2
              id="discovery-heading"
              className="font-display text-foreground mt-1 text-lg font-bold tracking-[-0.025em]"
            >
              Accommodation options
            </h2>
          </div>
        </div>

        {campusQuery.isPending || discoveryQuery.isPending ? <DiscoverySkeletons /> : null}

        {campusQuery.isError || discoveryQuery.isError ? (
          <FeedbackState
            title="We could not load accommodation"
            description="Check your connection and try again."
            actionLabel="Try again"
            onAction={() => {
              void campusQuery.refetch()
              void discoveryQuery.refetch()
            }}
          />
        ) : null}

        {discoveryQuery.data?.results.length ? (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 md:gap-5">
            {discoveryQuery.data.results.map((result) => (
              <DiscoveryCard key={`${result.property.id}:${result.unit_type.id}`} result={result} />
            ))}
          </div>
        ) : null}

        {discoveryQuery.isSuccess && discoveryQuery.data.results.length === 0 ? (
          <FeedbackState
            title="No accommodation is available yet"
            description="New FUTA inventory will appear here as agents make offers available."
          />
        ) : null}
      </section>
    </AppShell>
  )
}

function DiscoverySkeletons() {
  return (
    <div
      className="grid grid-cols-1 gap-4 md:grid-cols-2 md:gap-5"
      aria-label="Loading accommodation"
    >
      {[0, 1, 2, 3].map((item) => (
        <div key={item} className="bg-surface rounded-card overflow-hidden">
          <Skeleton className="aspect-[16/10] w-full rounded-none" />
          <div className="space-y-3 p-4">
            <Skeleton className="h-5 w-3/5" />
            <Skeleton className="h-3 w-2/5" />
            <Skeleton className="h-3 w-4/5" />
            <div className="border-border flex justify-between border-t pt-3">
              <Skeleton className="h-6 w-28" />
              <Skeleton className="h-5 w-20" />
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}

function resultCountLabel(count: number) {
  return `${count.toLocaleString('en-NG')} ${count === 1 ? 'option' : 'options'} around FUTA`
}
