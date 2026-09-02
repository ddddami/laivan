import {
  Outlet,
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from '@tanstack/react-router'
import { fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import type { PropertyDetail } from '../../api/client'
import { UnitDetail } from './unit-detail'

const property: PropertyDetail = {
  id: 'property-1',
  campus_id: 'campus-1',
  name: 'Alice Lodge',
  area: 'Obanla',
  landmark: 'Near South Gate',
  description: 'Building-level description must stay on the property page.',
  version: 1,
  media: [
    {
      id: 'property-media-1',
      property_id: 'property-1',
      property_unit_type_id: null,
      agent_offer_id: null,
      uploaded_by_agent_id: 'agent-1',
      url: 'https://media.example.test/property.jpg',
      thumbnail_url: 'https://media.example.test/property-thumb.webp',
      medium_url: 'https://media.example.test/property-medium.webp',
      kind: 'image',
      caption: 'Alice Lodge compound',
      content_type: 'image/jpeg',
      size_bytes: 2048,
      created_at: '2026-05-01T10:00:00Z',
    },
  ],
  created_at: '2026-05-01T10:00:00Z',
  updated_at: '2026-05-02T10:00:00Z',
  unit_types: [
    {
      id: 'unit-1',
      property_id: 'property-1',
      category: 'self_contained',
      name: 'Premium self-contained',
      description: 'Private room with its own facilities.',
      notes: 'Top-floor corner unit.',
      bedroom_count: 1,
      has_parlour: false,
      bathroom_type: 'private',
      kitchen_type: 'private',
      media: [
        {
          id: 'unit-media-1',
          property_id: null,
          property_unit_type_id: 'unit-1',
          agent_offer_id: null,
          uploaded_by_agent_id: 'agent-1',
          url: 'https://media.example.test/unit.jpg',
          thumbnail_url: 'https://media.example.test/unit-thumb.webp',
          medium_url: 'https://media.example.test/unit-medium.webp',
          kind: 'image',
          caption: 'Bright self-contained room',
          content_type: 'image/jpeg',
          size_bytes: 1024,
          created_at: '2026-05-01T10:00:00Z',
        },
      ],
      agent_offers: [
        {
          id: 'offer-paused',
          property_unit_type_id: 'unit-1',
          agent_id: 'agent-2',
          title: 'Corner unit offer',
          description: 'This offer is temporarily paused.',
          notes: 'Inspection schedule is not open.',
          price_naira: 340_000,
          status: 'paused',
          version: 1,
          agent: { id: 'agent-2', display_name: 'Bola Ajayi' },
          media: [],
          created_at: '2026-05-01T10:00:00Z',
          updated_at: '2026-05-03T10:00:00Z',
        },
        {
          id: 'offer-available',
          property_unit_type_id: 'unit-1',
          agent_id: 'agent-1',
          title: 'Available top-floor unit',
          description: 'Annual offer for the selected unit.',
          notes: 'Inspection on weekdays.',
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
    {
      id: 'unit-2',
      property_id: 'property-1',
      category: 'single_room',
      name: 'Single room',
      description: '',
      bedroom_count: 1,
      has_parlour: false,
      bathroom_type: 'shared',
      kitchen_type: 'shared',
      media: [],
      agent_offers: [
        {
          id: 'wrong-offer',
          property_unit_type_id: 'unit-2',
          agent_id: 'agent-3',
          title: 'Different unit offer',
          description: '',
          price_naira: 180_000,
          status: 'available',
          version: 1,
          agent: { id: 'agent-3', display_name: 'Wrong Unit Agent' },
          media: [],
          created_at: '2026-05-01T10:00:00Z',
          updated_at: '2026-05-02T10:00:00Z',
        },
      ],
      created_at: '2026-05-01T10:00:00Z',
      updated_at: '2026-05-02T10:00:00Z',
    },
  ],
}

describe('UnitDetail', () => {
  it('keeps selected unit facts above only that unit’s competing agent offers', async () => {
    await renderUnitDetail()

    expect(screen.getByRole('heading', { name: 'Premium self-contained' })).toBeInTheDocument()
    expect(screen.getByText('Top-floor corner unit.')).toBeInTheDocument()
    expect(screen.getByText('1 available', { exact: false })).toBeInTheDocument()
    expect(screen.getByLabelText('Ade Martins offer')).toBeInTheDocument()
    expect(screen.getByLabelText('Bola Ajayi offer')).toBeInTheDocument()
    expect(screen.queryByText('Wrong Unit Agent')).not.toBeInTheDocument()
    expect(
      screen.queryByText('Building-level description must stay on the property page.'),
    ).not.toBeInTheDocument()
  })

  it('renders media attached to the selected unit type', async () => {
    await renderUnitDetail()

    const image = screen.getByRole('img', { name: 'Bright self-contained room' })
    expect(image).toHaveAttribute('src', 'https://media.example.test/unit-medium.webp')
  })

  it('falls back to clearly labelled property media when the unit has no photos', async () => {
    await renderUnitDetail(property.unit_types[1]!)

    const image = screen.getByRole('img', { name: 'Alice Lodge compound' })
    expect(image).toHaveAttribute('src', 'https://media.example.test/property-medium.webp')
    expect(screen.getByText('Property context photos')).toBeInTheDocument()
  })

  it('visually demotes paused offers while preserving factual comparison details', async () => {
    await renderUnitDetail()

    const availableOffer = screen.getByLabelText('Ade Martins offer')
    const pausedOffer = screen.getByLabelText('Bola Ajayi offer')

    expect(within(availableOffer).getByText('Available')).toBeInTheDocument()
    expect(within(availableOffer).getByText('₦350,000')).toBeInTheDocument()
    expect(within(availableOffer).getByText('Inspection on weekdays.')).toBeInTheDocument()
    expect(within(pausedOffer).getByText('Paused')).toBeInTheDocument()
    expect(within(pausedOffer).getByText('₦340,000')).toBeInTheDocument()
    expect(within(pausedOffer).getByText('Updated 3 May 2026')).toBeInTheDocument()
    expect(
      within(pausedOffer).queryByRole('button', { name: 'Explore request options' }),
    ).not.toBeInTheDocument()
    fireEvent.click(within(pausedOffer).getByRole('button', { name: 'Preview save option' }))
    expect(screen.getByRole('button', { name: /Save property/ })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Request an inspection/ })).not.toBeInTheDocument()
  })

  it('previews structured workflows without submitting or exposing agent contact', async () => {
    await renderUnitDetail()

    fireEvent.click(screen.getByRole('button', { name: 'Explore request options' }))

    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('Alice Lodge · Premium self-contained')).toBeInTheDocument()
    expect(screen.getByText('via Ade Martins')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Save property/ })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Request an inspection/ }))

    expect(
      screen.getByText(
        'Inspection requests will collect a preferred time and preserve their workflow status.',
      ),
    ).toBeInTheDocument()
    expect(screen.getByText(/Nothing in this preview is submitted or saved/)).toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /WhatsApp/i })).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Close preview' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})

async function renderUnitDetail(unit = property.unit_types[0]!) {
  const rootRoute = createRootRoute({ component: Outlet })
  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    component: () => <UnitDetail property={property} unit={unit} />,
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
  return render(<RouterProvider router={router} />)
}
