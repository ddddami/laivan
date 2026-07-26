import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider, createMemoryHistory, createRouter } from '@tanstack/react-router'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { routeTree } from '../routeTree.gen'

describe('public detail navigation', () => {
  it('returns directly from an accommodation result to the preserved discovery URL', async () => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    const router = createRouter({
      routeTree,
      history: createMemoryHistory({
        initialEntries: ['/?area=Obanla&availability=all&sort=-created_at&page=2'],
      }),
    })

    await router.load()
    await router.navigate({
      to: '/properties/$propertyId/unit-types/$unitTypeId',
      params: { propertyId: 'property-1', unitTypeId: 'unit-1' },
      state: (current) => ({ ...current, entryPoint: 'discovery' }),
    })

    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Back to results' }))
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/')
      expect(router.state.location.search).toEqual({
        area: 'Obanla',
        availability: 'all',
        sort: '-created_at',
        page: 2,
      })
    })
  })

  it('returns from a unit through its property to the preserved discovery URL', async () => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    const router = createRouter({
      routeTree,
      history: createMemoryHistory({
        initialEntries: ['/?area=Obanla&availability=all&sort=-created_at&page=2'],
      }),
    })

    await router.load()
    await router.navigate({
      to: '/properties/$propertyId',
      params: { propertyId: 'property-1' },
    })
    await router.navigate({
      to: '/properties/$propertyId/unit-types/$unitTypeId',
      params: { propertyId: 'property-1', unitTypeId: 'unit-1' },
    })

    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Back to property' }))
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/properties/property-1')
    })

    fireEvent.click(screen.getByRole('button', { name: 'Back to results' }))
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/')
      expect(router.state.location.search).toEqual({
        area: 'Obanla',
        availability: 'all',
        sort: '-created_at',
        page: 2,
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
