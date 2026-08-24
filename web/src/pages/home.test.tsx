import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  Outlet,
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from '@tanstack/react-router'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import type { DiscoveryResponse, PublicApiClient } from '../api/client'
import { normalizeDiscoverySearch, type DiscoverySearch } from '../features/discovery/search'
import { Home } from './home'

const discoveryResponse: DiscoveryResponse = {
  results: [
    {
      property: {
        id: 'property-1',
        name: 'Alice Lodge',
        area: 'Obanla',
        landmark: 'Near South Gate',
      },
      unit_type: {
        id: 'unit-1',
        category: 'self_contained',
        name: 'Self-contained',
        bedroom_count: 1,
        bathroom_type: 'private',
        kitchen_type: 'private',
      },
      pricing: { lowest_price_naira: 350_000 },
      offer_summary: { available_offer_count: 2 },
      thumbnail_url: null,
      thumbnail_source: 'none',
      created_at: '2026-05-01T10:00:00Z',
      updated_at: '2026-05-02T10:00:00Z',
    },
  ],
  metadata: {
    current_page: 3,
    page_size: 20,
    first_page: 1,
    last_page: 4,
    total_records: 67,
  },
}

async function renderHome(initialEntry: string, apiClient: PublicApiClient) {
  const rootRoute = createRootRoute({ component: Outlet })
  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    validateSearch: normalizeDiscoverySearch,
    component: TestHome,
  })
  const propertyRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/properties/$propertyId',
    component: () => null,
  })
  const router = createRouter({
    routeTree: rootRoute.addChildren([indexRoute, propertyRoute]),
    history: createMemoryHistory({ initialEntries: [initialEntry] }),
  })
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  function TestHome() {
    const search = indexRoute.useSearch()
    const navigate = indexRoute.useNavigate()
    return (
      <Home
        search={search}
        apiClient={apiClient}
        onSearchChange={(next: DiscoverySearch) => void navigate({ search: next })}
      />
    )
  }

  await router.load()
  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
  return router
}

describe('Home route behavior', () => {
  it('recovers URL filters, sends them to discovery, and resets the page on change', async () => {
    const discover = vi.fn(async () => discoveryResponse)
    const apiClient: PublicApiClient = {
      getCampus: async () => ({
        id: 'campus-1',
        slug: 'futa',
        name: 'Federal University of Technology, Akure',
        short_name: 'FUTA',
      }),
      discover,
      getProperty: async () => {
        throw new Error('Property is not used in this test')
      },
    }
    const router = await renderHome('/?area=Obanla&page=3', apiClient)

    await waitFor(() => {
      expect(discover).toHaveBeenCalledWith(
        expect.objectContaining({
          campus_id: 'campus-1',
          area: 'Obanla',
          page: 3,
        }),
      )
    })

    fireEvent.click(screen.getAllByRole('button', { name: 'Self-contained' })[0]!)

    await waitFor(() => {
      expect(router.state.location.search).toEqual(
        expect.objectContaining({
          area: 'Obanla',
          category: 'self_contained',
          page: 1,
        }),
      )
      expect(discover).toHaveBeenLastCalledWith(
        expect.objectContaining({
          area: 'Obanla',
          category: 'self_contained',
          page: 1,
        }),
      )
    })
  })

  it('searches by lodge name without clearing active filters', async () => {
    const discover = vi.fn(async () => discoveryResponse)
    const apiClient: PublicApiClient = {
      getCampus: async () => ({
        id: 'campus-1',
        slug: 'futa',
        name: 'Federal University of Technology, Akure',
        short_name: 'FUTA',
      }),
      discover,
      getProperty: async () => {
        throw new Error('Property is not used in this test')
      },
    }
    const router = await renderHome(
      '/?q=Alice%20Lodge&area=Obanla&category=self_contained&page=3',
      apiClient,
    )

    await waitFor(() => {
      expect(discover).toHaveBeenCalledWith(
        expect.objectContaining({
          q: 'Alice Lodge',
          area: 'Obanla',
          category: 'self_contained',
          page: 3,
        }),
      )
    })

    fireEvent.change(screen.getByRole('searchbox', { name: 'Search accommodation' }), {
      target: { value: 'Blue Roof' },
    })
    fireEvent.submit(screen.getByRole('search'))

    await waitFor(() => {
      expect(router.state.location.search).toEqual(
        expect.objectContaining({
          q: 'Blue Roof',
          area: 'Obanla',
          category: 'self_contained',
          page: 1,
        }),
      )
      expect(discover).toHaveBeenLastCalledWith(
        expect.objectContaining({
          q: 'Blue Roof',
          area: 'Obanla',
          category: 'self_contained',
          page: 1,
        }),
      )
    })
  })

  it('clears an invalid price range without requiring URL editing', async () => {
    const apiClient: PublicApiClient = {
      getCampus: async () => ({
        id: 'campus-1',
        slug: 'futa',
        name: 'Federal University of Technology, Akure',
        short_name: 'FUTA',
      }),
      discover: async () => discoveryResponse,
      getProperty: async () => {
        throw new Error('Property is not used in this test')
      },
    }
    const router = await renderHome('/?min_price=500000&max_price=200000', apiClient)

    expect(
      screen.getByText('Minimum price cannot be higher than maximum price.'),
    ).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Clear filters' }))

    await waitFor(() => {
      expect(router.state.location.search).toEqual({
        availability: 'available',
        sort: 'recommended',
        page: 1,
      })
    })
  })

  it('retries campus resolution without requesting discovery with an empty campus ID', async () => {
    const getCampus = vi
      .fn<PublicApiClient['getCampus']>()
      .mockRejectedValueOnce(new Error('Campus unavailable'))
      .mockResolvedValueOnce({
        id: 'campus-1',
        slug: 'futa',
        name: 'Federal University of Technology, Akure',
        short_name: 'FUTA',
      })
    const discover = vi.fn(async () => discoveryResponse)
    const apiClient: PublicApiClient = {
      getCampus,
      discover,
      getProperty: async () => {
        throw new Error('Property is not used in this test')
      },
    }
    await renderHome('/', apiClient)

    expect(await screen.findByText('We could not load accommodation')).toBeInTheDocument()
    expect(screen.queryByLabelText('Loading accommodation')).not.toBeInTheDocument()
    expect(discover).not.toHaveBeenCalled()

    fireEvent.click(screen.getByRole('button', { name: 'Try again' }))

    await waitFor(() => {
      expect(getCampus).toHaveBeenCalledTimes(2)
      expect(discover).toHaveBeenCalledWith(expect.objectContaining({ campus_id: 'campus-1' }))
    })
  })
})
