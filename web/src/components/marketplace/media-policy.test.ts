import { describe, expect, it } from 'vitest'

import type { Media } from '../../api/client'
import { selectAccommodationMedia } from './media-policy'

const unitImage = (id: string): Media => ({
  id,
  property_id: null,
  property_unit_type_id: 'unit-1',
  agent_offer_id: null,
  uploaded_by_agent_id: 'agent-1',
  url: `https://media.example.test/${id}.jpg`,
  thumbnail_url: `https://media.example.test/${id}-thumb.webp`,
  medium_url: `https://media.example.test/${id}-medium.webp`,
  kind: 'image',
  caption: id,
  content_type: 'image/jpeg',
  size_bytes: 1024,
  created_at: '2026-05-01T10:00:00Z',
})

const propertyImage = (id: string): Media => ({
  ...unitImage(id),
  property_id: 'property-1',
  property_unit_type_id: null,
})

describe('selectAccommodationMedia', () => {
  it('prefers all unit-type images over property context images', () => {
    const selection = selectAccommodationMedia(
      [unitImage('unit-1'), unitImage('unit-2')],
      [propertyImage('property-1')],
    )

    expect(selection.source).toBe('unit_type')
    expect(selection.items.map((item) => item.id)).toEqual(['unit-1', 'unit-2'])
  })

  it('uses property images when the unit type has no images', () => {
    const selection = selectAccommodationMedia([], [propertyImage('property-1')])

    expect(selection.source).toBe('property')
    expect(selection.items.map((item) => item.id)).toEqual(['property-1'])
  })

  it('does not use offer images as aggregate accommodation media', () => {
    const offerImage = {
      ...unitImage('offer-1'),
      property_unit_type_id: null,
      agent_offer_id: 'offer-1',
    }
    const selection = selectAccommodationMedia([], [offerImage])

    expect(selection.source).toBe('none')
    expect(selection.items).toHaveLength(0)
  })

  it('ignores non-image media at every layer', () => {
    const video = { ...unitImage('unit-video'), kind: 'video' as const }
    const selection = selectAccommodationMedia([video], [video])

    expect(selection.source).toBe('none')
    expect(selection.items).toHaveLength(0)
  })
})
