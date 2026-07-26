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

import type { PropertyDetail as PropertyDetailData } from '../../api/client'
import { PropertyDetail } from './property-detail'

const property: PropertyDetailData = {
  id: '550e8400-e29b-41d4-a716-446655440010',
  campus_id: '550e8400-e29b-41d4-a716-446655440002',
  name: 'Alice Lodge',
  area: 'Obanla',
  landmark: 'Near South Gate',
  description: 'A quiet building with borehole water.',
  media: [],
  created_at: '2026-05-01T10:00:00Z',
  updated_at: '2026-05-02T10:00:00Z',
  unit_types: [
    {
      id: '550e8400-e29b-41d4-a716-446655440020',
      property_id: '550e8400-e29b-41d4-a716-446655440010',
      category: 'self_contained',
      name: 'Premium self-contained',
      description: 'Private room and facilities.',
      bedroom_count: 1,
      has_parlour: false,
      bathroom_type: 'private',
      kitchen_type: 'private',
      media: [],
      agent_offers: [
        {
          id: '550e8400-e29b-41d4-a716-446655440030',
          property_unit_type_id: '550e8400-e29b-41d4-a716-446655440020',
          agent_id: '550e8400-e29b-41d4-a716-446655440040',
          title: 'Available self-contained',
          description: '',
          price_naira: 350_000,
          status: 'available',
          agent: {
            id: '550e8400-e29b-41d4-a716-446655440040',
            display_name: 'Ade Agent',
          },
          media: [],
          created_at: '2026-05-01T10:00:00Z',
          updated_at: '2026-05-02T10:00:00Z',
        },
      ],
      created_at: '2026-05-01T10:00:00Z',
      updated_at: '2026-05-02T10:00:00Z',
    },
    {
      id: '550e8400-e29b-41d4-a716-446655440021',
      property_id: '550e8400-e29b-41d4-a716-446655440010',
      category: 'single_room',
      name: 'Single room',
      description: '',
      bedroom_count: 1,
      has_parlour: false,
      bathroom_type: 'shared',
      kitchen_type: 'shared',
      media: [],
      agent_offers: [],
      created_at: '2026-05-01T10:00:00Z',
      updated_at: '2026-05-02T10:00:00Z',
    },
  ],
}

describe('PropertyDetail', () => {
  it('keeps building facts above distinct unit types with nested destinations', async () => {
    const rootRoute = createRootRoute({ component: Outlet })
    const indexRoute = createRoute({
      getParentRoute: () => rootRoute,
      path: '/',
      component: () => <PropertyDetail property={property} />,
    })
    const unitRoute = createRoute({
      getParentRoute: () => rootRoute,
      path: '/properties/$propertyId/unit-types/$unitTypeId',
      component: () => null,
    })
    const router = createRouter({
      routeTree: rootRoute.addChildren([indexRoute, unitRoute]),
      history: createMemoryHistory({ initialEntries: ['/'] }),
    })

    await router.load()
    render(<RouterProvider router={router} />)

    expect(screen.getByRole('heading', { name: 'Alice Lodge' })).toBeInTheDocument()
    expect(screen.getByText('A quiet building with borehole water.')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Premium self-contained' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Single room' })).toBeInTheDocument()
    expect(
      screen.getByRole('img', {
        name: 'No property photos available for Alice Lodge',
      }),
    ).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Premium self-contained/ })).toHaveAttribute(
      'href',
      '/properties/550e8400-e29b-41d4-a716-446655440010/unit-types/550e8400-e29b-41d4-a716-446655440020',
    )
    expect(screen.getByText('Price unavailable')).toBeInTheDocument()
  })
})
