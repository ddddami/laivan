import { describe, expect, it } from 'vitest'

import {
  activeFilterCount,
  defaultDiscoverySearch,
  discoverySearchError,
  normalizeDiscoverySearch,
} from './search'

describe('discovery search', () => {
  it('normalizes valid shareable filters', () => {
    expect(
      normalizeDiscoverySearch({
        category: 'self_contained, single_room,self_contained',
        area: ' South Gate ',
        min_price: '200000',
        max_price: 500_000,
        bathroom_type: 'private',
        kitchen_type: 'none',
        has_parlour: 'false',
        availability: 'all',
        sort: '-created_at',
        page: '3',
      }),
    ).toEqual({
      category: 'self_contained,single_room',
      area: 'South Gate',
      min_price: 200_000,
      max_price: 500_000,
      bathroom_type: 'private',
      kitchen_type: 'none',
      has_parlour: false,
      availability: 'all',
      sort: '-created_at',
      page: 3,
    })
  })

  it('falls back safely for invalid enum, price, boolean, and page values', () => {
    expect(
      normalizeDiscoverySearch({
        category: 'castle,single_room',
        min_price: '-1',
        max_price: '3.5',
        has_parlour: 'maybe',
        availability: 'confirmed',
        sort: 'popular',
        page: 0,
      }),
    ).toEqual({
      ...defaultDiscoverySearch,
      category: 'single_room',
      area: undefined,
      min_price: undefined,
      max_price: undefined,
      bathroom_type: undefined,
      kitchen_type: undefined,
      has_parlour: undefined,
    })
  })

  it('keeps a reversed valid price range visible for user correction', () => {
    const search = normalizeDiscoverySearch({ min_price: 500_000, max_price: 200_000 })

    expect(discoverySearchError(search)).toBe('Minimum price cannot be higher than maximum price.')
  })

  it('counts explicit marketplace filters without counting defaults', () => {
    expect(activeFilterCount(defaultDiscoverySearch)).toBe(0)
    expect(
      activeFilterCount({
        ...defaultDiscoverySearch,
        category: 'self_contained',
        area: 'Obanla',
        has_parlour: false,
        availability: 'all',
      }),
    ).toBe(4)
  })
})
