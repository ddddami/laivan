import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider, createMemoryHistory, createRouter } from '@tanstack/react-router'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { DiscoveryResponse, PropertyDetail } from '../../api/client'
import { publicApiClient } from '../../api/queries'
import { routeTree } from '../../routeTree.gen'

afterEach(() => vi.restoreAllMocks())

describe('public detail navigation', () => {
  it('returns from discovery through an accommodation and property to the original results', async () => {
    vi.spyOn(publicApiClient, 'getCampus').mockResolvedValue({
      id: 'campus-1',
      slug: 'futa',
      name: 'Federal University of Technology, Akure',
      short_name: 'FUTA',
    })
    vi.spyOn(publicApiClient, 'discover').mockResolvedValue(discoveryResponse)
    vi.spyOn(publicApiClient, 'getProperty').mockResolvedValue(property)
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    const router = createRouter({
      routeTree,
      history: createMemoryHistory({
        initialEntries: ['/?area=Obanla&availability=all&sort=-created_at&page=2'],
      }),
      scrollRestoration: true,
    })

    await router.load()
    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    )

    vi.spyOn(window, 'scrollX', 'get').mockReturnValue(0)
    vi.spyOn(window, 'scrollY', 'get').mockReturnValue(640)
    fireEvent.scroll(document)

    fireEvent.click(await screen.findByRole('link', { name: /Self-contained.*Alice Lodge/ }))
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/properties/property-1/unit-types/unit-1')
    })
    fireEvent.click(await screen.findByRole('link', { name: 'Alice Lodge' }))
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/properties/property-1')
    })
    expect(await screen.findByRole('heading', { name: 'Alice Lodge' })).toBeInTheDocument()

    const scrollTo = vi.mocked(window.scrollTo)
    scrollTo.mockClear()
    fireEvent.click(await screen.findByRole('button', { name: 'Back to results' }))
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/')
      expect(router.state.location.search).toEqual({
        area: 'Obanla',
        availability: 'all',
        sort: '-created_at',
        page: 2,
      })
    })
    await waitFor(() => {
      expect(scrollTo).toHaveBeenCalledWith({
        top: 640,
        left: 0,
        behavior: undefined,
      })
    })
  })

  it('uses safe in-app fallbacks for a directly opened unit URL', async () => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    const router = createRouter({
      routeTree,
      history: createMemoryHistory({
        initialEntries: ['/properties/property-1/unit-types/unit-1'],
      }),
    })

    await router.load()
    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Back to property' }))
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/properties/property-1')
      expect(router.state.location.state.__TSR_index).toBe(0)
    })

    fireEvent.click(screen.getByRole('button', { name: 'Back to results' }))
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/')
    })
  })
})

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
    current_page: 2,
    page_size: 20,
    first_page: 1,
    last_page: 2,
    total_records: 21,
  },
}

const property: PropertyDetail = {
  id: 'property-1',
  campus_id: 'campus-1',
  name: 'Alice Lodge',
  area: 'Obanla',
  landmark: 'Near South Gate',
  description: 'A factual property description.',
  version: 1,
  media: [],
  unit_types: [
    {
      id: 'unit-1',
      property_id: 'property-1',
      category: 'self_contained',
      name: 'Self-contained',
      description: 'Private room with its own facilities.',
      bedroom_count: 1,
      has_parlour: false,
      bathroom_type: 'private',
      kitchen_type: 'private',
      media: [],
      agent_offers: [
        {
          id: 'offer-1',
          property_unit_type_id: 'unit-1',
          agent_id: 'agent-1',
          title: 'Alice self-contained',
          description: 'Annual offer for this accommodation.',
          price_naira: 350_000,
          status: 'available',
          version: 1,
          agent: { id: 'agent-1', display_name: 'Ade Martins' },
          media: [],
          created_at: '2026-05-01T10:00:00Z',
          updated_at: '2026-05-02T10:00:00Z',
        },
      ],
      created_at: '2026-05-01T10:00:00Z',
      updated_at: '2026-05-02T10:00:00Z',
    },
  ],
  created_at: '2026-05-01T10:00:00Z',
  updated_at: '2026-05-02T10:00:00Z',
}
