import { createFileRoute } from '@tanstack/react-router'

import type { DiscoverySearch } from '../features/discovery/search'
import { normalizeDiscoverySearch } from '../features/discovery/search'
import { Home } from '../pages/home'

export const Route = createFileRoute('/')({
  validateSearch: normalizeDiscoverySearch,
  component: DiscoveryRoute,
})

function DiscoveryRoute() {
  const search = Route.useSearch()
  const navigate = Route.useNavigate()

  function updateSearch(next: DiscoverySearch) {
    void navigate({ search: next })
  }

  return <Home search={search} onSearchChange={updateSearch} />
}
