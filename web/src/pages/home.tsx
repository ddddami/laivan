import { useQuery } from '@tanstack/react-query'
import { FilterHorizontalIcon, Search01Icon } from '@hugeicons/core-free-icons'
import { type FormEvent, useEffect, useState } from 'react'

import type { PublicApiClient } from '../api/client'
import { campusQueryOptions, discoveryQueryOptions } from '../api/queries'
import { AppShell } from '../components/app-shell'
import { DiscoveryCard } from '../components/marketplace/discovery-card'
import { FilterChip } from '../components/ui/filter-chip'
import { FeedbackState } from '../components/ui/feedback-state'
import { IconButton } from '../components/ui/icon-button'
import { ProductIcon } from '../components/ui/product-icon'
import { Sheet } from '../components/ui/sheet'
import { Skeleton } from '../components/ui/skeleton'
import { StatusLine } from '../components/ui/status-line'
import { DiscoveryFilters } from '../features/discovery/discovery-filters'
import { Pagination } from '../features/discovery/pagination'
import {
  activeFilterCount,
  type DiscoverySearch,
  discoverySearchError,
  type DiscoverySort,
} from '../features/discovery/search'

type HomeProps = {
  search: DiscoverySearch
  onSearchChange: (search: DiscoverySearch) => void
  apiClient?: PublicApiClient
}

export function Home({ search, onSearchChange, apiClient }: HomeProps) {
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [query, setQuery] = useState(search.q ?? '')
  const searchError = discoverySearchError(search)
  const filterCount = activeFilterCount(search)
  const campusQuery = useQuery(campusQueryOptions('futa', apiClient))
  const discoveryQuery = useQuery({
    ...discoveryQueryOptions(
      {
        campus_id: campusQuery.data?.id ?? '',
        category: search.category,
        q: search.q,
        area: search.area,
        min_price: search.min_price,
        max_price: search.max_price,
        bathroom_type: search.bathroom_type,
        kitchen_type: search.kitchen_type,
        has_parlour: search.has_parlour,
        availability: search.availability,
        sort: search.sort,
        page: search.page,
        page_size: 20,
      },
      apiClient,
    ),
    enabled: campusQuery.isSuccess && !searchError,
  })
  const hardError =
    (campusQuery.isError && !campusQuery.data) || (discoveryQuery.isError && !discoveryQuery.data)

  useEffect(() => {
    setQuery(search.q ?? '')
  }, [search.q])

  function submitSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    onSearchChange({
      ...search,
      q: query.trim() || undefined,
      page: 1,
    })
  }

  function selectCategory(category?: string) {
    onSearchChange({
      ...search,
      category,
      page: 1,
    })
  }

  function clearFilters() {
    onSearchChange({
      q: search.q,
      availability: 'available',
      sort: search.sort,
      page: 1,
    })
  }

  function clearSearch() {
    setQuery('')
    onSearchChange({
      ...search,
      q: undefined,
      page: 1,
    })
  }

  function changePage(page: number) {
    onSearchChange({
      ...search,
      page,
    })
  }

  return (
    <AppShell>
      <section className="border-border border-b pb-5 sm:pb-7">
        <h1 className="font-display text-foreground max-w-2xl text-[2rem] leading-[0.98] font-bold tracking-[-0.045em] sm:text-5xl">
          Find accommodation
          <br />
          near <span className="text-muted">FUTA, Akure.</span>
        </h1>
      </section>

      <section className="border-border border-b py-4">
        <div className="flex items-center gap-2">
          <form role="search" className="relative min-w-0 flex-1" onSubmit={submitSearch}>
            <ProductIcon
              icon={Search01Icon}
              size={17}
              className="text-muted pointer-events-none absolute top-1/2 left-3.5 -translate-y-1/2"
            />
            <label htmlFor="accommodation-search" className="sr-only">
              Search accommodation
            </label>
            <input
              id="accommodation-search"
              type="search"
              value={query}
              maxLength={100}
              onChange={(event) => setQuery(event.target.value)}
              className="focus-ring bg-surface text-foreground placeholder:text-muted rounded-control min-h-11 w-full pl-10 text-sm"
              placeholder="Search lodges, areas, or room types"
              autoComplete="off"
            />
          </form>
          <IconButton
            icon={FilterHorizontalIcon}
            label={filterCount > 0 ? `Open filters, ${filterCount} active` : 'Open filters'}
            className={`${
              filterCount > 0 ? 'bg-foreground text-background hover:bg-foreground/85' : ''
            } lg:hidden`}
            onClick={() => setFiltersOpen(true)}
          />
        </div>

        <div className="scrollbar-none -mx-1 mt-3 flex gap-2 overflow-x-auto px-1">
          <FilterChip selected={!search.category} onClick={() => selectCategory()}>
            All
          </FilterChip>
          <FilterChip
            selected={search.category === 'single_room'}
            onClick={() => selectCategory('single_room')}
          >
            Single room
          </FilterChip>
          <FilterChip
            selected={search.category === 'self_contained'}
            onClick={() => selectCategory('self_contained')}
          >
            Self-contained
          </FilterChip>
          <FilterChip
            selected={search.category === 'room_and_parlour'}
            onClick={() => selectCategory('room_and_parlour')}
          >
            Room and parlour
          </FilterChip>
        </div>
      </section>

      <section className="pt-5 sm:pt-7" aria-labelledby="discovery-heading">
        <div className="mb-4 flex items-end justify-between gap-4">
          <div>
            <p className="font-body text-muted text-xs">
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
          <label className="font-body text-muted text-xs">
            <span className="sr-only">Sort accommodation</span>
            <select
              value={search.sort}
              className="focus-ring bg-surface text-foreground rounded-control min-h-11 border-0 px-3 text-xs font-semibold"
              onChange={(event) =>
                onSearchChange({
                  ...search,
                  sort: event.target.value as DiscoverySort,
                  page: 1,
                })
              }
            >
              <option value="recommended">Recommended</option>
              <option value="-created_at">Newest</option>
              <option value="lowest_price_naira">Lowest price</option>
              <option value="-lowest_price_naira">Highest price</option>
            </select>
          </label>
        </div>

        {searchError ? (
          <FeedbackState
            title="Check your price range"
            description={searchError}
            actionLabel="Clear filters"
            onAction={clearFilters}
          />
        ) : null}

        {!searchError &&
        !discoveryQuery.data &&
        (campusQuery.isPending || (campusQuery.isSuccess && discoveryQuery.isPending)) ? (
          <DiscoverySkeletons />
        ) : null}

        {!searchError && hardError ? (
          <FeedbackState
            title="We could not load accommodation"
            description="Check your connection and try again."
            actionLabel="Try again"
            onAction={() => {
              if (campusQuery.isError) {
                void campusQuery.refetch()
              } else {
                void discoveryQuery.refetch()
              }
            }}
          />
        ) : null}

        {!searchError && discoveryQuery.isError && discoveryQuery.data ? (
          <StatusLine
            tone="warning"
            actionLabel="Try again"
            onAction={() => void discoveryQuery.refetch()}
          >
            Showing saved results because the latest update could not be loaded.
          </StatusLine>
        ) : null}

        {!searchError &&
        discoveryQuery.data &&
        discoveryQuery.isFetching &&
        !discoveryQuery.isError ? (
          <StatusLine>
            {discoveryQuery.isPlaceholderData
              ? 'Loading the updated result set…'
              : 'Checking for newer accommodation information…'}
          </StatusLine>
        ) : null}

        <div
          className={
            searchError
              ? 'hidden'
              : 'lg:grid lg:grid-cols-[17rem_minmax(0,1fr)] lg:items-start lg:gap-7'
          }
        >
          <aside className="border-border rounded-card sticky top-5 hidden border lg:block">
            <div className="border-border border-b px-5 py-4">
              <h3 className="font-display text-foreground text-base font-bold">Filters</h3>
              <p className="font-body text-muted mt-1 text-xs">
                {filterCount} {filterCount === 1 ? 'filter' : 'filters'} active
              </p>
            </div>
            <DiscoveryFilters
              key={JSON.stringify(search)}
              search={search}
              onApply={onSearchChange}
            />
          </aside>

          <div>
            {discoveryQuery.data?.results.length ? (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2 md:gap-5 lg:grid-cols-1 xl:grid-cols-2">
                {discoveryQuery.data.results.map((result) => (
                  <DiscoveryCard
                    key={`${result.property.id}:${result.unit_type.id}`}
                    result={result}
                  />
                ))}
              </div>
            ) : null}

            {discoveryQuery.isSuccess && discoveryQuery.data.results.length === 0 ? (
              <FeedbackState
                title={
                  filterCount > 0
                    ? 'No accommodation matches'
                    : search.q
                      ? 'No accommodation found'
                      : 'No accommodation is available yet'
                }
                description={
                  filterCount > 0
                    ? 'Try widening your price range, changing the area, or clearing a filter.'
                    : search.q
                      ? 'Try another lodge, area, landmark, or room type.'
                      : 'New FUTA inventory will appear here as agents make offers available.'
                }
                actionLabel={
                  filterCount > 0 ? 'Clear filters' : search.q ? 'Clear search' : undefined
                }
                onAction={filterCount > 0 ? clearFilters : search.q ? clearSearch : undefined}
              />
            ) : null}

            {discoveryQuery.data && discoveryQuery.data.results.length > 0 ? (
              <Pagination
                currentPage={discoveryQuery.data.metadata.current_page}
                lastPage={discoveryQuery.data.metadata.last_page}
                pageSize={discoveryQuery.data.metadata.page_size}
                totalRecords={discoveryQuery.data.metadata.total_records}
                onPageChange={changePage}
              />
            ) : null}
          </div>
        </div>
      </section>

      <Sheet
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        title="Filters"
        description={`${filterCount} ${filterCount === 1 ? 'filter' : 'filters'} active`}
      >
        <DiscoveryFilters
          key={`${filtersOpen}:${JSON.stringify(search)}`}
          search={search}
          onApply={onSearchChange}
          onClose={() => setFiltersOpen(false)}
        />
      </Sheet>
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
