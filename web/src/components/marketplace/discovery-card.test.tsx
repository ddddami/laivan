import {
  Outlet,
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from '@tanstack/react-router'
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import type { DiscoveryResult } from '../../api/client'
import { DiscoveryCard } from './discovery-card'

const result: DiscoveryResult = {
  property: {
    id: '550e8400-e29b-41d4-a716-446655440010',
    name: 'Alice Lodge',
    area: 'Obanla',
    landmark: 'Near South Gate',
  },
  unit_type: {
    id: '550e8400-e29b-41d4-a716-446655440020',
    category: 'self_contained',
    name: 'Self-contained',
    bedroom_count: 1,
    bathroom_type: 'private',
    kitchen_type: 'private',
  },
  pricing: { lowest_price_naira: 350_000 },
  offer_summary: { available_offer_count: 2 },
  thumbnail_url: null,
  created_at: '2026-05-01T10:00:00Z',
  updated_at: '2026-05-02T10:00:00Z',
}

describe('DiscoveryCard', () => {
  it('presents the property and unit hierarchy as a semantic property link', async () => {
    const rootRoute = createRootRoute({ component: Outlet })
    const indexRoute = createRoute({
      getParentRoute: () => rootRoute,
      path: '/',
      component: () => <DiscoveryCard result={result} />,
    })
    const propertyRoute = createRoute({
      getParentRoute: () => rootRoute,
      path: '/properties/$propertyId',
      component: () => null,
    })
    const router = createRouter({
      routeTree: rootRoute.addChildren([indexRoute, propertyRoute]),
      history: createMemoryHistory({ initialEntries: ['/'] }),
    })

    await router.load()
    render(<RouterProvider router={router} />)

    expect(screen.getByRole('heading', { name: 'Alice Lodge' })).toBeInTheDocument()
    expect(screen.getByText('Self-contained')).toBeInTheDocument()
    expect(screen.getByText('Obanla · Near South Gate')).toBeInTheDocument()
    expect(screen.getByText('₦350,000')).toBeInTheDocument()
    expect(screen.getByText('2 offers')).toBeInTheDocument()
    expect(screen.getByRole('img', { name: 'No photos available' })).toBeInTheDocument()
    expect(screen.getByRole('link')).toHaveAttribute(
      'href',
      '/properties/550e8400-e29b-41d4-a716-446655440010',
    )
  })
})
