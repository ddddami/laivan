import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider, createMemoryHistory, createRouter } from '@tanstack/react-router'
import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { ApiError, type PropertyDetail } from '../../api/client'
import { publicApiClient } from '../../api/queries'
import { routeTree } from '../../routeTree.gen'

const property: PropertyDetail = {
  id: 'property-1',
  campus_id: 'campus-1',
  name: 'Alice Lodge',
  area: 'Obanla',
  landmark: 'Near South Gate',
  description: 'A factual property description.',
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
  unit_types: [
    {
      id: 'unit-1',
      property_id: 'property-1',
      category: 'self_contained',
      name: 'Premium self-contained',
      description: 'The selected unit.',
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
          id: 'offer-1',
          property_unit_type_id: 'unit-1',
          agent_id: 'agent-1',
          title: 'Selected unit offer',
          description: 'Offer attached to the selected unit.',
          price_naira: 350_000,
          status: 'available',
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
      name: 'Standard single room',
      description: 'A different unit.',
      bedroom_count: 1,
      has_parlour: false,
      bathroom_type: 'shared',
      kitchen_type: 'shared',
      media: [],
      agent_offers: [
        {
          id: 'offer-2',
          property_unit_type_id: 'unit-2',
          agent_id: 'agent-2',
          title: 'Different unit offer',
          description: 'Must not appear for unit one.',
          price_naira: 180_000,
          status: 'available',
          agent: { id: 'agent-2', display_name: 'Bola Ajayi' },
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

afterEach(() => vi.restoreAllMocks())

describe('property and unit routes', () => {
  it('loads the requested property through the public API client', async () => {
    const getProperty = vi.spyOn(publicApiClient, 'getProperty').mockResolvedValue(property)
    await renderRoute('/properties/property-1')

    expect(await screen.findByRole('heading', { name: 'Alice Lodge' })).toBeInTheDocument()
    expect(
      screen.getByRole('heading', { name: 'Accommodation options at Alice Lodge' }),
    ).toBeInTheDocument()
    expect(screen.getByText('Accommodation types')).toBeInTheDocument()
    expect(screen.getByText('A factual property description.')).toBeInTheDocument()
    expect(getProperty).toHaveBeenCalledWith('property-1')
  })

  it('selects only the requested unit and its offers from the property response', async () => {
    vi.spyOn(publicApiClient, 'getProperty').mockResolvedValue(property)
    await renderRoute('/properties/property-1/unit-types/unit-1')

    expect(
      await screen.findByRole('heading', { name: 'Premium self-contained' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Alice Lodge' })).toHaveAttribute(
      'href',
      '/properties/property-1',
    )
    expect(screen.getByRole('img', { name: 'Bright self-contained room' })).toHaveAttribute(
      'src',
      'https://media.example.test/unit-medium.webp',
    )
    expect(screen.getByLabelText('Ade Martins offer')).toBeInTheDocument()
    expect(screen.queryByText('Bola Ajayi')).not.toBeInTheDocument()
  })

  it('shows property context media when the requested unit has no unit photos', async () => {
    vi.spyOn(publicApiClient, 'getProperty').mockResolvedValue(property)
    await renderRoute('/properties/property-1/unit-types/unit-2')

    expect(await screen.findByRole('heading', { name: 'Standard single room' })).toBeInTheDocument()
    expect(screen.getByRole('img', { name: 'Alice Lodge compound' })).toHaveAttribute(
      'src',
      'https://media.example.test/property-medium.webp',
    )
    expect(screen.getByText('Property context photos')).toBeInTheDocument()
  })

  it('distinguishes a missing property from an unknown unit within a property', async () => {
    const getProperty = vi.spyOn(publicApiClient, 'getProperty').mockRejectedValue(
      new ApiError({
        kind: 'response',
        code: 'property_not_found',
        message: 'Property not found',
        status: 404,
      }),
    )
    const missingProperty = await renderRoute('/properties/missing')
    expect(await screen.findByText('Property not found')).toBeInTheDocument()

    missingProperty.unmount()
    getProperty.mockResolvedValue(property)
    await renderRoute('/properties/property-1/unit-types/missing-unit')
    expect(await screen.findByText('Unit type not found')).toBeInTheDocument()
    expect(screen.getByText('This unit type is not part of Alice Lodge.')).toBeInTheDocument()
  })

  it('retries a failed property request', async () => {
    const getProperty = vi
      .spyOn(publicApiClient, 'getProperty')
      .mockRejectedValueOnce(new Error('Temporary failure'))
      .mockResolvedValueOnce(property)
    await renderRoute('/properties/property-1')

    expect(await screen.findByText('Property unavailable')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }))

    expect(await screen.findByRole('heading', { name: 'Alice Lodge' })).toBeInTheDocument()
    expect(getProperty).toHaveBeenCalledTimes(2)
  })
})

async function renderRoute(initialEntry: string) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const router = createRouter({
    routeTree,
    history: createMemoryHistory({ initialEntries: [initialEntry] }),
  })
  await router.load()
  return render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}
