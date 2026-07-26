import { render, screen, within } from '@testing-library/react'
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
  media: [],
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
      media: [],
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
  it('keeps selected unit facts above only that unit’s competing agent offers', () => {
    render(<UnitDetail property={property} unit={property.unit_types[0]!} />)

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

  it('visually demotes paused offers while preserving factual comparison details', () => {
    render(<UnitDetail property={property} unit={property.unit_types[0]!} />)

    const availableOffer = screen.getByLabelText('Ade Martins offer')
    const pausedOffer = screen.getByLabelText('Bola Ajayi offer')

    expect(within(availableOffer).getByText('Available')).toBeInTheDocument()
    expect(within(availableOffer).getByText('₦350,000')).toBeInTheDocument()
    expect(within(availableOffer).getByText('Inspection on weekdays.')).toBeInTheDocument()
    expect(within(pausedOffer).getByText('Paused')).toBeInTheDocument()
    expect(within(pausedOffer).getByText('₦340,000')).toBeInTheDocument()
    expect(within(pausedOffer).getByText('Updated 3 May 2026')).toBeInTheDocument()
  })
})
